// Copyright (C) 2020-2025 Alkira Inc. All Rights Reserved.

package tenant

import (
	"fmt"
)

type InventoryCXP struct {
	Id                string             `json:"id"`
	Name              string             `json:"name"`
	Provider          string             `json:"provider"`
	ProviderRegion    string             `json:"providerRegion"`
	State             string             `json:"state"`
	AvailabilityZones map[string]string  `json:"availabilityZones"`
	Geolocation       map[string]float64 `json:"geolocation"`
}

func NewInventoryCXP(ac *AlkiraClient) *AlkiraApi[InventoryCXP] {
	uri := fmt.Sprintf("%s/inventory/cxps", ac.URI)
	api := &AlkiraApi[InventoryCXP]{ac, uri}
	return api
}
