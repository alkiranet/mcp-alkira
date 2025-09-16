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
	)
}

func AuditLogs() mcp.Tool {
	return mcp.NewTool("notification_auditLogs",
		mcp.WithDescription("Get all audit logs with optional filter: 'status' or 'type'"),
		mcp.WithString("type",
			mcp.Description("Audit Log Type, which defines the type of the audit log."),
			mcp.Enum("USER_SETTINGS", "USER_ACCESS", "TENANT_SETTINGS",
				"NETWORK_PROVISIONING", "NETWORK_CONFIGURATION", "OPERATIONS",
				"MISCELLANEOUS"),
		),
		mcp.WithString("status",
			mcp.Description("Audit Log Status, like: 'FAILED', 'SUCCEEDED'"),
		),
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
	)
}
