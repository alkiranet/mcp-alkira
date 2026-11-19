// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"time"
)

// WriteOptions holds per-call options for Core.Create.
type WriteOptions struct{}

// ListEdits holds the add and remove maps of an update. Each key names a
// field and each value is a JSON array of items. add appends to a `set` field
// the items it does not hold yet; remove drops items from a `set` field, or
// drops elements by id from a top-level IDMatchedArrays field. Both apply to
// the list stored at the time of the call.
type ListEdits struct {
	Add    map[string]json.RawMessage
	Remove map[string]json.RawMessage
}

// Core implements Create, Get, Update, Delete, ListPaged and Confirm for one
// family of type T described by a Descriptor. Every method runs under the
// context deadline, or DefaultWriteTimeout (AK_MCP_WRITE_TIMEOUT) when the
// context has none.
type Core[T any] struct {
	ac    *AlkiraClient
	d     Descriptor
	rules []FieldRule
}

// NewCore returns a Core for T. It fails when T's ak tags or the descriptor are invalid.
func NewCore[T any](ac *AlkiraClient, d Descriptor) (*Core[T], error) {
	if ac == nil {
		return nil, errors.New("core: nil client")
	}
	if err := d.validate(); err != nil {
		return nil, err
	}
	rules, err := FieldRules[T]()
	if err != nil {
		return nil, err
	}
	return &Core[T]{ac: ac, d: d, rules: rules}, nil
}

// Descriptor returns the descriptor the core was built with.
func (c *Core[T]) Descriptor() Descriptor { return c.d }

// StateError refuses a write because of the stored state: a soft-deleted
// row (DocState), or an entity or network state the state guard rejects.
type StateError struct {
	ID           string
	DocState     string // MARKED_FOR_DELETION or DELETED when the row is soft-deleted
	State        string // entity state
	NetworkState string // set when the network state refused the update
}

func (e *StateError) Error() string {
	switch {
	case e.DocState != "":
		return fmt.Sprintf("id %s is %s", e.ID, e.DocState)
	case e.NetworkState != "":
		return fmt.Sprintf("id %s cannot change while the network state is %s and its state is %s", e.ID, e.NetworkState, e.State)
	}
	return fmt.Sprintf("id %s cannot change while its state is %s", e.ID, e.State)
}

// Is reports a match for ErrStateRefused.
func (e *StateError) Is(target error) bool { return target == ErrStateRefused }

// DuplicateError refuses a create whose uniqueness key is taken.
type DuplicateError struct {
	Key   string
	Value string
	ID    string // id of the existing object
	Type  string // connector type, for the connector name space
}

func (e *DuplicateError) Error() string {
	if e.Type != "" {
		return fmt.Sprintf("%s %q already exists (%s id %s)", e.Key, e.Value, e.Type, e.ID)
	}
	return fmt.Sprintf("%s %q already exists (id %s)", e.Key, e.Value, e.ID)
}

// Is reports a match for ErrDuplicate.
func (e *DuplicateError) Is(target error) bool { return target == ErrDuplicate }

// Outcome is the result of the confirming read after a write whose response was lost.
type Outcome string

// A write that failed before it was sent returns an error that errors.Is
// ErrNotSent, never an *OutcomeError: nothing was applied.
const (
	// OutcomeApplied: the confirming GET holds every patched value, added
	// item and removed item of the PUT.
	OutcomeApplied Outcome = "applied"
	// OutcomeMayStillApply: the confirming GET does not show the PUT. The
	// backend can commit after the client gives up (connector locks wait up
	// to 60 s), so this does not prove the PUT failed.
	OutcomeMayStillApply Outcome = "may still apply"
	// OutcomeCreated: the uniqueness-key lookup found the object.
	OutcomeCreated Outcome = "created"
	// OutcomeMayStillBeCreated: the lookup did not find the object. A POST
	// that was sent is never reported as not created.
	OutcomeMayStillBeCreated Outcome = "may still be created"
	// OutcomeDeleted: the confirming GET answered the family's not-found
	// signature, or the row is marked for deletion.
	OutcomeDeleted Outcome = "deleted"
	// OutcomeUnknown: the confirming read failed, or found the object after a DELETE.
	OutcomeUnknown Outcome = "unknown"
)

// OutcomeError reports a write whose response was lost (post-send transport
// failure, deadline, a 5xx on PUT or DELETE, or 500 "Request failed" on
// POST) and what the confirming read found. It matches ErrOutcomeUnknown
// unless the read settled the outcome (applied, created, deleted).
type OutcomeError struct {
	Method  string
	ID      string // entity id; for OutcomeCreated, the id found by the lookup
	Outcome Outcome
	Current []byte // body of the confirming read, when it returned one
	Err     error  // the write error
	// FlagsUnavailable reports that the kill switch could not be checked.
	FlagsUnavailable bool
}

func (e *OutcomeError) Error() string {
	return fmt.Sprintf("%s id %s: outcome %s: %v", e.Method, e.ID, e.Outcome, e.Err)
}

func (e *OutcomeError) Is(target error) bool {
	if target != ErrOutcomeUnknown {
		return false
	}
	switch e.Outcome {
	case OutcomeApplied, OutcomeCreated, OutcomeDeleted:
		return false
	}
	return true
}

func (e *OutcomeError) Unwrap() error { return e.Err }

