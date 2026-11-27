// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ConnectorAzureExpressRouteConfig is the write model of an Azure
// ExpressRoute connector (TPS model/azureexpressroute/AzureExpressRouteConnector.java;
// GET /connectors reports its type as EXPRESS_ROUTE). The fields follow the
// portal save body (karrawa azureExpressConfig/add/index.js:420-489); optional
// fields are pointers or omitempty slices, so a zero value is never sent.
// The backend reads each new or PENDING instance's circuit from Azure while
// it saves (AzureExpressRouteConnectorServiceImpl.java:290-293), and refuses
// creates and updates while the tenant feature EXPRESS-ROUTE-CONNECTOR-DISABLE
// is on (:164-165, :1324-1331).
type ConnectorAzureExpressRouteConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=AzureExpressRouteConnector.java:80-84"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,charset=description,src=AzureExpressRouteConnector.java:86-88"`
	CXP         string      `json:"cxp" ak:"required,maxlen=128,trimmed,src=AzureExpressRouteConnector.java:93-96"`
	Group       *string     `json:"group,omitempty" ak:"maxlen=256,trimmed,charset=name,src=AzureExpressRouteConnector.java:101-104"`
	// Size defaults to LARGE in the backend; XSMALL, 5XSMALL, 4LARGE and
	// 20LARGE and above are refused (AzureExpressRouteConnector.java:114-117).
	Size        *string `json:"size,omitempty" ak:"default=LARGE,enum=SMALL|MEDIUM|LARGE|2LARGE|5LARGE|10LARGE,src=AzureExpressRouteConnector.java:114-117|Defaults.java:148-150"`
	BillingTags []int64 `json:"billingTags,omitempty" ak:"flag=BILLING-TAGS"`
	// Enabled is sent as true by the portal unless changed (azureExpressConfig/utils.js:85).
	Enabled *bool `json:"enabled,omitempty" ak:"default=true,src=utils.js:85"`
	// VHubPrefix is a /23 that every ExpressRoute connector on the CXP
	// shares; it cannot change once the connector is provisioned
	// (AzureExpressRouteConnectorServiceImpl.java:175-187).
	VHubPrefix string `json:"vhubPrefix" ak:"required,src=AzureExpressRouteConnector.java:130-132"`
	// TunnelProtocol defaults to VXLAN_GPE (AzureExpressRouteConnectorServiceImpl.java:287-289);
	// the backend refuses a switch between VXLAN_GPE and the others once
	// the connector is provisioned.
	TunnelProtocol *string                                    `json:"tunnelProtocol,omitempty" ak:"default=VXLAN_GPE,enum=VXLAN|VXLAN_GPE|IPSEC,src=Defaults.java:1396-1398|utils.js:92"`
	Instances      []ConnectorAzureExpressRouteInstanceConfig `json:"instances" ak:"required,minitems=1,src=AzureExpressRouteConnector.java:134-138"`
	// SegmentOptions holds one entry per segment, keyed by segmentName.
	SegmentOptions []ConnectorAzureExpressRouteSegmentConfig `json:"segmentOptions" ak:"required,minitems=1,distinct=segmentName,src=AzureExpressRouteConnector.java:140-144|AzureExpressRouteConnectorValidator.java:156-164"`

	GroupID         *int64 `json:"groupId,omitempty" ak:"readonly"`
	ImplicitGroupID *int64 `json:"implicitGroupId,omitempty" ak:"readonly"`
	Primary         *bool  `json:"primary,omitempty" ak:"readonly"`
	LastEnabled     *int64 `json:"lastEnabled,omitempty" ak:"readonly"`
	// State and PreviousState come from the GET /connectors row.
	State         string `json:"state,omitempty" ak:"readonly"`
	PreviousState string `json:"previousState,omitempty" ak:"readonly"`
}

// ConnectorAzureExpressRouteInstanceConfig is one connector instance
// (AzureExpressRouteInstance.java). An update must list the id of every
// stored instance; one without an id is a new instance.
type ConnectorAzureExpressRouteInstanceConfig struct {
	ID   json.Number `json:"id,omitempty"`
	Name string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=AzureExpressRouteInstance.java:35-39"`
	// ExpressRouteCircuitID is the Azure resource id of the circuit; it
	// cannot change once the instance is provisioned.
	ExpressRouteCircuitID string `json:"expressRouteCircuitId" ak:"required,trimmed,src=AzureExpressRouteInstance.java:41-43"`
	// CredentialID is an Azure credential that can read the circuit.
	CredentialID    string `json:"credentialId" ak:"required,src=AzureExpressRouteInstance.java:54-55"`
	RedundantRouter *bool  `json:"redundantRouter,omitempty"`
	// LoopbackSubnet is a /26 the backend reserves on save
	// (AzureExpressRouteConnectorServiceImpl.java:301-309).
	LoopbackSubnet string `json:"loopbackSubnet" ak:"required,src=AzureExpressRouteInstance.java:51-52|AzureExpressRouteConnectorServiceImpl.java:301-309"`
	// GatewayMacAddresses and Vnis are set with tunnelProtocol VXLAN only.
	GatewayMacAddresses []string `json:"gatewayMacAddresses,omitempty" ak:"maxitems=2,src=AzureExpressRouteConnectorValidator.java:63-80"`
	Vnis                []int64  `json:"vnis,omitempty" ak:"maxitems=2,range=16773023..16777215,src=AzureExpressRouteInstance.java:61-64"`
	// SegmentOptions holds the customer gateways per segment; set with
	// tunnelProtocol IPSEC only.
	SegmentOptions []ConnectorAzureExpressRouteInstanceSegment `json:"segmentOptions,omitempty"`

	ExpressRouteCircuitName *string `json:"expressRouteCircuitName,omitempty" ak:"readonly"`
}

