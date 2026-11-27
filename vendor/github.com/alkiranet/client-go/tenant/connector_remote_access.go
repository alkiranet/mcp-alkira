// Copyright (C) 2022-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
)

type ConnectorRemoteAccessTemplate struct {
	AdvancedOptions       ConnectorRemoteAccessAdvancedOptions  `json:"advancedOptions"`
	Arguments             []ConnectorRemoteAccessArguments      `json:"arguments"`
	AuthenticationOptions ConnectorRemoteAccessAuthOptions      `json:"authenticationOptions"`
	Id                    json.Number                           `json:"id,omitempty"`
	Name                  string                                `json:"name"`
	SegmentOptions        []ConnectorRemoteAccessSegmentOptions `json:"segmentOptions"`
	Segments              []string                              `json:"segments,omitempty"`
	BannerText            string                                `json:"bannerText,omitempty"`
}

type ConnectorRemoteAccessAdvancedOptions struct {
	EnableDynamicRegionMapping bool   `json:"enableDynamicRegionMapping"`
	MaxActiveUsersThreshold    int    `json:"maxActiveUsersThreshold"`
	NameServer                 string `json:"nameServer"`
	FallbackToTcp              bool   `json:"fallbackToTcp"`
}

type ConnectorRemoteAccessArguments struct {
	BillingTags []int  `json:"billingTags,omitempty"`
	Cxp         string `json:"cxp"`
	Size        string `json:"size"`
}

type ConnectorRemoteAccessAuthOptions struct {
	LdapSettings   *ConnectorRemoteAccessLdapSettings `json:"ldapSettings,omitempty"`
	SupportedModes []string                           `json:"supportedModes"`
}

type ConnectorRemoteAccessLdapSettings struct {
	BindUserDomain      string `json:"bindUserDomain,omitempty"`
	DestinationAddress  string `json:"destinationAddress,omitempty"`
	LdapType            string `json:"ldapType,omitempty"`
	ManagementSegmentId int    `json:"managementSegmentId,omitempty"`
	SearchScopeDomain   string `json:"searchScopeDomain,omitempty"`
}

type ConnectorRemoteAccessSegmentOptions struct {
	SegmentId         int                                      `json:"segmentId"`
	UserGroupMappings []ConnectorRemoteAccessUserGroupMappings `json:"userGroupMappings"`
}

type ConnectorRemoteAccessUserGroupMappings struct {
	BillingTag          int                                       `json:"billingTag,omitempty"`
	Name                string                                    `json:"name"`
	PrefixListId        int                                       `json:"prefixListId"`
	SplitTunneling      bool                                      `json:"splitTunneling"`
	CxpToSubnetsMapping []ConnectorRemoteAccessCxpToSubnetMapping `json:"cxpToSubnetsMapping"`
}

type ConnectorRemoteAccessCxpToSubnetMapping struct {
	Cxp     string   `json:"cxp"`
	Subnets []string `json:"subnets"`
}

// NewConnectorRemoteAccessTemplate
func NewConnectorRemoteAccessTemplate(ac *AlkiraClient) *AlkiraApi[ConnectorRemoteAccessTemplate] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/alkira-remote-access-connector-templates", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[ConnectorRemoteAccessTemplate]{ac, uri, PaginationOn}
	return api
}

