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

// AuditLogFilter selects, orders and pages audit logs. Empty fields are left
// off the request so the API applies its own default.
type AuditLogFilter struct {
	Status string
	Type   string
	Offset string
	Limit  string

	// SortBy is "field:direction", joined with "|" for several fields, for
	// example "createdAt:desc". Fields: type, createdAt, initiator, ipAddress.
	SortBy string

	// StartTime and EndTime bound createdAt, inclusive, in Unix seconds.
	StartTime string
	EndTime   string
}

// auditLogsURI builds the audit log list URI for filter f.
func (ac *AlkiraClient) auditLogsURI(f AuditLogFilter) (string, error) {

	uri, err := url.Parse(fmt.Sprintf("%s/api/auditlogs?paginated=true", ac.URI))

	if err != nil {
		return "", fmt.Errorf("failed to parse URI: %v", err)
	}

	q := uri.Query()

	for _, p := range []struct{ key, value string }{
		{"status", f.Status},
		{"type", f.Type},
		{"offset", f.Offset},
		{"limit", f.Limit},
		{"sortBy", f.SortBy},
		{"startTime", f.StartTime},
		{"endTime", f.EndTime},
	} {
		if p.value != "" {
			q.Add(p.key, p.value)
		}
	}

	uri.RawQuery = q.Encode()
	return uri.String(), nil
}

// GetAuditLogs get audit logs with optional filter with log status or
// log type
func (ac *AlkiraClient) GetAuditLogs(auditStatus string, auditType string, offset string, limit string) (string, error) {

	uri, err := ac.auditLogsURI(AuditLogFilter{Status: auditStatus, Type: auditType, Offset: offset, Limit: limit})

	if err != nil {
		return "", fmt.Errorf("GetAuditLogs: %v", err)
	}

	data, err := ac.Get(uri)

	return string(data), err
}

// GetAuditLogSummary get audit logs with optional filter with log status or
// log type
func (ac *AlkiraClient) GetAuditLogSummary(auditStatus string, auditType string, offset string, limit string) (string, error) {
	return ac.GetAuditLogSummaryFiltered(AuditLogFilter{Status: auditStatus, Type: auditType, Offset: offset, Limit: limit})
}

// GetAuditLogSummaryFiltered gets one page of audit logs selected and ordered
// by filter f.
func (ac *AlkiraClient) GetAuditLogSummaryFiltered(f AuditLogFilter) (string, error) {

	uri, err := ac.auditLogsURI(f)

	if err != nil {
		return "", fmt.Errorf("GetAuditLogSummary: %v", err)
	}

	data, err := ac.Get(uri)

	if err != nil {
		return "", fmt.Errorf("GetAuditLogSummary: %v", err)
	}

	var result DataWithPagination[AuditLogEntry]

	err = json.Unmarshal(data, &result)

	if err != nil {
		return "", fmt.Errorf("GetAuditLogSummary: failed to unmarshal: %v", err)
	}

	// Marshal the summary data
	summary, err := json.Marshal(result)

	if err != nil {
		return "", fmt.Errorf("GetAuditLogSummary: failed to marshal: %v", err)
	}

	return string(summary), nil
}
