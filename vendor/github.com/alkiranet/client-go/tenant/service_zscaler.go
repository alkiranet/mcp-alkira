// Copyright (C) 2022-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

type ServiceZscaler struct {
	BillingTags           []int               `json:"billingTags"`
	Cxp                   string              `json:"cxp"`
	Description           string              `json:"description"`
	Id                    json.Number         `json:"id,omitempty"`           // only set on response
	InternalName          string              `json:"internalName,omitempty"` // only set on response
	IpsecConfiguration    *ZscalerIpSecConfig `json:"ipsecConfiguration"`
	Name                  string              `json:"name"`
	PrimaryPublicEdgeIp   string              `json:"primaryPublicEdgeIp"`
	SecondaryPublicEdgeIp string              `json:"secondaryPublicEdgeIp"`
	Segments              []string            `json:"segments"`
	Size                  string              `json:"size"`
	TunnelType            string              `json:"tunnelType"`
}

type ZscalerIpSecConfig struct {
	EspDhGroupNumber       string `json:"espDhGroupNumber"`
	EspEncryptionAlgorithm string `json:"espEncryptionAlgorithm"`
	EspIntegrityAlgorithm  string `json:"espIntegrityAlgorithm"`
	HealthCheckType        string `json:"healthCheckType"`
	HttpProbeUrl           string `json:"httpProbeUrl"`
	IkeDhGroupNumber       string `json:"ikeDhGroupNumber"`
	IkeEncryptionAlgorithm string `json:"ikeEncryptionAlgorithm"`
	IkeIntegrityAlgorithm  string `json:"ikeIntegrityAlgorithm"`
	LocalFqdnId            string `json:"localFqdnId"`
	PreSharedKey           string `json:"preSharedKey"`
	PingProbeIp            string `json:"pingProbeIp,omitempty"`
}

type ServiceZscalerSummary struct {
	Cxp          string      `json:"cxp"`
	Id           json.Number `json:"id,omitempty"`           // only set on response
	InternalName string      `json:"internalName,omitempty"` // only set on response
	Name         string      `json:"name"`
	Segments     []string    `json:"segments"`
}

// NewServiceZscaler new service zscaler
func NewServiceZscaler(ac *AlkiraClient) *AlkiraApi[ServiceZscalerSummary] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/zscaler-internet-access-services", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[ServiceZscalerSummary]{ac, uri, PaginationOn}
	return api
}

// ServiceZscalerConfig is the write model of a Zscaler internet access
// service (TPS model/ZscalerInternetAccessService.java). The backend keeps
// the stored cxp on update whatever the request sends
// (ServiceLogicImpl.java:619-624, :891).
type ServiceZscalerConfig struct {
	ID                    json.Number          `json:"id,omitempty" ak:"readonly"`
	Name                  string               `json:"name" ak:"required,maxlen=60,charset=hyphen,src=ZscalerInternetAccessService.java:50-53"`
	Description           *string              `json:"description,omitempty" ak:"maxlen=512,charset=description,src=ZscalerInternetAccessService.java:55-57"`
	CXP                   string               `json:"cxp" ak:"required,immutable,maxlen=128,trimmed,src=ZscalerInternetAccessService.java:59-62"`
	Size                  *string              `json:"size,omitempty" ak:"default=SMALL,enum=SMALL|MEDIUM|LARGE|2LARGE,src=ZscalerInternetAccessService.java:64-67"`
	BillingTags           []int64              `json:"billingTags,omitempty" ak:"maxitems=5,flag=BILLING-TAGS,src=ZscalerInternetAccessService.java:69-72"`
	Segments              []string             `json:"segments" ak:"required,minitems=1,maxitems=1,trimmed,src=ZscalerInternetAccessService.java:74-80"`
	PrimaryPublicEdgeIP   string               `json:"primaryPublicEdgeIp" ak:"required,src=ZscalerInternetAccessService.java:86-88"`
	SecondaryPublicEdgeIP string               `json:"secondaryPublicEdgeIp" ak:"required,src=ZscalerInternetAccessService.java:90-92"`
	TunnelType            *string              `json:"tunnelType,omitempty" ak:"default=IPSEC,enum=IPSEC,src=ZscalerInternetAccessServiceValidator.java:20-24"`
	IpsecConfiguration    *ServiceZscalerIpsec `json:"ipsecConfiguration" ak:"required,src=ZscalerInternetAccessServiceValidator.java:26-30"`

	InternalName string  `json:"internalName,omitempty" ak:"readonly"`
	SegmentIDs   []int64 `json:"segmentIds,omitempty" ak:"readonly"`
	// State and PreviousState come from the GET /services row (Descriptor.ServiceType).
	State         string `json:"state,omitempty" ak:"readonly"`
	PreviousState string `json:"previousState,omitempty" ak:"readonly"`
}

