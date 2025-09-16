package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ConnectorAzureVnetGetAll() mcp.Tool {
	return mcp.NewTool("connector_azure_vnet_get_all",
		mcp.WithDescription("Get all Azure VNET connectors."),
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

func ConnectorAzureVnetGetById() mcp.Tool {
	return mcp.NewTool("connector_azure_vnet_get_by_id",
		mcp.WithDescription("Get details of one Azure VNET connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func ConnectorAzureVnetGetByName() mcp.Tool {
	return mcp.NewTool("connector_azure_vnet_get_by_name",
		mcp.WithDescription("Get details of one Azure Vnet connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}

func ConnectorAzureVnetGetTotal() mcp.Tool {
	return mcp.NewTool("connector_azure_vnet_get_total",
		mcp.WithDescription("Get total numbers of Azure VNET connectors."),
	)
}
