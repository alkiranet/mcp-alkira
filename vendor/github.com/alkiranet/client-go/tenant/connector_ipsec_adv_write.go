// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ConnectorIPsecAdvConfig is the write model of an advanced IPsec connector
// (TPS model/ipsec/AdvancedIPSecConnector.java; GET /connectors reports its
// type as ADV_IP_SEC). Tunnel pre-shared keys and the BGP key are
// write-only: a create may set them, a patch cannot set them, and an update resends the stored values, so the
// stored keys stay in place. A gateway or tunnel sent in an update without
// its stored id is new, and the backend generates a new key for each new
// tunnel (AdvancedIPSecConnectorServiceImpl.java:497-498, :538-540).
type ConnectorIPsecAdvConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=AdvancedIPSecConnector.java:73-77"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,charset=description,src=AdvancedIPSecConnector.java:79-81"`
	CXP         string      `json:"cxp" ak:"required,maxlen=128,trimmed,src=AdvancedIPSecConnector.java:92-95"`
	Group       *string     `json:"group,omitempty" ak:"maxlen=256,trimmed,charset=name,src=AdvancedIPSecConnector.java:100-103"`
	// Size has no backend default (AdvancedIPSecConnector.java:113-115).
	Size    string `json:"size" ak:"required,enum=SMALL|MEDIUM|LARGE|2LARGE|4LARGE|5LARGE|10LARGE,src=AdvancedIPSecConnector.java:113-115|Defaults.java:148-150"`
	Enabled *bool  `json:"enabled,omitempty" ak:"flag=CXP-FAILOVER-REGION"`
	// BillingTags is connector-level in this family, unlike the IPsec
	// connector's per-site tags.
	BillingTags           []int64 `json:"billingTags,omitempty"`
	AdvertiseDefaultRoute *bool   `json:"advertiseDefaultRoute,omitempty" ak:"required,src=AdvancedIPSecConnector.java:133-134"`
	AdvertiseOnPremRoutes *bool   `json:"advertiseOnPremRoutes,omitempty" ak:"required,src=AdvancedIPSecConnector.java:136-137"`
	DestinationType       string  `json:"destinationType" ak:"required,enum=IPSEC_ENDPOINT|AWS_VPN_CONNECTION|AZURE_VPN_CONNECTION,src=AdvancedIPSecConnector.java:139-140|:309-313"`
	// Segment is the segment name; the backend refuses a name it does not
	// know (AdvancedIPSecConnectorServiceImpl.java:260-262).
	Segment string `json:"segment" ak:"required,trimmed,src=AdvancedIPSecConnector.java:142-143"`
	// TunnelsPerGateway is the tunnel count of every gateway; a gateway sent
	// without tunnels gets that many default tunnels
	// (AdvancedIPSecConnectorValidator.java:76-89).
	TunnelsPerGateway *int64                     `json:"tunnelsPerGateway,omitempty" ak:"required,min=1,src=AdvancedIPSecConnectorValidator.java:76-89"`
	Gateways          []ConnectorIPsecAdvGateway `json:"gateways" ak:"required,distinct=name,src=AdvancedIPSecConnectorValidator.java:147-153"`
	// VPNMode is ROUTE_BASED only: the backend refuses POLICY_BASED for this
	// family (AdvancedIPSecConnectorValidator.java:128-131).
	VPNMode        string                        `json:"vpnMode" ak:"required,enum=ROUTE_BASED,src=AdvancedIPSecConnector.java:156-157|AdvancedIPSecConnectorValidator.java:128-131"`
	RoutingOptions *ConnectorIPsecRoutingOptions `json:"routingOptions,omitempty" ak:"required,src=AdvancedIPSecConnectorValidator.java:91-93"`
	// ScaleGroupID cannot change once the connector is provisioned
	// (AdvancedIPSecConnectorServiceImpl.java:294-300).
	ScaleGroupID *string `json:"scaleGroupId,omitempty" ak:"immutable,flag=SCALE-GROUPS-FOR-USER-AND-SITES"`

	TenantID        *int64          `json:"tenantId,omitempty" ak:"readonly"`
	TenantNetworkID *int64          `json:"tenantNetworkId,omitempty" ak:"readonly"`
	GroupID         *int64          `json:"groupId,omitempty" ak:"readonly"`
	ImplicitGroupID *int64          `json:"implicitGroupId,omitempty" ak:"readonly"`
	SegmentID       *int64          `json:"segmentId,omitempty" ak:"readonly"`
	IPSecScaleGroup *string         `json:"ipSecScaleGroup,omitempty" ak:"readonly"`
	LastEnabled     *int64          `json:"lastEnabled,omitempty" ak:"readonly"`
	Type            string          `json:"type,omitempty" ak:"readonly"`
	ValidationState json.RawMessage `json:"validationState,omitempty" ak:"readonly"`
	State           string          `json:"state,omitempty" ak:"readonly"`
	PreviousState   string          `json:"previousState,omitempty" ak:"readonly"`
}

// ConnectorIPsecAdvGateway is one customer gateway (TPS
// model/ipsec/IPSecGateway.java). A gateway id the connector does not hold
// is refused (AdvancedIPSecConnectorServiceImpl.java:476-479).
type ConnectorIPsecAdvGateway struct {
	ID           json.Number               `json:"id,omitempty"`
	Name         string                    `json:"name" ak:"required,src=IPSecGateway.java:32-33"`
	CustomerGwIP string                    `json:"customerGwIp" ak:"required,src=IPSecGateway.java:37-40"`
	HAMode       *string                   `json:"haMode,omitempty" ak:"enum=ACTIVE|STANDBY,src=HAMode.java:8-11"`
	Tunnels      []ConnectorIPsecAdvTunnel `json:"tunnels,omitempty"`

	// GatewayNo is the gateway's position, set on every save
	// (AdvancedIPSecConnectorServiceImpl.java:423-426).
	GatewayNo   *int64 `json:"gatewayNo,omitempty" ak:"readonly"`
	ConnectorID *int64 `json:"connectorId,omitempty" ak:"readonly"`
}

// ConnectorIPsecAdvTunnel is one gateway tunnel (TPS
// model/ipsec/IPSecTunnel.java). Its id is a string. advanced and profileId
// go together (AdvancedIPSecConnectorValidator.java:33-41).
type ConnectorIPsecAdvTunnel struct {
	ID string `json:"id,omitempty"`
	// PreSharedKey is write-only: a create may set it, as the portal's
	// optional key field does; left out or blank, the backend generates a
	// random key (AdvancedIPSecConnectorServiceImpl.java:538-540).
	PreSharedKey *string                     `json:"preSharedKey,omitempty" ak:"writeonly,maxlen=64,charset=psk,src=IPSecTunnel.java:33-36"`
	ProfileID    *int64                      `json:"profileId,omitempty"`
	Advanced     *ConnectorIPsecSiteAdvanced `json:"advanced,omitempty"`
	// A tunnel without cxpEnd and customerEnd saves, but the provision of
	// such a connector fails in resource-manager and leaves the reservation
	// retrying, so a create needs both ends.
	CustomerEnd *ConnectorIPsecAdvCustomerEnd `json:"customerEnd,omitempty" ak:"required,src=AdvancedIPSecConnectorValidator.java:64-74"`
	CXPEnd      *ConnectorIPsecAdvCXPEnd      `json:"cxpEnd,omitempty" ak:"required,src=AdvancedIPSecConnectorValidator.java:64-74"`

	// TunnelNo is the tunnel's position, set on every save
	// (AdvancedIPSecConnectorServiceImpl.java:509-510).
	TunnelNo *int64 `json:"tunnelNo,omitempty" ak:"readonly"`
}

// ConnectorIPsecAdvCustomerEnd is a tunnel's customer end. With
// overlayIpReservationId set, the backend replaces overlayIp with the
// reservation's address (AdvancedIPSecConnectorServiceImpl.java:513-527).
type ConnectorIPsecAdvCustomerEnd struct {
	OverlayIP              *string `json:"overlayIp,omitempty"`
	OverlayIPReservationID *string `json:"overlayIpReservationId,omitempty"`

	PublicIP *string `json:"publicIp,omitempty" ak:"readonly"`
}

// ConnectorIPsecAdvCXPEnd is a tunnel's CXP end: two IP reservation ids; the
// addresses are filled in from them (AdvancedIPSecConnectorServiceImpl.java:529-536).
type ConnectorIPsecAdvCXPEnd struct {
	PublicIPReservationID  string `json:"publicIpReservationId" ak:"required,src=IPSecTunnel.java:109-110"`
	OverlayIPReservationID string `json:"overlayIpReservationId" ak:"required,src=IPSecTunnel.java:116-117"`

	PublicIP  *string `json:"publicIp,omitempty" ak:"readonly"`
	OverlayIP *string `json:"overlayIp,omitempty" ak:"readonly"`
	NodeID    *string `json:"nodeId,omitempty" ak:"readonly"`
}

// connectorIPsecAdvDescriptor: GET, PUT and DELETE answer a missing id with
// 404 AK000003, as every connector does. gateways identifies the family in
// the GET body and ADV_IP_SEC in the GET /connectors row, which also carries
// the state the D13 guard reads. Gateways and their tunnels merge by id: a
// patch must list every stored gateway and tunnel id, and the stored key of
// each listed tunnel is resent. With the async flag on, TPS skips its own
// unknown-tunnel-id check and keeps an unknown id as a new tunnel with a new
// key (AdvancedIPSecConnectorServiceImpl.java:496-505), so the id match here
// is the only refusal of an unknown id. With
// RELEASE-ASYNC-NETWORK-ENTITY-VALIDATION-PHASE-1 on, the portal saves with
// ?async=true (karrawa src/_api/ipsec.js:27-31, :39-48).
var connectorIPsecAdvDescriptor = Descriptor{
	Collection:    "adv-ipsec-connectors",
	FamilyType:    FamilyCheck{Field: "gateways"},
	ConnectorType: "ADV_IP_SEC",
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
	IDMatchedArrays:    []string{"gateways", "gateways[].tunnels"},
	StateGuard:         StateGuardD13,
	Service:            ServiceTPS,
	CrossCheck:         checkConnectorIPsecAdv,
	AsyncCapable:       true,
	Flags: []FlagRule{
		{Flag: "SCALE-GROUPS-FOR-USER-AND-SITES", Effect: FlagIncludeField, Field: "scaleGroupId", UISource: "src/views/network/provision/objectConfig/ipsecConfig/utils.js:582|:675-677"},
		{Flag: "CXP-FAILOVER-REGION", Effect: FlagIncludeField, Field: "enabled", UISource: "src/views/network/provision/objectConfig/ipsecConfig/utils.js:581|:713-714"},
	},
}

// checkConnectorIPsecAdv checks the rules of AdvancedIPSecConnectorValidator
// that span fields: the routing side of a ROUTE_BASED connector (:91-127),
// advanced and profileId together with the identity values advanced needs
// (:33-53), a 0.0.0.0 gateway with tunnels that carry advanced (:24-28,
// :55-58), customized ends on every tunnel or on none (:64-74), the tunnel
// count of each gateway (:76-89), and one ACTIVE and at most one STANDBY
// gateway, an unset haMode being ACTIVE (:133-145), and unique gateway names
// (:147-153). It also refuses a name
// starting with ALK- (ConnectorServiceImpl.java:949-952).
func checkConnectorIPsecAdv(m map[string]any) error {
	var errs []error
	if name, _ := m["name"].(string); strings.HasPrefix(strings.ToUpper(name), "ALK-") {
		errs = append(errs, errors.New("name must not start with ALK-"))
	}
	if ro, ok := m["routingOptions"].(map[string]any); ok {
		st, _ := ro["staticRouting"].(map[string]any)
		dy, _ := ro["dynamicRouting"].(map[string]any)
		switch {
		case st == nil && dy == nil:
			errs = append(errs, errors.New("routingOptions must hold staticRouting or dynamicRouting"))
		case st != nil && st["prefixListId"] == nil:
			errs = append(errs, errors.New("routingOptions.staticRouting.prefixListId must be set"))
		case dy != nil && (dy["customerGwAsn"] == nil || dy["customerGwAsn"] == ""):
			errs = append(errs, errors.New("routingOptions.dynamicRouting.customerGwAsn must be set"))
		case st != nil && dy != nil && st["availability"] != nil && dy["availability"] != nil && st["availability"] != dy["availability"]:
			errs = append(errs, errors.New("routingOptions: staticRouting and dynamicRouting availability must be the same"))
		}
	}
	perGateway, hasCount := asInt(m["tunnelsPerGateway"])
	gateways, _ := m["gateways"].([]any)
	var customized, plain, active, standby int
	names := map[any]bool{}
	for gi, g := range gateways {
		gw, _ := g.(map[string]any)
		label := fmt.Sprintf("gateways[%d]", gi)
		if names[gw["name"]] {
			errs = append(errs, fmt.Errorf("%s: gateway name %v is used twice", label, gw["name"]))
		}
		names[gw["name"]] = true
		tunnels, _ := gw["tunnels"].([]any)
		switch gw["haMode"] {
		case nil, "ACTIVE":
			active++
		case "STANDBY":
			standby++
		}
		anyIP := gw["customerGwIp"] == "0.0.0.0"
		if anyIP && len(tunnels) == 0 {
			errs = append(errs, fmt.Errorf("%s: a gateway with customerGwIp 0.0.0.0 needs tunnels", label))
		}
		if len(tunnels) > 0 && hasCount && int64(len(tunnels)) != perGateway {
			errs = append(errs, fmt.Errorf("%s has %d tunnels; tunnelsPerGateway is %d", label, len(tunnels), perGateway))
		}
		if len(tunnels) == 0 {
			plain++
		}
		for ti, e := range tunnels {
			tn, _ := e.(map[string]any)
			tl := fmt.Sprintf("%s.tunnels[%d]", label, ti)
			adv, _ := tn["advanced"].(map[string]any)
			switch {
			case adv == nil && tn["profileId"] != nil:
				errs = append(errs, fmt.Errorf("%s: advanced must be set when profileId is set", tl))
			case adv != nil && tn["profileId"] == nil:
				errs = append(errs, fmt.Errorf("%s: profileId must be set when advanced is set", tl))
			}
			if adv != nil && (adv["ikeVersion"] == nil || adv["remoteAuthType"] == nil || adv["remoteAuthValue"] == nil) {
				errs = append(errs, fmt.Errorf("%s: advanced needs ikeVersion, remoteAuthType and remoteAuthValue", tl))
			}
			if anyIP && adv == nil {
				errs = append(errs, fmt.Errorf("%s: advanced must be set when customerGwIp is 0.0.0.0", tl))
			}
			hasCXP, hasCustomer := tn["cxpEnd"] != nil, tn["customerEnd"] != nil
			switch {
			case hasCXP && hasCustomer:
				customized++
			case !hasCXP && !hasCustomer:
				plain++
			default:
				errs = append(errs, fmt.Errorf("%s: cxpEnd and customerEnd are set together", tl))
			}
		}
	}
	if customized > 0 && plain > 0 {
		errs = append(errs, errors.New("gateways: every tunnel carries cxpEnd and customerEnd, or none does"))
	}
	if len(gateways) > 0 && active == 0 {
		errs = append(errs, errors.New("gateways: at least one gateway must have haMode ACTIVE"))
	}
	if standby > 1 {
		errs = append(errs, errors.New("gateways: at most one gateway can have haMode STANDBY"))
	}
	return errors.Join(errs...)
}

// asInt reads a JSON number held as float64 or json.Number.
func asInt(v any) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case json.Number:
		i, err := n.Int64()
		return i, err == nil
	}
	return 0, false
}

// ConnectorIPsecAdvDescriptor returns the advanced IPsec connector write descriptor.
func ConnectorIPsecAdvDescriptor() Descriptor { return connectorIPsecAdvDescriptor.clone() }

// NewConnectorIPsecAdvCore returns the write core for advanced IPsec connectors.
func NewConnectorIPsecAdvCore(ac *AlkiraClient) (*Core[ConnectorIPsecAdvConfig], error) {
	return NewCore[ConnectorIPsecAdvConfig](ac, connectorIPsecAdvDescriptor)
}

// CreateConnectorIPsecAdv creates an advanced IPsec connector. Tunnel
// pre-shared keys are optional, as in the portal: a tunnel without a key
// gets one the backend generates.
func (ac *AlkiraClient) CreateConnectorIPsecAdv(ctx context.Context, c *ConnectorIPsecAdvConfig) (*ConnectorIPsecAdvConfig, *WriteResult, error) {
	core, err := NewConnectorIPsecAdvCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, c, WriteOptions{})
}

// GetConnectorIPsecAdv reads one advanced IPsec connector and its raw body.
func (ac *AlkiraClient) GetConnectorIPsecAdv(ctx context.Context, id string) (*ConnectorIPsecAdvConfig, []byte, error) {
	core, err := NewConnectorIPsecAdvCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

// UpdateConnectorIPsecAdv applies a merge patch to an advanced IPsec
// connector; the stored tunnel keys and BGP key are resent unchanged.
func (ac *AlkiraClient) UpdateConnectorIPsecAdv(ctx context.Context, id string, patch json.RawMessage) (*ConnectorIPsecAdvConfig, *WriteResult, error) {
	core, err := NewConnectorIPsecAdvCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

// DeleteConnectorIPsecAdv deletes an advanced IPsec connector; a deployed
// one is marked for deletion until the next provision.
func (ac *AlkiraClient) DeleteConnectorIPsecAdv(ctx context.Context, id string) (*WriteResult, error) {
	core, err := NewConnectorIPsecAdvCore(ac)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

// ListConnectorIPsecAdvPaged reads every advanced IPsec connector through paged reads.
func (ac *AlkiraClient) ListConnectorIPsecAdvPaged(ctx context.Context) ([]ConnectorIPsecAdvConfig, error) {
	core, err := NewConnectorIPsecAdvCore(ac)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
