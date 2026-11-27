// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// ConnectorIPsecConfig is the write model of an IPsec connector (TPS
// model/IPsecConnector.java; GET /connectors reports its type as IP_SEC).
// Pre-shared keys and the BGP key are write-only: a create may set them, a
// patch cannot, and an update resends the stored values, so the stored keys
// stay in place. A site sent in an update without its stored id is a new
// site, for which the backend generates new keys.
type ConnectorIPsecConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=IPsecConnector.java:65-69"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,charset=description,src=IPsecConnector.java:71-73"`
	// Segments holds exactly one segment name (IPsecConnector.java:75-80).
	Segments []string `json:"segments" ak:"required,src=IPsecConnector.java:75-80"`
	CXP      string   `json:"cxp" ak:"required,maxlen=128,trimmed,src=IPsecConnector.java:86-89"`
	Group    *string  `json:"group,omitempty" ak:"maxlen=256,trimmed,charset=name,src=IPsecConnector.java:94-97"`
	// Size defaults to SMALL in the backend (IPsecConnector.java:107-110).
	Size  *string              `json:"size,omitempty" ak:"enum=SMALL|MEDIUM|LARGE|2LARGE|4LARGE|5LARGE,src=IPsecConnector.java:107-110|Defaults.java:148-150"`
	Sites []ConnectorIPsecSite `json:"sites" ak:"required,src=IPsecConnector.java:113-114"`
	// VPNMode selects the side the backend keeps: routingOptions for
	// ROUTE_BASED, policyOptions for POLICY_BASED (IPsecConnectorServiceImpl.java:783-787).
	// An absent vpnMode is ROUTE_BASED (IPsecConnector.java:123). Switching
	// mode takes vpnMode and the new side in one patch.
	VPNMode        *string                       `json:"vpnMode,omitempty" ak:"enum=ROUTE_BASED|POLICY_BASED,src=IPsecConnector.java:123|VPNMode.java:8-10"`
	RoutingOptions *ConnectorIPsecRoutingOptions `json:"routingOptions,omitempty" ak:"exclusive=mode:route"`
	PolicyOptions  *ConnectorIPsecPolicyOptions  `json:"policyOptions,omitempty" ak:"exclusive=mode:policy"`
	Enabled        *bool                         `json:"enabled,omitempty" ak:"flag=CXP-FAILOVER-REGION"`
	SecondaryCXPs  []string                      `json:"secondaryCXPs,omitempty" ak:"flag=CXP-FAILOVER-REGION"`
	ScaleGroupID   *string                       `json:"scaleGroupId,omitempty" ak:"flag=SCALE-GROUPS-FOR-USER-AND-SITES"`
	// SegmentOptions is keyed by segment name.
	SegmentOptions map[string]ConnectorIPsecSegmentOptions `json:"segmentOptions,omitempty"`

	GroupID            *int64          `json:"groupId,omitempty" ak:"readonly"`
	ImplicitGroupID    *int64          `json:"implicitGroupId,omitempty" ak:"readonly"`
	PrimaryConnectorID *int64          `json:"primaryConnectorId,omitempty" ak:"readonly,rederived"`
	Primary            *bool           `json:"primary,omitempty" ak:"readonly"`
	LastEnabled        *int64          `json:"lastEnabled,omitempty" ak:"readonly"`
	Type               string          `json:"type,omitempty" ak:"readonly"`
	ValidationState    json.RawMessage `json:"validationState,omitempty" ak:"readonly"`
	ConfigStatus       json.RawMessage `json:"configStatus,omitempty" ak:"readonly"`
	State              string          `json:"state,omitempty" ak:"readonly"`
	PreviousState      string          `json:"previousState,omitempty" ak:"readonly"`
}

// ConnectorIPsecSite is one IPsec connector site (TPS model/IPsecConnectorSite.java).
type ConnectorIPsecSite struct {
	ID            json.Number `json:"id,omitempty"`
	Name          string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=IPsecConnectorSite.java:44-48"`
	GatewayIPType *string     `json:"gatewayIpType,omitempty" ak:"enum=STATIC|DYNAMIC,src=IPsecConnectorSite.java:98-101"`
	GatewayNo     *int64      `json:"gatewayNo,omitempty"`
	CustomerGwIP  *string     `json:"customerGwIp,omitempty"`
	// PresharedKeys are write-only: a create may set one or two keys, as the
	// portal's optional key fields do; left out or blank, the backend
	// generates a random key per slot (IPsecConnectorServiceImpl.java:625-633,
	// :668-674). A patch cannot set them.
	PresharedKeys          []string                    `json:"presharedKeys,omitempty" ak:"writeonly,maxitems=2,maxlen=64,charset=psk,src=IPsecConnectorSite.java:63-69"`
	BillingTags            []int64                     `json:"billingTags,omitempty"`
	EnableTunnelRedundancy *bool                       `json:"enableTunnelRedundancy,omitempty"`
	HAMode                 *string                     `json:"haMode,omitempty" ak:"enum=ACTIVE|STANDBY,src=HAMode.java:8-11"`
	Advanced               *ConnectorIPsecSiteAdvanced `json:"advanced,omitempty"`

	ConnectorID *int64 `json:"connectorId,omitempty" ak:"readonly"`
	State       string `json:"state,omitempty" ak:"readonly"`
}

