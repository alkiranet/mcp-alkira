// Copyright (C) 2022-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type FortinetInstance struct {
	Name         string `json:"name"`
	Id           int    `json:"id,omitempty"`
	HostName     string `json:"hostName"`
	SerialNumber string `json:"serialNumber"`
	CredentialId string `json:"credentialId"`
}

type FortinetManagmentServer struct {
	IpAddress string `json:"ipAddress"`
	Segment   string `json:"segment"`
}

type FortinetInstanceConfig struct {
	ManagementIp string `json:"managementIP"`
	SerialNumber string `json:"serialNumber"`
}

type ServiceFortinet struct {
	AutoScale        string                   `json:"autoScale,omitempty"`
	BillingTags      []int                    `json:"billingTags"`
	CredentialId     string                   `json:"credentialId"`
	Cxp              string                   `json:"cxp"`
	Id               json.Number              `json:"id"`
	Instances        []FortinetInstance       `json:"instances"`
	InternalName     string                   `json:"internalName"`
	LicenseType      string                   `json:"licenseType"`
	ManagementServer *FortinetManagmentServer `json:"managementServer"`
	MaxInstanceCount int                      `json:"maxInstanceCount"`
	MinInstanceCount int                      `json:"minInstanceCount"`
	Name             string                   `json:"name"`
	Scheme           string                   `json:"scheme,omitempty"`
	Segments         []string                 `json:"segments"`
	SegmentOptions   SegmentNameToZone        `json:"segmentOptions"`
	Size             string                   `json:"size"`
	State            string                   `json:"state,omitempty"`
	TunnelProtocol   string                   `json:"tunnelProtocol"`
	Version          string                   `json:"version"`
	Description      string                   `json:"description,omitempty"`
}

type ServiceFortinetSummary struct {
	Cxp          string      `json:"cxp"`
	Id           json.Number `json:"id"`
	InternalName string      `json:"internalName"`
	LicenseType  string      `json:"licenseType"`
	Name         string      `json:"name"`
	Segments     []string    `json:"segments"`
	State        string      `json:"state,omitempty"`
	Version      string      `json:"version"`
}

// NewServiceFortinet new service fortinet
func NewServiceFortinet(ac *AlkiraClient) *AlkiraApi[ServiceFortinetSummary] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/ftnt-fw-services", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[ServiceFortinetSummary]{ac, uri, PaginationOn}
	return api
}

