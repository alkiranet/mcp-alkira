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
	"slices"
	"strconv"
	"strings"
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
	// A gated field whose zero value is marshalled would count as set on every create.
	zero, err := json.Marshal(new(T))
	if err != nil {
		return nil, err
	}
	zm, err := decodeObject(zero)
	if err != nil {
		return nil, err
	}
	for _, f := range d.Flags {
		if f.Effect != FlagIncludeField {
			continue
		}
		i := slices.IndexFunc(rules, func(r FieldRule) bool { return r.Name == f.Field })
		switch {
		case i < 0:
			return nil, fmt.Errorf("descriptor %s: flag rule field %q is not a field of the type", d.Collection, f.Field)
		case zm[f.Field] != nil:
			return nil, fmt.Errorf("descriptor %s: flag rule field %q must be a pointer or omitempty", d.Collection, f.Field)
		case f.ClearStored && !slices.Contains([]reflect.Kind{reflect.Slice, reflect.Bool}, indirect(rules[i].Type).Kind()):
			return nil, fmt.Errorf("descriptor %s: ClearStored field %q is not a list or bool", d.Collection, f.Field)
		}
	}
	return &Core[T]{ac: ac, d: d, rules: rules}, nil
}

// Descriptor returns the descriptor the core was built with.
func (c *Core[T]) Descriptor() Descriptor { return c.d.clone() }

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

// Kinds of object a DuplicateError names.
const (
	// DuplicateKindConnector: a connector of type DuplicateError.Type.
	DuplicateKindConnector = "connector"
	// DuplicateKindGroup: a group, which shares the connector name space.
	DuplicateKindGroup = "group"
	// DuplicateKindService: a service of type DuplicateError.Type.
	DuplicateKindService = "service"
)

// DuplicateError refuses a create whose uniqueness key is taken.
type DuplicateError struct {
	Key   string
	Value string
	ID    string // id of the existing object
	Type  string // connector or service type, for those name spaces
	Group bool   // the existing object is the group named by Uniqueness.GroupSuffix
	// Service: the existing object is a service of type Type.
	Service bool
	// Kind is the kind of the existing object when it lies outside the
	// family's own collection: DuplicateKindConnector, DuplicateKindGroup or
	// DuplicateKindService. It is empty for an object of the same family.
	Kind string
}

