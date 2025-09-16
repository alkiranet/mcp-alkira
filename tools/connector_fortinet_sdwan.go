package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ConnectorFortinetSdwanGetAll() mcp.Tool {
	return mcp.NewTool("connector_fortinet_sdwan_get_all",
		mcp.WithDescription("Get all Fortinet SD-WAN connectors"),
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

func ConnectorFortinetSdwanGetById() mcp.Tool {
	return mcp.NewTool("connector_fortinet_sdwan_get_by_id",
		mcp.WithDescription("Get details of one Fortinet SD-WAN connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func ConnectorFortinetSdwanGetByName() mcp.Tool {
	return mcp.NewTool("connector_fortinet_sdwan_get_by_name",
		mcp.WithDescription("Get details of one Fortinet SD-WAN connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}

func ConnectorFortinetSdwanGetTotal() mcp.Tool {
	return mcp.NewTool("connector_fortinet_sdwan_get_total",
		mcp.WithDescription("Get total numbers of Fortinet SD-WAN connectors."),
	)
}