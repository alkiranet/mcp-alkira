// Copyright (C) 2020-2025 Alkira Inc. All Rights Reserved.

package tenant

import (
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
