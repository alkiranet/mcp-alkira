// Copyright (C) 2022-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

type CheckpointInstance struct {
	Id             int    `json:"id,omitempty"` // response only
	Name           string `json:"name"`
	InternalName   string `json:"internalName,omitempty"` // response only
	CredentialId   string `json:"credentialId"`
	TrafficEnabled bool   `json:"trafficEnabled"`
}

type CheckpointInstanceConfig struct {
	// The response is string data the entire body of the response
	// whould be interpreted together. There is no json structure.
	Data string
}

type CheckpointManagementServer struct {
	ConfigurationMode string   `json:"configurationMode"`
	CredentialId      string   `json:"credentialId"`
	Domain            string   `json:"domain"`
	GlobalCidrListId  int      `json:"globalCidrListId"`
	Ips               []string `json:"ips"`
	Reachability      string   `json:"reachability"`
	Segment           string   `json:"segment,omitempty"`
	Type              string   `json:"type"`
	UserName          string   `json:"userName"`
}

type ServiceCheckpoint struct {
	AutoScale        string                      `json:"autoScale"`
	BillingTags      []int                       `json:"billingTags"`
	Cxp              string                      `json:"cxp"`
	CredentialId     string                      `json:"credentialId"`
	Description      string                      `json:"description"`
	Id               json.Number                 `json:"id,omitempty"` // response only
	Instances        []CheckpointInstance        `json:"instances"`
	InternalName     string                      `json:"internalName"`
	LicenseType      string                      `json:"licenseType"`
	ManagementServer *CheckpointManagementServer `json:"managementServer"`
	MaxInstanceCount int                         `json:"maxInstanceCount"`
	MinInstanceCount int                         `json:"minInstanceCount"`
	Name             string                      `json:"name"`
	PdpIps           []string                    `json:"pdpIps,omitempty"`
	Segments         []string                    `json:"segments"`
	SegmentOptions   SegmentNameToZone           `json:"segmentOptions"`
	Size             string                      `json:"size"`
	TunnelProtocol   string                      `json:"tunnelProtocol"`
	Version          string                      `json:"version"`
}

type ServiceCheckpointSummary struct {
	BillingTags  []int       `json:"billingTags"`
	Cxp          string      `json:"cxp"`
	Id           json.Number `json:"id,omitempty"`
	InternalName string      `json:"internalName"`
	LicenseType  string      `json:"licenseType"`
	Name         string      `json:"name"`
	Segments     []string    `json:"segments"`
	Version      string      `json:"version"`
}

// NewServiceCheckpoint new service checkpoint
func NewServiceCheckpoint(ac *AlkiraClient) *AlkiraApi[ServiceCheckpointSummary] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/chkp-fw-services", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[ServiceCheckpointSummary]{ac, uri, PaginationOn}
	return api
}

