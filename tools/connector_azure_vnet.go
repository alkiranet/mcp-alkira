package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAzureVnetConnectors() mcp.Tool {
	return mcp.NewTool("getAzureVnetConnectors",
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

func GetAzureVnetConnectorById() mcp.Tool {
	return mcp.NewTool("getAzureVnetConnectorById",
		mcp.WithDescription("Get details of one Azure VNET connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func GetAzureVnetConnectorByName() mcp.Tool {
	return mcp.NewTool("getAzureVnetConnectorByName",
		mcp.WithDescription("Get details of one Azure Vnet connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}
