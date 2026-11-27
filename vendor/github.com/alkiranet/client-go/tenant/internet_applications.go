// Copyright (C) 2020-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type InternetApplicationSnatIpv4 struct {
	StartIp string `json:"startIp"`
	EndIp   string `json:"endIp"`
}

type InternetApplicationTargets struct {
	Type       string   `json:"type"`
	Value      string   `json:"value"`
	Ports      []int    `json:"ports"`
	PortRanges []string `json:"portRanges,omitempty"`
}

type InternetApplication struct {
	BillingTags                   []int                          `json:"billingTags"`
	BiDirectionalAvailabilityZone string                         `json:"biDirectionalAvailabilityZone,omitempty"`
	ByoipId                       int                            `json:"byoipId,omitempty"`
	ConnectorId                   int                            `json:"connectorId"`
	ConnectorType                 string                         `json:"connectorType"`
	Description                   string                         `json:"description,omitempty"`
	FqdnPrefix                    string                         `json:"fqdnPrefix"`
	Group                         string                         `json:"group,omitempty"`
	Id                            json.Number                    `json:"id,omitempty"` // response only
	InboundConnectorId            string                         `json:"inboundConnectorId,omitempty"`
	InboundConnectorType          string                         `json:"inboundConnectorType,omitempty"`
	InboundInternetGroupId        json.Number                    `json:"inboundInternetGroupId,omitempty"`
	InternetProtocol              string                         `json:"internetProtocol"`
	Name                          string                         `json:"name"`
	PublicIps                     []string                       `json:"publicIps"`
	SegmentName                   string                         `json:"segmentName"`
	SnatIpv4Ranges                []*InternetApplicationSnatIpv4 `json:"snatIpv4Ranges,omitempty"`
	Size                          string                         `json:"size"`
	Targets                       []InternetApplicationTargets   `json:"targets,omitempty"`
	IlbCredentialId               string                         `json:"ilbCredentialId,omitempty"`
}

// NewInternetApplication new internet application
func NewInternetApplication(ac *AlkiraClient) *AlkiraApi[InternetApplication] {
	uri := fmt.Sprintf("%s/v1/tenantnetworks/%s/internet-applications", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[InternetApplication]{ac, uri, PaginationOn}
	return api
}

// InternetApplicationConfig is the write model of an internet application
// (TPS model/InternetApplication.java). A PUT replaces the whole object
// (InternetApplicationServiceImpl.java:578-658), so an update resends every
// stored field. The backend creates or reuses one server-owned INB_INT
// connector per segment and CXP for it (:513-515); that connector is not
// written here.
type InternetApplicationConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,charset=slug,src=InternetApplication.java:69-73"`
	Description *string     `json:"description,omitempty" ak:"maxlen=255,trimmed,charset=name,src=InternetApplication.java:75-78"`
	// FqdnPrefix is required outside China CXPs and refused inside them
	// (InternetApplicationServiceImpl.java:712-714, :969-971).
	FqdnPrefix  *string `json:"fqdnPrefix,omitempty" ak:"maxlen=255,charset=alnum,src=InternetApplication.java:80-82"`
	SegmentName string  `json:"segmentName" ak:"required,src=InternetApplication.java:93-94"`
	ConnectorID *int64  `json:"connectorId" ak:"required,src=InternetApplication.java:98-99"`
	// ConnectorType must equal the type of connectorId when set
	// (InternetApplicationServiceImpl.java:749-751).
	ConnectorType *string                     `json:"connectorType,omitempty"`
	Size          *string                     `json:"size,omitempty" ak:"default=SMALL,enum=SMALL|MEDIUM|LARGE|2LARGE|4LARGE|5LARGE,src=InternetApplication.java:104-107|EntitySizeValidator.java:14-22"`
	BillingTags   []int64                     `json:"billingTags,omitempty" ak:"flag=BILLING-TAGS"`
	Targets       []InternetApplicationTarget `json:"targets" ak:"required,src=InternetApplication.java:182-183|InternetApplicationValidator.java:32-38"`

	InboundConnectorType          *string                    `json:"inboundConnectorType,omitempty" ak:"enum=DEFAULT|AKAMAI_PROLEXIC,flag=AKAMAI-PRO-PHASE-1,src=InternetApplication.java:185-187|:224-226"`
	InboundConnectorID            *int64                     `json:"inboundConnectorId,omitempty" ak:"flag=AKAMAI-PRO-PHASE-1"`
	PublicIPs                     []string                   `json:"publicIps,omitempty"`
	ByoipID                       *int64                     `json:"byoipId,omitempty"`
	InternetProtocol              *string                    `json:"internetProtocol,omitempty" ak:"default=IPV4,enum=IPV4|IPV6|BOTH,flag=IFA-IPV6-SUPPORT,src=InternetApplication.java:197-198|:254-258"`
	SnatIPv4Ranges                []InternetApplicationRange `json:"snatIpv4Ranges,omitempty"`
	BiDirectionalAvailabilityZone *string                    `json:"biDirectionalAvailabilityZone,omitempty" ak:"enum=AZ0|AZ1,src=InternetApplication.java:204|:306-308"`
	IlbCredentialID               *string                    `json:"ilbCredentialId,omitempty"`

	TenantID                   json.Number `json:"tenantId,omitempty" ak:"readonly"`
	TenantNetworkID            json.Number `json:"tenantNetworkId,omitempty" ak:"readonly"`
	SegmentID                  *int64      `json:"segmentId,omitempty" ak:"readonly"`
	SessionLimit               *int64      `json:"sessionLimit,omitempty" ak:"readonly"`
	InternalName               string      `json:"internalName,omitempty" ak:"readonly"`
	ProductID                  string      `json:"productId,omitempty" ak:"readonly"`
	ConnectorName              string      `json:"connectorName,omitempty" ak:"readonly"`
	CXP                        string      `json:"cxp,omitempty" ak:"readonly"`
	AppID                      *int64      `json:"appId,omitempty" ak:"readonly"`
	InlineServices             []any       `json:"inlineServices,omitempty" ak:"readonly"`
	State                      string      `json:"state,omitempty" ak:"readonly"`
	UpdatedAt                  json.Number `json:"updatedAt,omitempty" ak:"readonly"`
	UpdatedBy                  string      `json:"updatedBy,omitempty" ak:"readonly"`
	ConnectorGroupID           *int64      `json:"connectorGroupId,omitempty" ak:"readonly"`
	InboundInternetGroupID     *int64      `json:"inboundInternetGroupId,omitempty" ak:"readonly"`
	InboundInternetConnectorID *int64      `json:"inboundInternetConnectorId,omitempty" ak:"readonly"`
	FQDN                       string      `json:"fqdn,omitempty" ak:"readonly"`
	CXPPublicIPs               []any       `json:"cxpPublicIps,omitempty" ak:"readonly"`
	TrafficEnabled             *bool       `json:"trafficEnabled,omitempty" ak:"readonly"`
	DocState                   string      `json:"docState,omitempty" ak:"readonly"`
	LastConfigUpdatedAt        json.Number `json:"lastConfigUpdatedAt,omitempty" ak:"readonly"`
}

