// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ValidationRunning is WriteResult.ValidationState while an async
// validation has not reached a terminal state.
const ValidationRunning = "validating"

// ValidationUnknown is WriteResult.ValidationState of an async create whose
// response, and with it the validation state id, was lost; a name lookup
// found the object. The validation can still be running or have FAILED, so
// the caller must check the object's validation state before provisioning.
const ValidationUnknown = "unknown"

// Terminal async validation states.
const (
	ValidationSucceeded = "SUCCESS"
	ValidationFailed    = "FAILED"
)

// Poll schedule for an async write's validation state: every 10 s, the
// interval alkira-client-go uses, for about 20 s, inside the call deadline.
var (
	validationPollInterval = 10 * time.Second
	validationPollBudget   = 20 * time.Second
	// validationGetAllowance is the time left for the state GET after a wait;
	// a wait that would not leave it before the deadline is not started.
	validationGetAllowance = 2 * time.Second
)

// ValidationErrorDetails is the errorDetails block of a validation state.
type ValidationErrorDetails struct {
	Code    string       `json:"code,omitempty"`
	ID      string       `json:"id,omitempty"`
	Message string       `json:"message,omitempty"`
	Errors  []FieldError `json:"errors,omitempty"`
}

// ValidationState is one async validation from
// GET /tenantnetworks/{id}/validation-states/{TYPE:SUBTYPE:ID}.
type ValidationState struct {
	ID                   int64                   `json:"id"`
	NetworkEntityType    string                  `json:"networkEntityType"`
	NetworkEntitySubType string                  `json:"networkEntitySubType,omitempty"`
	NetworkEntityID      string                  `json:"networkEntityId"`
	State                string                  `json:"state"` // PENDING, IN_PROGRESS, FAILED or SUCCESS
	PreviousState        string                  `json:"previousState,omitempty"`
	ErrorDetails         *ValidationErrorDetails `json:"errorDetails,omitempty"`
	UpdatedAt            int64                   `json:"updatedAt,omitempty"`
}

// GetValidationState reads one async validation state by the composite id
// of the x-ak-validation-state-id header (TYPE:SUBTYPE:ID). A state that
// does not exist returns ErrNotFound.
func (ac *AlkiraClient) GetValidationState(ctx context.Context, id string) (*ValidationState, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: empty validation state id", ErrInvalidInput)
	}
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()
	raw, err := ac.doGet(ctx, ac.URI+"/tenantnetworks/"+ac.TenantNetworkId+"/validation-states/"+pathSegment(id))
	if err != nil {
		return nil, err
	}
	// TPS answers 200 with an empty or null body when no state row exists.
	if t := bytes.TrimSpace(raw); len(t) == 0 || string(t) == "null" {
		return nil, fmt.Errorf("validation state %s: %w", id, ErrNotFound)
	}
	var vs ValidationState
	if err := json.Unmarshal(raw, &vs); err != nil {
		return nil, fmt.Errorf("validation state %s: %w", id, err)
	}
	return &vs, nil
}

// awaitValidation polls the validation state of an async write until it is
// terminal, the poll budget is spent or the next wait would pass the
// deadline. res keeps ValidationRunning unless a terminal state is read. A
// state read that a retry cannot change (not found, or a 4xx other than 408
// and 429) ends the poll and is kept in res.ValidationReadError; other read
// errors are retried at the next interval.
func (ac *AlkiraClient) awaitValidation(ctx context.Context, res *WriteResult) {
	res.ValidationState = ValidationRunning
	for waited := validationPollInterval; waited <= validationPollBudget; waited += validationPollInterval {
		if deadline, ok := ctx.Deadline(); ok && deadline.Sub(clockNow()) <= validationPollInterval+validationGetAllowance {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-clockAfter(validationPollInterval):
		}
		vs, err := ac.GetValidationState(ctx, res.ValidationStateID)
		if err != nil {
			if validationReadFinal(err) {
				res.ValidationReadError = err
				return
			}
			logf("DEBUG", "validation state %s: %v", res.ValidationStateID, err)
			continue
		}
		if vs.State == ValidationSucceeded || vs.State == ValidationFailed {
			res.ValidationState = vs.State
			res.ValidationErrors = vs.ErrorDetails
			return
		}
	}
}

// validationReadFinal reports whether a failed validation state read would
// fail the same way on the next poll.
func validationReadFinal(err error) bool {
	if errors.Is(err, ErrNotFound) {
		return true
	}
	var ae *APIError
	return errors.As(err, &ae) && ae.Status >= 400 && ae.Status < 500 && ae.Status != 408 && ae.Status != 429
}
