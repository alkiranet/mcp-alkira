// Copyright (C) 2022-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
)

type ByoipExtraAttributes struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
	PublicKey string `json:"publicKey"`
}

type Byoip struct {
	ExtraAttributes ByoipExtraAttributes `json:"extraAttributes"`
	CloudProvider   string               `json:"cloudProvider"`
	Prefix          string               `json:"prefix"`
	Cxp             string               `json:"cxp"`
	Description     string               `json:"description"`
	Id              json.Number          `json:"id,omitempty"`
	DoNotAdvertise  bool                 `json:"doNotAdvertise"`
}

// NewByoip new BYOIP
func NewByoip(ac *AlkiraClient) *AlkiraApi[Byoip] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/byoips", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[Byoip]{ac, uri, PaginationOn}

	return api
}

// ByoipConfig is the write model of a BYOIP prefix (TPS model/BYOIP.java).
// It has no name: prefix identifies it and is unique across every tenant
// (entity/BYOIPEntity.java:37-40). An update changes description, and
// extraAttributes only while no connector or application uses the prefix;
// the backend ignores a changed prefix, cxp or cloudProvider
// (BYOIPServiceImpl.java:229-248). While no connector uses the prefix, an
// update copies onboardingStatus, poolId and extraAttributes from the
// request (:233-237), and onboardingStatus defaults to PENDING
// (BYOIP.java:59-61), so these are tagged keep: a create or patch cannot set
// them, and an update resends the stored values.
type ByoipConfig struct {
	ID              json.Number            `json:"id,omitempty" ak:"readonly"`
	Prefix          string                 `json:"prefix" ak:"required,immutable,src=BYOIP.java:33-36"`
	CXP             string                 `json:"cxp" ak:"required,immutable,src=BYOIP.java:38-40"`
	CloudProvider   *string                `json:"cloudProvider,omitempty" ak:"default=AWS,enum=AWS|AZURE,immutable,src=BYOIP.java:44-45|:124-131"`
	Description     *string                `json:"description,omitempty" ak:"maxlen=512,trimmed,charset=name,src=BYOIP.java:47-50"`
	ExtraAttributes *ByoipExtraAttrsConfig `json:"extraAttributes" ak:"required,src=BYOIP.java:31|BYOIPAWSExtraAttributes.java:21-29"`
	DoNotAdvertise  *bool                  `json:"doNotAdvertise,omitempty" ak:"flag=AKAMAI-PRO-PHASE-1"`

	ExtraAttributesType string      `json:"extraAttributesType,omitempty" ak:"readonly"`
	OnboardingStatus    string      `json:"onboardingStatus,omitempty" ak:"readonly,keep"`
	UsageCount          *int64      `json:"usageCount,omitempty" ak:"readonly"`
	CreatedAt           json.Number `json:"createdAt,omitempty" ak:"readonly"`
	PoolID              string      `json:"poolId,omitempty" ak:"readonly,keep"`
	ErrorMessage        string      `json:"errorMessage,omitempty" ak:"readonly"`
	LastConfigUpdatedAt json.Number `json:"lastConfigUpdatedAt,omitempty" ak:"readonly"`
}

// ByoipExtraAttrsConfig is the signed ROA of a BYOIP prefix; the backend
// checks the signature with the public key (BYOIPValidator.java:26-70).
type ByoipExtraAttrsConfig struct {
	Message   string `json:"message" ak:"required,src=BYOIPAWSExtraAttributes.java:21-29"`
	Signature string `json:"signature" ak:"required,src=BYOIPAWSExtraAttributes.java:21-29"`
	PublicKey string `json:"publicKey" ak:"required,src=BYOIPAWSExtraAttributes.java:21-29"`

	TenantCustomPrefixName string `json:"tenantCustomPrefixName,omitempty" ak:"readonly,keep"`
	TenantPublicPrefixName string `json:"tenantPublicPrefixName,omitempty" ak:"readonly,keep"`
}

// byoipDescriptor: GET and PUT answer a missing id with 404 AK000003
// (BYOIPServiceImpl.java:139-140, AbstractNetworkCRUDService.java:208),
// DELETE too (:291). A DELETE is refused with 400 AK000002 while an internet
// connector uses the prefix (BYOIPServiceImpl.java:272-279). A duplicate
// prefix is 400 constraint.notunique (:106-112); the backend looks it up
// across every tenant, so the guard, which reads this network's list, can
// miss a prefix another network holds. BYOIP carries no state
// (DiffableNetworkEntity.java:14-20).
var byoipDescriptor = Descriptor{
	Collection: "byoips",
	InUse:      []Signature{{Status: 400, Code: "AK000002", Message: "already associated with"}},
	Duplicate:  []Signature{{Status: 400, FieldCode: "constraint.notunique"}},

	Provisionable: true,
	DocState:      DocStateMissingIsAdded,
	Uniqueness:    Uniqueness{Key: "prefix", Scope: ScopeCollection},
	Service:       ServiceTPS,
	Flags: []FlagRule{
		{Flag: "AKAMAI-PRO-PHASE-1", Effect: FlagIncludeField, Field: "doNotAdvertise", UISource: "src/views/byoip/index.js:139|createOrEditBYOIP.js:71-73"},
	},
}

// ByoipDescriptor returns the BYOIP write descriptor.
func ByoipDescriptor() Descriptor { return byoipDescriptor.clone() }

// NewByoipCore returns the write core for BYOIP prefixes.
func NewByoipCore(ac *AlkiraClient) (*Core[ByoipConfig], error) {
	return NewCore[ByoipConfig](ac, byoipDescriptor)
}

// CreateByoip creates a BYOIP prefix.
func (ac *AlkiraClient) CreateByoip(ctx context.Context, b *ByoipConfig) (*ByoipConfig, *WriteResult, error) {
	c, err := NewByoipCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Create(ctx, b, WriteOptions{})
}

// GetByoip reads one BYOIP prefix and its raw body.
func (ac *AlkiraClient) GetByoip(ctx context.Context, id string) (*ByoipConfig, []byte, error) {
	c, err := NewByoipCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Get(ctx, id)
}

// UpdateByoip applies a merge patch to a BYOIP prefix.
func (ac *AlkiraClient) UpdateByoip(ctx context.Context, id string, patch json.RawMessage) (*ByoipConfig, *WriteResult, error) {
	c, err := NewByoipCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Update(ctx, id, patch, ListEdits{})
}

// DeleteByoip deletes a BYOIP prefix.
func (ac *AlkiraClient) DeleteByoip(ctx context.Context, id string) (*WriteResult, error) {
	c, err := NewByoipCore(ac)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, id)
}

// ListByoipPaged reads every BYOIP prefix of the network through paged reads.
func (ac *AlkiraClient) ListByoipPaged(ctx context.Context) ([]ByoipConfig, error) {
	c, err := NewByoipCore(ac)
	if err != nil {
		return nil, err
	}
	return c.ListPaged(ctx)
}
