package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetIPSecConnectors() mcp.Tool {
	return mcp.NewTool("getIPSecConnectors",
		mcp.WithDescription("Get all IPSec connectors. By default, 50 " +
			"connectors will be returned once."),
		mcp.WithString("offset",
			mcp.Description("Pagination offset"),
			mcp.DefaultString("0"),
		),
		mcp.WithString("limit",
			mcp.Description("Pagination limit"),
			mcp.DefaultString("50"),
		),
	)
}

func GetIPSecConnectorById() mcp.Tool {
	return mcp.NewTool("getIPSecConnectorById",
		mcp.WithDescription("Get details of one IPSec connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func GetIPSecConnectorByName() mcp.Tool {
	return mcp.NewTool("getIPSecConnectorByName",
		mcp.WithDescription("Get details of one IPSec connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}

func GetIPSecAdvConnectors() mcp.Tool {
	return mcp.NewTool("getIPSecAdvConnectors",
		mcp.WithDescription("Get all Advanced IPSec connectors."),
		mcp.WithString("offset",
			mcp.Description("Pagination offset"),
			mcp.DefaultString("0"),
		),
		mcp.WithString("limit",
			mcp.Description("Pagination limit"),
			mcp.DefaultString("50"),
		),
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
		mcp.WithDescription("Get all IPSec Tunnel Profiles"),
		mcp.WithString("offset",
			mcp.Description("Pagination offset"),
			mcp.DefaultString("0"),
		),
		mcp.WithString("limit",
			mcp.Description("Pagination limit"),
			mcp.DefaultString("50"),
		),
	)
}
