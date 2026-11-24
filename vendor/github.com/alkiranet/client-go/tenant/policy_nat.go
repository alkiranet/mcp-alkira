// Copyright (C) 2021-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
)

type NATPolicy struct {
	Name                               string      `json:"name"`
	Description                        string      `json:"description"`
	Type                               string      `json:"type"`
	Segment                            string      `json:"segment"`
	IncludedGroups                     []int       `json:"includedGroups"`
	ExcludedGroups                     []int       `json:"excludedGroups"`
	Id                                 json.Number `json:"id,omitempty"`
	NATRuleIds                         []int       `json:"natRuleIds"`
	Category                           string      `json:"category"`
	AllowOverlappingTranslatedPrefixes *bool       `json:"allowOverlappingTranslatedPrefixes"`
}

// A summary of a NAT policy (other fields are ignored)
type NATPolicySummary struct {
	Id         json.Number `json:"id"`
	Name       string      `json:"name"`
	Segment    string      `json:"segment"`
	NATRuleIds []int       `json:"natRuleIds"`
}

// NewNATPolicy new NAT policy
func NewNATPolicy(ac *AlkiraClient) *AlkiraApi[NATPolicySummary] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/nat-policies", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[NATPolicySummary]{ac, uri, PaginationOn}
	return api
}

