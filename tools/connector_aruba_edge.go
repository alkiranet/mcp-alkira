package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetArubaEdgeConnectors() mcp.Tool {
	return mcp.NewTool("getArubaEdgeConnectors",
		mcp.WithDescription("Get all Aruba Edge Connect SD-WAN connectors"),
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

func GetArubaEdgeConnectorById() mcp.Tool {
	return mcp.NewTool("getArubaEdgeConnectorById",
		mcp.WithDescription("Get details of one Aruba Edge Connect SD-WAN connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func GetArubaEdgeConnectorByName() mcp.Tool {
	return mcp.NewTool("getArubaEdgeConnectorByName",
		mcp.WithDescription("Get details of one Aruba Edge Connect SD-WAN connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}