// Copyright (C) 2020-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
)

//
// The concept of group is not clearly defined in the system. It seems
// like the group API also lists all other "types" of groups as well.
//

// A summary of a group (other fields are ignored from the original
// response)
type Group struct {
	Id           json.Number `json:"id"`
	Name         string      `json:"name"`
	Description  string      `json:"description,omitempty"`
	InternalName string      `json:"internalName"`
	Type         string      `json:"type"`
}

func NewGroup(ac *AlkiraClient) *AlkiraApi[Group] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/groups", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[Group]{ac, uri, PaginationOn}
	return api
}

// GroupConfig is the write model of an explicit group (TPS model/Group.java).
// The backend sets type EXPLICIT on create and an update changes only name
// and description (GroupServiceImpl.java:293, :307-308), so every other field
// is read-only.
type GroupConfig struct {
	ID              json.Number `json:"id,omitempty" ak:"readonly"`
	TenantID        json.Number `json:"tenantId,omitempty" ak:"readonly"`
	TenantNetworkID json.Number `json:"tenantNetworkId,omitempty" ak:"readonly"`
	Name            string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=Group.java:64-68"`
	Description     *string     `json:"description,omitempty" ak:"maxlen=512,trimmed,charset=name,src=Group.java:70-73"`
	InternalName    string      `json:"internalName,omitempty" ak:"readonly"`

	Implicit            *bool       `json:"implicit,omitempty" ak:"readonly"`
	Tags                []int64     `json:"tags,omitempty" ak:"readonly"`
	Type                string      `json:"type,omitempty" ak:"readonly"`
	CXP                 string      `json:"cxp,omitempty" ak:"readonly"`
	TagID               *int64      `json:"tagId,omitempty" ak:"readonly"`
	FirewallZoneIDs     []int64     `json:"firewallZoneIds,omitempty" ak:"readonly"`
	SystemGenerated     *bool       `json:"systemGenerated,omitempty" ak:"readonly"`
	RoutingTags         []any       `json:"routingTags,omitempty" ak:"readonly"`
	SourceID            string      `json:"sourceId,omitempty" ak:"readonly"`
	LastConfigUpdatedAt json.Number `json:"lastConfigUpdatedAt,omitempty" ak:"readonly"`
}

// groupDescriptor: GET answers a missing id with 404 AK000001
// (GroupServiceImpl.java:444), PUT with 404 AK000003 (:302) and DELETE with
// 400 AK000002 "Group with id: '<id>' was not found" (:354,
// messages.properties:3). A DELETE refused for a connector is
// 400 group.00000010 (:366), for a firewall zone 400 group.00000006 (:379),
// for an InterCXP route policy 400 AK000002 (:393), and for a segment
// resource or policy 409 (:362, :368). Group and connector names share one
// name space (:321-324). TPS groups carry no state (model/Group.java), so
// there is no state guard. The collection also lists the groups the backend
// creates (connector groups, "<segment>-ALL-Group"); TPS refuses an update
// or delete of a group whose type is not EXPLICIT (400 group.00000001,
// group.00000003; :301-302, :353-354), so the family check refuses them first;
// GetGroup returns them.
var groupDescriptor = Descriptor{
	Collection:     "groups",
	FamilyType:     FamilyCheck{Field: "type", Value: "EXPLICIT"},
	NotFound:       []Signature{{Status: 404, Code: "AK000001"}, {Status: 404, Code: "AK000003"}},
	DeleteNotFound: []Signature{{Status: 400, Code: "AK000002", Message: "Group with id: '"}},
	InUse: []Signature{
		{Status: 400, FieldCode: "group.00000010"},
		{Status: 400, FieldCode: "group.00000006"},
		{Status: 400, Code: "AK000002", Message: "in use by an InterCXP route policy"},
		{Status: 409, Code: "AK000001"},
	},
	Duplicate: []Signature{{Status: 400, FieldCode: "group.00000002"}},

	Provisionable: true,
	DocState:      DocStateMissingIsAdded,
	Uniqueness:    Uniqueness{Key: "name", Scope: ScopeGroupsConnectors},
	Service:       ServiceTPS,
}

// GroupDescriptor returns the group write descriptor.
func GroupDescriptor() Descriptor { return groupDescriptor.clone() }

// NewGroupCore returns the write core for groups.
func NewGroupCore(ac *AlkiraClient) (*Core[GroupConfig], error) {
	return NewCore[GroupConfig](ac, groupDescriptor)
}

// CreateGroup creates an explicit group.
func (ac *AlkiraClient) CreateGroup(ctx context.Context, g *GroupConfig) (*GroupConfig, *WriteResult, error) {
	c, err := NewGroupCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Create(ctx, g, WriteOptions{})
}

// GetGroup reads one group and its raw body.
func (ac *AlkiraClient) GetGroup(ctx context.Context, id string) (*GroupConfig, []byte, error) {
	c, err := NewGroupCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Get(ctx, id)
}

// UpdateGroup applies a merge patch to a group.
func (ac *AlkiraClient) UpdateGroup(ctx context.Context, id string, patch json.RawMessage) (*GroupConfig, *WriteResult, error) {
	c, err := NewGroupCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Update(ctx, id, patch, ListEdits{})
}

// DeleteGroup deletes a group.
func (ac *AlkiraClient) DeleteGroup(ctx context.Context, id string) (*WriteResult, error) {
	c, err := NewGroupCore(ac)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, id)
}

// ListGroupPaged reads every group, implicit ones included, through paged reads.
func (ac *AlkiraClient) ListGroupPaged(ctx context.Context) ([]GroupConfig, error) {
	c, err := NewGroupCore(ac)
	if err != nil {
		return nil, err
	}
	return c.ListPaged(ctx)
}

//
// User Group
//

type UserGroup struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewUserGroup(ac *AlkiraClient) *AlkiraApi[UserGroup] {
	uri := fmt.Sprintf("%s/user-groups", ac.URI)
	api := &AlkiraApi[UserGroup]{ac, uri, PaginationOff}
	return api
}

//
// Inter Connector Communication Group
//

type InterConnectorCommunicationGroup struct {
	Id                           json.Number `json:"id"`
	Name                         string      `json:"name"`
	Description                  string      `json:"description"`
	Segment                      string      `json:"segment"`
	Cxp                          string      `json:"cxp"`
	ConnectorProviderRegion      string      `json:"connectorProviderRegion"`
	ConnectorType                string      `json:"connectorType"`
	VirtualNetworkManagerAzureId int         `json:"azureVirtualNetworkManagerId"`
}

func NewInterConnectorCommunicationGroup(ac *AlkiraClient) *AlkiraApi[InterConnectorCommunicationGroup] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/inter-connector-communication-groups", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[InterConnectorCommunicationGroup]{ac, uri, PaginationOn}
	return api
}
