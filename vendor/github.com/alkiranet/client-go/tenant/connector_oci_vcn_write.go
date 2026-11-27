// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
)

// ConnectorOciVcnConfig is the write model of an OCI VCN connector (TPS
// model/oci/OCIVCNConnector.java; GET /connectors reports its type as
// OCI_VCN). The fields follow the portal create body (karrawa
// ociConfig/add/AddOCIConnector.js:424-444); optional fields are pointers or
// omitempty slices, so a zero value is never sent. The backend reads the VCN
// from OCI on every save (OCIVCNConnectorServiceImpl.java:116-121), so a save
// needs a live VCN and an OCI credential that can read it.
type ConnectorOciVcnConfig struct {
	ID json.Number `json:"id,omitempty" ak:"readonly"`
	// Name also takes "/" (OCIVCNConnector.java:63), which charset=name does not.
	Name        string  `json:"name" ak:"required,maxlen=256,trimmed,src=OCIVCNConnector.java:60-64"`
	Description *string `json:"description,omitempty" ak:"maxlen=512,charset=description,src=OCIVCNConnector.java:66-68"`
	CXP         string  `json:"cxp" ak:"required,maxlen=128,trimmed,src=OCIVCNConnector.java:73-76"`
	// Segments holds exactly one segment name (OCIVCNConnector.java:80-87).
	Segments []string `json:"segments" ak:"required,src=OCIVCNConnector.java:80-87"`
	// VCNID is the OCI id of the VCN; one connector per VCN in a network
	// (OCIVCNConnectorServiceImpl.java:231-235).
	VCNID          string  `json:"vcnId" ak:"required,src=OCIVCNConnector.java:89-90"`
	CustomerRegion string  `json:"customerRegion" ak:"required,trimmed,src=OCIVCNConnector.java:92-94"`
	CredentialID   string  `json:"credentialId" ak:"required,trimmed,src=OCIVCNConnector.java:96-98"`
	Group          *string `json:"group,omitempty" ak:"maxlen=256,trimmed,charset=name,src=OCIVCNConnector.java:102-105"`
	BillingTags    []int64 `json:"billingTags,omitempty" ak:"flag=BILLING-TAGS"`
	// Size defaults to SMALL in the backend (OCIVCNConnector.java:126-129).
	Size *string `json:"size,omitempty" ak:"enum=XSMALL|5XSMALL|SMALL|MEDIUM|LARGE,src=OCIVCNConnector.java:126-129|Defaults.java:148-150"`
	// VCNRouting left out makes the backend build both branches from the
	// live VCN and mark their entries defaultRouting=true
	// (OCIVCNConnectorServiceImpl.java:123-130, :385-408).
	VCNRouting    *ConnectorOciVcnRoutingOptions `json:"vcnRouting,omitempty"`
	Enabled       *bool                          `json:"enabled,omitempty"`
	SecondaryCXPs []string                       `json:"secondaryCXPs,omitempty"`

	GroupID         *int64  `json:"groupId,omitempty" ak:"readonly"`
	ImplicitGroupID *int64  `json:"implicitGroupId,omitempty" ak:"readonly"`
	DRGID           *string `json:"drgId,omitempty" ak:"readonly"`
	// Primary is set by the backend; the portal does not send it.
	Primary            *bool           `json:"primary,omitempty" ak:"readonly"`
	LastEnabled        *int64          `json:"lastEnabled,omitempty" ak:"readonly"`
	PrimaryConnectorID *int64          `json:"primaryConnectorId,omitempty" ak:"readonly,rederived"`
	VCNCompartmentName *string         `json:"vcnCompartmentName,omitempty" ak:"readonly,rederived"`
	ConfigStatus       json.RawMessage `json:"configStatus,omitempty" ak:"readonly"`
	// State and PreviousState come from the GET /connectors row.
	State         string `json:"state,omitempty" ak:"readonly"`
	PreviousState string `json:"previousState,omitempty" ak:"readonly"`
}

// ConnectorOciVcnRoutingOptions is OCIVCNConnector.vcnRouting (TPS model/cloud/Routing.java).
type ConnectorOciVcnRoutingOptions struct {
	ExportToCXPOptions   *ConnectorOciVcnExport `json:"exportToCXPOptions,omitempty"`
	ImportFromCXPOptions *ConnectorOciVcnImport `json:"importFromCXPOptions,omitempty"`
}

// ConnectorOciVcnExport is vcnRouting.exportToCXPOptions (ExportToCXPOptions.java).
type ConnectorOciVcnExport struct {
	UserInputPrefixes []ConnectorOciVcnPrefix `json:"userInputPrefixes,omitempty"`
	RouteExportMode   *string                 `json:"routeExportMode,omitempty" ak:"enum=USER_INPUT_PREFIXES,src=ExportToCXPOptions.java:27-28"`
}

