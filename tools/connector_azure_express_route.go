package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ConnectorAzureExpressRouteGetAll() mcp.Tool {
	return mcp.NewTool("connector_azure_express_route_get_all",
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

func ConnectorAzureExpressRouteGetById() mcp.Tool {
	return mcp.NewTool("connector_azure_express_route_get_by_id",
		mcp.WithDescription("Get details of one Azure ExpressRoute connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func ConnectorAzureExpressRouteGetByName() mcp.Tool {
	return mcp.NewTool("connector_azure_express_route_get_by_name",
		mcp.WithDescription("Get details of one Azure ExpressRoute connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}

func ConnectorAzureExpressRouteGetTotal() mcp.Tool {
	return mcp.NewTool("connector_azure_express_route_get_total",
		mcp.WithDescription("Get total numbers of Azure ExpressRoute connectors."),
	)
}