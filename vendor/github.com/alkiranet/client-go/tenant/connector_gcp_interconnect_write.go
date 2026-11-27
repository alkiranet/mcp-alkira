// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
)

// ConnectorGcpInterconnectConfig is the write model of a GCP Interconnect
// connector (TPS model/gcpinterconnect/GCPInterconnectConnector.java; GET
// /connectors reports its type as GCP_INTERCONNECT). The fields follow the
// portal body (karrawa gcpInterconnectConfig/addOrEdit/index.js:221-288);
// optional fields are pointers or omitempty slices, so a zero value is never
// sent. The segments are those of the instances' segmentOptions.
type ConnectorGcpInterconnectConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=GCPInterconnectConnector.java:71-75"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,charset=description,src=GCPInterconnectConnector.java:77-79"`
	CXP         string      `json:"cxp" ak:"required,maxlen=128,trimmed,src=GCPInterconnectConnector.java:84-87"`
	Group       *string     `json:"group,omitempty" ak:"maxlen=256,trimmed,charset=name,src=GCPInterconnectConnector.java:92-95"`
	// Size defaults to LARGE in the backend (GCPInterconnectConnector.java:105-108).
	Size        *string `json:"size,omitempty" ak:"enum=SMALL|MEDIUM|LARGE|2LARGE|5LARGE|10LARGE|20LARGE|30LARGE|40LARGE|50LARGE,src=GCPInterconnectConnector.java:105-108|Defaults.java:148-150"`
	BillingTags []int64 `json:"billingTags,omitempty" ak:"flag=BILLING-TAGS"`
	// Enabled defaults to true in the backend (GCPInterconnectConnector.java:115-116).
	Enabled *bool `json:"enabled,omitempty"`
	// LoopbackPrefixes are reserved at save; a prefix cannot be removed once
	// the connector is provisioned (GCPInterconnectConnectorServiceImpl.java:135-148).
	LoopbackPrefixes []string                           `json:"loopbackPrefixes" ak:"required,minitems=1,src=GCPInterconnectConnector.java:121-125"`
	Instances        []ConnectorGcpInterconnectInstance `json:"instances" ak:"required,minitems=1,src=GCPInterconnectConnector.java:129-131"`
	// TunnelProtocol defaults to GRE in the backend
	// (GCPInterconnectConnector.java:137-139) and cannot change once the connector is provisioned
	// (GCPInterconnectConnectorServiceImpl.java:153-155). IPSEC, VXLAN and
	// VXLAN_GPE each need their tenant feature on the CXP (:231-239).
	TunnelProtocol *string `json:"tunnelProtocol,omitempty" ak:"enum=IPSEC|GRE|VXLAN_GPE|VXLAN,src=GCPInterconnectConnector.java:137-139|Defaults.java:1390-1394"`
	ScaleGroupID   *string `json:"scaleGroupId,omitempty" ak:"flag=SUPPORT-SCALE-GROUP-FOR-DX-CONNECTOR"`

	GroupID         *int64          `json:"groupId,omitempty" ak:"readonly"`
	ImplicitGroupID *int64          `json:"implicitGroupId,omitempty" ak:"readonly"`
	Primary         *bool           `json:"primary,omitempty" ak:"readonly"`
	LastEnabled     *int64          `json:"lastEnabled,omitempty" ak:"readonly"`
	ConfigStatus    json.RawMessage `json:"configStatus,omitempty" ak:"readonly"`
	// State and PreviousState come from the GET /connectors row.
	State         string `json:"state,omitempty" ak:"readonly"`
	PreviousState string `json:"previousState,omitempty" ak:"readonly"`
}

// ConnectorGcpInterconnectInstance is one connector instance, a VLAN
// attachment (GCPInterconnectConnectorInstance.java). The BGP key is secret:
// a create can set it, a patch cannot, and an update resends the stored
// value of each instance it keeps. The ASNs are numbers or numeric strings.
type ConnectorGcpInterconnectInstance struct {
	ID   json.Number `json:"id,omitempty"`
	Name string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=GCPInterconnectConnectorInstance.java:36-40"`
	// GCPEdgeAvailabilityDomain cannot change once the connector is
	// provisioned (GCPInterconnectConnectorServiceImpl.java:156-162).
	GCPEdgeAvailabilityDomain *string                                 `json:"gcpEdgeAvailabilityDomain,omitempty" ak:"enum=AVAILABILITY_DOMAIN_1|AVAILABILITY_DOMAIN_2,src=GCPInterconnectConnectorInstance.java:42|Defaults.java:1594-1597"`
	SegmentOptions            []ConnectorGcpInterconnectSegmentOption `json:"segmentOptions" ak:"required,minitems=1,src=GCPInterconnectConnectorInstance.java:46-48"`
	CustomerASN               any                                     `json:"customerAsn" ak:"required,src=GCPInterconnectConnectorInstance.java:50-52"`
	// LocalASN is kept only while RELEASE-EXPLICIT-LOCAL-ASN-FOR-BRANCH-CONNECTOR
	// is on; the backend clears it otherwise (GCPInterconnectConnectorServiceImpl.java:185-192).
	LocalASN         any     `json:"localAsn,omitempty"`
	BGPAuthKeyAlkira *string `json:"bgpAuthKeyAlkira,omitempty" ak:"secret,maxlen=80,trimmed,src=GCPInterconnectConnectorInstance.java:57-60"`
	// GatewayMacAddress is set exactly when tunnelProtocol is VXLAN.
	GatewayMacAddress *string `json:"gatewayMacAddress,omitempty"`
	// VNI is set exactly when tunnelProtocol is VXLAN or VXLAN_GPE.
	VNI *int64 `json:"vni,omitempty" ak:"range=16773023..16777215,src=GCPInterconnectConnectorInstance.java:65-66"`
}

