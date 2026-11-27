// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ConnectorAzureVnetConfig is the write model of an Azure VNet connector (TPS
// model/AzureVNETConnector.java; GET /connectors reports its type as
// AZURE_VNET). The fields follow the portal create body (karrawa
// azureConfig/add/AddAzureConnector.js:388-437); optional fields are pointers
// or omitempty slices, so a zero value is never sent. An unset
// connectionMode is chosen by the backend from the CXP and the VNet
// (AzureVNETConnectorServiceImpl.java:354-356).
type ConnectorAzureVnetConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=AzureVNETConnector.java:111-115"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,charset=description,src=AzureVNETConnector.java:117-119"`
	// VNetID is the Azure resource id of the VNet; it cannot change
	// (AzureVNETConnectorServiceImpl.java:1534-1535).
	VNetID string `json:"vnetId" ak:"required,immutable,src=AzureVNETConnector.java:138-139|AzureVNETConnectorServiceImpl.java:1534-1535"`
	CXP    string `json:"cxp" ak:"required,maxlen=128,trimmed,src=AzureVNETConnector.java:144-147"`
	// CredentialID is an Azure VNet credential whose subscription matches the
	// VNet's (AzureVNETConnectorServiceImpl.java:316-318).
	CredentialID string `json:"credentialId" ak:"required,src=AzureVNETConnector.java:152-153"`
	// ResourceGroupName is the resource group of the VNet, kept for display;
	// the backend does not use it (AzureVNETConnector.java:155-156).
	ResourceGroupName string `json:"resourceGroupName,omitempty"`
	// Segments holds exactly one segment name (AzureVNETConnector.java:182-190).
	Segments    []string `json:"segments" ak:"required,src=AzureVNETConnector.java:182-190"`
	Group       *string  `json:"group,omitempty" ak:"maxlen=256,trimmed,charset=name,src=AzureVNETConnector.java:124-127"`
	BillingTags []int64  `json:"billingTags,omitempty" ak:"flag=BILLING-TAGS"`
	// Size defaults to SMALL in the backend (AzureVNETConnector.java:202-205).
	Size               *string  `json:"size,omitempty" ak:"enum=XSMALL|5XSMALL|SMALL|MEDIUM|LARGE|2LARGE|4LARGE|5LARGE|10LARGE,src=AzureVNETConnector.java:202-205|Defaults.java:148-150"`
	NativeServiceNames []string `json:"nativeServiceNames,omitempty"`
	ServiceTags        []string `json:"serviceTags,omitempty"`
	UDRListIDs         []int64  `json:"udrListIds,omitempty"`
	// Enabled is sent as true by the portal unless changed
	// (AddAzureConnector.js:402).
	Enabled                           *bool                      `json:"enabled,omitempty" ak:"default=true,src=AddAzureConnector.js:402"`
	SecondaryCXPs                     []string                   `json:"secondaryCXPs,omitempty"`
	ConnectionMode                    *string                    `json:"connectionMode,omitempty" ak:"default=VNET_GATEWAY,enum=VNET_GATEWAY|VNET_PEERING,src=AzureVNETConnector.java:829-832|AddAzureConnector.js:403-406"`
	CustomerASN                       any                        `json:"customerAsn,omitempty"`
	ScaleGroupID                      *string                    `json:"scaleGroupId,omitempty" ak:"flag=SCALE-GROUPS-FOR-CLOUD-CONNECTORS"`
	CXPPeeringGatewayID               *int64                     `json:"cxpPeeringGatewayId,omitempty" ak:"flag=ATH-ON-AWS-CXP"`
	DirectInterVnetCommunicationGroup *string                    `json:"directInterVnetCommunicationGroup,omitempty"`
	VNetRouting                       *ConnectorAzureVnetRouting `json:"vnetRouting,omitempty"`

	GroupID                             *int64          `json:"groupId,omitempty" ak:"readonly"`
	ImplicitGroupID                     *int64          `json:"implicitGroupId,omitempty" ak:"readonly"`
	DirectInterVnetCommunicationGroupID *int64          `json:"directInterVnetCommunicationGroupId,omitempty" ak:"readonly"`
	CustomerRegion                      *string         `json:"customerRegion,omitempty" ak:"readonly,rederived"`
	SubscriptionID                      *string         `json:"subscriptionId,omitempty" ak:"readonly,rederived"`
	RoutingMode                         *string         `json:"routingMode,omitempty" ak:"readonly,rederived"`
	PrimaryConnectorID                  *int64          `json:"primaryConnectorId,omitempty" ak:"readonly,rederived"`
	Primary                             *bool           `json:"primary,omitempty" ak:"readonly"`
	LastEnabled                         *int64          `json:"lastEnabled,omitempty" ak:"readonly"`
	TagID                               *int64          `json:"tagId,omitempty" ak:"readonly"`
	VGWID                               *string         `json:"vgwId,omitempty" ak:"readonly"`
	IsActiveActiveEnabled               *bool           `json:"isActiveActiveEnabled,omitempty" ak:"readonly"`
	GatewaySubnetPrefix                 *string         `json:"gatewaySubnetPrefix,omitempty" ak:"readonly"`
	GatewaySKU                          *string         `json:"gatewaySku,omitempty" ak:"readonly"`
	GatewayGeneration                   *string         `json:"gatewayGeneration,omitempty" ak:"readonly"`
	ObservedGatewayGeneration           *string         `json:"observedGatewayGeneration,omitempty" ak:"readonly"`
	ConfigStatus                        json.RawMessage `json:"configStatus,omitempty" ak:"readonly"`
	ValidationState                     json.RawMessage `json:"validationState,omitempty" ak:"readonly"`
	// State and PreviousState come from the GET /connectors row.
	State         string `json:"state,omitempty" ak:"readonly"`
	PreviousState string `json:"previousState,omitempty" ak:"readonly"`
}

// ConnectorAzureVnetRouting is AzureVNETConnector.vnetRouting
// (AzureVNETConnector.java:279-297).
type ConnectorAzureVnetRouting struct {
	ExportToCXPOptions   *ConnectorAzureVnetExport  `json:"exportToCXPOptions,omitempty"`
	ImportFromCXPOptions *ConnectorAzureVnetImport  `json:"importFromCXPOptions,omitempty"`
	ServiceRoutes        *ConnectorAzureVnetEntries `json:"serviceRoutes,omitempty"`
	UDRLists             *ConnectorAzureVnetEntries `json:"udrLists,omitempty"`
}

// ConnectorAzureVnetExport is vnetRouting.exportToCXPOptions.
type ConnectorAzureVnetExport struct {
	UserInputPrefixes []ConnectorAzureVnetPrefix `json:"userInputPrefixes"`
}

// ConnectorAzureVnetPrefix is one exported CIDR or subnet; ID is the Azure
// subnet id of a SUBNET entry.
type ConnectorAzureVnetPrefix struct {
	ID             *string  `json:"id,omitempty"`
	Value          *string  `json:"value,omitempty"`
	Values         []string `json:"values,omitempty"`
	Type           string   `json:"type" ak:"required,enum=CIDR|SUBNET,src=AzureVNETConnector.java:558-561"`
	DefaultRouting *bool    `json:"defaultRouting,omitempty"`
}

// ConnectorAzureVnetImport is vnetRouting.importFromCXPOptions. The VNet-wide
// prefixListIds and routeImportMode apply when no CIDR or subnet is selected;
// cidrs and subnets carry the same settings per selected entry. A patch that
// sets one side drops the other side's stored fields.
type ConnectorAzureVnetImport struct {
	PrefixListIDs   []int64                         `json:"prefixListIds,omitempty" ak:"exclusive=import:vnet"`
	RouteImportMode *string                         `json:"routeImportMode,omitempty" ak:"exclusive=import:vnet,enum=ADVERTISE_DEFAULT_ROUTE|OVERRIDE_DEFAULT_ROUTE|ADVERTISE_CUSTOM_PREFIX,src=AzureVNETConnector.java:552-556"`
	CIDRs           []ConnectorAzureVnetImportEntry `json:"cidrs,omitempty" ak:"exclusive=import:entries"`
	Subnets         []ConnectorAzureVnetImportEntry `json:"subnets,omitempty" ak:"exclusive=import:entries"`
}

// ConnectorAzureVnetImportEntry is one importFromCXPOptions cidrs or subnets
// entry; ID is set on subnets only.
type ConnectorAzureVnetImportEntry struct {
	ID              *string  `json:"id,omitempty"`
	Value           *string  `json:"value,omitempty"`
	Values          []string `json:"values,omitempty"`
	PrefixListIDs   []int64  `json:"prefixListIds,omitempty"`
	RouteImportMode *string  `json:"routeImportMode,omitempty" ak:"enum=ADVERTISE_DEFAULT_ROUTE|OVERRIDE_DEFAULT_ROUTE|ADVERTISE_CUSTOM_PREFIX,src=AzureVNETConnector.java:552-556"`
}

// ConnectorAzureVnetEntries is vnetRouting.serviceRoutes or udrLists: per
// CIDR and per subnet settings.
type ConnectorAzureVnetEntries struct {
	CIDRs   []ConnectorAzureVnetEntry `json:"cidrs,omitempty"`
	Subnets []ConnectorAzureVnetEntry `json:"subnets,omitempty"`
}

// ConnectorAzureVnetEntry is one serviceRoutes or udrLists entry; ID is set
// on subnets only.
type ConnectorAzureVnetEntry struct {
	ID                 *string  `json:"id,omitempty"`
	Value              *string  `json:"value,omitempty"`
	Values             []string `json:"values,omitempty"`
	NativeServiceNames []string `json:"nativeServiceNames,omitempty"`
	ServiceTags        []string `json:"serviceTags,omitempty"`
	UDRListIDs         []int64  `json:"udrListIds,omitempty"`
}

// connectorAzureVnetDescriptor: GET, PUT and DELETE answer a missing id with
// 404 AK000003, as every connector does. vnetId identifies the family in the
// GET body and AZURE_VNET in the GET /connectors row, which also carries the
// state the D13 guard reads. With RELEASE-ASYNC-NETWORK-ENTITY-VALIDATION-PHASE-1
// on, the portal saves with ?async=true (karrawa src/_api/azure.js:87-107).
var connectorAzureVnetDescriptor = Descriptor{
	Collection:    "azurevnetconnectors",
	FamilyType:    FamilyCheck{Field: "vnetId"},
	ConnectorType: "AZURE_VNET",
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
	CrossCheck:         checkConnectorAzureVnet,
	AsyncCapable:       true,
	Flags: []FlagRule{
		{Flag: "BILLING-TAGS", Effect: FlagIncludeField, Field: "billingTags", UISource: "src/views/network/provision/objectConfig/azureConfig/add/AddAzureConnector.js:116-117|:398-400"},
		{Flag: "SCALE-GROUPS-FOR-CLOUD-CONNECTORS", Effect: FlagIncludeField, Field: "scaleGroupId", UISource: "src/views/network/provision/objectConfig/azureConfig/add/AddAzureConnector.js:118|:408-410"},
		{Flag: "ATH-ON-AWS-CXP", Effect: FlagIncludeField, Field: "cxpPeeringGatewayId", UISource: "src/views/network/provision/objectConfig/azureConfig/add/AddAzureConnector.js:123|:412-414"},
	},
}

// checkConnectorAzureVnet adds to checkConnector: resourceGroupName must be
// the resource group named in vnetId, and a VNET_PEERING connector cannot be
// saved disabled (RT connector-mutations/RUN-SUMMARY.md:119).
func checkConnectorAzureVnet(m map[string]any) error {
	errs := []error{checkConnector(m)}
	vnet, _ := m["vnetId"].(string)
	rg, _ := m["resourceGroupName"].(string)
	if want, ok := azureResourceGroup(vnet); ok && rg != "" && !strings.EqualFold(rg, want) {
		errs = append(errs, fmt.Errorf("resourceGroupName %q is not the resource group of vnetId (%s)", rg, want))
	}
	if m["connectionMode"] == "VNET_PEERING" && m["enabled"] == false {
		errs = append(errs, errors.New("a VNET_PEERING connector cannot be disabled"))
	}
	return errors.Join(errs...)
}

// azureResourceGroup returns the resourceGroups segment of an Azure resource id.
func azureResourceGroup(id string) (string, bool) {
	parts := strings.Split(id, "/")
	for i := 0; i+1 < len(parts); i++ {
		if strings.EqualFold(parts[i], "resourceGroups") {
			return parts[i+1], true
		}
	}
	return "", false
}

// ConnectorAzureVnetDescriptor returns the Azure VNet connector write descriptor.
func ConnectorAzureVnetDescriptor() Descriptor { return connectorAzureVnetDescriptor.clone() }

// NewConnectorAzureVnetCore returns the write core for Azure VNet connectors.
func NewConnectorAzureVnetCore(ac *AlkiraClient) (*Core[ConnectorAzureVnetConfig], error) {
	return NewCore[ConnectorAzureVnetConfig](ac, connectorAzureVnetDescriptor)
}

// CreateConnectorAzureVnet creates an Azure VNet connector.
func (ac *AlkiraClient) CreateConnectorAzureVnet(ctx context.Context, c *ConnectorAzureVnetConfig) (*ConnectorAzureVnetConfig, *WriteResult, error) {
	core, err := NewConnectorAzureVnetCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, c, WriteOptions{})
}

// GetConnectorAzureVnet reads one Azure VNet connector and its raw body.
func (ac *AlkiraClient) GetConnectorAzureVnet(ctx context.Context, id string) (*ConnectorAzureVnetConfig, []byte, error) {
	core, err := NewConnectorAzureVnetCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

// UpdateConnectorAzureVnet applies a merge patch to an Azure VNet connector.
func (ac *AlkiraClient) UpdateConnectorAzureVnet(ctx context.Context, id string, patch json.RawMessage) (*ConnectorAzureVnetConfig, *WriteResult, error) {
	core, err := NewConnectorAzureVnetCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

// DeleteConnectorAzureVnet deletes an Azure VNet connector; a deployed one is
// marked for deletion until the next provision.
func (ac *AlkiraClient) DeleteConnectorAzureVnet(ctx context.Context, id string) (*WriteResult, error) {
	core, err := NewConnectorAzureVnetCore(ac)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

// ListConnectorAzureVnetPaged reads every Azure VNet connector through paged reads.
func (ac *AlkiraClient) ListConnectorAzureVnetPaged(ctx context.Context) ([]ConnectorAzureVnetConfig, error) {
	core, err := NewConnectorAzureVnetCore(ac)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
