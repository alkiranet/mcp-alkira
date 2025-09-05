// Copyright (C) 2020-2025 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
)

type PrefixListRange struct {
	Prefix      string `json:"prefix"`
	Le          int    `json:"le,omitempty"`
	Ge          int    `json:"ge,omitempty"`
	Description string `json:"description,omitempty"`
}

type PrefixListDetails struct {
	Description string `json:"description,omitempty"`
}

type PrefixList struct {
	Description   string                        `json:"description,omitempty"`
	Id            json.Number                   `json:"id,omitempty"`
	Name          string                        `json:"name"`
	Prefixes      []string                      `json:"prefixes"`
	PrefixDetails map[string]*PrefixListDetails `json:"prefixDetails,omitempty"`
	PrefixRanges  []PrefixListRange             `json:"prefixRanges,omitempty"`
	Type          string                        `json:"type,omitempty"`
}

type PrefixListSummary struct {
	Id       json.Number `json:"id"`
	Name     string      `json:"name"`
	Prefixes []string    `json:"prefixes"`
	Type     string      `json:"type,omitempty"`
}

// NewPrefixList new prefix list
func NewPrefixList(ac *AlkiraClient) *AlkiraApi[PrefixListSummary] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/policy/prefixlists", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[PrefixListSummary]{ac, uri, PaginationOff}
	return api
}