// ConnectorRemoteAccessTemplateConfig is the write model of a remote access
// connector template (TPS model/RemoteAccessConnectorTemplate.java). The
// backend creates one REMOTE_ACCESS connector per arguments[].cxp and owns
// the certificates and keys; the template body carries only their ids.
type ConnectorRemoteAccessTemplateConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,trimmed,src=RemoteAccessConnectorTemplate.java:63-67"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,charset=description,src=RemoteAccessConnectorTemplate.java:69-71"`
	// Segments holds segment names; each needs an entry in segmentOptions
	// (RemoteAccessConnectorTemplateServiceImpl.java:900-909).
	Segments              []string                         `json:"segments" ak:"required,minitems=1,trimmed,src=RemoteAccessConnectorTemplate.java:76"`
	SegmentIDs            []int64                          `json:"segmentIds,omitempty"`
	SegmentOptions        []ConnectorRemoteAccessSegmentIn `json:"segmentOptions" ak:"required,minitems=1,src=RemoteAccessConnectorTemplateServiceImpl.java:686"`
	AuthenticationOptions *ConnectorRemoteAccessAuthIn     `json:"authenticationOptions" ak:"required,src=RemoteAccessConnectorTemplateServiceImpl.java:694"`
	AdvancedOptions       *ConnectorRemoteAccessAdvancedIn `json:"advancedOptions" ak:"required,src=RemoteAccessConnectorTemplateServiceImpl.java:690"`
	Arguments             []ConnectorRemoteAccessArgIn     `json:"arguments" ak:"required,minitems=1,distinct=cxp,src=RemoteAccessConnectorTemplateServiceImpl.java:698|ConnectorTemplateValidatorImpl.java:245-249"`
	// SamlIDPMetadata is required while SAML is a supported mode (:883-897).
	SamlIDPMetadata *string `json:"samlIDPMetadata,omitempty"`
	// DeviceCertificate is required for always-on or pre-logon; the backend
	// drops it while ZTNA-PHASE-2 is off (:776-793, :858-862).
	DeviceCertificate *string `json:"deviceCertificate,omitempty" ak:"flag=ZTNA-PHASE-2"`
	BannerText        *string `json:"bannerText,omitempty" ak:"maxlen=2048,src=RemoteAccessConnectorTemplate.java:122-124"`

	InternalName            string `json:"internalName,omitempty" ak:"readonly"`
	Version                 *int64 `json:"version,omitempty" ak:"readonly"`
	OpenVPNStaticKeysID     *int64 `json:"openVPNStaticKeysId,omitempty" ak:"readonly"`
	ClientCertificateID     *int64 `json:"clientCertificateId,omitempty" ak:"readonly"`
	ClientRootCertificateID *int64 `json:"clientRootCertificateId,omitempty" ak:"readonly"`
	SamlCertificateID       *int64 `json:"samlCertificateId,omitempty" ak:"readonly"`
	CertificatesGeneratedAt *int64 `json:"certificatesGeneratedAt,omitempty" ak:"readonly"`
	CertificatesExpireAt    *int64 `json:"certificatesExpireAt,omitempty" ak:"readonly"`
	State                   string `json:"state,omitempty" ak:"readonly"`
	DocState                string `json:"docState,omitempty" ak:"readonly"`
	// ReferencedPrefixListIDs is in the GET body only; a PUT that carries it
	// is refused with 400 "Failed to parse request" (dev, tenant 239).
	ReferencedPrefixListIDs []int64 `json:"referencedPrefixListIds,omitempty" ak:"readonly"`
}

// ConnectorRemoteAccessSegmentIn is one segmentOptions entry: segmentId or
// the segment name, which the backend resolves into each other (:2426-2434).
type ConnectorRemoteAccessSegmentIn struct {
	SegmentID         *int64                           `json:"segmentId,omitempty"`
	Name              *string                          `json:"name,omitempty"`
	UserGroupMappings []ConnectorRemoteAccessMappingIn `json:"userGroupMappings" ak:"required,minitems=1,src=RemoteAccessConnectorTemplateServiceImpl.java:900-909"`
}

// ConnectorRemoteAccessMappingIn maps an existing user group to subnets per CXP.
type ConnectorRemoteAccessMappingIn struct {
	Name                string                          `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=RemoteAccessConnectorTemplate.java:236-240"`
	SplitTunneling      *bool                           `json:"splitTunneling" ak:"required,src=RemoteAccessConnectorTemplate.java:243-244"`
	BillingTag          *int64                          `json:"billingTag,omitempty"`
	PrefixListID        *int64                          `json:"prefixListId,omitempty"`
	CxpToSubnetsMapping []ConnectorRemoteAccessSubnetIn `json:"cxpToSubnetsMapping" ak:"required,minitems=1,src=RemoteAccessConnectorTemplate.java:249"`
	// UserGroupType is set by the backend from the always-on settings (:757-773).
	UserGroupType string `json:"userGroupType,omitempty" ak:"readonly"`
}

