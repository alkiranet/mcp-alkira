// Copyright (C) 2022-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
)

type SegmentResource struct {
	Id            json.Number                  `json:"id"`
	Name          string                       `json:"name"`
	Description   string                       `json:"description,omitempty"`
	Segment       string                       `json:"segment"`
	GroupId       int                          `json:"groupId,omitempty"` // response only
	GroupPrefixes []SegmentResourceGroupPrefix `json:"groupPrefixes"`
}

type SegmentResourceGroupPrefix struct {
	GroupId      int `json:"groupId"`
	PrefixListId int `json:"prefixListId"`
}

// NewSegmentResource new segment resource
func NewSegmentResource(ac *AlkiraClient) *AlkiraApi[SegmentResource] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/segment-resources", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[SegmentResource]{ac, uri, PaginationOn}
	return api
}

// SegmentResourceConfig is the write model of a segment resource (TPS
// model/SegmentResource.java). Optional fields are pointers, so a zero value
// is never sent. The resource carries no state and no docState
// (SegmentResourceEntity.java:44-78).
type SegmentResourceConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=SegmentResource.java:58-62"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,trimmed,charset=name,src=SegmentResource.java:64-67"`
	// Segment is the segment name (SegmentResourceServiceImpl.java:382).
	Segment       string                             `json:"segment" ak:"required,trimmed,charset=name,src=SegmentResource.java:69-72"`
	GroupPrefixes []SegmentResourceConfigGroupPrefix `json:"groupPrefixes" ak:"required,src=SegmentResource.java:74-76"`

	// GroupID is the implicit "<name>-Group" group the backend creates with
	// the resource (SegmentResourceServiceImpl.java:405-426).
	GroupID             *int64      `json:"groupId,omitempty" ak:"readonly"`
	SystemGenerated     *bool       `json:"systemGenerated,omitempty" ak:"readonly"`
	SegmentID           *int64      `json:"segmentId,omitempty" ak:"readonly"`
	LastConfigUpdatedAt json.Number `json:"lastConfigUpdatedAt,omitempty" ak:"readonly"`
	SubConnectors       []any       `json:"subConnectors,omitempty" ak:"readonly"`
	SubServiceGroups    []any       `json:"subServiceGroups,omitempty" ak:"readonly"`
}

// SegmentResourceConfigGroupPrefix maps a group to a prefix list; -1 means
// any group or any prefix list (SegmentResource.java:51-52).
type SegmentResourceConfigGroupPrefix struct {
	GroupID      *int64 `json:"groupId" ak:"required,min=-1,src=SegmentResource.java:106-108"`
	PrefixListID *int64 `json:"prefixListId" ak:"required,min=-1,src=SegmentResource.java:110-112"`
}

// segmentResourceDescriptor: GET and PUT answer a missing id with 404
// AK000003 (SegmentResourceServiceImpl.java:278-283,
// AbstractNetworkCRUDService.java:208), DELETE too (:291). A DELETE is
// refused with 409 while route policies use the resource (:1306-1313), with
// 400 AK000002 for an InterCXP route policy (:1315-1318), with 400
// constraint.forbidden while shares hold it (:1321-1322; the same code
// refuses deleting a system-generated resource, :1303-1304) and with 409 for an
// external reference (NetworkConfigEntityListener.java:83-97). A duplicate
// name, or a name whose "-Group" group exists, is 400 constraint.exists
// (:390-391, :416-419, :497-498, :545-550). The resource has no state, so
// there is no state guard.
var segmentResourceDescriptor = Descriptor{
	Collection: "segment-resources",
	InUse: []Signature{
		{Status: 409, Code: "AK000001"},
		{Status: 400, Code: "AK000002", Message: "in use by an InterCXP route policy"},
		{Status: 400, FieldCode: "constraint.forbidden"},
	},
	Duplicate: []Signature{{Status: 400, FieldCode: "constraint.exists"}},

	Provisionable: true,
	DocState:      DocStateMissingIsAdded,
	Uniqueness:    Uniqueness{Key: "name", Scope: ScopeCollection, GroupSuffix: "-Group"},
	Service:       ServiceTPS,
}

// SegmentResourceDescriptor returns the segment resource write descriptor.
func SegmentResourceDescriptor() Descriptor { return segmentResourceDescriptor.clone() }

// NewSegmentResourceCore returns the write core for segment resources.
func NewSegmentResourceCore(ac *AlkiraClient) (*Core[SegmentResourceConfig], error) {
	return NewCore[SegmentResourceConfig](ac, segmentResourceDescriptor)
}

// CreateSegmentResource creates a segment resource.
func (ac *AlkiraClient) CreateSegmentResource(ctx context.Context, r *SegmentResourceConfig) (*SegmentResourceConfig, *WriteResult, error) {
	c, err := NewSegmentResourceCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Create(ctx, r, WriteOptions{})
}

// GetSegmentResource reads one segment resource and its raw body.
func (ac *AlkiraClient) GetSegmentResource(ctx context.Context, id string) (*SegmentResourceConfig, []byte, error) {
	c, err := NewSegmentResourceCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Get(ctx, id)
}

// UpdateSegmentResource applies a merge patch to a segment resource.
func (ac *AlkiraClient) UpdateSegmentResource(ctx context.Context, id string, patch json.RawMessage) (*SegmentResourceConfig, *WriteResult, error) {
	c, err := NewSegmentResourceCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Update(ctx, id, patch, ListEdits{})
}

// DeleteSegmentResource deletes a segment resource and its implicit group.
func (ac *AlkiraClient) DeleteSegmentResource(ctx context.Context, id string) (*WriteResult, error) {
	c, err := NewSegmentResourceCore(ac)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, id)
}

// ListSegmentResourcePaged reads every segment resource through paged reads.
func (ac *AlkiraClient) ListSegmentResourcePaged(ctx context.Context) ([]SegmentResourceConfig, error) {
	c, err := NewSegmentResourceCore(ac)
	if err != nil {
		return nil, err
	}
	return c.ListPaged(ctx)
}