// NatPolicyConfig is the write model of a NAT policy (TPS model/NATPolicy.java).
// A PUT replaces the stored policy: an omitted field is cleared
// (PHASE-NAT-PROBE-RESULTS.md:115), so the merge resends the full GET body.
// natRuleIds that name no rule are dropped by the backend without an error
// (NATPolicyServiceImpl.java:413-416). The model has no direction field.
type NatPolicyConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=NATPolicy.java:65-69|NATPolicyEntity.java:63-65|PHASE-NAT-PROBE-RESULTS.md:166-167"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,trimmed,charset=name,src=NATPolicy.java:71-74|PHASE-NAT-PROBE-RESULTS.md:172-173"`
	Type        *string     `json:"type,omitempty" ak:"default=INTRA_SEGMENT,enum=INTRA_SEGMENT|INTER_SEGMENT,src=NATPolicy.java:76-77|NATPolicy.java:146-150|natPolicies/index.js:518"`
	Segment     string      `json:"segment" ak:"required,trimmed,charset=name,src=NATPolicy.java:79-82"`
	SegmentID   json.Number `json:"segmentId,omitempty" ak:"readonly"`

	IncludedGroups []int64 `json:"includedGroups,omitempty" ak:"required,src=NATPolicy.java:87|NATPolicyServiceImpl.java:407"`
	ExcludedGroups []int64 `json:"excludedGroups,omitempty"`
	// A missing list fails the create with a 500 (NATPolicyServiceImpl.java:413).
	NatRuleIDs []int64 `json:"natRuleIds,omitempty" ak:"required,src=NATPolicy.java:117-119|NATPolicyServiceImpl.java:413"`
	Category   *string `json:"category,omitempty" ak:"default=DEFAULT,src=NATPolicy.java:121|natPolicies/index.js:101|:522"`

	AllowOverlappingTranslatedPrefixes          *bool `json:"allowOverlappingTranslatedPrefixes,omitempty" ak:"flag=NAT-POLICY-ALLOW-OVERLAPPING-TRANSLATED-PREFIXES,src=NATPolicy.java:133|NATPolicyServiceValidator.java:521|natPolicies/index.js:523-529"`
	AllowOverlappingOutboundDestinationPrefixes *bool `json:"allowOverlappingOutboundDestinationPrefixes,omitempty" ak:"flag=RELEASE-NAT-POLICY-OUTBOUND-DNAT,src=NATPolicy.java:142|NATPolicyServiceValidator.java:162-176|natPolicies/index.js:530-536"`

	State               string      `json:"state,omitempty" ak:"readonly"`
	PreviousState       string      `json:"previousState,omitempty" ak:"readonly"`
	DocState            string      `json:"docState,omitempty" ak:"readonly"`
	LastConfigUpdatedAt json.Number `json:"lastConfigUpdatedAt,omitempty" ak:"readonly"`
	UpdatedBy           string      `json:"updatedBy,omitempty" ak:"readonly"`
}

// natPolicyDescriptor: GET and PUT answer a missing id with 404 AK000003 "nat
// policy <id> not found" and DELETE with 404 AK000003
// (NATPolicyServiceImpl.java:267, :381; PHASE-NAT-PROBE-RESULTS.md:24-26). A
// DELETE of a policy still referred externally is 409
// (NetworkConfigEntityListener.java:96). The backend accepts a second policy
// with the same name (PHASE-NAT-PROBE-RESULTS.md:47), so there is no
// duplicate signature; the create guard pages the collection, since ?name=
// answers 400 once two policies share a name (:48). A delete of a provisioned
// policy marks it MARKED_FOR_DELETION (:548), which a plain GET hides
// (:380-381). Updates carry no state check in TPS (only delete calls
// ProvisionedEntityValidator, :543).
//
// allowOverlappingOutboundDestinationPrefixes is gated by two rules, so it is
// sent only while both flags are on, as the portal does
// (natPolicies/index.js:246-248); the backend needs only
// RELEASE-NAT-POLICY-OUTBOUND-DNAT for true (NATPolicyServiceValidator.java:167-170).
var natPolicyDescriptor = Descriptor{
	Collection: "nat-policies",
	InUse:      []Signature{{Status: 409, Code: "AK000001"}},

	Provisionable:      true,
	MarkedRowLifecycle: true,
	DocState:           DocStateMissingIsAdded,
	Uniqueness:         Uniqueness{Key: "name", Scope: ScopeCollection},
	Service:            ServiceTPS,
	Flags: []FlagRule{
		{Flag: "NAT-POLICY-ALLOW-OVERLAPPING-TRANSLATED-PREFIXES", Effect: FlagIncludeField, Field: "allowOverlappingTranslatedPrefixes", UISource: "src/views/network/provision/policies/natPolicies/index.js:243-245|:523-529"},
		{Flag: "RELEASE-NAT-POLICY-OUTBOUND", Effect: FlagIncludeField, Field: "allowOverlappingOutboundDestinationPrefixes", UISource: "src/views/network/provision/policies/natPolicies/index.js:246-248|:530-536"},
		{Flag: "RELEASE-NAT-POLICY-OUTBOUND-DNAT", Effect: FlagIncludeField, Field: "allowOverlappingOutboundDestinationPrefixes", UISource: "src/views/network/provision/policies/natPolicies/index.js:246-248|:530-536"},
	},
}

// NatPolicyDescriptor returns the NAT policy write descriptor.
func NatPolicyDescriptor() Descriptor { return natPolicyDescriptor.clone() }

// NewNatPolicyCore returns the write core for NAT policies.
func NewNatPolicyCore(ac *AlkiraClient) (*Core[NatPolicyConfig], error) {
	return NewCore[NatPolicyConfig](ac, natPolicyDescriptor)
}

// CreateNatPolicy creates a NAT policy.
func (ac *AlkiraClient) CreateNatPolicy(ctx context.Context, p *NatPolicyConfig) (*NatPolicyConfig, *WriteResult, error) {
	c, err := NewNatPolicyCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Create(ctx, p, WriteOptions{})
}

// GetNatPolicy reads one NAT policy and its raw body.
func (ac *AlkiraClient) GetNatPolicy(ctx context.Context, id string) (*NatPolicyConfig, []byte, error) {
	c, err := NewNatPolicyCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Get(ctx, id)
}

// UpdateNatPolicy applies a merge patch to a NAT policy.
func (ac *AlkiraClient) UpdateNatPolicy(ctx context.Context, id string, patch json.RawMessage) (*NatPolicyConfig, *WriteResult, error) {
	c, err := NewNatPolicyCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Update(ctx, id, patch, ListEdits{})
}

// DeleteNatPolicy deletes a NAT policy.
func (ac *AlkiraClient) DeleteNatPolicy(ctx context.Context, id string) (*WriteResult, error) {
	c, err := NewNatPolicyCore(ac)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, id)
}

// ListNatPolicyPaged reads every NAT policy through paged reads; rows that
// share a name are all returned.
func (ac *AlkiraClient) ListNatPolicyPaged(ctx context.Context) ([]NatPolicyConfig, error) {
	c, err := NewNatPolicyCore(ac)
	if err != nil {
		return nil, err
	}
	return c.ListPaged(ctx)
}
