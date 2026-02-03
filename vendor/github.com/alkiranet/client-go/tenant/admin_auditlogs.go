// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type AuditLogResponse struct {
	Data []AuditLogEntry `json:"data"`
}

type AuditLogEntry struct {
	CreatedAt   int64  `json:"createdAt"`
	Description string `json:"description"`
	ID          string `json:"id"`
	Initiator   string `json:"initiator"`
	IPAddress   string `json:"ipAddress"`
	Status      string `json:"status"`
	Type        string `json:"type"`
}

// GetAuditLogs get audit logs with optional filter with log status or
// log type
func (ac *AlkiraClient) GetAuditLogs(auditStatus string, auditType string, offset string, limit string) (string, error) {

	baseUri := fmt.Sprintf("%s/api/auditlogs?paginated=true", ac.URI)

	uri, err := url.Parse(baseUri)

	if err != nil {
		return "", fmt.Errorf("GetAuditLogs: failed to parse URI: %v", err)
	}

	// Process optional query parameters
	q := uri.Query()

	if auditStatus != "" {
		q.Add("status", auditStatus)
	}
	if auditType != "" {
		q.Add("type", auditType)
	}
	if offset != "" {
		q.Add("offset", offset)
	}
	if limit != "" {
		q.Add("limit", limit)
	}

	// GET
	uri.RawQuery = q.Encode()
	data, err := ac.Get(uri.String())

	return string(data), err
}

// GetAuditLogSummary get audit logs with optional filter with log status or
// log type
func (ac *AlkiraClient) GetAuditLogSummary(auditStatus string, auditType string, offset string, limit string) (string, error) {

	baseUri := fmt.Sprintf("%s/api/auditlogs?paginated=true", ac.URI)

	uri, err := url.Parse(baseUri)

	if err != nil {
		return "", fmt.Errorf("GetAuditLogSummary: failed to parse URI: %v", err)
	}

	// Process optional query parameters
	q := uri.Query()

	if auditStatus != "" {
		q.Add("status", auditStatus)
	}
	if auditType != "" {
		q.Add("type", auditType)
	}
	if offset != "" {
		q.Add("offset", offset)
	}
	if limit != "" {
		q.Add("limit", limit)
	}

	// GET
	uri.RawQuery = q.Encode()
	data, err := ac.Get(uri.String())

	var result DataWithPagination[AuditLogEntry]

	err = json.Unmarshal([]byte(data), &result)

	if err != nil {
		return "", fmt.Errorf("GetAuditLogSummary: failed to unmarshal: %v", err)
	}

	// Marshal the summary data
	summary, err := json.Marshal(result)

	if err != nil {
		return "", fmt.Errorf("GetAuditLogSummary: failed to marshal: %v", err)
	}

	return string(summary), err
}
