package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetVmwareSdwanConnectors() mcp.Tool {
	return mcp.NewTool("getVmwareSdwanConnectors",
		mcp.WithDescription("Get all VMware SD-WAN (VeloCloud) connectors"),
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

func GetVmwareSdwanConnectorById() mcp.Tool {
	return mcp.NewTool("getVmwareSdwanConnectorById",
		mcp.WithDescription("Get details of one VMware SD-WAN (VeloCloud) connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func GetVmwareSdwanConnectorByName() mcp.Tool {
	return mcp.NewTool("getVmwareSdwanConnectorByName",
		mcp.WithDescription("Get details of one VMware SD-WAN (VeloCloud) connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}