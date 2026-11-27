// Copyright (C) 2020-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
)

type ServiceF5Lb struct {
	SegmentOptions   F5SegmentOption `json:"segmentOptions"`
	Description      string          `json:"description,omitempty"`
	Cxp              string          `json:"cxp"`
	Size             string          `json:"size"`
	Id               json.Number     `json:"id,omitempty"`
	Name             string          `json:"name"`
	ServiceGroupName string          `json:"serviceGroupName"`
	ImplicitGroupId  int             `json:"implicitGroupId,omitempty"` // response only
	Instances        []F5Instance    `json:"instances"`
	Segments         []string        `json:"segments"`
	BillingTags      []int           `json:"billingTags,omitempty"`
	PrefixListId     int             `json:"prefixListId,omitempty"`
	GlobalCidrListId int             `json:"globalCidrListId"`
}

type F5Instance struct {
	Deployment               F5InstanceDeployment `json:"deployment"`
	Name                     string               `json:"name"`
	RegistrationCredentialId string               `json:"registrationCredentialId,omitempty"`
	CredentialId             string               `json:"credentialId"`
	HostNameFqdn             string               `json:"hostNameFqdn"`
	LicenseType              string               `json:"licenseType"`
	Version                  string               `json:"version"`
	Id                       int                  `json:"id,omitempty"` // RESPONSE ONLY
}

type F5SegmentOption map[string]F5SegmentSubOption
type F5SegmentSubOption struct {
	NatPoolPrefixLength int `json:"natPoolPrefixLength,omitempty"`
	ElbNicCount         int `json:"elbNicCount"`
}

type F5InstanceDeployment struct {
	Option string `json:"option,omitempty"`
	Type   string `json:"type"`
}

func NewServiceF5Lb(ac *AlkiraClient) *AlkiraApi[ServiceF5Lb] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/f5-lb-services", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[ServiceF5Lb]{ac, uri, PaginationOn}
	return api

}

