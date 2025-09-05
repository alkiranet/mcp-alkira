// Copyright (C) 2021-2025 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
)

type NatPolicy struct {
	Name                               string      `json:"name"`
	Description                        string      `json:"description"`
	Type                               string      `json:"type"`
	Segment                            string      `json:"segment"`
	IncludedGroups                     []int       `json:"includedGroups"`
	ExcludedGroups                     []int       `json:"excludedGroups"`
	Id                                 json.Number `json:"id,omitempty"`
	NatRuleIds                         []int       `json:"natRuleIds"`
	Category                           string      `json:"category"`
	AllowOverlappingTranslatedPrefixes *bool       `json:"allowOverlappingTranslatedPrefixes"`
}

// A summary of a NAT policy (other fields are ignored)
type NatPolicySummary struct {
	Id         json.Number `json:"id"`
	Name       string      `json:"name"`
	Segment    string      `json:"segment"`
	NatRuleIds []int       `json:"natRuleIds"`
}

// NewNatPolicy new nat policy
func NewNatPolicy(ac *AlkiraClient) *AlkiraApi[NatPolicySummary] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/nat-policies", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[NatPolicySummary]{ac, uri, PaginationOn}
	return api
}
