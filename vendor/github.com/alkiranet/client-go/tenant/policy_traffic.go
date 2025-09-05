// Copyright (C) 2020-2025 Alkira Inc. All Rights Reserved.

package tenant

import (
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