// ConnectorAzureExpressRouteInstanceSegment is one instance segmentOptions
// entry (AzureExpressRouteInstanceSegmentOptions.java).
type ConnectorAzureExpressRouteInstanceSegment struct {
	SegmentName      string                              `json:"segmentName" ak:"required,maxlen=256,trimmed,src=AzureExpressRouteInstanceSegmentOptions.java:25-28"`
	CustomerGateways []ConnectorAzureExpressRouteGateway `json:"customerGateways,omitempty" ak:"maxitems=2,src=AzureExpressRouteInstanceSegmentOptions.java:32-33"`

	SegmentID *int64 `json:"segmentId,omitempty" ak:"readonly"`
}

// ConnectorAzureExpressRouteGateway is one customer gateway
// (AzureExpressRouteCustomerGateway.java); gateway names are unique per connector.
type ConnectorAzureExpressRouteGateway struct {
	Name    string                             `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=AzureExpressRouteCustomerGateway.java:27-31"`
	Tunnels []ConnectorAzureExpressRouteTunnel `json:"tunnels" ak:"required,minitems=1,src=AzureExpressRouteCustomerGateway.java:35-38"`

	GatewayNo *int64 `json:"gatewayNo,omitempty" ak:"readonly"`
}

// ConnectorAzureExpressRouteTunnel is one IPsec tunnel
// (AzureExpressRouteInstanceTunnel.java). The backend gives a new tunnel its
// id; an update keeps a stored tunnel by its id. initiator false needs
// ikeVersion, profileId, remoteAuthType and remoteAuthValue, and those four
// are set together or not at all (AzureExpressRouteConnectorValidator.java:178-195).
// The pre-shared key is secret: a create or patch cannot set it, and an
// update resends the stored value.
type ConnectorAzureExpressRouteTunnel struct {
	ID        *string `json:"id,omitempty" ak:"maxlen=256,trimmed,charset=name,src=AzureExpressRouteInstanceTunnel.java:33-35"`
	Name      string  `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=AzureExpressRouteInstanceTunnel.java:27-31"`
	Initiator *bool   `json:"initiator,omitempty"`
	ProfileID *int64  `json:"profileId,omitempty"`
	// IKEVersion accepts IKEv2 only (AzureExpressRouteConnectorValidator.java:131-134).
	IKEVersion *string `json:"ikeVersion,omitempty" ak:"enum=IKEv2,src=AzureExpressRouteInstanceTunnel.java:63-64"`
	// RemoteAuthType accepts FQDN only, and RemoteAuthValue is that IKE
	// identity, not a key (AzureExpressRouteConnectorValidator.java:123-130).
	RemoteAuthType  *string `json:"remoteAuthType,omitempty" ak:"enum=FQDN,src=IPsecConnectorInstanceAdvancedAttributes.java:447-449"`
	RemoteAuthValue *string `json:"remoteAuthValue,omitempty" ak:"maxlen=256,trimmed,src=AzureExpressRouteInstanceTunnel.java:53-55"`
	PreSharedKey    *string `json:"preSharedKey,omitempty" ak:"secret"`

	TunnelNo   *int64 `json:"tunnelNo,omitempty" ak:"readonly"`
	TunnelAzNo *int64 `json:"tunnelAzNo,omitempty" ak:"readonly"`
}

