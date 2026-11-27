// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"strings"
)

// ConnectorAwsDirectConnectConfig is the write model of an AWS Direct
// Connect connector (TPS model/DirectConnectConnector.java; GET /connectors
// reports its type as DIRECT_CONNECT). The fields follow the portal save
// body (karrawa awsDirectConfig/AddConnect/index.js:485-579); optional
// fields are pointers or omitempty slices, so a zero value is never sent.
// The instance BGP keys are secret: the GET returns them, a patch cannot set
// them, and an update resends the stored values.
type ConnectorAwsDirectConnectConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=DirectConnectConnector.java:66-70"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,charset=description,src=DirectConnectConnector.java:72-74"`
	CXP         string      `json:"cxp" ak:"required,maxlen=128,trimmed,src=DirectConnectConnector.java:76-79"`
	Group       *string     `json:"group,omitempty" ak:"maxlen=256,trimmed,charset=name,src=DirectConnectConnector.java:81-84"`
	Size        string      `json:"size" ak:"required,trimmed,enum=XSMALL|5XSMALL|SMALL|MEDIUM|LARGE|2LARGE|4LARGE|5LARGE|10LARGE|20LARGE,src=DirectConnectConnector.java:94-97|Defaults.java:148-150"`
	// Instances merge by id: an update must list every stored instance id,
	// and the stored BGP keys of each listed instance are resent.
	Instances   []ConnectorAwsDirectConnectInstanceConfig `json:"instances" ak:"required,minitems=1,distinct=name|connectionId,src=DirectConnectConnector.java:99-102|DirectConnectConnectorValidator.java:160-167|DirectConnectConnectorServiceImpl.java:366-371"`
	BillingTags []int64                                   `json:"billingTags,omitempty" ak:"flag=BILLING-TAGS"`
	// Enabled is sent as true by the portal unless changed (AddConnect/index.js:89, :563).
	Enabled *bool `json:"enabled,omitempty" ak:"default=true,src=AddConnect/index.js:89|:563"`
	// TunnelProtocol defaults to GRE (DirectConnectConnector.java:114-115) and
	// cannot change once the connector is provisioned
	// (DirectConnectConnectorServiceImpl.java:279-283).
	TunnelProtocol *string `json:"tunnelProtocol,omitempty" ak:"enum=GRE|VXLAN_GPE|VXLAN|IPSEC,src=DirectConnectConnectorServiceImpl.java:82|:223-225"`
	// ScaleGroupID also needs SCALE-GROUPS-FOR-USER-AND-SITES in the backend,
	// which refuses it with 403 otherwise (DirectConnectConnectorServiceImpl.java:607-610).
	ScaleGroupID *string `json:"scaleGroupId,omitempty" ak:"flag=SUPPORT-SCALE-GROUP-FOR-DX-CONNECTOR"`
	// LoopbackPrefixes are /26 blocks (DirectConnectConnector.java:119-121).
	LoopbackPrefixes []string `json:"loopbackPrefixes,omitempty" ak:"flag=RELEASE-CUSTOM-SCALE-OPTIONS"`

	GroupID         *int64          `json:"groupId,omitempty" ak:"readonly"`
	ImplicitGroupID *int64          `json:"implicitGroupId,omitempty" ak:"readonly"`
	LastEnabled     *int64          `json:"lastEnabled,omitempty" ak:"readonly"`
	ConfigStatus    json.RawMessage `json:"configStatus,omitempty" ak:"readonly"`
	// State and PreviousState come from the GET /connectors row.
	State         string `json:"state,omitempty" ak:"readonly"`
	PreviousState string `json:"previousState,omitempty" ak:"readonly"`
}

// ConnectorAwsDirectConnectInstanceConfig is one Direct Connect connection of the
// connector (TPS model/DirectConnectInstance.java). The ASNs are numbers or
// numeric strings (DirectConnectInstance.java:57, :63).
type ConnectorAwsDirectConnectInstanceConfig struct {
	ID   json.Number `json:"id,omitempty"`
	Name string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=DirectConnectInstance.java:41-45"`
	// ConnectionID is an AWS Direct Connect connection the credential can
	// read: the backend reads it from AWS on every save
	// (DirectConnectConnectorServiceImpl.java:374-419).
	ConnectionID string `json:"connectionId" ak:"required,trimmed,src=DirectConnectInstance.java:47-49"`
	DCGatewayASN any    `json:"dcGatewayAsn" ak:"required,src=DirectConnectInstance.java:54-57"`
	UnderlayASN  any    `json:"underlayAsn" ak:"required,src=DirectConnectInstance.java:59-63"`
	// UnderlayPrefix is a /30; with it, onPremUnderlayIp and awsUnderlayIp
	// are both set (DirectConnectConnectorServiceImpl.java:341-355).
	UnderlayPrefix   *string `json:"underlayPrefix,omitempty"`
	OnPremUnderlayIP *string `json:"onPremUnderlayIp,omitempty"`
	AwsUnderlayIP    *string `json:"awsUnderlayIp,omitempty"`
	BGPAuthKey       *string `json:"bgpAuthKey,omitempty" ak:"secret"`
	BGPAuthKeyAlkira *string `json:"bgpAuthKeyAlkira,omitempty" ak:"secret"`
	// VLAN, when unset, is read from the AWS connection
	// (DirectConnectConnectorServiceImpl.java:400-405).
	VLAN           *int64 `json:"vlan,omitempty" ak:"range=1..4094,src=DirectConnectInstance.java:84-85"`
	CustomerRegion string `json:"customerRegion" ak:"required,src=DirectConnectInstance.java:87-89"`
	// CredentialID is the id of a stored AWS credential.
	CredentialID   string                                   `json:"credentialId" ak:"required,trimmed,src=DirectConnectInstance.java:91-93"`
	SegmentOptions []ConnectorAwsDirectConnectSegmentConfig `json:"segmentOptions" ak:"required,minitems=1,maxitems=10,distinct=segmentName,src=DirectConnectInstance.java:95-99|DirectConnectConnectorValidator.java:150-158"`
	// GatewayMacAddress and VNI are set with tunnelProtocol VXLAN only
	// (DirectConnectConnectorServiceImpl.java:329-339).
	GatewayMacAddress *string `json:"gatewayMacAddress,omitempty"`
	VNI               *int64  `json:"vni,omitempty" ak:"range=16773023..16777215,src=DirectConnectInstance.java:104-105"`

	ConnectionName *string `json:"connectionName,omitempty" ak:"readonly"`
}