// ConnectorIPsecSiteAdvanced is a site's IKE and ESP settings
// (IPsecConnectorInstanceAdvancedAttributes.java:44-107). The auth values
// are IKE identities, not keys; a DYNAMIC site needs remoteAuthValue
// (IPSecConnectorValidator.java:109-112).
type ConnectorIPsecSiteAdvanced struct {
	Initiator               *bool    `json:"initiator,omitempty"`
	IKEVersion              *string  `json:"ikeVersion,omitempty"`
	IKERekeyTime            *int64   `json:"ikeRekeyTime,omitempty"`
	IKEOverTime             *int64   `json:"ikeOverTime,omitempty"`
	IKERandomTime           *int64   `json:"ikeRandomTime,omitempty"`
	ESPRekeyTime            *int64   `json:"espRekeyTime,omitempty"`
	ESPLifeTime             *int64   `json:"espLifeTime,omitempty"`
	ESPRandomTime           *int64   `json:"espRandomTime,omitempty"`
	DPDTimeout              *int64   `json:"dPDTimeout,omitempty"`
	DPDDelay                *int64   `json:"dPDDelay,omitempty"`
	ReplayWindowSize        *int64   `json:"replayWindowSize,omitempty"`
	IKEEncryptionAlgorithms []string `json:"ikeEncryptionAlgorithms,omitempty"`
	IKEIntegrityAlgorithms  []string `json:"ikeIntegrityAlgorithms,omitempty"`
	IKEDHGroupNumbers       []string `json:"ikeDHGroupNumbers,omitempty"`
	ESPEncryptionAlgorithms []string `json:"espEncryptionAlgorithms,omitempty"`
	ESPIntegrityAlgorithms  []string `json:"espIntegrityAlgorithms,omitempty"`
	ESPDHGroupNumbers       []string `json:"espDHGroupNumbers,omitempty"`
	LocalAuthType           *string  `json:"localAuthType,omitempty"`
	LocalAuthValue          *string  `json:"localAuthValue,omitempty"`
	RemoteAuthType          *string  `json:"remoteAuthType,omitempty"`
	RemoteAuthValue         *string  `json:"remoteAuthValue,omitempty"`
}

// ConnectorIPsecRoutingOptions is routingOptions of a ROUTE_BASED connector.
// A patch that sets dynamicRouting to null moves the connector to static
// routing and drops the stored BGP key.
type ConnectorIPsecRoutingOptions struct {
	StaticRouting  *ConnectorIPsecStaticRouting  `json:"staticRouting,omitempty"`
	DynamicRouting *ConnectorIPsecDynamicRouting `json:"dynamicRouting,omitempty" ak:"droppable"`
}