// ConnectorGcpInterconnectSegmentOption is one instance segmentOptions entry
// (GCPInterconnectInstanceSegmentOptions.java).
type ConnectorGcpInterconnectSegmentOption struct {
	SegmentName           string                                    `json:"segmentName" ak:"required,maxlen=256,trimmed,src=GCPInterconnectInstanceSegmentOptions.java:29-32"`
	CustomerGateways      []ConnectorGcpInterconnectCustomerGateway `json:"customerGateways" ak:"required,minitems=1,src=GCPInterconnectInstanceSegmentOptions.java:38-42"`
	DisableInternetExit   *bool                                     `json:"disableInternetExit" ak:"required,src=GCPInterconnectInstanceSegmentOptions.java:44-45"`
	AdvertiseOnPremRoutes *bool                                     `json:"advertiseOnPremRoutes" ak:"required,src=GCPInterconnectInstanceSegmentOptions.java:47-48"`
	SegmentID             *int64                                    `json:"segmentId,omitempty" ak:"readonly"`
}

// ConnectorGcpInterconnectCustomerGateway is one customer gateway
// (GCPInterconnectCustomerGateway.java); tunnelCount defaults to 2, the
// only value the backend takes. LoopbackIP needs
// GCP-INTERCONNECT-CGW-LOOPBACK-IP on; the backend refuses it otherwise
// (GCPInterconnectConnectorServiceImpl.java:169-172).
type ConnectorGcpInterconnectCustomerGateway struct {
	LoopbackIP  *string `json:"loopbackIp,omitempty"`
	TunnelCount *int64  `json:"tunnelCount,omitempty" ak:"range=2..2,src=GCPInterconnectCustomerGateway.java:27-30"`
}

// connectorGcpInterconnectDescriptor: GET, PUT and DELETE answer a missing
// id with 404 AK000003, as every connector does. loopbackPrefixes identifies
// the family in the GET body and GCP_INTERCONNECT in the GET /connectors row,
// which also carries the state the D13 guard reads. Instances merge by id: a
// patch must list every stored instance id, and the stored BGP key of each
// listed instance is resent.
var connectorGcpInterconnectDescriptor = Descriptor{
	Collection:    "gcp-interconnect-connectors",
	FamilyType:    FamilyCheck{Field: "loopbackPrefixes"},
	ConnectorType: "GCP_INTERCONNECT",
	InUse: []Signature{
		{Status: 400, Code: "AK000002", Message: "associated with"},
		{Status: 409, Code: "AK000001", Message: "is still referred externally"},
	},
	Duplicate: []Signature{
		{Status: 400, FieldCode: "constraint.notunique"},
		{Status: 400, FieldCode: "connector.00000003"},
	},
	Provisionable:      true,
	MarkedRowLifecycle: true,
	DocState:           DocStatePlainGet,
	Uniqueness:         Uniqueness{Key: "name", Scope: ScopeConnectorsGroups},
	IDMatchedArrays:    []string{"instances"},
	StateGuard:         StateGuardD13,
	Service:            ServiceTPS,
	CrossCheck:         checkConnectorGcpInterconnect,
	Flags: []FlagRule{
		{Flag: "BILLING-TAGS", Effect: FlagIncludeField, Field: "billingTags", UISource: "src/views/network/provision/objectConfig/gcpInterconnectConfig/addOrEdit/index.js:54|:282"},
		{Flag: "SUPPORT-SCALE-GROUP-FOR-DX-CONNECTOR", Effect: FlagIncludeField, Field: "scaleGroupId", UISource: "src/views/network/provision/objectConfig/gcpInterconnectConfig/addOrEdit/index.js:55|:283-285"},
	},
}