// ServiceCheckpointConfig is the write model of a Check Point firewall
// service (TPS model/CheckPointFWService.java; GET /services type CHKPFW).
// Credentials are ids of /credentials/chkp-fw, chkp-fw-management-server
// and chkp-fw-instance credentials, which cannot be updated in place
// (credentials-java ChkpFWCredentialService.java:36): a new secret is a new
// credential id.
type ServiceCheckpointConfig struct {
	ID           json.Number `json:"id,omitempty" ak:"readonly"`
	Name         string      `json:"name" ak:"required,maxlen=60,charset=hyphen,src=CheckPointFWService.java:68-71"`
	Description  *string     `json:"description,omitempty" ak:"maxlen=512,charset=description,src=CheckPointFWService.java:73-75"`
	CredentialID string      `json:"credentialId" ak:"required,src=CheckPointFWService.java:77"`
	// CXP is kept from the stored service on update (ServiceLogicImpl.java:891).
	CXP string `json:"cxp" ak:"required,immutable,maxlen=128,trimmed,src=CheckPointFWService.java:80-83"`
	// PdpIps cannot be combined with AllowList (CheckPointFWServiceValidator.java:67-71).
	PdpIps           []string                          `json:"pdpIps,omitempty"`
	ManagementServer *CheckpointManagementServerConfig `json:"managementServer,omitempty" ak:"required,src=CheckPointFWServiceValidator.java:17-21"`
	// Segments holds exactly one segment name.
	Segments []string `json:"segments" ak:"required,minitems=1,maxitems=1,src=CheckPointFWService.java:91-92"`
	// SegmentOptions holds one entry, keyed by segments[0], whose zonesToGroups has the zone DEFAULT only.
	SegmentOptions   map[string]FirewallSegmentOptions `json:"segmentOptions" ak:"required,src=CheckPointFWService.java:94-101"`
	LicenseType      string                            `json:"licenseType" ak:"required,enum=BRING_YOUR_OWN|PAY_AS_YOU_GO,src=CheckPointFWService.java:103|Defaults.java:671-673"`
	Version          string                            `json:"version" ak:"required,maxlen=128,src=CheckPointFWService.java:106-108"`
	Size             *string                           `json:"size,omitempty" ak:"default=SMALL,enum=SMALL|MEDIUM|LARGE,src=CheckPointFWService.java:110-113"`
	BillingTags      []int64                           `json:"billingTags,omitempty"`
	AutoScale        *string                           `json:"autoScale,omitempty" ak:"required,enum=ON|OFF,src=CheckPointFWService.java:120|Service.java:176-179"`
	MaxInstanceCount *int64                            `json:"maxInstanceCount,omitempty" ak:"required,min=1,src=CheckPointFWService.java:123-126"`
	MinInstanceCount *int64                            `json:"minInstanceCount,omitempty" ak:"required,min=1,src=CheckPointFWService.java:128-129"`
	Instances        []ServiceCheckpointInstanceConfig `json:"instances" ak:"required,minitems=1,src=CheckPointFWService.java:131|CheckPointFWServiceValidator.java:36"`
	TunnelProtocol   *string                           `json:"tunnelProtocol,omitempty" ak:"default=IPSEC,enum=IPSEC|GRE|VXLAN_GPE|VXLAN,src=CheckPointFWService.java:134-136|Defaults.java:1383-1387"`
	AllowList        []string                          `json:"allowList,omitempty" ak:"flag=RELEASE-SERVICE-CHKPFW-ALLOW-LIST"`

	InternalName string `json:"internalName,omitempty" ak:"readonly"`
	// State and PreviousState come from the GET /services row (Descriptor.ServiceType).
	State         string `json:"state,omitempty" ak:"readonly"`
	PreviousState string `json:"previousState,omitempty" ak:"readonly"`
}

// CheckpointManagementServerConfig is CheckPointFWService.managementServer.
// userName and credentialId are required and kept only in AUTOMATED mode,
// domain with AUTOMATED and MDS, segment with PRIVATE reachability
// (CheckPointFWServiceValidator.java:45-64; CheckPointFWServiceLogicImpl.java:451-461).
type CheckpointManagementServerConfig struct {
	ConfigurationMode string   `json:"configurationMode" ak:"required,enum=MANUAL|AUTOMATED,src=CheckPointFWService.java:223|:253-256"`
	Type              string   `json:"type" ak:"required,enum=SMS|MDS,src=CheckPointFWService.java:226|:258-261"`
	Reachability      string   `json:"reachability" ak:"required,enum=PRIVATE|PUBLIC,src=CheckPointFWService.java:229|:263-266"`
	Ips               []string `json:"ips" ak:"required,minitems=1,maxitems=2,src=CheckPointFWService.java:232-235"`
	Domain            *string  `json:"domain,omitempty"`
	UserName          *string  `json:"userName,omitempty" ak:"maxlen=32,charset=hyphen,src=CheckPointFWService.java:240-242"`
	CredentialID      *string  `json:"credentialId,omitempty"`
	Segment           *string  `json:"segment,omitempty"`
	// GlobalCidrListID is a global CIDR list of the service CXP holding /24 prefixes
	// (ManagementServerGlobalCidrValidator.java:29-38).
	GlobalCidrListID *int64 `json:"globalCidrListId,omitempty" ak:"required,src=CheckPointFWService.java:250"`

	SegmentID *int64 `json:"segmentId,omitempty" ak:"readonly"`
}

// ServiceCheckpointInstanceConfig is one Check Point instance
// (CheckPointFWServiceInstance.java); an update keeps an instance by sending its id.
type ServiceCheckpointInstanceConfig struct {
	ID             json.Number `json:"id,omitempty"`
	Name           string      `json:"name" ak:"required,maxlen=256,src=CheckPointFWServiceInstance.java:32-35"`
	CredentialID   string      `json:"credentialId" ak:"required,src=CheckPointFWServiceInstance.java:37"`
	TrafficEnabled *bool       `json:"trafficEnabled,omitempty"`

	InternalName *string `json:"internalName,omitempty" ak:"readonly"`
	HostName     *string `json:"hostName,omitempty" ak:"readonly"`
}