// F5LBServiceConfig is the write model of an F5 load balancer service,
// copied from terraform-provider-alkira-v3 internal/client/service_f5_lb.go:5-84
// and checked against TPS model/f5/F5LBService.java. The portal rebuilds
// the body from its form on create and edit (karrawa f5Config/utils.js:24-108);
// an update here resends the stored body with the patch applied.
type F5LBServiceConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,charset=slug,src=F5LBService.java:63-66"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,charset=name,src=F5LBService.java:68-70"`
	CXP         string      `json:"cxp" ak:"required,immutable,maxlen=128,trimmed,src=F5LBService.java:72-75|F5LBServiceLogicImpl.java:754-756"`
	// Size cannot change while any instance is PAY_AS_YOU_GO (F5LBServiceLogicImpl.java:779-782).
	Size                *string `json:"size,omitempty" ak:"default=SMALL,enum=SMALL|MEDIUM|LARGE|2LARGE|5LARGE,src=F5LBService.java:77-80|Defaults.java:122-134"`
	ServiceGroupName    string  `json:"serviceGroupName" ak:"required,maxlen=255,trimmed,charset=name,src=F5LBService.java:82-86"`
	ILBServiceGroupName *string `json:"ilbServiceGroupName,omitempty" ak:"maxlen=255,trimmed,charset=name,src=F5LBService.java:88-91"`
	// BillingTags: a PUT without it clears the stored tags (V3 service_f5_lb.go:22-24).
	BillingTags      []int64                             `json:"billingTags,omitempty" ak:"maxitems=5,flag=BILLING-TAGS,src=F5LBService.java:96-99"`
	PrefixListID     *int64                              `json:"prefixListId,omitempty"`
	Instances        []F5LBInstanceConfig                `json:"instances" ak:"required,minitems=1,maxitems=2,src=F5LBService.java:104-106"`
	Segments         []string                            `json:"segments" ak:"required,minitems=1,trimmed,src=F5LBService.java:108-112"`
	SegmentOptions   map[string]F5LBSegmentOptionsConfig `json:"segmentOptions" ak:"required,src=F5LBService.java:114-119"`
	GlobalCidrListID *int64                              `json:"globalCidrListId" ak:"required,immutable,src=F5LBService.java:130-131|F5LBServiceLogicImpl.java:758-760"`
	// TunnelProtocol is forced to VXLAN on Azure and defaults to GRE on AWS
	// (F5LBServiceLogicImpl.java:1063-1089).
	TunnelProtocol *string `json:"tunnelProtocol,omitempty" ak:"enum=GRE|VXLAN,src=F5LBService.java:133|F5LBServiceValidator.java:22-31"`

	InternalName        string  `json:"internalName,omitempty" ak:"readonly"`
	ELBServiceGroupName string  `json:"elbServiceGroupName,omitempty" ak:"readonly"`
	ProviderName        string  `json:"providerName,omitempty" ak:"readonly"`
	SegmentIDs          []int64 `json:"segmentIds,omitempty" ak:"readonly"`
	SecondaryCidrUUID   string  `json:"secondaryCidrUuid,omitempty" ak:"readonly"`
	ImplicitGroupID     *int64  `json:"implicitGroupId,omitempty" ak:"readonly"`
	ILBImplicitGroupID  *int64  `json:"ilbImplicitGroupId,omitempty" ak:"readonly"`
	ELBImplicitGroupID  *int64  `json:"elbImplicitGroupId,omitempty" ak:"readonly"`
	// State and PreviousState come from the GET /services row (Descriptor.ServiceType).
	State         string `json:"state,omitempty" ak:"readonly"`
	PreviousState string `json:"previousState,omitempty" ak:"readonly"`
}

// F5LBSegmentOptionsConfig is one segmentOptions entry, keyed by a name of
// segments (F5LBServiceLogicImpl.java:1158-1173). elbBgpOptions and lbType
// need RELEASE-SERVICE-F5LB-BGP and the ILB flag of the CXP's cloud; the
// backend refuses or drops them otherwise (:912-937, :966-969).
type F5LBSegmentOptionsConfig struct {
	ELBNicCount         *int64                `json:"elbNicCount" ak:"required,range=1..6,src=F5LBService.java:156-159"`
	NatPoolPrefixLength *int64                `json:"natPoolPrefixLength,omitempty" ak:"range=0..32,src=F5LBService.java:153-155"`
	LBType              []string              `json:"lbType,omitempty" ak:"enum=ELB|ILB,src=F5LBService.java:162|F5LBServiceValidator.java:44-57"`
	ELBBgpOptions       *F5LBBgpOptionsConfig `json:"elbBgpOptions,omitempty"`
	SegmentID           *int64                `json:"segmentId,omitempty" ak:"readonly"`
}

// F5LBBgpOptionsConfig is segmentOptions.<name>.elbBgpOptions.
type F5LBBgpOptionsConfig struct {
	AdvertiseToCXPPrefixListID *int64   `json:"advertiseToCXPPrefixListId" ak:"required,src=F5LBService.java:182|F5LBServiceValidator.java:59-96"`
	SegmentID                  *int64   `json:"segmentId,omitempty" ak:"readonly"`
	AdvertisedPrefixes         []string `json:"advertisedPrefixes,omitempty" ak:"readonly"`
}

// F5LBInstanceConfig is one F5 instance (TPS F5LBServiceInstance.java). An
// update matches instances by id: one without id is created, and a stored
// one must be kept or removed by id. The instance credential is used by
// this instance only and is deleted with it.
type F5LBInstanceConfig struct {
	ID                       *int64                `json:"id,omitempty" ak:"readonly"`
	Name                     string                `json:"name" ak:"required,maxlen=256,charset=slug,src=F5LBServiceInstance.java:42-45"`
	CredentialID             string                `json:"credentialId" ak:"required,immutable,src=F5LBServiceInstance.java:49-50|F5LBServiceLogicImpl.java:808-815"`
	RegistrationCredentialID *string               `json:"registrationCredentialId,omitempty" ak:"immutable,src=F5LBServiceInstance.java:47|F5LBServiceValidator.java:116-131"`
	LicenseType              string                `json:"licenseType" ak:"required,immutable,enum=BRING_YOUR_OWN|PAY_AS_YOU_GO,src=F5LBServiceInstance.java:52-53|F5LBServiceValidator.java:32-35"`
	Version                  string                `json:"version" ak:"required,src=F5LBServiceInstance.java:55-57"`
	Deployment               *F5LBDeploymentConfig `json:"deployment" ak:"required,src=F5LBServiceInstance.java:59-61"`
	HostNameFqdn             string                `json:"hostNameFqdn" ak:"required,maxlen=255,src=F5LBServiceInstance.java:63-66"`
	AvailabilityZone         *int64                `json:"availabilityZone,omitempty" ak:"range=0..1,src=F5LBServiceInstance.java:82-84"`

	InternalName      string          `json:"internalName,omitempty" ak:"readonly"`
	ProductID         string          `json:"productId,omitempty" ak:"readonly"`
	SegmentSubnetMap  json.RawMessage `json:"segmentSubnetMap,omitempty" ak:"readonly"`
	SegmentIlbInfoMap json.RawMessage `json:"segmentIlbInfoMap,omitempty" ak:"readonly"`
	RouteDomainIDs    json.RawMessage `json:"routeDomainIds,omitempty" ak:"readonly"`
	Metadata          json.RawMessage `json:"metadata,omitempty" ak:"readonly"`
}

// F5LBDeploymentConfig is an instance's deployment. BRING_YOUR_OWN takes
// LTM_DNS and PAY_AS_YOU_GO one of GOOD, BETTER, BEST; the backend sets
// option itself (F5LBServiceValidator.java:97-106, :121-137; F5LBServiceLogicImpl.java:880-888).
type F5LBDeploymentConfig struct {
	Type   string  `json:"type" ak:"required,immutable,enum=LTM_DNS|GOOD|BETTER|BEST|ALL,src=F5LBServiceInstance.java:123-143|F5LBServiceLogicImpl.java:824-831"`
	Option *string `json:"option,omitempty" ak:"readonly"`
}

// f5LBServiceDescriptor: services share one id space and a GET of another
// service's id is not refused (ServiceLogicImpl.java:1354-1365; V3
// service_f5_lb.go:82-84), so providerName, which only F5 bodies carry
// (F5LBServiceLogicImpl.java:365), identifies the family. The type, state
// and previousState come from GET /services/{id} (ServiceType F5LB,
// Defaults.java:458), so an update runs the D13 state pre-check, as TPS
// does (ProvisionedEntityValidatorImpl.java:34-73). Names are unique across
// every service type, so the duplicate guard reads GET /services; the
// backend answers 409 with no errors[] (ServiceLogicImpl.java:2174-2184).
// A DELETE is refused with 400 constraint.invalid while vServer endpoints
// remain or a policy uses the service group (F5LBServiceLogicImpl.java:538-541,
// :702-710); the D13 delete refusal carries the same code and top-level
// message (ProvisionedEntityValidatorImpl.java:55-61), so only the 409 of an
// external reference is classified as in use.
var f5LBServiceDescriptor = Descriptor{
	Collection: "f5-lb-services",
	FamilyType: FamilyCheck{Field: "providerName"},
	InUse: []Signature{
		{Status: 409, Code: "AK000001", Message: "is still referred externally"},
	},
	Duplicate:          []Signature{{Status: 409, Code: "AK000001", Message: "already exists"}},
	Provisionable:      true,
	MarkedRowLifecycle: true,
	DocState:           DocStatePlainGet,
	ServiceType:        "F5LB",
	Uniqueness:         Uniqueness{Key: "name", Scope: ScopeServices},
	IDMatchedArrays:    []string{"instances"},
	StateGuard:         StateGuardD13,
	Service:            ServiceTPS,
	Flags: []FlagRule{
		{Flag: "BILLING-TAGS", Effect: FlagIncludeField, Field: "billingTags", UISource: "src/views/network/provision/objectConfig/f5Config/addOrEdit/index.js:306-308|f5Config/utils.js:73"},
	},
}

// F5LBServiceDescriptor returns the F5 load balancer service write descriptor.
func F5LBServiceDescriptor() Descriptor { return f5LBServiceDescriptor.clone() }

// CreateF5LBService creates an F5 load balancer service.
func (ac *AlkiraClient) CreateF5LBService(ctx context.Context, s *F5LBServiceConfig) (*F5LBServiceConfig, *WriteResult, error) {
	core, err := NewCore[F5LBServiceConfig](ac, f5LBServiceDescriptor)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, s, WriteOptions{})
}

// GetF5LBService reads one F5 load balancer service and its raw body.
func (ac *AlkiraClient) GetF5LBService(ctx context.Context, id string) (*F5LBServiceConfig, []byte, error) {
	core, err := NewCore[F5LBServiceConfig](ac, f5LBServiceDescriptor)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

// UpdateF5LBService applies a merge patch to an F5 load balancer service.
func (ac *AlkiraClient) UpdateF5LBService(ctx context.Context, id string, patch json.RawMessage) (*F5LBServiceConfig, *WriteResult, error) {
	core, err := NewCore[F5LBServiceConfig](ac, f5LBServiceDescriptor)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

// DeleteF5LBService deletes an F5 load balancer service; a deployed one is
// marked for deletion until the next provision.
func (ac *AlkiraClient) DeleteF5LBService(ctx context.Context, id string) (*WriteResult, error) {
	core, err := NewCore[F5LBServiceConfig](ac, f5LBServiceDescriptor)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

// ListF5LBServicePaged reads every F5 load balancer service through paged reads.
func (ac *AlkiraClient) ListF5LBServicePaged(ctx context.Context) ([]F5LBServiceConfig, error) {
	core, err := NewCore[F5LBServiceConfig](ac, f5LBServiceDescriptor)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
