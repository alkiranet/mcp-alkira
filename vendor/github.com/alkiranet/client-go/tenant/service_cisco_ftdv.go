// Copyright (C) 2022-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

type CiscoFTDvInstance struct {
	Id             int    `json:"id,omitempty"`           // response only
	CredentialId   string `json:"credentialId,omitempty"` // response only
	InternalName   string `json:"internalName,omitempty"` // response only
	State          string `json:"state,omitempty"`        // response only
	Hostname       string `json:"hostName"`
	LicenseType    string `json:"licenseType"`
	Version        string `json:"version"`
	TrafficEnabled bool   `json:"trafficEnabled"`
}

type CiscoFTDvManagementServer struct {
	IPAddress string `json:"ipAddress"`
	Segment   string `json:"segment"`
	SegmentId int    `json:"segmentId"`
}

type ServiceCiscoFTDv struct {
	Id               json.Number               `json:"id,omitempty"` // response only
	Name             string                    `json:"name"`
	GlobalCidrListId int                       `json:"globalCidrListId"`
	Size             string                    `json:"size"`
	CredentialId     string                    `json:"credentialId,omitempty"` // response only
	Cxp              string                    `json:"cxp"`
	ManagementServer CiscoFTDvManagementServer `json:"managementServer"`
	IpAllowList      []string                  `json:"servicesIpAllowList"`
	MaxInstanceCount int                       `json:"maxInstanceCount"`
	MinInstanceCount int                       `json:"minInstanceCount"`
	Segments         []string                  `json:"segments"`
	SegmentOptions   SegmentNameToZone         `json:"segmentOptions,omitempty"`
	Instances        []CiscoFTDvInstance       `json:"instances"`
	BillingTags      []int                     `json:"billingTags"`
	TunnelProtocol   string                    `json:"tunnelProtocol"`
	AutoScale        string                    `json:"autoScale"`
	InternalName     string                    `json:"internalName,omitempty"` // response only
	State            string                    `json:"state,omitempty"`        // response only
	Description      string                    `json:"description,omitempty"`
}

type ServiceCiscoFTDvSummary struct {
	Id               json.Number `json:"id"`
	Name             string      `json:"name"`
	GlobalCidrListId int         `json:"globalCidrListId"`
	Size             string      `json:"size"`
	Cxp              string      `json:"cxp"`
	Segments         []string    `json:"segments"`
	BillingTags      []int       `json:"billingTags"`
	InternalName     string      `json:"internalName,omitempty"`
	State            string      `json:"state,omitempty"`
}

// NewServiceCiscoFTDv new service cisco FTDv
func NewServiceCiscoFTDv(ac *AlkiraClient) *AlkiraApi[ServiceCiscoFTDvSummary] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/cisco-ftdv-fw-services", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[ServiceCiscoFTDvSummary]{ac, uri, PaginationOn}
	return api
}

