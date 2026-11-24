// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

var ztaProfileNotFound = []Signature{{Status: 400, FieldCode: "constraint.notexist"}}

// ZtaProfile is a zero trust access profile (user-manager
// /zero-trust-access-profiles), in the alkira-client-go shape
// (ACG zta_profile.go:9-15). A traffic policy names profiles in ztaProfileIds.
type ZtaProfile struct {
	Id   string `json:"id,omitempty"` // only set on response
	Name string `json:"name"`
}

// NewZtaProfile returns the read API of zero trust access profiles; the
// collection is tenant level and paged by default.
func NewZtaProfile(ac *AlkiraClient) *AlkiraApi[ZtaProfile] {
	uri := fmt.Sprintf("%s/zero-trust-access-profiles", ac.URI)
	return &AlkiraApi[ZtaProfile]{ac, uri, PaginationOn}
}

// GetZtaProfile reads one zero trust access profile and its raw body. A
// missing or deleted id, which user-manager answers with 400 and field code
// constraint.notexist (ZeroTrustAccessProfileServiceImpl.java:199-205;
// UserManagerExceptionHandler.java:56-63), returns ErrNotFound.
func (ac *AlkiraClient) GetZtaProfile(ctx context.Context, id string) (*ZtaProfile, []byte, error) {
	if id == "" {
		return nil, nil, fmt.Errorf("%w: empty id", ErrInvalidInput)
	}
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()
	raw, err := ac.doGet(ctx, ac.URI+"/zero-trust-access-profiles/"+pathSegment(id))
	if err != nil {
		var ae *APIError
		if errors.As(err, &ae) && ae.kind == nil && ae.matches(ztaProfileNotFound) {
			ae.kind = ErrNotFound
		}
		return nil, nil, err
	}
	var p ZtaProfile
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, raw, fmt.Errorf("zero trust access profile %s: %w", id, err)
	}
	return &p, raw, nil
}

// ListZtaProfilePaged reads every zero trust access profile through paged
// reads; pagination.hits is the total (ZeroTrustAccessProfileServiceImpl.java:141-144).
func (ac *AlkiraClient) ListZtaProfilePaged(ctx context.Context) ([]ZtaProfile, error) {
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()
	rows, err := ac.listPaged(ctx, ac.URI+"/zero-trust-access-profiles", PagingTotalHits)
	if err != nil {
		return nil, err
	}
	out := make([]ZtaProfile, len(rows))
	for i, r := range rows {
		if err := json.Unmarshal(r, &out[i]); err != nil {
			return nil, fmt.Errorf("zero trust access profiles: row %d: %w", i, err)
		}
	}
	return out, nil
}
