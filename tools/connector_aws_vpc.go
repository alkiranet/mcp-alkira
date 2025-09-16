package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ConnectorAwsVpcGetAll() mcp.Tool {
	return mcp.NewTool("connector_aws_vpc_get_all",
		mcp.WithDescription("Get all AWS VPC connectors."),
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

func ConnectorAwsVpcGetById() mcp.Tool {
	return mcp.NewTool("connector_aws_vpc_get_by_id",
		mcp.WithDescription("Get details of one AWS VPC connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func ConnectorAwsVpcGetByName() mcp.Tool {
	return mcp.NewTool("connector_aws_vpc_get_by_name",
		mcp.WithDescription("Get details of one AWS VPC connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}

func ConnectorAwsVpcGetTotal() mcp.Tool {
	return mcp.NewTool("connector_aws_vpc_get_total",
		mcp.WithDescription("Get total numbers of AWS VPC connectors."),
	)
}