// Create validates obj against its ak tags, refuses a duplicate of the
// uniqueness key, and POSTs obj with its unset and read-only fields left out.
// WriteResult.ID is the id of the response body, or, when the body has none,
// the id the uniqueness-key lookup finds. On a lost response whose lookup
// finds the object, the *OutcomeError (OutcomeCreated) comes with a
// WriteResult carrying that id.
// The returned object is nil when the response body is empty or does not
// decode into T.
func (c *Core[T]) Create(ctx context.Context, obj *T, opts WriteOptions) (*T, *WriteResult, error) {
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()

	if obj == nil {
		return nil, nil, fmt.Errorf("%w: nil object", ErrInvalidInput)
	}
	if err := validateValue(c.rules, reflect.ValueOf(obj), false, ""); err != nil {
		return nil, nil, err
	}
	raw, err := json.Marshal(obj)
	if err != nil {
		return nil, nil, fmt.Errorf("create: %w", err)
	}
	m, err := decodeObject(raw)
	if err != nil {
		return nil, nil, fmt.Errorf("create: %w", err)
	}
	stripReadOnly(c.rules, m, false)
	body, err := json.Marshal(m)
	if err != nil {
		return nil, nil, fmt.Errorf("create: %w", err)
	}
	flagsUnavailable, err := c.ac.CheckWritesAllowed(ctx)
	if err != nil {
		return nil, nil, err
	}

	key := c.keyValue(m)
	if key != "" {
		dup, err := c.lookup(ctx, key)
		if err != nil {
			return nil, nil, fmt.Errorf("duplicate check: %w", err)
		}
		if dup != nil {
			return nil, nil, dup
		}
	}

	res, err := c.ac.doWrite(ctx, "POST", c.d.collectionURI(c.ac), body)
	if err != nil {
		if !lostResponse(err, "POST") {
			return nil, nil, c.classify("POST", err)
		}
		oe := &OutcomeError{Method: "POST", Outcome: OutcomeUnknown, Err: err, FlagsUnavailable: flagsUnavailable}
		if key == "" {
			return nil, nil, oe
		}
		cctx, ccancel := confirmContext(ctx)
		defer ccancel()
		dup, lerr := c.lookup(cctx, key)
		switch {
		case lerr != nil:
		case dup != nil:
			oe.Outcome, oe.ID = OutcomeCreated, dup.ID
			return nil, &WriteResult{ID: dup.ID, RequestID: requestIdOf(err), FlagsUnavailable: flagsUnavailable}, oe
		default:
			oe.Outcome = OutcomeMayStillBeCreated
		}
		return nil, nil, oe
	}
	res.FlagsUnavailable = flagsUnavailable
	res.ID = idOf(res.Body)
	if res.ID == "" && key != "" {
		cctx, ccancel := confirmContext(ctx)
		defer ccancel()
		if dup, lerr := c.lookup(cctx, key); lerr == nil && dup != nil {
			res.ID = dup.ID
		}
	}
	return decodeEntity[T](res.Body), res, nil
}

// Get reads one object. It refuses an id of another family
// (ErrFamilyMismatch) and a soft-deleted row (*StateError), and returns the
// raw body next to the decoded object.
func (c *Core[T]) Get(ctx context.Context, id string) (*T, []byte, error) {
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()

	raw, _, err := c.preGet(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, raw, fmt.Errorf("GET id %s: %w", id, err)
	}
	return &out, raw, nil
}