// InternetApplicationTarget is one InternetApplicationTarget; set exactly one
// of ports and portRanges (InternetApplicationServiceImpl.java:729-730).
type InternetApplicationTarget struct {
	Type             *string  `json:"type" ak:"required,enum=IP|ILB_NAME|INTERNAL_DNS,src=InternetApplicationTarget.java:29-30|:45-49"`
	Value            *string  `json:"value,omitempty"`
	PolicyFqdnListID *int64   `json:"policyFqdnListId,omitempty"`
	Ports            []int64  `json:"ports,omitempty" ak:"range=-1..65534,src=InternetApplicationTarget.java:36-39"`
	PortRanges       []string `json:"portRanges,omitempty"`
}

// InternetApplicationRange is one snatIpv4Ranges entry (IPRange.java:24-28).
type InternetApplicationRange struct {
	StartIP string `json:"startIp" ak:"required,src=IPRange.java:24-28"`
	EndIP   string `json:"endIp" ak:"required,src=IPRange.java:24-28"`
}

// internetApplicationDescriptor: the API serves the family under
// /v1/tenantnetworks (web-apigw strips /v1, Utils.java:11-25). GET and PUT
// answer a missing id with 404 AK000003 (InternetApplicationServiceImpl.java:1262-1268,
// AbstractNetworkCRUDService.java:208), DELETE too (:291). A DELETE held by
// a policy, a probe or an external reference is 409 AK000001 (:1323,
// :1333-1335, :1346-1347; PolicyServiceProviderImpl.java:201-204). A
// duplicate name is 400 ifa.00000001 (:717-719). Deleting a deployed
// application marks it for deletion (:1408-1417); a GET by id still returns
// the marked row with its docState (:1262-1268). A FAILED application is not
// updated but re-provisioned as stored (AbstractNetworkCRUDService.java:214-218).
var internetApplicationDescriptor = Descriptor{
	Collection: "internet-applications",
	V1Path:     true,
	InUse:      []Signature{{Status: 409, Code: "AK000001"}},
	Duplicate:  []Signature{{Status: 400, FieldCode: "ifa.00000001"}},

	Provisionable:      true,
	MarkedRowLifecycle: true,
	DocState:           DocStateMissingIsAdded,
	Uniqueness:         Uniqueness{Key: "name", Scope: ScopeCollection},
	StateGuard:         StateGuardFailedOnly,
	Service:            ServiceTPS,
	CrossCheck:         checkInternetApplication,
	Flags: []FlagRule{
		{Flag: "BILLING-TAGS", Effect: FlagIncludeField, Field: "billingTags", UISource: "src/views/applications/internetApplications/AddInternetApplication.js:1411-1418"},
		{Flag: "IFA-IPV6-SUPPORT", Effect: FlagIncludeField, Field: "internetProtocol", UISource: "src/views/applications/internetApplications/AddInternetApplication.js:68|:905-910"},
		{Flag: "AKAMAI-PRO-PHASE-1", Effect: FlagIncludeField, Field: "inboundConnectorType", UISource: "src/views/applications/internetApplications/AddInternetApplication.js:881-895"},
		{Flag: "AKAMAI-PRO-PHASE-1", Effect: FlagIncludeField, Field: "inboundConnectorId", UISource: "src/views/applications/internetApplications/AddInternetApplication.js:881-895"},
	},
}

