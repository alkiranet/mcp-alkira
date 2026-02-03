// Copyright (C) 2020-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
)

type ServicePANSummary struct {
	Id           json.Number `json:"id"`
	Name         string      `json:"name"`
	InternalName string      `json:"internalName"`
	CXP          string      `json:"cxp"`
	LicenseType  string      `json:"licenseType"`
}

// NewServicePAN new service pan
func NewServicePAN(ac *AlkiraClient) *AlkiraApi[ServicePANSummary] {
	uri := fmt.Sprintf("%s/v1/tenantnetworks/%s/panfwservices", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[ServicePANSummary]{ac, uri, PaginationOn}
	return api
}
