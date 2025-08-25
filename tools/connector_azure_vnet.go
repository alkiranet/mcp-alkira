package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAzureVnetConnectors() mcp.Tool {
	return mcp.NewTool("getAzureVnetConnectors",
		mcp.WithDescription("Get all Azure VNET connectors."),
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
