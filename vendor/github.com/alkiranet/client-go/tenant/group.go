// Copyright (C) 2020-2025 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
)

// A summary of a group (other fields are ignored)
type Group struct {
	Id           json.Number `json:"id"`
	Name         string      `json:"name"`
	Description  string      `json:"description,omitempty"`
	InternalName string      `json:"internalName"`
	Type         string      `json:"type"`
}

func NewGroup(ac *AlkiraClient) *AlkiraApi[Group] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/groups", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[Group]{ac, uri, PaginationOn}
	return api
}
