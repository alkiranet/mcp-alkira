package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAwsVpcConnectors() mcp.Tool {
	return mcp.NewTool("getAwsVpcConnectors",
		mcp.WithDescription("Get all AWS VPC connectors."),
	)
}

func GetAwsVpcConnectorById() mcp.Tool {
	return mcp.NewTool("getAwsVpcConnectorById",
		mcp.WithDescription("Get details of one AWS VPC connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func GetAwsVpcConnectorByName() mcp.Tool {
	return mcp.NewTool("getAwsVpcConnectorByName",
		mcp.WithDescription("Get details of one AWS VPC connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}
