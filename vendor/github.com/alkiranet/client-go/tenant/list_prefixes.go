// Copyright (C) 2020-2025 Alkira Inc. All Rights Reserved.

package tenant

import (
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
