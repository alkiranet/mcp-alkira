// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// ConnectorCiscoSdwanConfig is the write model of a Cisco SD-WAN connector
// (TPS model/CiscoSDWANIngress.java; GET /connectors reports its type as
// SD_WAN). The fields follow the portal create body (karrawa
// sdwanConfig/AddConnector/index.js:576-604); optional fields are pointers
// or omitempty slices, so a zero value is never sent.
type ConnectorCiscoSdwanConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=CiscoSDWANIngress.java:72-76"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,charset=description,src=CiscoSDWANIngress.java:78-80"`
	CXP         string      `json:"cxp" ak:"required,maxlen=128,trimmed,src=CiscoSDWANIngress.java:64-67"`
	Group       *string     `json:"group,omitempty" ak:"maxlen=256,trimmed,charset=name,src=CiscoSDWANIngress.java:85-88"`
	// CiscoEdgeInfo lists the edges; hostNames are distinct
	// (CiscoSDWANIngressServiceImpl.java:345-375).
	CiscoEdgeInfo        []ConnectorCiscoSdwanEdge       `json:"ciscoEdgeInfo" ak:"required,distinct=hostName,src=CiscoSDWANIngress.java:98-100|CiscoSDWANIngressServiceImpl.java:374-376"`
	CiscoEdgeVRFMappings []ConnectorCiscoSdwanVRFMapping `json:"ciscoEdgeVRFMappings" ak:"required,distinct=vrf|segmentId,src=CiscoSDWANIngress.java:102-104|CiscoSDWANIngressServiceImpl.java:322-329"`
	Size                 string                          `json:"size" ak:"required,enum=SMALL|MEDIUM|LARGE|2LARGE|5LARGE|10LARGE,src=CiscoSDWANIngress.java:106-109|Defaults.java:148-150"`
	// Version must name an image of the type, CXP and size
	// (CiscoSDWANIngressServiceImpl.java:126-138).
	Version     string   `json:"version" ak:"required,src=CiscoSDWANIngress.java:111-113"`
	BillingTags []int64  `json:"billingTags,omitempty" ak:"flag=BILLING-TAGS"`
	AllowList   []string `json:"allowList,omitempty" ak:"flag=RELEASE-CONNECTOR-CISCO-SDWAN-ALLOW-LIST"`
	// Type is the edge platform; the backend stores VEDGE when it is unset and
	// refuses a change (CiscoSDWANIngressServiceImpl.java:205-207).
	Type *string `json:"type,omitempty" ak:"immutable,enum=VEDGE|CSR|CAT8000V,src=CiscoSDWANIngress.java:136-137|Defaults.java:1407-1415|CiscoSDWANIngressServiceImpl.java:205-207"`
	// TunnelProtocol is IPSEC, or GRE on a CAT8000V outside Azure
	// (CiscoSDWANIngressServiceImpl.java:270-288).
	TunnelProtocol *string `json:"tunnelProtocol,omitempty" ak:"enum=IPSEC|GRE,src=CiscoSDWANIngress.java:143|CiscoSDWANIngressServiceImpl.java:270-279"`
	Enabled        *bool   `json:"enabled,omitempty"`

	GroupID         *int64          `json:"groupId,omitempty" ak:"readonly"`
	ImplicitGroupID *int64          `json:"implicitGroupId,omitempty" ak:"readonly"`
	LastEnabled     *int64          `json:"lastEnabled,omitempty" ak:"readonly"`
	ValidationState json.RawMessage `json:"validationState,omitempty" ak:"readonly"`
	// State and PreviousState come from the GET /connectors row.
	State         string `json:"state,omitempty" ak:"readonly"`
	PreviousState string `json:"previousState,omitempty" ak:"readonly"`
}

// ConnectorCiscoSdwanEdge is one ciscoEdgeInfo entry (CiscoEdgeInfo.java:21-44).
// The backend names it <connector>-<hostName> (CiscoEdgeInfo.java:46-51).
// After the connector is provisioned the cloud-init file, host and SSH key
// pair cannot change (CiscoSDWANIngressServiceImpl.java:415-445).
type ConnectorCiscoSdwanEdge struct {
	ID   json.Number `json:"id,omitempty"`
	Name *string     `json:"name,omitempty" ak:"readonly"`
	// CloudInitFile is the cloud-init content of the edge; CSR and CAT8000V
	// need org, vbond, uuid and otp in it (CiscoSDWANIngressServiceImpl.java:300-318, :356-370).
	CloudInitFile     string  `json:"cloudInitFile" ak:"required,masked,src=CiscoSDWANIngressServiceImpl.java:356-358"`
	CloudInitFileName *string `json:"cloudInitFileName,omitempty" ak:"readonly"`
	// CredentialID names a ciscosdwan credential (userName and password).
	CredentialID string `json:"credentialId" ak:"required,trimmed,src=CiscoEdgeInfo.java:34-36"`
	HostName     string `json:"hostName" ak:"required,maxlen=32,trimmed,src=CiscoEdgeInfo.java:38-41"`
	// SSHKeyPairCredentialID names a keypair credential; VEDGE only
	// (CiscoSDWANIngressServiceImpl.java:232-242).
	SSHKeyPairCredentialID *string `json:"sshKeyPairCredentialId,omitempty" ak:"trimmed,src=CiscoEdgeInfo.java:43-44"`
}

// ConnectorCiscoSdwanVRFMapping is one ciscoEdgeVRFMappings entry
// (CiscoEdgeVRFMapping.java:23-42). VRFs 0 and 512 are reserved
// (CiscoSDWANIngressServiceImpl.java:331-334), and a mapping cannot change
// after the connector is provisioned (:397-413).
type ConnectorCiscoSdwanVRFMapping struct {
	Vrf       *int64 `json:"vrf,omitempty" ak:"required,range=1..511|513..65530,src=CiscoEdgeVRFMapping.java:25-28|CiscoSDWANIngressServiceImpl.java:331-334"`
	SegmentID *int64 `json:"segmentId,omitempty" ak:"required,src=CiscoEdgeVRFMapping.java:29-30"`
	// CustomerAsn must differ from the segment's ASN
	// (CiscoSDWANIngressServiceImpl.java:336-341).
	CustomerAsn *int64 `json:"customerAsn,omitempty" ak:"range=1..4294967295,src=CiscoEdgeVRFMapping.java:32-33"`
	// DisableInternetExit defaults to true, the portal's "internet exit: No"
	// (karrawa sdwanConfig/AddConnector/index.js:36); the backend default is false (CiscoEdgeVRFMapping.java:35-37).
	DisableInternetExit   *bool `json:"disableInternetExit,omitempty" ak:"default=true"`
	AdvertiseOnPremRoutes *bool `json:"advertiseOnPremRoutes,omitempty"`
}

// connectorCiscoSdwanDescriptor: GET, PUT and DELETE answer a missing id
// with 404 AK000003, as every connector does. ciscoEdgeInfo identifies the
// family in the GET body and SD_WAN in the GET /connectors row, which also
// carries the state the D13 guard reads. Edges merge by id. TPS can validate
// this family asynchronously (CiscoSDWANIngressServiceImpl.java:965), but the
// portal never sends ?async=true for it (karrawa src/_api/sdwan.js:69-71,
// :100-105), so writes stay synchronous.
var connectorCiscoSdwanDescriptor = Descriptor{
	Collection:         "ciscosdwaningresses",
	FamilyType:         FamilyCheck{Field: "ciscoEdgeInfo"},
	ConnectorType:      "SD_WAN",
	InUse:              sdwanInUse,
	Duplicate:          sdwanDuplicate,
	Provisionable:      true,
	MarkedRowLifecycle: true,
	DocState:           DocStatePlainGet,
	Uniqueness:         Uniqueness{Key: "name", Scope: ScopeConnectorsGroups},
	IDMatchedArrays:    []string{"ciscoEdgeInfo"},
	StateGuard:         StateGuardD13,
	Service:            ServiceTPS,
	CrossCheck:         checkConnectorCiscoSdwan,
	Flags: []FlagRule{
		{Flag: "BILLING-TAGS", Effect: FlagIncludeField, Field: "billingTags", UISource: "src/views/network/provision/objectConfig/sdwanConfig/AddConnector/index.js:94|:584"},
		{Flag: "RELEASE-CONNECTOR-CISCO-SDWAN-ALLOW-LIST", Effect: FlagIncludeField, Field: "allowList", UISource: "src/views/network/provision/objectConfig/sdwanConfig/AddConnector/index.js:301-303|:590-600"},
	},
}

// checkConnectorCiscoSdwan adds to the name rule the platform rules of the
// backend: allowList on CAT8000V only (CiscoSDWANIngressServiceImpl.java:121-124),
// no SSH key pair on CSR or CAT8000V (:239-241), and IPSEC as the only tunnel
// protocol outside CAT8000V (:275-278). An unset type is VEDGE.
func checkConnectorCiscoSdwan(m map[string]any) error {
	errs := []error{checkConnectorName(m)}
	typ, _ := m["type"].(string)
	if typ == "" {
		typ = "VEDGE"
	}
	if l, _ := m["allowList"].([]any); len(l) > 0 && typ != "CAT8000V" {
		errs = append(errs, fmt.Errorf("allowList is supported on type CAT8000V only, not %s", typ))
	}
	if tp, _ := m["tunnelProtocol"].(string); tp != "" && tp != "IPSEC" && typ != "CAT8000V" {
		errs = append(errs, fmt.Errorf("tunnelProtocol %s needs type CAT8000V; type %s supports IPSEC only", tp, typ))
	}
	if typ == "CSR" || typ == "CAT8000V" {
		edges, _ := m["ciscoEdgeInfo"].([]any)
		for _, e := range edges {
			edge, _ := e.(map[string]any)
			if k, _ := edge["sshKeyPairCredentialId"].(string); k != "" {
				errs = append(errs, fmt.Errorf("edge %v: sshKeyPairCredentialId is not supported on type %s", edge["hostName"], typ))
			}
		}
	}
	return errors.Join(errs...)
}

// ConnectorCiscoSdwanDescriptor returns the Cisco SD-WAN connector write descriptor.
func ConnectorCiscoSdwanDescriptor() Descriptor { return connectorCiscoSdwanDescriptor.clone() }

// NewConnectorCiscoSdwanCore returns the write core for Cisco SD-WAN connectors.
func NewConnectorCiscoSdwanCore(ac *AlkiraClient) (*Core[ConnectorCiscoSdwanConfig], error) {
	return NewCore[ConnectorCiscoSdwanConfig](ac, connectorCiscoSdwanDescriptor)
}

// CreateConnectorCiscoSdwan creates a Cisco SD-WAN connector.
func (ac *AlkiraClient) CreateConnectorCiscoSdwan(ctx context.Context, c *ConnectorCiscoSdwanConfig) (*ConnectorCiscoSdwanConfig, *WriteResult, error) {
	core, err := NewConnectorCiscoSdwanCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, c, WriteOptions{})
}

// GetConnectorCiscoSdwan reads one Cisco SD-WAN connector and its raw body.
func (ac *AlkiraClient) GetConnectorCiscoSdwan(ctx context.Context, id string) (*ConnectorCiscoSdwanConfig, []byte, error) {
	core, err := NewConnectorCiscoSdwanCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

// UpdateConnectorCiscoSdwan applies a merge patch to a Cisco SD-WAN connector.
func (ac *AlkiraClient) UpdateConnectorCiscoSdwan(ctx context.Context, id string, patch json.RawMessage) (*ConnectorCiscoSdwanConfig, *WriteResult, error) {
	core, err := NewConnectorCiscoSdwanCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

// DeleteConnectorCiscoSdwan deletes a Cisco SD-WAN connector; a deployed one
// is marked for deletion until the next provision.
func (ac *AlkiraClient) DeleteConnectorCiscoSdwan(ctx context.Context, id string) (*WriteResult, error) {
	core, err := NewConnectorCiscoSdwanCore(ac)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

// ListConnectorCiscoSdwanPaged reads every Cisco SD-WAN connector through paged reads.
func (ac *AlkiraClient) ListConnectorCiscoSdwanPaged(ctx context.Context) ([]ConnectorCiscoSdwanConfig, error) {
	core, err := NewConnectorCiscoSdwanCore(ac)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