func (e *DuplicateError) Error() string {
	if e.Group {
		return fmt.Sprintf("group %q already exists (id %s)", e.Value, e.ID)
	}
	if e.Kind == DuplicateKindGroup {
		return fmt.Sprintf("the name %q is taken by group id %s; connector and group names share one name space", e.Value, e.ID)
	}
	if e.Service {
		return fmt.Sprintf("a service named %q already exists (%s id %s)", e.Value, e.Type, e.ID)
	}
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
	// FlagsApplied, FieldsDropped and FieldsCleared are those of the lost
	// POST or PUT, as in WriteResult; FlagsApplied names AsyncValidationFlag when the
	// request asked for async validation.
	FlagsApplied  []string
	FieldsDropped []string
	FieldsCleared []string
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
	start := time.Now()
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
	flags, err := c.ac.WriteFeatures(ctx)
	if err != nil {
		return nil, nil, err
	}
	sh, err := c.shape(flags, func(f string) bool { return m[f] != nil }, func(f string) { delete(m, f) })
	if err != nil {
		return nil, nil, err
	}
	if err := c.refuseCleared(m, flags); err != nil {
		return nil, nil, err
	}
	c.applyDefaults(m, flags, &sh)
	if err := c.crossCheck(m); err != nil {
		return nil, nil, err
	}
	body, err := json.Marshal(m)
	if err != nil {
		return nil, nil, fmt.Errorf("create: %w", err)
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
	if err := checkBudget(ctx, start); err != nil {
		return nil, nil, err
	}

	res, err := c.ac.doWrite(ctx, "POST", withQuery(c.d.collectionURI(c.ac), sh.query()), body)
	if err != nil {
		if !lostResponse(err, "POST") {
			return nil, nil, c.classify("POST", err)
		}
		oe := &OutcomeError{Method: "POST", Outcome: OutcomeUnknown, Err: err,
			FlagsApplied: sh.flags(true), FieldsDropped: sh.dropped, FieldsCleared: sh.cleared}
		if key == "" {
			return nil, nil, oe
		}
		cctx, ccancel := confirmContext(ctx)
		defer ccancel()
		dup, lerr := c.confirmLookup(cctx, key)
		switch {
		case lerr != nil:
		case dup != nil:
			oe.Outcome, oe.ID = OutcomeCreated, dup.ID
			res := &WriteResult{ID: dup.ID, RequestID: requestIdOf(err), Saved: true}
			sh.report(res, true)
			if sh.async {
				res.ValidationState = ValidationUnknown
			}
			return nil, res, oe
		default:
			oe.Outcome = OutcomeMayStillBeCreated
		}
		return nil, nil, oe
	}
	res.Saved = true
	sh.report(res, true)
	res.ID = idOf(res.Body)
	if res.ID == "" && key != "" {
		cctx, ccancel := confirmContext(ctx)
		defer ccancel()
		if dup, lerr := c.confirmLookup(cctx, key); lerr == nil && dup != nil {
			res.ID = dup.ID
		}
	}
	if res.ValidationStateID != "" {
		c.ac.awaitValidation(ctx, res)
	}
	return decodeEntity[T](res.Body), res, nil
}

// Get reads one object. It refuses an id of another family
// (ErrFamilyMismatch) and a soft-deleted row (*StateError), and returns the
// raw body next to the decoded object.
func (c *Core[T]) Get(ctx context.Context, id string) (*T, []byte, error) {
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()

	raw, _, err := c.preGet(ctx, id, false)
	if err != nil {
		return nil, nil, err
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, raw, fmt.Errorf("GET id %s: %w", id, err)
	}
	return &out, raw, nil
}

// liveInList confirms that id is a live row of the paged collection, which
// leaves deleted rows out.
func (c *Core[T]) liveInList(ctx context.Context, id string) error {
	rows, err := c.ac.listPaged(ctx, c.d.collectionURI(c.ac), c.d.Paging)
	if err != nil {
		return fmt.Errorf("list %s: %w", c.d.Collection, c.classify("GET", err))
	}
	for _, raw := range rows {
		var row struct {
			ID any `json:"id"`
		}
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.UseNumber()
		if dec.Decode(&row) == nil && fmt.Sprint(row.ID) == id {
			return nil
		}
	}
	return &StateError{ID: id, DocState: "DELETED"}
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
	start := time.Now()
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
	flags, err := c.ac.WriteFeatures(ctx)
	if err != nil {
		return nil, nil, err
	}
	if err := c.refuseClearedPatch(pm, flags); err != nil {
		return nil, nil, err
	}
	// Flag rules see only the fields the patch names, explicit null
	// included since it removes the stored value, and the items the list
	// edits add (a `set` field changes through add only); a flag-gated
	// field held in the stored body is resent unchanged unless its rule
	// clears it.
	sh, err := c.shape(flags,
		func(f string) bool {
			_, ok := pm[f]
			return ok || len(le.add[f]) > 0
		},
		func(f string) { delete(pm, f); delete(le.add, f) })
	if err != nil {
		return nil, nil, err
	}
	pre, err := c.preGetFull(ctx, id, true)
	if err != nil {
		return nil, nil, err
	}
	raw, stored := pre.raw, pre.m
	if err := c.stateGuard(ctx, id, stored, contains(flags, SelectiveProvFlag)); err != nil {
		return nil, nil, err
	}
	merged, absent, err := c.merge(stored, pm, le)
	if err != nil {
		return nil, nil, err
	}
	c.clearFlagged(merged, flags, &sh)
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
	if p := changedField(c.rules, before, merged, "", func(r FieldRule) bool { return r.Immutable }, false); p != "" {
		return nil, nil, fmt.Errorf("%w: %s cannot be changed after create", ErrInvalidInput, p)
	}
	// Secrets held in stored fields the exclusive drop removed go with them.
	secretBase := cloneJSON(before).(map[string]any)
	dropExclusive(c.rules, secretBase, decodePatch(pm))
	if p := changedField(c.rules, secretBase, merged, "", func(r FieldRule) bool { return r.Secret }, true); p != "" {
		return nil, nil, fmt.Errorf("%w: the update would change or drop the stored secret %s", ErrInvalidInput, p)
	}
	if err := c.crossCheck(merged); err != nil {
		return nil, nil, err
	}
	// List edits that change nothing send no PUT: a PUT can move the entity
	// back to PENDING.
	if len(pm) == 0 && c.editsUnchanged(stored, merged, le) {
		res := &WriteResult{Body: raw, Previous: raw, Absent: absent, Unchanged: true}
		sh.report(res, false)
		return decodeEntity[T](raw), res, nil
	}
	body, err := json.Marshal(merged)
	if err != nil {
		return nil, nil, fmt.Errorf("update: %w", err)
	}
	if err := c.unchangedSince(ctx, id, pre.entity); err != nil {
		return nil, nil, err
	}
	if err := checkBudget(ctx, start); err != nil {
		return nil, nil, err
	}

	res, err := c.ac.doWrite(ctx, "PUT", withQuery(c.d.entityURI(c.ac, id), sh.query()), body)
	if err != nil {
		if !lostResponse(err, "PUT") {
			return nil, nil, c.classify("PUT", err)
		}
		oe := &OutcomeError{Method: "PUT", ID: id, Outcome: OutcomeUnknown, Err: err,
			FlagsApplied: sh.flags(true), FieldsDropped: sh.dropped, FieldsCleared: sh.cleared}
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
	res.Saved = true
	sh.report(res, true)
	if res.ValidationStateID != "" {
		c.ac.awaitValidation(ctx, res)
	}
	return decodeEntity[T](res.Body), res, nil
}

// DeleteOptions holds per-call options for Core.DeleteWith.
type DeleteOptions struct {
	// Check, when set, runs on the pre-GET body before the DELETE is sent. A
	// non-nil error is returned as is and no DELETE is sent.
	Check func(stored []byte) error
}

// Delete deletes one object after a GET that refuses another family's id
// and a soft-deleted row. An answer that matches the family's DELETE
// not-found signature counts as deleted, whatever its status.
func (c *Core[T]) Delete(ctx context.Context, id string) (*WriteResult, error) {
	return c.DeleteWith(ctx, id, DeleteOptions{})
}

// DeleteWith is Delete with options.
func (c *Core[T]) DeleteWith(ctx context.Context, id string, opts DeleteOptions) (*WriteResult, error) {
	start := time.Now()
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()

	if id == "" {
		return nil, fmt.Errorf("%w: empty id", ErrInvalidInput)
	}
	flags, err := c.ac.WriteFeatures(ctx)
	if err != nil {
		return nil, err
	}
	// A family behind a FlagTypeAvailable flag cannot be deleted either.
	if _, err := c.shape(flags, func(string) bool { return false }, func(string) {}); err != nil {
		return nil, err
	}
	raw, _, err := c.preGet(ctx, id, true)
	if err != nil {
		return nil, err
	}
	if opts.Check != nil {
		if err := opts.Check(raw); err != nil {
			return nil, err
		}
	}
	if err := checkBudget(ctx, start); err != nil {
		return nil, err
	}
	res, err := c.ac.doWrite(ctx, "DELETE", c.d.entityURI(c.ac, id), nil)
	if err != nil {
		var ae *APIError
		if errors.As(err, &ae) && ae.kind == nil && ae.matches(c.d.notFound("DELETE")) {
			return &WriteResult{Body: []byte(ae.Body), Status: ae.Status, RequestID: ae.RequestID, Previous: raw}, nil
		}
		if !lostResponse(err, "DELETE") {
			return nil, c.classify("DELETE", err)
		}
		oe := &OutcomeError{Method: "DELETE", ID: id, Outcome: OutcomeUnknown, Err: err}
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
	return res, nil
}

// ListPaged reads the whole collection with ?paginated=true&limit=100,
// paging by offset until the end that the descriptor's Paging mode reads.
func (c *Core[T]) ListPaged(ctx context.Context) ([]T, error) {
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()

	rows, err := c.ac.listPaged(ctx, c.d.collectionURI(c.ac), c.d.Paging)
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

	raw, _, err := c.preGet(ctx, id, false)
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

// preRead is the result of a pre-GET: raw is the body the write builds on
// (with the GET /connectors or GET /services row state copied in), m its
// decoded form, and entity the family GET body as received.
type preRead struct {
	raw    []byte
	m      map[string]any
	entity []byte
}

// preGet reads one entity and runs the family and docState checks (D4, BC-R2).
// The FamilyCheck value is checked only before a write.
func (c *Core[T]) preGet(ctx context.Context, id string, write bool) ([]byte, map[string]any, error) {
	p, err := c.preGetFull(ctx, id, write)
	return p.raw, p.m, err
}

// entityQuery is the query of the family GET of a pre-GET.
func (c *Core[T]) entityQuery() url.Values {
	if c.d.MarkedRowLifecycle && c.d.DocState != DocStatePlainGet {
		return url.Values{"includeMarkedForDeletion": {"true"}}
	}
	return nil
}

// unchangedSince re-reads the family GET of id right before a PUT and
// refuses the write, with nothing sent, when the body differs from the one
// the update was built on: another writer changed the object in between,
// and the PUT would overwrite that change with the stored body it was
// merged from.
func (c *Core[T]) unchangedSince(ctx context.Context, id string, entity []byte) error {
	cur, err := c.ac.doGet(ctx, withQuery(c.d.entityURI(c.ac, id), c.entityQuery()))
	if err != nil {
		if cerr := ctx.Err(); cerr != nil {
			return fmt.Errorf("re-read of id %s before the write: %w", id, cerr)
		}
		return fmt.Errorf("%w: re-read of id %s before the write: %w", ErrNotSent, id, c.classify("GET", err))
	}
	before, err1 := decodeObject(entity)
	after, err2 := decodeObject(cur)
	if err1 != nil || err2 != nil || !jsonEqual(dropPerRead(before), dropPerRead(after)) {
		return fmt.Errorf("%w: id %s %w; read it again and retry the update", ErrNotSent, id, ErrChangedSinceRead)
	}
	return nil
}

// perReadFields are fields the backend fills anew on every GET, so they
// differ between two reads of an unchanged object (GET /global-cidr-lists/{id}
// answers a new lastConfigUpdatedAt each time).
var perReadFields = []string{"lastConfigUpdatedAt", "lastModifiedAt"}

// dropPerRead removes perReadFields from v at any depth and returns v.
func dropPerRead(v any) any {
	switch x := v.(type) {
	case map[string]any:
		for _, k := range perReadFields {
			delete(x, k)
		}
		for _, e := range x {
			dropPerRead(e)
		}
	case []any:
		for _, e := range x {
			dropPerRead(e)
		}
	}
	return v
}

func (c *Core[T]) preGetFull(ctx context.Context, id string, write bool) (preRead, error) {
	raw, m, entity, err := c.preGetBody(ctx, id, write)
	return preRead{raw: raw, m: m, entity: entity}, err
}

func (c *Core[T]) preGetBody(ctx context.Context, id string, write bool) ([]byte, map[string]any, []byte, error) {
	if id == "" {
		return nil, nil, nil, fmt.Errorf("%w: empty id", ErrInvalidInput)
	}
	var row *serviceRef
	if c.d.ServiceType != "" {
		r, err := c.serviceRow(ctx, id)
		if err != nil {
			return nil, nil, nil, err
		}
		row = r
	}
	uri := c.d.entityURI(c.ac, id)
	raw, err := c.ac.doGet(ctx, withQuery(uri, c.entityQuery()))
	if err != nil {
		err = c.classify("GET", err)
		if c.d.DocState == DocStatePlainGet && c.d.MarkedRowLifecycle && errors.Is(err, ErrNotFound) {
			marked, merr := c.ac.doGet(ctx, withQuery(uri, url.Values{"includeMarkedForDeletion": {"true"}}))
			if merr == nil {
				return marked, nil, nil, &StateError{ID: id, DocState: "MARKED_FOR_DELETION"}
			}
		}
		return nil, nil, nil, err
	}
	entity := raw
	m, err := decodeObject(raw)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("GET id %s: %w", id, err)
	}
	if f := c.d.FamilyType; f.Field != "" {
		v, ok := m[f.Field]
		if (!ok && (write || !f.ReadMissing)) || (ok && write && f.Value != "" && fmt.Sprint(v) != f.Value) {
			return nil, nil, nil, &FamilyError{ID: id, Collection: c.d.Collection, Field: f.Field, Got: v, Want: f.Value}
		}
	}
	if c.d.ConnectorType != "" {
		if raw, err = c.connectorRow(ctx, id, m); err != nil {
			return nil, nil, nil, err
		}
	}
	if row != nil {
		if raw, err = withRowState(m, id, row.State, row.PreviousState); err != nil {
			return nil, nil, nil, err
		}
	}
	if c.d.DocState != DocStatePlainGet {
		v, ok := m["docState"]
		ds, isString := v.(string)
		switch {
		case !ok && c.d.DocState == DocStateField:
			return nil, nil, nil, fmt.Errorf("%w: GET id %s has no docState", ErrStateRefused, id)
		case ok && !isString:
			return nil, nil, nil, fmt.Errorf("%w: GET id %s has a docState that is not a string", ErrStateRefused, id)
		case ds == "MARKED_FOR_DELETION" || ds == "DELETED":
			return raw, m, entity, &StateError{ID: id, DocState: ds}
		}
	}
	if c.d.LiveInList {
		if err := c.liveInList(ctx, id); err != nil {
			return raw, m, entity, err
		}
	}
	return raw, m, entity, nil
}

// connectorRow checks id against its GET /connectors/{id} row and copies
// the row's state and previousState into m. It returns m encoded again. The
// single GET reads the database; the GET /connectors list answers from a
// cached network view by default (staleOk), which can miss a just-created
// row or carry an old state.
func (c *Core[T]) connectorRow(ctx context.Context, id string, m map[string]any) ([]byte, error) {
	data, err := c.ac.doGet(ctx, c.ac.URI+"/tenantnetworks/"+c.ac.TenantNetworkId+"/connectors/"+pathSegment(id))
	var ae *APIError
	switch {
	case errors.As(err, &ae) && ae.Status == 404:
		return nil, fmt.Errorf("%w: id %s is not a connector", ErrFamilyMismatch, id)
	case err != nil:
		return nil, fmt.Errorf("connector type: %w", c.classify("GET", err))
	}
	var r ConnectorRef
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("GET connector %s: %w", id, err)
	}
	if r.Type != c.d.ConnectorType {
		return nil, fmt.Errorf("%w: id %s is a %s connector, not %s", ErrFamilyMismatch, id, r.Type, c.d.ConnectorType)
	}
	return withRowState(m, id, r.State, r.PreviousState)
}

// withRowState copies a GET /connectors or GET /services row's state and
// previousState into m and returns m encoded again.
func withRowState(m map[string]any, id, state, prev string) ([]byte, error) {
	for k, v := range map[string]string{"state": state, "previousState": prev} {
		if v != "" {
			m[k] = v
		}
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("GET id %s: %w", id, err)
	}
	return raw, nil
}

// serviceRef is one GET /services row.
type serviceRef struct {
	ID            json.Number `json:"id"`
	Name          string      `json:"name"`
	Type          string      `json:"type"`
	State         string      `json:"state"`
	PreviousState string      `json:"previousState"`
}

// serviceRow reads GET /services/{id} with marked rows and refuses an id of
// another service type. The row reads the database, as the family GET does.
func (c *Core[T]) serviceRow(ctx context.Context, id string) (*serviceRef, error) {
	uri := c.ac.URI + "/tenantnetworks/" + c.ac.TenantNetworkId + "/services/" + pathSegment(id)
	data, err := c.ac.doGet(ctx, withQuery(uri, url.Values{"includeMarkedForDeletion": {"true"}}))
	if err != nil {
		return nil, fmt.Errorf("service type: %w", c.classify("GET", err))
	}
	var r serviceRef
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("GET service %s: %w", id, err)
	}
	if r.Type != c.d.ServiceType {
		return nil, &FamilyError{ID: id, Collection: c.d.Collection, Field: "type", Got: r.Type, Want: c.d.ServiceType}
	}
	return &r, nil
}

// stateGuard runs the descriptor's state guard on the pre-GET body. The
// network state is read only when the entity state alone does not decide
// (BC-R10). With selectiveProv false the D13 guard also refuses an
// IN_PROGRESS network, as TPS does with SELECTIVE-PROV off.
func (c *Core[T]) stateGuard(ctx context.Context, id string, stored map[string]any, selectiveProv bool) error {
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
		if nw.State == "FAILED" || nw.State == "IN_PROGRESS" && !selectiveProv {
			return &StateError{ID: id, State: state, NetworkState: nw.State}
		}
	}
	return nil
}

