package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func TenantResourceUsages() mcp.Tool {
	return mcp.NewTool("tenant_resource_usages",
		mcp.WithDescription("Get resource usages with optional filter 'category' or 'type', or 'scope'"),
		mcp.WithString("category",
			mcp.Description("Resource category"),
		),
		mcp.WithString("type",
			mcp.Description("Resource type"),
		),
		mcp.WithString("scope",
			mcp.Description("Resource scope"),
		),
	)
}

func TenantResourceLimits() mcp.Tool {
	return mcp.NewTool("tenant_resource_limits",
		mcp.WithDescription("Get limits of all resources"),
	)
}
