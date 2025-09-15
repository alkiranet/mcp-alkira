// Copyright (C) 2021-2025 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
)

type ExtendedCommunityList struct {
	Description string      `json:"description"`
	Id          json.Number `json:"id,omitempty"`
	Name        string      `json:"name"`
	Values      []string    `json:"values"`
}

func NewListExtendedCommunity(ac *AlkiraClient) *AlkiraApi[ExtendedCommunityList] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/extended-community-lists", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[ExtendedCommunityList]{ac, uri, PaginationOn}
	return api
}