// ConnectorIPsecStaticRouting is routingOptions.staticRouting.
type ConnectorIPsecStaticRouting struct {
	PrefixListID *int64  `json:"prefixListId,omitempty"`
	Availability *string `json:"availability,omitempty"`
}

// ConnectorIPsecDynamicRouting is routingOptions.dynamicRouting. The ASNs
// are numbers or numeric strings (DynamicRouting.java:23, :32).
type ConnectorIPsecDynamicRouting struct {
	CustomerGwASN any `json:"customerGwAsn,omitempty"`
	LocalASN      any `json:"localAsn,omitempty"`
	// BGPAuthKeyAlkira is write-only and optional, as in the portal; left
	// out, BGP runs without authentication. The backend refuses a key
	// shorter than 6 characters (DynamicRoutingValidator.java:16-24); that
	// rule is left to it so a stored shorter key never blocks an update.
	BGPAuthKeyAlkira *string `json:"bgpAuthKeyAlkira,omitempty" ak:"writeonly,maxlen=80,trimmed,src=DynamicRoutingValidator.java:16-24"`
	Availability     *string `json:"availability,omitempty"`
}

// ConnectorIPsecPolicyOptions is policyOptions of a POLICY_BASED connector.
type ConnectorIPsecPolicyOptions struct {
	BranchTSPrefixListIDs []int64                         `json:"branchTSPrefixListIds,omitempty" ak:"minitems=1,src=PolicyOptions.java:24-25"`
	CXPTSPrefixListIDs    []int64                         `json:"cxpTSPrefixListIds,omitempty" ak:"minitems=1,src=PolicyOptions.java:27-28"`
	TSPairs               []ConnectorIPsecTrafficSelector `json:"tsPairs,omitempty"`
}

// ConnectorIPsecTrafficSelector is one policyOptions.tsPairs entry
// (TPS model/ipsec/TrafficSelector.java).
type ConnectorIPsecTrafficSelector struct {
	SourcePrefix string `json:"sourcePrefix,omitempty"`
	DestPrefix   string `json:"destPrefix,omitempty"`
}

// ConnectorIPsecSegmentOptions is one segmentOptions entry.
type ConnectorIPsecSegmentOptions struct {
	DisableInternetExit   *bool `json:"disableInternetExit,omitempty"`
	AdvertiseOnPremRoutes *bool `json:"advertiseOnPremRoutes,omitempty"`
}

// connectorIPsecDescriptor: GET, PUT and DELETE answer a missing id with 404
// AK000003, as every connector does. sites identifies the family in the GET
// body and IP_SEC in the GET /connectors row, which also carries the state
// the D13 guard reads. Sites merge by id: a patch must list every stored site
// id, and the stored keys of each listed site are resent. With
// RELEASE-ASYNC-NETWORK-ENTITY-VALIDATION-PHASE-1 on, the portal saves with
// ?async=true (karrawa src/_api/ipsec.js:21-37).
var connectorIPsecDescriptor = Descriptor{
	Collection:    "ipsecconnectors",
	FamilyType:    FamilyCheck{Field: "sites"},
	ConnectorType: "IP_SEC",
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
	IDMatchedArrays:    []string{"sites"},
	StateGuard:         StateGuardD13,
	Service:            ServiceTPS,
	CrossCheck:         checkConnectorIPsec,
	AsyncCapable:       true,
	Flags: []FlagRule{
		{Flag: "SCALE-GROUPS-FOR-USER-AND-SITES", Effect: FlagIncludeField, Field: "scaleGroupId", UISource: "src/views/network/provision/objectConfig/ipsecConfig/utils.js:582|:675-677"},
		{Flag: "CXP-FAILOVER-REGION", Effect: FlagIncludeField, Field: "enabled", UISource: "src/views/network/provision/objectConfig/ipsecConfig/utils.js:581|:713-714"},
		{Flag: "CXP-FAILOVER-REGION", Effect: FlagIncludeField, Field: "secondaryCXPs", UISource: "src/views/network/provision/objectConfig/ipsecConfig/utils.js:581|:713-717"},
	},
}