// ConnectorAzureExpressRouteSegmentConfig is one connector segmentOptions entry
// (AzureExpressRouteSegmentOptions.java). The ASNs are numbers or numeric
// strings. The backend ignores localAsn while the tenant feature
// RELEASE-EXPLICIT-LOCAL-ASN-FOR-BRANCH-CONNECTOR is off
// (AzureExpressRouteConnectorServiceImpl.java:1172-1179).
type ConnectorAzureExpressRouteSegmentConfig struct {
	SegmentName           string `json:"segmentName" ak:"required,maxlen=256,trimmed,src=AzureExpressRouteSegmentOptions.java:30-33"`
	CustomerASN           any    `json:"customerAsn" ak:"required"`
	LocalASN              any    `json:"localAsn,omitempty"`
	DisableInternetExit   *bool  `json:"disableInternetExit,omitempty"`
	AdvertiseOnPremRoutes *bool  `json:"advertiseOnPremRoutes,omitempty"`

	SegmentID *int64 `json:"segmentId,omitempty" ak:"readonly"`
}

// connectorAzureExpressRouteDescriptor: GET, PUT and DELETE answer a missing
// id with 404 AK000003, as every connector does. vhubPrefix identifies the
// family in the GET body and EXPRESS_ROUTE in the GET /connectors row, which
// also carries the state the D13 guard reads. Instances merge by id; the
// segment options, gateways and tunnels below them merge by segment name,
// gateway name and tunnel id, so stored tunnel ids and keys are resent.
var connectorAzureExpressRouteDescriptor = Descriptor{
	Collection:    "azure-express-route-connectors",
	V1Path:        true,
	FamilyType:    FamilyCheck{Field: "vhubPrefix"},
	ConnectorType: "EXPRESS_ROUTE",
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
	KeyMatchedArrays: map[string]string{
		"segmentOptions":                                          "segmentName",
		"instances[].segmentOptions":                              "segmentName",
		"instances[].segmentOptions[].customerGateways":           "name",
		"instances[].segmentOptions[].customerGateways[].tunnels": "id",
	},
	StateGuard: StateGuardD13,
	Service:    ServiceTPS,
	CrossCheck: checkConnectorAzureExpressRoute,
	Flags: []FlagRule{
		{Flag: "BILLING-TAGS", Effect: FlagIncludeField, Field: "billingTags", UISource: "src/views/network/provision/objectConfig/azureExpressConfig/add/index.js:176-178|:426"},
	},
}

