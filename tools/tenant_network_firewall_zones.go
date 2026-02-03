package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func TenantNetworkFirewallZones() mcp.Tool {
	return mcp.NewTool("tenant_network_firewall_zones",
		mcp.WithDescription("Get all fireall zones of the tenant network."),
	)
}
