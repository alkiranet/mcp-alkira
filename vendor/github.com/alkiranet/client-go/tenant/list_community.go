// Copyright (C) 2021-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
)

type CommunityList struct {
	Description string      `json:"description"`
	Id          json.Number `json:"id,omitempty"`
	Name        string      `json:"name"`
	Values      []string    `json:"values"`
}

func NewListCommunity(ac *AlkiraClient) *AlkiraApi[CommunityList] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/community-lists", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[CommunityList]{ac, uri, PaginationOn}
	return api
}

// communityListDescriptor: every verb answers a missing id with 404 AK000003
// (CommunityListServiceImpl.java:116). A DELETE of a list held by a route
// policy is 400 AK000002 (:158-166), one held by an external reference 409
// (NetworkConfigEntityListener.java:83-97); a duplicate name is 400
// constraint.exists on name (:140). The body carries no docState.
var communityListDescriptor = Descriptor{
	Collection: "community-lists",
	InUse:      []Signature{{Status: 400, Code: "AK000002"}, {Status: 409, Code: "AK000001"}},
	Duplicate:  []Signature{{Status: 400, FieldCode: "constraint.exists"}},

	Provisionable: true,
	DocState:      DocStateMissingIsAdded,
	Uniqueness:    Uniqueness{Key: "name", Scope: ScopeCollection},
	Service:       ServiceTPS,
}

// CommunityListDescriptor returns the community list write descriptor.
func CommunityListDescriptor() Descriptor { return communityListDescriptor.clone() }

// NewCommunityListCore returns the write core for community lists.
func NewCommunityListCore(ac *AlkiraClient) (*Core[ValueListConfig], error) {
	return NewCore[ValueListConfig](ac, communityListDescriptor)
}

// CreateCommunityList creates a community list.
func (ac *AlkiraClient) CreateCommunityList(ctx context.Context, l *ValueListConfig) (*ValueListConfig, *WriteResult, error) {
	c, err := NewCommunityListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Create(ctx, l, WriteOptions{})
}

// GetCommunityList reads one community list and its raw body.
func (ac *AlkiraClient) GetCommunityList(ctx context.Context, id string) (*ValueListConfig, []byte, error) {
	c, err := NewCommunityListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Get(ctx, id)
}

// UpdateCommunityList applies a merge patch to a community list.
func (ac *AlkiraClient) UpdateCommunityList(ctx context.Context, id string, patch json.RawMessage) (*ValueListConfig, *WriteResult, error) {
	c, err := NewCommunityListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Update(ctx, id, patch, ListEdits{})
}

// DeleteCommunityList deletes a community list.
func (ac *AlkiraClient) DeleteCommunityList(ctx context.Context, id string) (*WriteResult, error) {
	c, err := NewCommunityListCore(ac)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, id)
}

// ListCommunityListPaged reads every community list through paged reads.
func (ac *AlkiraClient) ListCommunityListPaged(ctx context.Context) ([]ValueListConfig, error) {
	c, err := NewCommunityListCore(ac)
	if err != nil {
		return nil, err
	}
	return c.ListPaged(ctx)
}
