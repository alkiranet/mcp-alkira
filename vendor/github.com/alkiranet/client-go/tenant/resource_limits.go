// Copyright (C) 2025 Alkira Inc. All Rights Reserved.

package tenant

import (
	"fmt"
)

// GetResourceLimits get resource limits of the tenant
func (ac *AlkiraClient) GetResourceLimits() (string, error) {
	uri := fmt.Sprintf("%s/api/resourcelimits", ac.URI)
	data, err := ac.Get(uri)
	return string(data), err
}
