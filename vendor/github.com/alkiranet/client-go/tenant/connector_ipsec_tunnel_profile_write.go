// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// IPsecTunnelProfileConfig is the write model of an IPsec tunnel profile
// (TPS model/ipsec/IPSecTunnelProfile.java): the IKE and ESP algorithms an
// advanced IPsec connector tunnel refers to by profileId. The algorithm
// names are the backend enum constants, "_" removed where the backend
// removes it (EnumValidator.java:37-43).
type IPsecTunnelProfileConfig struct {
	ID                 json.Number                  `json:"id,omitempty" ak:"readonly"`
	Name               string                       `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=IPSecTunnelProfile.java:45-49"`
	Description        *string                      `json:"description,omitempty" ak:"maxlen=1024,src=IPSecTunnelProfile.java:51-52"`
	IPSecConfiguration *IPsecTunnelProfileESPConfig `json:"ipSecConfiguration,omitempty" ak:"required,src=AddIkeProfile/utils.js:38-42"`
	IKEConfiguration   *IPsecTunnelProfileIKEConfig `json:"ikeConfiguration,omitempty" ak:"required,src=AddIkeProfile/utils.js:43-47"`

	TenantID        *int64 `json:"tenantId,omitempty" ak:"readonly"`
	TenantNetworkID *int64 `json:"tenantNetworkId,omitempty" ak:"readonly"`
	UpdatedAt       *int64 `json:"updatedAt,omitempty" ak:"readonly"`
}

// IPsecTunnelProfileESPConfig is ipSecConfiguration, the ESP algorithms. A
// GCM encryption algorithm carries its own integrity, so integrityAlgorithm
// is optional.
type IPsecTunnelProfileESPConfig struct {
	EncryptionAlgorithm string  `json:"encryptionAlgorithm" ak:"required,enum=AES256CBC|AES192CBC|AES128CBC|3DESCBC|NULL|AES128GCM16|AES192GCM16|AES256GCM16,src=IPSecTunnelProfile.java:89-91|IPsecConnectorInstanceAdvancedAttributes.java:245-253"`
	IntegrityAlgorithm  *string `json:"integrityAlgorithm,omitempty" ak:"enum=SHA1|SHA256|SHA384|SHA512|MD5,src=IPSecTunnelProfile.java:93-94|IPsecConnectorInstanceAdvancedAttributes.java:299-304"`
	DHGroup             string  `json:"dhGroup" ak:"required,enum=MODP1024|MODP1536|MODP2048|MODP3072|MODP4096|MODP6144|MODP8192|ECP256|ECP384|ECP521|CURVE25519|NONE,src=IPSecTunnelProfile.java:96-98|IPsecConnectorInstanceAdvancedAttributes.java:344-356"`
}

// IPsecTunnelProfileIKEConfig is ikeConfiguration, the IKE algorithms.
type IPsecTunnelProfileIKEConfig struct {
	EncryptionAlgorithm string `json:"encryptionAlgorithm" ak:"required,enum=AES256CBC|AES192CBC|AES128CBC|3DESCBC,src=IPSecTunnelProfile.java:108-110|IPsecConnectorInstanceAdvancedAttributes.java:111-115"`
	IntegrityAlgorithm  string `json:"integrityAlgorithm" ak:"required,enum=SHA1|SHA256|SHA384|SHA512|MD5,src=IPSecTunnelProfile.java:112-114|IPsecConnectorInstanceAdvancedAttributes.java:153-158"`
	DHGroup             string `json:"dhGroup" ak:"required,enum=MODP1024|MODP1536|MODP2048|MODP3072|MODP4096|MODP6144|MODP8192|ECP256|ECP384|ECP521|CURVE25519,src=IPSecTunnelProfile.java:116-118|IPsecConnectorInstanceAdvancedAttributes.java:196-207"`
}

// iPsecTunnelProfileDescriptor: every verb answers a missing id with 404
// AK000003 (IPSecTunnelProfileServiceImpl.java:128-130, :167-168, :178-179).
// A DELETE of a profile an advanced IPsec or ExpressRoute connector refers to
// is 400 AK000002 (:181-195); POST /ipsec-tunnel-profiles/reference-counts
// counts the referrers. A duplicate name in the network is 400 group.00000002
// on name (:320-326). A profile is not deployed: it has no state, and the
// delete removes the row.
var iPsecTunnelProfileDescriptor = Descriptor{
	Collection: "ipsec-tunnel-profiles",
	InUse:      []Signature{{Status: 400, Code: "AK000002", Message: "still referenced by connector"}},
	Duplicate:  []Signature{{Status: 400, FieldCode: "group.00000002"}},
	DocState:   DocStateMissingIsAdded,
	Uniqueness: Uniqueness{Key: "name", Scope: ScopeCollection},
	Service:    ServiceTPS,
	CrossCheck: checkIPsecTunnelProfile,
}

