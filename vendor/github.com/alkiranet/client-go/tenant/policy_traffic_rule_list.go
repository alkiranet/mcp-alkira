// Copyright (C) 2020-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
)

type TrafficPolicyRuleListSummary struct {
	Id   json.Number `json:"id"`
	Name string      `json:"name"`
	Type string      `json:"type"`
}

type TrafficPolicyRuleList struct {
	Description string                      `json:"description"`
	Id          json.Number                 `json:"id,omitempty"`
	Name        string                      `json:"name"`
	Rules       []TrafficPolicyRuleListRule `json:"rules"`
}

type TrafficPolicyRuleListRule struct {
	Priority int `json:"priority"`
	RuleId   int `json:"ruleId"`
}

// NewTrafficPolicyRuleList new traffic policy rule list
func NewTrafficPolicyRuleList(ac *AlkiraClient) *AlkiraApi[TrafficPolicyRuleListSummary] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/policy/rulelists", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[TrafficPolicyRuleListSummary]{ac, uri, PaginationOn}
	return api
}

// TrafficPolicyRuleListConfig is the write model of a traffic policy rule list
// (policy-service model/RuleList.java). The body has no state and no
// docState; a rule list held only by policies marked for deletion is marked
// DELETED on delete, and a plain GET hides it (RuleListServiceImpl.java:244-246,
// :367-371).
type TrafficPolicyRuleListConfig struct {
	ID              json.Number `json:"id,omitempty" ak:"readonly"`
	TenantID        json.Number `json:"tenantId,omitempty" ak:"readonly"`
	TenantNetworkID json.Number `json:"tenantNetworkId,omitempty" ak:"readonly"`
	LastUpdatedDate json.Number `json:"lastUpdatedDate,omitempty" ak:"readonly"`
	Name            string      `json:"name" ak:"required,maxlen=255,trimmed,charset=name,src=RuleList.java:43-47"`
	Description     *string     `json:"description,omitempty" ak:"maxlen=255,trimmed,charset=name,src=RuleList.java:48-51"`
	// Priorities and rule ids must each be distinct (RuleListServiceImpl.java:163-168, :180-183).
	// The portal never sends a repeat: each rule is one row of a selection
	// grid and priorities follow list position (rulesSelection/index.js:38-69;
	// createRuleList.js:181-182).
	Rules []TrafficPolicyRuleListConfigRule `json:"rules,omitempty" ak:"required,minitems=1,distinct=priority|ruleId,src=RuleList.java:52-54|RuleListServiceImpl.java:163-168|RuleListServiceImpl.java:180-183|rulesSelection/index.js:38-69|createRuleList.js:181-182"`

	Type                string      `json:"type,omitempty" ak:"readonly"`
	UpdatedBy           string      `json:"updatedBy,omitempty" ak:"readonly"`
	LastConfigUpdatedAt json.Number `json:"lastConfigUpdatedAt,omitempty" ak:"readonly"`
}

// TrafficPolicyRuleListConfigRule is one RuleList.rules entry (OrderedRule.java). The
// portal numbers priorities from 1 in list order (createRuleList.js:181-182).
type TrafficPolicyRuleListConfigRule struct {
	RuleID   *int64 `json:"ruleId,omitempty" ak:"required,src=OrderedRule.java:19-20"`
	Priority *int64 `json:"priority,omitempty" ak:"required,min=1,src=OrderedRule.java:21-22|createRuleList.js:181-182"`
}

// trafficPolicyRuleListDescriptor: GET, PUT and DELETE answer a missing id with 404
// AK000003 "rule list with id <id> is not present" (RuleListServiceImpl.java:212,
// :258, :347, :357); a PUT of a rule list that is not USER_CONFIGURED is the
// same 404 (:268, :347). A DELETE of a rule list used by a policy is 400
// rulelist.00000004 (:248); a duplicate name is 400 rulelist.00000001 (:178).
// PUT and DELETE answer 202.
var trafficPolicyRuleListDescriptor = Descriptor{
	Collection: "policy/rulelists",
	InUse:      []Signature{{Status: 400, FieldCode: "rulelist.00000004"}},
	Duplicate:  []Signature{{Status: 400, FieldCode: "rulelist.00000001"}},

	Provisionable: true,
	DocState:      DocStateMissingIsAdded,
	Uniqueness:    Uniqueness{Key: "name", Scope: ScopeCollection},
	Service:       ServicePolicy,
}

// TrafficPolicyRuleListDescriptor returns the traffic policy rule list write descriptor.
func TrafficPolicyRuleListDescriptor() Descriptor { return trafficPolicyRuleListDescriptor.clone() }

// NewTrafficPolicyRuleListCore returns the write core for traffic policy rule lists.
func NewTrafficPolicyRuleListCore(ac *AlkiraClient) (*Core[TrafficPolicyRuleListConfig], error) {
	return NewCore[TrafficPolicyRuleListConfig](ac, trafficPolicyRuleListDescriptor)
}

// CreateTrafficPolicyRuleList creates a traffic policy rule list.
func (ac *AlkiraClient) CreateTrafficPolicyRuleList(ctx context.Context, l *TrafficPolicyRuleListConfig) (*TrafficPolicyRuleListConfig, *WriteResult, error) {
	c, err := NewTrafficPolicyRuleListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Create(ctx, l, WriteOptions{})
}

// GetTrafficPolicyRuleList reads one traffic policy rule list and its raw body.
func (ac *AlkiraClient) GetTrafficPolicyRuleList(ctx context.Context, id string) (*TrafficPolicyRuleListConfig, []byte, error) {
	c, err := NewTrafficPolicyRuleListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Get(ctx, id)
}

// UpdateTrafficPolicyRuleList applies a merge patch to a traffic policy rule list.
func (ac *AlkiraClient) UpdateTrafficPolicyRuleList(ctx context.Context, id string, patch json.RawMessage) (*TrafficPolicyRuleListConfig, *WriteResult, error) {
	c, err := NewTrafficPolicyRuleListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Update(ctx, id, patch, ListEdits{})
}

// DeleteTrafficPolicyRuleList deletes a traffic policy rule list.
func (ac *AlkiraClient) DeleteTrafficPolicyRuleList(ctx context.Context, id string) (*WriteResult, error) {
	c, err := NewTrafficPolicyRuleListCore(ac)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, id)
}

// ListTrafficPolicyRuleListPaged reads every traffic policy rule list through paged reads.
func (ac *AlkiraClient) ListTrafficPolicyRuleListPaged(ctx context.Context) ([]TrafficPolicyRuleListConfig, error) {
	c, err := NewTrafficPolicyRuleListCore(ac)
	if err != nil {
		return nil, err
	}
	return c.ListPaged(ctx)
}