// ConnectorRemoteAccessSubnetIn is one cxpToSubnetsMapping entry.
type ConnectorRemoteAccessSubnetIn struct {
	CXP     string   `json:"cxp" ak:"required,maxlen=256,trimmed,charset=name,src=RemoteAccessConnectorTemplate.java:267-271"`
	Subnets []string `json:"subnets" ak:"required,minitems=1,trimmed,src=RemoteAccessConnectorTemplate.java:273-276"`
}

// ConnectorRemoteAccessAuthIn is RemoteAccessAuthenticationOptions; LDAP
// needs ldapSettings with a credential referenced by id.
type ConnectorRemoteAccessAuthIn struct {
	SupportedModes []string                     `json:"supportedModes" ak:"required,minitems=1,enum=LOCAL|LDAP|SAML,src=RemoteAccessAuthenticationOptions.java:26-38"`
	LdapSettings   *ConnectorRemoteAccessLdapIn `json:"ldapSettings,omitempty"`
}

// ConnectorRemoteAccessLdapIn is the LDAP settings of a template
// (ConnectorTemplateValidatorImpl.java:260-329).
type ConnectorRemoteAccessLdapIn struct {
	LdapType            *string `json:"ldapType,omitempty" ak:"enum=OPEN_LDAP|ACTIVE_DIRECTORY,src=RemoteAccessAuthenticationOptions.java:48-49"`
	ManagementSegmentID *int64  `json:"managementSegmentId,omitempty"`
	ManagementSegment   *string `json:"managementSegment,omitempty"`
	DestinationAddress  *string `json:"destinationAddress,omitempty"`
	BindUserDomain      *string `json:"bindUserDomain,omitempty"`
	SearchScopeDomain   *string `json:"searchScopeDomain,omitempty"`
	CredentialID        *string `json:"credentialId,omitempty"`
	CredentialUpdatedAt *int64  `json:"credentialUpdatedAt,omitempty" ak:"readonly"`
}

// ConnectorRemoteAccessAdvancedIn is RemoteAccessConnectorAdvancedOptions.
// The backend drops cxpToNameServerMapping, fallbackToTcp, alwaysOnEnabled,
// preLogonEnabled and zeroTrustAccessPosture while their flags are off
// (RemoteAccessConnectorTemplateServiceImpl.java:777-856).
type ConnectorRemoteAccessAdvancedIn struct {
	EnableDynamicRegionMapping *bool           `json:"enableDynamicRegionMapping,omitempty" ak:"default=true,src=remoteAccessConfig/add/index.js:78"`
	MaxActiveUsersThreshold    *int64          `json:"maxActiveUsersThreshold,omitempty" ak:"default=80,src=constants/index.js:1467"`
	NameServer                 *string         `json:"nameServer,omitempty"`
	CxpToNameServerMapping     json.RawMessage `json:"cxpToNameServerMapping,omitempty"`
	FallbackToTcp              *bool           `json:"fallbackToTcp,omitempty"`
	AlwaysOnEnabled            *bool           `json:"alwaysOnEnabled,omitempty"`
	PreLogonEnabled            *bool           `json:"preLogonEnabled,omitempty"`
	ZeroTrustAccessPosture     json.RawMessage `json:"zeroTrustAccessPosture,omitempty"`
	UDPPort                    *int64          `json:"udpPort,omitempty" ak:"readonly"`
	TCPPort                    *int64          `json:"tcpPort,omitempty" ak:"readonly"`
}

// ConnectorRemoteAccessArgIn is one arguments entry: a CXP the template
// deploys a connector in. The backend default size is SMALL.
type ConnectorRemoteAccessArgIn struct {
	CXP         string  `json:"cxp" ak:"required,trimmed,src=RemoteAccessConnectorTemplate.java:288-292"`
	Size        *string `json:"size,omitempty" ak:"enum=SMALL|MEDIUM|LARGE|2LARGE|4LARGE|5LARGE,src=RemoteAccessConnectorTemplate.java:294-296|ConnectorTemplateValidatorImpl.java:217-221"`
	BillingTags []int64 `json:"billingTags,omitempty" ak:"maxitems=5,src=RemoteAccessConnectorTemplate.java:298-300"`
}