// shaping is what the flag rules did to one create or update.
type shaping struct {
	applied, dropped, cleared []string
	async                     bool
}

func (s shaping) query() url.Values {
	if !s.async {
		return nil
	}
	return url.Values{"async": {"true"}}
}

// flags names the flags that changed the request; async validation changes
// it only when the request is sent.
func (s shaping) flags(sent bool) []string {
	if s.async && sent && !contains(s.applied, AsyncValidationFlag) {
		return append(slices.Clone(s.applied), AsyncValidationFlag)
	}
	return s.applied
}

func (s shaping) report(res *WriteResult, sent bool) {
	res.FlagsApplied, res.FieldsDropped, res.FieldsCleared = s.flags(sent), s.dropped, s.cleared
}

func (s *shaping) mark(flags ...string) {
	for _, f := range flags {
		if !contains(s.applied, f) {
			s.applied = append(s.applied, f)
		}
	}
}

// shape applies the descriptor's flag rules: it refuses a type or a refused
// field whose flag is off (ErrNotEnabled), drops the other gated fields set
// while their flag is off, and selects async validation.
func (c *Core[T]) shape(flags []string, sets func(field string) bool, drop func(field string)) (shaping, error) {
	var s shaping
	refuse := func(what string, r FlagRule) error {
		return fmt.Errorf("%s: %w (needs %s)", what, ErrNotEnabled, r.needs())
	}
	for _, r := range c.d.Flags {
		on := r.on(flags)
		switch {
		case r.Effect == FlagTypeAvailable && !on:
			return s, refuse(c.d.Collection, r)
		case r.Effect != FlagIncludeField || !sets(r.Field):
		case on:
			s.mark(r.deciding(flags)...)
		case r.Refuse:
			return s, refuse(r.Field, r)
		default:
			drop(r.Field)
			s.dropped = append(s.dropped, r.Field)
			s.mark(r.deciding(flags)...)
		}
	}
	s.async = c.d.AsyncCapable && contains(flags, AsyncValidationFlag)
	return s, nil
}

