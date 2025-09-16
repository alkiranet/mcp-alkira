package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ConnectorVmwareSdwanGetAll() mcp.Tool {
	return mcp.NewTool("connector_vmware_sdwan_get_all",
		mcp.WithDescription("Get all VMware SD-WAN (VeloCloud) connectors"),
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

func ConnectorVmwareSdwanGetById() mcp.Tool {
	return mcp.NewTool("connector_vmware_sdwan_get_by_id",
		mcp.WithDescription("Get details of one VMware SD-WAN (VeloCloud) connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
	)
}

func ConnectorVmwareSdwanGetByName() mcp.Tool {
	return mcp.NewTool("connector_vmware_sdwan_get_by_name",
		mcp.WithDescription("Get details of one VMware SD-WAN (VeloCloud) connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
	)
}

func ConnectorVmwareSdwanGetTotal() mcp.Tool {
	return mcp.NewTool("connector_vmware_sdwan_get_total",
		mcp.WithDescription("Get total numbers of VMware SD-WAN (VeloCloud) connectors."),
	)
}