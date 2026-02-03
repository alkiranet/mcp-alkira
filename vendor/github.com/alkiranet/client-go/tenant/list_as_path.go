// Copyright (C) 2021-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
)

type AsPathList struct {
	Description string      `json:"description"`
	Id          json.Number `json:"id,omitempty"`
	Name        string      `json:"name"`
	Values      []string    `json:"values"`
}

func NewListAsPath(ac *AlkiraClient) *AlkiraApi[AsPathList] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/as-path-lists", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[AsPathList]{ac, uri, PaginationOn}
	return api
}
