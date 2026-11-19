// Copyright (C) 2020-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

type PrefixList struct {
	Id   json.Number `json:"id"`
	Name string      `json:"name"`
}

// NewPrefixList new prefix list
func NewPrefixLists(ac *AlkiraClient) *AlkiraApi[PrefixList] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/policy/prefixlists", ac.URI, ac.TenantNetworkId)
	return &AlkiraApi[PrefixList]{ac, uri, PaginationOff}
}

// Special structure to return prefix lists that contains specific
// prefix
type PrefixListCompact struct {
	Id       json.Number `json:"id"`
	Name     string      `json:"name"`
	Prefixes []string    `json:"prefixes"`
	Type     string      `json:"type,omitempty"`
}

// GetByPrefix
//
// This function will retrieve all prefix lists and return prefix
// lists that contains specific prefix before API has the
// functionality.
func (aa *AlkiraApi[PrefixList]) GetByPrefix(prefix string) (string, error) {

	uri, err := url.Parse(aa.Uri)

	if err != nil {
		return "", fmt.Errorf("GetByPrefix: failed to parse URI %s: %v", aa.Uri, err)
	}

	// Disable pagination so we could read all resources at once
	q := uri.Query()
	q.Add("paginated", "false")

	uri.RawQuery = q.Encode()
	data, err := aa.Client.Get(uri.String())

	if err != nil {
		return "", fmt.Errorf("GetByPrefix: failed to get prefix list: %v", err)
	}

	logf("TRACE", "GetByPrefix: payload size %d", len(data))

	var lists []PrefixListCompact
	err = json.Unmarshal([]byte(data), &lists)

	if err != nil {
		return "", fmt.Errorf("GetByPrefix: failed to unmarshal: %v", err)
	}

	var result []PrefixListCompact

	for _, list := range lists {
		if len(list.Prefixes) > 0 {
			for _, onePrefix := range list.Prefixes {
				if prefix == onePrefix {
					result = append(result, list)
				}
			}
		}
	}

	// Marshal the summary data
	prefixLists, err := json.Marshal(result)

	if err != nil {
		return "", fmt.Errorf("GetByPrefix: failed to marshal: %v", err)
	}

	return string(prefixLists), err
}

// PrefixListConfig is the write model of a prefix list (policy-service
// model/PrefixList.java). type is derived from prefixRanges on every save
// (PrefixListServiceImpl.java:139, :275), so it is read-only. prefixes and
// prefixRanges carry no maxitems: their limit is the tenant's resource limit,
// unlimited when unset (SoftLimitValidator.java:48-53,
// ResourceLimitsServiceImpl.java:86-97), not the annotation's 1024 and 128.
type PrefixListConfig struct {
	ID              json.Number `json:"id,omitempty" ak:"readonly"`
	TenantID        json.Number `json:"tenantId,omitempty" ak:"readonly"`
	TenantNetworkID json.Number `json:"tenantNetworkId,omitempty" ak:"readonly"`
	Name            string      `json:"name" ak:"required,maxlen=255,trimmed,charset=name,src=PrefixList.java:44-48"`
	Description     *string     `json:"description,omitempty" ak:"maxlen=255,trimmed,charset=name,src=PrefixList.java:49-52"`

	Prefixes      []string                          `json:"prefixes,omitempty"`
	PrefixDetails map[string]PrefixListConfigDetail `json:"prefixDetails,omitempty"`
	PrefixRanges  []PrefixListConfigRange           `json:"prefixRanges,omitempty"`

	Type                string      `json:"type,omitempty" ak:"readonly"`
	LastUpdatedDate     json.Number `json:"lastUpdatedDate,omitempty" ak:"readonly"`
	LastConfigUpdatedAt json.Number `json:"lastConfigUpdatedAt,omitempty" ak:"readonly"`
}

// PrefixListConfigDetail is one prefixDetails entry (PrefixDetail.java).
type PrefixListConfigDetail struct {
	Description *string `json:"description,omitempty" ak:"maxlen=255,trimmed,charset=name,src=PrefixDetail.java:21-24"`
}

// PrefixListConfigRange is one prefixRanges entry (PrefixRange.java).
type PrefixListConfigRange struct {
	Prefix      string  `json:"prefix"`
	LE          *int64  `json:"le,omitempty"`
	GE          *int64  `json:"ge,omitempty"`
	Description *string `json:"description,omitempty" ak:"maxlen=255,trimmed,charset=name,src=PrefixRange.java:26-29"`
}

// prefixListDescriptor: policy-service answers a missing id with 404
// AK000003 and no id on GET, PUT and DELETE (PrefixListServiceImpl.java:167,
// :200, :289, PolicyExceptionHandler.java:52-53). A DELETE of a list referred
// by another object is 400 prefixlist.00000003 (:183), one used by an active
// rule 400 prefixlist.00000001 (:191); a duplicate name is 400
// prefixlist.00000002 (:223). PUT and DELETE answer 202. A GET by id returns
// rows whatever their docState (:308-311), and the body omits docState: a
// list deleted while rules still refer to it stays as a DELETED residue row
// (:153-189). The paged list leaves DELETED rows out
// (PrefixListRepository.java:22-33), so LiveInList refuses the residue.
var prefixListDescriptor = Descriptor{
	Collection: "policy/prefixlists",
	InUse:      []Signature{{Status: 400, FieldCode: "prefixlist.00000003"}, {Status: 400, FieldCode: "prefixlist.00000001"}},
	Duplicate:  []Signature{{Status: 400, FieldCode: "prefixlist.00000002"}},

	Provisionable: true,
	DocState:      DocStateMissingIsAdded,
	LiveInList:    true,
	Uniqueness:    Uniqueness{Key: "name", Scope: ScopeCollection},
	Service:       ServicePolicy,
}

// PrefixListDescriptor returns the prefix list write descriptor.
func PrefixListDescriptor() Descriptor { return prefixListDescriptor.clone() }

// NewPrefixListCore returns the write core for prefix lists.
func NewPrefixListCore(ac *AlkiraClient) (*Core[PrefixListConfig], error) {
	return NewCore[PrefixListConfig](ac, prefixListDescriptor)
}

// CreatePrefixList creates a prefix list.
func (ac *AlkiraClient) CreatePrefixList(ctx context.Context, p *PrefixListConfig) (*PrefixListConfig, *WriteResult, error) {
	c, err := NewPrefixListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Create(ctx, p, WriteOptions{})
}

// GetPrefixList reads one prefix list and its raw body.
func (ac *AlkiraClient) GetPrefixList(ctx context.Context, id string) (*PrefixListConfig, []byte, error) {
	c, err := NewPrefixListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Get(ctx, id)
}

// UpdatePrefixList applies a merge patch to a prefix list.
func (ac *AlkiraClient) UpdatePrefixList(ctx context.Context, id string, patch json.RawMessage) (*PrefixListConfig, *WriteResult, error) {
	c, err := NewPrefixListCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Update(ctx, id, patch, ListEdits{})
}

// DeletePrefixList deletes a prefix list.
func (ac *AlkiraClient) DeletePrefixList(ctx context.Context, id string) (*WriteResult, error) {
	c, err := NewPrefixListCore(ac)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, id)
}

// ListPrefixListPaged reads every prefix list through paged reads.
func (ac *AlkiraClient) ListPrefixListPaged(ctx context.Context) ([]PrefixListConfig, error) {
	c, err := NewPrefixListCore(ac)
	if err != nil {
		return nil, err
	}
	return c.ListPaged(ctx)
}
