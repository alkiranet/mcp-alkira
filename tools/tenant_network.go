package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetTenantNetworkSummary() mcp.Tool {
	return mcp.NewTool("getTenantNetworkSummary",
		mcp.WithDescription("Get tenant network summary of the tenant that "+
			"contains some basic infos of the tenant network, especially the "+
			"number of connectors and services of the tenant network."),
	)
}
