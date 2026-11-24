// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"fmt"
	"net/url"
	"strings"
)

// Signature identifies one backend error response by status plus the
// ErrorDetails code. Code matches the top-level code and FieldCode matches the
// code of any errors[] entry; at least one of them is set. Bodies that are not
// ErrorDetails JSON (empty, HTML, a gateway {"message":...} without code)
// never match.
type Signature struct {
	Status    int
	Code      string
	FieldCode string
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

// Service is the backend that owns a family.
type Service int

const (
	ServiceTPS Service = iota
	ServicePolicy
	ServiceTenantBilling
)

// FamilyCheck identifies a family from its GET body: Field must be present
// and, when Value is set, equal Value. A sibling family's id answers 200 on
// another family's path, so the check runs before every PUT and DELETE.
type FamilyCheck struct {
	Field string
	Value string
}

// FlagEffect is what a feature flag changes in a create or update.
type FlagEffect int

const (
	// FlagIncludeField: Field is sent only while the flag is on. With the flag
	// off a value the request sets is dropped, or refused when Refuse is set.
	FlagIncludeField FlagEffect = iota + 1
	// FlagTypeAvailable: the family can be written only while the flag is on.
	FlagTypeAvailable
)

// FlagRule is one way a feature flag shapes a family's writes, as the portal does.
type FlagRule struct {
	Flag   string
	Effect FlagEffect
	// Field is the top-level JSON field of a FlagIncludeField rule. Its zero
	// value must be left out of the JSON body (a pointer or omitempty). On a
	// `set` field the rule gates the items an update adds.
	Field string
	// Refuse refuses, rather than drops, a FlagIncludeField value set while the flag is off.
	Refuse bool
	// UISource is the portal file:line that branches on the flag.
	UISource string
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
	// Flags lists the feature flags that shape creates and updates.
	Flags []FlagRule
	// AsyncCapable: with AsyncValidationFlag on, creates and updates send
	// ?async=true and wait for the validation state.
	AsyncCapable bool
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
		case f.Flag == "" || f.Effect < FlagIncludeField || f.Effect > FlagTypeAvailable:
			return fmt.Errorf("descriptor %s: flag rule %+v needs a flag and an effect", d.Collection, f)
		case (f.Effect == FlagIncludeField) != (f.Field != ""):
			return fmt.Errorf("descriptor %s: flag rule %+v: a field goes with FlagIncludeField only", d.Collection, f)
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
