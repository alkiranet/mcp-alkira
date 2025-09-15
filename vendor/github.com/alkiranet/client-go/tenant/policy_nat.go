// Copyright (C) 2021-2025 Alkira Inc. All Rights Reserved.

package tenant

import (
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