// atPath returns the object holding the last element of a dotted JSON path
// in m, and that element's key; parent is nil when the path stops early.
func atPath(m map[string]any, path string) (parent map[string]any, key string) {
	parts := strings.Split(path, ".")
	parent = m
	for _, p := range parts[:len(parts)-1] {
		parent, _ = parent[p].(map[string]any)
	}
	return parent, parts[len(parts)-1]
}

// refuseCleared refuses a request body m that sets a FlagClearField list
// non-empty while the rule is on: the backend refuses the list, and sending
// it empty would widen what the request asks for.
func (c *Core[T]) refuseCleared(m map[string]any, flags []string) error {
	for _, r := range c.d.Flags {
		if r.Effect != FlagClearField || !r.on(flags) {
			continue
		}
		parent, key := atPath(m, r.Field)
		if l, ok := parent[key].([]any); ok && len(l) > 0 {
			return fmt.Errorf("%w: %s must be empty while the tenant feature %s is on", ErrInvalidInput, r.Field, r.needs())
		}
	}
	return nil
}

// refuseClearedPatch is refuseCleared for a merge patch.
func (c *Core[T]) refuseClearedPatch(pm map[string]json.RawMessage, flags []string) error {
	m := make(map[string]any, len(pm))
	for k, raw := range pm {
		if v, err := decodeAny(raw); err == nil {
			m[k] = v
		}
	}
	return c.refuseCleared(m, flags)
}