// checkConnectorIPsec adds to checkConnector: the options side must match
// vpnMode, since the backend keeps only the side vpnMode names and refuses a
// body without it (IPSecConnectorValidator.java:25-90;
// IPsecConnectorServiceImpl.java:783-787); ROUTE_BASED needs static or
// dynamic routing; and a DYNAMIC site needs
// advanced.remoteAuthValue (IPSecConnectorValidator.java:109-112).
func checkConnectorIPsec(m map[string]any) error {
	errs := []error{checkConnector(m)}
	mode, _ := m["vpnMode"].(string)
	if mode == "" {
		mode = "ROUTE_BASED"
	}
	want, other := "routingOptions", "policyOptions"
	if mode == "POLICY_BASED" {
		want, other = other, want
	}
	if m[want] == nil {
		errs = append(errs, fmt.Errorf("%s must be set when vpnMode is %s", want, mode))
	}
	if ro, ok := m["routingOptions"].(map[string]any); ok && mode == "ROUTE_BASED" && ro["staticRouting"] == nil && ro["dynamicRouting"] == nil {
		errs = append(errs, errors.New("routingOptions must hold staticRouting or dynamicRouting"))
	}
	if m[other] != nil {
		errs = append(errs, fmt.Errorf("%s is set but vpnMode is %s; a mode switch sets vpnMode and %s in one patch", other, mode, other))
	}
	sites, _ := m["sites"].([]any)
	for _, e := range sites {
		site, _ := e.(map[string]any)
		if site["gatewayIpType"] != "DYNAMIC" {
			continue
		}
		adv, _ := site["advanced"].(map[string]any)
		if v, _ := adv["remoteAuthValue"].(string); v == "" {
			errs = append(errs, fmt.Errorf("site %v: advanced.remoteAuthValue must be set when gatewayIpType is DYNAMIC", site["name"]))
		}
	}
	return errors.Join(errs...)
}

// ConnectorIPsecDescriptor returns the IPsec connector write descriptor.
func ConnectorIPsecDescriptor() Descriptor { return connectorIPsecDescriptor.clone() }

// NewConnectorIPsecCore returns the write core for IPsec connectors.
func NewConnectorIPsecCore(ac *AlkiraClient) (*Core[ConnectorIPsecConfig], error) {
	return NewCore[ConnectorIPsecConfig](ac, connectorIPsecDescriptor)
}

// CreateConnectorIPsec creates an IPsec connector. Pre-shared keys and the
// BGP key are optional, as in the portal: a site without keys gets keys the
// backend generates.
func (ac *AlkiraClient) CreateConnectorIPsec(ctx context.Context, c *ConnectorIPsecConfig) (*ConnectorIPsecConfig, *WriteResult, error) {
	core, err := NewConnectorIPsecCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, c, WriteOptions{})
}

// GetConnectorIPsec reads one IPsec connector and its raw body.
func (ac *AlkiraClient) GetConnectorIPsec(ctx context.Context, id string) (*ConnectorIPsecConfig, []byte, error) {
	core, err := NewConnectorIPsecCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

// UpdateConnectorIPsec applies a merge patch to an IPsec connector; the
// stored pre-shared keys are resent unchanged.
func (ac *AlkiraClient) UpdateConnectorIPsec(ctx context.Context, id string, patch json.RawMessage) (*ConnectorIPsecConfig, *WriteResult, error) {
	core, err := NewConnectorIPsecCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

// DeleteConnectorIPsec deletes an IPsec connector; a deployed one is marked
// for deletion until the next provision.
func (ac *AlkiraClient) DeleteConnectorIPsec(ctx context.Context, id string) (*WriteResult, error) {
	core, err := NewConnectorIPsecCore(ac)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

// ListConnectorIPsecPaged reads every IPsec connector through paged reads.
func (ac *AlkiraClient) ListConnectorIPsecPaged(ctx context.Context) ([]ConnectorIPsecConfig, error) {
	core, err := NewConnectorIPsecCore(ac)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