// Update applies an RFC 7386 merge patch and the list edits. One GET serves
// the docState, family and state checks; the PUT body is the stored body
// without read-only fields, merged with the patch, so unknown and secret
// fields are resent unchanged. Arrays in IDMatchedArrays merge element by id,
// and RoutingProvenance fields are left out when every stored entry is
// default routing. A patch cannot set a `set` field: an array in a merge
// patch replaces the stored list, so lists change through edits only.
// WriteResult.Absent names the removed items the stored list did not hold.
// A nil or empty patch with edits is allowed.
func (c *Core[T]) Update(ctx context.Context, id string, patch json.RawMessage, edits ListEdits) (*T, *WriteResult, error) {
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()

	if len(bytes.TrimSpace(patch)) == 0 {
		patch = json.RawMessage(`{}`)
	}
	pm, err := validatePatch(c.rules, patch)
	if err != nil {
		return nil, nil, err
	}
	le, err := c.validateEdits(edits)
	if err != nil {
		return nil, nil, err
	}
	if len(pm) == 0 && len(le.add) == 0 && len(le.remove) == 0 {
		return nil, nil, fmt.Errorf("%w: nothing to change", ErrInvalidInput)
	}
	if id == "" {
		return nil, nil, fmt.Errorf("%w: empty id", ErrInvalidInput)
	}
	flagsUnavailable, err := c.ac.CheckWritesAllowed(ctx)
	if err != nil {
		return nil, nil, err
	}
	raw, stored, err := c.preGet(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if err := c.stateGuard(ctx, id, stored); err != nil {
		return nil, nil, err
	}
	merged, absent, err := c.merge(stored, pm, le)
	if err != nil {
		return nil, nil, err
	}
	// Elements removed by id are the only stored elements merged may lack.
	before := stored
	if len(le.remove) > 0 {
		before = cloneJSON(stored).(map[string]any)
		for k, ids := range le.remove {
			if l, held := before[k]; held && c.idMatchedTop(k) {
				before[k], _ = removeByID(l, ids)
			}
		}
	}
	if p := changedField(c.rules, before, merged, "", func(r FieldRule) bool { return r.Immutable }); p != "" {
		return nil, nil, fmt.Errorf("%w: %s cannot be changed after create", ErrInvalidInput, p)
	}
	if p := changedField(c.rules, before, merged, "", func(r FieldRule) bool { return r.Secret }); p != "" {
		return nil, nil, fmt.Errorf("%w: the update would change or drop the stored secret %s", ErrInvalidInput, p)
	}
	// List edits that change nothing send no PUT: a PUT can move the entity
	// back to PENDING.
	if len(pm) == 0 && c.editsUnchanged(stored, merged, le) {
		return decodeEntity[T](raw), &WriteResult{Body: raw, Previous: raw, Absent: absent, Unchanged: true, FlagsUnavailable: flagsUnavailable}, nil
	}
	body, err := json.Marshal(merged)
	if err != nil {
		return nil, nil, fmt.Errorf("update: %w", err)
	}

	res, err := c.ac.doWrite(ctx, "PUT", c.d.entityURI(c.ac, id), body)
	if err != nil {
		if !lostResponse(err, "PUT") {
			return nil, nil, c.classify("PUT", err)
		}
		oe := &OutcomeError{Method: "PUT", ID: id, Outcome: OutcomeUnknown, Err: err, FlagsUnavailable: flagsUnavailable}
		cctx, ccancel := confirmContext(ctx)
		defer ccancel()
		if exists, cur, cerr := c.Confirm(cctx, id); cerr == nil && exists {
			oe.Current = cur
			oe.Outcome = OutcomeMayStillApply
			if c.applied(cur, pm, le) {
				oe.Outcome = OutcomeApplied
			}
		}
		return nil, nil, oe
	}
	res.Previous = raw
	res.Absent = absent
	res.FlagsUnavailable = flagsUnavailable
	return decodeEntity[T](res.Body), res, nil
}

// DeleteOptions holds per-call options for Core.DeleteWith.
type DeleteOptions struct {
	// Check, when set, runs on the pre-GET body before the DELETE is sent. A
	// non-nil error is returned as is and no DELETE is sent.
	Check func(stored []byte) error
}

// Delete deletes one object after a GET that refuses another family's id
// and a soft-deleted row. A 404 counts as deleted only when it matches the
// family's DELETE not-found signature.
func (c *Core[T]) Delete(ctx context.Context, id string) (*WriteResult, error) {
	return c.DeleteWith(ctx, id, DeleteOptions{})
}

// DeleteWith is Delete with options.
func (c *Core[T]) DeleteWith(ctx context.Context, id string, opts DeleteOptions) (*WriteResult, error) {
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()

	if id == "" {
		return nil, fmt.Errorf("%w: empty id", ErrInvalidInput)
	}
	flagsUnavailable, err := c.ac.CheckWritesAllowed(ctx)
	if err != nil {
		return nil, err
	}
	raw, _, err := c.preGet(ctx, id)
	if err != nil {
		return nil, err
	}
	if opts.Check != nil {
		if err := opts.Check(raw); err != nil {
			return nil, err
		}
	}
	res, err := c.ac.doWrite(ctx, "DELETE", c.d.entityURI(c.ac, id), nil)
	if err != nil {
		var ae *APIError
		if errors.As(err, &ae) && ae.Status == 404 && ae.matches(c.d.notFound("DELETE")) {
			return &WriteResult{Body: []byte(ae.Body), Status: 404, RequestID: ae.RequestID, Previous: raw, FlagsUnavailable: flagsUnavailable}, nil
		}
		if !lostResponse(err, "DELETE") {
			return nil, c.classify("DELETE", err)
		}
		oe := &OutcomeError{Method: "DELETE", ID: id, Outcome: OutcomeUnknown, Err: err, FlagsUnavailable: flagsUnavailable}
		cctx, ccancel := confirmContext(ctx)
		defer ccancel()
		if exists, cur, cerr := c.Confirm(cctx, id); cerr == nil {
			oe.Current = cur
			if !exists {
				oe.Outcome = OutcomeDeleted
			}
		}
		return nil, oe
	}
	res.Previous = raw
	res.FlagsUnavailable = flagsUnavailable
	return res, nil
}

// ListPaged reads the whole collection with ?paginated=true&limit=100,
// paging by offset until pagination.hits.
func (c *Core[T]) ListPaged(ctx context.Context) ([]T, error) {
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()

	rows, err := c.ac.listPaged(ctx, c.d.collectionURI(c.ac))
	if err != nil {
		return nil, err
	}
	out := make([]T, len(rows))
	for i, r := range rows {
		if err := json.Unmarshal(r, &out[i]); err != nil {
			return nil, fmt.Errorf("list: row %d: %w", i, err)
		}
	}
	return out, nil
}

// Confirm sends one confirming GET. exists is false when the response
// matches the family's not-found signature or the row is soft-deleted; any
// other failure, an HTML or gateway 404 included, is returned as err.
func (c *Core[T]) Confirm(ctx context.Context, id string) (exists bool, body []byte, err error) {
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()

	raw, _, err := c.preGet(ctx, id)
	var se *StateError
	switch {
	case errors.Is(err, ErrNotFound):
		return false, nil, nil
	case errors.As(err, &se) && se.DocState != "":
		return false, raw, nil
	case err != nil:
		return false, nil, err
	}
	return true, raw, nil
}

// preGet reads one entity and runs the family and docState checks (D4, BC-R2).
func (c *Core[T]) preGet(ctx context.Context, id string) ([]byte, map[string]any, error) {
	if id == "" {
		return nil, nil, fmt.Errorf("%w: empty id", ErrInvalidInput)
	}
	uri := c.d.entityURI(c.ac, id)
	var q url.Values
	if c.d.MarkedRowLifecycle && c.d.DocState != DocStatePlainGet {
		q = url.Values{"includeMarkedForDeletion": {"true"}}
	}
	raw, err := c.ac.doGet(ctx, withQuery(uri, q))
	if err != nil {
		err = c.classify("GET", err)
		if c.d.DocState == DocStatePlainGet && c.d.MarkedRowLifecycle && errors.Is(err, ErrNotFound) {
			marked, merr := c.ac.doGet(ctx, withQuery(uri, url.Values{"includeMarkedForDeletion": {"true"}}))
			if merr == nil {
				return marked, nil, &StateError{ID: id, DocState: "MARKED_FOR_DELETION"}
			}
		}
		return nil, nil, err
	}
	m, err := decodeObject(raw)
	if err != nil {
		return nil, nil, fmt.Errorf("GET id %s: %w", id, err)
	}
	if f := c.d.FamilyType; f.Field != "" {
		v, ok := m[f.Field]
		if !ok || (f.Value != "" && fmt.Sprint(v) != f.Value) {
			return nil, nil, fmt.Errorf("%w: id %s is not a %s object", ErrFamilyMismatch, id, c.d.Collection)
		}
	}
	if c.d.DocState != DocStatePlainGet {
		v, ok := m["docState"]
		ds, isString := v.(string)
		switch {
		case !ok && c.d.DocState == DocStateField:
			return nil, nil, fmt.Errorf("%w: GET id %s has no docState", ErrStateRefused, id)
		case ok && !isString:
			return nil, nil, fmt.Errorf("%w: GET id %s has a docState that is not a string", ErrStateRefused, id)
		case ds == "MARKED_FOR_DELETION" || ds == "DELETED":
			return raw, m, &StateError{ID: id, DocState: ds}
		}
	}
	return raw, m, nil
}

// stateGuard runs the descriptor's state guard on the pre-GET body. The
// network state is read only when the entity state alone does not decide
// (BC-R10).
func (c *Core[T]) stateGuard(ctx context.Context, id string, stored map[string]any) error {
	if c.d.StateGuard == StateGuardNone {
		return nil
	}
	v, ok := stored["state"]
	state, isString := v.(string)
	if ok && !isString {
		return fmt.Errorf("%w: GET id %s has a state that is not a string", ErrStateRefused, id)
	}
	switch c.d.StateGuard {
	case StateGuardFailedOnly:
		if state == "FAILED" {
			return &StateError{ID: id, State: state}
		}
	case StateGuardD13:
		switch state {
		case "FAILED", "SCHEDULED", "IN_PROGRESS":
			return &StateError{ID: id, State: state}
		case "PENDING":
			return nil
		}
		raw, err := c.ac.doGet(ctx, c.ac.URI+"/tenantnetworks/"+c.ac.TenantNetworkId)
		if err != nil {
			return fmt.Errorf("network state: %w", err)
		}
		var nw struct {
			State string `json:"state"`
		}
		if err := json.Unmarshal(raw, &nw); err != nil {
			return fmt.Errorf("%w: network state: %v", ErrStateRefused, err)
		}
		if nw.State == "FAILED" {
			return &StateError{ID: id, State: state, NetworkState: nw.State}
		}
	}
	return nil
}

// merge builds the PUT body from the stored body, the patch and the list
// edits. It returns the removed items the stored lists did not hold.
func (c *Core[T]) merge(stored map[string]any, pm map[string]json.RawMessage, le listEdits) (map[string]any, map[string][]any, error) {
	base := cloneJSON(stored).(map[string]any)
	stripReadOnly(c.rules, base, false)
	for _, f := range c.d.RoutingProvenance {
		if _, set := pm[f]; !set && allDefaultRouting(base[f]) {
			delete(base, f)
		}
	}
	absent := map[string][]any{}
	// Elements removed by id leave the stored array before the merge, so the
	// every-stored-id rule of mergeByID does not ask for them.
	for k, ids := range le.remove {
		if !c.idMatchedTop(k) {
			continue
		}
		if err := listOrMissing(base, k); err != nil {
			return nil, nil, err
		}
		kept, missing := removeByID(base[k], ids)
		if _, held := base[k]; held {
			base[k] = kept
		}
		if len(missing) > 0 {
			absent[k] = missing
		}
	}
	patch := make(map[string]any, len(pm))
	for k, raw := range pm {
		v, err := decodeAny(raw)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %s: %v", ErrInvalidInput, k, err)
		}
		patch[k] = v
	}
	idPaths := make(map[string]bool, len(c.d.IDMatchedArrays))
	for _, p := range c.d.IDMatchedArrays {
		idPaths[p] = true
	}
	mv, err := mergeValue(base, patch, "", idPaths)
	if err != nil {
		return nil, nil, err
	}
	out := mv.(map[string]any)
	for k, items := range le.remove {
		if c.idMatchedTop(k) {
			continue
		}
		if err := listOrMissing(out, k); err != nil {
			return nil, nil, err
		}
		list, held := out[k].([]any)
		kept := make([]any, 0, len(list))
		for _, e := range list {
			if !containsJSON(items, e) {
				kept = append(kept, e)
			}
		}
		for _, it := range items {
			if !containsJSON(list, it) {
				absent[k] = append(absent[k], it)
			}
		}
		if held {
			out[k] = kept
		}
	}
	for k, items := range le.add {
		if err := listOrMissing(out, k); err != nil {
			return nil, nil, err
		}
		list, _ := out[k].([]any)
		list = append([]any{}, list...)
		for _, it := range items {
			if !containsJSON(list, it) {
				list = append(list, it)
			}
		}
		out[k] = list
	}
	for _, r := range c.rules {
		if l, ok := out[r.Name].([]any); ok && r.Required && (r.Set || c.idMatchedTop(r.Name)) && len(l) == 0 && (le.remove[r.Name] != nil) {
			return nil, nil, fmt.Errorf("%w: %s is required; removing every item would empty it", ErrInvalidInput, r.Name)
		}
	}
	if len(absent) == 0 {
		absent = nil
	}
	return out, absent, nil
}

