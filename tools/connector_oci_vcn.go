package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetOciVcnConnectors() mcp.Tool {
	return mcp.NewTool("getOciVcnConnectors",
		mcp.WithDescription("Get all Oracle Cloud Infrastructure VCN connectors"),
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

func GetOciVcnConnectorById() mcp.Tool {
	return mcp.NewTool("getOciVcnConnectorById",
		mcp.WithDescription("Get details of one Oracle Cloud Infrastructure VCN connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func GetOciVcnConnectorByName() mcp.Tool {
	return mcp.NewTool("getOciVcnConnectorByName",
		mcp.WithDescription("Get details of one Oracle Cloud Infrastructure VCN connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}