// checkIPsecTunnelProfile refuses a configuration without one of the
// algorithms the backend needs (IPSecTunnelProfile.java:89-118), which a
// patch can remove, and a GCM ESP algorithm with dhGroup NONE
// (IPSecTunnelProfileValidator.java:14-20).
func checkIPsecTunnelProfile(m map[string]any) error {
	var errs []error
	need := func(block string, fields ...string) map[string]any {
		b, _ := m[block].(map[string]any)
		for _, f := range fields {
			if v, _ := b[f].(string); v == "" {
				errs = append(errs, fmt.Errorf("%s.%s must be set", block, f))
			}
		}
		return b
	}
	esp := need("ipSecConfiguration", "encryptionAlgorithm", "dhGroup")
	need("ikeConfiguration", "encryptionAlgorithm", "integrityAlgorithm", "dhGroup")
	if esp["dhGroup"] == "NONE" {
		switch esp["encryptionAlgorithm"] {
		case "AES128GCM16", "AES192GCM16", "AES256GCM16":
			errs = append(errs, fmt.Errorf("ipSecConfiguration.encryptionAlgorithm %v is not allowed with dhGroup NONE", esp["encryptionAlgorithm"]))
		}
	}
	return errors.Join(errs...)
}

// IPsecTunnelProfileDescriptor returns the IPsec tunnel profile write descriptor.
func IPsecTunnelProfileDescriptor() Descriptor { return iPsecTunnelProfileDescriptor.clone() }

// NewIPsecTunnelProfileCore returns the write core for IPsec tunnel profiles.
func NewIPsecTunnelProfileCore(ac *AlkiraClient) (*Core[IPsecTunnelProfileConfig], error) {
	return NewCore[IPsecTunnelProfileConfig](ac, iPsecTunnelProfileDescriptor)
}

// CreateIPsecTunnelProfile creates an IPsec tunnel profile.
func (ac *AlkiraClient) CreateIPsecTunnelProfile(ctx context.Context, p *IPsecTunnelProfileConfig) (*IPsecTunnelProfileConfig, *WriteResult, error) {
	core, err := NewIPsecTunnelProfileCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, p, WriteOptions{})
}

// GetIPsecTunnelProfile reads one IPsec tunnel profile and its raw body.
func (ac *AlkiraClient) GetIPsecTunnelProfile(ctx context.Context, id string) (*IPsecTunnelProfileConfig, []byte, error) {
	core, err := NewIPsecTunnelProfileCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

// GetIPsecTunnelProfileByName pages the collection and returns every profile
// named name; none is ErrNotFound.
func (ac *AlkiraClient) GetIPsecTunnelProfileByName(ctx context.Context, name string) ([]IPsecTunnelProfileConfig, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: empty name", ErrInvalidInput)
	}
	all, err := ac.ListIPsecTunnelProfilePaged(ctx)
	if err != nil {
		return nil, err
	}
	var out []IPsecTunnelProfileConfig
	for _, p := range all {
		if p.Name == name {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%w: no IPsec tunnel profile named %q", ErrNotFound, name)
	}
	return out, nil
}

// UpdateIPsecTunnelProfile applies a merge patch to an IPsec tunnel profile.
func (ac *AlkiraClient) UpdateIPsecTunnelProfile(ctx context.Context, id string, patch json.RawMessage) (*IPsecTunnelProfileConfig, *WriteResult, error) {
	core, err := NewIPsecTunnelProfileCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

// DeleteIPsecTunnelProfile deletes an IPsec tunnel profile; a profile a
// connector tunnel refers to is refused as in use.
func (ac *AlkiraClient) DeleteIPsecTunnelProfile(ctx context.Context, id string) (*WriteResult, error) {
	core, err := NewIPsecTunnelProfileCore(ac)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

// ListIPsecTunnelProfilePaged reads every IPsec tunnel profile through paged reads.
func (ac *AlkiraClient) ListIPsecTunnelProfilePaged(ctx context.Context) ([]IPsecTunnelProfileConfig, error) {
	core, err := NewIPsecTunnelProfileCore(ac)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
