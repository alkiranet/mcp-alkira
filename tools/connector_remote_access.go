package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ConnectorRemoteAccessGetAll() mcp.Tool {
	return mcp.NewTool("connector_remote_access_get_all",
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

func ConnectorRemoteAccessGetById() mcp.Tool {
	return mcp.NewTool("connector_remote_access_get_by_id",
		mcp.WithDescription("Get details of one Remote Access connector by " +
			"its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func ConnectorRemoteAccessGetByName() mcp.Tool {
	return mcp.NewTool("connector_remote_access_get_by_name",
		mcp.WithDescription("Get details of one Remote Access connector by " +
			"its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}

func ConnectorRemoteAccessGetTotal() mcp.Tool {
	return mcp.NewTool("connector_remote_access_get_total",
		mcp.WithDescription("Get total numbers of Remote Access connectors."),
	)
}
