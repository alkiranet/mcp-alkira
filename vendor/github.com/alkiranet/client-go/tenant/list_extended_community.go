// Copyright (C) 2021-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
)

type ExtendedCommunityList struct {
	Description string      `json:"description"`
	Id          json.Number `json:"id,omitempty"`
	Name        string      `json:"name"`
	Values      []string    `json:"values"`
}

func NewListExtendedCommunity(ac *AlkiraClient) *AlkiraApi[ExtendedCommunityList] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/extended-community-lists", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[ExtendedCommunityList]{ac, uri, PaginationOn}
	return api
}

// extendedCommunityListDescriptor: every verb answers a missing id with 404
// AK000003 (ExtendedCommunityListServiceImpl.java:108). A DELETE of a list
// held by a route policy is 400 AK000002 (:161-169), one held by an external
// reference 409 (NetworkConfigEntityListener.java:83-97); a duplicate name is
// 400 constraint.exists on name (:133). Values that mix rt or soo values with
// regex values are refused before the request (:136-144).
var extendedCommunityListDescriptor = Descriptor{
	Collection: "extended-community-lists",
	InUse:      []Signature{{Status: 400, Code: "AK000002"}, {Status: 409, Code: "AK000001"}},
	Duplicate:  []Signature{{Status: 400, FieldCode: "constraint.exists"}},

	Provisionable: true,
	DocState:      DocStateMissingIsAdded,
	Uniqueness:    Uniqueness{Key: "name", Scope: ScopeCollection},
	Service:       ServiceTPS,
	CrossCheck:    checkExtCommunityMix,
}

// ExtendedCommunityListDescriptor returns the extended community list write descriptor.
func ExtendedCommunityListDescriptor() Descriptor { return extendedCommunityListDescriptor.clone() }

// NewExtendedCommunityListCore returns the write core for extended community lists.
func NewExtendedCommunityListCore(ac *AlkiraClient) (*Core[ValueListConfig], error) {
	return NewCore[ValueListConfig](ac, extendedCommunityListDescriptor)
}

// CreateExtendedCommunityList creates an extended community list.
func (ac *AlkiraClient) CreateExtendedCommunityList(ctx context.Context, l *ValueListConfig) (*ValueListConfig, *WriteResult, error) {
	c, err := NewExtendedCommunityListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Create(ctx, l, WriteOptions{})
}

// GetExtendedCommunityList reads one extended community list and its raw body.
func (ac *AlkiraClient) GetExtendedCommunityList(ctx context.Context, id string) (*ValueListConfig, []byte, error) {
	c, err := NewExtendedCommunityListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Get(ctx, id)
}

// UpdateExtendedCommunityList applies a merge patch to an extended community list.
func (ac *AlkiraClient) UpdateExtendedCommunityList(ctx context.Context, id string, patch json.RawMessage) (*ValueListConfig, *WriteResult, error) {
	c, err := NewExtendedCommunityListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Update(ctx, id, patch, ListEdits{})
}

// DeleteExtendedCommunityList deletes an extended community list.
func (ac *AlkiraClient) DeleteExtendedCommunityList(ctx context.Context, id string) (*WriteResult, error) {
	c, err := NewExtendedCommunityListCore(ac)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, id)
}

// ListExtendedCommunityListPaged reads every extended community list through paged reads.
func (ac *AlkiraClient) ListExtendedCommunityListPaged(ctx context.Context) ([]ValueListConfig, error) {
	c, err := NewExtendedCommunityListCore(ac)
	if err != nil {
		return nil, err
	}
	return c.ListPaged(ctx)
}
