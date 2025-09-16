package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ConnectorCiscoSdwanGetAll() mcp.Tool {
	return mcp.NewTool("connector_cisco_sdwan_get_all",
		mcp.WithDescription("Get all Cisco SD-WAN connectors"),
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

func ConnectorCiscoSdwanGetById() mcp.Tool {
	return mcp.NewTool("connector_cisco_sdwan_get_by_id",
		mcp.WithDescription("Get details of one Cisco SD-WAN connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func ConnectorCiscoSdwanGetByName() mcp.Tool {
	return mcp.NewTool("connector_cisco_sdwan_get_by_name",
		mcp.WithDescription("Get details of one Cisco SD-WAN connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}

func ConnectorCiscoSdwanGetTotal() mcp.Tool {
	return mcp.NewTool("connector_cisco_sdwan_get_total",
		mcp.WithDescription("Get total numbers of Cisco SD-WAN connectors."),
	)
}