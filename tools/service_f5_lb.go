package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ServiceF5LBGetAll() mcp.Tool {
	return mcp.NewTool("service_f5_lb_get_all",
		mcp.WithDescription("Get all F5 Load Balancer services"),
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

func ServiceF5LBGetById() mcp.Tool {
	return mcp.NewTool("service_f5_lb_get_by_id",
		mcp.WithDescription("Get details of one F5 Load Balancer service " +
			"by its ID."),
		mcp.WithString("serviceId",
			mcp.Required(),
			mcp.Description("Service ID."),
		),
	)
}

func ServiceF5LBGetByName() mcp.Tool {
	return mcp.NewTool("service_f5_lb_get_by_name",
		mcp.WithDescription("Get details of one F5 Load Balancer service " +
			"by its name."),
		mcp.WithString("serviceName",
			mcp.Required(),
			mcp.Description("Service Name."),
		),
	)
}

func ServiceF5LBGetTotal() mcp.Tool {
	return mcp.NewTool("service_f5_lb_get_total",
		mcp.WithDescription("Get total number of F5 Load Balancer services"),
	)
}
