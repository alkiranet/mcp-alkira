// Copyright (C) 2023-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
)

type DnsServerList struct {
	Id           json.Number `json:"id,omitempty"`
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	DnsServerIps []string    `json:"dnsServerIps"`
	Segment      string      `json:"segment"`
}

// NewDnsServerList new DNS server list
func NewDnsServerList(ac *AlkiraClient) *AlkiraApi[DnsServerList] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/dns-server-lists", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[DnsServerList]{ac, uri, PaginationOn}
	return api
}

// DnsServerListConfig is the write model of a DNS server list (TPS
// model/DNSServerList.java). segment is the segment name; the backend
// resolves it to segmentId on every save (DNSServerListServiceImpl.java:114-120,
// :145-151) and reads it back from segmentId, so a deleted segment leaves
// segmentId set and segment absent. dnsServerIps carries no item limit: the
// default is one IP, but the tenant's resource limits can raise it
// (DNSServerList.java:46-47).
type DnsServerListConfig struct {
	ID           json.Number `json:"id,omitempty" ak:"readonly"`
	Name         string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=DNSServerList.java:34-38"`
	Description  *string     `json:"description,omitempty" ak:"maxlen=512,trimmed,charset=name,src=DNSServerList.java:40-43"`
	DnsServerIps []string    `json:"dnsServerIps,omitempty" ak:"required,src=DNSServerList.java:45-50"`
	Segment      string      `json:"segment,omitempty" ak:"required,src=CreateDNSServerList.js:62|:72"`

	SegmentID           json.Number `json:"segmentId,omitempty" ak:"readonly"`
	LastConfigUpdatedAt json.Number `json:"lastConfigUpdatedAt,omitempty" ak:"readonly"`
}

// dnsServerListDescriptor: every verb answers a missing id with 404 AK000003
// (DNSServerListServiceImpl.java:101). A DELETE of a list a policy FQDN list
// refers to is 400 AK000002 (:160-162), one held by an external reference 409
// (NetworkConfigEntityListener.java:83-97); a duplicate name is 400
// constraint.exists on name (:133).
var dnsServerListDescriptor = Descriptor{
	Collection: "dns-server-lists",
	InUse:      []Signature{{Status: 400, Code: "AK000002"}, {Status: 409, Code: "AK000001"}},
	Duplicate:  []Signature{{Status: 400, FieldCode: "constraint.exists"}},

	Provisionable: true,
	DocState:      DocStateMissingIsAdded,
	Uniqueness:    Uniqueness{Key: "name", Scope: ScopeCollection},
	Service:       ServiceTPS,
}

// DnsServerListDescriptor returns the DNS server list write descriptor.
func DnsServerListDescriptor() Descriptor { return dnsServerListDescriptor.clone() }

// NewDnsServerListCore returns the write core for DNS server lists.
func NewDnsServerListCore(ac *AlkiraClient) (*Core[DnsServerListConfig], error) {
	return NewCore[DnsServerListConfig](ac, dnsServerListDescriptor)
}

// CreateDnsServerList creates a DNS server list.
func (ac *AlkiraClient) CreateDnsServerList(ctx context.Context, l *DnsServerListConfig) (*DnsServerListConfig, *WriteResult, error) {
	c, err := NewDnsServerListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Create(ctx, l, WriteOptions{})
}

// GetDnsServerList reads one DNS server list and its raw body.
func (ac *AlkiraClient) GetDnsServerList(ctx context.Context, id string) (*DnsServerListConfig, []byte, error) {
	c, err := NewDnsServerListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Get(ctx, id)
}

// UpdateDnsServerList applies a merge patch to a DNS server list.
func (ac *AlkiraClient) UpdateDnsServerList(ctx context.Context, id string, patch json.RawMessage) (*DnsServerListConfig, *WriteResult, error) {
	c, err := NewDnsServerListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Update(ctx, id, patch, ListEdits{})
}

// DeleteDnsServerList deletes a DNS server list.
func (ac *AlkiraClient) DeleteDnsServerList(ctx context.Context, id string) (*WriteResult, error) {
	c, err := NewDnsServerListCore(ac)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, id)
}

// ListDnsServerListPaged reads every DNS server list through paged reads.
func (ac *AlkiraClient) ListDnsServerListPaged(ctx context.Context) ([]DnsServerListConfig, error) {
	c, err := NewDnsServerListCore(ac)
	if err != nil {
		return nil, err
	}
	return c.ListPaged(ctx)
}