// ConnectorAwsDirectConnectSegmentConfig is one segmentOptions entry of an
// instance (TPS model/DirectConnectSegmentOptions.java). With tunnelProtocol
// IPSEC the backend allocates the loopback IPs and needs the two counts;
// otherwise the three loopback IPs are set and the counts are dropped
// (DirectConnectConnectorValidator.java:45-131).
type ConnectorAwsDirectConnectSegmentConfig struct {
	SegmentName string `json:"segmentName" ak:"required,maxlen=256,trimmed,src=DirectConnectSegmentOptions.java:43-46"`
	CustomerASN any    `json:"customerAsn" ak:"required,src=DirectConnectSegmentOptions.java:48-50"`
	// LocalASN is kept only while RELEASE-EXPLICIT-LOCAL-ASN-FOR-BRANCH-CONNECTOR
	// is on (DirectConnectConnectorServiceImpl.java:932-936).
	LocalASN                         any     `json:"localAsn,omitempty"`
	CustomerLoopbackIP               *string `json:"customerLoopbackIp,omitempty"`
	AlkLoopbackIP1                   *string `json:"alkLoopbackIp1,omitempty"`
	AlkLoopbackIP2                   *string `json:"alkLoopbackIp2,omitempty"`
	LoopbackSubnet                   string  `json:"loopbackSubnet" ak:"required,src=DirectConnectSegmentOptions.java:61-63"`
	DisableInternetExit              *bool   `json:"disableInternetExit,omitempty"`
	AdvertiseOnPremRoutes            *bool   `json:"advertiseOnPremRoutes,omitempty"`
	NumOfCustomerLoopbackIPs         *int64  `json:"numOfCustomerLoopbackIps,omitempty" ak:"range=1..4,src=DirectConnectConnectorValidator.java:71-76"`
	TunnelCountPerCustomerLoopbackIP *int64  `json:"tunnelCountPerCustomerLoopbackIp,omitempty" ak:"enum=2|4,src=DirectConnectConnectorValidator.java:59-69"`

	GeneratedTunnelLoopbackIPs json.RawMessage `json:"generatedTunnelLoopbackIps,omitempty" ak:"readonly"`
}

