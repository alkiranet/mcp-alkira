package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAzureExpressRouteConnectors() mcp.Tool {
	return mcp.NewTool("getAzureExpressRouteConnectors",
		mcp.WithDescription("Get all Azure ExpressRoute connectors"),
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

func GetAzureExpressRouteConnectorById() mcp.Tool {
	return mcp.NewTool("getAzureExpressRouteConnectorById",
		mcp.WithDescription("Get details of one Azure ExpressRoute connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func GetAzureExpressRouteConnectorByName() mcp.Tool {
	return mcp.NewTool("getAzureExpressRouteConnectorByName",
		mcp.WithDescription("Get details of one Azure ExpressRoute connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}