// listOrMissing refuses a list edit on a stored value that is not a list.
func listOrMissing(m map[string]any, k string) error {
	v, ok := m[k]
	if _, isList := v.([]any); ok && v != nil && !isList {
		return fmt.Errorf("%w: %s is stored as %s, not a list", ErrInvalidInput, k, jsonText(v))
	}
	return nil
}

// editsUnchanged reports whether every edited list in merged equals the stored one.
func (c *Core[T]) editsUnchanged(stored, merged map[string]any, le listEdits) bool {
	for _, edits := range []map[string][]any{le.add, le.remove} {
		for k := range edits {
			if !jsonEqual(stored[k], merged[k]) {
				return false
			}
		}
	}
	return true
}

// listEdits is ListEdits after validation, items decoded.
type listEdits struct {
	add, remove map[string][]any
}

// validateEdits checks the list edits against the field rules: add names a
// `set` field and its items pass the field's element rules; remove names a
// `set` field or a top-level id-matched array; secret fields take no edits;
// an item cannot be both added and removed.
func (c *Core[T]) validateEdits(e ListEdits) (listEdits, error) {
	out := listEdits{add: map[string][]any{}, remove: map[string][]any{}}
	byName := make(map[string]FieldRule, len(c.rules))
	for _, r := range c.rules {
		byName[r.Name] = r
	}
	var errs []error
	decode := func(verb, k string, raw json.RawMessage) []any {
		v, err := decodeAny(raw)
		items, ok := v.([]any)
		if err != nil || !ok || len(items) == 0 {
			errs = append(errs, fmt.Errorf("%w: %s.%s must be a non-empty JSON array", ErrInvalidInput, verb, k))
			return nil
		}
		return items
	}
	checkSet := func(verb, k string, raw json.RawMessage) bool {
		r, ok := byName[k]
		switch {
		case !ok || !r.Set:
			errs = append(errs, fmt.Errorf("%w: %s.%s: %s is not a list that %s applies to", ErrInvalidInput, verb, k, k, verb))
			return false
		case r.ReadOnly || r.Immutable:
			errs = append(errs, fmt.Errorf("%w: %s.%s: %s cannot be changed", ErrInvalidInput, verb, k, k))
			return false
		case r.Secret:
			errs = append(errs, fmt.Errorf("%w: %s.%s: %s is a secret and cannot be changed", ErrInvalidInput, verb, k, k))
			return false
		}
		pv := reflect.New(r.Type)
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		if err := dec.Decode(pv.Interface()); err != nil {
			errs = append(errs, fmt.Errorf("%w: %s.%s: %v", ErrInvalidInput, verb, k, err))
			return false
		}
		// Removed items skip the element rules, so stored values that break
		// today's rules can still be removed.
		if verb == "add" {
			if err := validateField(r, pv.Elem(), true, verb+"."+k); err != nil {
				errs = append(errs, err)
				return false
			}
		}
		return true
	}
	for k, raw := range e.Add {
		if items := decode("add", k, raw); items != nil && checkSet("add", k, raw) {
			out.add[k] = items
		}
	}
	for k, raw := range e.Remove {
		items := decode("remove", k, raw)
		if items == nil {
			continue
		}
		if c.idMatchedTop(k) {
			if r, ok := byName[k]; ok && (r.ReadOnly || r.Immutable || r.Secret) {
				errs = append(errs, fmt.Errorf("%w: remove.%s: %s cannot be changed", ErrInvalidInput, k, k))
				continue
			}
			for i, it := range items {
				switch it.(type) {
				case string, json.Number:
				default:
					errs = append(errs, fmt.Errorf("%w: remove.%s[%d] must be an element id", ErrInvalidInput, k, i))
				}
			}
			out.remove[k] = items
			continue
		}
		if checkSet("remove", k, raw) {
			out.remove[k] = items
		}
	}
	for k, items := range out.add {
		for _, it := range items {
			if containsJSON(out.remove[k], it) {
				errs = append(errs, fmt.Errorf("%w: %s: %s is in both add and remove", ErrInvalidInput, k, jsonText(it)))
			}
		}
	}
	return out, errors.Join(errs...)
}