// checkConnectorAzureExpressRoute mirrors the backend's cross-field rules
// (AzureExpressRouteConnectorValidator.java): circuits unique across
// instances (:48-54); gatewayMacAddresses and vnis not set with VXLAN_GPE,
// and with VXLAN one per router (:63-99); with IPSEC, each instance holds
// the connector's segments with one customer gateway per router
// (:101-113, :201-204), gateway and tunnel names are unique per connector
// (:145-154) and the tunnel IKE settings go together (:178-195). The router
// count is checked only when redundantRouter is set, as the backend does.
func checkConnectorAzureExpressRoute(m map[string]any) error {
	var errs []error
	if name, _ := m["name"].(string); strings.HasPrefix(strings.ToUpper(name), "ALK-") {
		errs = append(errs, errors.New("name must not start with ALK-"))
	}
	proto, _ := m["tunnelProtocol"].(string)
	gpe := proto == "" || proto == "VXLAN_GPE"
	segs := map[string]bool{}
	for _, e := range erObjects(m["segmentOptions"]) {
		if n, _ := e["segmentName"].(string); n != "" {
			segs[n] = true
		}
	}
	circuits, gateways, tunnels := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, inst := range erObjects(m["instances"]) {
		name := inst["name"]
		if c, _ := inst["expressRouteCircuitId"].(string); c != "" {
			if circuits[c] {
				errs = append(errs, fmt.Errorf("instance %v: expressRouteCircuitId %s is used by another instance", name, c))
			}
			circuits[c] = true
		}
		routers := 0 // unknown while redundantRouter is unset
		switch inst["redundantRouter"] {
		case false:
			routers = 1
		case true:
			routers = 2
		}
		macs, _ := inst["gatewayMacAddresses"].([]any)
		vnis, _ := inst["vnis"].([]any)
		switch {
		case gpe && (len(macs) > 0 || len(vnis) > 0):
			errs = append(errs, fmt.Errorf("instance %v: gatewayMacAddresses and vnis are set with tunnelProtocol VXLAN only", name))
		case proto == "VXLAN" && len(macs) == 0:
			errs = append(errs, fmt.Errorf("instance %v: gatewayMacAddresses must be set with tunnelProtocol VXLAN", name))
		case proto == "VXLAN" && routers > 0 && len(macs) != routers:
			errs = append(errs, fmt.Errorf("instance %v: gatewayMacAddresses must hold %d address(es), one per router", name, routers))
		case proto == "VXLAN" && routers > 0 && len(vnis) > 0 && len(vnis) != routers:
			errs = append(errs, fmt.Errorf("instance %v: vnis must hold 0 or %d value(s), one per router", name, routers))
		}
		if proto != "IPSEC" {
			continue
		}
		isegs := erObjects(inst["segmentOptions"])
		if len(isegs) == 0 {
			errs = append(errs, fmt.Errorf("instance %v: segmentOptions must be set with tunnelProtocol IPSEC", name))
		}
		got := map[string]bool{}
		for _, so := range isegs {
			sn, _ := so["segmentName"].(string)
			got[sn] = true
			gws := erObjects(so["customerGateways"])
			if (routers > 0 && len(gws) != routers) || len(gws) == 0 {
				errs = append(errs, fmt.Errorf("instance %v segment %s: customerGateways must hold one gateway per router", name, sn))
			}
			for _, gw := range gws {
				gn, _ := gw["name"].(string)
				if gateways[gn] {
					errs = append(errs, fmt.Errorf("customer gateway name %q is used twice", gn))
				}
				gateways[gn] = true
				for _, tun := range erObjects(gw["tunnels"]) {
					tn, _ := tun["name"].(string)
					if tunnels[tn] {
						errs = append(errs, fmt.Errorf("tunnel name %q is used twice", tn))
					}
					tunnels[tn] = true
					errs = append(errs, checkERTunnel(tn, tun))
				}
			}
		}
		if len(segs) > 0 && !sameKeys(got, segs) {
			errs = append(errs, fmt.Errorf("instance %v: segmentOptions must name the same segments as the connector segmentOptions", name))
		}
	}
	return errors.Join(errs...)
}

