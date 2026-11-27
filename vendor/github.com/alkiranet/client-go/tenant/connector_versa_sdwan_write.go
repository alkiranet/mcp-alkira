// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
)

// ConnectorVersaSdwanConfig is the write model of a Versa SD-WAN connector
// (TPS model/versasdwan/VersaSDWANConnector.java; GET /connectors reports its
// type as VERSA_SDWAN). The fields follow the portal save body (karrawa
// versaSDWANConfig/utils.js:66-99); optional fields are pointers or
// omitempty slices, so a zero value is never sent. The connector references
// no credential (VersaSDWANConnector.java:170-172).
type ConnectorVersaSdwanConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=VersaSDWANConnector.java:79-83"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,charset=description,src=VersaSDWANConnector.java:85-87"`
	CXP         string      `json:"cxp" ak:"required,maxlen=128,trimmed,src=VersaSDWANConnector.java:89-92"`
	Group       *string     `json:"group,omitempty" ak:"maxlen=256,trimmed,charset=name,src=VersaSDWANConnector.java:94-97"`
	// RemoteID is an email address (VersaSDWANConnector.java:107-109).
	RemoteID string `json:"remoteId" ak:"required,src=VersaSDWANConnector.java:107-109"`
	// The shared keys are write-only and required on create: a key left out
	// or sent as null is stored as the backend's fixed default "1234"
	// (VersaSDWANConnector.java:111-120; verified on tenant 239 for both),
	// which is also what the portal's optional key fields produce when left
	// blank (versaSDWANConfig/addOrEdit/VersaKeysContainer.js:34-46,
	// utils.js:72-78). A patch cannot set them and an update resends the
	// stored values.
	RemotePublicSharedKey *string `json:"remotePublicSharedKey,omitempty" ak:"required,writeonly,maxlen=64,src=VersaSDWANConnector.java:111-113"`
	LocalID               string  `json:"localId" ak:"required,src=VersaSDWANConnector.java:115-116"`
	LocalPublicSharedKey  *string `json:"localPublicSharedKey,omitempty" ak:"required,writeonly,maxlen=64,src=VersaSDWANConnector.java:118-120"`
	// GlobalTenantID defaults to 1 in the backend (VersaSDWANConnector.java:122-123).
	GlobalTenantID *int64 `json:"globalTenantId,omitempty"`
	// VersaControllerHost is required, as in the portal
	// (versaSDWANConfig/addOrEdit/VersaNameAttributes.js:42).
	VersaControllerHost   *string                         `json:"versaControllerHost,omitempty" ak:"required,trimmed,src=VersaNameAttributes.js:42|VersaSDWANConnector.java:123"`
	Instances             []ConnectorVersaSdwanInstance   `json:"instances" ak:"required,src=VersaSDWANConnector.java:127-130"`
	Size                  string                          `json:"size" ak:"required,enum=SMALL|MEDIUM|LARGE|2LARGE|5LARGE,src=VersaSDWANConnector.java:132-134|Defaults.java:148-150"`
	BillingTags           []int64                         `json:"billingTags,omitempty" ak:"flag=BILLING-TAGS"`
	VersaSDWANVRFMappings []ConnectorVersaSdwanVRFMapping `json:"versaSDWANVRFMappings" ak:"required,src=VersaSDWANConnector.java:141-143"`
	// TunnelProtocol defaults to IPSEC, the only protocol the backend accepts
	// for Versa on AWS and Azure (VersaSDWANConnectorValidator.java:27-37).
	TunnelProtocol *string  `json:"tunnelProtocol,omitempty" ak:"enum=IPSEC,src=VersaSDWANConnector.java:145-146|VersaSDWANConnectorValidator.java:27-37"`
	AllowList      []string `json:"allowList,omitempty" ak:"flag=RELEASE-CONNECTOR-VERSA-SDWAN-ALLOW-LIST"`
	Enabled        *bool    `json:"enabled,omitempty"`

	GroupID         *int64 `json:"groupId,omitempty" ak:"readonly"`
	ImplicitGroupID *int64 `json:"implicitGroupId,omitempty" ak:"readonly"`
	// State and PreviousState come from the GET /connectors row.
	State         string `json:"state,omitempty" ak:"readonly"`
	PreviousState string `json:"previousState,omitempty" ak:"readonly"`
}

// ConnectorVersaSdwanInstance is one Versa edge
// (VersaSDWANConnectorInstance.java:21-38); the backend names it
// <connector>-<hostName> (VersaSDWANConnectorServiceImpl.java:152).
type ConnectorVersaSdwanInstance struct {
	ID           json.Number `json:"id,omitempty"`
	Name         *string     `json:"name,omitempty" ak:"readonly"`
	SerialNumber *string     `json:"serialNumber,omitempty" ak:"charset=hyphen,src=VersaSDWANConnectorInstance.java:28-29"`
	HostName     string      `json:"hostName" ak:"required,maxlen=32,trimmed,src=VersaSDWANConnectorInstance.java:31-34"`
	// Version must name a Versa image of the CXP and size
	// (VersaSDWANConnectorServiceImpl.java:77-89).
	Version string `json:"version" ak:"required,src=VersaSDWANConnectorInstance.java:36-37"`
}

// ConnectorVersaSdwanVRFMapping is one versaSDWANVRFMappings entry
// (VersaSDWANVRFMapping.java:27-50).
type ConnectorVersaSdwanVRFMapping struct {
	VrfName       string `json:"vrfName" ak:"required,maxlen=32,trimmed,charset=name,src=VersaSDWANVRFMapping.java:29-33"`
	SegmentID     *int64 `json:"segmentId,omitempty" ak:"required,src=VersaSDWANVRFMapping.java:35-36"`
	GatewayBgpAsn *int64 `json:"gatewayBgpAsn,omitempty" ak:"required,range=1..4294967295,src=VersaSDWANVRFMapping.java:38-41"`
	// DisableInternetExit defaults to true, the portal's "internet exit: No"
	// (karrawa versaSDWANConfig/addOrEdit/reducer.js:35); the backend default is false (VersaSDWANVRFMapping.java:43-45).
	DisableInternetExit   *bool `json:"disableInternetExit,omitempty" ak:"default=true"`
	AdvertiseOnPremRoutes *bool `json:"advertiseOnPremRoutes,omitempty"`
}

// connectorVersaSdwanDescriptor: GET, PUT and DELETE answer a missing id
// with 404 AK000003, as every connector does. versaSDWANVRFMappings
// identifies the family in the GET body and VERSA_SDWAN in the GET
// /connectors row, which also carries the state the D13 guard reads.
// Instances merge by id. The portal saves without ?async=true (karrawa
// src/_api/versa.js:3-19).
var connectorVersaSdwanDescriptor = Descriptor{
	Collection:         "versa-sdwan-connectors",
	FamilyType:         FamilyCheck{Field: "versaSDWANVRFMappings"},
	ConnectorType:      "VERSA_SDWAN",
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
		{Flag: "BILLING-TAGS", Effect: FlagIncludeField, Field: "billingTags", UISource: "src/views/network/provision/objectConfig/versaSDWANConfig/addOrEdit/index.js:186-188|src/views/network/provision/objectConfig/versaSDWANConfig/utils.js:95"},
		{Flag: "RELEASE-CONNECTOR-VERSA-SDWAN-ALLOW-LIST", Effect: FlagIncludeField, Field: "allowList", UISource: "src/views/network/provision/objectConfig/versaSDWANConfig/addOrEdit/index.js:47-49|src/views/network/provision/objectConfig/versaSDWANConfig/utils.js:61-64"},
	},
}

// ConnectorVersaSdwanDescriptor returns the Versa SD-WAN connector write descriptor.
func ConnectorVersaSdwanDescriptor() Descriptor { return connectorVersaSdwanDescriptor.clone() }

// NewConnectorVersaSdwanCore returns the write core for Versa SD-WAN connectors.
func NewConnectorVersaSdwanCore(ac *AlkiraClient) (*Core[ConnectorVersaSdwanConfig], error) {
	return NewCore[ConnectorVersaSdwanConfig](ac, connectorVersaSdwanDescriptor)
}

// CreateConnectorVersaSdwan creates a Versa SD-WAN connector.
func (ac *AlkiraClient) CreateConnectorVersaSdwan(ctx context.Context, c *ConnectorVersaSdwanConfig) (*ConnectorVersaSdwanConfig, *WriteResult, error) {
	core, err := NewConnectorVersaSdwanCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, c, WriteOptions{})
}

// GetConnectorVersaSdwan reads one Versa SD-WAN connector and its raw body.
func (ac *AlkiraClient) GetConnectorVersaSdwan(ctx context.Context, id string) (*ConnectorVersaSdwanConfig, []byte, error) {
	core, err := NewConnectorVersaSdwanCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

// UpdateConnectorVersaSdwan applies a merge patch to a Versa SD-WAN connector.
func (ac *AlkiraClient) UpdateConnectorVersaSdwan(ctx context.Context, id string, patch json.RawMessage) (*ConnectorVersaSdwanConfig, *WriteResult, error) {
	core, err := NewConnectorVersaSdwanCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

// DeleteConnectorVersaSdwan deletes a Versa SD-WAN connector; a deployed one
// is marked for deletion until the next provision.
func (ac *AlkiraClient) DeleteConnectorVersaSdwan(ctx context.Context, id string) (*WriteResult, error) {
	core, err := NewConnectorVersaSdwanCore(ac)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

// ListConnectorVersaSdwanPaged reads every Versa SD-WAN connector through paged reads.
func (ac *AlkiraClient) ListConnectorVersaSdwanPaged(ctx context.Context) ([]ConnectorVersaSdwanConfig, error) {
	core, err := NewConnectorVersaSdwanCore(ac)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
