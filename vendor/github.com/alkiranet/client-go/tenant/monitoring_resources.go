// Copyright (C) 2025 Alkira Inc. All Rights Reserved.

package tenant

import (
	"fmt"
	"net/url"
)

// GetResourceUsage get resource usages of the tenant
//
// Optional query parameters could be added to filter responses.
//
// resourceCategory resource category
// resourceType resource type
// resourceScope resource scope
func (ac *AlkiraClient) GetResourceUsages(resourceCategory string, resourceType string, resourceScope string) (string, error) {

	baseUri := fmt.Sprintf("%s/api/resourceusage", ac.URI)

	uri, err := url.Parse(baseUri)

	if err != nil {
		return "", fmt.Errorf("GetResourceUsage: failed to parse URI: %v", err)
	}

	// Process optional query parameters
	q := uri.Query()

	if resourceCategory != "" {
		q.Add("category", resourceCategory)
	}

	if resourceType != "" {
		q.Add("type", resourceType)
	}

	if resourceScope != "" {
		q.Add("scope", resourceScope)
	}

	// GET
	uri.RawQuery = q.Encode()
	data, err := ac.Get(uri.String())

	return string(data), err
}

// GetResourceLimits get resource limits of the tenant
func (ac *AlkiraClient) GetResourceLimits() (string, error) {
	uri := fmt.Sprintf("%s/api/resourcelimits", ac.URI)
	data, err := ac.Get(uri)
	return string(data), err
}
