// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
)

// ConnectorRef is one row of GET /connectors.
type ConnectorRef struct {
	ID   json.Number `json:"id"`
	Name string      `json:"name"`
	Type string      `json:"type"`
}

// ListConnectors returns every connector of every family from
// GET /tenantnetworks/{id}/connectors, one complete array that ignores
// paging parameters. Rows of a just-deleted connector stay in it for some
// seconds; confirm a row with a single GET before relying on it.
func (ac *AlkiraClient) ListConnectors(ctx context.Context) ([]ConnectorRef, error) {
	raw, err := ac.doGet(ctx, ac.URI+"/tenantnetworks/"+ac.TenantNetworkId+"/connectors")
	if err != nil {
		return nil, err
	}
	var rows []ConnectorRef
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("GET connectors: %w", err)
	}
	return rows, nil
}
