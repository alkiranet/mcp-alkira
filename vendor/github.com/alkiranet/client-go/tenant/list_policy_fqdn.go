// Copyright (C) 2023-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
)

type PolicyFqdnList struct {
	Id              json.Number `json:"id,omitempty"`
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	Fqdns           []string    `json:"fqdns"`
	DnsServerListId int         `json:"dnsServerListId"`
}

// NewPolicyFqdnList new global cidr list
func NewPolicyFqdnList(ac *AlkiraClient) *AlkiraApi[PolicyFqdnList] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/policy-fqdn-lists", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[PolicyFqdnList]{ac, uri, PaginationOn}
	return api
}

// PolicyFqdnListConfig is the write model of a policy FQDN list (TPS
// model/PolicyFqdnList.java). fqdns keeps its order. It carries no item
// limit: the default is one FQDN, but the tenant's resource limits can raise
// it (PolicyFqdnList.java:46-47). metadata holds the resolved addresses the
// backend maintains (PolicyFqdnListServiceImpl.java:236-251).
type PolicyFqdnListConfig struct {
	ID              json.Number `json:"id,omitempty" ak:"readonly"`
	Name            string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=PolicyFqdnList.java:34-38"`
	Description     *string     `json:"description,omitempty" ak:"maxlen=512,trimmed,charset=name,src=PolicyFqdnList.java:40-43"`
	Fqdns           []string    `json:"fqdns,omitempty" ak:"required,src=PolicyFqdnList.java:45-50"`
	DnsServerListID *int64      `json:"dnsServerListId,omitempty" ak:"required,src=PolicyFqdnList.java:52-53"`

	Metadata            any         `json:"metadata,omitempty" ak:"readonly"`
	LastConfigUpdatedAt json.Number `json:"lastConfigUpdatedAt,omitempty" ak:"readonly"`
}

// policyFqdnListDescriptor: every verb answers a missing id with 404
// AK000003 (PolicyFqdnListServiceImpl.java:115). A DELETE of a list a NAT
// rule or internet application refers to is 400 AK000002 (:165-178), one
// held by an external reference 409 (NetworkConfigEntityListener.java:83-97);
// a duplicate name is 400 constraint.exists on name (:140).
var policyFqdnListDescriptor = Descriptor{
	Collection: "policy-fqdn-lists",
	InUse:      []Signature{{Status: 400, Code: "AK000002"}, {Status: 409, Code: "AK000001"}},
	Duplicate:  []Signature{{Status: 400, FieldCode: "constraint.exists"}},

	Provisionable: true,
	DocState:      DocStateMissingIsAdded,
	Uniqueness:    Uniqueness{Key: "name", Scope: ScopeCollection},
	Service:       ServiceTPS,
}

// PolicyFqdnListDescriptor returns the policy FQDN list write descriptor.
func PolicyFqdnListDescriptor() Descriptor { return policyFqdnListDescriptor.clone() }

// NewPolicyFqdnListCore returns the write core for policy FQDN lists.
func NewPolicyFqdnListCore(ac *AlkiraClient) (*Core[PolicyFqdnListConfig], error) {
	return NewCore[PolicyFqdnListConfig](ac, policyFqdnListDescriptor)
}

// CreatePolicyFqdnList creates a policy FQDN list.
func (ac *AlkiraClient) CreatePolicyFqdnList(ctx context.Context, l *PolicyFqdnListConfig) (*PolicyFqdnListConfig, *WriteResult, error) {
	c, err := NewPolicyFqdnListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Create(ctx, l, WriteOptions{})
}

// GetPolicyFqdnList reads one policy FQDN list and its raw body.
func (ac *AlkiraClient) GetPolicyFqdnList(ctx context.Context, id string) (*PolicyFqdnListConfig, []byte, error) {
	c, err := NewPolicyFqdnListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Get(ctx, id)
}

// UpdatePolicyFqdnList applies a merge patch to a policy FQDN list.
func (ac *AlkiraClient) UpdatePolicyFqdnList(ctx context.Context, id string, patch json.RawMessage) (*PolicyFqdnListConfig, *WriteResult, error) {
	c, err := NewPolicyFqdnListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Update(ctx, id, patch, ListEdits{})
}

// DeletePolicyFqdnList deletes a policy FQDN list.
func (ac *AlkiraClient) DeletePolicyFqdnList(ctx context.Context, id string) (*WriteResult, error) {
	c, err := NewPolicyFqdnListCore(ac)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, id)
}

// ListPolicyFqdnListPaged reads every policy FQDN list through paged reads.
func (ac *AlkiraClient) ListPolicyFqdnListPaged(ctx context.Context) ([]PolicyFqdnListConfig, error) {
	c, err := NewPolicyFqdnListCore(ac)
	if err != nil {
		return nil, err
	}
	return c.ListPaged(ctx)
}