// idMatchedTop reports whether k is a top-level IDMatchedArrays field.
func (c *Core[T]) idMatchedTop(k string) bool {
	for _, p := range c.d.IDMatchedArrays {
		if p == k {
			return true
		}
	}
	return false
}

// removeByID drops the elements of list whose id is in ids and returns the
// ids no element held.
func removeByID(list any, ids []any) (any, []any) {
	stored, _ := list.([]any)
	drop := map[string]bool{}
	for _, id := range ids {
		drop[fmt.Sprint(id)] = true
	}
	kept := make([]any, 0, len(stored))
	found := map[string]bool{}
	for _, e := range stored {
		if m, ok := e.(map[string]any); ok && m["id"] != nil && drop[fmt.Sprint(m["id"])] {
			found[fmt.Sprint(m["id"])] = true
			continue
		}
		kept = append(kept, e)
	}
	var missing []any
	for _, id := range ids {
		if !found[fmt.Sprint(id)] {
			missing = append(missing, id)
		}
	}
	if list == nil && len(kept) == 0 {
		return nil, missing
	}
	return kept, missing
}

func containsJSON(list []any, v any) bool {
	for _, e := range list {
		if jsonEqual(e, v) {
			return true
		}
	}
	return false
}

func jsonText(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// mergeValue applies RFC 7386 to target, except that arrays at idPaths merge
// element by id: the patch must carry the id of every stored element, an
// element with a known id merges into it, one without id is new, and an
// unknown id is refused.
func mergeValue(target, patch any, path string, idPaths map[string]bool) (any, error) {
	switch p := patch.(type) {
	case map[string]any:
		t, ok := target.(map[string]any)
		if !ok {
			t = map[string]any{}
		} else {
			cp := make(map[string]any, len(t))
			for k, v := range t {
				cp[k] = v
			}
			t = cp
		}
		for k, v := range p {
			if v == nil {
				delete(t, k)
				continue
			}
			mv, err := mergeValue(t[k], v, joinPath(path, k), idPaths)
			if err != nil {
				return nil, err
			}
			t[k] = mv
		}
		return t, nil
	case []any:
		if idPaths[path] {
			return mergeByID(target, p, path, idPaths)
		}
	}
	return patch, nil
}

func joinPath(path, k string) string {
	if path == "" {
		return k
	}
	return path + "." + k
}

func mergeByID(target any, patch []any, path string, idPaths map[string]bool) (any, error) {
	stored, _ := target.([]any)
	byID := map[string]any{}
	for _, e := range stored {
		if m, ok := e.(map[string]any); ok && m["id"] != nil {
			byID[fmt.Sprint(m["id"])] = m
		}
	}
	sent := map[string]bool{}
	out := make([]any, 0, len(patch))
	for i, e := range patch {
		pm, ok := e.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%w: %s[%d] must be an object", ErrInvalidInput, path, i)
		}
		var base any
		if id, has := pm["id"]; has && id != nil {
			key := fmt.Sprint(id)
			s, known := byID[key]
			if !known {
				return nil, fmt.Errorf("%w: %s[%d] has unknown id %s", ErrInvalidInput, path, i, key)
			}
			if sent[key] {
				return nil, fmt.Errorf("%w: %s[%d] repeats id %s", ErrInvalidInput, path, i, key)
			}
			sent[key] = true
			base = s
		}
		mv, err := mergeValue(base, pm, path+"[]", idPaths)
		if err != nil {
			return nil, err
		}
		out = append(out, mv)
	}
	for id := range byID {
		if !sent[id] {
			return nil, fmt.Errorf("%w: %s must carry the id of every stored element; id %s is missing", ErrInvalidInput, path, id)
		}
	}
	return out, nil
}

