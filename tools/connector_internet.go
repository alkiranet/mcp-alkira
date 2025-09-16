package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ConnectorInternetGetAll() mcp.Tool {
	return mcp.NewTool("connector_internet_get_all",
		mcp.WithDescription("Get all internet connectors"),
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

func ConnectorInternetGetById() mcp.Tool {
	return mcp.NewTool("connector_internet_get_by_id",
		mcp.WithDescription("Get details of one internet connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func ConnectorInternetGetByName() mcp.Tool {
	return mcp.NewTool("connector_internet_get_by_name",
		mcp.WithDescription("Get details of one internet connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}

func ConnectorInternetGetTotal() mcp.Tool {
	return mcp.NewTool("connector_internet_get_total",
		mcp.WithDescription("Get total number of internet connectors."),
	)
}
