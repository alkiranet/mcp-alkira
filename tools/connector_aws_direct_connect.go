package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ConnectorAwsDirectConnectGetAll() mcp.Tool {
	return mcp.NewTool("connector_aws_direct_connect_get_all",
		mcp.WithDescription("Get all AWS Direct Connect connectors"),
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

func ConnectorAwsDirectConnectGetById() mcp.Tool {
	return mcp.NewTool("connector_aws_direct_connect_get_by_id",
		mcp.WithDescription("Get details of one AWS Direct Connect connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func ConnectorAwsDirectConnectGetByName() mcp.Tool {
	return mcp.NewTool("connector_aws_direct_connect_get_by_name",
		mcp.WithDescription("Get details of one AWS Direct Connect connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}

func ConnectorAwsDirectConnectGetTotal() mcp.Tool {
	return mcp.NewTool("connector_aws_direct_connect_get_total",
		mcp.WithDescription("Get total numbers of AWS Direct Connect connectors."),
	)
}