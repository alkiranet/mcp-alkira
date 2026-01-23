// Copyright (C) 2022-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
)

type SegmentResource struct {
	Id            json.Number                  `json:"id"`
	Name          string                       `json:"name"`
	Description   string                       `json:"description,omitempty"`
	Segment       string                       `json:"segment"`
	GroupId       int                          `json:"groupId,omitempty"` // response only
	GroupPrefixes []SegmentResourceGroupPrefix `json:"groupPrefixes"`
}

type SegmentResourceGroupPrefix struct {
	GroupId      int `json:"groupId"`
	PrefixListId int `json:"prefixListId"`
}

// NewSegmentResource new segment resource
func NewSegmentResource(ac *AlkiraClient) *AlkiraApi[SegmentResource] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/segment-resources", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[SegmentResource]{ac, uri, PaginationOn}
	return api
}
