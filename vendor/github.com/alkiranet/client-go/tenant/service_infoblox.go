// Copyright (C) 2022-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
)

type ServiceInfoblox struct {
	AnyCast                     InfobloxAnycast    `json:"anycast"`
	BillingTags                 []int              `json:"billingTags"`
	Cxp                         string             `json:"cxp"`
	Description                 string             `json:"description,omitempty"`
	GlobalCidrListId            int                `json:"globalCidrListId"`
	GridMaster                  InfobloxGridMaster `json:"gridMaster"`
	Id                          json.Number        `json:"id,omitempty"` // RESPONSE ONLY
	Instances                   []InfobloxInstance `json:"instances"`
	InternalName                string             `json:"internalName,omitempty"`
	LicenseType                 string             `json:"licenseType,omitempty"`
	Name                        string             `json:"name"`
	Segments                    []string           `json:"segments"`
	ServiceGroupId              int                `json:"serviceGroupId,omitempty"`              // RESPONSE ONLY
	ServiceGroupImplicitGroupId int                `json:"serviceGroupImplicitGroupId,omitempty"` // RESPONSE ONLY
	ServiceGroupName            string             `json:"serviceGroupName"`
	Size                        string             `json:"size,omitempty"`
	AllowListId                 int                `json:"allowListId,omitempty"`
}

type InfobloxAnycast struct {
	BackupCxps []string `json:"backupCxps,omitempty"`
	Enabled    bool     `json:"enabled"`
	Ips        []string `json:"ips,omitempty"`
}

type InfobloxGridMaster struct {
	External                 bool   `json:"external,omitempty"`
	GridMasterCredentialId   string `json:"gridMasterCredentialId"`
	Ip                       string `json:"ip,omitempty"`
	Name                     string `json:"name"`
	SharedSecretCredentialId string `json:"sharedSecretCredentialId"`
}

type InfobloxInstance struct {
	AnyCastEnabled     bool        `json:"anyCastEnabled"`
	ConfiguredMasterIp string      `json:"configuredMasterIp,omitempty"`
	CredentialId       string      `json:"credentialId"`
	HostName           string      `json:"hostName"`
	Id                 json.Number `json:"id,omitempty"`
	InternalName       string      `json:"internalName,omitempty"`
	LanPrefix          string      `json:"lanPrefix,omitempty"`
	ManagementPrefix   string      `json:"managementPrefix,omitempty"`
	Model              string      `json:"model"`
	Name               string      `json:"name,omitempty"`
	ProductId          string      `json:"productId,omitempty"`
	PublicIp           string      `json:"publicIp,omitempty"`
	Type               string      `json:"type"`
	Version            string      `json:"version,omitempty"`
}

type ServiceInfobloxSummary struct {
	Cxp          string      `json:"cxp"`
	Id           json.Number `json:"id,omitempty"`
	InternalName string      `json:"internalName,omitempty"`
	LicenseType  string      `json:"licenseType,omitempty"`
	Name         string      `json:"name"`
}

// NewServiceInfoblox new service infoblox
func NewServiceInfoblox(ac *AlkiraClient) *AlkiraApi[ServiceInfobloxSummary] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/infoblox-services", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[ServiceInfobloxSummary]{ac, uri, PaginationOn}
	return api
}

