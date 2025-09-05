package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetRemoteAccessConnectors() mcp.Tool {
	return mcp.NewTool("getRemoteAccessConnectors",
		mcp.WithDescription("Get all Remote Access Template connectors"),
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

func GetRemoteAccessConnectorById() mcp.Tool {
	return mcp.NewTool("getRemoteAccessConnectorById",
		mcp.WithDescription("Get details of one Remote Access connector by " +
			"its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func GetRemoteAccessConnectorByName() mcp.Tool {
	return mcp.NewTool("getRemoteAccessConnectorByName",
		mcp.WithDescription("Get details of one Remote Access connector by " +
			"its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}