// changedField returns the path of the first field that pick selects whose
// merged value differs from the stored one, or "". Read-only fields are
// skipped, since the merge strips them. Array elements pair by id, or by
// position when the stored element has no id; a stored element with no pair
// compares against nothing. A merged element with no stored pair is new and
// has nothing to compare.
func changedField(rules []FieldRule, stored, merged any, path string, pick func(FieldRule) bool) string {
	switch s := stored.(type) {
	case []any:
		m, _ := merged.([]any)
		byID := map[string]any{}
		for _, e := range m {
			if id := elemID(e); id != "" {
				byID[id] = e
			}
		}
		for i, se := range s {
			var me any
			if id := elemID(se); id != "" {
				me = byID[id]
			} else if i < len(m) && elemID(m[i]) == "" {
				me = m[i]
			}
			if p := changedField(rules, se, me, fmt.Sprintf("%s[%d]", path, i), pick); p != "" {
				return p
			}
		}
	case map[string]any:
		m, _ := merged.(map[string]any)
		for _, r := range rules {
			if r.ReadOnly {
				continue
			}
			sv, mv, p := s[r.Name], m[r.Name], joinPath(path, r.Name)
			if pick(r) {
				if !jsonEqual(sv, mv) {
					return p
				}
				continue
			}
			if r.Fields == nil || sv == nil {
				continue
			}
			if indirect(r.Type).Kind() == reflect.Map {
				sm, _ := sv.(map[string]any)
				mm, _ := mv.(map[string]any)
				for k, x := range sm {
					if q := changedField(r.Fields, x, mm[k], p+"["+k+"]", pick); q != "" {
						return q
					}
				}
				continue
			}
			if q := changedField(r.Fields, sv, mv, p, pick); q != "" {
				return q
			}
		}
	}
	return ""
}

func elemID(e any) string {
	if m, ok := e.(map[string]any); ok && m["id"] != nil {
		return fmt.Sprint(m["id"])
	}
	return ""
}

// cloneJSON deep-copies a decoded JSON value.
func cloneJSON(v any) any {
	switch x := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(x))
		for k, e := range x {
			m[k] = cloneJSON(e)
		}
		return m
	case []any:
		l := make([]any, len(x))
		for i, e := range x {
			l[i] = cloneJSON(e)
		}
		return l
	}
	return v
}