// ConnectorOciVcnPrefix is one exported prefix: a CIDR of the VCN or a
// subnet, whose ID is the OCI subnet id (UserInputPrefix.java).
type ConnectorOciVcnPrefix struct {
	ID             *string `json:"id,omitempty"`
	Value          string  `json:"value" ak:"required,src=UserInputPrefix.java:26"`
	Type           string  `json:"type" ak:"required,enum=CIDR|SUBNET,src=UserInputPrefix.java:59-61"`
	DefaultRouting *bool   `json:"defaultRouting,omitempty" ak:"readonly"`
}

// ConnectorOciVcnImport is vcnRouting.importFromCXPOptions (ImportFromCXPOptions.java).
type ConnectorOciVcnImport struct {
	RouteTables []ConnectorOciVcnRouteTable `json:"routeTables,omitempty"`
}

// ConnectorOciVcnRouteTable is one VCN route table, ID its OCI id
// (RouteTableImportOptions.java).
type ConnectorOciVcnRouteTable struct {
	ID              string  `json:"id" ak:"required,src=RouteTableImportOptions.java:23"`
	PrefixListIDs   []int64 `json:"prefixListIds,omitempty"`
	RouteImportMode string  `json:"routeImportMode" ak:"required,enum=ADVERTISE_DEFAULT_ROUTE|OVERRIDE_DEFAULT_ROUTE|ADVERTISE_CUSTOM_PREFIX,src=ImportFromCXPOptions.java:33-36"`
	DefaultRouting  *bool   `json:"defaultRouting,omitempty" ak:"readonly"`
}

// connectorOciVcnDescriptor: GET, PUT and DELETE answer a missing id with
// 404 AK000003, as every connector does. vcnId identifies the family in the
// GET body and OCI_VCN in the GET /connectors row, which also carries the
// state the D13 guard reads. An update that leaves vcnRouting out while every
// stored routing entry is backend-built omits it, so the backend rebuilds it
// from the VCN rather than storing it as user input.
var connectorOciVcnDescriptor = Descriptor{
	Collection:    "oci-vcn-connectors",
	FamilyType:    FamilyCheck{Field: "vcnId"},
	ConnectorType: "OCI_VCN",
	InUse: []Signature{
		{Status: 400, Code: "AK000002", Message: "associated with"},
		{Status: 409, Code: "AK000001", Message: "is still referred externally"},
	},
	// constraint.exists: another connector in the network already holds the
	// VCN (OCIVCNConnectorServiceImpl.java:231-235).
	Duplicate: []Signature{
		{Status: 400, FieldCode: "constraint.notunique"},
		{Status: 400, FieldCode: "connector.00000003"},
		{Status: 400, FieldCode: "constraint.exists"},
	},
	Provisionable:      true,
	MarkedRowLifecycle: true,
	DocState:           DocStatePlainGet,
	Uniqueness:         Uniqueness{Key: "name", Scope: ScopeConnectorsGroups},
	StateGuard:         StateGuardD13,
	RoutingProvenance:  []string{"vcnRouting"},
	Service:            ServiceTPS,
	CrossCheck:         checkConnector,
	Flags: []FlagRule{
		{Flag: "BILLING-TAGS", Effect: FlagIncludeField, Field: "billingTags", UISource: "src/views/network/provision/objectConfig/ociConfig/add/AddOCIConnector.js:45-46|:439"},
	},
}

// ConnectorOciVcnDescriptor returns the OCI VCN connector write descriptor.
func ConnectorOciVcnDescriptor() Descriptor { return connectorOciVcnDescriptor.clone() }

// NewConnectorOciVcnCore returns the write core for OCI VCN connectors.
func NewConnectorOciVcnCore(ac *AlkiraClient) (*Core[ConnectorOciVcnConfig], error) {
	return NewCore[ConnectorOciVcnConfig](ac, connectorOciVcnDescriptor)
}

// CreateConnectorOciVcn creates an OCI VCN connector.
func (ac *AlkiraClient) CreateConnectorOciVcn(ctx context.Context, c *ConnectorOciVcnConfig) (*ConnectorOciVcnConfig, *WriteResult, error) {
	core, err := NewConnectorOciVcnCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, c, WriteOptions{})
}

// GetConnectorOciVcn reads one OCI VCN connector and its raw body.
func (ac *AlkiraClient) GetConnectorOciVcn(ctx context.Context, id string) (*ConnectorOciVcnConfig, []byte, error) {
	core, err := NewConnectorOciVcnCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

// UpdateConnectorOciVcn applies a merge patch to an OCI VCN connector.
func (ac *AlkiraClient) UpdateConnectorOciVcn(ctx context.Context, id string, patch json.RawMessage) (*ConnectorOciVcnConfig, *WriteResult, error) {
	core, err := NewConnectorOciVcnCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

// DeleteConnectorOciVcn deletes an OCI VCN connector; a deployed one is
// marked for deletion until the next provision.
func (ac *AlkiraClient) DeleteConnectorOciVcn(ctx context.Context, id string) (*WriteResult, error) {
	core, err := NewConnectorOciVcnCore(ac)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

// ListConnectorOciVcnPaged reads every OCI VCN connector through paged reads.
func (ac *AlkiraClient) ListConnectorOciVcnPaged(ctx context.Context) ([]ConnectorOciVcnConfig, error) {
	core, err := NewConnectorOciVcnCore(ac)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
