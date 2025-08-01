package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAlerts() mcp.Tool {
	return mcp.NewTool("getAlerts",
		mcp.WithDescription("Get all alerts with optional filter 'status' or 'type', or 'prioirty'"),
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

func GetAuditLogs() mcp.Tool {
	return mcp.NewTool("getAuditLogs",
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

func GetJobs() mcp.Tool {
	return mcp.NewTool("getJobs",
		mcp.WithDescription("Get all jobs with optional filter: 'status' or 'type'"),
        mcp.WithString("type",
            mcp.Description("Job Type, which defines the type of the job."),
			mcp.Enum("USER_SETTINGS", "USER_ACCESS", "TENANT_SETTINGS",
				"NETWORK_PROVISIONING", "NETWORK_CONFIGURATION", "OPERATIONS",
				"MISCELLANEOUS"),
        ),
		mcp.WithString("status",
            mcp.Description("Job Status, which describe the status of the job."),
			mcp.Enum("PENDING", "IN_PROGRESS", "SUCCESS", "FAILED"),
        ),
    )
}

func GetAllHealth() mcp.Tool {
	return mcp.NewTool("getAllHealth",
		mcp.WithDescription("Get health status of all resources"),
    )
}

func GetHealthOfConnector() mcp.Tool {
	return mcp.NewTool("getHealthOfConnector",
		mcp.WithDescription("Get health status of a connector by ID"),
		mcp.WithString("id",
			mcp.Required(),
            mcp.Description("Connector ID."),
        ),
    )
}

func GetHealthOfConnectorInstance() mcp.Tool {
	return mcp.NewTool("getHealthOfConnectorInstance",
		mcp.WithDescription("Get health status of a connector instance by ID"),
		mcp.WithString("id",
			mcp.Required(),
            mcp.Description("Connector ID."),
        ),
		mcp.WithString("instanceId",
			mcp.Required(),
            mcp.Description("Instance ID."),
        ),
    )
}

func GetHealthOfService() mcp.Tool {
	return mcp.NewTool("getHealthOfService",
		mcp.WithDescription("Get health status of a service by ID"),
		mcp.WithString("id",
			mcp.Required(),
            mcp.Description("Service ID."),
        ),
    )
}

func GetHealthOfServiceInstance() mcp.Tool {
	return mcp.NewTool("getHealthOfServiceInstance",
		mcp.WithDescription("Get health status of a service instance by ID"),
		mcp.WithString("id",
			mcp.Required(),
            mcp.Description("Service ID."),
        ),
		mcp.WithString("instanceId",
			mcp.Required(),
            mcp.Description("Instance ID."),
        ),
    )
}

