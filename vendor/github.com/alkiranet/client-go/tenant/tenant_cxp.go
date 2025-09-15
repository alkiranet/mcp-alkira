// Copyright (C) 2020-2025 Alkira Inc. All Rights Reserved.

package tenant

import (
	"fmt"
)

type CXP struct {
	Id                string            `json:"id"`
	Name              string            `json:"name"`
	Provider          string            `json:"provider"`
	ProviderRegion    string            `json:"providerRegion"`
	State             string            `json:"state"`
	AvailabilityZones map[string]string `json:"availabilityZones,omitempty"`
}

func NewCXP(ac *AlkiraClient) *AlkiraApi[CXP] {
	uri := fmt.Sprintf("%s/inventory/cxps", ac.URI)
	api := &AlkiraApi[CXP]{ac, uri, PaginationOff}
	return api
}
