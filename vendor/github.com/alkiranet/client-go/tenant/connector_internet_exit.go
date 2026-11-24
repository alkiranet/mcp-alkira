// Copyright (C) 2020-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ENUM for keys: "DEFAULT", "SRC_IP"
type AlgorithmAttributes struct {
	Keys string `json:"keys"`
}

type TrafficDistribution struct {
	Algorithm           string              `json:"algorithm"`
	AlgorithmAttributes AlgorithmAttributes `json:"algorithmAttributes"`
}

type ConnectorInternet struct {
	BillingTags         []int                `json:"billingTags"`
	ByoipId             int                  `json:"byoipId,omitempty"`
	CXP                 string               `json:"cxp"`
	Description         string               `json:"description"`
	Group               string               `json:"group,omitempty"`
	Enabled             bool                 `json:"enabled"`
	Id                  json.Number          `json:"id,omitempty"`              // response only
	ImplicitGroupId     int                  `json:"implicitGroupId,omitempty"` // response only
	PublicIps           []string             `json:"publicIps,omitempty"`
	Name                string               `json:"name"`
	NumOfPublicIPs      int                  `json:"numOfPublicIPs,omitempty"`
	Segments            []string             `json:"segments"`
	TrafficDistribution *TrafficDistribution `json:"trafficDistribution,omitempty"`
	EgressIpTypes       []string             `json:"egressIPTypes"`
}

func NewConnectorInternet(ac *AlkiraClient) *AlkiraApi[ConnectorInternet] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/internetconnectors", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[ConnectorInternet]{ac, uri, PaginationOn}
	return api
}