// connectorAwsDirectConnectDescriptor: GET, PUT and DELETE answer a missing
// id with 404 AK000003, as every connector does. instances identifies the
// family in the GET body and DIRECT_CONNECT in the GET /connectors row,
// which also carries the state the D13 guard reads. Instances merge by id.
// The backend reads every instance's connection from AWS on each save, so a
// create or update needs a Direct Connect connection the credential can see.
var connectorAwsDirectConnectDescriptor = Descriptor{
	Collection:    "directconnectconnectors",
	FamilyType:    FamilyCheck{Field: "instances"},
	ConnectorType: "DIRECT_CONNECT",
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
	CrossCheck:         checkConnectorAwsDirectConnect,
	Flags: []FlagRule{
		{Flag: "BILLING-TAGS", Effect: FlagIncludeField, Field: "billingTags", UISource: "src/views/network/provision/objectConfig/awsDirectConfig/AddConnect/index.js:105-106|:565"},
		{Flag: "SUPPORT-SCALE-GROUP-FOR-DX-CONNECTOR", Effect: FlagIncludeField, Field: "scaleGroupId", UISource: "src/views/network/provision/objectConfig/awsDirectConfig/AddConnect/index.js:92-94|:567-569"},
		{Flag: "RELEASE-CUSTOM-SCALE-OPTIONS", Effect: FlagIncludeField, Field: "loopbackPrefixes", UISource: "src/views/network/provision/objectConfig/awsDirectConfig/AddConnect/index.js:96-98|:570-575"},
	},
}

// checkConnectorAwsDirectConnect refuses, with no request, the rules the
// backend checks across fields: a name starting with ALK- (the implicit
// group name, GroupServiceImpl.java:1205-1208); 20LARGE without a VXLAN
// protocol (DirectConnectConnectorServiceImpl.java:217-220); vni and
// gatewayMacAddress set exactly when tunnelProtocol is VXLAN, and one
// segment per instance with VXLAN (:329-339; DirectConnectConnectorValidator.java:36-43);
// the underlay prefix and both underlay IPs set together, the IPs inside the
// prefix (:341-355); one dcGatewayAsn across instances, different from each
// underlayAsn (:357-364); and distinct underlay prefixes
// (DirectConnectConnectorValidator.java:136-148).
func checkConnectorAwsDirectConnect(m map[string]any) error {
	var errs []error
	if name, _ := m["name"].(string); strings.HasPrefix(strings.ToUpper(name), "ALK-") {
		errs = append(errs, errors.New("name must not start with ALK-"))
	}
	proto, _ := m["tunnelProtocol"].(string)
	if proto == "" {
		proto = "GRE"
	}
	vxlan := proto == "VXLAN"
	if m["size"] == "20LARGE" && proto != "VXLAN" && proto != "VXLAN_GPE" {
		errs = append(errs, fmt.Errorf("size 20LARGE needs tunnelProtocol VXLAN or VXLAN_GPE, not %s", proto))
	}
	instances, _ := m["instances"].([]any)
	var dcASN string
	prefixes := map[string]bool{}
	for _, e := range instances {
		in, _ := e.(map[string]any)
		name := fmt.Sprint(in["name"])
		if vxlan && in["vni"] == nil {
			errs = append(errs, fmt.Errorf("instance %s: vni must be set when tunnelProtocol is VXLAN", name))
		}
		if mac, _ := in["gatewayMacAddress"].(string); vxlan != (mac != "") {
			errs = append(errs, fmt.Errorf("instance %s: gatewayMacAddress must be set when tunnelProtocol is VXLAN, and only then", name))
		}
		if so, _ := in["segmentOptions"].([]any); vxlan && len(so) != 1 {
			errs = append(errs, fmt.Errorf("instance %s: exactly one segmentOptions entry is allowed when tunnelProtocol is VXLAN", name))
		}
		errs = append(errs, checkUnderlay(name, in, prefixes))
		asn := fmt.Sprint(in["dcGatewayAsn"])
		if dcASN == "" {
			dcASN = asn
		} else if asn != dcASN {
			errs = append(errs, fmt.Errorf("instance %s: dcGatewayAsn %s must be the same in every instance (%s)", name, asn, dcASN))
		}
		if asn == fmt.Sprint(in["underlayAsn"]) {
			errs = append(errs, fmt.Errorf("instance %s: dcGatewayAsn must differ from underlayAsn", name))
		}
	}
	return errors.Join(errs...)
}

