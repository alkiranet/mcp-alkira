// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
)

// ConnectorAwsVpcConfig is the write model of an AWS VPC connector (TPS
// model/AWSVPCConnector.java; GET /connectors reports its type as AWS_VPC).
// The fields follow the portal create body (karrawa
// awsConfig/add/AddAWSConnector.js:426-470); optional fields are pointers or
// omitempty slices, so a zero value is never sent.
type ConnectorAwsVpcConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,trimmed,src=AWSVPCConnector.java:58-62"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,charset=description,src=AWSVPCConnector.java:64-66"`
	CXP         string      `json:"cxp" ak:"required,maxlen=128,trimmed,src=AWSVPCConnector.java:71-74"`
	// Segments holds exactly one segment name (AWSVPCConnector.java:78-85).
	Segments []string `json:"segments" ak:"required,src=AWSVPCConnector.java:78-85"`
	// VPCID is checked against the live VPC at save time
	// (AWSVPCConnectorServiceImpl.java:130-145) and cannot change
	// (AWSVPCConnectorServiceImpl.java:341).
	VPCID          string  `json:"vpcId" ak:"required,immutable,src=AWSVPCConnector.java:87-88|AWSVPCConnectorServiceImpl.java:341"`
	VPCOwnerID     *string `json:"vpcOwnerId,omitempty"`
	CustomerRegion string  `json:"customerRegion" ak:"required,trimmed,src=AWSVPCConnector.java:95-98"`
	// CustomerName is cleared by an update that leaves it out; the update
	// resends the stored value.
	CustomerName *string `json:"customerName,omitempty" ak:"trimmed,src=AWSVPCConnector.java:92-93"`
	CredentialID string  `json:"credentialId" ak:"required,trimmed,src=AWSVPCConnector.java:100-102"`
	Group        *string `json:"group,omitempty" ak:"maxlen=256,trimmed,charset=name,src=AWSVPCConnector.java:106-109"`
	BillingTags  []int64 `json:"billingTags,omitempty" ak:"flag=BILLING-TAGS"`
	// Size defaults to SMALL in the backend (AWSVPCConnector.java:130-133).
	Size       *string                 `json:"size,omitempty" ak:"enum=XSMALL|5XSMALL|SMALL|MEDIUM|LARGE|2LARGE|5LARGE|10LARGE|20LARGE,src=AWSVPCConnector.java:130-133|Defaults.java:148-150"`
	VPCRouting *ConnectorAwsVpcRouting `json:"vpcRouting,omitempty"`
	// Enabled is sent by the portal only while CXP-FAILOVER-REGION is on; the
	// backend sets an unset value to true (AWSVPCConnectorServiceImpl.java:468).
	Enabled                            *bool                          `json:"enabled,omitempty" ak:"flag=CXP-FAILOVER-REGION"`
	SecondaryCXPs                      []string                       `json:"secondaryCXPs,omitempty"`
	DirectInterVPCCommunicationEnabled *bool                          `json:"directInterVPCCommunicationEnabled,omitempty"`
	DirectInterVPCCommunicationGroup   *string                        `json:"directInterVPCCommunicationGroup,omitempty"`
	TGWAttachments                     []ConnectorAwsVpcTGWAttachment `json:"tgwAttachments,omitempty"`
	// TGWConnectEnabled cannot change once the connector is provisioned
	// (AWSVPCConnectorServiceImpl.java:345); it is refused as immutable.
	TGWConnectEnabled *bool   `json:"tgwConnectEnabled,omitempty" ak:"immutable,src=AWSVPCConnectorServiceImpl.java:345"`
	ScaleGroupID      *string `json:"scaleGroupId,omitempty" ak:"flag=SCALE-GROUPS-FOR-CLOUD-CONNECTORS"`

	GroupID                            *int64          `json:"groupId,omitempty" ak:"readonly"`
	ImplicitGroupID                    *int64          `json:"implicitGroupId,omitempty" ak:"readonly"`
	DirectInterVPCCommunicationGroupID *int64          `json:"directInterVPCCommunicationGroupId,omitempty" ak:"readonly"`
	VGWID                              *string         `json:"vgwId,omitempty" ak:"readonly,rederived"`
	PrimaryConnectorID                 *int64          `json:"primaryConnectorId,omitempty" ak:"readonly,rederived"`
	Primary                            *bool           `json:"primary,omitempty" ak:"readonly"`
	LastEnabled                        *int64          `json:"lastEnabled,omitempty" ak:"readonly"`
	ConfigStatus                       json.RawMessage `json:"configStatus,omitempty" ak:"readonly"`
	// State and PreviousState come from the GET /connectors row.
	State         string `json:"state,omitempty" ak:"readonly"`
	PreviousState string `json:"previousState,omitempty" ak:"readonly"`
}

// ConnectorAwsVpcRouting is AWSVPCConnector.vpcRouting
// (AWSVPCConnector.java:193-304). Left out, the backend builds both branches
// from the live VPC and marks their entries defaultRouting=true.
type ConnectorAwsVpcRouting struct {
	ExportToCXPOptions   *ConnectorAwsVpcExport `json:"exportToCXPOptions,omitempty"`
	ImportFromCXPOptions *ConnectorAwsVpcImport `json:"importFromCXPOptions,omitempty"`
}

// ConnectorAwsVpcExport is vpcRouting.exportToCXPOptions.
type ConnectorAwsVpcExport struct {
	UserInputPrefixes []ConnectorAwsVpcPrefix `json:"userInputPrefixes"`
	RouteExportMode   *string                 `json:"routeExportMode,omitempty" ak:"enum=USER_INPUT_PREFIXES,src=AWSVPCConnector.java:321-325"`
}

// ConnectorAwsVpcPrefix is one exported prefix: a CIDR of the VPC, a subnet
// (ID is the AWS subnet id) or an overlay subnet.
type ConnectorAwsVpcPrefix struct {
	ID             *string `json:"id,omitempty"`
	Value          string  `json:"value" ak:"required,src=AWSVPCConnector.java:281"`
	Type           string  `json:"type" ak:"required,enum=CIDR|SUBNET|OVERLAY_SUBNETS,src=AWSVPCConnector.java:315-319"`
	DefaultRouting *bool   `json:"defaultRouting,omitempty" ak:"readonly"`
}

// ConnectorAwsVpcImport is vpcRouting.importFromCXPOptions.
type ConnectorAwsVpcImport struct {
	RouteTables []ConnectorAwsVpcRouteTable `json:"routeTables"`
}

// ConnectorAwsVpcRouteTable is one VPC route table (ID is the AWS id).
type ConnectorAwsVpcRouteTable struct {
	ID              string  `json:"id" ak:"required,src=AWSVPCConnector.java:296"`
	PrefixListIDs   []int64 `json:"prefixListIds"`
	RouteImportMode string  `json:"routeImportMode" ak:"required,enum=ADVERTISE_DEFAULT_ROUTE|OVERRIDE_DEFAULT_ROUTE|ADVERTISE_CUSTOM_PREFIX,src=AWSVPCConnector.java:329-333"`
	DefaultRouting  *bool   `json:"defaultRouting,omitempty" ak:"readonly"`
}

// ConnectorAwsVpcTGWAttachment is one tgwAttachments entry.
type ConnectorAwsVpcTGWAttachment struct {
	AvailabilityZone string `json:"availabilityZone" ak:"required,src=AWSVPCConnector.java:350"`
	SubnetID         string `json:"subnetId" ak:"required,src=AWSVPCConnector.java:351"`
}

// connectorAwsVpcDescriptor: GET, PUT and DELETE answer a missing id with
// 404 AK000003, as every connector does. vpcId identifies the family in the
// GET body and AWS_VPC in the GET /connectors row, which also carries the
// state the D13 guard reads. An update that leaves vpcRouting out while every
// stored routing entry is backend-built omits it, so the backend rebuilds it
// from the VPC rather than storing it as user input.
var connectorAwsVpcDescriptor = Descriptor{
	Collection:    "awsvpcconnectors",
	FamilyType:    FamilyCheck{Field: "vpcId"},
	ConnectorType: "AWS_VPC",
	InUse: []Signature{
		{Status: 400, Code: "AK000002", Message: "associated with"},
		{Status: 409, Code: "AK000001", Message: "is still referred externally"},
	},
	// awsvpcconnector.00000004: another connector already holds the VPC
	// (AWSVPCConnectorServiceImpl.java:353).
	Duplicate: []Signature{
		{Status: 400, FieldCode: "constraint.notunique"},
		{Status: 400, FieldCode: "connector.00000003"},
		{Status: 400, FieldCode: "awsvpcconnector.00000004"},
	},
	Provisionable:      true,
	MarkedRowLifecycle: true,
	DocState:           DocStatePlainGet,
	Uniqueness:         Uniqueness{Key: "name", Scope: ScopeConnectorsGroups},
	StateGuard:         StateGuardD13,
	RoutingProvenance:  []string{"vpcRouting"},
	Service:            ServiceTPS,
	CrossCheck:         checkConnector,
	Flags: []FlagRule{
		{Flag: "BILLING-TAGS", Effect: FlagIncludeField, Field: "billingTags", UISource: "src/views/network/provision/objectConfig/awsConfig/add/AddAWSConnector.js:53|:443"},
		{Flag: "CXP-FAILOVER-REGION", Effect: FlagIncludeField, Field: "enabled", UISource: "src/views/network/provision/objectConfig/awsConfig/add/AddAWSConnector.js:71|:441"},
		{Flag: "SCALE-GROUPS-FOR-CLOUD-CONNECTORS", Effect: FlagIncludeField, Field: "scaleGroupId", UISource: "src/views/network/provision/objectConfig/awsConfig/add/AddAWSConnector.js:74|:470-472"},
	},
}

// ConnectorAwsVpcDescriptor returns the AWS VPC connector write descriptor.
func ConnectorAwsVpcDescriptor() Descriptor { return connectorAwsVpcDescriptor.clone() }

// NewConnectorAwsVpcCore returns the write core for AWS VPC connectors.
func NewConnectorAwsVpcCore(ac *AlkiraClient) (*Core[ConnectorAwsVpcConfig], error) {
	return NewCore[ConnectorAwsVpcConfig](ac, connectorAwsVpcDescriptor)
}

// CreateConnectorAwsVpc creates an AWS VPC connector.
func (ac *AlkiraClient) CreateConnectorAwsVpc(ctx context.Context, c *ConnectorAwsVpcConfig) (*ConnectorAwsVpcConfig, *WriteResult, error) {
	core, err := NewConnectorAwsVpcCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, c, WriteOptions{})
}

// GetConnectorAwsVpc reads one AWS VPC connector and its raw body.
func (ac *AlkiraClient) GetConnectorAwsVpc(ctx context.Context, id string) (*ConnectorAwsVpcConfig, []byte, error) {
	core, err := NewConnectorAwsVpcCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

// UpdateConnectorAwsVpc applies a merge patch to an AWS VPC connector.
func (ac *AlkiraClient) UpdateConnectorAwsVpc(ctx context.Context, id string, patch json.RawMessage) (*ConnectorAwsVpcConfig, *WriteResult, error) {
	core, err := NewConnectorAwsVpcCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

// DeleteConnectorAwsVpc deletes an AWS VPC connector; a deployed one is
// marked for deletion until the next provision.
func (ac *AlkiraClient) DeleteConnectorAwsVpc(ctx context.Context, id string) (*WriteResult, error) {
	core, err := NewConnectorAwsVpcCore(ac)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

// ListConnectorAwsVpcPaged reads every AWS VPC connector through paged reads.
func (ac *AlkiraClient) ListConnectorAwsVpcPaged(ctx context.Context) ([]ConnectorAwsVpcConfig, error) {
	core, err := NewConnectorAwsVpcCore(ac)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