// ConnectorInternetConfig is the write model of an internet connector (TPS
// model/InternetConnector.java; GET /connectors reports its type as SAAS,
// InternetConnectorServiceImpl.java:270). Optional fields are pointers or
// omitempty slices, so a zero value is never sent; a create sends an unset
// field that has default= with that value, as the portal does
// (AddInternet/index.js:265-296).
type ConnectorInternetConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=InternetConnector.java:50-54"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,charset=description,src=InternetConnector.java:56-58"`
	CXP         string      `json:"cxp" ak:"required,maxlen=128,trimmed,src=InternetConnector.java:78-81"`
	// Segments holds exactly one segment name (InternetConnector.java:86-92).
	Segments    []string `json:"segments" ak:"required,src=InternetConnector.java:86-92"`
	Group       *string  `json:"group,omitempty" ak:"maxlen=256,trimmed,src=InternetConnector.java:63-65"`
	BillingTags []int64  `json:"billingTags,omitempty" ak:"flag=BILLING-TAGS"`
	// Size is not sent by the portal; the backend default is SMALL
	// (InternetConnector.java:99-102).
	Size                *string                         `json:"size,omitempty" ak:"enum=SMALL|MEDIUM|LARGE|2LARGE|4LARGE,src=InternetConnector.java:99-102|Defaults.java:127-150"`
	TrafficDistribution *ConnectorInternetTrafficConfig `json:"trafficDistribution,omitempty"`
	NumOfPublicIPs      *int64                          `json:"numOfPublicIPs,omitempty" ak:"default=2,enum=2|4|6|8,src=InternetConnectorServiceImpl.java:98|:115-121"`
	Enabled             *bool                           `json:"enabled,omitempty" ak:"default=true,src=ConnectorServiceImpl.java:978-980"`
	ByoipID             *int64                          `json:"byoipId,omitempty"`
	PublicIPs           []string                        `json:"publicIps,omitempty"`
	EgressIPTypes       []string                        `json:"egressIPTypes,omitempty" ak:"enum=BYOIP|ALKIRA_PUBLIC_IP,src=InternetConnector.java:151-157"`

	GroupID         *int64 `json:"groupId,omitempty" ak:"readonly"`
	ImplicitGroupID *int64 `json:"implicitGroupId,omitempty" ak:"readonly"`
	// The GET body repeats numOfPublicIPs and egressIPTypes under these
	// spellings (InternetConnector.java:159-167); they are left out of a PUT
	// so a patch of the canonical field is the only value sent.
	NumOfPublicIPsAlias   *int64   `json:"numOfPublicIps,omitempty" ak:"readonly"`
	EgressIPTypesAlias    []string `json:"egressIpTypes,omitempty" ak:"readonly"`
	NotEligibleForScaling *bool    `json:"notEligibleForScaling,omitempty" ak:"readonly"`
	// State and PreviousState come from the GET /connectors row
	// (Descriptor.ConnectorType); the internet connector GET body has neither.
	State         string `json:"state,omitempty" ak:"readonly"`
	PreviousState string `json:"previousState,omitempty" ak:"readonly"`
}

// ConnectorInternetTrafficConfig is InternetConnector.trafficDistribution;
// the portal always sends HASHING with keys DEFAULT or SRC_IP.
type ConnectorInternetTrafficConfig struct {
	Algorithm           *string                             `json:"algorithm,omitempty" ak:"default=HASHING,enum=HASHING,src=InternetConnector.java:104-141|AddInternet/index.js:265-296"`
	AlgorithmAttributes *ConnectorInternetTrafficKeysConfig `json:"algorithmAttributes,omitempty"`
}

// ConnectorInternetTrafficKeysConfig is trafficDistribution.algorithmAttributes.
type ConnectorInternetTrafficKeysConfig struct {
	Keys *string `json:"keys,omitempty" ak:"default=DEFAULT,enum=DEFAULT|SRC_IP,src=InternetConnector.java:104-141|AddInternet/index.js:265-296"`
}

// connectorInternetDescriptor: GET and PUT answer a missing id with 404
// AK000003 (ConnectorServiceImpl.java:5548-5549; AbstractConnectorService.java:571-575),
// DELETE too (:887). Connector ids share one id space and a GET of a sibling
// family's id answers 200 in this family's shape (AbstractConnectorService.java:686-688),
// so numOfPublicIPs, which the backend stores for every internet connector
// (InternetConnectorServiceImpl.java:98), and the SAAS type of the GET
// /connectors row (:270) identify the family; the row also carries the state
// the D13 guard reads. A connector
// name is unique across every connector type and group, marked rows included
// (ConnectorServiceImpl.java:5598-5603, :938-940). A DELETE is refused with
// 400 AK000002 while internet applications, segment resources or NAT
// policies use the connector (:2485-2499; InternetConnectorServiceImpl.java:517-537)
// and with 409 for an external reference (NetworkConfigEntityListener.java:83-97).
// An update runs the D13 state check (ProvisionedEntityValidatorImpl.java:34-76);
// a deployed connector is marked for deletion on delete and a plain GET then
// answers 404 (ConnectorServiceImpl.java:2513-2525, :761-762).
var connectorInternetDescriptor = Descriptor{
	Collection:    "internetconnectors",
	FamilyType:    FamilyCheck{Field: "numOfPublicIPs"},
	ConnectorType: "SAAS",
	InUse: []Signature{
		{Status: 400, Code: "AK000002", Message: "associated with"},
		{Status: 400, Code: "AK000002", Message: "Internet exit is mandatory"},
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
	CrossCheck:         checkConnectorInternet,
	Flags: []FlagRule{
		{Flag: "BILLING-TAGS", Effect: FlagIncludeField, Field: "billingTags", UISource: "src/views/network/provision/objectConfig/internetConfig/AddInternet/index.js:369|:279"},
	},
}

// checkConnectorInternet refuses a name starting with ALK- in any case
// (GroupServiceImpl.java:1205-1207; ConnectorServiceImpl.java:949-952) and
// more or fewer than one segment (InternetConnector.java:86-92).
func checkConnectorInternet(m map[string]any) error {
	var errs []error
	if name, _ := m["name"].(string); strings.HasPrefix(strings.ToUpper(name), "ALK-") {
		errs = append(errs, errors.New("name must not start with ALK-"))
	}
	if segs, _ := m["segments"].([]any); len(segs) != 1 {
		errs = append(errs, fmt.Errorf("segments must hold exactly one segment name, not %d", len(segs)))
	}
	return errors.Join(errs...)
}

// ConnectorInternetDescriptor returns the internet connector write descriptor.
func ConnectorInternetDescriptor() Descriptor { return connectorInternetDescriptor.clone() }

// NewConnectorInternetCore returns the write core for internet connectors.
func NewConnectorInternetCore(ac *AlkiraClient) (*Core[ConnectorInternetConfig], error) {
	return NewCore[ConnectorInternetConfig](ac, connectorInternetDescriptor)
}

// CreateConnectorInternet creates an internet connector.
func (ac *AlkiraClient) CreateConnectorInternet(ctx context.Context, c *ConnectorInternetConfig) (*ConnectorInternetConfig, *WriteResult, error) {
	core, err := NewConnectorInternetCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, c, WriteOptions{})
}

// GetConnectorInternet reads one internet connector and its raw body.
func (ac *AlkiraClient) GetConnectorInternet(ctx context.Context, id string) (*ConnectorInternetConfig, []byte, error) {
	core, err := NewConnectorInternetCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

// UpdateConnectorInternet applies a merge patch to an internet connector.
func (ac *AlkiraClient) UpdateConnectorInternet(ctx context.Context, id string, patch json.RawMessage) (*ConnectorInternetConfig, *WriteResult, error) {
	core, err := NewConnectorInternetCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

// DeleteConnectorInternet deletes an internet connector; a deployed one is
// marked for deletion until the next provision.
func (ac *AlkiraClient) DeleteConnectorInternet(ctx context.Context, id string) (*WriteResult, error) {
	core, err := NewConnectorInternetCore(ac)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

// ListConnectorInternetPaged reads every internet connector through paged reads.
func (ac *AlkiraClient) ListConnectorInternetPaged(ctx context.Context) ([]ConnectorInternetConfig, error) {
	core, err := NewConnectorInternetCore(ac)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
