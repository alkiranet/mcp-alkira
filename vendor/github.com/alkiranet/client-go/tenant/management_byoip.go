// Copyright (C) 2022-2025 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
)

type ByoipExtraAttributes struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
	PublicKey string `json:"publicKey"`
}

type Byoip struct {
	ExtraAttributes ByoipExtraAttributes `json:"extraAttributes"`
	CloudProvider   string               `json:"cloudProvider"`
	Prefix          string               `json:"prefix"`
	Cxp             string               `json:"cxp"`
	Description     string               `json:"description"`
	Id              json.Number          `json:"id,omitempty"`
	DoNotAdvertise  bool                 `json:"doNotAdvertise"`
}

// NewByoip new BYOIP
func NewByoip(ac *AlkiraClient) *AlkiraApi[Byoip] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/byoips", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[Byoip]{ac, uri}

	return api
}
