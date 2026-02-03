// Copyright (C) 2022-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
)

type CheckpointInstance struct {
	Id             int    `json:"id,omitempty"` // response only
	Name           string `json:"name"`
	InternalName   string `json:"internalName,omitempty"` // response only
	CredentialId   string `json:"credentialId"`
	TrafficEnabled bool   `json:"trafficEnabled"`
}

type CheckpointInstanceConfig struct {
	// The response is string data the entire body of the response
	// whould be interpreted together. There is no json structure.
	Data string
}

type CheckpointManagementServer struct {
	ConfigurationMode string   `json:"configurationMode"`
	CredentialId      string   `json:"credentialId"`
	Domain            string   `json:"domain"`
	GlobalCidrListId  int      `json:"globalCidrListId"`
	Ips               []string `json:"ips"`
	Reachability      string   `json:"reachability"`
	Segment           string   `json:"segment,omitempty"`
	Type              string   `json:"type"`
	UserName          string   `json:"userName"`
}

type ServiceCheckpoint struct {
	AutoScale        string                      `json:"autoScale"`
	BillingTags      []int                       `json:"billingTags"`
	Cxp              string                      `json:"cxp"`
	CredentialId     string                      `json:"credentialId"`
	Description      string                      `json:"description"`
	Id               json.Number                 `json:"id,omitempty"` // response only
	Instances        []CheckpointInstance        `json:"instances"`
	InternalName     string                      `json:"internalName"`
	LicenseType      string                      `json:"licenseType"`
	ManagementServer *CheckpointManagementServer `json:"managementServer"`
	MaxInstanceCount int                         `json:"maxInstanceCount"`
	MinInstanceCount int                         `json:"minInstanceCount"`
	Name             string                      `json:"name"`
	PdpIps           []string                    `json:"pdpIps,omitempty"`
	Segments         []string                    `json:"segments"`
	SegmentOptions   SegmentNameToZone           `json:"segmentOptions"`
	Size             string                      `json:"size"`
	TunnelProtocol   string                      `json:"tunnelProtocol"`
	Version          string                      `json:"version"`
}

type ServiceCheckpointSummary struct {
	BillingTags  []int       `json:"billingTags"`
	Cxp          string      `json:"cxp"`
	Id           json.Number `json:"id,omitempty"`
	InternalName string      `json:"internalName"`
	LicenseType  string      `json:"licenseType"`
	Name         string      `json:"name"`
	Segments     []string    `json:"segments"`
	Version      string      `json:"version"`
}

// NewServiceCheckpoint new service checkpoint
func NewServiceCheckpoint(ac *AlkiraClient) *AlkiraApi[ServiceCheckpointSummary] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/chkp-fw-services", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[ServiceCheckpointSummary]{ac, uri, PaginationOn}
	return api
}
