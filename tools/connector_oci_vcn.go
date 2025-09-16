package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ConnectorOciVcnGetAll() mcp.Tool {
	return mcp.NewTool("connector_oci_vcn_get_all",
		mcp.WithDescription("Get all Oracle Cloud Infrastructure VCN connectors"),
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

func ConnectorOciVcnGetById() mcp.Tool {
	return mcp.NewTool("connector_oci_vcn_get_by_id",
		mcp.WithDescription("Get details of one Oracle Cloud Infrastructure VCN connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func ConnectorOciVcnGetByName() mcp.Tool {
	return mcp.NewTool("connector_oci_vcn_get_by_name",
		mcp.WithDescription("Get details of one Oracle Cloud Infrastructure VCN connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}

func ConnectorOciVcnGetTotal() mcp.Tool {
	return mcp.NewTool("connector_oci_vcn_get_total",
		mcp.WithDescription("Get total numbers of Oracle Cloud Infrastructure VCN connectors."),
	)
}