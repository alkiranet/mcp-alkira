// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// ConnectorFortinetSdwanConfig is the write model of a Fortinet SD-WAN
// connector (TPS model/fortinetsdwan/FortinetSDWANConnector.java; GET
// /connectors reports its type as FORTINET_SDWAN). The fields follow the
// portal save body (karrawa fortinetSdwanConfig/add/utils.js:91-166);
// optional fields are pointers or omitempty slices, so a zero value is never
// sent.
type ConnectorFortinetSdwanConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=FortinetSDWANConnector.java:73-77"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,charset=description,src=FortinetSDWANConnector.java:79-81"`
	CXP         string      `json:"cxp" ak:"required,maxlen=128,trimmed,src=FortinetSDWANConnector.java:83-86"`
	Group       *string     `json:"group,omitempty" ak:"maxlen=256,trimmed,charset=name,src=FortinetSDWANConnector.java:88-91"`
	// Instances are the FortiGate edges. The backend matches them by id: an
	// edge left out is removed with its credential, and an edge without id is
	// new (RT fortinet-sdwan/PHASE-A-RESULTS.md:20-21).
	Instances []ConnectorFortinetSdwanEdge `json:"instances" ak:"required,src=FortinetSDWANConnector.java:101-104"`
	Size      string                       `json:"size" ak:"required,enum=SMALL|MEDIUM|LARGE|2LARGE,src=FortinetSDWANConnector.java:106-108|Defaults.java:148-150"`
	// AllowList entries must not overlap (FortinetSDWANConnectorValidator.java:69-132).
	AllowList            []string                           `json:"allowList,omitempty"`
	BillingTags          []int64                            `json:"billingTags,omitempty"`
	FtntSDWANVRFMappings []ConnectorFortinetSdwanVRFMapping `json:"ftntSDWANVRFMappings" ak:"required,src=FortinetSDWANConnector.java:119-121"`
	// TunnelProtocol defaults to IPSEC; GRE is refused on an Azure CXP
	// (FortinetSDWANConnectorValidator.java:51-67).
	TunnelProtocol *string `json:"tunnelProtocol,omitempty" ak:"enum=IPSEC|GRE,src=FortinetSDWANConnector.java:123-124|FortinetSDWANConnectorValidator.java:54-60"`
	Enabled        *bool   `json:"enabled,omitempty"`

	GroupID         *int64 `json:"groupId,omitempty" ak:"readonly"`
	ImplicitGroupID *int64 `json:"implicitGroupId,omitempty" ak:"readonly"`
	// State and PreviousState come from the GET /connectors row.
	State         string `json:"state,omitempty" ak:"readonly"`
	PreviousState string `json:"previousState,omitempty" ak:"readonly"`
}

// ConnectorFortinetSdwanEdge is one FortiGate edge
// (FortinetSDWANConnectorInstance.java:24-51). The backend names it
// <connector>-<hostName>; a renamed edge is a new edge with a new id, and the
// edge credential is single-use: the backend deletes it with its edge
// (RT fortinet-sdwan/PHASE-A-RESULTS.md:20-21, PHASE-D-RESULTS.md:9-10). The
// backend drops serialNumber on a PAY_AS_YOU_GO edge
// (RT BACKEND-FOLLOWUPS.md:310-313).
type ConnectorFortinetSdwanEdge struct {
	ID           json.Number `json:"id,omitempty"`
	Name         *string     `json:"name,omitempty" ak:"readonly"`
	BucketNameID *string     `json:"bucketNameId,omitempty" ak:"readonly"`
	// LicenseType defaults to PAY_AS_YOU_GO and must match the licenseType the
	// credential holds (FortinetSDWANConnectorServiceImpl.java:439-451).
	LicenseType  *string `json:"licenseType,omitempty" ak:"enum=BRING_YOUR_OWN|PAY_AS_YOU_GO,src=FortinetSDWANConnectorInstance.java:34-35|Defaults.java:682-684"`
	SerialNumber *string `json:"serialNumber,omitempty" ak:"charset=alnum,src=FortinetSDWANConnectorInstance.java:37-39"`
	// CredentialID names a ftnt-sdwan-connector-instance credential holding
	// userName, password, licenseType and, for BRING_YOUR_OWN, licenseKey
	// (FortinetSDWANConnectorServiceImpl.java:411-461).
	CredentialID string `json:"credentialId" ak:"required,src=FortinetSDWANConnectorInstance.java:41-42"`
	HostName     string `json:"hostName" ak:"required,maxlen=24,trimmed,src=FortinetSDWANConnectorInstance.java:44-47"`
	// Version must name a FortiGate image of the CXP, size and licence
	// (FortinetSDWANConnectorServiceImpl.java:464-476).
	Version string `json:"version" ak:"required,src=FortinetSDWANConnectorInstance.java:49-50"`
}

// ConnectorFortinetSdwanVRFMapping is one ftntSDWANVRFMappings entry
// (FortinetSDWANVRFMapping.java:24-47).
type ConnectorFortinetSdwanVRFMapping struct {
	Vrf       *int64 `json:"vrf,omitempty" ak:"required,range=1..31,src=FortinetSDWANVRFMapping.java:27-30"`
	SegmentID *int64 `json:"segmentId,omitempty" ak:"required,src=FortinetSDWANVRFMapping.java:32-33"`
	// GatewayBgpAsn must differ from the segment's ASN
	// (FortinetSDWANConnectorServiceImpl.java:505-512).
	GatewayBgpAsn *int64 `json:"gatewayBgpAsn,omitempty" ak:"required,range=1..4294967295,src=FortinetSDWANVRFMapping.java:35-38"`
	// DisableInternetExit defaults to true, the portal's "internet exit: No"
	// (karrawa fortinetSdwanConfig/add/utils.js:85); the backend default is false (FortinetSDWANVRFMapping.java:40-42).
	DisableInternetExit   *bool `json:"disableInternetExit,omitempty" ak:"default=true"`
	AdvertiseOnPremRoutes *bool `json:"advertiseOnPremRoutes,omitempty"`
}

// connectorFortinetSdwanDescriptor: GET, PUT and DELETE answer a missing id
// with 404 AK000003, as every connector does. ftntSDWANVRFMappings
// identifies the family in the GET body and FORTINET_SDWAN in the GET
// /connectors row, which also carries the state the D13 guard reads.
// Instances merge by id. The portal saves without ?async=true (karrawa
// src/_api/fortinetSdwan.js:15-25).
var connectorFortinetSdwanDescriptor = Descriptor{
	Collection:         "ftnt-sdwan-connectors",
	FamilyType:         FamilyCheck{Field: "ftntSDWANVRFMappings"},
	ConnectorType:      "FORTINET_SDWAN",
	InUse:              sdwanInUse,
	Duplicate:          sdwanDuplicate,
	Provisionable:      true,
	MarkedRowLifecycle: true,
	DocState:           DocStatePlainGet,
	Uniqueness:         Uniqueness{Key: "name", Scope: ScopeConnectorsGroups},
	IDMatchedArrays:    []string{"instances"},
	StateGuard:         StateGuardD13,
	Service:            ServiceTPS,
	CrossCheck:         checkConnectorFortinetSdwan,
}

// checkConnectorFortinetSdwan adds to the name rule: a BRING_YOUR_OWN edge
// needs a serialNumber (FortinetSDWANConnectorValidator.java:34-49).
func checkConnectorFortinetSdwan(m map[string]any) error {
	errs := []error{checkConnectorName(m)}
	instances, _ := m["instances"].([]any)
	for _, e := range instances {
		in, _ := e.(map[string]any)
		if sn, _ := in["serialNumber"].(string); in["licenseType"] == "BRING_YOUR_OWN" && sn == "" {
			errs = append(errs, fmt.Errorf("instance %v: serialNumber must be set when licenseType is BRING_YOUR_OWN", in["hostName"]))
		}
	}
	return errors.Join(errs...)
}

// ConnectorFortinetSdwanDescriptor returns the Fortinet SD-WAN connector write descriptor.
func ConnectorFortinetSdwanDescriptor() Descriptor { return connectorFortinetSdwanDescriptor.clone() }

// NewConnectorFortinetSdwanCore returns the write core for Fortinet SD-WAN connectors.
func NewConnectorFortinetSdwanCore(ac *AlkiraClient) (*Core[ConnectorFortinetSdwanConfig], error) {
	return NewCore[ConnectorFortinetSdwanConfig](ac, connectorFortinetSdwanDescriptor)
}

// CreateConnectorFortinetSdwan creates a Fortinet SD-WAN connector.
func (ac *AlkiraClient) CreateConnectorFortinetSdwan(ctx context.Context, c *ConnectorFortinetSdwanConfig) (*ConnectorFortinetSdwanConfig, *WriteResult, error) {
	core, err := NewConnectorFortinetSdwanCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, c, WriteOptions{})
}

// GetConnectorFortinetSdwan reads one Fortinet SD-WAN connector and its raw body.
func (ac *AlkiraClient) GetConnectorFortinetSdwan(ctx context.Context, id string) (*ConnectorFortinetSdwanConfig, []byte, error) {
	core, err := NewConnectorFortinetSdwanCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

// UpdateConnectorFortinetSdwan applies a merge patch to a Fortinet SD-WAN connector.
func (ac *AlkiraClient) UpdateConnectorFortinetSdwan(ctx context.Context, id string, patch json.RawMessage) (*ConnectorFortinetSdwanConfig, *WriteResult, error) {
	core, err := NewConnectorFortinetSdwanCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

// DeleteConnectorFortinetSdwan deletes a Fortinet SD-WAN connector; a
// deployed one is marked for deletion until the next provision.
func (ac *AlkiraClient) DeleteConnectorFortinetSdwan(ctx context.Context, id string) (*WriteResult, error) {
	core, err := NewConnectorFortinetSdwanCore(ac)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

// ListConnectorFortinetSdwanPaged reads every Fortinet SD-WAN connector through paged reads.
func (ac *AlkiraClient) ListConnectorFortinetSdwanPaged(ctx context.Context) ([]ConnectorFortinetSdwanConfig, error) {
	core, err := NewConnectorFortinetSdwanCore(ac)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
