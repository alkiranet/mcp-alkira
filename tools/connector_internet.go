package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetInternetConnectors() mcp.Tool {
	return mcp.NewTool("getInternetConnectors",
		mcp.WithDescription("Get all internet connectors"),
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

func GetInternetConnectorById() mcp.Tool {
	return mcp.NewTool("getInternetConnectorById",
		mcp.WithDescription("Get details of one internet connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func GetInternetConnectorByName() mcp.Tool {
	return mcp.NewTool("getInternetConnectorByName",
		mcp.WithDescription("Get details of one internet connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}
