// Copyright (C) 2020-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
)

type TrafficPolicyRuleSummary struct {
	Description string      `json:"description"`
	Id          json.Number `json:"id,omitempty"`
	Name        string      `json:"name"`
}

type TrafficPolicyRule struct {
	Description    string                   `json:"description"`
	Id             json.Number              `json:"id,omitempty"`
	MatchCondition PolicyRuleMatchCondition `json:"matchCondition"`
	Name           string                   `json:"name"`
	RuleAction     PolicyRuleAction         `json:"ruleAction"`
}

type PolicyRuleMatchCondition struct {
	ApplicationList       []int    `json:"applicationList"`
	Dscp                  string   `json:"dscp"`
	DstIp                 string   `json:"dstIp,omitempty"`
	DstPortList           []string `json:"dstPortList,omitempty"`
	DstPrefixListId       int      `json:"dstPrefixListId,omitempty"`
	InternetApplicationId int      `json:"internetApplicationId,omitempty"`
	Protocol              string   `json:"protocol"`
	SrcIp                 string   `json:"srcIp,omitempty"`
	SrcPortList           []string `json:"srcPortList,omitempty"`
	SrcPrefixListId       int      `json:"srcPrefixListId,omitempty"`
}

type PolicyRuleAction struct {
	Action          string   `json:"action"`
	ServiceTypeList []string `json:"serviceTypeList"`
	ServiceList     []int    `json:"serviceList"`
	FlowCollectors  []int    `json:"flowCollectors,omitempty"`
}

// NewTrafficPolicyRule new traffic policy rule
func NewTrafficPolicyRule(ac *AlkiraClient) *AlkiraApi[TrafficPolicyRuleSummary] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/policy/rules", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[TrafficPolicyRuleSummary]{ac, uri, PaginationOff}
	return api
}

// TrafficPolicyRuleConfig is the write model of a traffic policy rule (policy-service
// model/Rule.java). A PUT overwrites every match and action field from the
// body (RuleServiceImpl.java:374-387), so the merge resends the full GET
// body; applicationFamilyList is cleared when a body leaves it out
// (V3 internal/client/policy_rule.go:15-17). The backend refuses a rule
// whose source is neither or both of srcIp and srcPrefixListId, or whose
// destination is not exactly one of dstIp, dstPrefixListId and
// internetApplicationId (RuleServiceImpl.java:709-717); the portal sends
// "any" for an unset srcIp or dstIp (createRule/index.js:166-190).
type TrafficPolicyRuleConfig struct {
	ID              json.Number `json:"id,omitempty" ak:"readonly"`
	TenantID        json.Number `json:"tenantId,omitempty" ak:"readonly"`
	TenantNetworkID json.Number `json:"tenantNetworkId,omitempty" ak:"readonly"`
	Name            string      `json:"name" ak:"required,maxlen=255,trimmed,charset=name,src=Rule.java:41-45"`
	Description     *string     `json:"description,omitempty" ak:"maxlen=255,trimmed,charset=name,src=Rule.java:46-49"`
	InternalName    string      `json:"internalName,omitempty" ak:"readonly"`

	MatchCondition *TrafficPolicyRuleConfigMatch  `json:"matchCondition,omitempty" ak:"required,src=Rule.java:54-56|RuleServiceImpl.java:690-692"`
	RuleAction     *TrafficPolicyRuleConfigAction `json:"ruleAction,omitempty" ak:"required,src=Rule.java:57-59|RuleServiceImpl.java:693-695"`

	State               string      `json:"state,omitempty" ak:"readonly"`
	DocState            string      `json:"docState,omitempty" ak:"readonly"`
	Type                string      `json:"type,omitempty" ak:"readonly"`
	LastUpdatedDate     json.Number `json:"lastUpdatedDate,omitempty" ak:"readonly"`
	UpdatedBy           string      `json:"updatedBy,omitempty" ak:"readonly"`
	LastConfigUpdatedAt json.Number `json:"lastConfigUpdatedAt,omitempty" ak:"readonly"`
}

// TrafficPolicyRuleConfigMatch is Rule.matchCondition (MatchCondition.java). The port
// lists are sets: the backend stores and returns them in its own order.
// A dstPortList is needed unless internetApplicationId is set
// (RuleServiceImpl.java:699-701); applicationFamilyList is refused while
// DEPRECATE-QOSMOS-APPLICATION-FAMILY is on (:718-722), so with the flag on a
// create or patch that sets it non-empty is refused, and an update sends a
// stored list empty, as the portal does. Sending a requested list empty
// would widen a DROP rule to the whole 5-tuple.
type TrafficPolicyRuleConfigMatch struct {
	SrcIP                 *string  `json:"srcIp,omitempty"`
	DstIP                 *string  `json:"dstIp,omitempty"`
	Dscp                  *string  `json:"dscp,omitempty" ak:"default=any,src=MatchCondition.java:32-33|DSCPValidator.java:13-16|createRule/index.js:198-200"`
	Protocol              *string  `json:"protocol,omitempty" ak:"required,enum=icmp|tcp|udp|any,src=MatchCondition.java:35-37"`
	SrcPrefixListID       *int64   `json:"srcPrefixListId,omitempty"`
	DstPrefixListID       *int64   `json:"dstPrefixListId,omitempty"`
	SrcPortList           []string `json:"srcPortList,omitempty" ak:"required,minitems=1,src=MatchCondition.java:42-44|RuleServiceImpl.java:696-698|createRule/index.js:195-197"`
	DstPortList           []string `json:"dstPortList,omitempty"`
	ApplicationList       []int64  `json:"applicationList,omitempty" ak:"maxitems=1,src=MatchCondition.java:50-52"`
	ApplicationFamilyList []int64  `json:"applicationFamilyList,omitempty" ak:"maxitems=1,src=MatchCondition.java:53-55"`
	InternetApplicationID *int64   `json:"internetApplicationId,omitempty"`
}