// ServiceFortinetConfig is the write model of a Fortinet firewall service
// (TPS model/fortinetfwservice/FortinetFWService.java; GET /services type
// FTNTFW). TPS refuses autoScale ON (FortinetFWServiceLogicImpl.java:809-811)
// and keeps the stored managementServer.segment on update (:200). After the
// first provision name, cxp, licenseType, tunnelProtocol, version and
// credentialId cannot change (:796-807).
type ServiceFortinetConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=22,charset=hyphen,src=FortinetFWService.java:77-82"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,charset=description,src=FortinetFWService.java:84-86"`
	Size        *string     `json:"size,omitempty" ak:"default=SMALL,enum=SMALL|MEDIUM|LARGE|2LARGE|5LARGE,src=FortinetFWService.java:88-91|Defaults.java:122-145"`
	CXP         string      `json:"cxp" ak:"required,immutable,maxlen=128,trimmed,src=FortinetFWService.java:93-96"`
	Version     string      `json:"version" ak:"required,maxlen=32,src=FortinetFWService.java:98-100"`
	// CredentialID is a /credentials/ftntfw credential.
	CredentialID     string                            `json:"credentialId" ak:"required,src=FortinetFWService.java:102-103"`
	ManagementServer *FortinetManagementServerConfig   `json:"managementServer,omitempty" ak:"required,src=FortinetFWServiceValidator.java:33-43"`
	LicenseType      string                            `json:"licenseType" ak:"required,enum=BRING_YOUR_OWN|PAY_AS_YOU_GO,src=FortinetFWService.java:108-109|Defaults.java:671-673"`
	Instances        []ServiceFortinetInstanceConfig   `json:"instances" ak:"required,minitems=1,src=FortinetFWService.java:111-112|FortinetFWServiceLogicImpl.java:852-854"`
	AutoScale        *string                           `json:"autoScale,omitempty" ak:"default=OFF,enum=OFF,src=FortinetFWService.java:114-117|FortinetFWServiceLogicImpl.java:809-811"`
	MaxInstanceCount *int64                            `json:"maxInstanceCount,omitempty" ak:"required,min=1,max=8,src=FortinetFWService.java:119-121"`
	MinInstanceCount *int64                            `json:"minInstanceCount,omitempty" ak:"required,min=1,src=FortinetFWService.java:123-124"`
	Segments         []string                          `json:"segments" ak:"required,minitems=1,maxitems=31,src=FortinetFWService.java:126-127"`
	SegmentOptions   map[string]FirewallSegmentOptions `json:"segmentOptions,omitempty"`
	BillingTags      []int64                           `json:"billingTags,omitempty" ak:"flag=BILLING-TAGS"`
	TunnelProtocol   *string                           `json:"tunnelProtocol,omitempty" ak:"default=IPSEC,enum=IPSEC|GRE|VXLAN_GPE|VXLAN,src=FortinetFWService.java:140-142|Defaults.java:1383-1387"`
	// AllowList set while the flag is off is answered 403 (AllowListFeatureFlagGuard.java:13-17).
	AllowList []string `json:"allowList,omitempty" ak:"flag=RELEASE-SERVICE-FTNTFW-ALLOW-LIST"`
	// Scheme TERM_BASED with BRING_YOUR_OWN needs serialNumber and credentialId on each instance.
	Scheme *string `json:"scheme,omitempty" ak:"default=TERM_BASED,enum=TERM_BASED|POINT_BASED,src=FortinetFWService.java:160-166"`

	InternalName  string `json:"internalName,omitempty" ak:"readonly"`
	State         string `json:"state,omitempty" ak:"readonly"`
	PreviousState string `json:"previousState,omitempty" ak:"readonly"`
}

// FortinetManagementServerConfig is FortinetFWService.managementServer;
// segment names one of segments.
type FortinetManagementServerConfig struct {
	IPAddress *string `json:"ipAddress,omitempty"`
	Segment   string  `json:"segment" ak:"required,immutable,maxlen=256,src=FortinetFWService.java:229-234"`
}

// ServiceFortinetInstanceConfig is one Fortinet instance
// (FortinetFWServiceInstance.java); an update keeps an instance by sending
// its id. TPS stores hostName as the instance name.
type ServiceFortinetInstanceConfig struct {
	ID           json.Number `json:"id,omitempty"`
	Name         string      `json:"name" ak:"required,maxlen=256,src=FortinetFWServiceInstance.java:39-43"`
	HostName     string      `json:"hostName" ak:"required,maxlen=256,src=FortinetFWServiceInstance.java:45-48"`
	SerialNumber *string     `json:"serialNumber,omitempty"`
	// CredentialID is a /credentials/ftntfw-instance credential, required for BRING_YOUR_OWN.
	CredentialID   *string `json:"credentialId,omitempty"`
	TrafficEnabled *bool   `json:"trafficEnabled,omitempty"`

	UniqueID     *string `json:"uniqueId,omitempty" ak:"readonly"`
	InternalName *string `json:"internalName,omitempty" ak:"readonly"`
	State        *string `json:"state,omitempty" ak:"readonly"`
}

// serviceFortinetDescriptor: the shared firewall rules (firewallDescriptor)
// and the portal's flag shaping; Fortinet saves synchronously.
var serviceFortinetDescriptor = func() Descriptor {
	d := firewallDescriptor("ftnt-fw-services", "FTNTFW")
	d.CrossCheck = checkServiceFortinet
	const add = "src/views/network/provision/objectConfig/fortinetConfig/add/index.js"
	d.Flags = []FlagRule{
		{Flag: "BILLING-TAGS", Effect: FlagIncludeField, Field: "billingTags", UISource: add + ":288|:961"},
		{Flag: "RELEASE-SERVICE-FTNTFW-ALLOW-LIST", Effect: FlagIncludeField, Field: "allowList", Refuse: true, UISource: add + ":304|:965-968"},
	}
	return d
}()

// checkServiceFortinet refuses what FortinetFWServiceValidator.java:39-86
// refuses: instance counts, min other than max, a management segment outside
// segments, a BRING_YOUR_OWN instance without credentialId or, with scheme
// TERM_BASED, without serialNumber, and GRE on an Azure CXP.
func checkServiceFortinet(m map[string]any) error {
	errs := append(checkInstanceCounts(m, true), checkManagementSegment(m)...)
	if m["licenseType"] == "BRING_YOUR_OWN" {
		termBased := m["scheme"] == nil || m["scheme"] == "TERM_BASED"
		insts, _ := m["instances"].([]any)
		for i, e := range insts {
			in, _ := e.(map[string]any)
			if s, _ := in["credentialId"].(string); s == "" {
				errs = append(errs, fmt.Errorf("instances[%d].credentialId is required with BRING_YOUR_OWN", i))
			}
			if s, _ := in["serialNumber"].(string); s == "" && termBased {
				errs = append(errs, fmt.Errorf("instances[%d].serialNumber is required with BRING_YOUR_OWN and scheme TERM_BASED", i))
			}
		}
	}
	if cxp, _ := m["cxp"].(string); m["tunnelProtocol"] == "GRE" && strings.Contains(strings.ToLower(cxp), "azure") {
		errs = append(errs, errors.New("tunnelProtocol GRE is not available on an Azure CXP"))
	}
	return errors.Join(errs...)
}

// ServiceFortinetDescriptor returns the Fortinet firewall service write descriptor.
func ServiceFortinetDescriptor() Descriptor { return serviceFortinetDescriptor.clone() }

// CreateServiceFortinet creates a Fortinet firewall service.
func (ac *AlkiraClient) CreateServiceFortinet(ctx context.Context, s *ServiceFortinetConfig) (*ServiceFortinetConfig, *WriteResult, error) {
	return serviceCreate(ctx, ac, serviceFortinetDescriptor, s)
}

// GetServiceFortinet reads one Fortinet firewall service and its raw body.
func (ac *AlkiraClient) GetServiceFortinet(ctx context.Context, id string) (*ServiceFortinetConfig, []byte, error) {
	return serviceGet[ServiceFortinetConfig](ctx, ac, serviceFortinetDescriptor, id)
}

// UpdateServiceFortinet applies a merge patch to a Fortinet firewall service.
func (ac *AlkiraClient) UpdateServiceFortinet(ctx context.Context, id string, patch json.RawMessage) (*ServiceFortinetConfig, *WriteResult, error) {
	return serviceUpdate[ServiceFortinetConfig](ctx, ac, serviceFortinetDescriptor, id, patch)
}

// DeleteServiceFortinet deletes a Fortinet firewall service; a deployed one
// is marked for deletion until the next provision.
func (ac *AlkiraClient) DeleteServiceFortinet(ctx context.Context, id string) (*WriteResult, error) {
	return serviceDelete[ServiceFortinetConfig](ctx, ac, serviceFortinetDescriptor, id)
}

// ListServiceFortinetPaged reads every Fortinet firewall service through paged reads.
func (ac *AlkiraClient) ListServiceFortinetPaged(ctx context.Context) ([]ServiceFortinetConfig, error) {
	return serviceList[ServiceFortinetConfig](ctx, ac, serviceFortinetDescriptor)
}
