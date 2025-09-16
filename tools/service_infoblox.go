package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ServiceInfobloxGetAll() mcp.Tool {
	return mcp.NewTool("service_infoblox_get_all",
		mcp.WithDescription("Get all Infoblox DNS/DHCP services"),
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

func ServiceInfobloxGetById() mcp.Tool {
	return mcp.NewTool("service_infoblox_get_by_id",
		mcp.WithDescription("Get details of one Infoblox DNS/DHCP service " +
			"by its ID."),
		mcp.WithString("serviceId",
			mcp.Required(),
			mcp.Description("Service ID."),
		),
	)
}

func ServiceInfobloxGetByName() mcp.Tool {
	return mcp.NewTool("service_infoblox_get_by_name",
		mcp.WithDescription("Get details of one Infoblox DNS/DHCP service " +
			"by its name."),
		mcp.WithString("serviceName",
			mcp.Required(),
			mcp.Description("Service Name."),
		),
	)
}

func ServiceInfobloxGetTotal() mcp.Tool {
	return mcp.NewTool("service_infoblox_get_total",
		mcp.WithDescription("Get total number of Infoblox DNS/DHCP services"),
	)
}