// checkInternetApplication holds the rules tags cannot express: the name
// neither starts nor ends with - or _ (InternetApplication.java:69-73); an
// INTERNAL_DNS target names a policy FQDN list and no value, any other
// target a value and no list (InternetApplicationValidator.java:111-149);
// each target sets exactly one of ports and portRanges
// (InternetApplicationServiceImpl.java:729-730).
func checkInternetApplication(m map[string]any) error {
	var errs []error
	if name, _ := m["name"].(string); strings.HasPrefix(name, "-") || strings.HasPrefix(name, "_") || strings.HasSuffix(name, "-") || strings.HasSuffix(name, "_") {
		errs = append(errs, errors.New("name must not start or end with - or _"))
	}
	targets, _ := m["targets"].([]any)
	for i, t := range targets {
		tm, ok := t.(map[string]any)
		if !ok {
			continue
		}
		set := func(k string) bool {
			v, ok := tm[k]
			if s, isString := v.(string); isString {
				return s != ""
			}
			if l, isList := v.([]any); isList {
				return len(l) > 0
			}
			return ok && v != nil
		}
		if tm["type"] == "INTERNAL_DNS" {
			if !set("policyFqdnListId") || set("value") {
				errs = append(errs, fmt.Errorf("targets[%d]: an INTERNAL_DNS target needs policyFqdnListId and no value", i))
			}
		} else if !set("value") || set("policyFqdnListId") {
			errs = append(errs, fmt.Errorf("targets[%d]: a %v target needs a value and no policyFqdnListId", i, tm["type"]))
		}
		if set("ports") == set("portRanges") {
			errs = append(errs, fmt.Errorf("targets[%d]: set exactly one of ports and portRanges", i))
		}
	}
	return errors.Join(errs...)
}

// InternetApplicationDescriptor returns the internet application write descriptor.
func InternetApplicationDescriptor() Descriptor { return internetApplicationDescriptor.clone() }

// NewInternetApplicationCore returns the write core for internet applications.
func NewInternetApplicationCore(ac *AlkiraClient) (*Core[InternetApplicationConfig], error) {
	return NewCore[InternetApplicationConfig](ac, internetApplicationDescriptor)
}

// CreateInternetApplication creates an internet application.
func (ac *AlkiraClient) CreateInternetApplication(ctx context.Context, a *InternetApplicationConfig) (*InternetApplicationConfig, *WriteResult, error) {
	c, err := NewInternetApplicationCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Create(ctx, a, WriteOptions{})
}

// GetInternetApplication reads one internet application and its raw body.
func (ac *AlkiraClient) GetInternetApplication(ctx context.Context, id string) (*InternetApplicationConfig, []byte, error) {
	c, err := NewInternetApplicationCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Get(ctx, id)
}

// UpdateInternetApplication applies a merge patch to an internet application.
func (ac *AlkiraClient) UpdateInternetApplication(ctx context.Context, id string, patch json.RawMessage) (*InternetApplicationConfig, *WriteResult, error) {
	c, err := NewInternetApplicationCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Update(ctx, id, patch, ListEdits{})
}

// DeleteInternetApplication deletes an internet application.
func (ac *AlkiraClient) DeleteInternetApplication(ctx context.Context, id string) (*WriteResult, error) {
	c, err := NewInternetApplicationCore(ac)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, id)
}

// ListInternetApplicationPaged reads every internet application through paged reads.
func (ac *AlkiraClient) ListInternetApplicationPaged(ctx context.Context) ([]InternetApplicationConfig, error) {
	c, err := NewInternetApplicationCore(ac)
	if err != nil {
		return nil, err
	}
	return c.ListPaged(ctx)
}
