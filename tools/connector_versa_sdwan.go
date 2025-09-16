package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ConnectorVersaSdwanGetAll() mcp.Tool {
	return mcp.NewTool("connector_versa_sdwan_get_all",
		mcp.WithDescription("Get all Versa SD-WAN connectors"),
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

func ConnectorVersaSdwanGetById() mcp.Tool {
	return mcp.NewTool("connector_versa_sdwan_get_by_id",
		mcp.WithDescription("Get details of one Versa SD-WAN connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func ConnectorVersaSdwanGetByName() mcp.Tool {
	return mcp.NewTool("connector_versa_sdwan_get_by_name",
		mcp.WithDescription("Get details of one Versa SD-WAN connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}

func ConnectorVersaSdwanGetTotal() mcp.Tool {
	return mcp.NewTool("connector_versa_sdwan_get_total",
		mcp.WithDescription("Get total numbers of Versa SD-WAN connectors."),
	)
}