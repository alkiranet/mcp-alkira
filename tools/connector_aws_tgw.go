package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ConnectorAwsTgwGetAll() mcp.Tool {
	return mcp.NewTool("connector_aws_tgw_get_all",
		mcp.WithDescription("Get all AWS Transit Gateway connectors"),
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

func ConnectorAwsTgwGetById() mcp.Tool {
	return mcp.NewTool("connector_aws_tgw_get_by_id",
		mcp.WithDescription("Get details of one AWS Transit Gateway connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func ConnectorAwsTgwGetByName() mcp.Tool {
	return mcp.NewTool("connector_aws_tgw_get_by_name",
		mcp.WithDescription("Get details of one AWS Transit Gateway connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}

func ConnectorAwsTgwGetTotal() mcp.Tool {
	return mcp.NewTool("connector_aws_tgw_get_total",
		mcp.WithDescription("Get total numbers of AWS Transit Gateway connectors."),
	)
}