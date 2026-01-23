// Copyright (C) 2021-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
)

type CommunityList struct {
	Description string      `json:"description"`
	Id          json.Number `json:"id,omitempty"`
	Name        string      `json:"name"`
	Values      []string    `json:"values"`
}

func NewListCommunity(ac *AlkiraClient) *AlkiraApi[CommunityList] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/community-lists", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[CommunityList]{ac, uri, PaginationOn}
	return api
}
