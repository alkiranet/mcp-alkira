package handlers

import (
	"context"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/mark3labs/mcp-go/mcp"
)

func Alerts(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		s, _ := request.RequireString("status")
		t, _ := request.RequireString("type")
		p, _ := request.RequireString("priority")

		offset, limit, err := pagingArgs(request, "", "")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get resources
		alerts, err := client.GetAlertsSummary(s, t, p, offset, limit)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(alerts), nil
	}
}

func AuditLogs(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		filter := ak.AuditLogFilter{
			Status: request.GetString("status", ""),
			Type:   request.GetString("type", ""),
		}

		var err error
		if filter.SortBy, err = auditLogSortBy(request.GetString("sortBy", "")); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if filter.Offset, filter.Limit, err = pagingArgs(request, "0", "50"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if filter.StartTime, err = unixSecondsArg(request, "startTime"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if filter.EndTime, err = unixSecondsArg(request, "endTime"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if err := auditLogWindow(filter); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get resources
		auditLogs, err := client.GetAuditLogSummaryFiltered(filter)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(auditLogs), nil
	}
}

// auditLogSortFields are the audit log fields the notification service sorts
// on. It drops any other field without an error, which leaves the page unordered.
var auditLogSortFields = []string{"createdAt", "type", "initiator", "ipAddress"}

// auditLogSortBy checks a "field:direction" list joined with "|" against
// auditLogSortFields. Empty means newest first, since without a sort the API
// does not order by time and recent events can sit past the first page.
func auditLogSortBy(sortBy string) (string, error) {
	if sortBy == "" {
		return "createdAt:desc", nil
	}
	for _, term := range strings.Split(sortBy, "|") {
		field, dir, hasDir := strings.Cut(term, ":")
		if !slices.Contains(auditLogSortFields, field) {
			return "", fmt.Errorf("sortBy field %q is not supported; use one of %s (case-sensitive)", field, strings.Join(auditLogSortFields, ", "))
		}
		if hasDir && dir != "asc" && dir != "desc" {
			return "", fmt.Errorf("sortBy direction %q for %s must be asc or desc", dir, field)
		}
	}
	return sortBy, nil
}

// maxUnixSeconds is year 5138 in Unix seconds. Larger values are taken to be
// Unix milliseconds, which the API would read as a far-future second.
const maxUnixSeconds = 1e11

// unixSecondsArg reads a time in Unix seconds given as a JSON number or a
// decimal string. It returns "" when the argument is absent or empty.
func unixSecondsArg(request mcp.CallToolRequest, name string) (string, error) {
	var n int64
	switch v := request.GetArguments()[name].(type) {
	case nil:
		return "", nil
	case float64:
		if v != math.Trunc(v) || v < 0 || v > maxUnixSeconds {
			return "", unixSecondsError(name, strconv.FormatFloat(v, 'f', -1, 64))
		}
		n = int64(v)
	case string:
		if v == "" {
			return "", nil
		}
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil || parsed < 0 || parsed > maxUnixSeconds {
			return "", unixSecondsError(name, strconv.Quote(v))
		}
		n = parsed
	default:
		return "", unixSecondsError(name, fmt.Sprintf("%#v", v))
	}
	return strconv.FormatInt(n, 10), nil
}

func unixSecondsError(name, got string) error {
	return fmt.Errorf("%s must be a whole number of Unix seconds from 0 to %d, got %s; divide a millisecond timestamp by 1000", name, int64(maxUnixSeconds), got)
}

// auditLogMaxWindow is the deepest result the audit log search serves: offset + limit.
const auditLogMaxWindow = 10000

// auditLogWindow refuses a page the audit log search cannot serve and a time
// window that ends before it starts. Offset and times are already non-negative.
func auditLogWindow(f ak.AuditLogFilter) error {
	offset, _ := strconv.Atoi(f.Offset)
	limit, _ := strconv.Atoi(f.Limit)
	if limit < 1 {
		return fmt.Errorf("limit must be at least 1, got %d", limit)
	}
	if offset+limit > auditLogMaxWindow {
		return fmt.Errorf("offset + limit must be at most %d, got %d; narrow startTime and endTime to reach older events", auditLogMaxWindow, offset+limit)
	}
	if f.StartTime != "" && f.EndTime != "" {
		start, _ := strconv.Atoi(f.StartTime)
		end, _ := strconv.Atoi(f.EndTime)
		if start > end {
			return fmt.Errorf("startTime %d is after endTime %d", start, end)
		}
	}
	return nil
}

func Jobs(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		s, err := request.RequireString("status")
		if err != nil {
			s = ""
		}

		t, err := request.RequireString("type")
		if err != nil {
			t = ""
		}

		offset, limit, err := pagingArgs(request, "", "")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get resources
		jobs, err := client.GetJobs(s, t, offset, limit)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(jobs), nil
	}
}
