// Copyright (C) 2021-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
)

type RoutePolicySummary struct {
	Id             json.Number `json:"id"`
	Name           string      `json:"name"`
	Direction      string      `json:"direction"`
	Segment        string      `json:"segment"`
	IncludedGroups []int       `json:"includedGroups"`
	ExcludedGroups []int       `json:"excludedGroups,omitempty"`
}

type RoutePolicy struct {
	Name                          string             `json:"name"`
	Description                   string             `json:"description"`
	Enabled                       bool               `json:"enabled"`
	Direction                     string             `json:"direction"`
	Segment                       string             `json:"segment"`
	IncludedGroups                []int              `json:"includedGroups"`
	ExcludedGroups                []int              `json:"excludedGroups,omitempty"`
	Id                            json.Number        `json:"id,omitempty"` // response only
	AdvertiseInternetExit         *bool              `json:"advertiseInternetExit"`
	AdvertiseOnPremRoutes         bool               `json:"advertiseOnPremRoutes,omitempty"`
	AdvertiseCustomRoutesPrefixId int                `json:"advertiseCustomRoutesPrefixId,omitempty"`
	EnableASOverride              *bool              `json:"enableASOverride,omitempty"`
	TargetConnectorCategory       string             `json:"targetConnectorCategory,omitempty"`
	SourceRoutesPrefixListId      int                `json:"sourceRoutesPrefixListId,omitempty"`
	Rules                         []RoutePolicyRules `json:"rules,omitempty"`
}

type RoutePolicyRules struct {
	Action                       string                                        `json:"action"`
	Name                         string                                        `json:"name"`
	Match                        RoutePolicyRulesMatch                         `json:"match"`
	SequenceNo                   int                                           `json:"sequenceNo,omitempty"` // response only
	Set                          *RoutePolicyRulesSet                          `json:"set,omitempty"`
	InterCxpRoutesRedistribution *RoutePolicyRulesInterCxpRoutesRedistribution `json:"interCxpRoutesRedistribution,omitempty"`
}

type RoutePolicyRulesMatch struct {
	All                      bool     `json:"all"`
	AsPathListIds            []int    `json:"asPathListIds,omitempty"`
	CommunityListIds         []int    `json:"communityListIds,omitempty"`
	ExtendedCommunityListIds []int    `json:"extendedCommunityListIds,omitempty"`
	PrefixListIds            []int    `json:"prefixListIds,omitempty"`
	Cxps                     []string `json:"cxps,omitempty"`
	SegmentResourceIds       []int    `json:"segmentResourceIds,omitempty"`
	ConnectorGroupIds        []int    `json:"connectorGroupIds,omitempty"`
}

type RoutePolicyRulesSet struct {
	AsPathPrepend               string `json:"asPathPrepend,omitempty"`
	Community                   string `json:"community,omitempty"`
	ExtendedCommunity           string `json:"extendedCommunity,omitempty"`
	Med                         int    `json:"med,omitempty"`
	AsPathReplaceWithSegmentAsn string `json:"asPathReplaceWithSegmentAsn,omitempty"`
}

type RoutePolicyRulesInterCxpRoutesRedistribution struct {
	DistributionType        string   `json:"distributionType,omitempty"`
	RedistributeAsSecondary bool     `json:"redistributeAsSecondary,omitempty"`
	RestrictedCxps          []string `json:"restrictedCxps,omitempty"`
}

// NewRoutePolicy new route policy
func NewRoutePolicy(ac *AlkiraClient) *AlkiraApi[RoutePolicySummary] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/route-policies", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[RoutePolicySummary]{ac, uri, PaginationOn}
	return api
}