// ServiceInfobloxConfig is the write model of an Infoblox service (TPS
// model/infoblox/InfobloxService.java). The portal rebuilds the body from
// its form on create and edit (karrawa InfoBloxComponent.js:660-705);
// an update here resends the stored body with the patch applied.
type ServiceInfobloxConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=64,charset=hyphen,src=InfobloxService.java:70-72|constants/index.js:129"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,charset=description,src=InfobloxService.java:74-76"`
	// Size applies only to a service with a NIOS_X instance; the backend
	// refuses any other value than SMALL without one (InfobloxServiceLogicImpl.java:397-403).
	Size             *string                     `json:"size,omitempty" ak:"default=SMALL,enum=SMALL|MEDIUM|LARGE|2LARGE,src=InfobloxService.java:81-87|InfoBloxComponent.js:686"`
	CXP              string                      `json:"cxp" ak:"required,immutable,maxlen=128,trimmed,src=InfobloxService.java:89-92|ServiceLogicImpl.java:891"`
	Segments         []string                    `json:"segments" ak:"required,minitems=1,maxitems=1,trimmed,src=InfobloxService.java:94-98"`
	BillingTags      []int64                     `json:"billingTags,omitempty" ak:"maxitems=5,flag=BILLING-TAGS,src=InfobloxService.java:100-102"`
	GlobalCidrListID *int64                      `json:"globalCidrListId" ak:"required,src=InfobloxService.java:104-105"`
	Anycast          *ServiceInfobloxAnycast     `json:"anycast,omitempty"`
	ServiceGroupName string                      `json:"serviceGroupName" ak:"required,trimmed,src=InfobloxService.java:116-118"`
	GridMaster       *ServiceInfobloxGridMaster  `json:"gridMaster,omitempty"`
	AllowListID      *int64                      `json:"allowListId,omitempty"`
	Instances        []ServiceInfobloxInstanceIn `json:"instances" ak:"required,minitems=1,maxitems=8,src=InfobloxService.java:134-138"`

	InternalName                string `json:"internalName,omitempty" ak:"readonly"`
	State                       string `json:"state,omitempty" ak:"readonly"`
	PreviousState               string `json:"previousState,omitempty" ak:"readonly"`
	ServiceGroupID              *int64 `json:"serviceGroupId,omitempty" ak:"readonly"`
	ServiceGroupImplicitGroupID *int64 `json:"serviceGroupImplicitGroupId,omitempty" ak:"readonly"`
	LicenseType                 string `json:"licenseType,omitempty" ak:"readonly"`
}

// ServiceInfobloxAnycast is InfobloxService.Anycast. The portal always
// sends it; with enabled false the backend clears ips and backupCxps
// (InfobloxServiceLogicImpl.java:257-261).
type ServiceInfobloxAnycast struct {
	Enabled       *bool    `json:"enabled,omitempty" ak:"default=false,src=InfoBloxComponent.js:696-703"`
	Ips           []string `json:"ips,omitempty" ak:"maxitems=1,src=InfobloxService.java:249-251"`
	BackupCxps    []string `json:"backupCxps,omitempty" ak:"maxitems=1,src=InfobloxService.java:254-256"`
	RoutePolicyID *int64   `json:"routePolicyId,omitempty" ak:"readonly"`
	PrefixListID  *int64   `json:"prefixListId,omitempty" ak:"readonly"`
}

// ServiceInfobloxGridMaster is InfobloxService.GridMaster: required with
// name and both credential ids for a service with NIOS instances, refused
// for a NIOS_X-only one (InfobloxServiceLogicImpl.java:476-496).
type ServiceInfobloxGridMaster struct {
	IP                       *string `json:"ip,omitempty"`
	Name                     *string `json:"name,omitempty" ak:"maxlen=256,charset=name,src=InfobloxService.java:227-229"`
	GridMasterCredentialID   *string `json:"gridMasterCredentialId,omitempty"`
	SharedSecretCredentialID *string `json:"sharedSecretCredentialId,omitempty"`
	External                 *bool   `json:"external,omitempty" ak:"readonly"`
}

// ServiceInfobloxInstanceIn is one Infoblox instance (TPS
// InfobloxServiceInstance.java). An update matches instances by id; one
// sent without id is created. platform absent means NIOS; NIOS_X takes a
// join token credential and no model or type (InfoBloxServiceInstanceValidator.java:51-67).
type ServiceInfobloxInstanceIn struct {
	ID             *int64  `json:"id,omitempty" ak:"readonly"`
	HostName       string  `json:"hostName" ak:"required,maxlen=256,trimmed,src=InfobloxServiceInstance.java:44-47"`
	CredentialID   string  `json:"credentialId" ak:"required,src=InfobloxServiceInstance.java:49-50"`
	Model          *string `json:"model,omitempty" ak:"trimmed,src=InfobloxServiceInstance.java:54-55"`
	Type           *string `json:"type,omitempty" ak:"enum=MASTER|MASTER_CANDIDATE|MEMBER,src=InfobloxServiceInstance.java:58|:107-111"`
	Platform       *string `json:"platform,omitempty" ak:"immutable,enum=NIOS|NIOS_X,src=InfobloxServiceInstance.java:62|:114-117"`
	AnyCastEnabled *bool   `json:"anyCastEnabled,omitempty"`
	Version        string  `json:"version" ak:"required,src=InfoBloxServiceInstanceValidator.java:44-47"`

	Name               string `json:"name,omitempty" ak:"readonly"`
	InternalName       string `json:"internalName,omitempty" ak:"readonly"`
	State              string `json:"state,omitempty" ak:"readonly"`
	ManagementPrefix   string `json:"managementPrefix,omitempty" ak:"readonly"`
	LanPrefix          string `json:"lanPrefix,omitempty" ak:"readonly"`
	CxpBgpIP           string `json:"cxpBgpIp,omitempty" ak:"readonly"`
	CxpBgpAsn          string `json:"cxpBgpAsn,omitempty" ak:"readonly"`
	InstanceBgpIP      string `json:"instanceBgpIp,omitempty" ak:"readonly"`
	InstanceBgpAsn     string `json:"instanceBgpAsn,omitempty" ak:"readonly"`
	ConfiguredMasterIP string `json:"configuredMasterIp,omitempty" ak:"readonly"`
	PublicIP           string `json:"publicIp,omitempty" ak:"readonly"`
	ProductID          string `json:"productId,omitempty" ak:"readonly"`
}

// serviceInfobloxDescriptor: services share one id space and a GET of
// another service's id answers 200 here (ServiceLogicImpl.java:1354-1365),
// so anycast, which only Infoblox bodies carry, identifies the family, and
// GET /services/{id} (ServiceType INFOBLOX, Defaults.java:453) refuses
// another type's id before the family GET and supplies state and
// previousState. A
// missing id answers 404 on GET, PUT and DELETE (:1358-1364, :864-865,
// :1101-1102). Service names are unique across every service type of the
// network, so the duplicate guard reads GET /services; the backend answers
// 409 (ServiceLogicImpl.java:2174-2184). A DELETE is refused with
// 400 constraint.invalid while the service group is used by a policy or
// zone, or the service is another service's anycast backup
// (InfobloxServiceLogicImpl.java:787-814); the D13 delete refusal carries the
// same code and top-level message (ProvisionedEntityValidatorImpl.java:55-61),
// so only the 409 of an external reference is classified as in use.
var serviceInfobloxDescriptor = Descriptor{
	Collection: "infoblox-services",
	FamilyType: FamilyCheck{Field: "anycast"},
	InUse: []Signature{
		{Status: 409, Code: "AK000001", Message: "is still referred externally"},
	},
	Duplicate:          []Signature{{Status: 409, Code: "AK000001", Message: "already exists"}},
	Provisionable:      true,
	MarkedRowLifecycle: true,
	DocState:           DocStatePlainGet,
	ServiceType:        "INFOBLOX",
	Uniqueness:         Uniqueness{Key: "name", Scope: ScopeServices},
	IDMatchedArrays:    []string{"instances"},
	StateGuard:         StateGuardD13,
	Service:            ServiceTPS,
	Flags: []FlagRule{
		{Flag: "BILLING-TAGS", Effect: FlagIncludeField, Field: "billingTags", UISource: "src/views/network/provision/objectConfig/inflobloxConfig/InfoBloxComponent.js:211|:677"},
	},
}

// ServiceInfobloxDescriptor returns the Infoblox service write descriptor.
func ServiceInfobloxDescriptor() Descriptor { return serviceInfobloxDescriptor.clone() }

// CreateServiceInfoblox creates an Infoblox service.
func (ac *AlkiraClient) CreateServiceInfoblox(ctx context.Context, s *ServiceInfobloxConfig) (*ServiceInfobloxConfig, *WriteResult, error) {
	core, err := NewCore[ServiceInfobloxConfig](ac, serviceInfobloxDescriptor)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, s, WriteOptions{})
}

// GetServiceInfoblox reads one Infoblox service and its raw body.
func (ac *AlkiraClient) GetServiceInfoblox(ctx context.Context, id string) (*ServiceInfobloxConfig, []byte, error) {
	core, err := NewCore[ServiceInfobloxConfig](ac, serviceInfobloxDescriptor)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

// UpdateServiceInfoblox applies a merge patch to an Infoblox service.
func (ac *AlkiraClient) UpdateServiceInfoblox(ctx context.Context, id string, patch json.RawMessage) (*ServiceInfobloxConfig, *WriteResult, error) {
	core, err := NewCore[ServiceInfobloxConfig](ac, serviceInfobloxDescriptor)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

// DeleteServiceInfoblox deletes an Infoblox service; a deployed one is
// marked for deletion until the next provision.
func (ac *AlkiraClient) DeleteServiceInfoblox(ctx context.Context, id string) (*WriteResult, error) {
	core, err := NewCore[ServiceInfobloxConfig](ac, serviceInfobloxDescriptor)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

// ListServiceInfobloxPaged reads every Infoblox service through paged reads.
func (ac *AlkiraClient) ListServiceInfobloxPaged(ctx context.Context) ([]ServiceInfobloxConfig, error) {
	core, err := NewCore[ServiceInfobloxConfig](ac, serviceInfobloxDescriptor)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
