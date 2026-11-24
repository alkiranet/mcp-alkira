// Copyright (C) 2020-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
)

type BillingTag struct {
	Id          json.Number `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
}

// NewBillingTag
func NewBillingTag(ac *AlkiraClient) *AlkiraApi[BillingTag] {
	uri := fmt.Sprintf("%s/tags", ac.URI)
	api := &AlkiraApi[BillingTag]{ac, uri, PaginationOn}

	return api
}

// BillingTagConfig is the write model of a billing tag (tenant-billing
// model/Tag.java). An update changes only name and description
// (TagServiceImpl.java:133-134).
type BillingTagConfig struct {
	ID            json.Number `json:"id,omitempty" ak:"readonly"`
	TenantID      json.Number `json:"tenantId,omitempty" ak:"readonly"`
	Name          string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=Tag.java:29-33"`
	Description   *string     `json:"description,omitempty" ak:"maxlen=256,trimmed,charset=name,src=Tag.java:35-38"`
	Instances     []string    `json:"instances,omitempty" ak:"readonly"`
	UpdatedBy     string      `json:"updatedBy,omitempty" ak:"readonly"`
	LastUpdatedAt json.Number `json:"lastUpdatedAt,omitempty" ak:"readonly"`
}

// billingTagDescriptor: tenant-level /tags on tenant-billing. Every verb
// answers a missing id with 404 AK000003 (TagServiceImpl.java:89-90, :129-130,
// :146-147). The tag-count limit is a 404 too (TagServiceImpl.java:105-106),
// classified ErrLimitExceeded by its message before any signature applies.
var billingTagDescriptor = Descriptor{
	Collection:  "tags",
	TenantLevel: true,
	InUse:       []Signature{{Status: 400, FieldCode: "tag.00000002"}},
	Duplicate:   []Signature{{Status: 400, FieldCode: "tag.00000001"}},

	DocState:   DocStateMissingIsAdded,
	Uniqueness: Uniqueness{Key: "name", Scope: ScopeCollection},
	Service:    ServiceTenantBilling,
	Paging:     PagingPageRows,
	Flags: []FlagRule{
		{Flag: "BILLING-VIEW", Effect: FlagTypeAvailable, UISource: "src/views/settings/index.js:23|:43-45"},
		{Flag: "BILLING-TAGS", Effect: FlagTypeAvailable, UISource: "src/views/settings/index.js:24-26|:43-45"},
	},
}

// BillingTagDescriptor returns the billing tag write descriptor.
func BillingTagDescriptor() Descriptor { return billingTagDescriptor.clone() }

// NewBillingTagCore returns the write core for billing tags.
func NewBillingTagCore(ac *AlkiraClient) (*Core[BillingTagConfig], error) {
	return NewCore[BillingTagConfig](ac, billingTagDescriptor)
}

// CreateBillingTag creates a billing tag.
func (ac *AlkiraClient) CreateBillingTag(ctx context.Context, t *BillingTagConfig) (*BillingTagConfig, *WriteResult, error) {
	c, err := NewBillingTagCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Create(ctx, t, WriteOptions{})
}

// GetBillingTag reads one billing tag and its raw body.
func (ac *AlkiraClient) GetBillingTag(ctx context.Context, id string) (*BillingTagConfig, []byte, error) {
	c, err := NewBillingTagCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Get(ctx, id)
}

// UpdateBillingTag applies a merge patch to a billing tag.
func (ac *AlkiraClient) UpdateBillingTag(ctx context.Context, id string, patch json.RawMessage) (*BillingTagConfig, *WriteResult, error) {
	c, err := NewBillingTagCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Update(ctx, id, patch, ListEdits{})
}

// DeleteBillingTag deletes a billing tag.
func (ac *AlkiraClient) DeleteBillingTag(ctx context.Context, id string) (*WriteResult, error) {
	c, err := NewBillingTagCore(ac)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, id)
}

// ListBillingTagPaged reads every billing tag, paging until a short page.
func (ac *AlkiraClient) ListBillingTagPaged(ctx context.Context) ([]BillingTagConfig, error) {
	c, err := NewBillingTagCore(ac)
	if err != nil {
		return nil, err
	}
	return c.ListPaged(ctx)
}