// RoutePolicyConfig is the write model of a route policy (TPS
// model/routepolicy/RoutePolicy.java). type is not modelled: the backend
// reads it as USER_CONFIGURED when unset (RoutePolicy.java:140-142), and the
// merge resends the stored value. targetConnectorCategory and
// sourceRoutesPrefixListId are sent only while
// RELEASE-AZURE-VGW-AND-GCP-VPC-ROUTE-POLICY is on, as the portal does.
type RoutePolicyConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=RoutePolicy.java:68-72|RoutePolicyEntity.java:56-58"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,trimmed,charset=name,src=RoutePolicy.java:74-77|RoutePolicyEntity.java:66-67"`
	Segment     string      `json:"segment" ak:"required,maxlen=256,trimmed,charset=name,src=RoutePolicy.java:79-83"`
	SegmentID   json.Number `json:"segmentId,omitempty" ak:"readonly"`
	Direction   *string     `json:"direction,omitempty" ak:"required,immutable,enum=INBOUND|OUTBOUND,src=RoutePolicy.java:87-88|Defaults.java:647-650|RoutePolicyServiceImpl.java:566"`
	Enabled     *bool       `json:"enabled,omitempty" ak:"required,src=RoutePolicy.java:90-91"`

	AdvertiseInternetExit         *bool  `json:"advertiseInternetExit,omitempty"`
	AdvertiseOnPremRoutes         *bool  `json:"advertiseOnPremRoutes,omitempty"`
	EnableASOverride              *bool  `json:"enableASOverride,omitempty"`
	AdvertiseCustomRoutesPrefixID *int64 `json:"advertiseCustomRoutesPrefixId,omitempty"`

	IncludedGroups []int64                 `json:"includedGroups,omitempty" ak:"required,src=RoutePolicy.java:101-102"`
	ExcludedGroups []int64                 `json:"excludedGroups,omitempty"`
	Rules          []RoutePolicyConfigRule `json:"rules,omitempty"`

	TargetConnectorCategory  *string `json:"targetConnectorCategory,omitempty" ak:"enum=USERS_AND_SITES|CLOUD,flag=RELEASE-AZURE-VGW-AND-GCP-VPC-ROUTE-POLICY,src=RoutePolicy.java:136|Defaults.java:225-229|CreateRoutePolicy.js:144-146"`
	SourceRoutesPrefixListID *int64  `json:"sourceRoutesPrefixListId,omitempty" ak:"flag=RELEASE-AZURE-VGW-AND-GCP-VPC-ROUTE-POLICY,src=RoutePolicy.java:138|CreateRoutePolicy.js:111-114|:162-165"`

	State               string      `json:"state,omitempty" ak:"readonly"`
	PreviousState       string      `json:"previousState,omitempty" ak:"readonly"`
	DocState            string      `json:"docState,omitempty" ak:"readonly"`
	LastConfigUpdatedAt json.Number `json:"lastConfigUpdatedAt,omitempty" ak:"readonly"`
	UpdatedBy           string      `json:"updatedBy,omitempty" ak:"readonly"`
}

// RoutePolicyConfigRule is one RoutePolicyRule.
type RoutePolicyConfigRule struct {
	SequenceNo                   *int64                                   `json:"sequenceNo,omitempty"`
	Name                         string                                   `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=RoutePolicyRule.java:38-42|RoutePolicyRuleEntity.java:49-50"`
	Match                        *RoutePolicyConfigMatch                  `json:"match,omitempty" ak:"required,src=RoutePolicyRule.java:44-46"`
	Action                       *string                                  `json:"action,omitempty" ak:"required,enum=ALLOW|DENY|ALLOW_W_SET,src=RoutePolicyRule.java:48-49|Defaults.java:652-656"`
	Set                          *RoutePolicyConfigSet                    `json:"set,omitempty"`
	InterCxpRoutesRedistribution *RoutePolicyConfigInterCxpRedistribution `json:"interCxpRoutesRedistribution,omitempty"`
}

// RoutePolicyConfigMatch is RoutePolicyRuleMatch. connectorGroupIds, cxps and
// segmentResourceIds carry tenant resource limits, not maxitems
// (RoutePolicyRuleMatch.java:37-47 @SoftLimit).
type RoutePolicyConfigMatch struct {
	All                      *bool    `json:"all,omitempty"`
	PrefixListIDs            []int64  `json:"prefixListIds,omitempty" ak:"maxitems=1,src=RoutePolicyRuleMatch.java:25-26"`
	CommunityListIDs         []int64  `json:"communityListIds,omitempty" ak:"maxitems=1,src=RoutePolicyRuleMatch.java:28-29"`
	ExtendedCommunityListIDs []int64  `json:"extendedCommunityListIds,omitempty" ak:"maxitems=1,src=RoutePolicyRuleMatch.java:31-32"`
	AsPathListIDs            []int64  `json:"asPathListIds,omitempty" ak:"maxitems=1,src=RoutePolicyRuleMatch.java:34-35"`
	ConnectorGroupIDs        []int64  `json:"connectorGroupIds,omitempty"`
	Cxps                     []string `json:"cxps,omitempty"`
	SegmentResourceIDs       []int64  `json:"segmentResourceIds,omitempty"`
}

// RoutePolicyConfigSet is RoutePolicySetParams.
type RoutePolicyConfigSet struct {
	AsPathPrepend               *string `json:"asPathPrepend,omitempty"`
	Community                   *string `json:"community,omitempty"`
	ExtendedCommunity           *string `json:"extendedCommunity,omitempty"`
	Med                         *int64  `json:"med,omitempty"`
	AsPathReplaceWithSegmentAsn *string `json:"asPathReplaceWithSegmentAsn,omitempty"`
}

