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

type NatPolicySummary struct {
	Name string      `json:"name"`
	Type string      `json:"type"`
	Id   json.Number `json:"id"`
}

// NewNatPolicy new nat policy
func NewNatPolicy(ac *AlkiraClient) *AlkiraApi[NatPolicy] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/nat-policies", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[NatPolicy]{ac, uri}
	return api
}

// NewNatPoliciesSummary new nat policies summary
func NewNatPolicySummary(ac *AlkiraClient) *AlkiraApi[NatPolicySummary] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/nat-policies", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[NatPolicySummary]{ac, uri}
	return api
}