// allDefaultRouting reports whether v is a non-empty list of objects that all
// carry defaultRouting=true.
func allDefaultRouting(v any) bool {
	list, ok := v.([]any)
	if !ok || len(list) == 0 {
		return false
	}
	for _, e := range list {
		m, ok := e.(map[string]any)
		if !ok || m["defaultRouting"] != true {
			return false
		}
	}
	return true
}

// applied reports whether the confirming read holds the PUT: every patched
// path has the sent value (a null path is absent), every added item is in
// its list and every removed item or element id is gone. Fields the patch
// does not touch are ignored, since the backend changes state and
// re-derives fields.
func (c *Core[T]) applied(current []byte, pm map[string]json.RawMessage, le listEdits) bool {
	cur, err := decodeObject(current)
	if err != nil {
		return false
	}
	idPaths := make(map[string]bool, len(c.d.IDMatchedArrays))
	for _, p := range c.d.IDMatchedArrays {
		idPaths[p] = true
	}
	for k, raw := range pm {
		pv, err := decodeAny(raw)
		if err != nil || !patchHolds(cur, map[string]any{k: pv}, "", idPaths) {
			return false
		}
	}
	for k, items := range le.add {
		list, _ := cur[k].([]any)
		for _, it := range items {
			if !containsJSON(list, it) {
				return false
			}
		}
	}
	for k, items := range le.remove {
		if c.idMatchedTop(k) {
			if _, missing := removeByID(cur[k], items); len(missing) != len(items) {
				return false
			}
			continue
		}
		list, _ := cur[k].([]any)
		for _, it := range items {
			if containsJSON(list, it) {
				return false
			}
		}
	}
	return true
}

// patchHolds reports whether cur carries every value of merge patch p.
// Arrays at idPaths are checked element by element through the ids; a patch
// element without an id (a new element) cannot be found, so it never holds.
func patchHolds(cur, p any, path string, idPaths map[string]bool) bool {
	switch pv := p.(type) {
	case map[string]any:
		cm, ok := cur.(map[string]any)
		if !ok {
			return false
		}
		for k, v := range pv {
			if v == nil {
				if cm[k] != nil {
					return false
				}
				continue
			}
			if !patchHolds(cm[k], v, joinPath(path, k), idPaths) {
				return false
			}
		}
		return true
	case []any:
		if !idPaths[path] {
			break
		}
		cl, _ := cur.([]any)
		byID := map[string]any{}
		for _, e := range cl {
			if m, ok := e.(map[string]any); ok && m["id"] != nil {
				byID[fmt.Sprint(m["id"])] = m
			}
		}
		for _, e := range pv {
			em, ok := e.(map[string]any)
			if !ok || em["id"] == nil {
				return false
			}
			ce, ok := byID[fmt.Sprint(em["id"])]
			if !ok || !patchHolds(ce, em, path+"[]", idPaths) {
				return false
			}
		}
		return true
	}
	return jsonEqual(cur, p)
}

// classify sets the descriptor's classification on an *APIError that the
// transport left unclassified.
func (c *Core[T]) classify(method string, err error) error {
	var ae *APIError
	if !errors.As(err, &ae) || ae.kind != nil {
		return err
	}
	switch {
	case ae.matches(c.d.notFound(method)):
		ae.kind = ErrNotFound
	case method == "DELETE" && ae.matches(c.d.InUse):
		ae.kind = ErrInUse
	case (method == "POST" || method == "PUT") && ae.matches(c.d.Duplicate):
		ae.kind = ErrDuplicate
	}
	return err
}

// lostResponse reports whether a write error leaves the outcome open: a
// deadline while the request was in flight, a post-send failure, any 5xx on
// PUT or DELETE, and on POST the gateway's 500 "Request failed" or a 502,
// 503 or 504 from the gateway or ingress, which can answer after the backend
// committed. Other POST 500s come from the backend and leave nothing
// created. A deadline reached before the request was sent is not lost.
func lostResponse(err error, method string) bool {
	if errors.Is(err, ErrPostSend) || errors.Is(err, errDeadlineInFlight) {
		return true
	}
	var ae *APIError
	if !errors.As(err, &ae) || ae.Status < 500 {
		return false
	}
	if method == "POST" {
		return ae.Status == 500 && ae.Message == "Request failed" || ae.Status >= 502 && ae.Status <= 504
	}
	return true
}