// RoutePolicyConfigInterCxpRedistribution is RoutePolicyInterCxpRoutesRedistribution.
type RoutePolicyConfigInterCxpRedistribution struct {
	DistributionType        *string  `json:"distributionType,omitempty" ak:"enum=ALL|LOCAL_ONLY|RESTRICTED_CXPS,src=RoutePolicyInterCxpRoutesRedistribution.java:25-29"`
	RestrictedCxps          []string `json:"restrictedCxps,omitempty"`
	RedistributeAsSecondary *bool    `json:"redistributeAsSecondary,omitempty"`
}

// routePolicyDescriptor: GET answers a missing id with 404 AK000003 "route
// policy <id> not found" and PUT and DELETE with 404 AK000003 "routePolicy
// with id <id> not found" (RoutePolicyServiceImpl.java:313, :431). A DELETE of
// a policy still referred externally is 409 (NetworkConfigEntityListener.java:96);
// a duplicate name is 400 constraint.exists (:504). An update runs
// ProvisionedEntityValidator.onModify (:1195), the D13 rule
// (ProvisionedEntityValidatorImpl.java:34-73). A delete of a provisioned
// policy marks it MARKED_FOR_DELETION (:1216), which a plain GET hides (:312-313).
var routePolicyDescriptor = Descriptor{
	Collection: "route-policies",
	InUse:      []Signature{{Status: 409, Code: "AK000001"}},
	Duplicate:  []Signature{{Status: 400, Code: "AK000001", FieldCode: "constraint.exists"}},

	Provisionable:      true,
	MarkedRowLifecycle: true,
	DocState:           DocStateMissingIsAdded,
	Uniqueness:         Uniqueness{Key: "name", Scope: ScopeCollection},
	StateGuard:         StateGuardD13,
	Service:            ServiceTPS,
	Flags: []FlagRule{
		{Flag: "RELEASE-AZURE-VGW-AND-GCP-VPC-ROUTE-POLICY", Effect: FlagIncludeField, Field: "targetConnectorCategory", UISource: "src/views/network/provision/policies/routePolicies/CreateRoutePolicy.js:103-105|:144-146"},
		{Flag: "RELEASE-AZURE-VGW-AND-GCP-VPC-ROUTE-POLICY", Effect: FlagIncludeField, Field: "sourceRoutesPrefixListId", UISource: "src/views/network/provision/policies/routePolicies/CreateRoutePolicy.js:103-105|:111-114|:162-165"},
	},
}

// RoutePolicyDescriptor returns the route policy write descriptor.
func RoutePolicyDescriptor() Descriptor { return routePolicyDescriptor.clone() }

// NewRoutePolicyCore returns the write core for route policies.
func NewRoutePolicyCore(ac *AlkiraClient) (*Core[RoutePolicyConfig], error) {
	return NewCore[RoutePolicyConfig](ac, routePolicyDescriptor)
}

// CreateRoutePolicy creates a route policy.
func (ac *AlkiraClient) CreateRoutePolicy(ctx context.Context, p *RoutePolicyConfig) (*RoutePolicyConfig, *WriteResult, error) {
	c, err := NewRoutePolicyCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Create(ctx, p, WriteOptions{})
}

// GetRoutePolicy reads one route policy and its raw body.
func (ac *AlkiraClient) GetRoutePolicy(ctx context.Context, id string) (*RoutePolicyConfig, []byte, error) {
	c, err := NewRoutePolicyCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Get(ctx, id)
}

// UpdateRoutePolicy applies a merge patch to a route policy.
func (ac *AlkiraClient) UpdateRoutePolicy(ctx context.Context, id string, patch json.RawMessage) (*RoutePolicyConfig, *WriteResult, error) {
	c, err := NewRoutePolicyCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Update(ctx, id, patch, ListEdits{})
}

// DeleteRoutePolicy deletes a route policy.
func (ac *AlkiraClient) DeleteRoutePolicy(ctx context.Context, id string) (*WriteResult, error) {
	c, err := NewRoutePolicyCore(ac)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, id)
}

// ListRoutePolicyPaged reads every route policy through paged reads.
func (ac *AlkiraClient) ListRoutePolicyPaged(ctx context.Context) ([]RoutePolicyConfig, error) {
	c, err := NewRoutePolicyCore(ac)
	if err != nil {
		return nil, err
	}
	return c.ListPaged(ctx)
}
