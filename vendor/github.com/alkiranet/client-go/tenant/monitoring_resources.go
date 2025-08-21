// Copyright (C) 2025 Alkira Inc. All Rights Reserved.

package tenant

import (
	"fmt"
	"net/url"
)

// GetResourceUsage get alerts with optional query parameters
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
	data, _, err := ac.get(uri.String())

	return string(data), err
}


// GetResourceUsage get alerts with optional query parameters
func (ac *AlkiraClient) GetResourceLimits() (string, error) {
	uri := fmt.Sprintf("%s/api/resourcelimits", ac.URI)
	data, _, err := ac.get(uri)
	return string(data), err
}
