package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetGcpVpcConnectors() mcp.Tool {
	return mcp.NewTool("getGcpVpcConnectors",
		mcp.WithDescription("Get all Google Cloud VPC connectors"),
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

func GetGcpVpcConnectorById() mcp.Tool {
	return mcp.NewTool("getGcpVpcConnectorById",
		mcp.WithDescription("Get details of one Google Cloud VPC connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func GetGcpVpcConnectorByName() mcp.Tool {
	return mcp.NewTool("getGcpVpcConnectorByName",
		mcp.WithDescription("Get details of one Google Cloud VPC connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}