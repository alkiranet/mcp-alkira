package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func Alerts() mcp.Tool {
	return mcp.NewTool("notification_alerts",
		mcp.WithDescription("Get all alerts with optional filter 'status' or 'type', or 'priority'"),
		mcp.WithString("type",
			mcp.Description("Alert Type, like: 'NETWORK_STATUS'"),
			mcp.Enum("NETWORK_STATUS", "NETWORK_PROVISIONING", "OPERATIONS",
				"MISCELLANEOUS", "PREFIX_OVERLAP"),
		),
		mcp.WithString("status",
			mcp.Description("Alert Status, like: 'ACTIVE', 'RESOLVED'"),
		),
		mcp.WithString("priority",
			mcp.Description("Alert Priority"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func AuditLogs() mcp.Tool {
	return mcp.NewTool("notification_auditLogs",
		mcp.WithDescription("Get one page of audit logs, newest first by default, with optional filters: 'status', 'type', 'startTime', 'endTime'. Page with 'offset' and 'limit'; offset + limit must be at most 10000. The response 'pagination.hits' is the match count, capped at 10000; to reach older events, narrow with 'startTime' and 'endTime' instead of paging further. To walk several pages, pass the same 'endTime' on every call, such as the current time or the newest 'createdAt' on the first page; without it, new events shift the pages and events repeat or go missing. 'createdAt' is whole seconds and events in the same second have no fixed order, so drop duplicates by 'id'"),
		mcp.WithString("type",
			mcp.Description("Audit Log Type, which defines the type of the audit log."),
			mcp.Enum("USER_SETTINGS", "USER_ACCESS", "TENANT_SETTINGS",
				"NETWORK_PROVISIONING", "NETWORK_CONFIGURATION", "OPERATIONS",
				"MISCELLANEOUS"),
		),
		mcp.WithString("status",
			mcp.Description("Audit Log Status, like: 'FAILED', 'SUCCEEDED'"),
		),
		mcp.WithString("offset",
			mcp.Description("Pagination offset"),
			mcp.DefaultString("0"),
		),
		mcp.WithString("limit",
			mcp.Description("Pagination limit, from 1 to 10000"),
			mcp.DefaultString("50"),
		),
		mcp.WithString("sortBy",
			mcp.Description("Sort order as 'field:direction', several joined with '|'. Fields (case-sensitive): createdAt, type, initiator, ipAddress. Directions: asc, desc; desc when omitted"),
			mcp.DefaultString("createdAt:desc"),
		),
		mcp.WithString("startTime",
			mcp.Description("Only audit logs created at or after this time, in Unix seconds (not milliseconds)"),
		),
		mcp.WithString("endTime",
			mcp.Description("Only audit logs created at or before this time, in Unix seconds (not milliseconds); not before startTime. Keep it fixed while paging"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func Jobs() mcp.Tool {
	return mcp.NewTool("notification_jobs",
		mcp.WithDescription("Get all jobs with optional filter: 'status' or 'type'"),
		mcp.WithString("type",
			mcp.Description("Job Type, which defines the type of the job."),
			mcp.Enum("USER_SETTINGS", "USER_ACCESS", "TENANT_SETTINGS",
				"NETWORK_PROVISIONING", "NETWORK_CONFIGURATION", "OPERATIONS",
				"MISCELLANEOUS"),
		),
		mcp.WithString("status",
			mcp.Description("Job Status, which describes the status of the job."),
			mcp.Enum("PENDING", "IN_PROGRESS", "SUCCESS", "FAILED"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}