// ServiceZscalerIpsec is ZscalerInternetAccessService.IPSecConfiguration;
// the defaults are the portal's (ziaConfig/add/index.js:108-132).
type ServiceZscalerIpsec struct {
	LocalFqdnID string `json:"localFqdnId" ak:"required,src=ZscalerInternetAccessService.java:108-109"`
	// PreSharedKey must match the key set in ZIA. It is write-only: a
	// create must set it, as the portal requires
	// (ziaConfig/SecurityConfiguration.js:139-148; TPS stores a missing key
	// as null and the tunnel never comes up), a patch cannot, and an update
	// resends the stored value.
	PreSharedKey           *string `json:"preSharedKey,omitempty" ak:"required,writeonly,maxlen=40,charset=psk,src=ZscalerInternetAccessService.java:111-113"`
	IkeEncryptionAlgorithm *string `json:"ikeEncryptionAlgorithm,omitempty" ak:"default=AES256CBC,enum=AES256CBC,src=ZscalerInternetAccessService.java:115-116|:143-145"`
	IkeIntegrityAlgorithm  *string `json:"ikeIntegrityAlgorithm,omitempty" ak:"default=SHA256,enum=SHA256,src=ZscalerInternetAccessService.java:118-119|:147-149"`
	IkeDhGroupNumber       *string `json:"ikeDhGroupNumber,omitempty" ak:"default=MODP1024,enum=MODP1024|MODP2048,src=ZscalerInternetAccessService.java:121-122|:151-154"`
	EspEncryptionAlgorithm *string `json:"espEncryptionAlgorithm,omitempty" ak:"default=NULL,enum=NULL|AES256CBC,src=ZscalerInternetAccessService.java:124-125|:156-159"`
	EspIntegrityAlgorithm  *string `json:"espIntegrityAlgorithm,omitempty" ak:"default=MD5,enum=MD5|SHA256,src=ZscalerInternetAccessService.java:127-128|:161-164"`
	EspDhGroupNumber       *string `json:"espDhGroupNumber,omitempty" ak:"default=MODP1024,enum=MODP1024|MODP2048,src=ZscalerInternetAccessService.java:130-131|:167-170"`
	HealthCheckType        *string `json:"healthCheckType,omitempty" ak:"default=IKE_STATUS,enum=IKE_STATUS|PING_PROBE|HTTP_PROBE,src=ZscalerInternetAccessService.java:133-134|:172-176"`
	PingProbeIP            *string `json:"pingProbeIp,omitempty"`
	HTTPProbeURL           *string `json:"httpProbeUrl,omitempty"`
}

