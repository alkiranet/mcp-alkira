// Copyright (C) 2021-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
)

type GlobalCidrList struct {
	Description string      `json:"description"`
	CXP         string      `json:"cxp"`
	Id          json.Number `json:"id,omitempty"`
	Name        string      `json:"name"`
	Tags        []string    `json:"tags,omitempty"`
	Values      []string    `json:"values"`
}

// NewGlobalCidrList new global cidr list
func NewGlobalCidrList(ac *AlkiraClient) *AlkiraApi[GlobalCidrList] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/global-cidr-lists", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[GlobalCidrList]{ac, uri, PaginationOn}
	return api
}

// GlobalCidrListConfig is the write model of a global CIDR list (TPS
// model/GlobalCidrList.java). cxp is required: the backend accepts a list
// without it, but every consumer needs one and the portal always sends it.
// values carries no item limit: the default is one CIDR per list, but the
// tenant's resource limits can raise it (GlobalCidrList.java:53-54). An
// update sends the whole row (GlobalCidrListServiceImpl.java:148-160), so
// tags is resent from the stored object unless the update changes it.
type GlobalCidrListConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=GlobalCidrList.java:46-50"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,trimmed,charset=name,src=GlobalCidrList.java:62-65"`
	CXP         string      `json:"cxp,omitempty" ak:"required,src=CreateSecondaryCidrList.js:118|:129"`
	Values      []string    `json:"values,omitempty" ak:"required,trimmed,src=GlobalCidrList.java:52-58"`
	Tags        []string    `json:"tags,omitempty" ak:"set"`

	LastConfigUpdatedAt json.Number `json:"lastConfigUpdatedAt,omitempty" ak:"readonly"`
}

// globalCidrListDescriptor: every verb answers a missing id with 404
// AK000003 (GlobalCidrListServiceImpl.java:268-270). A DELETE of a list a
// service refers to is 400 AK000002 (:174-180), one held by an external
// reference 409 (NetworkConfigEntityListener.java:83-97). A duplicate name is
// 400 constraint.invalid on name (:308-311), the code every other value and
// CXP refusal of a save shares, so it carries no duplicate signature; the
// create guard reads the collection first.
var globalCidrListDescriptor = Descriptor{
	Collection: "global-cidr-lists",
	InUse:      []Signature{{Status: 400, Code: "AK000002"}, {Status: 409, Code: "AK000001"}},

	Provisionable: true,
	DocState:      DocStateMissingIsAdded,
	Uniqueness:    Uniqueness{Key: "name", Scope: ScopeCollection},
	Service:       ServiceTPS,
}

// GlobalCidrListDescriptor returns the global CIDR list write descriptor.
func GlobalCidrListDescriptor() Descriptor { return globalCidrListDescriptor.clone() }

// NewGlobalCidrListCore returns the write core for global CIDR lists.
func NewGlobalCidrListCore(ac *AlkiraClient) (*Core[GlobalCidrListConfig], error) {
	return NewCore[GlobalCidrListConfig](ac, globalCidrListDescriptor)
}

// CreateGlobalCidrList creates a global CIDR list.
func (ac *AlkiraClient) CreateGlobalCidrList(ctx context.Context, l *GlobalCidrListConfig) (*GlobalCidrListConfig, *WriteResult, error) {
	c, err := NewGlobalCidrListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Create(ctx, l, WriteOptions{})
}

// GetGlobalCidrList reads one global CIDR list and its raw body.
func (ac *AlkiraClient) GetGlobalCidrList(ctx context.Context, id string) (*GlobalCidrListConfig, []byte, error) {
	c, err := NewGlobalCidrListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Get(ctx, id)
}

// UpdateGlobalCidrList applies a merge patch to a global CIDR list.
func (ac *AlkiraClient) UpdateGlobalCidrList(ctx context.Context, id string, patch json.RawMessage) (*GlobalCidrListConfig, *WriteResult, error) {
	c, err := NewGlobalCidrListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Update(ctx, id, patch, ListEdits{})
}

// DeleteGlobalCidrList deletes a global CIDR list.
func (ac *AlkiraClient) DeleteGlobalCidrList(ctx context.Context, id string) (*WriteResult, error) {
	c, err := NewGlobalCidrListCore(ac)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, id)
}

// ListGlobalCidrListPaged reads every global CIDR list through paged reads.
func (ac *AlkiraClient) ListGlobalCidrListPaged(ctx context.Context) ([]GlobalCidrListConfig, error) {
	c, err := NewGlobalCidrListCore(ac)
	if err != nil {
		return nil, err
	}
	return c.ListPaged(ctx)
}
