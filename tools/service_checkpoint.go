package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ServiceCheckPointGetAll() mcp.Tool {
	return mcp.NewTool("service_check_point_get_all",
		mcp.WithDescription("Get all Check Point firewall services"),
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

func ServiceCheckPointGetById() mcp.Tool {
	return mcp.NewTool("service_check_point_get_by_id",
		mcp.WithDescription("Get details of one Check Point firewall service " +
			"by its ID."),
		mcp.WithString("serviceId",
			mcp.Required(),
			mcp.Description("Service ID."),
		),
	)
}

func ServiceCheckPointGetByName() mcp.Tool {
	return mcp.NewTool("service_check_point_get_by_name",
		mcp.WithDescription("Get details of one Check Point firewall service " +
			"by its name."),
		mcp.WithString("serviceName",
			mcp.Required(),
			mcp.Description("Service Name."),
		),
	)
}

func ServiceCheckPointGetTotal() mcp.Tool {
	return mcp.NewTool("service_check_point_get_total",
		mcp.WithDescription("Get total number of Check Point firewall services"),
	)
}
