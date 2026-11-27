// Copyright (C) 2020-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
)

type ServicePANSummary struct {
	Id           json.Number `json:"id"`
	Name         string      `json:"name"`
	InternalName string      `json:"internalName"`
	CXP          string      `json:"cxp"`
	LicenseType  string      `json:"licenseType"`
}

// NewServicePAN new service pan
func NewServicePAN(ac *AlkiraClient) *AlkiraApi[ServicePANSummary] {
	uri := fmt.Sprintf("%s/v1/tenantnetworks/%s/panfwservices", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[ServicePANSummary]{ac, uri, PaginationOn}
	return api
}

// ServicePanConfig is the write model of a PAN firewall service (TPS
// model/PANFWService.java; GET /services type PAN). Optional fields are
// pointers or omitempty, so a zero value is never sent. TPS keeps the stored
// cxp, licence, Panorama, management segment, credential and SCM fields on
// update (PANFWServiceLogicImpl.java:654-697), so they are immutable here.
// licenseKey, which the portal also sends, is not a TPS field; the licence
// key lives in the /credentials/pan credential.
type ServicePanConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=60,charset=hyphen,src=PANFWService.java:71-74"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,charset=description,src=PANFWService.java:76-77"`
	CXP         string      `json:"cxp" ak:"required,immutable,maxlen=128,trimmed,src=PANFWService.java:80-82"`
	// Segments are segment ids and include ManagementSegment (PanFwServiceValidator.java:154).
	// Item and instance counts above 1 are tenant soft limits that TPS checks
	// (SoftLimitValidator.java:60-69), so no tag caps them.
	Segments            []int64  `json:"segments" ak:"required,minitems=1,src=PANFWServiceLogicImpl.java:277"`
	ManagementSegment   *int64   `json:"managementSegment,omitempty" ak:"required,immutable,min=1,src=PANFWService.java:100-101"`
	PanoramaEnabled     *bool    `json:"panoramaEnabled,omitempty" ak:"immutable,src=PANFWServiceLogicImpl.java:654-697"`
	PanoramaIPAddresses []string `json:"panoramaIPAddresses,omitempty"`
	PanoramaDeviceGroup *string  `json:"panoramaDeviceGroup,omitempty" ak:"immutable,src=PANFWServiceLogicImpl.java:654-697"`
	PanoramaTemplate    *string  `json:"panoramaTemplate,omitempty" ak:"immutable,src=PANFWServiceLogicImpl.java:654-697"`
	ScmEnabled          *bool    `json:"scmEnabled,omitempty" ak:"immutable,flag=RELEASE-SERVICE-PAN-SCM-ADVANCED-ROUTING,src=PANFWServiceLogicImpl.java:654-697"`
	ScmFolder           *string  `json:"scmFolder,omitempty" ak:"immutable,flag=RELEASE-SERVICE-PAN-SCM-ADVANCED-ROUTING,src=PANFWServiceLogicImpl.java:654-697"`
	RoutingType         *string  `json:"routingType,omitempty" ak:"immutable,enum=|advanced,flag=RELEASE-SERVICE-PAN-SCM-ADVANCED-ROUTING,src=PANFWService.java:111-112"`
	LicenseType         string   `json:"licenseType" ak:"required,immutable,enum=BRING_YOUR_OWN|PAY_AS_YOU_GO,src=PANFWService.java:121|Defaults.java:671-673"`
	// SubLicenseType is required for BRING_YOUR_OWN and refused otherwise (PANFWServiceLogicImpl.java:256-262).
	SubLicenseType *string `json:"subLicenseType,omitempty" ak:"immutable,enum=CREDIT_BASED|MODEL_BASED,src=PANFWService.java:215-217"`
	Version        string  `json:"version" ak:"required,src=PanFwServiceValidator.java:123"`
	Type           *string `json:"type,omitempty"`
	Bundle         *string `json:"bundle,omitempty" ak:"enum=VM_SERIES_BUNDLE_1|VM_SERIES_BUNDLE_2|PAN_VM_300_BUNDLE_2,src=Defaults.java:1420-1423"`
	// CredentialID is a /credentials/pan credential.
	CredentialID     string  `json:"credentialId" ak:"required,immutable,src=PANFWService.java:129-130"`
	Size             *string `json:"size,omitempty" ak:"default=SMALL,enum=SMALL|MEDIUM|LARGE|2LARGE|4LARGE|5LARGE,src=PANFWService.java:140-143|Defaults.java:124-134"`
	MaxInstanceCount *int64  `json:"maxInstanceCount,omitempty" ak:"required,min=1,src=PANFWService.java:114-117"`
	// MinInstanceCount defaults to 1, the least the portal sends
	// (panConfig/AddFirewall/index.js:1226); the backend's primitive int has
	// no minimum and stores an omitted value as 0 (PANFWService.java:118).
	MinInstanceCount *int64  `json:"minInstanceCount,omitempty" ak:"default=1,min=1,src=PANFWService.java:118|AddFirewall/index.js:1226"`
	TunnelProtocol   *string `json:"tunnelProtocol,omitempty" ak:"default=IPSEC,enum=IPSEC|GRE|VXLAN_GPE|VXLAN,src=PANFWService.java:156-157|Defaults.java:1383"`
	// SegmentOptions maps a segment name to its zones and groups.
	SegmentOptions map[string]FirewallSegmentOptions `json:"segmentOptions,omitempty"`
	BillingTags    []int64                           `json:"billingTags,omitempty" ak:"flag=BILLING-TAGS"`
	// AllowList set while the flag is off is answered 403 (AllowListFeatureFlagGuard.java:13-17).
	AllowList                   []string                          `json:"allowList,omitempty" ak:"flag=RELEASE-SERVICE-PAN-ALLOW-LIST"`
	GlobalProtectEnabled        *bool                             `json:"globalProtectEnabled,omitempty"`
	GlobalProtectSegmentOptions map[string]PanGlobalProtectConfig `json:"globalProtectSegmentOptions,omitempty"`
	MasterKeyEnabled            *bool                             `json:"masterKeyEnabled,omitempty"`
	// MasterKeyCredentialID is a /credentials/pan-masterkey credential.
	MasterKeyCredentialID *string `json:"masterKeyCredentialId,omitempty"`
	// RegistrationCredentialID is a /credentials/pan-registration credential,
	// required from version 9.1 (PANFWServiceLogicImpl.java:723-732).
	RegistrationCredentialID *string                    `json:"registrationCredentialId,omitempty"`
	Instances                []ServicePanInstanceConfig `json:"instances" ak:"required,minitems=1,src=PANFWService.java:132-133|PANFWServiceLogicImpl.java:273"`

	InternalName       string          `json:"internalName,omitempty" ak:"readonly"`
	Az                 *string         `json:"az,omitempty" ak:"readonly"`
	ManagementIP       *string         `json:"managementIP,omitempty" ak:"readonly"`
	PanWarmBootEnabled *bool           `json:"panWarmBootEnabled,omitempty" ak:"readonly"`
	ValidationState    json.RawMessage `json:"validationState,omitempty" ak:"readonly"`
	// State and PreviousState come from the GET /services row (Descriptor.ServiceType).
	State         string `json:"state,omitempty" ak:"readonly"`
	PreviousState string `json:"previousState,omitempty" ak:"readonly"`
}

// ServicePanInstanceConfig is one PAN instance (PANFWServiceInstance.java);
// an update keeps an instance by sending its id.
type ServicePanInstanceConfig struct {
	ID   json.Number `json:"id,omitempty"`
	Name string      `json:"name" ak:"required,maxlen=256,trimmed,src=PANFWServiceInstance.java:43-47"`
	// CredentialID is a /credentials/paninstance credential.
	CredentialID                string                                    `json:"credentialId" ak:"required,src=PANFWServiceInstance.java:49-50"`
	TrafficEnabled              *bool                                     `json:"trafficEnabled,omitempty"`
	GlobalProtectSegmentOptions map[string]PanGlobalProtectInstanceConfig `json:"globalProtectSegmentOptions,omitempty"`

	UniqueID         *string `json:"uniqueId,omitempty" ak:"readonly"`
	InternalName     *string `json:"internalName,omitempty" ak:"readonly"`
	HostName         *string `json:"hostName,omitempty" ak:"readonly"`
	Dormant          *bool   `json:"dormant,omitempty" ak:"readonly"`
	MasterKeyEnabled *bool   `json:"masterKeyEnabled,omitempty" ak:"readonly"`
}

// PanGlobalProtectConfig is a globalProtectSegmentOptions entry, keyed by segment name.
type PanGlobalProtectConfig struct {
	RemoteUserZoneName *string `json:"remoteUserZoneName,omitempty"`
	PortalFqdnPrefix   *string `json:"portalFqdnPrefix,omitempty" ak:"maxlen=63,src=PanFwServiceValidator.java:27-68"`
	ServiceGroupName   *string `json:"serviceGroupName,omitempty"`

	SegmentID          *int64 `json:"segmentId,omitempty" ak:"readonly"`
	RemoteUserTunnelID *int64 `json:"remoteUserTunnelId,omitempty" ak:"readonly"`
	ServiceGroupID     *int64 `json:"serviceGroupId,omitempty" ak:"readonly"`
}

// PanGlobalProtectInstanceConfig is an instance's globalProtectSegmentOptions entry.
type PanGlobalProtectInstanceConfig struct {
	PortalEnabled  *bool  `json:"portalEnabled,omitempty"`
	GatewayEnabled *bool  `json:"gatewayEnabled,omitempty"`
	PrefixListID   *int64 `json:"prefixListId,omitempty"`

	SegmentID *int64 `json:"segmentId,omitempty" ak:"readonly"`
}

// servicePanDescriptor: the shared firewall rules (firewallDescriptor), the
// portal's flag shaping, and async validation (PANFWServiceLogicImpl.java:594-597).
var servicePanDescriptor = func() Descriptor {
	d := firewallDescriptor("panfwservices", "PAN")
	d.AsyncCapable = true
	d.CrossCheck = checkServicePan
	const add = "src/views/network/provision/objectConfig/panConfig/AddFirewall/index.js"
	d.Flags = []FlagRule{
		{Flag: "RELEASE-SERVICE-PAN-ALLOW-LIST", Effect: FlagIncludeField, Field: "allowList", Refuse: true, UISource: add + ":163|:1240-1243"},
		{Flag: "BILLING-TAGS", Effect: FlagIncludeField, Field: "billingTags", UISource: add + ":227|:1228"},
		{Flag: "RELEASE-SERVICE-PAN-SCM-ADVANCED-ROUTING", Effect: FlagIncludeField, Field: "scmEnabled", UISource: add + ":1235-1239"},
		{Flag: "RELEASE-SERVICE-PAN-SCM-ADVANCED-ROUTING", Effect: FlagIncludeField, Field: "routingType", UISource: add + ":1235-1239"},
		{Flag: "RELEASE-SERVICE-PAN-SCM-ADVANCED-ROUTING", Effect: FlagIncludeField, Field: "scmFolder", UISource: add + ":1235-1239"},
	}
	return d
}()

// checkServicePan refuses instance counts the validator refuses and a
// managementSegment missing from segments (PanFwServiceValidator.java:138-154).
func checkServicePan(m map[string]any) error {
	errs := checkInstanceCounts(m, false)
	if ms, ok := m["managementSegment"].(json.Number); ok {
		segs, _ := m["segments"].([]any)
		if !slices.Contains(segs, any(ms)) {
			errs = append(errs, fmt.Errorf("segments must include managementSegment %s", ms))
		}
	}
	return errors.Join(errs...)
}

// ServicePanDescriptor returns the PAN firewall service write descriptor.
func ServicePanDescriptor() Descriptor { return servicePanDescriptor.clone() }

// CreateServicePan creates a PAN firewall service.
func (ac *AlkiraClient) CreateServicePan(ctx context.Context, s *ServicePanConfig) (*ServicePanConfig, *WriteResult, error) {
	return serviceCreate(ctx, ac, servicePanDescriptor, s)
}

// GetServicePan reads one PAN firewall service and its raw body.
func (ac *AlkiraClient) GetServicePan(ctx context.Context, id string) (*ServicePanConfig, []byte, error) {
	return serviceGet[ServicePanConfig](ctx, ac, servicePanDescriptor, id)
}

// UpdateServicePan applies a merge patch to a PAN firewall service.
func (ac *AlkiraClient) UpdateServicePan(ctx context.Context, id string, patch json.RawMessage) (*ServicePanConfig, *WriteResult, error) {
	return serviceUpdate[ServicePanConfig](ctx, ac, servicePanDescriptor, id, patch)
}

// DeleteServicePan deletes a PAN firewall service; a deployed one is marked
// for deletion until the next provision.
func (ac *AlkiraClient) DeleteServicePan(ctx context.Context, id string) (*WriteResult, error) {
	return serviceDelete[ServicePanConfig](ctx, ac, servicePanDescriptor, id)
}

// ListServicePanPaged reads every PAN firewall service through paged reads.
func (ac *AlkiraClient) ListServicePanPaged(ctx context.Context) ([]ServicePanConfig, error) {
	return serviceList[ServicePanConfig](ctx, ac, servicePanDescriptor)
}
