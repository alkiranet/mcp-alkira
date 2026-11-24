// Copyright (C) 2022-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

type SegmentResourceShare struct {
	Id                  json.Number `json:"id,omitempty"` // response only
	Name                string      `json:"name"`
	Description         string      `json:"description,omitempty"`
	ServiceList         []int       `json:"serviceList"`
	DesignatedSegment   string      `json:"designatedSegment"`
	EndAResources       []int       `json:"endAResources"`
	EndBResources       []int       `json:"endBResources"`
	EndARouteLimit      int         `json:"endARouteLimit"`
	EndBRouteLimit      int         `json:"endBRouteLimit"`
	FromEnd             string      `json:"fromEnd,omitempty"`
	Direction           string      `json:"direction"`
	DesignatedSegmentId int         `json:"designatedSegmentId,omitempty"` // response only
	EndASegmentId       int         `json:"endASegmentId,omitempty"`       // response only
	EndBSegmentId       int         `json:"endBSegmentId,omitempty"`       // response only
	RuleListId          int         `json:"ruleListId,omitempty"`
}

// NewSegmentResourceShare new segment resource share
func NewSegmentResourceShare(ac *AlkiraClient) *AlkiraApi[SegmentResourceShare] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/segment-resource-shares", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[SegmentResourceShare]{ac, uri, PaginationOn}
	return api
}

