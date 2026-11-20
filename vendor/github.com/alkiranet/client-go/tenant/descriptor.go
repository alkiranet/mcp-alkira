// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
)

// Signature identifies one backend error response by status plus the
// ErrorDetails code. Code matches the top-level code and FieldCode matches the
// code of any errors[] entry; at least one of them is set. Message, when set,
// must occur in ErrorDetails.message; it separates answers that share a status
// and code, such as a TPS not-found and a validation 400 (both 400 AK000002).
// Bodies that are not ErrorDetails JSON (empty, HTML, a gateway
// {"message":...} without code) never match.
type Signature struct {
	Status    int
	Code      string
	FieldCode string
	Message   string
}

// DocStateMode says how a family reports a soft-deleted row.
type DocStateMode int

const (
	// DocStateField: the body carries docState; a missing value refuses the write.
	DocStateField DocStateMode = iota
	// DocStateMissingIsAdded: docState is omitted while the row is ADDED.
	DocStateMissingIsAdded
	// DocStatePlainGet: the body never carries docState. A plain GET that
	// answers not-found followed by an includeMarkedForDeletion GET that
	// answers 200 means the row is marked for deletion.
	DocStatePlainGet
)

// UniquenessScope is the name space the duplicate guard reads before a create.
type UniquenessScope int

const (
	// ScopeNone disables the duplicate guard.
	ScopeNone UniquenessScope = iota
	// ScopeCollection reads the family's own paged collection.
	ScopeCollection
	// ScopeConnectorsGroups reads GET /connectors (every connector family) and the paged groups.
	ScopeConnectorsGroups
	// ScopeGroupsConnectors reads the family's own paged collection and GET /connectors.
	ScopeGroupsConnectors
)

// Uniqueness is the field a create must not duplicate and where to look for it.
type Uniqueness struct {
	Key   string // JSON field name, usually "name"
	Scope UniquenessScope
	// GroupSuffix, when set, also refuses a create while a group named the key
	// value plus GroupSuffix exists: the backend creates that implicit group
	// with the object and refuses the create when the name is taken.
	GroupSuffix string
}

// StateGuard selects the entity-state check an update runs before its PUT.
type StateGuard int

const (
	// StateGuardNone sends the PUT whatever the entity state.
	StateGuardNone StateGuard = iota
	// StateGuardFailedOnly refuses an update of a FAILED entity.
	StateGuardFailedOnly
	// StateGuardD13 refuses an update of a FAILED, SCHEDULED or IN_PROGRESS
	// entity, and of a non-PENDING entity on a FAILED network. With
	// SELECTIVE-PROV off in the write's flag list (the cached list when the
	// flag read fails) it also refuses a non-PENDING entity while the network
	// is IN_PROGRESS, as TPS does.
	StateGuardD13
)

// Paging says what pagination.hits of a paged collection read counts.
type Paging int

const (
	// PagingTotalHits: hits is the collection total; reads stop at offset >= hits.
	PagingTotalHits Paging = iota
	// PagingPageRows: hits is the rows on the returned page (tenant-billing
	// /tags, TagServiceImpl.java:70-84, :183-196); reads stop on a page
	// shorter than the limit.
	PagingPageRows
)

// Service is the backend that owns a family.
type Service int

const (
	ServiceTPS Service = iota
	ServicePolicy
	ServiceTenantBilling
)

// FamilyCheck identifies a family from its GET body: Field must be present
// and, before a PUT or DELETE, equal Value when Value is set. A sibling
// family's id answers 200 on another family's path, so every read checks
// Field; a path can list rows the backend does not let a write change, so
// writes also check Value while reads return those rows. A body that fails
// the check is refused with a *FamilyError.
type FamilyCheck struct {
	Field string
	Value string
}

// FamilyError refuses an id whose GET body fails the descriptor's FamilyCheck.
type FamilyError struct {
	ID         string
	Collection string
	Field      string
	// Got is the stored value of Field; nil when the body lacks Field.
	Got any
	// Want is FamilyCheck.Value; empty when only the presence of Field is checked.
	Want string
}

func (e *FamilyError) Error() string {
	if e.Got == nil || e.Want == "" {
		return fmt.Sprintf("%v: id %s is not a %s object", ErrFamilyMismatch, e.ID, e.Collection)
	}
	return fmt.Sprintf("%v: id %s in %s has %s %v, not %s", ErrFamilyMismatch, e.ID, e.Collection, e.Field, e.Got, e.Want)
}

// Is reports a match for ErrFamilyMismatch.
func (e *FamilyError) Is(target error) bool { return target == ErrFamilyMismatch }

// FlagEffect is what a feature flag changes in a create or update.
type FlagEffect int

