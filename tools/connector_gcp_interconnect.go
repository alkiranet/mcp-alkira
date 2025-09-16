package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ConnectorGcpInterconnectGetAll() mcp.Tool {
	return mcp.NewTool("connector_gcp_interconnect_get_all",
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

func ConnectorGcpInterconnectGetById() mcp.Tool {
	return mcp.NewTool("connector_gcp_interconnect_get_by_id",
		mcp.WithDescription("Get details of one Google Cloud Interconnect " +
			"connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func ConnectorGcpInterconnectGetByName() mcp.Tool {
	return mcp.NewTool("connector_gcp_interconnect_get_by_name",
		mcp.WithDescription("Get details of one Google Cloud Interconnect " +
			"connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}

func ConnectorGcpInterconnectGetTotal() mcp.Tool {
	return mcp.NewTool("connector_gcp_interconnect_get_total",
		mcp.WithDescription("Get total numbers of Google Cloud Interconnect connectors."),
	)
}