// checkConnectorGcpInterconnect mirrors GCPInterconnectConnectorValidator.java:27-66:
// segment names are unique within an instance and the same in every
// instance; gatewayMacAddress is set exactly for VXLAN; vni exactly for
// VXLAN and VXLAN_GPE. The name must not start with ALK-.
func checkConnectorGcpInterconnect(m map[string]any) error {
	var errs []error
	if name, _ := m["name"].(string); strings.HasPrefix(strings.ToUpper(name), "ALK-") {
		errs = append(errs, errors.New("name must not start with ALK-"))
	}
	proto, _ := m["tunnelProtocol"].(string)
	if proto == "" {
		proto = "GRE"
	}
	instances, _ := m["instances"].([]any)
	var first []string
	for i, e := range instances {
		inst, _ := e.(map[string]any)
		opts, _ := inst["segmentOptions"].([]any)
		var segs []string
		for _, o := range opts {
			so, _ := o.(map[string]any)
			s, _ := so["segmentName"].(string)
			segs = append(segs, s)
		}
		slices.Sort(segs)
		if len(slices.Compact(slices.Clone(segs))) != len(segs) {
			errs = append(errs, fmt.Errorf("instances[%d]: segment names must be unique across segmentOptions", i))
		}
		if i == 0 {
			first = segs
		} else if !slices.Equal(first, segs) {
			errs = append(errs, fmt.Errorf("instances[%d]: every instance must list the same segments in segmentOptions", i))
		}
		mac, _ := inst["gatewayMacAddress"].(string)
		switch {
		case proto != "VXLAN" && mac != "":
			errs = append(errs, fmt.Errorf("instances[%d]: gatewayMacAddress is set only when tunnelProtocol is VXLAN", i))
		case proto == "VXLAN" && mac == "":
			errs = append(errs, fmt.Errorf("instances[%d]: gatewayMacAddress must be set when tunnelProtocol is VXLAN", i))
		}
		vxlan := proto == "VXLAN" || proto == "VXLAN_GPE"
		switch hasVNI := inst["vni"] != nil; {
		case hasVNI && !vxlan:
			errs = append(errs, fmt.Errorf("instances[%d]: vni is set only when tunnelProtocol is VXLAN or VXLAN_GPE", i))
		case !hasVNI && vxlan:
			errs = append(errs, fmt.Errorf("instances[%d]: vni must be set when tunnelProtocol is %s", i, proto))
		}
	}
	return errors.Join(errs...)
}

// ConnectorGcpInterconnectDescriptor returns the GCP Interconnect connector write descriptor.
func ConnectorGcpInterconnectDescriptor() Descriptor {
	return connectorGcpInterconnectDescriptor.clone()
}

// NewConnectorGcpInterconnectCore returns the write core for GCP Interconnect connectors.
func NewConnectorGcpInterconnectCore(ac *AlkiraClient) (*Core[ConnectorGcpInterconnectConfig], error) {
	return NewCore[ConnectorGcpInterconnectConfig](ac, connectorGcpInterconnectDescriptor)
}

// CreateConnectorGcpInterconnect creates a GCP Interconnect connector.
func (ac *AlkiraClient) CreateConnectorGcpInterconnect(ctx context.Context, c *ConnectorGcpInterconnectConfig) (*ConnectorGcpInterconnectConfig, *WriteResult, error) {
	core, err := NewConnectorGcpInterconnectCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, c, WriteOptions{})
}

// GetConnectorGcpInterconnect reads one GCP Interconnect connector and its raw body.
func (ac *AlkiraClient) GetConnectorGcpInterconnect(ctx context.Context, id string) (*ConnectorGcpInterconnectConfig, []byte, error) {
	core, err := NewConnectorGcpInterconnectCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

// UpdateConnectorGcpInterconnect applies a merge patch to a GCP Interconnect connector.
func (ac *AlkiraClient) UpdateConnectorGcpInterconnect(ctx context.Context, id string, patch json.RawMessage) (*ConnectorGcpInterconnectConfig, *WriteResult, error) {
	core, err := NewConnectorGcpInterconnectCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

// DeleteConnectorGcpInterconnect deletes a GCP Interconnect connector; a
// deployed one is marked for deletion until the next provision.
func (ac *AlkiraClient) DeleteConnectorGcpInterconnect(ctx context.Context, id string) (*WriteResult, error) {
	core, err := NewConnectorGcpInterconnectCore(ac)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

// ListConnectorGcpInterconnectPaged reads every GCP Interconnect connector through paged reads.
func (ac *AlkiraClient) ListConnectorGcpInterconnectPaged(ctx context.Context) ([]ConnectorGcpInterconnectConfig, error) {
	core, err := NewConnectorGcpInterconnectCore(ac)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