const (
	// FlagIncludeField: Field is sent only while the flag is on. With the flag
	// off a value the request sets is dropped, or refused when Refuse is set,
	// and with ClearStored set an update sends the stored value empty.
	FlagIncludeField FlagEffect = iota + 1
	// FlagTypeAvailable: the family can be written only while the flag is on.
	FlagTypeAvailable
	// FlagClearField: while the flag is on, the list at Field, a dotted JSON
	// path that may name a nested field, must be empty. A create or patch
	// that sets it non-empty is refused (ErrInvalidInput naming the flag), and
	// an update sends a stored non-empty list as [], as the portal does; the
	// flag is then reported in FlagsApplied and Field in FieldsCleared.
	FlagClearField
)

// FlagRule is one way a feature flag shapes a family's writes, as the portal
// does. A rule is on while Flag is on and, when AnyOf is set, at least one
// of AnyOf is on too. Rules combine with AND: a field that two
// FlagIncludeField rules gate is sent only while both are on, and a family
// with two FlagTypeAvailable rules needs both flags.
type FlagRule struct {
	Flag string
	// AnyOf, when set, also needs one of these flags on.
	AnyOf  []string
	Effect FlagEffect
	// Field is the top-level JSON field of a FlagIncludeField rule. Its zero
	// value must be left out of the JSON body (a pointer or omitempty). On a
	// `set` field the rule gates the items an update adds.
	Field string
	// Refuse refuses, rather than drops, a FlagIncludeField value set while the flag is off.
	Refuse bool
	// ClearStored, on a FlagIncludeField rule over a list or bool field, makes
	// an update send the stored value as [] or false while the rule is off,
	// for a backend that refuses a stored value once the flag is off. The
	// rule's flags are then reported in FlagsApplied and Field in FieldsCleared.
	ClearStored bool
	// UISource is the portal file:line that branches on the flag.
	UISource string
}

// on reports whether the rule's flags are on in flags.
func (r FlagRule) on(flags []string) bool {
	return contains(flags, r.Flag) && (len(r.AnyOf) == 0 || slices.ContainsFunc(r.AnyOf, func(f string) bool { return contains(flags, f) }))
}

// needs names the flags the rule needs, for a refusal.
func (r FlagRule) needs() string {
	if len(r.AnyOf) == 0 {
		return r.Flag
	}
	return r.Flag + " and one of " + strings.Join(r.AnyOf, ", ")
}

// deciding names the flags that put the rule in its state for flags: when
// on, Flag and the AnyOf flags that are on; when off, Flag if it is off,
// else every AnyOf flag.
func (r FlagRule) deciding(flags []string) []string {
	if !contains(flags, r.Flag) {
		return []string{r.Flag}
	}
	if !r.on(flags) {
		return slices.Clone(r.AnyOf)
	}
	out := []string{r.Flag}
	for _, f := range r.AnyOf {
		if contains(flags, f) {
			out = append(out, f)
		}
	}
	return out
}

// Descriptor describes one writable family for Core.
type Descriptor struct {
	// Collection is the path below /tenantnetworks/{id}/, or below the API
	// root when TenantLevel is set, for example "segments" or "policy/prefixlists".
	Collection  string
	TenantLevel bool
	// V1Path serves the collection under /v1/tenantnetworks/{id}/.
	V1Path     bool
	FamilyType FamilyCheck
	// ConnectorType is the family's type in GET /connectors. When set, every
	// pre-GET also reads GET /connectors/{id}: an id with no row or a row of
	// another type is refused (ErrFamilyMismatch), and the row's state and
	// previousState, which the family's own GET body does not carry, are
	// added to the stored body for the state guard and the deployed check.
	ConnectorType string

	// NotFound classifies GET and PUT; DeleteNotFound classifies DELETE.
	// nil means 404 with code AK000003.
	NotFound       []Signature
	DeleteNotFound []Signature
	InUse          []Signature
	Duplicate      []Signature

	Provisionable bool
	// MarkedRowLifecycle: rows soft-delete to MARKED_FOR_DELETION and a GET by
	// id needs includeMarkedForDeletion=true to see them.
	MarkedRowLifecycle bool
	DocState           DocStateMode
	// EmptyStateDeployed counts an object without state as deployed
	// (policy-service PolicyServiceImpl.java:501, RuleServiceImpl.java:458).
	EmptyStateDeployed bool
	Uniqueness         Uniqueness
	// IDMatchedArrays lists arrays whose elements the backend matches by id,
	// as dotted JSON paths; "[]" steps into each element ("gateways[].tunnels").
	IDMatchedArrays []string
	StateGuard      StateGuard
	// RoutingProvenance lists top-level fields left out of the PUT when every
	// stored entry has defaultRouting=true and the patch does not set them.
	RoutingProvenance []string
	Service           Service
	Paging            Paging
	// Flags lists the feature flags that shape creates and updates.
	Flags []FlagRule
	// CrossCheck, when set, checks a rule that spans fields and that tags
	// cannot express. It runs on the create body and on the merged PUT body
	// before the request is sent; an error refuses the write as
	// ErrInvalidInput.
	CrossCheck func(body map[string]any) error
	// AsyncCapable: with AsyncValidationFlag on, creates and updates send
	// ?async=true and wait for the validation state.
	AsyncCapable bool
}

