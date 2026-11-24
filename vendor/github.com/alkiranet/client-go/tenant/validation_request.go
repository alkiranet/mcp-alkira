// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// DryRunValidationFlag enables dry-run validation requests (karrawa
// 5d18e83fd diffV2/index.js gates the validate action on it).
const DryRunValidationFlag = "RELEASE-DRY-RUN-NETWORK-VALIDATION"

// ValidatedChange is one change a dry-run validation covered.
type ValidatedChange struct {
	ChangeType string `json:"changeType"`
	EntityType string `json:"entityType"`
	EntityID   string `json:"entityId"`
	EntityName string `json:"entityName,omitempty"`
}

// ValidationRequestError is one error a dry-run validation reported.
type ValidationRequestError struct {
	Message        string `json:"message"`
	ExceptionClass string `json:"exceptionClass,omitempty"`
}

// ValidationRequest is one dry-run validation
// (TPS NetworkValidationResult.java). State is IN_PROGRESS, SUCCESS or FAILED.
type ValidationRequest struct {
	ValidationRequestID   string                   `json:"validationRequestId"`
	State                 string                   `json:"state"`
	StartedAt             int64                    `json:"startedAt,omitempty"`
	CompletedAt           int64                    `json:"completedAt,omitempty"`
	Initiator             string                   `json:"initiator,omitempty"`
	BaseSnapshotRequestID string                   `json:"baseSnapshotRequestId,omitempty"`
	Changes               []ValidatedChange        `json:"changes,omitempty"`
	Errors                []ValidationRequestError `json:"errors,omitempty"`
}

// CreateValidationRequest starts a dry-run validation of the selected
// changes with POST /tenantnetworks/{id}/validation-requests, the body the
// portal sends (the provision body without cleanup options), and returns the
// x-validation-request-id. Nothing is saved or provisioned.
func (ac *AlkiraClient) CreateValidationRequest(ctx context.Context, changes []ProvisionChange) (string, *WriteResult, error) {
	if len(changes) == 0 {
		return "", nil, fmt.Errorf("%w: no changes selected", ErrInvalidInput)
	}
	body, err := json.Marshal(changes)
	if err != nil {
		return "", nil, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()
	res, err := ac.doWrite(ctx, "POST", ac.networkURI("validation-requests"), body)
	if err != nil {
		return "", nil, err
	}
	id := res.ValidationRequestID
	if id == "" {
		var vr ValidationRequest
		if json.Unmarshal(res.Body, &vr) == nil {
			id = vr.ValidationRequestID
		}
	}
	return id, res, nil
}

// GetValidationRequest reads one dry-run validation by id.
func (ac *AlkiraClient) GetValidationRequest(ctx context.Context, id string) (*ValidationRequest, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: empty validation request id", ErrInvalidInput)
	}
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()
	raw, err := ac.doGet(ctx, ac.networkURI("validation-requests", pathSegment(id)))
	if err != nil {
		return nil, err
	}
	var vr ValidationRequest
	if err := json.Unmarshal(raw, &vr); err != nil {
		return nil, fmt.Errorf("validation request %s: %w", id, err)
	}
	return &vr, nil
}

// ListValidationStates reads every async validation state in the given
// states with ?state=...&paginated=true, the read the portal makes
// (karrawa topology.js:117-131), paging by offset until pagination.hits.
func (ac *AlkiraClient) ListValidationStates(ctx context.Context, states ...string) ([]ValidationState, error) {
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()
	var filter url.Values
	if len(states) > 0 {
		filter = url.Values{"state": states}
	}
	rows, err := ac.listPagedWith(ctx, ac.networkURI("validation-states"), PagingTotalHits, filter)
	if err != nil {
		return nil, err
	}
	out := make([]ValidationState, len(rows))
	for i, r := range rows {
		if err := json.Unmarshal(r, &out[i]); err != nil {
			return nil, fmt.Errorf("validation states: row %d: %w", i, err)
		}
	}
	return out, nil
}
