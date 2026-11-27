// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
)

// ConnectorVmwareSdwanConfig is the write model of a VMware SD-WAN connector
// (TPS model/vmwaresdwan/VMWareSdWanConnector.java; GET /connectors reports
// its type as VMWARE_SDWAN). The fields follow the portal save body (karrawa
// vmwareConfig/VMwareComponent.js:338-379); optional fields are pointers or
// omitempty slices, so a zero value is never sent.
type ConnectorVmwareSdwanConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=VMWareSdWanConnector.java:88-92"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,charset=description,src=VMWareSdWanConnector.java:94-96"`
	CXP         string      `json:"cxp" ak:"required,maxlen=128,trimmed,src=VMWareSdWanConnector.java:80-83"`
	Group       *string     `json:"group,omitempty" ak:"maxlen=256,trimmed,charset=name,src=VMWareSdWanConnector.java:101-104"`
	// Instances are the VMware edges; each credentialId names a
	// vmware-sdwan-connector-instance credential holding the activation code
	// (VmWareSdWanConnectorServiceImpl.java:108-121).
	Instances []ConnectorVmwareSdwanInstance `json:"instances" ak:"required,distinct=hostName,src=VMWareSdWanConnector.java:114-117|VMWareSDWANConnectorValidator.java:40-57"`
	Size      string                         `json:"size" ak:"required,enum=XSMALL|5XSMALL|SMALL|MEDIUM|LARGE|2LARGE|4LARGE|5LARGE|10LARGE|20LARGE,src=VMWareSdWanConnector.java:119-121|Defaults.java:148-150"`
	// Version must name a VMware SD-WAN image of the CXP
	// (VmWareSdWanConnectorServiceImpl.java:510-521).
	Version string `json:"version" ak:"required,src=VMWareSdWanConnector.java:123-124"`
	// OrchestratorHostAddress is an http or https URL
	// (VMWareSdWanConnector.java:126-130).
	OrchestratorHostAddress string  `json:"orchestratorHostAddress" ak:"required,maxlen=256,src=VMWareSdWanConnector.java:126-130"`
	BillingTags             []int64 `json:"billingTags,omitempty"`
	// VmWareSdWanVRFMappings holds exactly one mapping
	// (VMWareSDWANConnectorValidator.java:59-64).
	VmWareSdWanVRFMappings []ConnectorVmwareSdwanVRFMapping `json:"vmWareSdWanVRFMappings" ak:"required,maxitems=1,src=VMWareSdWanConnector.java:137-140|VMWareSDWANConnectorValidator.java:59-64"`
	// TunnelProtocol must be IPSEC, and the backend has no default for it
	// (VMWareSDWANConnectorValidator.java:33-38); the portal always sends it
	// (VMwareComponent.js:345).
	TunnelProtocol *string  `json:"tunnelProtocol,omitempty" ak:"default=IPSEC,enum=IPSEC,src=VMwareComponent.js:345|VMWareSDWANConnectorValidator.java:33-38"`
	AllowList      []string `json:"allowList,omitempty" ak:"flag=RELEASE-CONNECTOR-VMWARE-SDWAN-ALLOW-LIST"`
	Enabled        *bool    `json:"enabled,omitempty"`

	GroupID         *int64 `json:"groupId,omitempty" ak:"readonly"`
	ImplicitGroupID *int64 `json:"implicitGroupId,omitempty" ak:"readonly"`
	LastEnabled     *int64 `json:"lastEnabled,omitempty" ak:"readonly"`
	// State and PreviousState come from the GET /connectors row.
	State         string `json:"state,omitempty" ak:"readonly"`
	PreviousState string `json:"previousState,omitempty" ak:"readonly"`
}

// ConnectorVmwareSdwanInstance is one VMware edge
// (VMWareSdWanConnectorInstance.java:21-35). The backend names it
// <connector>-<hostName> (VmWareSdWanConnectorServiceImpl.java:188); after
// the edge is provisioned its credentialId and hostName cannot change
// (:210-223).
type ConnectorVmwareSdwanInstance struct {
	ID           json.Number `json:"id,omitempty"`
	Name         *string     `json:"name,omitempty" ak:"readonly"`
	CredentialID string      `json:"credentialId" ak:"required,trimmed,src=VMWareSdWanConnectorInstance.java:27-29"`
	HostName     string      `json:"hostName" ak:"required,maxlen=32,trimmed,src=VMWareSdWanConnectorInstance.java:31-34"`
}

// ConnectorVmwareSdwanVRFMapping is one vmWareSdWanVRFMappings entry
// (VMWareSdWanVRFMapping.java:25-47).
type ConnectorVmwareSdwanVRFMapping struct {
	VmWareSdWanSegmentName string `json:"vmWareSdWanSegmentName" ak:"required,maxlen=32,trimmed,charset=name,src=VMWareSdWanVRFMapping.java:26-30"`
	SegmentID              *int64 `json:"segmentId,omitempty" ak:"required,src=VMWareSdWanVRFMapping.java:32-33"`
	GatewayBgpAsn          *int64 `json:"gatewayBgpAsn,omitempty" ak:"required,range=1..4294967295,src=VMWareSdWanVRFMapping.java:35-38"`
	// DisableInternetExit defaults to true, the portal's "internet exit: No"
	// (karrawa vmwareConfig/VMwareComponent.js:48); the backend default is false (VMWareSdWanVRFMapping.java:40-42).
	DisableInternetExit   *bool `json:"disableInternetExit,omitempty" ak:"default=true"`
	AdvertiseOnPremRoutes *bool `json:"advertiseOnPremRoutes,omitempty"`
}

// connectorVmwareSdwanDescriptor: GET, PUT and DELETE answer a missing id
// with 404 AK000003, as every connector does. vmWareSdWanVRFMappings
// identifies the family in the GET body and VMWARE_SDWAN in the GET
// /connectors row, which also carries the state the D13 guard reads.
// Instances merge by id. The portal saves without ?async=true (karrawa
// src/_api/vmware.js:3-19). The backend refuses an allowList while
// RELEASE-CONNECTOR-VMWARE-SDWAN-ALLOW-LIST is off (VMWareSdWanConnector.java:243-245).
var connectorVmwareSdwanDescriptor = Descriptor{
	Collection:         "vmware-sdwan-connectors",
	FamilyType:         FamilyCheck{Field: "vmWareSdWanVRFMappings"},
	ConnectorType:      "VMWARE_SDWAN",
	InUse:              sdwanInUse,
	Duplicate:          sdwanDuplicate,
	Provisionable:      true,
	MarkedRowLifecycle: true,
	DocState:           DocStatePlainGet,
	Uniqueness:         Uniqueness{Key: "name", Scope: ScopeConnectorsGroups},
	IDMatchedArrays:    []string{"instances"},
	StateGuard:         StateGuardD13,
	Service:            ServiceTPS,
	CrossCheck:         checkConnectorName,
	Flags: []FlagRule{
		{Flag: "RELEASE-CONNECTOR-VMWARE-SDWAN-ALLOW-LIST", Effect: FlagIncludeField, Field: "allowList", UISource: "src/views/network/provision/objectConfig/vmwareConfig/VMwareComponent.js:145|:377-379"},
	},
}

// ConnectorVmwareSdwanDescriptor returns the VMware SD-WAN connector write descriptor.
func ConnectorVmwareSdwanDescriptor() Descriptor { return connectorVmwareSdwanDescriptor.clone() }

// NewConnectorVmwareSdwanCore returns the write core for VMware SD-WAN connectors.
func NewConnectorVmwareSdwanCore(ac *AlkiraClient) (*Core[ConnectorVmwareSdwanConfig], error) {
	return NewCore[ConnectorVmwareSdwanConfig](ac, connectorVmwareSdwanDescriptor)
}

// CreateConnectorVmwareSdwan creates a VMware SD-WAN connector.
func (ac *AlkiraClient) CreateConnectorVmwareSdwan(ctx context.Context, c *ConnectorVmwareSdwanConfig) (*ConnectorVmwareSdwanConfig, *WriteResult, error) {
	core, err := NewConnectorVmwareSdwanCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, c, WriteOptions{})
}

// GetConnectorVmwareSdwan reads one VMware SD-WAN connector and its raw body.
func (ac *AlkiraClient) GetConnectorVmwareSdwan(ctx context.Context, id string) (*ConnectorVmwareSdwanConfig, []byte, error) {
	core, err := NewConnectorVmwareSdwanCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

// UpdateConnectorVmwareSdwan applies a merge patch to a VMware SD-WAN connector.
func (ac *AlkiraClient) UpdateConnectorVmwareSdwan(ctx context.Context, id string, patch json.RawMessage) (*ConnectorVmwareSdwanConfig, *WriteResult, error) {
	core, err := NewConnectorVmwareSdwanCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

// DeleteConnectorVmwareSdwan deletes a VMware SD-WAN connector; a deployed
// one is marked for deletion until the next provision.
func (ac *AlkiraClient) DeleteConnectorVmwareSdwan(ctx context.Context, id string) (*WriteResult, error) {
	core, err := NewConnectorVmwareSdwanCore(ac)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

// ListConnectorVmwareSdwanPaged reads every VMware SD-WAN connector through paged reads.
func (ac *AlkiraClient) ListConnectorVmwareSdwanPaged(ctx context.Context) ([]ConnectorVmwareSdwanConfig, error) {
	core, err := NewConnectorVmwareSdwanCore(ac)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
