// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
)

// ConnectorAwsTgwConfig is the write model of an AWS TGW connector (TPS
// model/awstgw/AWSTGWConnector.java; GET /connectors reports its type as
// AWS_TGW). The fields follow the portal save body (karrawa
// awsTgwConfig/addOrEdit/index.js:221-233); optional fields are pointers or
// omitempty slices, so a zero value is never sent.
type ConnectorAwsTgwConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,charset=slug,src=AWSTGWConnector.java:64-67"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,charset=description,src=AWSTGWConnector.java:69-71"`
	CXP         string      `json:"cxp" ak:"required,maxlen=128,trimmed,src=AWSTGWConnector.java:56-59"`
	Group       *string     `json:"group,omitempty" ak:"maxlen=256,trimmed,charset=name,src=AWSTGWConnector.java:76-79"`
	// Segments holds exactly one segment name (AWSTGWConnector.java:89-96).
	Segments []string `json:"segments" ak:"required,src=AWSTGWConnector.java:89-96"`
	// Size defaults to SMALL in the backend (AWSTGWConnector.java:102-105).
	Size *string `json:"size,omitempty" ak:"enum=XSMALL|5XSMALL|SMALL|MEDIUM|LARGE|2LARGE|4LARGE|5LARGE|10LARGE|20LARGE,src=AWSTGWConnector.java:102-105|Defaults.java:148-150"`
	// AwsTgwPeeringAttachmentID is an ACTIVE AWS TGW peering attachment of the
	// connector's CXP that no other connector uses
	// (AWSTGWConnectorServiceImpl.java:289-330); it cannot change
	// (AWSTGWConnectorServiceImpl.java:333-337).
	AwsTgwPeeringAttachmentID *int64  `json:"awsTgwPeeringAttachmentId,omitempty" ak:"required,immutable,src=AWSTGWConnector.java:107-108|AWSTGWConnectorServiceImpl.java:333-337"`
	BillingTags               []int64 `json:"billingTags,omitempty" ak:"flag=BILLING-TAGS"`
	// StaticRoutes holds at most one prefix list id (AWSTGWConnector.java:118-122).
	StaticRoutes []int64 `json:"staticRoutes,omitempty"`
	// Enabled is sent as true by the portal unless changed (addOrEdit/index.js:228).
	Enabled *bool `json:"enabled,omitempty" ak:"default=true,src=AWSTGWConnector.java:124-126|addOrEdit/index.js:228"`
	// ScaleGroupID needs RELEASE-SCALE-GROUP-FOR-TGW-CONNECTOR; without it the
	// backend refuses the save with 403 (AWSTGWConnectorServiceImpl.java:382-384).
	ScaleGroupID *string `json:"scaleGroupId,omitempty" ak:"flag=RELEASE-SCALE-GROUP-FOR-TGW-CONNECTOR"`

	GroupID                      *int64          `json:"groupId,omitempty" ak:"readonly"`
	ImplicitGroupID              *int64          `json:"implicitGroupId,omitempty" ak:"readonly"`
	SegmentIDs                   []int64         `json:"segmentIds,omitempty" ak:"readonly"`
	AwsTgwID                     *int64          `json:"awsTgwId,omitempty" ak:"readonly,rederived"`
	TransitGatewaySegmentGroupID *string         `json:"transitGatewaySegmentGroupId,omitempty" ak:"readonly,rederived"`
	StaticRouteValues            []string        `json:"staticRouteValues,omitempty" ak:"readonly,rederived"`
	Primary                      *bool           `json:"primary,omitempty" ak:"readonly"`
	ConfigStatus                 json.RawMessage `json:"configStatus,omitempty" ak:"readonly"`
	// State and PreviousState come from the GET /connectors row.
	State         string `json:"state,omitempty" ak:"readonly"`
	PreviousState string `json:"previousState,omitempty" ak:"readonly"`
}

// connectorAwsTgwDescriptor: GET, PUT and DELETE answer a missing id with
// 404 AK000003, as every connector does. awsTgwPeeringAttachmentId
// identifies the family in the GET body and AWS_TGW in the GET /connectors
// row, which also carries the state the D13 guard reads.
var connectorAwsTgwDescriptor = Descriptor{
	Collection:    "aws-tgw-connectors",
	FamilyType:    FamilyCheck{Field: "awsTgwPeeringAttachmentId"},
	ConnectorType: "AWS_TGW",
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
	StateGuard:         StateGuardD13,
	Service:            ServiceTPS,
	CrossCheck:         checkConnector,
	Flags: []FlagRule{
		{Flag: "BILLING-TAGS", Effect: FlagIncludeField, Field: "billingTags", UISource: "src/views/network/provision/objectConfig/awsTgwConfig/addOrEdit/index.js:60|:229|:556"},
		{Flag: "RELEASE-SCALE-GROUP-FOR-TGW-CONNECTOR", Effect: FlagIncludeField, Field: "scaleGroupId", UISource: "src/views/network/provision/objectConfig/awsTgwConfig/addOrEdit/index.js:47-49|:232"},
	},
}

// ConnectorAwsTgwDescriptor returns the AWS TGW connector write descriptor.
func ConnectorAwsTgwDescriptor() Descriptor { return connectorAwsTgwDescriptor.clone() }

// NewConnectorAwsTgwCore returns the write core for AWS TGW connectors.
func NewConnectorAwsTgwCore(ac *AlkiraClient) (*Core[ConnectorAwsTgwConfig], error) {
	return NewCore[ConnectorAwsTgwConfig](ac, connectorAwsTgwDescriptor)
}

// CreateConnectorAwsTgw creates an AWS TGW connector.
func (ac *AlkiraClient) CreateConnectorAwsTgw(ctx context.Context, c *ConnectorAwsTgwConfig) (*ConnectorAwsTgwConfig, *WriteResult, error) {
	core, err := NewConnectorAwsTgwCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, c, WriteOptions{})
}

// GetConnectorAwsTgw reads one AWS TGW connector and its raw body.
func (ac *AlkiraClient) GetConnectorAwsTgw(ctx context.Context, id string) (*ConnectorAwsTgwConfig, []byte, error) {
	core, err := NewConnectorAwsTgwCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

// UpdateConnectorAwsTgw applies a merge patch to an AWS TGW connector.
func (ac *AlkiraClient) UpdateConnectorAwsTgw(ctx context.Context, id string, patch json.RawMessage) (*ConnectorAwsTgwConfig, *WriteResult, error) {
	core, err := NewConnectorAwsTgwCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

// DeleteConnectorAwsTgw deletes an AWS TGW connector; a deployed one is
// marked for deletion until the next provision.
func (ac *AlkiraClient) DeleteConnectorAwsTgw(ctx context.Context, id string) (*WriteResult, error) {
	core, err := NewConnectorAwsTgwCore(ac)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

// ListConnectorAwsTgwPaged reads every AWS TGW connector through paged reads.
func (ac *AlkiraClient) ListConnectorAwsTgwPaged(ctx context.Context) ([]ConnectorAwsTgwConfig, error) {
	core, err := NewConnectorAwsTgwCore(ac)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
