// Copyright (C) 2020-2025 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
)

type ConnectorAwsVpcSummary struct {
	Id              json.Number `json:"id"`
	Name            string      `json:"name"`
	Segments        []string    `json:"segments"`
	Group           string      `json:"group,omitempty"`
	ImplicitGroupId int         `json:"implicitGroupId,omitempty"`
	ScaleGroupId    string      `json:"scaleGroupId,omitempty"`
	Cxp             string      `json:"cxp"`
	VpcId           string      `json:"vpcId"`
}

// NewConnectorAwsVpc new connector-aws-vpc
func NewConnectorAwsVpc(ac *AlkiraClient) *AlkiraApi[ConnectorAwsVpcSummary] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/awsvpcconnectors", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[ConnectorAwsVpcSummary]{ac, uri, PaginationOn}
	return api
}
