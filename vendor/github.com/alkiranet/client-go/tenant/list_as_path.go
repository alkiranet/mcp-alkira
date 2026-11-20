// Copyright (C) 2021-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
)

type AsPathList struct {
	Description string      `json:"description"`
	Id          json.Number `json:"id,omitempty"`
	Name        string      `json:"name"`
	Values      []string    `json:"values"`
}

func NewListAsPath(ac *AlkiraClient) *AlkiraApi[AsPathList] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/as-path-lists", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[AsPathList]{ac, uri, PaginationOn}
	return api
}

// asPathListDescriptor: every verb answers a missing id with 404 AK000003,
// and so does a GET or PUT of a row marked for deletion
// (AsPathListServiceImpl.java:129-130). A DELETE of a list held by a route
// policy is 400 AK000002 (:174-181), one held by an external reference 409
// (NetworkConfigEntityListener.java:83-97); a duplicate name is 400
// constraint.exists on name (:154). A list held only by rules of deleted
// route policies is marked for deletion, not removed (:188-191). The body
// always carries docState (AsPathList.java:57-60).
var asPathListDescriptor = Descriptor{
	Collection: "as-path-lists",
	InUse:      []Signature{{Status: 400, Code: "AK000002"}, {Status: 409, Code: "AK000001"}},
	Duplicate:  []Signature{{Status: 400, FieldCode: "constraint.exists"}},

	Provisionable: true,
	DocState:      DocStateField,
	Uniqueness:    Uniqueness{Key: "name", Scope: ScopeCollection},
	Service:       ServiceTPS,
}

// AsPathListDescriptor returns the AS-path list write descriptor.
func AsPathListDescriptor() Descriptor { return asPathListDescriptor.clone() }

// NewAsPathListCore returns the write core for AS-path lists.
func NewAsPathListCore(ac *AlkiraClient) (*Core[ValueListConfig], error) {
	return NewCore[ValueListConfig](ac, asPathListDescriptor)
}

// CreateAsPathList creates an AS-path list.
func (ac *AlkiraClient) CreateAsPathList(ctx context.Context, l *ValueListConfig) (*ValueListConfig, *WriteResult, error) {
	c, err := NewAsPathListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Create(ctx, l, WriteOptions{})
}

// GetAsPathList reads one AS-path list and its raw body.
func (ac *AlkiraClient) GetAsPathList(ctx context.Context, id string) (*ValueListConfig, []byte, error) {
	c, err := NewAsPathListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Get(ctx, id)
}

// UpdateAsPathList applies a merge patch to an AS-path list.
func (ac *AlkiraClient) UpdateAsPathList(ctx context.Context, id string, patch json.RawMessage) (*ValueListConfig, *WriteResult, error) {
	c, err := NewAsPathListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Update(ctx, id, patch, ListEdits{})
}

// DeleteAsPathList deletes an AS-path list.
func (ac *AlkiraClient) DeleteAsPathList(ctx context.Context, id string) (*WriteResult, error) {
	c, err := NewAsPathListCore(ac)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, id)
}

// ListAsPathListPaged reads every AS-path list through paged reads.
func (ac *AlkiraClient) ListAsPathListPaged(ctx context.Context) ([]ValueListConfig, error) {
	c, err := NewAsPathListCore(ac)
	if err != nil {
		return nil, err
	}
	return c.ListPaged(ctx)
}
