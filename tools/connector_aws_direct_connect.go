package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAwsDirectConnectConnectors() mcp.Tool {
	return mcp.NewTool("getAwsDirectConnectConnectors",
		mcp.WithDescription("Get all AWS Direct Connect connectors"),
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

func GetAwsDirectConnectConnectorById() mcp.Tool {
	return mcp.NewTool("getAwsDirectConnectConnectorById",
		mcp.WithDescription("Get details of one AWS Direct Connect connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func GetAwsDirectConnectConnectorByName() mcp.Tool {
	return mcp.NewTool("getAwsDirectConnectConnectorByName",
		mcp.WithDescription("Get details of one AWS Direct Connect connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}