package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetFortinetSdwanConnectors() mcp.Tool {
	return mcp.NewTool("getFortinetSdwanConnectors",
		mcp.WithDescription("Get all Fortinet SD-WAN connectors"),
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

func GetFortinetSdwanConnectorById() mcp.Tool {
	return mcp.NewTool("getFortinetSdwanConnectorById",
		mcp.WithDescription("Get details of one Fortinet SD-WAN connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func GetFortinetSdwanConnectorByName() mcp.Tool {
	return mcp.NewTool("getFortinetSdwanConnectorByName",
		mcp.WithDescription("Get details of one Fortinet SD-WAN connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}