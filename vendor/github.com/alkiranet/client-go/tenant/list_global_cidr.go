// Copyright (C) 2021-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
)

type GlobalCidrList struct {
	Description string      `json:"description"`
	CXP         string      `json:"cxp"`
	Id          json.Number `json:"id,omitempty"`
	Name        string      `json:"name"`
	Tags        []string    `json:"tags,omitempty"`
	Values      []string    `json:"values"`
}

// NewGlobalCidrList new global cidr list
func NewGlobalCidrList(ac *AlkiraClient) *AlkiraApi[GlobalCidrList] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/global-cidr-lists", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[GlobalCidrList]{ac, uri, PaginationOn}
	return api
}
