package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetIPSecConnectors() mcp.Tool {
	return mcp.NewTool("getIPSecConnectors",
		mcp.WithDescription("Get all IPSec connectors. The returned data " +
			"will be with pagination enabled by default."),
		mcp.WithString("offset",
			mcp.Description("Offset of paginated data will be returned."),
		),
		mcp.WithString("limit",
			mcp.Description("Limit of paginated data will be returned. If not " +
				"provided, default value is 10."),
		),
	)
}

func GetIPSecConnectorById() mcp.Tool {
	return mcp.NewTool("getIPSecConnectorById",
		mcp.WithDescription("Get details of one IPSec connectors by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func GetIPSecConnectorByName() mcp.Tool {
	return mcp.NewTool("getIPSecConnectorByName",
		mcp.WithDescription("Get details of one IPSec connectors by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}

func GetIPSecAdvConnectors() mcp.Tool {
	return mcp.NewTool("getIPSecAdvConnectors",
		mcp.WithDescription("Get all Advanced IPSec connectors."),
	)
}

func GetIPSecAdvConnectorById() mcp.Tool {
	return mcp.NewTool("getIPSecAdvConnectorById",
		mcp.WithDescription("Get details of one Advanced IPSec connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func GetIPSecAdvConnectorByName() mcp.Tool {
	return mcp.NewTool("getIPSecAdvConnectorByName",
		mcp.WithDescription("Get details of one Advanced IPSec connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}

func GetIPSecConnectorTunnelProfile() mcp.Tool {
	return mcp.NewTool("getIPSecConnectorTunnelProfile",
		mcp.WithDescription("Get all IPSec Tunnel Profile connectors"),
	)
}