// checkERTunnel: ikeVersion, profileId, remoteAuthType and remoteAuthValue
// go together, initiator must be set with them, and initiator false needs
// them (AzureExpressRouteConnectorValidator.java:178-195).
func checkERTunnel(name string, t map[string]any) error {
	n := 0
	for _, k := range []string{"ikeVersion", "profileId", "remoteAuthType", "remoteAuthValue"} {
		if v := t[k]; v != nil && v != "" {
			n++
		}
	}
	switch {
	case n > 0 && t["initiator"] == nil:
		return fmt.Errorf("tunnel %q: initiator must be set with ikeVersion, profileId, remoteAuthType or remoteAuthValue", name)
	case n > 0 && n < 4:
		return fmt.Errorf("tunnel %q: set all or none of ikeVersion, profileId, remoteAuthType and remoteAuthValue", name)
	case n == 0 && t["initiator"] == false:
		return fmt.Errorf("tunnel %q: initiator false needs ikeVersion, profileId, remoteAuthType and remoteAuthValue", name)
	}
	return nil
}

// erObjects returns the object elements of a decoded JSON array.
func erObjects(v any) []map[string]any {
	a, _ := v.([]any)
	out := make([]map[string]any, 0, len(a))
	for _, e := range a {
		if m, ok := e.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

// sameKeys reports whether a and b hold the same keys.
func sameKeys(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

// ConnectorAzureExpressRouteDescriptor returns the Azure ExpressRoute connector write descriptor.
func ConnectorAzureExpressRouteDescriptor() Descriptor {
	return connectorAzureExpressRouteDescriptor.clone()
}

// NewConnectorAzureExpressRouteCore returns the write core for Azure ExpressRoute connectors.
func NewConnectorAzureExpressRouteCore(ac *AlkiraClient) (*Core[ConnectorAzureExpressRouteConfig], error) {
	return NewCore[ConnectorAzureExpressRouteConfig](ac, connectorAzureExpressRouteDescriptor)
}

// CreateConnectorAzureExpressRoute creates an Azure ExpressRoute connector.
func (ac *AlkiraClient) CreateConnectorAzureExpressRoute(ctx context.Context, c *ConnectorAzureExpressRouteConfig) (*ConnectorAzureExpressRouteConfig, *WriteResult, error) {
	core, err := NewConnectorAzureExpressRouteCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, c, WriteOptions{})
}

// GetConnectorAzureExpressRoute reads one Azure ExpressRoute connector and its raw body.
func (ac *AlkiraClient) GetConnectorAzureExpressRoute(ctx context.Context, id string) (*ConnectorAzureExpressRouteConfig, []byte, error) {
	core, err := NewConnectorAzureExpressRouteCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

// UpdateConnectorAzureExpressRoute applies a merge patch to an Azure
// ExpressRoute connector; stored tunnel keys are resent unchanged.
func (ac *AlkiraClient) UpdateConnectorAzureExpressRoute(ctx context.Context, id string, patch json.RawMessage) (*ConnectorAzureExpressRouteConfig, *WriteResult, error) {
	core, err := NewConnectorAzureExpressRouteCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

// DeleteConnectorAzureExpressRoute deletes an Azure ExpressRoute connector;
// a deployed one is marked for deletion until the next provision.
func (ac *AlkiraClient) DeleteConnectorAzureExpressRoute(ctx context.Context, id string) (*WriteResult, error) {
	core, err := NewConnectorAzureExpressRouteCore(ac)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

// ListConnectorAzureExpressRoutePaged reads every Azure ExpressRoute connector through paged reads.
func (ac *AlkiraClient) ListConnectorAzureExpressRoutePaged(ctx context.Context) ([]ConnectorAzureExpressRouteConfig, error) {
	core, err := NewConnectorAzureExpressRouteCore(ac)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
