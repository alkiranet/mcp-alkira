// Copyright (C) 2024-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
)

type PeeringGatewayAwsTgwAttachment struct {
	Name             string      `json:"name"`
	Description      string      `json:"description,omitempty"`
	Requestor        string      `json:"requestor"`
	PeerAwsRegion    string      `json:"peerAwsRegion"`
	PeerAwsTgwId     string      `json:"peerAwsTgwId"`
	PeerAwsAccountId string      `json:"peerAwsAccountId"`
	AwsTgwId         int         `json:"awsTgwId"`
	Id               json.Number `json:"id,omitempty"`    // response only
	State            string      `json:"state,omitempty"` // response only
}

// NewConnectorPeeringGatewayAwsTgwAttachment new peering gateway aws tgw attachment
func NewPeeringGatewayAwsTgwAttachment(ac *AlkiraClient) *AlkiraApi[PeeringGatewayAwsTgwAttachment] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/aws-tgw-peering-attachments", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[PeeringGatewayAwsTgwAttachment]{ac, uri, PaginationOn}
	return api
}
