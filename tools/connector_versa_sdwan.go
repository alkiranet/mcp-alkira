package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetVersaSdwanConnectors() mcp.Tool {
	return mcp.NewTool("getVersaSdwanConnectors",
		mcp.WithDescription("Get all Versa SD-WAN connectors"),
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

func GetVersaSdwanConnectorById() mcp.Tool {
	return mcp.NewTool("getVersaSdwanConnectorById",
		mcp.WithDescription("Get details of one Versa SD-WAN connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func GetVersaSdwanConnectorByName() mcp.Tool {
	return mcp.NewTool("getVersaSdwanConnectorByName",
		mcp.WithDescription("Get details of one Versa SD-WAN connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}