// serviceCheckpointDescriptor: the shared firewall rules (firewallDescriptor)
// and the portal's allow-list shaping; Check Point saves synchronously.
var serviceCheckpointDescriptor = func() Descriptor {
	d := firewallDescriptor("chkp-fw-services", "CHKPFW")
	d.CrossCheck = checkServiceCheckpoint
	d.Flags = []FlagRule{
		{Flag: "RELEASE-SERVICE-CHKPFW-ALLOW-LIST", Effect: FlagIncludeField, Field: "allowList", Refuse: true,
			UISource: "src/views/network/provision/objectConfig/checkpointConfig/add/index.js:297|:597-600"},
	}
	return d
}()

// checkServiceCheckpoint refuses what CheckPointFWServiceValidator.java:24-91
// refuses: instance counts, min other than max with autoScale OFF, the
// management server fields each mode needs, pdpIps with allowList, and
// segmentOptions other than one DEFAULT zone on segments[0].
func checkServiceCheckpoint(m map[string]any) error {
	errs := checkInstanceCounts(m, m["autoScale"] == "OFF")
	if p, _ := m["pdpIps"].([]any); len(p) > 0 {
		if a, _ := m["allowList"].([]any); len(a) > 0 {
			errs = append(errs, errors.New("pdpIps and allowList cannot both be set"))
		}
	}
	if ms, _ := m["managementServer"].(map[string]any); ms != nil {
		set := func(k string) bool { v, _ := ms[k].(string); return v != "" }
		if ms["configurationMode"] == "AUTOMATED" && (!set("userName") || !set("credentialId")) {
			errs = append(errs, errors.New("managementServer.userName and managementServer.credentialId are required in AUTOMATED mode"))
		}
		if ms["configurationMode"] == "AUTOMATED" && ms["type"] == "MDS" && !set("domain") {
			errs = append(errs, errors.New("managementServer.domain is required in AUTOMATED mode with type MDS"))
		}
		if ms["reachability"] == "PRIVATE" && !set("segment") {
			errs = append(errs, errors.New("managementServer.segment is required with PRIVATE reachability"))
		}
	}
	segs, _ := m["segments"].([]any)
	opts, _ := m["segmentOptions"].(map[string]any)
	for k, v := range opts {
		vm, _ := v.(map[string]any)
		zones, _ := vm["zonesToGroups"].(map[string]any)
		if len(segs) != 1 || k != segs[0] || len(zones) != 1 || zones["DEFAULT"] == nil {
			errs = append(errs, errors.New("segmentOptions must hold one entry for the service segment with the single zone DEFAULT"))
		}
	}
	return errors.Join(errs...)
}

// ServiceCheckpointDescriptor returns the Check Point firewall service write descriptor.
func ServiceCheckpointDescriptor() Descriptor { return serviceCheckpointDescriptor.clone() }

// CreateServiceCheckpoint creates a Check Point firewall service.
func (ac *AlkiraClient) CreateServiceCheckpoint(ctx context.Context, s *ServiceCheckpointConfig) (*ServiceCheckpointConfig, *WriteResult, error) {
	return serviceCreate(ctx, ac, serviceCheckpointDescriptor, s)
}

// GetServiceCheckpoint reads one Check Point firewall service and its raw body.
func (ac *AlkiraClient) GetServiceCheckpoint(ctx context.Context, id string) (*ServiceCheckpointConfig, []byte, error) {
	return serviceGet[ServiceCheckpointConfig](ctx, ac, serviceCheckpointDescriptor, id)
}

// UpdateServiceCheckpoint applies a merge patch to a Check Point firewall service.
func (ac *AlkiraClient) UpdateServiceCheckpoint(ctx context.Context, id string, patch json.RawMessage) (*ServiceCheckpointConfig, *WriteResult, error) {
	return serviceUpdate[ServiceCheckpointConfig](ctx, ac, serviceCheckpointDescriptor, id, patch)
}

// DeleteServiceCheckpoint deletes a Check Point firewall service; a deployed
// one is marked for deletion until the next provision.
func (ac *AlkiraClient) DeleteServiceCheckpoint(ctx context.Context, id string) (*WriteResult, error) {
	return serviceDelete[ServiceCheckpointConfig](ctx, ac, serviceCheckpointDescriptor, id)
}

// ListServiceCheckpointPaged reads every Check Point firewall service through paged reads.
func (ac *AlkiraClient) ListServiceCheckpointPaged(ctx context.Context) ([]ServiceCheckpointConfig, error) {
	return serviceList[ServiceCheckpointConfig](ctx, ac, serviceCheckpointDescriptor)
}
