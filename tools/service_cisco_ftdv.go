package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ServiceCiscoFTDvGetAll() mcp.Tool {
	return mcp.NewTool("service_cisco_ftdv_get_all",
		mcp.WithDescription("Get all Cisco Firepower Threat Defense virtual " +
			"services"),
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

func ServiceCiscoFTDvGetById() mcp.Tool {
	return mcp.NewTool("service_cisco_ftdv_get_by_id",
		mcp.WithDescription("Get details of one Cisco Firepower Threat " +
			"Defense virtual service by its ID."),
		mcp.WithString("serviceId",
			mcp.Required(),
			mcp.Description("Service ID."),
		),
	)
}

func ServiceCiscoFTDvGetByName() mcp.Tool {
	return mcp.NewTool("service_cisco_ftdv_get_by_name",
		mcp.WithDescription("Get details of one Cisco Firepower Threat " +
			"Defense virtual service by its name."),
		mcp.WithString("serviceName",
			mcp.Required(),
			mcp.Description("Service Name."),
		),
	)
}

func ServiceCiscoFTDvGetTotal() mcp.Tool {
	return mcp.NewTool("service_cisco_ftdv_get_total",
		mcp.WithDescription("Get total number of Cisco Firepower Threat Defense virtual services"),
	)
}