// SegmentResourceShareConfig is the write model of a segment resource share
// (TPS model/SegmentResourceShare.java). Optional fields are pointers or
// omitempty slices, so a zero value is never sent. The backend creates and
// updates an INTER_SEGMENT policy for the share asynchronously
// (SegmentResourceShareServiceImpl.java:311-314, :531-534); it is not shown.
type SegmentResourceShareConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=SegmentResourceShare.java:57-61"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,trimmed,charset=name,src=SegmentResourceShare.java:63-66"`
	// DesignatedSegment is a segment name; it cannot change after the first
	// provision (SegmentResourceShareServiceImpl.java:467-474).
	DesignatedSegment string  `json:"designatedSegment" ak:"required,maxlen=256,trimmed,charset=name,src=SegmentResourceShare.java:68-72"`
	EndAResources     []int64 `json:"endAResources" ak:"required,src=SegmentResourceShare.java:75-79"`
	EndBResources     []int64 `json:"endBResources" ak:"required,src=SegmentResourceShare.java:82-86"`
	EndARouteLimit    *int64  `json:"endARouteLimit,omitempty" ak:"default=100,min=1,src=SegmentResourceShare.java:88-93"`
	EndBRouteLimit    *int64  `json:"endBRouteLimit,omitempty" ak:"default=100,min=1,src=SegmentResourceShare.java:95-100"`
	Direction         *string `json:"direction" ak:"required,enum=UNIDIRECTIONAL|BIDIRECTIONAL,src=SegmentResourceShare.java:102-104|:150-153"`
	// FromEnd is the end traffic starts from; refused on a bidirectional share
	// (SegmentResourceShareServiceImpl.java:319-322).
	FromEnd *string `json:"fromEnd,omitempty" ak:"enum=A|B,src=SegmentResourceShare.java:106|:311-314"`
	// Deprecated by the backend in favour of serviceList.
	ServiceID *int64 `json:"serviceId,omitempty" ak:"expired=1780272000,src=SegmentResourceShare.java:120-122"`
	// ServiceList holds firewall service ids; [0] or an empty list means no
	// firewall and [-1] any firewall (SegmentResourceShare.java:51-52,
	// SegmentResourceShareServiceImpl.java:928-930).
	ServiceList []int64 `json:"serviceList,omitempty" ak:"min=-1,src=SegmentResourceShare.java:124-128"`
	RuleListID  *int64  `json:"ruleListId,omitempty"`

	State               string      `json:"state,omitempty" ak:"readonly"`
	DesignatedSegmentID *int64      `json:"designatedSegmentId,omitempty" ak:"readonly"`
	EndASegmentID       *int64      `json:"endASegmentId,omitempty" ak:"readonly"`
	EndBSegmentID       *int64      `json:"endBSegmentId,omitempty" ak:"readonly"`
	LastConfigUpdatedAt json.Number `json:"lastConfigUpdatedAt,omitempty" ak:"readonly"`
}

// segmentResourceShareDescriptor: GET and PUT answer a missing id with 404
// AK000003 (SegmentResourceShareServiceImpl.java:239-246,
// AbstractNetworkCRUDService.java:208), DELETE too (:291). The share delete
// has no refusal of its own (:570-601); an external reference refuses it
// with 409 (NetworkConfigEntityListener.java:83-97). A duplicate name is 400
// constraint.notunique (:993-1002). A FAILED share is not updated but
// re-provisioned as stored (AbstractNetworkCRUDService.java:214-218).
var segmentResourceShareDescriptor = Descriptor{
	Collection: "segment-resource-shares",
	InUse:      []Signature{{Status: 409, Code: "AK000001", Message: "is still referred externally"}},
	Duplicate:  []Signature{{Status: 400, FieldCode: "constraint.notunique"}},

	Provisionable: true,
	DocState:      DocStateMissingIsAdded,
	Uniqueness:    Uniqueness{Key: "name", Scope: ScopeCollection},
	StateGuard:    StateGuardFailedOnly,
	Service:       ServiceTPS,
	CrossCheck:    checkSegmentResourceShare,
}

// checkSegmentResourceShare refuses fromEnd on a bidirectional share
// (SegmentResourceShareServiceImpl.java:319-322) and a ruleListId together
// with a serviceId or a non-empty serviceList (:556-558). It reports every
// rule the body breaks.
func checkSegmentResourceShare(m map[string]any) error {
	var errs []error
	if m["direction"] == "BIDIRECTIONAL" && m["fromEnd"] != nil {
		errs = append(errs, errors.New("fromEnd must not be set on a BIDIRECTIONAL share"))
	}
	if m["ruleListId"] != nil {
		if l, ok := m["serviceList"].([]any); m["serviceId"] != nil || ok && len(l) > 0 {
			errs = append(errs, errors.New("set either ruleListId or serviceId/serviceList, not both"))
		}
	}
	return errors.Join(errs...)
}

// SegmentResourceShareDescriptor returns the segment resource share write descriptor.
func SegmentResourceShareDescriptor() Descriptor { return segmentResourceShareDescriptor.clone() }

// NewSegmentResourceShareCore returns the write core for segment resource shares.
func NewSegmentResourceShareCore(ac *AlkiraClient) (*Core[SegmentResourceShareConfig], error) {
	return NewCore[SegmentResourceShareConfig](ac, segmentResourceShareDescriptor)
}

// CreateSegmentResourceShare creates a segment resource share.
func (ac *AlkiraClient) CreateSegmentResourceShare(ctx context.Context, s *SegmentResourceShareConfig) (*SegmentResourceShareConfig, *WriteResult, error) {
	c, err := NewSegmentResourceShareCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Create(ctx, s, WriteOptions{})
}

// GetSegmentResourceShare reads one segment resource share and its raw body.
func (ac *AlkiraClient) GetSegmentResourceShare(ctx context.Context, id string) (*SegmentResourceShareConfig, []byte, error) {
	c, err := NewSegmentResourceShareCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Get(ctx, id)
}

// UpdateSegmentResourceShare applies a merge patch to a segment resource share.
func (ac *AlkiraClient) UpdateSegmentResourceShare(ctx context.Context, id string, patch json.RawMessage) (*SegmentResourceShareConfig, *WriteResult, error) {
	c, err := NewSegmentResourceShareCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Update(ctx, id, patch, ListEdits{})
}

// DeleteSegmentResourceShare deletes a segment resource share.
func (ac *AlkiraClient) DeleteSegmentResourceShare(ctx context.Context, id string) (*WriteResult, error) {
	c, err := NewSegmentResourceShareCore(ac)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, id)
}

// ListSegmentResourceSharePaged reads every segment resource share through paged reads.
func (ac *AlkiraClient) ListSegmentResourceSharePaged(ctx context.Context) ([]SegmentResourceShareConfig, error) {
	c, err := NewSegmentResourceShareCore(ac)
	if err != nil {
		return nil, err
	}
	return c.ListPaged(ctx)
}
