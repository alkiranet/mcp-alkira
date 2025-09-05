package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAwsTgwConnectors() mcp.Tool {
	return mcp.NewTool("getAwsTgwConnectors",
		mcp.WithDescription("Get all AWS Transit Gateway connectors"),
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

func GetAwsTgwConnectorById() mcp.Tool {
	return mcp.NewTool("getAwsTgwConnectorById",
		mcp.WithDescription("Get details of one AWS Transit Gateway connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func GetAwsTgwConnectorByName() mcp.Tool {
	return mcp.NewTool("getAwsTgwConnectorByName",
		mcp.WithDescription("Get details of one AWS Transit Gateway connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}