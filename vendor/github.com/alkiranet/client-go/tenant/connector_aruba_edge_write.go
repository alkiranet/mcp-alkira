// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ConnectorArubaEdgeConfig is the write model of an Aruba EdgeConnect
// connector (TPS model/ArubaEdgeConnector.java; GET /connectors reports its
// type as ARUBA_EDGE_CONNECT). The fields follow the portal save body
// (karrawa silverPeakConfig/silverPeakComponent/index.js:531-575); optional
// fields are pointers or omitempty slices, so a zero value is never sent.
type ConnectorArubaEdgeConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=ArubaEdgeConnector.java:80-84"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,charset=description,src=ArubaEdgeConnector.java:86-88"`
	CXP         string      `json:"cxp" ak:"required,maxlen=128,trimmed,src=ArubaEdgeConnector.java:72-75"`
	Group       *string     `json:"group,omitempty" ak:"maxlen=256,trimmed,charset=name,src=ArubaEdgeConnector.java:93-96"`
	// Instances are the EdgeConnect appliances; each credentialId names an
	// aruba-edge-connector-instances credential that holds the account key
	// (ArubaEdgeConnectorServiceImpl.java:117-129).
	Instances []ConnectorArubaEdgeInstance `json:"instances" ak:"required,distinct=hostName,src=ArubaEdgeConnector.java:106-109|ArubaEdgeConnectorValidator.java:21-35"`
	Size      string                       `json:"size" ak:"required,enum=XSMALL|5XSMALL|SMALL|MEDIUM|LARGE|2LARGE|4LARGE|5LARGE|10LARGE|20LARGE,src=ArubaEdgeConnector.java:111-113|Defaults.java:148-150"`
	// Version must name an Aruba EdgeConnect image of the CXP, size and boost
	// mode (ArubaEdgeConnectorServiceImpl.java:723-739).
	Version     string   `json:"version" ak:"required,src=ArubaEdgeConnector.java:115-117"`
	BillingTags []int64  `json:"billingTags,omitempty"`
	AllowList   []string `json:"allowList,omitempty" ak:"flag=RELEASE-CONNECTOR-ARUBA-EDGE-CONNECT-ALLOW-LIST"`
	// ArubaEdgeVRFMappings maps EdgeConnect segments to Alkira segment ids;
	// more than one needs version 9 or later
	// (ArubaEdgeConnectorServiceImpl.java:136-142).
	ArubaEdgeVRFMappings []ConnectorArubaEdgeVRFMapping `json:"arubaEdgeVRFMappings" ak:"required,distinct=arubaEdgeConnectSegmentName|alkiraSegmentId,src=ArubaEdgeConnector.java:140-142|ArubaEdgeConnectorServiceImpl.java:660-667"`
	BoostMode            *bool                          `json:"boostMode,omitempty" ak:"required,src=ArubaEdgeConnector.java:144-145"`
	// TunnelProtocol defaults to GRE in the backend, which refuses GRE on an
	// Azure CXP (ArubaEdgeConnectorServiceImpl.java:711-719).
	TunnelProtocol *string `json:"tunnelProtocol,omitempty" ak:"enum=IPSEC|GRE,src=ArubaEdgeConnector.java:147-148|Defaults.java:1399-1401"`
	Enabled        *bool   `json:"enabled,omitempty"`
	ScaleGroupID   *string `json:"scaleGroupId,omitempty" ak:"flag=SCALE-GROUPS-FOR-USER-AND-SITES"`

	GroupID         *int64 `json:"groupId,omitempty" ak:"readonly"`
	ImplicitGroupID *int64 `json:"implicitGroupId,omitempty" ak:"readonly"`
	LastEnabled     *int64 `json:"lastEnabled,omitempty" ak:"readonly"`
	// State and PreviousState come from the GET /connectors row.
	State         string `json:"state,omitempty" ak:"readonly"`
	PreviousState string `json:"previousState,omitempty" ak:"readonly"`
}

// ConnectorArubaEdgeInstance is one EdgeConnect appliance
// (ArubaEdgeConnectorInstance.java:21-45). The backend names it
// <connector>-<hostName> (ArubaEdgeConnectorServiceImpl.java:233). After the
// instance is provisioned accountName, credentialId, hostName and siteTag
// cannot change (ArubaEdgeConnectorServiceImpl.java:265-276).
type ConnectorArubaEdgeInstance struct {
	ID           json.Number `json:"id,omitempty"`
	Name         *string     `json:"name,omitempty" ak:"readonly"`
	SiteTag      string      `json:"siteTag" ak:"required,maxlen=128,trimmed,src=ArubaEdgeConnectorInstance.java:27-30"`
	CredentialID string      `json:"credentialId" ak:"required,trimmed,src=ArubaEdgeConnectorInstance.java:32-34"`
	HostName     string      `json:"hostName" ak:"required,maxlen=32,trimmed,src=ArubaEdgeConnectorInstance.java:36-39"`
	AccountName  string      `json:"accountName" ak:"required,maxlen=128,trimmed,src=ArubaEdgeConnectorInstance.java:41-44"`
}

// ConnectorArubaEdgeVRFMapping is one arubaEdgeVRFMappings entry
// (ArubaEdgeVRFMapping.java:26-47). The mapping cannot change after the
// connector is provisioned (ArubaEdgeConnectorServiceImpl.java:672-687).
type ConnectorArubaEdgeVRFMapping struct {
	ArubaEdgeConnectSegmentName string `json:"arubaEdgeConnectSegmentName" ak:"required,maxlen=63,trimmed,charset=name,src=ArubaEdgeVRFMapping.java:27-31"`
	AlkiraSegmentID             *int64 `json:"alkiraSegmentId,omitempty" ak:"required,src=ArubaEdgeVRFMapping.java:33-34"`
	// GatewayBgpAsn must differ from the segment's ASN
	// (ArubaEdgeConnectorServiceImpl.java:159-162).
	GatewayBgpAsn *int64 `json:"gatewayBgpAsn,omitempty" ak:"required,range=1..4294967295,src=ArubaEdgeVRFMapping.java:36-38"`
	// DisableInternetExit defaults to true, the portal's "internet exit: No"
	// (karrawa silverPeakComponent/index.js:157); the backend default is false (ArubaEdgeVRFMapping.java:40-42).
	DisableInternetExit   *bool `json:"disableInternetExit,omitempty" ak:"default=true"`
	AdvertiseOnPremRoutes *bool `json:"advertiseOnPremRoutes,omitempty"`
}

// connectorArubaEdgeDescriptor: GET, PUT and DELETE answer a missing id with
// 404 AK000003, as every connector does. arubaEdgeVRFMappings identifies the
// family in the GET body and ARUBA_EDGE_CONNECT in the GET /connectors row,
// which also carries the state the D13 guard reads. Instances merge by id:
// the backend keeps an instance whose id the body carries and creates one
// without (ArubaEdgeConnectorServiceImpl.java:256-278). The portal saves
// without ?async=true (karrawa src/_api/sdwan.js:183-205).
var connectorArubaEdgeDescriptor = Descriptor{
	Collection:         "aruba-edge-connectors",
	FamilyType:         FamilyCheck{Field: "arubaEdgeVRFMappings"},
	ConnectorType:      "ARUBA_EDGE_CONNECT",
	InUse:              sdwanInUse,
	Duplicate:          sdwanDuplicate,
	Provisionable:      true,
	MarkedRowLifecycle: true,
	DocState:           DocStatePlainGet,
	Uniqueness:         Uniqueness{Key: "name", Scope: ScopeConnectorsGroups},
	IDMatchedArrays:    []string{"instances"},
	StateGuard:         StateGuardD13,
	Service:            ServiceTPS,
	CrossCheck:         checkConnectorArubaEdge,
	Flags: []FlagRule{
		{Flag: "SCALE-GROUPS-FOR-USER-AND-SITES", Effect: FlagIncludeField, Field: "scaleGroupId", UISource: "src/views/network/provision/objectConfig/silverPeakConfig/silverPeakComponent/index.js:183-185|:544-545"},
		{Flag: "RELEASE-CONNECTOR-ARUBA-EDGE-CONNECT-ALLOW-LIST", Effect: FlagIncludeField, Field: "allowList", UISource: "src/views/network/provision/objectConfig/silverPeakConfig/silverPeakComponent/index.js:187-189|:546-549"},
	},
}

// checkConnectorArubaEdge adds to the name rule: more than one VRF mapping
// needs a version whose major number is 9 or later
// (ArubaEdgeConnectorServiceImpl.java:136-142).
func checkConnectorArubaEdge(m map[string]any) error {
	errs := []error{checkConnectorName(m)}
	maps, _ := m["arubaEdgeVRFMappings"].([]any)
	version, _ := m["version"].(string)
	if major, err := strconv.Atoi(strings.SplitN(version, ".", 2)[0]); len(maps) > 1 && err == nil && major < 9 {
		errs = append(errs, fmt.Errorf("version %s supports one arubaEdgeVRFMappings entry; more need version 9 or later", version))
	}
	return errors.Join(errs...)
}

// ConnectorArubaEdgeDescriptor returns the Aruba EdgeConnect connector write descriptor.
func ConnectorArubaEdgeDescriptor() Descriptor { return connectorArubaEdgeDescriptor.clone() }

// NewConnectorArubaEdgeCore returns the write core for Aruba EdgeConnect connectors.
func NewConnectorArubaEdgeCore(ac *AlkiraClient) (*Core[ConnectorArubaEdgeConfig], error) {
	return NewCore[ConnectorArubaEdgeConfig](ac, connectorArubaEdgeDescriptor)
}

// CreateConnectorArubaEdge creates an Aruba EdgeConnect connector.
func (ac *AlkiraClient) CreateConnectorArubaEdge(ctx context.Context, c *ConnectorArubaEdgeConfig) (*ConnectorArubaEdgeConfig, *WriteResult, error) {
	core, err := NewConnectorArubaEdgeCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, c, WriteOptions{})
}

// GetConnectorArubaEdge reads one Aruba EdgeConnect connector and its raw body.
func (ac *AlkiraClient) GetConnectorArubaEdge(ctx context.Context, id string) (*ConnectorArubaEdgeConfig, []byte, error) {
	core, err := NewConnectorArubaEdgeCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

// UpdateConnectorArubaEdge applies a merge patch to an Aruba EdgeConnect connector.
func (ac *AlkiraClient) UpdateConnectorArubaEdge(ctx context.Context, id string, patch json.RawMessage) (*ConnectorArubaEdgeConfig, *WriteResult, error) {
	core, err := NewConnectorArubaEdgeCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

// DeleteConnectorArubaEdge deletes an Aruba EdgeConnect connector; a
// deployed one is marked for deletion until the next provision.
func (ac *AlkiraClient) DeleteConnectorArubaEdge(ctx context.Context, id string) (*WriteResult, error) {
	core, err := NewConnectorArubaEdgeCore(ac)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

// ListConnectorArubaEdgePaged reads every Aruba EdgeConnect connector through paged reads.
func (ac *AlkiraClient) ListConnectorArubaEdgePaged(ctx context.Context) ([]ConnectorArubaEdgeConfig, error) {
	core, err := NewConnectorArubaEdgeCore(ac)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
