// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// StaleNetworkForChangesFlag lets the change lists answer from the cached
// network (karrawa 5d18e83fd diffV2/index.js:406-438 sends staleOk=true with it).
const StaleNetworkForChangesFlag = "USE-STALE-NETWORK-FOR-CHANGES"

// NetworkChangeEntity is the entity of one pending change (TPS
// NetworkEntity.java; policy-service NetworkPolicyEntity.java).
type NetworkChangeEntity struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	Type              string          `json:"type"`
	SubType           string          `json:"subType,omitempty"`
	LastModifiedAt    int64           `json:"lastModifiedAt,omitempty"`
	Properties        json.RawMessage `json:"properties,omitempty"`
	ImplicitReference bool            `json:"implicitReference,omitempty"`
	ProvisionedState  string          `json:"provisionedState,omitempty"`
	DocState          string          `json:"docState,omitempty"`
}

// NetworkChange is one pending change since the last successful provision
// (TPS NetworkEntityDiff.java; policy-service NetworkPolicyEntityDiff.java).
type NetworkChange struct {
	Entity                NetworkChangeEntity `json:"entity"`
	ChangeType            string              `json:"changeType"` // ADDED, MODIFIED or DELETED
	BestEffortAllowed     bool                `json:"bestEffortAllowed,omitempty"`
	AvailableCleanupTypes []string            `json:"availableCleanupTypes,omitempty"`
	LastUpdatedByUserID   string              `json:"lastUpdatedByUserId,omitempty"`
	LastUpdatedBy         string              `json:"lastUpdatedBy,omitempty"`
	ImplicitReference     bool                `json:"implicitReference,omitempty"`
}

// GetNetworkChanges reads the TPS pending changes,
// GET /tenantnetworks/{id}/changes?since=LAST_SUCCESSFUL_PROVISION.
func (ac *AlkiraClient) GetNetworkChanges(ctx context.Context, staleOk bool) ([]NetworkChange, error) {
	return ac.changes(ctx, ac.networkURI("changes"), staleOk)
}

// GetNetworkPolicyChanges reads the policy-service pending changes,
// GET /tenantnetworks/{id}/networkpolicy/changes?since=LAST_SUCCESSFUL_PROVISION.
// Policy-service families appear only here, not in GetNetworkChanges.
func (ac *AlkiraClient) GetNetworkPolicyChanges(ctx context.Context, staleOk bool) ([]NetworkChange, error) {
	return ac.changes(ctx, ac.networkURI("networkpolicy", "changes"), staleOk)
}

func (ac *AlkiraClient) changes(ctx context.Context, uri string, staleOk bool) ([]NetworkChange, error) {
	q := url.Values{"since": {"LAST_SUCCESSFUL_PROVISION"}}
	if staleOk {
		q.Set("staleOk", "true")
	}
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()
	raw, err := ac.doGet(ctx, withQuery(uri, q))
	if err != nil {
		return nil, err
	}
	var out []NetworkChange
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("GET %s: %w", uri, err)
	}
	return out, nil
}

// GetSelfUserID returns the id of the user the client authenticates as,
// from GET /user/self, the read the portal makes (karrawa src/_api/users.js:7).
// It is the value of lastUpdatedByUserId on that user's changes.
func (ac *AlkiraClient) GetSelfUserID(ctx context.Context) (string, error) {
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()
	raw, err := ac.doGet(ctx, ac.URI+"/user/self")
	if err != nil {
		return "", err
	}
	var u struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &u); err != nil {
		return "", fmt.Errorf("user self: %w", err)
	}
	if u.ID == "" {
		return "", fmt.Errorf("user self: %w: no id", ErrNotFound)
	}
	return u.ID, nil
}
