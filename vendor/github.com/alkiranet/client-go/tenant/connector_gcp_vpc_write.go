// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// ConnectorGcpVpcConfig is the write model of a GCP VPC connector (TPS
// model/GCPVPCConnector.java; GET /connectors reports its type as GCP_VPC).
// The fields follow the portal create body (karrawa
// gcpConfig/add/AddGCPConnector.js:274-310); optional fields are pointers or
// omitempty slices, so a zero value is never sent. The backend reads the VPC
// from GCP on every save (GCPVPCConnectorServiceImpl.java:200), so a save
// needs a live VPC and a GCP credential that can read it.
type ConnectorGcpVpcConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=GCPVPCConnector.java:61-65"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,charset=description,src=GCPVPCConnector.java:67-69"`
	// Size defaults to SMALL in the backend (GCPVPCConnector.java:71-74).
	Size *string `json:"size,omitempty" ak:"enum=XSMALL|5XSMALL|SMALL|MEDIUM|LARGE|2LARGE|4LARGE|5LARGE|10LARGE,src=GCPVPCConnector.java:71-74|Defaults.java:148-150"`
	// Segments holds exactly one segment name (GCPVPCConnector.java:76-83).
	Segments []string `json:"segments" ak:"required,src=GCPVPCConnector.java:76-83"`
	CXP      string   `json:"cxp" ak:"required,maxlen=128,trimmed,src=GCPVPCConnector.java:87-90"`
	// CustomerRegion is the GCP region of the connector. The model does not
	// mark it required, but the save reads it unchecked and a missing value
	// answers 500 (GCPVPCConnectorServiceImpl.java:233, :256).
	CustomerRegion string `json:"customerRegion" ak:"required,src=GCPVPCConnector.java:92|GCPVPCConnectorServiceImpl.java:233|:256"`
	// VPCName is the GCP network name; it cannot change
	// (GCPVPCConnectorServiceImpl.java:210-211).
	VPCName string `json:"vpcName" ak:"required,immutable,src=GCPVPCConnector.java:97-98|GCPVPCConnectorServiceImpl.java:210-211"`
	// ProjectID defaults to the credential's project
	// (GCPVPCConnector.java:296-302) and cannot change
	// (GCPVPCConnectorServiceImpl.java:213-216).
	ProjectID    *string `json:"projectId,omitempty" ak:"immutable,src=GCPVPCConnectorServiceImpl.java:213-216"`
	Group        *string `json:"group,omitempty" ak:"maxlen=256,trimmed,charset=name,src=GCPVPCConnector.java:105-108"`
	CredentialID string  `json:"credentialId" ak:"required,src=GCPVPCConnector.java:118-119"`
	BillingTags  []int64 `json:"billingTags,omitempty" ak:"flag=BILLING-TAGS"`
	// GCPRouting is returned as stored; left out, the backend stores an
	// import of ADVERTISE_DEFAULT_ROUTE and no export options
	// (GCPVPCConnector.java:130-131, :220-224).
	GCPRouting *ConnectorGcpVpcRoutingOptions `json:"gcpRouting,omitempty"`
	// Enabled: the backend stores an unset value as true
	// (GCPVPCConnectorServiceImpl.java:814).
	Enabled       *bool    `json:"enabled,omitempty"`
	SecondaryCXPs []string `json:"secondaryCXPs,omitempty"`
	// CustomerASN defaults to 64522 (GCPVPCConnectorServiceImpl.java:84, :811).
	CustomerASN  *int64  `json:"customerAsn,omitempty" ak:"range=64512..65534|4200000000..4294967294,src=GCPVPCConnector.java:150-151|Defaults.java:1490-1496"`
	ScaleGroupID *string `json:"scaleGroupId,omitempty" ak:"flag=SCALE-GROUPS-FOR-CLOUD-CONNECTORS"`

	// VPCID and RoutingMode are read from the live VPC on every save
	// (GCPVPCConnectorServiceImpl.java:201, :251).
	VPCID              *string         `json:"vpcId,omitempty" ak:"readonly,rederived"`
	RoutingMode        *string         `json:"routingMode,omitempty" ak:"readonly,rederived"`
	GroupID            *int64          `json:"groupId,omitempty" ak:"readonly"`
	ImplicitGroupID    *int64          `json:"implicitGroupId,omitempty" ak:"readonly"`
	PrimaryConnectorID *int64          `json:"primaryConnectorId,omitempty" ak:"readonly,rederived"`
	Primary            *bool           `json:"primary,omitempty" ak:"readonly"`
	LastEnabled        *int64          `json:"lastEnabled,omitempty" ak:"readonly"`
	ConfigStatus       json.RawMessage `json:"configStatus,omitempty" ak:"readonly"`
	// State and PreviousState come from the GET /connectors row.
	State         string `json:"state,omitempty" ak:"readonly"`
	PreviousState string `json:"previousState,omitempty" ak:"readonly"`
}

// ConnectorGcpVpcRoutingOptions is GCPVPCConnector.gcpRouting.
type ConnectorGcpVpcRoutingOptions struct {
	ExportToCXPOptions   *ConnectorGcpVpcExport `json:"exportToCXPOptions,omitempty"`
	ImportFromCXPOptions *ConnectorGcpVpcImport `json:"importFromCXPOptions,omitempty"`
}

// ConnectorGcpVpcExport is gcpRouting.exportToCXPOptions: every subnet, or
// the subnets listed (GCPVPCConnectorServiceImpl.java:266-271).
type ConnectorGcpVpcExport struct {
	ExportAllSubnets  *bool                   `json:"exportAllSubnets,omitempty"`
	UserInputPrefixes []ConnectorGcpVpcPrefix `json:"userInputPrefixes,omitempty"`
}

// ConnectorGcpVpcPrefix is one exported subnet: ID is the GCP subnet id and
// FqID its projects/<p>/regions/<r>/subnetworks/<name> path; one of them is
// needed (GCPVPCConnectorServiceImpl.java:275-277).
type ConnectorGcpVpcPrefix struct {
	ID    *string `json:"id,omitempty"`
	FqID  *string `json:"fqId,omitempty"`
	Value string  `json:"value" ak:"required,src=GCPVPCConnectorServiceImpl.java:281-285"`
	Type  string  `json:"type" ak:"required,enum=SUBNET,src=GCPVPCConnectorServiceImpl.java:273|:297-299"`
}

// ConnectorGcpVpcImport is gcpRouting.importFromCXPOptions.
type ConnectorGcpVpcImport struct {
	PrefixListIDs   []int64 `json:"prefixListIds,omitempty"`
	RouteImportMode *string `json:"routeImportMode,omitempty" ak:"enum=ADVERTISE_DEFAULT_ROUTE|ADVERTISE_CUSTOM_PREFIX,src=GCPVPCConnector.java:256-259"`
}

// connectorGcpVpcDescriptor: GET, PUT and DELETE answer a missing id with
// 404 AK000003, as every connector does. vpcName identifies the family in
// the GET body and GCP_VPC in the GET /connectors row, which also carries
// the state the D13 guard reads. The routing is returned as stored, so it
// carries no routing provenance.
var connectorGcpVpcDescriptor = Descriptor{
	Collection:    "gcpvpcconnectors",
	FamilyType:    FamilyCheck{Field: "vpcName"},
	ConnectorType: "GCP_VPC",
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
	CrossCheck:         checkConnectorGcpVpc,
	Flags: []FlagRule{
		{Flag: "BILLING-TAGS", Effect: FlagIncludeField, Field: "billingTags", UISource: "src/views/network/provision/objectConfig/gcpConfig/add/AddGCPConnector.js:60-61|:286"},
		{Flag: "SCALE-GROUPS-FOR-CLOUD-CONNECTORS", Effect: FlagIncludeField, Field: "scaleGroupId", UISource: "src/views/network/provision/objectConfig/gcpConfig/add/AddGCPConnector.js:62|:289-291"},
	},
}

// checkConnectorGcpVpc adds to checkConnector the routing rules the backend
// applies at save: exportAllSubnets true takes no subnets and false needs
// some (GCPVPCConnectorServiceImpl.java:266-271), a subnet needs id or fqId
// (:275-277), and prefixListIds go with ADVERTISE_CUSTOM_PREFIX only
// (:688-693).
func checkConnectorGcpVpc(m map[string]any) error {
	errs := []error{checkConnector(m)}
	routing, _ := m["gcpRouting"].(map[string]any)
	if exp, ok := routing["exportToCXPOptions"].(map[string]any); ok {
		prefixes, _ := exp["userInputPrefixes"].([]any)
		switch exp["exportAllSubnets"] {
		case true:
			if len(prefixes) > 0 {
				errs = append(errs, errors.New("gcpRouting.exportToCXPOptions: userInputPrefixes must be empty when exportAllSubnets is true"))
			}
		case false:
			if len(prefixes) == 0 {
				errs = append(errs, errors.New("gcpRouting.exportToCXPOptions: userInputPrefixes must be set when exportAllSubnets is false"))
			}
		}
		for i, e := range prefixes {
			p, _ := e.(map[string]any)
			id, _ := p["id"].(string)
			fq, _ := p["fqId"].(string)
			if id == "" && fq == "" {
				errs = append(errs, fmt.Errorf("gcpRouting.exportToCXPOptions.userInputPrefixes[%d]: a subnet needs id or fqId", i))
			}
		}
	}
	if imp, ok := routing["importFromCXPOptions"].(map[string]any); ok {
		lists, _ := imp["prefixListIds"].([]any)
		mode, _ := imp["routeImportMode"].(string)
		if mode == "" {
			mode = "ADVERTISE_DEFAULT_ROUTE"
		}
		if mode == "ADVERTISE_DEFAULT_ROUTE" && len(lists) > 0 {
			errs = append(errs, errors.New("gcpRouting.importFromCXPOptions: prefixListIds must be empty when routeImportMode is ADVERTISE_DEFAULT_ROUTE"))
		}
		if mode == "ADVERTISE_CUSTOM_PREFIX" && len(lists) == 0 {
			errs = append(errs, errors.New("gcpRouting.importFromCXPOptions: prefixListIds must be set when routeImportMode is ADVERTISE_CUSTOM_PREFIX"))
		}
	}
	return errors.Join(errs...)
}

// ConnectorGcpVpcDescriptor returns the GCP VPC connector write descriptor.
func ConnectorGcpVpcDescriptor() Descriptor { return connectorGcpVpcDescriptor.clone() }

// NewConnectorGcpVpcCore returns the write core for GCP VPC connectors.
func NewConnectorGcpVpcCore(ac *AlkiraClient) (*Core[ConnectorGcpVpcConfig], error) {
	return NewCore[ConnectorGcpVpcConfig](ac, connectorGcpVpcDescriptor)
}

// CreateConnectorGcpVpc creates a GCP VPC connector.
func (ac *AlkiraClient) CreateConnectorGcpVpc(ctx context.Context, c *ConnectorGcpVpcConfig) (*ConnectorGcpVpcConfig, *WriteResult, error) {
	core, err := NewConnectorGcpVpcCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, c, WriteOptions{})
}

// GetConnectorGcpVpc reads one GCP VPC connector and its raw body.
func (ac *AlkiraClient) GetConnectorGcpVpc(ctx context.Context, id string) (*ConnectorGcpVpcConfig, []byte, error) {
	core, err := NewConnectorGcpVpcCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

// UpdateConnectorGcpVpc applies a merge patch to a GCP VPC connector.
func (ac *AlkiraClient) UpdateConnectorGcpVpc(ctx context.Context, id string, patch json.RawMessage) (*ConnectorGcpVpcConfig, *WriteResult, error) {
	core, err := NewConnectorGcpVpcCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

// DeleteConnectorGcpVpc deletes a GCP VPC connector; a deployed one is
// marked for deletion until the next provision.
func (ac *AlkiraClient) DeleteConnectorGcpVpc(ctx context.Context, id string) (*WriteResult, error) {
	core, err := NewConnectorGcpVpcCore(ac)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

// ListConnectorGcpVpcPaged reads every GCP VPC connector through paged reads.
func (ac *AlkiraClient) ListConnectorGcpVpcPaged(ctx context.Context) ([]ConnectorGcpVpcConfig, error) {
	core, err := NewConnectorGcpVpcCore(ac)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
