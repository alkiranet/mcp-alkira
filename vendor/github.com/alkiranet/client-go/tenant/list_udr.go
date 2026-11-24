// Copyright (C) 2024-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
)

type UdrListUdrs struct {
	Prefix       string `json:"prefix"`
	NextHopType  string `json:"nextHopType"`
	NextHopValue string `json:"nextHopValue"`
	Description  string `json:"description,omitempty"`
}

type UdrList struct {
	Name          string        `json:"name"`
	Description   string        `json:"description,omitempty"`
	CloudProvider string        `json:"cloudProvider"`
	Id            json.Number   `json:"id,omitempty"`
	Udrs          []UdrListUdrs `json:"udrs"`
}

// NewUdrList new UDR list
func NewUdrList(ac *AlkiraClient) *AlkiraApi[UdrList] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/udr-lists", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[UdrList]{ac, uri, PaginationOn}
	return api
}

// UdrListConfig is the write model of a UDR list (TPS model/UDRList.java).
// udrs carries no item limit: the default is 16 routes, but the tenant's
// resource limits can change it (UDRList.java:56).
type UdrListConfig struct {
	ID            json.Number          `json:"id,omitempty" ak:"readonly"`
	Name          string               `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=UDRList.java:41-45"`
	Description   *string              `json:"description,omitempty" ak:"maxlen=512,trimmed,charset=name,src=UDRList.java:47-50"`
	CloudProvider string               `json:"cloudProvider,omitempty" ak:"required,enum=AZURE,src=UDRList.java:52-53|UDRListServiceImpl.java:281-285"`
	Udrs          []UdrListConfigRoute `json:"udrs,omitempty" ak:"required"`

	LastConfigUpdatedAt json.Number `json:"lastConfigUpdatedAt,omitempty" ak:"readonly"`
}

// UdrListConfigRoute is one UDRList.udrs entry (model/UDR.java).
// nextHopValue must be empty for INTERNET, the only next hop type
// (UDRValidator.java:34-37), so it is never sent.
type UdrListConfigRoute struct {
	Prefix       string  `json:"prefix,omitempty" ak:"required"`
	NextHopType  string  `json:"nextHopType,omitempty" ak:"required,enum=INTERNET,src=UDR.java:29-30|:41-43"`
	NextHopValue *string `json:"nextHopValue,omitempty" ak:"readonly"`
	Description  *string `json:"description,omitempty" ak:"maxlen=512,trimmed,charset=name,src=UDR.java:34-39"`
}

// udrListDescriptor: every verb answers a missing id with 404 AK000003
// (UDRListServiceImpl.java:229). A DELETE of a list a connector refers to is
// 400 constraint.forbidden on udrList (:158-161); the backend keeps the
// reference for minutes after the connector is deleted, and the refusal is
// not retried. A duplicate name hits the database unique index
// (entity/UDRListEntity.java:34-36) and answers 409 AK000001 whose message is
// the raw database error (CustomExceptionHandler.java:197-205); the
// signature matches the database error text, never the constraint or table
// name.
var udrListDescriptor = Descriptor{
	Collection: "udr-lists",
	InUse:      []Signature{{Status: 400, FieldCode: "constraint.forbidden"}, {Status: 409, Code: "AK000001"}},
	Duplicate:  []Signature{{Status: 409, Code: "AK000001", Message: "duplicate key value violates unique constraint"}},

	Provisionable: true,
	DocState:      DocStateMissingIsAdded,
	Uniqueness:    Uniqueness{Key: "name", Scope: ScopeCollection},
	Service:       ServiceTPS,
}

// UdrListDescriptor returns the UDR list write descriptor.
func UdrListDescriptor() Descriptor { return udrListDescriptor.clone() }

// NewUdrListCore returns the write core for UDR lists.
func NewUdrListCore(ac *AlkiraClient) (*Core[UdrListConfig], error) {
	return NewCore[UdrListConfig](ac, udrListDescriptor)
}

// CreateUdrList creates a UDR list.
func (ac *AlkiraClient) CreateUdrList(ctx context.Context, l *UdrListConfig) (*UdrListConfig, *WriteResult, error) {
	c, err := NewUdrListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Create(ctx, l, WriteOptions{})
}

// GetUdrList reads one UDR list and its raw body.
func (ac *AlkiraClient) GetUdrList(ctx context.Context, id string) (*UdrListConfig, []byte, error) {
	c, err := NewUdrListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Get(ctx, id)
}

// UpdateUdrList applies a merge patch to a UDR list.
func (ac *AlkiraClient) UpdateUdrList(ctx context.Context, id string, patch json.RawMessage) (*UdrListConfig, *WriteResult, error) {
	c, err := NewUdrListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Update(ctx, id, patch, ListEdits{})
}

// DeleteUdrList deletes a UDR list.
func (ac *AlkiraClient) DeleteUdrList(ctx context.Context, id string) (*WriteResult, error) {
	c, err := NewUdrListCore(ac)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, id)
}

// ListUdrListPaged reads every UDR list through paged reads.
func (ac *AlkiraClient) ListUdrListPaged(ctx context.Context) ([]UdrListConfig, error) {
	c, err := NewUdrListCore(ac)
	if err != nil {
		return nil, err
	}
	return c.ListPaged(ctx)
}