// keyValue returns the uniqueness key of a create body, or "" when the
// guard is off or the key is unset.
func (c *Core[T]) keyValue(m map[string]any) string {
	if c.d.Uniqueness.Scope == ScopeNone {
		return ""
	}
	v, ok := m[c.d.Uniqueness.Key]
	if !ok || v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

// lookup finds an existing object holding the uniqueness key value, through
// paged reads and GET /connectors, never ?name= (D15).
func (c *Core[T]) lookup(ctx context.Context, value string) (*DuplicateError, error) {
	key := c.d.Uniqueness.Key
	pagedMatch := func(uri string) (*DuplicateError, error) {
		rows, err := c.ac.listPaged(ctx, uri)
		if err != nil {
			return nil, err
		}
		for _, r := range rows {
			m, err := decodeObject(r)
			if err != nil {
				return nil, err
			}
			if v, ok := m[key]; ok && v != nil && fmt.Sprint(v) == value {
				return &DuplicateError{Key: key, Value: value, ID: fmt.Sprint(m["id"])}, nil
			}
		}
		return nil, nil
	}
	groups := c.ac.URI + "/tenantnetworks/" + c.ac.TenantNetworkId + "/groups"
	switch c.d.Uniqueness.Scope {
	case ScopeCollection:
		return pagedMatch(c.d.collectionURI(c.ac))
	case ScopeConnectorsGroups:
		if dup, err := c.connectorMatch(ctx, value); dup != nil || err != nil {
			return dup, err
		}
		return pagedMatch(groups)
	case ScopeGroupsConnectors:
		if dup, err := pagedMatch(c.d.collectionURI(c.ac)); dup != nil || err != nil {
			return dup, err
		}
		return c.connectorMatch(ctx, value)
	}
	return nil, nil
}

// connectorMatch checks the connector name space. Deleted rows stay in
// GET /connectors for some seconds, so each same-named row is confirmed by
// a single GET and dropped when that answers the not-found signature.
func (c *Core[T]) connectorMatch(ctx context.Context, name string) (*DuplicateError, error) {
	rows, err := c.ac.ListConnectors(ctx)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		if r.Name != name {
			continue
		}
		uri := c.ac.URI + "/tenantnetworks/" + c.ac.TenantNetworkId + "/connectors/" + pathSegment(r.ID.String())
		if _, err := c.ac.doGet(ctx, uri); err != nil {
			var ae *APIError
			if errors.As(err, &ae) && ae.matches(defaultNotFound) {
				continue
			}
			return nil, err
		}
		return &DuplicateError{Key: c.d.Uniqueness.Key, Value: name, ID: r.ID.String(), Type: r.Type}, nil
	}
	return nil, nil
}

// ConfirmTimeout bounds the confirming GET or uniqueness-key lookup after a
// lost write. It runs on its own budget, detached from the write deadline,
// because the deadline firing in flight is one of the cases it resolves.
const ConfirmTimeout = 5 * time.Second

func confirmContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), ConfirmTimeout)
}

// requestIdOf returns the request id an *APIError carries, or "".
func requestIdOf(err error) string {
	var ae *APIError
	if errors.As(err, &ae) {
		return ae.RequestID
	}
	return ""
}

// idOf returns the top-level id of a JSON object body, or "".
func idOf(body []byte) string {
	m, err := decodeObject(body)
	if err != nil || m["id"] == nil {
		return ""
	}
	return fmt.Sprint(m["id"])
}

// Deployed reports whether a stored object has been provisioned, by the
// backend's own rule: the state is not PENDING and the previous state, when
// present, is not PENDING (TPS Defaults.ProvisionedState.isProvisioned; a
// first provision that is SCHEDULED keeps previousState PENDING). An object
// without state counts as deployed only when d.EmptyStateDeployed is set
// (policy-service marks such rows for deletion). A deployed object is
// marked for deletion on delete, not removed.
func Deployed(d Descriptor, stored []byte) (bool, error) {
	m, err := decodeObject(stored)
	if err != nil {
		return false, err
	}
	state, err := stringField(m, "state")
	if err != nil {
		return false, err
	}
	prev, err := stringField(m, "previousState")
	if err != nil {
		return false, err
	}
	if state == "" {
		return d.EmptyStateDeployed, nil
	}
	return state != "PENDING" && prev != "PENDING", nil
}

func stringField(m map[string]any, k string) (string, error) {
	v, ok := m[k]
	if !ok || v == nil {
		return "", nil
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("%w: %s is not a string", ErrInvalidInput, k)
	}
	return s, nil
}

// maxPages bounds a paged read; 1000 pages of 100 rows exceed every tenant limit.
const maxPages = 1000

// listPaged reads uri with ?paginated=true&limit=100 and offset until
// pagination.hits or an empty page.
func (ac *AlkiraClient) listPaged(ctx context.Context, uri string) ([]json.RawMessage, error) {
	var rows []json.RawMessage
	offset := 0
	for page := 0; page < maxPages; page++ {
		q := url.Values{"paginated": {"true"}, "limit": {"100"}, "offset": {fmt.Sprint(offset)}}
		raw, err := ac.doGet(ctx, withQuery(uri, q))
		if err != nil {
			return nil, err
		}
		var p struct {
			Data       []json.RawMessage `json:"data"`
			Pagination struct {
				Hits int `json:"hits"`
			} `json:"pagination"`
		}
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, fmt.Errorf("GET %s: page at offset %d: %w", uri, offset, err)
		}
		rows = append(rows, p.Data...)
		offset += len(p.Data)
		if len(p.Data) == 0 || offset >= p.Pagination.Hits {
			return rows, nil
		}
	}
	return nil, fmt.Errorf("GET %s: more than %d pages", uri, maxPages)
}

func decodeAny(raw []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}

func decodeObject(raw []byte) (map[string]any, error) {
	v, err := decodeAny(raw)
	if err != nil {
		return nil, fmt.Errorf("body is not JSON: %w", err)
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil, errors.New("body is not a JSON object")
	}
	return m, nil
}

// decodeEntity decodes a write response. It returns nil for an empty body or
// one that does not decode into T; WriteResult.Body keeps the bytes.
func decodeEntity[T any](body []byte) *T {
	var out T
	if len(bytes.TrimSpace(body)) == 0 || json.Unmarshal(body, &out) != nil {
		return nil
	}
	return &out
}

func jsonEqual(a, b any) bool {
	x, err1 := json.Marshal(a)
	y, err2 := json.Marshal(b)
	return err1 == nil && err2 == nil && string(x) == string(y)
}