// clone returns d with its own copies of the slices, so editing a returned
// descriptor never changes the package one.
func (d Descriptor) clone() Descriptor {
	d.NotFound = slices.Clone(d.NotFound)
	d.DeleteNotFound = slices.Clone(d.DeleteNotFound)
	d.InUse = slices.Clone(d.InUse)
	d.Duplicate = slices.Clone(d.Duplicate)
	d.IDMatchedArrays = slices.Clone(d.IDMatchedArrays)
	d.RoutingProvenance = slices.Clone(d.RoutingProvenance)
	d.Flags = slices.Clone(d.Flags)
	for i := range d.Flags {
		d.Flags[i].AnyOf = slices.Clone(d.Flags[i].AnyOf)
	}
	return d
}

var defaultNotFound = []Signature{{Status: 404, Code: "AK000003"}}

func (d *Descriptor) validate() error {
	if d.Collection == "" || strings.HasPrefix(d.Collection, "/") || strings.HasSuffix(d.Collection, "/") {
		return fmt.Errorf("descriptor: invalid collection %q", d.Collection)
	}
	if d.TenantLevel && d.V1Path {
		return fmt.Errorf("descriptor %s: TenantLevel and V1Path are exclusive", d.Collection)
	}
	if d.Uniqueness.Scope != ScopeNone && d.Uniqueness.Key == "" {
		return fmt.Errorf("descriptor %s: uniqueness scope without key", d.Collection)
	}
	// mergeByID handles a nested path only inside a parent it already merges
	// by id; without the parent, the patch array replaces the stored one and
	// drops every field the patch leaves out, stored keys included.
	listed := map[string]bool{}
	for _, p := range d.IDMatchedArrays {
		listed[p] = true
	}
	for _, p := range d.IDMatchedArrays {
		if i := strings.LastIndex(p, "[]."); i >= 0 && !listed[p[:i]] {
			return fmt.Errorf("descriptor %s: id-matched path %q needs its parent %q listed", d.Collection, p, p[:i])
		}
	}
	for _, f := range d.Flags {
		switch {
		case f.Flag == "" || f.Effect < FlagIncludeField || f.Effect > FlagClearField:
			return fmt.Errorf("descriptor %s: flag rule %+v needs a flag and an effect", d.Collection, f)
		case (f.Effect == FlagIncludeField || f.Effect == FlagClearField) != (f.Field != ""):
			return fmt.Errorf("descriptor %s: flag rule %+v: a field goes with FlagIncludeField and FlagClearField only", d.Collection, f)
		case f.ClearStored && f.Effect != FlagIncludeField:
			return fmt.Errorf("descriptor %s: flag rule %+v: ClearStored goes with FlagIncludeField only", d.Collection, f)
		case slices.Contains(f.AnyOf, "") || slices.Contains(f.AnyOf, f.Flag):
			return fmt.Errorf("descriptor %s: flag rule %+v: AnyOf holds an empty flag or Flag itself", d.Collection, f)
		}
	}
	for _, list := range [][]Signature{d.NotFound, d.DeleteNotFound, d.InUse, d.Duplicate} {
		for _, s := range list {
			if s.Status == 0 || (s.Code == "" && s.FieldCode == "") {
				return fmt.Errorf("descriptor %s: signature %+v needs a status and a code", d.Collection, s)
			}
		}
	}
	return nil
}

func (d *Descriptor) notFound(method string) []Signature {
	sigs := d.NotFound
	if method == "DELETE" {
		sigs = d.DeleteNotFound
	}
	if sigs == nil {
		return defaultNotFound
	}
	return sigs
}

// collectionURI is the collection path for client ac.
func (d *Descriptor) collectionURI(ac *AlkiraClient) string {
	switch {
	case d.TenantLevel:
		return ac.URI + "/" + d.Collection
	case d.V1Path:
		return ac.URI + "/v1/tenantnetworks/" + ac.TenantNetworkId + "/" + d.Collection
	}
	return ac.URI + "/tenantnetworks/" + ac.TenantNetworkId + "/" + d.Collection
}

// entityURI is the path of one entity; id is escaped as one path segment.
func (d *Descriptor) entityURI(ac *AlkiraClient, id string) string {
	return d.collectionURI(ac) + "/" + pathSegment(id)
}

// withQuery returns uri with q as its only query string.
func withQuery(uri string, q url.Values) string {
	if len(q) == 0 {
		return uri
	}
	return uri + "?" + q.Encode()
}
