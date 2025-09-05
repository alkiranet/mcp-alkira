package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetCiscoSdwanConnectors() mcp.Tool {
	return mcp.NewTool("getCiscoSdwanConnectors",
		mcp.WithDescription("Get all Cisco SD-WAN connectors"),
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

func GetCiscoSdwanConnectorById() mcp.Tool {
	return mcp.NewTool("getCiscoSdwanConnectorById",
		mcp.WithDescription("Get details of one Cisco SD-WAN connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func GetCiscoSdwanConnectorByName() mcp.Tool {
	return mcp.NewTool("getCiscoSdwanConnectorByName",
		mcp.WithDescription("Get details of one Cisco SD-WAN connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}