// clearFlagged runs on a merged update body, whose gated values the request
// cannot set (refuseCleared and shape keep them out): it sends empty every
// FlagClearField list whose rule is on and every ClearStored field whose
// rule is off that the stored body carries non-empty, and records the
// rule's flags in sh.applied and the field in sh.cleared.
func (c *Core[T]) clearFlagged(m map[string]any, flags []string, sh *shaping) {
	for _, r := range c.d.Flags {
		on := r.on(flags)
		if !(r.Effect == FlagClearField && on || r.ClearStored && !on) {
			continue
		}
		parent, key := atPath(m, r.Field)
		switch v := parent[key].(type) {
		case []any:
			if len(v) == 0 {
				continue
			}
			parent[key] = []any{}
		case bool:
			if !v {
				continue
			}
			parent[key] = false
		default:
			continue
		}
		sh.mark(r.deciding(flags)...)
		sh.cleared = append(sh.cleared, r.Field)
	}
}

// applyDefaults fills the unset fields of a create body that carry default=,
// nested struct fields included, so a create sends what the portal sends. A
// field that a FlagIncludeField rule gates stays unset while its flag is off;
// while the flag is on, a default sent into it adds the flag to sh.applied.
func (c *Core[T]) applyDefaults(m map[string]any, flags []string, sh *shaping) {
	gatedOff := map[string]bool{}
	var gatedOn []FlagRule
	for _, f := range c.d.Flags {
		switch {
		case f.Effect != FlagIncludeField:
		case !f.on(flags):
			gatedOff[f.Field] = true
		case m[f.Field] == nil:
			gatedOn = append(gatedOn, f)
		}
	}
	fillDefaults(c.rules, m, gatedOff)
	for _, f := range gatedOn {
		if m[f.Field] != nil {
			sh.mark(f.deciding(flags)...)
		}
	}
}

func fillDefaults(rules []FieldRule, m map[string]any, skip map[string]bool) {
	for _, r := range rules {
		t := indirect(r.Type)
		switch {
		case r.ReadOnly || skip[r.Name]:
		case t.Kind() == reflect.Slice && r.Fields != nil:
			// Defaults apply inside each element of a list of objects.
			list, _ := m[r.Name].([]any)
			for _, e := range list {
				if child, ok := e.(map[string]any); ok {
					fillDefaults(r.Fields, child, nil)
				}
			}
		case t.Kind() == reflect.Slice || t.Kind() == reflect.Map:
		case r.HasDefault:
			if m[r.Name] == nil {
				m[r.Name] = defaultValue(r.Default, t.Kind())
			}
		case r.Fields != nil && t.Kind() == reflect.Struct:
			child, ok := m[r.Name].(map[string]any)
			if !ok {
				if m[r.Name] != nil {
					continue
				}
				child = map[string]any{}
			}
			fillDefaults(r.Fields, child, nil)
			if len(child) > 0 {
				m[r.Name] = child
			}
		}
	}
}

// defaultValue is a default= value as JSON of kind k; parseTag checked it.
func defaultValue(v string, k reflect.Kind) any {
	switch {
	case k == reflect.Bool:
		b, _ := strconv.ParseBool(v)
		return b
	case isInt(k), k == reflect.Float32, k == reflect.Float64:
		return json.Number(v)
	}
	return v
}

