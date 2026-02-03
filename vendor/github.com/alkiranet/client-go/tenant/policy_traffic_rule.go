// Copyright (C) 2020-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
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