// serviceZscalerDescriptor: services share one id space and a GET of
// another service's id answers 200 here (ServiceLogicImpl.java:1354-1365),
// so primaryPublicEdgeIp, which only this family has, identifies it. The
// type, state and previousState come from GET /services/{id} (ServiceType
// ZIA, Defaults.java:456), so an update runs the D13 state pre-check, as
// TPS does (ProvisionedEntityValidatorImpl.java:34-73).
// Names are unique across every service type, so the duplicate guard reads
// GET /services; the backend answers 409 (:2174-2184). A missing
// id answers 404 on GET, PUT and DELETE (:1357-1365, :864-865;
// AbstractBaseService.java:366).
var serviceZscalerDescriptor = Descriptor{
	Collection: "zscaler-internet-access-services",
	FamilyType: FamilyCheck{Field: "primaryPublicEdgeIp"},
	InUse: []Signature{
		{Status: 409, Code: "AK000001", Message: "is still referred externally"},
	},
	Duplicate:          []Signature{{Status: 409, Code: "AK000001", Message: "already exists"}},
	Provisionable:      true,
	MarkedRowLifecycle: true,
	DocState:           DocStatePlainGet,
	ServiceType:        "ZIA",
	Uniqueness:         Uniqueness{Key: "name", Scope: ScopeServices},
	StateGuard:         StateGuardD13,
	Service:            ServiceTPS,
	CrossCheck:         checkServiceZscaler,
	Flags: []FlagRule{
		{Flag: "BILLING-TAGS", Effect: FlagIncludeField, Field: "billingTags", UISource: "src/views/network/provision/objectConfig/ziaConfig/add/index.js:139|:451-465"},
	},
}

// checkServiceZscaler holds the validator rules that span fields
// (ZscalerInternetAccessServiceValidator.java:15-18, :33-45).
func checkServiceZscaler(m map[string]any) error {
	var errs []error
	if p, _ := m["primaryPublicEdgeIp"].(string); p != "" && p == m["secondaryPublicEdgeIp"] {
		errs = append(errs, errors.New("primaryPublicEdgeIp and secondaryPublicEdgeIp must differ"))
	}
	ipsec, _ := m["ipsecConfiguration"].(map[string]any)
	switch ipsec["healthCheckType"] {
	case "PING_PROBE":
		if s, _ := ipsec["pingProbeIp"].(string); s == "" {
			errs = append(errs, errors.New("healthCheckType PING_PROBE needs ipsecConfiguration.pingProbeIp"))
		}
	case "HTTP_PROBE":
		if s, _ := ipsec["httpProbeUrl"].(string); s == "" {
			errs = append(errs, errors.New("healthCheckType HTTP_PROBE needs ipsecConfiguration.httpProbeUrl"))
		}
	}
	return errors.Join(errs...)
}

// ServiceZscalerDescriptor returns the Zscaler internet access service write descriptor.
func ServiceZscalerDescriptor() Descriptor { return serviceZscalerDescriptor.clone() }

// CreateServiceZscaler creates a Zscaler internet access service.
func (ac *AlkiraClient) CreateServiceZscaler(ctx context.Context, s *ServiceZscalerConfig) (*ServiceZscalerConfig, *WriteResult, error) {
	core, err := NewCore[ServiceZscalerConfig](ac, serviceZscalerDescriptor)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, s, WriteOptions{})
}

// GetServiceZscaler reads one Zscaler internet access service and its raw body.
func (ac *AlkiraClient) GetServiceZscaler(ctx context.Context, id string) (*ServiceZscalerConfig, []byte, error) {
	core, err := NewCore[ServiceZscalerConfig](ac, serviceZscalerDescriptor)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

// UpdateServiceZscaler applies a merge patch to a Zscaler internet access
// service; the stored pre-shared key is resent unchanged.
func (ac *AlkiraClient) UpdateServiceZscaler(ctx context.Context, id string, patch json.RawMessage) (*ServiceZscalerConfig, *WriteResult, error) {
	core, err := NewCore[ServiceZscalerConfig](ac, serviceZscalerDescriptor)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

// DeleteServiceZscaler deletes a Zscaler internet access service; a deployed
// one is marked for deletion until the next provision.
func (ac *AlkiraClient) DeleteServiceZscaler(ctx context.Context, id string) (*WriteResult, error) {
	core, err := NewCore[ServiceZscalerConfig](ac, serviceZscalerDescriptor)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

// ListServiceZscalerPaged reads every Zscaler internet access service through paged reads.
func (ac *AlkiraClient) ListServiceZscalerPaged(ctx context.Context) ([]ServiceZscalerConfig, error) {
	core, err := NewCore[ServiceZscalerConfig](ac, serviceZscalerDescriptor)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
