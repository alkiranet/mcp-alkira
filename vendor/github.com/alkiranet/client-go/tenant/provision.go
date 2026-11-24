// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode/utf16"
)

// Provision request states (api-docs NetworkProvisionedState.yaml:3-9), plus
// INITIALIZING, the state of a single-entity request before it joins a buffer.
const (
	ProvisionPending        = "PENDING"
	ProvisionScheduled      = "SCHEDULED"
	ProvisionInProgress     = "IN_PROGRESS"
	ProvisionSuccess        = "SUCCESS"
	ProvisionFailed         = "FAILED"
	ProvisionPartialSuccess = "PARTIAL_SUCCESS"
	ProvisionInitializing   = "INITIALIZING"
)

// MaxProvisionDescription is the longest provision description TPS accepts,
// in UTF-16 code units (TenantNetworkController.java provisionTenantNetwork:
// 400 when String.length() is above 128).
const MaxProvisionDescription = 128

// ProvisionErrorDetails is the errorDetails block of a provision request
// (TPS exception/ErrorDetails.java).
type ProvisionErrorDetails struct {
	Code     string         `json:"code,omitempty"`
	ID       string         `json:"id,omitempty"`
	Message  string         `json:"message,omitempty"`
	Errors   []FieldError   `json:"errors,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ContactSupport reports metadata.contactSupport=true: a hard failure that
// needs escalation, not a retry.
func (e *ProvisionErrorDetails) ContactSupport() bool {
	if e == nil {
		return false
	}
	v, _ := e.Metadata["contactSupport"].(bool)
	return v
}

// BufferedProvisionRequest is the buffer a provision request joined.
type BufferedProvisionRequest struct {
	ID                       int64  `json:"id"`
	State                    string `json:"state"`
	LastProvisionRequestTime int64  `json:"lastProvisionRequestTime,omitempty"`
}

// ProvisionChangeEntity identifies the entity of one selected change.
type ProvisionChangeEntity struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	SubType string `json:"subType,omitempty"`
}

// ProvisionChange is one element of a selective provision or dry-run
// validation body, as the portal builds it (karrawa 5d18e83fd
// diffV2/utils/dataTransform.js:182-226). TPS matches each element to a
// pending change by changeType, entity type and entity id
// (NetworkChangesCurator.java getCuratedChanges).
type ProvisionChange struct {
	ChangeType        string                `json:"changeType"`
	Entity            ProvisionChangeEntity `json:"entity"`
	SubType           string                `json:"subType,omitempty"`
	BestEffortAllowed bool                  `json:"bestEffortAllowed,omitempty"`
	CleanupType       string                `json:"cleanupType,omitempty"`
}

// ProvisionOptions are the arguments of ProvisionTenantNetworkWith.
type ProvisionOptions struct {
	// Description is sent as ?description= when set; at most 128 characters.
	Description string
	// Changes is the selected set; it must not be empty.
	Changes []ProvisionChange
}

// DescriptionLength is the length TPS measures: UTF-16 code units.
func DescriptionLength(s string) int { return len(utf16.Encode([]rune(s))) }

func (ac *AlkiraClient) networkURI(parts ...string) string {
	return ac.URI + "/tenantnetworks/" + ac.TenantNetworkId + "/" + strings.Join(parts, "/")
}

// ErrNoProvisionRequest: TPS accepted a provision but started no provision
// request, so its answer carries no x-provision-request-id
// (TenantNetworkController.java:524-532); the body is the network, not the
// request.
var ErrNoProvisionRequest = errors.New("provision accepted without a provision request id")

// ProvisionTenantNetworkWith sends POST /tenantnetworks/{id}/provision with
// selectedOnly=true and opts.Changes as the body, and returns the
// x-provision-request-id. Any 2xx carrying the header is success; a 2xx
// without it returns ErrNoProvisionRequest with the WriteResult. It never
// sends force or includeChanges, and never a body-less provision.
func (ac *AlkiraClient) ProvisionTenantNetworkWith(ctx context.Context, opts ProvisionOptions) (string, *WriteResult, error) {
	if len(opts.Changes) == 0 {
		return "", nil, fmt.Errorf("%w: no changes selected", ErrInvalidInput)
	}
	if DescriptionLength(opts.Description) > MaxProvisionDescription {
		return "", nil, fmt.Errorf("%w: description longer than %d characters", ErrInvalidInput, MaxProvisionDescription)
	}
	body, err := json.Marshal(opts.Changes)
	if err != nil {
		return "", nil, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	q := url.Values{"selectedOnly": {"true"}}
	if opts.Description != "" {
		q.Set("description", opts.Description)
	}
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()
	uri := withQuery(ac.networkURI("provision"), q)
	res, err := ac.doWrite(ctx, "POST", uri, body)
	if err != nil {
		return "", nil, err
	}
	if res.ProvisionRequestID == "" {
		return "", res, fmt.Errorf("POST %s (request id %s): %d response: %w", uri, res.RequestID, res.Status, ErrNoProvisionRequest)
	}
	return res.ProvisionRequestID, res, nil
}

// GetProvisionRequest reads one provision request, errorDetails included.
func (ac *AlkiraClient) GetProvisionRequest(ctx context.Context, id string) (*TenantNetworkProvisionRequest, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: empty provision request id", ErrInvalidInput)
	}
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()
	raw, err := ac.doGet(ctx, ac.networkURI("provision-requests", pathSegment(id)))
	if err != nil {
		return nil, err
	}
	var pr TenantNetworkProvisionRequest
	if err := json.Unmarshal(raw, &pr); err != nil {
		return nil, fmt.Errorf("provision request %s: %w", id, err)
	}
	return &pr, nil
}

// GetBufferedProvisionRequest reads the buffered provision request a
// provision request joined.
func (ac *AlkiraClient) GetBufferedProvisionRequest(ctx context.Context, id int64) (*BufferedProvisionRequest, error) {
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()
	raw, err := ac.doGet(ctx, ac.networkURI("buffered-provision-requests", fmt.Sprint(id)))
	if err != nil {
		return nil, err
	}
	var b BufferedProvisionRequest
	if err := json.Unmarshal(raw, &b); err != nil {
		return nil, fmt.Errorf("buffered provision request %d: %w", id, err)
	}
	return &b, nil
}

// ListProvisionRequests reads every provision request, or those in state
// when it is set; TPS binds one state value. The list needs ?paginated=true
// (the unpaginated GET answers 500) and caps a page at 100, so it sends
// limit=100 and pages by offset until pagination.hits. List rows carry no
// errorDetails; GetProvisionRequest reads them.
// TPS lists the newest first (ProvisionRequestServiceImpl.java:930), so a
// request filed while the pages are read shifts later rows by one place: a
// row read twice is returned once, and a request filed after the first page
// is read can be missing.
func (ac *AlkiraClient) ListProvisionRequests(ctx context.Context, state string) ([]TenantNetworkProvisionRequest, error) {
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()
	rows, err := ac.listPagedWith(ctx, ac.networkURI("provision-requests"), PagingTotalHits, stateFilter(state))
	if err != nil {
		return nil, err
	}
	reqs, err := decodeRequests(rows)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	out := reqs[:0]
	for _, r := range reqs {
		if !seen[r.Id] {
			seen[r.Id] = true
			out = append(out, r)
		}
	}
	return out, nil
}

// ListProvisionRequestsPage reads the first page of at most limit (1 to 100)
// provision requests, or those in state when it is set. TPS returns the newest
// first. hits is pagination.hits, the total the filter matches.
func (ac *AlkiraClient) ListProvisionRequestsPage(ctx context.Context, state string, limit int) ([]TenantNetworkProvisionRequest, int, error) {
	if limit < 1 || limit > 100 {
		return nil, 0, fmt.Errorf("%w: limit %d outside 1..100", ErrInvalidInput, limit)
	}
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()
	rows, hits, err := ac.listPage(ctx, ac.networkURI("provision-requests"), 0, limit, stateFilter(state))
	if err != nil {
		return nil, 0, err
	}
	out, err := decodeRequests(rows)
	return out, hits, err
}

func stateFilter(state string) url.Values {
	if state == "" {
		return nil
	}
	return url.Values{"state": {state}}
}

func decodeRequests(rows []json.RawMessage) ([]TenantNetworkProvisionRequest, error) {
	out := make([]TenantNetworkProvisionRequest, len(rows))
	for i, r := range rows {
		if err := json.Unmarshal(r, &out[i]); err != nil {
			return nil, fmt.Errorf("provision requests: row %d: %w", i, err)
		}
	}
	return out, nil
}

// GetLastProvisionState reads GET /tenantnetworks/{id}/last-provision-state.
// A 404 with no body means the network was never provisioned: provisioned is
// false and err is nil. Any other non-2xx, a 404 with a body included, is an error.
func (ac *AlkiraClient) GetLastProvisionState(ctx context.Context) (state string, provisioned bool, err error) {
	ctx, cancel := withWriteDeadline(ctx)
	defer cancel()
	raw, err := ac.doGet(ctx, ac.networkURI("last-provision-state"))
	var ae *APIError
	if errors.As(err, &ae) && ae.Status == 404 && strings.TrimSpace(ae.Body) == "" {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	var s struct {
		State string `json:"state"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(raw), &s); err != nil {
		return "", false, fmt.Errorf("last provision state: %w", err)
	}
	return s.State, true, nil
}
