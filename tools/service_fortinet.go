package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ServiceFortinetGetAll() mcp.Tool {
	return mcp.NewTool("service_fortinet_get_all",
		mcp.WithDescription("Get all Fortinet firewall services"),
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

func ServiceFortinetGetById() mcp.Tool {
	return mcp.NewTool("service_fortinet_get_by_id",
		mcp.WithDescription("Get details of one Fortinet firewall service " +
			"by its ID."),
		mcp.WithString("serviceId",
			mcp.Required(),
			mcp.Description("Service ID."),
		),
	)
}

func ServiceFortinetGetByName() mcp.Tool {
	return mcp.NewTool("service_fortinet_get_by_name",
		mcp.WithDescription("Get details of one Fortinet firewall service " +
			"by its name."),
		mcp.WithString("serviceName",
			mcp.Required(),
			mcp.Description("Service Name."),
		),
	)
}

func ServiceFortinetGetTotal() mcp.Tool {
	return mcp.NewTool("service_fortinet_get_total",
		mcp.WithDescription("Get total number of Fortinet firewall services"),
	)
}
