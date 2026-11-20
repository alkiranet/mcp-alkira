// Copyright (C) 2020-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
)

// A summary of the resource (other fields are ignored)
type TrafficPolicySummary struct {
	Id           json.Number `json:"id"`
	Name         string      `json:"name"`
	InternalName string      `json:"internalName"`
	RuleListId   int         `json:"ruleListId"`
	SegmentIds   []int       `json:"segmentIds"`
}

type TrafficPolicy struct {
	Description   string      `json:"description"`
	Enabled       bool        `json:"enabled"`
	FromGroups    []int       `json:"fromGroups"`
	Id            json.Number `json:"id,omitempty"`
	Name          string      `json:"name"`
	RuleListId    int         `json:"ruleListId"`
	SegmentIds    []int       `json:"segmentIds"`
	ToGroups      []int       `json:"toGroups"`
	ZTAProfileIds []string    `json:"ztaProfileIds"`
}

// NewTrafficPolicy new traffic policy
func NewTrafficPolicy(ac *AlkiraClient) *AlkiraApi[TrafficPolicySummary] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/policy/policies", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[TrafficPolicySummary]{ac, uri, PaginationOff}
	return api
}

// TrafficPolicyConfig is the write model of a traffic policy (policy-service
// model/Policy.java). A PUT sets every field from the body
// (PolicyServiceImpl.java:618-626), so the merge resends the full GET body.
// ztaProfileIds is sent only while RELEASE-ZTNA-POLICY-WITH-PROFILES is on,
// as the portal does (createPolicy.js:383-385); the backend keeps the ids
// only with the flag on (PolicyServiceImpl.java:254, :629), so a tenant
// without it drops them without an error (V3 docs/adr/ADR-3-conventions.md:165).
// Its limit of one profile is a tenant resource limit (Policy.java:98-99 @SoftLimit).
type TrafficPolicyConfig struct {
	ID              json.Number `json:"id,omitempty" ak:"readonly"`
	TenantID        json.Number `json:"tenantId,omitempty" ak:"readonly"`
	TenantNetworkID json.Number `json:"tenantNetworkId,omitempty" ak:"readonly"`
	Name            string      `json:"name" ak:"required,maxlen=255,trimmed,charset=name,src=Policy.java:57-61"`
	Description     *string     `json:"description,omitempty" ak:"maxlen=255,trimmed,charset=name,src=Policy.java:62-65"`

	SegmentIDs []int64 `json:"segmentIds,omitempty" ak:"required,minitems=1,maxitems=1,src=Policy.java:66-68"`
	FromGroups []int64 `json:"fromGroups,omitempty" ak:"required,minitems=1,src=Policy.java:69-71"`
	ToGroups   []int64 `json:"toGroups,omitempty" ak:"required,minitems=1,src=Policy.java:72-74"`
	// @NotNull alone lets 0 through, which names no rule list.
	RuleListID *int64 `json:"ruleListId,omitempty" ak:"required,min=1,src=Policy.java:75-76"`
	Enabled    *bool  `json:"enabled,omitempty" ak:"default=true,src=Policy.java:77-79|createPolicy.js:387-393"`

	Bidirectional *bool    `json:"bidirectional,omitempty"`
	ZtaProfileIDs []string `json:"ztaProfileIds,omitempty" ak:"flag=RELEASE-ZTNA-POLICY-WITH-PROFILES,src=Policy.java:98-100|createPolicy.js:383-385"`

	State               string      `json:"state,omitempty" ak:"readonly"`
	DocState            string      `json:"docState,omitempty" ak:"readonly"`
	Type                string      `json:"type,omitempty" ak:"readonly"`
	InternalName        string      `json:"internalName,omitempty" ak:"readonly"`
	LastUpdatedDate     json.Number `json:"lastUpdatedDate,omitempty" ak:"readonly"`
	UpdatedBy           string      `json:"updatedBy,omitempty" ak:"readonly"`
	LastConfigUpdatedAt json.Number `json:"lastConfigUpdatedAt,omitempty" ak:"readonly"`
}

// trafficPolicyDescriptor: GET, PUT and DELETE answer a missing id with 404
// AK000003 "policy with id <id> ... not present" (PolicyServiceImpl.java:568,
// :689, :699; PolicyExceptionHandler.java:52-53). A duplicate name is 400
// policy.00000002 (:372). A policy delete has no in-use refusal: the only
// DELETE refusal is policy.00000004 for a policy that is not USER_CONFIGURED
// (:494). A delete marks the policy MARKED_FOR_DELETION unless it is PENDING,
// a policy without state included, and a disabled deployed policy DELETED
// (:505-510). PUT and DELETE answer 202.
var trafficPolicyDescriptor = Descriptor{
	Collection: "policy/policies",
	Duplicate:  []Signature{{Status: 400, FieldCode: "policy.00000002"}},

	Provisionable:      true,
	MarkedRowLifecycle: true,
	EmptyStateDeployed: true,
	DocState:           DocStateMissingIsAdded,
	Uniqueness:         Uniqueness{Key: "name", Scope: ScopeCollection},
	Service:            ServicePolicy,
	Flags: []FlagRule{
		{Flag: "RELEASE-ZTNA-POLICY-WITH-PROFILES", Effect: FlagIncludeField, Field: "ztaProfileIds", UISource: "src/views/network/provision/policies/trafficPolicies/createPolicy.js:383-385"},
	},
}

// TrafficPolicyDescriptor returns the traffic policy write descriptor.
func TrafficPolicyDescriptor() Descriptor { return trafficPolicyDescriptor.clone() }

// NewTrafficPolicyCore returns the write core for traffic policies.
func NewTrafficPolicyCore(ac *AlkiraClient) (*Core[TrafficPolicyConfig], error) {
	return NewCore[TrafficPolicyConfig](ac, trafficPolicyDescriptor)
}

// CreateTrafficPolicy creates a traffic policy.
func (ac *AlkiraClient) CreateTrafficPolicy(ctx context.Context, p *TrafficPolicyConfig) (*TrafficPolicyConfig, *WriteResult, error) {
	c, err := NewTrafficPolicyCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Create(ctx, p, WriteOptions{})
}

// GetTrafficPolicy reads one traffic policy and its raw body.
func (ac *AlkiraClient) GetTrafficPolicy(ctx context.Context, id string) (*TrafficPolicyConfig, []byte, error) {
	c, err := NewTrafficPolicyCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Get(ctx, id)
}

// UpdateTrafficPolicy applies a merge patch to a traffic policy.
func (ac *AlkiraClient) UpdateTrafficPolicy(ctx context.Context, id string, patch json.RawMessage) (*TrafficPolicyConfig, *WriteResult, error) {
	c, err := NewTrafficPolicyCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Update(ctx, id, patch, ListEdits{})
}

// DeleteTrafficPolicy deletes a traffic policy.
func (ac *AlkiraClient) DeleteTrafficPolicy(ctx context.Context, id string) (*WriteResult, error) {
	c, err := NewTrafficPolicyCore(ac)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, id)
}

// ListTrafficPolicyPaged reads every traffic policy through paged reads.
func (ac *AlkiraClient) ListTrafficPolicyPaged(ctx context.Context) ([]TrafficPolicyConfig, error) {
	c, err := NewTrafficPolicyCore(ac)
	if err != nil {
		return nil, err
	}
	return c.ListPaged(ctx)
}