// checkUnderlay checks an instance's underlay prefix and IPs, and records the
// prefix in seen.
func checkUnderlay(name string, in map[string]any, seen map[string]bool) error {
	prefix, _ := in["underlayPrefix"].(string)
	onPrem, _ := in["onPremUnderlayIp"].(string)
	aws, _ := in["awsUnderlayIp"].(string)
	if onPrem == "" && aws == "" {
		return nil
	}
	if prefix == "" || onPrem == "" || aws == "" {
		return fmt.Errorf("instance %s: underlayPrefix, onPremUnderlayIp and awsUnderlayIp must all be set", name)
	}
	if seen[prefix] {
		return fmt.Errorf("instance %s: underlayPrefix %s is used by another instance", name, prefix)
	}
	seen[prefix] = true
	p, err := netip.ParsePrefix(prefix)
	if err != nil || p.Bits() != 30 {
		return fmt.Errorf("instance %s: underlayPrefix %s must be a /30 prefix", name, prefix)
	}
	// The two host addresses of the /30, as SubnetUtils with
	// setInclusiveHostCount(false) counts them.
	first := p.Masked().Addr().Next()
	for field, ip := range map[string]string{"onPremUnderlayIp": onPrem, "awsUnderlayIp": aws} {
		if a, err := netip.ParseAddr(ip); err != nil || a != first && a != first.Next() {
			return fmt.Errorf("instance %s: %s %s must be a host address of underlayPrefix %s", name, field, ip, prefix)
		}
	}
	return nil
}

// ConnectorAwsDirectConnectDescriptor returns the AWS Direct Connect connector write descriptor.
func ConnectorAwsDirectConnectDescriptor() Descriptor {
	return connectorAwsDirectConnectDescriptor.clone()
}

// NewConnectorAwsDirectConnectCore returns the write core for AWS Direct Connect connectors.
func NewConnectorAwsDirectConnectCore(ac *AlkiraClient) (*Core[ConnectorAwsDirectConnectConfig], error) {
	return NewCore[ConnectorAwsDirectConnectConfig](ac, connectorAwsDirectConnectDescriptor)
}

// CreateConnectorAwsDirectConnect creates an AWS Direct Connect connector.
func (ac *AlkiraClient) CreateConnectorAwsDirectConnect(ctx context.Context, c *ConnectorAwsDirectConnectConfig) (*ConnectorAwsDirectConnectConfig, *WriteResult, error) {
	core, err := NewConnectorAwsDirectConnectCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, c, WriteOptions{})
}

// GetConnectorAwsDirectConnect reads one AWS Direct Connect connector and its raw body.
func (ac *AlkiraClient) GetConnectorAwsDirectConnect(ctx context.Context, id string) (*ConnectorAwsDirectConnectConfig, []byte, error) {
	core, err := NewConnectorAwsDirectConnectCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

// UpdateConnectorAwsDirectConnect applies a merge patch to an AWS Direct Connect connector.
func (ac *AlkiraClient) UpdateConnectorAwsDirectConnect(ctx context.Context, id string, patch json.RawMessage) (*ConnectorAwsDirectConnectConfig, *WriteResult, error) {
	core, err := NewConnectorAwsDirectConnectCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

// DeleteConnectorAwsDirectConnect deletes an AWS Direct Connect connector; a
// deployed one is marked for deletion until the next provision.
func (ac *AlkiraClient) DeleteConnectorAwsDirectConnect(ctx context.Context, id string) (*WriteResult, error) {
	core, err := NewConnectorAwsDirectConnectCore(ac)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

// ListConnectorAwsDirectConnectPaged reads every AWS Direct Connect connector through paged reads.
func (ac *AlkiraClient) ListConnectorAwsDirectConnectPaged(ctx context.Context) ([]ConnectorAwsDirectConnectConfig, error) {
	core, err := NewConnectorAwsDirectConnectCore(ac)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