// ServiceCiscoFTDvConfig is the write model of a Cisco FTDv firewall service
// (TPS model/ciscoftdvfw/CiscoFTDvFWService.java; GET /services type
// CISCO_FTDV_FW). TPS refuses autoScale ON and GRE (CiscoFTDvFWServiceValidator.java:45, :72).
// After the first provision name, cxp, managementServer and globalCidrListId
// cannot change (CiscoFTDvFWServiceLogicImpl.java:689-696).
type ServiceCiscoFTDvConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=60,charset=hyphen,src=CiscoFTDvFWService.java:72-75"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,charset=description,src=CiscoFTDvFWService.java:77-79"`
	// GlobalCidrListID is a global CIDR list of the service CXP tagged
	// CISCO_FTDV_FW with prefixes of /24 or shorter (CiscoFTDvFWServiceLogicImpl.java:792-820).
	GlobalCidrListID *int64  `json:"globalCidrListId,omitempty" ak:"required,src=CiscoFTDvFWService.java:81"`
	Size             *string `json:"size,omitempty" ak:"default=SMALL,enum=SMALL|MEDIUM|LARGE|2LARGE,src=CiscoFTDvFWService.java:84-87|Defaults.java:122-145"`
	CXP              string  `json:"cxp" ak:"required,immutable,maxlen=128,trimmed,src=CiscoFTDvFWService.java:89-92"`
	// CredentialID is an optional /credentials/cisco-ftdv-fw (FMC) credential.
	CredentialID        *string                          `json:"credentialId,omitempty"`
	ManagementServer    *CiscoFTDvManagementServerConfig `json:"managementServer,omitempty" ak:"required,src=CiscoFTDvFWServiceValidator.java:31-40"`
	ServicesIPAllowList []string                         `json:"servicesIpAllowList,omitempty"`
	Instances           []ServiceCiscoFTDvInstanceConfig `json:"instances" ak:"required,minitems=1,src=CiscoFTDvFWService.java:104-105|CiscoFTDvFWServiceLogicImpl.java:711-712"`
	MaxInstanceCount    *int64                           `json:"maxInstanceCount,omitempty" ak:"required,min=1,src=CiscoFTDvFWService.java:107-110"`
	MinInstanceCount    *int64                           `json:"minInstanceCount,omitempty" ak:"required,min=1,src=CiscoFTDvFWService.java:112-113"`
	Segments            []string                         `json:"segments" ak:"required,minitems=1,src=CiscoFTDvFWService.java:115-117"`
	// SegmentOptions needs at least one zone on every segment (CiscoFTDvFWServiceZoneValidator.java:212-215).
	SegmentOptions map[string]FirewallSegmentOptions `json:"segmentOptions" ak:"required,src=CiscoFTDvFWService.java:119-124"`
	BillingTags    []int64                           `json:"billingTags,omitempty"`
	TunnelProtocol *string                           `json:"tunnelProtocol,omitempty" ak:"default=IPSEC,enum=IPSEC|VXLAN_GPE|VXLAN,src=CiscoFTDvFWService.java:131-133|CiscoFTDvFWServiceValidator.java:72"`
	AutoScale      *string                           `json:"autoScale,omitempty" ak:"default=OFF,enum=OFF,src=CiscoFTDvFWService.java:135-137|CiscoFTDvFWServiceValidator.java:45"`

	InternalName  string `json:"internalName,omitempty" ak:"readonly"`
	State         string `json:"state,omitempty" ak:"readonly"`
	PreviousState string `json:"previousState,omitempty" ak:"readonly"`
}

// CiscoFTDvManagementServerConfig is CiscoFTDvFWService.managementServer;
// segment names one of segments.
type CiscoFTDvManagementServerConfig struct {
	IPAddress *string `json:"ipAddress,omitempty"`
	Segment   string  `json:"segment" ak:"required,maxlen=256,src=CiscoFTDvFWService.java:145-154"`

	SegmentID *int64 `json:"segmentId,omitempty" ak:"readonly"`
}

// ServiceCiscoFTDvInstanceConfig is one FTDv instance
// (CiscoFTDvFWServiceInstance.java); an update keeps an instance by sending its id.
type ServiceCiscoFTDvInstanceConfig struct {
	ID          json.Number `json:"id,omitempty"`
	HostName    string      `json:"hostName" ak:"required,maxlen=256,src=CiscoFTDvFWServiceInstance.java:41-44"`
	LicenseType string      `json:"licenseType" ak:"required,enum=BRING_YOUR_OWN|PAY_AS_YOU_GO,src=CiscoFTDvFWServiceInstance.java:35|Defaults.java:671-673"`
	Version     string      `json:"version" ak:"required,maxlen=32,src=CiscoFTDvFWServiceInstance.java:49-51"`
	// CredentialID is a /credentials/cisco-ftdv-fw-instance credential.
	CredentialID   string `json:"credentialId" ak:"required,src=CiscoFTDvFWServiceInstance.java:46"`
	TrafficEnabled *bool  `json:"trafficEnabled,omitempty"`

	InternalName *string `json:"internalName,omitempty" ak:"readonly"`
	State        *string `json:"state,omitempty" ak:"readonly"`
	ProductID    *string `json:"productId,omitempty" ak:"readonly"`
}

// serviceCiscoFTDvDescriptor: the shared firewall rules (firewallDescriptor);
// the portal gates no FTDv field on a flag, and FTDv saves synchronously.
var serviceCiscoFTDvDescriptor = func() Descriptor {
	d := firewallDescriptor("cisco-ftdv-fw-services", "CISCO_FTDV_FW")
	d.CrossCheck = func(m map[string]any) error {
		return errors.Join(append(checkInstanceCounts(m, true), checkManagementSegment(m)...)...)
	}
	return d
}()

// ServiceCiscoFTDvDescriptor returns the Cisco FTDv firewall service write descriptor.
func ServiceCiscoFTDvDescriptor() Descriptor { return serviceCiscoFTDvDescriptor.clone() }

// CreateServiceCiscoFTDv creates a Cisco FTDv firewall service.
func (ac *AlkiraClient) CreateServiceCiscoFTDv(ctx context.Context, s *ServiceCiscoFTDvConfig) (*ServiceCiscoFTDvConfig, *WriteResult, error) {
	return serviceCreate(ctx, ac, serviceCiscoFTDvDescriptor, s)
}

// GetServiceCiscoFTDv reads one Cisco FTDv firewall service and its raw body.
func (ac *AlkiraClient) GetServiceCiscoFTDv(ctx context.Context, id string) (*ServiceCiscoFTDvConfig, []byte, error) {
	return serviceGet[ServiceCiscoFTDvConfig](ctx, ac, serviceCiscoFTDvDescriptor, id)
}

// UpdateServiceCiscoFTDv applies a merge patch to a Cisco FTDv firewall service.
func (ac *AlkiraClient) UpdateServiceCiscoFTDv(ctx context.Context, id string, patch json.RawMessage) (*ServiceCiscoFTDvConfig, *WriteResult, error) {
	return serviceUpdate[ServiceCiscoFTDvConfig](ctx, ac, serviceCiscoFTDvDescriptor, id, patch)
}

// DeleteServiceCiscoFTDv deletes a Cisco FTDv firewall service; a deployed
// one is marked for deletion until the next provision.
func (ac *AlkiraClient) DeleteServiceCiscoFTDv(ctx context.Context, id string) (*WriteResult, error) {
	return serviceDelete[ServiceCiscoFTDvConfig](ctx, ac, serviceCiscoFTDvDescriptor, id)
}

// ListServiceCiscoFTDvPaged reads every Cisco FTDv firewall service through paged reads.
func (ac *AlkiraClient) ListServiceCiscoFTDvPaged(ctx context.Context) ([]ServiceCiscoFTDvConfig, error) {
	return serviceList[ServiceCiscoFTDvConfig](ctx, ac, serviceCiscoFTDvDescriptor)
}