// connectorRemoteAccessTemplateDescriptor: GET and PUT answer a missing id
// with 404 "Could not find connector template" and DELETE with 404 "Could
// not find network entity" (ConnectorTemplateServiceImpl.java:124-125;
// AbstractNetworkCRUDService.java:276-291). Template ids have their own
// sequence (ConnectorTemplateEntity.java:57); arguments identifies the
// family. A name equal to another template's, ignoring case, is refused
// with 400 connectortemplate.00000030 (RemoteAccessConnectorTemplateServiceImpl.java:566-578).
// The template is not a provisioned entity, so TPS runs no state check on
// its update; a deployed template is marked for deletion on delete, and a
// delete that would drop the last connector of a user group used by a
// policy or firewall zone is refused with 400 (:2922-2931, :3003-3008).
var connectorRemoteAccessTemplateDescriptor = Descriptor{
	Collection: "alkira-remote-access-connector-templates",
	FamilyType: FamilyCheck{Field: "arguments"},
	InUse: []Signature{
		{Status: 400, Code: "AK000002", Message: "Cannot remove all remote access connectors"},
	},
	Duplicate:          []Signature{{Status: 400, FieldCode: "connectortemplate.00000030"}},
	Provisionable:      true,
	MarkedRowLifecycle: true,
	DocState:           DocStateField,
	Uniqueness:         Uniqueness{Key: "name", Scope: ScopeCollection},
	StateGuard:         StateGuardNone,
	Service:            ServiceTPS,
	Flags: []FlagRule{
		{Flag: "ZTNA-PHASE-2", Effect: FlagIncludeField, Field: "deviceCertificate", UISource: "src/views/network/provision/objectConfig/remoteAccessConfig/add/index.js:108|:783-787"},
	},
}

// ConnectorRemoteAccessTemplateDescriptor returns the remote access connector template write descriptor.
func ConnectorRemoteAccessTemplateDescriptor() Descriptor {
	return connectorRemoteAccessTemplateDescriptor.clone()
}

// CreateConnectorRemoteAccessTemplate creates a remote access connector template.
func (ac *AlkiraClient) CreateConnectorRemoteAccessTemplate(ctx context.Context, c *ConnectorRemoteAccessTemplateConfig) (*ConnectorRemoteAccessTemplateConfig, *WriteResult, error) {
	core, err := NewCore[ConnectorRemoteAccessTemplateConfig](ac, connectorRemoteAccessTemplateDescriptor)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, c, WriteOptions{})
}

// GetConnectorRemoteAccessTemplate reads one remote access connector template and its raw body.
func (ac *AlkiraClient) GetConnectorRemoteAccessTemplate(ctx context.Context, id string) (*ConnectorRemoteAccessTemplateConfig, []byte, error) {
	core, err := NewCore[ConnectorRemoteAccessTemplateConfig](ac, connectorRemoteAccessTemplateDescriptor)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

// UpdateConnectorRemoteAccessTemplate applies a merge patch to a remote access connector template.
func (ac *AlkiraClient) UpdateConnectorRemoteAccessTemplate(ctx context.Context, id string, patch json.RawMessage) (*ConnectorRemoteAccessTemplateConfig, *WriteResult, error) {
	core, err := NewCore[ConnectorRemoteAccessTemplateConfig](ac, connectorRemoteAccessTemplateDescriptor)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

// DeleteConnectorRemoteAccessTemplate deletes a remote access connector
// template and its connectors; a deployed one is marked for deletion until
// the next provision.
func (ac *AlkiraClient) DeleteConnectorRemoteAccessTemplate(ctx context.Context, id string) (*WriteResult, error) {
	core, err := NewCore[ConnectorRemoteAccessTemplateConfig](ac, connectorRemoteAccessTemplateDescriptor)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

// ListConnectorRemoteAccessTemplatePaged reads every remote access connector template through paged reads.
func (ac *AlkiraClient) ListConnectorRemoteAccessTemplatePaged(ctx context.Context) ([]ConnectorRemoteAccessTemplateConfig, error) {
	core, err := NewCore[ConnectorRemoteAccessTemplateConfig](ac, connectorRemoteAccessTemplateDescriptor)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
