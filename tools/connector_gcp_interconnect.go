package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetGcpInterconnectConnectors() mcp.Tool {
	return mcp.NewTool("getGcpInterconnectConnectors",
		mcp.WithDescription("Get all Google Cloud Interconnect connectors"),
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

func GetGcpInterconnectConnectorById() mcp.Tool {
	return mcp.NewTool("getGcpInterconnectConnectorById",
		mcp.WithDescription("Get details of one Google Cloud Interconnect " +
			"connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func GetGcpInterconnectConnectorByName() mcp.Tool {
	return mcp.NewTool("getGcpInterconnectConnectorByName",
		mcp.WithDescription("Get details of one Google Cloud Interconnect " +
			"connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}
