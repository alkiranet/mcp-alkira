// Copyright (C) 2020-2025 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
)

type BillingTag struct {
	Id          json.Number `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
}

// NewBillingTag
func NewBillingTag(ac *AlkiraClient) *AlkiraApi[BillingTag] {
	uri := fmt.Sprintf("%s/tags", ac.URI)
	api := &AlkiraApi[BillingTag]{ac, uri}

	return api
}