// merge builds the PUT body from the stored body, the patch and the list
// edits. It returns the removed items the stored lists did not hold.
func (c *Core[T]) merge(stored map[string]any, pm map[string]json.RawMessage, le listEdits) (map[string]any, map[string][]any, error) {
	base := cloneJSON(stored).(map[string]any)
	// defaultRouting can be a read-only field, so provenance is read before the strip.
	for _, f := range c.d.RoutingProvenance {
		if _, set := pm[f]; !set && allDefaultRouting(base[f]) {
			delete(base, f)
		}
	}
	if c.d.PrepareStored != nil {
		c.d.PrepareStored(base)
	}
	stripStored(c.rules, base)
	for _, r := range c.rules {
		if r.expired() {
			delete(base, r.Name)
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
	mv, err := mergeValue(base, patch, "", c.matchPaths())
	if err != nil {
		return nil, nil, err
	}
	out := mv.(map[string]any)
	dropExclusive(c.rules, out, patch)
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

// arrayMatch is how an array merges: by key, and with strict set by id with
// every stored id required (IDMatchedArrays).
type arrayMatch struct {
	key    string
	strict bool
}

// matchPaths returns the descriptor's id-matched and key-matched arrays by path.
func (c *Core[T]) matchPaths() map[string]arrayMatch {
	paths := make(map[string]arrayMatch, len(c.d.IDMatchedArrays)+len(c.d.KeyMatchedArrays))
	for _, p := range c.d.IDMatchedArrays {
		paths[p] = arrayMatch{key: "id", strict: true}
	}
	for p, k := range c.d.KeyMatchedArrays {
		paths[p] = arrayMatch{key: k}
	}
	return paths
}

// mergeValue applies RFC 7386 to target, except that id-matched arrays merge
// element by id: the patch must carry the id of every stored element, an
// element with a known id merges into it, one without id is new, and an
// unknown id is refused. Key-matched arrays merge element by key (mergeByKey).
func mergeValue(target, patch any, path string, paths map[string]arrayMatch) (any, error) {
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
			mv, err := mergeValue(t[k], v, joinPath(path, k), paths)
			if err != nil {
				return nil, err
			}
			t[k] = mv
		}
		return t, nil
	case []any:
		switch m, ok := paths[path]; {
		case ok && m.strict:
			return mergeByID(target, p, path, paths)
		case ok:
			return mergeByKey(target, p, path, m.key, paths)
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

func mergeByID(target any, patch []any, path string, paths map[string]arrayMatch) (any, error) {
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
		mv, err := mergeValue(base, pm, path+"[]", paths)
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

// mergeByKey merges each patch element into the stored element with the
// same key value; an element without a stored pair is new, and a stored
// element the patch leaves out is dropped.
func mergeByKey(target any, patch []any, path, key string, paths map[string]arrayMatch) (any, error) {
	stored, _ := target.([]any)
	byKey := map[string]any{}
	for _, e := range stored {
		if m, ok := e.(map[string]any); ok && m[key] != nil {
			byKey[fmt.Sprint(m[key])] = m
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
		if v, has := pm[key]; has && v != nil {
			k := fmt.Sprint(v)
			if sent[k] {
				return nil, fmt.Errorf("%w: %s[%d] repeats %s %s", ErrInvalidInput, path, i, key, k)
			}
			sent[k] = true
			base = byKey[k]
		}
		mv, err := mergeValue(base, pm, path+"[]", paths)
		if err != nil {
			return nil, err
		}
		out = append(out, mv)
	}
	return out, nil
}

// changedField returns the path of the first field that pick selects whose
// merged value differs from the stored one, or "". Read-only and expired
// fields are skipped, since the merge strips them. Array elements pair by
// id, or by position when the stored element has no id; a stored element
// with no pair compares against nothing. A merged element with no stored pair is new and
// has nothing to compare. With dropOK set, a droppable field merged no
// longer holds is not searched.
func changedField(rules []FieldRule, stored, merged any, path string, pick func(FieldRule) bool, dropOK bool) string {
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
			if p := changedField(rules, se, me, fmt.Sprintf("%s[%d]", path, i), pick, dropOK); p != "" {
				return p
			}
		}
	case map[string]any:
		m, _ := merged.(map[string]any)
		for _, r := range rules {
			if r.ReadOnly || r.expired() {
				continue
			}
			sv, mv, p := s[r.Name], m[r.Name], joinPath(path, r.Name)
			if pick(r) {
				if !jsonEqual(sv, mv) {
					return p
				}
				continue
			}
			if r.Fields == nil || sv == nil || dropOK && r.Droppable && mv == nil {
				continue
			}
			if indirect(r.Type).Kind() == reflect.Map {
				sm, _ := sv.(map[string]any)
				mm, _ := mv.(map[string]any)
				for k, x := range sm {
					if q := changedField(r.Fields, x, mm[k], p+"["+k+"]", pick, dropOK); q != "" {
						return q
					}
				}
				continue
			}
			if q := changedField(r.Fields, sv, mv, p, pick, dropOK); q != "" {
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

// allDefaultRouting reports whether v holds at least one routing entry and
// every entry has defaultRouting=true. An entry is an object in a list,
// found at any depth: v is a list of entries or an object whose lists hold
// them (AWS VPC vpcRouting.exportToCXPOptions.userInputPrefixes and
// importFromCXPOptions.routeTables).
func allDefaultRouting(v any) bool {
	n := 0
	var walk func(v any, inList bool) bool
	walk = func(v any, inList bool) bool {
		switch t := v.(type) {
		case []any:
			for _, e := range t {
				if !walk(e, true) {
					return false
				}
			}
		case map[string]any:
			if inList {
				n++
				return t["defaultRouting"] == true
			}
			for _, e := range t {
				if !walk(e, false) {
					return false
				}
			}
		}
		return true
	}
	return walk(v, false) && n > 0
}

// decodePatch decodes the top-level values of a validated patch; a value
// that does not decode is left out, since merge refuses it.
func decodePatch(pm map[string]json.RawMessage) map[string]any {
	patch := make(map[string]any, len(pm))
	for k, raw := range pm {
		if v, err := decodeAny(raw); err == nil {
			patch[k] = v
		}
	}
	return patch
}

// dropExclusive removes from merged the stored fields of an exclusive group
// whose other side the patch sets, at the top level and in nested objects
// the patch sets.
func dropExclusive(rules []FieldRule, merged, patch map[string]any) {
	sides := map[string]string{}
	for _, r := range rules {
		if v, set := patch[r.Name]; r.Exclusive != "" && set && v != nil {
			g, side := r.ExclusiveGroup()
			sides[g] = side
		}
	}
	for _, r := range rules {
		if r.Exclusive == "" {
			continue
		}
		g, side := r.ExclusiveGroup()
		if set, ok := sides[g]; ok && set != side {
			if _, inPatch := patch[r.Name]; !inPatch {
				delete(merged, r.Name)
			}
		}
	}
	for _, r := range rules {
		pm, ok1 := patch[r.Name].(map[string]any)
		mm, ok2 := merged[r.Name].(map[string]any)
		if r.Fields != nil && ok1 && ok2 {
			dropExclusive(r.Fields, mm, pm)
		}
	}
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
	paths := c.matchPaths()
	for k, raw := range pm {
		pv, err := decodeAny(raw)
		if err != nil || !patchHolds(cur, map[string]any{k: pv}, "", paths) {
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
// Id-matched and key-matched arrays are checked element by element through
// the id or key; a patch element without one (a new element) cannot be
// found, so it never holds.
func patchHolds(cur, p any, path string, paths map[string]arrayMatch) bool {
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
			if !patchHolds(cm[k], v, joinPath(path, k), paths) {
				return false
			}
		}
		return true
	case []any:
		m, matched := paths[path]
		if !matched {
			break
		}
		cl, _ := cur.([]any)
		byKey := map[string]any{}
		for _, e := range cl {
			if cm, ok := e.(map[string]any); ok && cm[m.key] != nil {
				byKey[fmt.Sprint(cm[m.key])] = cm
			}
		}
		for _, e := range pv {
			em, ok := e.(map[string]any)
			if !ok || em[m.key] == nil {
				return false
			}
			ce, ok := byKey[fmt.Sprint(em[m.key])]
			if !ok || !patchHolds(ce, em, path+"[]", paths) {
				return false
			}
		}
		return true
	}
	return jsonEqual(cur, p)
}

// confirmLookup finds the object a POST created by its uniqueness key. A
// group found through Uniqueness.GroupSuffix is the created object's
// implicit group, not the object, so it does not count.
func (c *Core[T]) confirmLookup(ctx context.Context, value string) (*DuplicateError, error) {
	dup, err := c.lookup(ctx, value)
	if dup != nil && dup.Group {
		return nil, err
	}
	return dup, err
}

// crossCheck runs the descriptor's CrossCheck on a request body. An error
// that is not already ErrInvalidInput is wrapped as one.
func (c *Core[T]) crossCheck(body map[string]any) error {
	if c.d.CrossCheck == nil {
		return nil
	}
	err := c.d.CrossCheck(body)
	if err == nil || errors.Is(err, ErrInvalidInput) {
		return err
	}
	return fmt.Errorf("%w: %w", ErrInvalidInput, err)
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

// lookup finds an existing object holding the uniqueness key value: family
// collections and groups through paged reads, the connector name space
// through GET /connectors?name=, services through GET /services.
func (c *Core[T]) lookup(ctx context.Context, value string) (*DuplicateError, error) {
	key := c.d.Uniqueness.Key
	pagedMatch := func(uri string, paging Paging) (*DuplicateError, error) {
		rows, err := c.ac.listPaged(ctx, uri, paging)
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
		dup, err := pagedMatch(c.d.collectionURI(c.ac), c.d.Paging)
		if dup != nil || err != nil || c.d.Uniqueness.GroupSuffix == "" {
			return dup, err
		}
		key, value = "name", value+c.d.Uniqueness.GroupSuffix
		dup, err = pagedMatch(groups, PagingTotalHits)
		if dup != nil {
			dup.Group, dup.Kind = true, DuplicateKindGroup
		}
		return dup, err
	case ScopeConnectorsGroups:
		if dup, err := c.connectorMatch(ctx, value); dup != nil || err != nil {
			return dup, err
		}
		dup, err := pagedMatch(groups, PagingTotalHits)
		if dup != nil {
			dup.Kind = DuplicateKindGroup
		}
		return dup, err
	case ScopeGroupsConnectors:
		if dup, err := pagedMatch(c.d.collectionURI(c.ac), c.d.Paging); dup != nil || err != nil {
			return dup, err
		}
		return c.connectorMatch(ctx, value)
	case ScopeServices:
		// TPS checks the name against every service row, marked ones included
		// (ServiceLogicImpl.java:2174-2184).
		uri := c.ac.URI + "/tenantnetworks/" + c.ac.TenantNetworkId + "/services"
		// staleOk=false: the default cached view can miss a service created seconds ago.
		raw, err := c.ac.doGet(ctx, withQuery(uri, url.Values{"includeMarkedForDeletion": {"true"}, "staleOk": {"false"}}))
		if err != nil {
			return nil, err
		}
		var rows []serviceRef
		if err := json.Unmarshal(raw, &rows); err != nil {
			return nil, fmt.Errorf("GET services: %w", err)
		}
		for _, r := range rows {
			if r.Name == value {
				return &DuplicateError{Key: key, Value: value, ID: r.ID.String(), Type: r.Type, Service: true, Kind: DuplicateKindService}, nil
			}
		}
	}
	return nil, nil
}

// connectorMatch checks the connector name space with
// GET /connectors?name=, which TPS answers from the database with the live
// row of that name, of any connector type (ConnectorController.java:87-91;
// ConnectorServiceImpl.getByName), so the read stays one row however many
// connectors the tenant has. A 5xx (TPS findOne fails when the name has
// more than one row) falls back to connectorListMatch.
func (c *Core[T]) connectorMatch(ctx context.Context, name string) (*DuplicateError, error) {
	uri := c.ac.URI + "/tenantnetworks/" + c.ac.TenantNetworkId + "/connectors"
	// No 5xx retries: the read client retries a 5xx up to 5 times with waits
	// of up to 30 s, which would use up the write deadline before the
	// fallback runs.
	raw, err := c.ac.doGetNo5xxRetry(ctx, withQuery(uri, url.Values{"name": {name}}))
	var ae *APIError
	switch {
	case errors.As(err, &ae) && ae.Status >= 500:
		return c.connectorListMatch(ctx, name)
	case err != nil:
		return nil, err
	}
	var rows []ConnectorRef
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("GET connectors by name: %w", err)
	}
	for _, r := range rows {
		if r.Name == name {
			return &DuplicateError{Key: c.d.Uniqueness.Key, Value: name, ID: r.ID.String(), Type: r.Type, Kind: DuplicateKindConnector}, nil
		}
	}
	return nil, nil
}

// connectorListMatch checks the connector name space through the whole
// GET /connectors list. Deleted rows stay in it for some seconds, so each
// same-named row is confirmed by a single GET and dropped when that answers
// the not-found signature.
func (c *Core[T]) connectorListMatch(ctx context.Context, name string) (*DuplicateError, error) {
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
		return &DuplicateError{Key: c.d.Uniqueness.Key, Value: name, ID: r.ID.String(), Type: r.Type, Kind: DuplicateKindConnector}, nil
	}
	return nil, nil
}

// MinWriteBudget is the time a POST, PUT or DELETE must have left before
// its deadline to be sent, capped at half of the call's whole budget so a
// short configured deadline still writes. A write with less left is refused
// with ErrNotSent and ErrBudgetTooSmall: sent that late, it would end with
// its outcome unknown instead of applied or refused.
var MinWriteBudget = 5 * time.Second

// ErrBudgetTooSmall refuses a write whose remaining deadline is below
// MinWriteBudget after the reads that precede it; nothing was sent.
var ErrBudgetTooSmall = errors.New("too little time left before the deadline to send the write")

// ErrChangedSinceRead refuses an update whose object changed between the
// read the update was built on and the re-read right before the PUT;
// nothing was sent. The returned error also matches ErrNotSent.
var ErrChangedSinceRead = errors.New("changed since it was read")

// checkBudget refuses the write when less than the minimum budget is left
// on ctx; start is when the call began.
func checkBudget(ctx context.Context, start time.Time) error {
	deadline, ok := ctx.Deadline()
	if !ok {
		return nil
	}
	need := MinWriteBudget
	if half := deadline.Sub(start) / 2; half < need {
		need = half
	}
	if left := time.Until(deadline); left < need {
		return fmt.Errorf("%w: %w (%s left, %s needed)", ErrNotSent, ErrBudgetTooSmall,
			max(left, 0).Round(time.Millisecond), need.Round(time.Millisecond))
	}
	return nil
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

// pageLimit is the page size of a paged read; services cap limit at 100.
const pageLimit = 100

// listPaged reads uri with ?paginated=true&limit=100 and offset until an
// empty page, or until pagination.hits (PagingTotalHits) or a short page
// (PagingPageRows).
func (ac *AlkiraClient) listPaged(ctx context.Context, uri string, paging Paging) ([]json.RawMessage, error) {
	return ac.listPagedWith(ctx, uri, paging, nil)
}

// listPagedWith is listPaged with filter added to every page query.
func (ac *AlkiraClient) listPagedWith(ctx context.Context, uri string, paging Paging, filter url.Values) ([]json.RawMessage, error) {
	var rows []json.RawMessage
	offset := 0
	for page := 0; page < maxPages; page++ {
		data, hits, err := ac.listPage(ctx, uri, offset, pageLimit, filter)
		if err != nil {
			return nil, err
		}
		rows = append(rows, data...)
		offset += len(data)
		end := offset >= hits
		if paging == PagingPageRows {
			end = len(data) < pageLimit
		}
		if len(data) == 0 || end {
			return rows, nil
		}
	}
	return nil, fmt.Errorf("GET %s: more than %d pages", uri, maxPages)
}

// listPage reads the page of at most limit rows at offset, with filter added
// to the query; hits is pagination.hits.
func (ac *AlkiraClient) listPage(ctx context.Context, uri string, offset, limit int, filter url.Values) ([]json.RawMessage, int, error) {
	q := url.Values{"paginated": {"true"}, "limit": {fmt.Sprint(limit)}, "offset": {fmt.Sprint(offset)}}
	for k, v := range filter {
		q[k] = v
	}
	raw, err := ac.doGet(ctx, withQuery(uri, q))
	if err != nil {
		return nil, 0, err
	}
	var p struct {
		Data       []json.RawMessage `json:"data"`
		Pagination struct {
			Hits int `json:"hits"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, 0, fmt.Errorf("GET %s: page at offset %d: %w", uri, offset, err)
	}
	return p.Data, p.Pagination.Hits, nil
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
