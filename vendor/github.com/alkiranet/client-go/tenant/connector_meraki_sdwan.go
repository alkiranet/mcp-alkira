// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
)

type MerakiSdwanSegmentOption struct {
	SegmentId             int  `json:"segmentId"`
	MerakiBgpAsn          int  `json:"merakiBgpAsn,omitempty"`
	AdvertiseOnPremRoutes bool `json:"advertiseOnPremRoutes"`
	DisableInternetExit   bool `json:"disableInternetExit"`
}

type MerakiSdwanInstance struct {
	Id           int    `json:"id,omitempty"`   // response only
	Name         string `json:"name,omitempty"` // response only
	HostName     string `json:"hostName"`
	CredentialId string `json:"credentialId"`
	Az           int    `json:"az,omitempty"`    // response only
	WanIp        string `json:"wanIp,omitempty"` // response only
}

type ConnectorMerakiSdwan struct {
	BillingTags     []int                      `json:"billingTags"`
	Cxp             string                     `json:"cxp"`
	Description     string                     `json:"description,omitempty"`
	Enabled         bool                       `json:"enabled"`
	Group           string                     `json:"group,omitempty"`
	GroupId         int                        `json:"groupId,omitempty"`         // response only
	Id              json.Number                `json:"id,omitempty"`              // response only
	ImplicitGroupId int                        `json:"implicitGroupId,omitempty"` // response only
	Instances       []MerakiSdwanInstance      `json:"instances"`
	Name            string                     `json:"name"`
	SegmentOptions  []MerakiSdwanSegmentOption `json:"segmentOptions"`
	Size            string                     `json:"size"`
	TunnelProtocol  string                     `json:"tunnelProtocol"`
	Version         string                     `json:"version"`
}

// NewConnectorMerakiSdwan new connector
func NewConnectorMerakiSdwan(ac *AlkiraClient) *AlkiraApi[ConnectorMerakiSdwan] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/meraki-sdwan-connectors", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[ConnectorMerakiSdwan]{ac, uri, PaginationOn}
	return api
}