// TrafficPolicyRuleConfigAction is Rule.ruleAction (RuleAction.java).
type TrafficPolicyRuleConfigAction struct {
	Action          *string  `json:"action,omitempty" ak:"required,enum=ALLOW|DROP,src=RuleAction.java:23-25"`
	ServiceTypeList []string `json:"serviceTypeList,omitempty"`
	ServiceList     []int64  `json:"serviceList,omitempty"`
	FlowCollectors  []int64  `json:"flowCollectors,omitempty"`
}

// trafficPolicyRuleDescriptor: GET, PUT and DELETE answer a missing id with 404 AK000003
// "rule with id <id> is not present" (RuleServiceImpl.java:264, :433, :488;
// PolicyExceptionHandler.java:52-53). A DELETE of a rule held by a rule list
// is 400 rule.00000003 (:457); a duplicate name is 400 rule.00000004 (:688).
// PUT and DELETE answer 202 (RT MOCK-EVIDENCE-MATRIX.md:1921). A delete marks
// the rule MARKED_FOR_DELETION unless it is PENDING, a rule without state
// included (:462-468); a rule deleted with ?provision=true can stay marked
// after its delete request is SUCCESS (rule 1514 on network 207, RT
// BACKEND-FOLLOWUPS.md:224-230). The body carries docState, ADDED for a live row.
var trafficPolicyRuleDescriptor = Descriptor{
	Collection: "policy/rules",
	InUse:      []Signature{{Status: 400, FieldCode: "rule.00000003"}},
	Duplicate:  []Signature{{Status: 400, FieldCode: "rule.00000004"}},

	Provisionable:      true,
	MarkedRowLifecycle: true,
	EmptyStateDeployed: true,
	DocState:           DocStateMissingIsAdded,
	Uniqueness:         Uniqueness{Key: "name", Scope: ScopeCollection},
	Service:            ServicePolicy,
	Flags: []FlagRule{
		{Flag: "DEPRECATE-QOSMOS-APPLICATION-FAMILY", Effect: FlagClearField, Field: "matchCondition.applicationFamilyList", UISource: "src/views/rules/rulesManagement/createRule/index.js:88-90|matchSelection.js:203-207|:296|:341|:369"},
	},
}

// TrafficPolicyRuleDescriptor returns the traffic policy rule write descriptor.
func TrafficPolicyRuleDescriptor() Descriptor { return trafficPolicyRuleDescriptor.clone() }

// NewTrafficPolicyRuleCore returns the write core for traffic policy rules.
func NewTrafficPolicyRuleCore(ac *AlkiraClient) (*Core[TrafficPolicyRuleConfig], error) {
	return NewCore[TrafficPolicyRuleConfig](ac, trafficPolicyRuleDescriptor)
}

// CreateTrafficPolicyRule creates a traffic policy rule.
func (ac *AlkiraClient) CreateTrafficPolicyRule(ctx context.Context, r *TrafficPolicyRuleConfig) (*TrafficPolicyRuleConfig, *WriteResult, error) {
	c, err := NewTrafficPolicyRuleCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Create(ctx, r, WriteOptions{})
}

// GetTrafficPolicyRule reads one traffic policy rule and its raw body.
func (ac *AlkiraClient) GetTrafficPolicyRule(ctx context.Context, id string) (*TrafficPolicyRuleConfig, []byte, error) {
	c, err := NewTrafficPolicyRuleCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Get(ctx, id)
}

// UpdateTrafficPolicyRule applies a merge patch to a traffic policy rule.
func (ac *AlkiraClient) UpdateTrafficPolicyRule(ctx context.Context, id string, patch json.RawMessage) (*TrafficPolicyRuleConfig, *WriteResult, error) {
	c, err := NewTrafficPolicyRuleCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Update(ctx, id, patch, ListEdits{})
}

// DeleteTrafficPolicyRule deletes a traffic policy rule.
func (ac *AlkiraClient) DeleteTrafficPolicyRule(ctx context.Context, id string) (*WriteResult, error) {
	c, err := NewTrafficPolicyRuleCore(ac)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, id)
}

// ListTrafficPolicyRulePaged reads every traffic policy rule through paged reads.
func (ac *AlkiraClient) ListTrafficPolicyRulePaged(ctx context.Context) ([]TrafficPolicyRuleConfig, error) {
	c, err := NewTrafficPolicyRuleCore(ac)
	if err != nil {
		return nil, err
	}
	return c.ListPaged(ctx)
}
