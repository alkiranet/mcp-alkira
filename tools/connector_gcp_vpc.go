package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ConnectorGcpVpcGetAll() mcp.Tool {
	return mcp.NewTool("connector_gcp_vpc_get_all",
		mcp.WithDescription("Get all Google Cloud VPC connectors"),
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

func ConnectorGcpVpcGetById() mcp.Tool {
	return mcp.NewTool("connector_gcp_vpc_get_by_id",
		mcp.WithDescription("Get details of one Google Cloud VPC connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func ConnectorGcpVpcGetByName() mcp.Tool {
	return mcp.NewTool("connector_gcp_vpc_get_by_name",
		mcp.WithDescription("Get details of one Google Cloud VPC connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}

func ConnectorGcpVpcGetTotal() mcp.Tool {
	return mcp.NewTool("connector_gcp_vpc_get_total",
		mcp.WithDescription("Get total number of Google Cloud VPC connector."),
	)
}
