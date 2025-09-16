package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ServiceZscalerGetAll() mcp.Tool {
	return mcp.NewTool("service_zscaler_get_all",
		mcp.WithDescription("Get all Zscaler security services"),
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

func ServiceZscalerGetById() mcp.Tool {
	return mcp.NewTool("service_zscaler_get_by_id",
		mcp.WithDescription("Get details of one Zscaler security service " +
			"by its ID."),
		mcp.WithString("serviceId",
			mcp.Required(),
			mcp.Description("Service ID."),
		),
	)
}

func ServiceZscalerGetByName() mcp.Tool {
	return mcp.NewTool("service_zscaler_get_by_name",
		mcp.WithDescription("Get details of one Zscaler security service " +
			"by its name."),
		mcp.WithString("serviceName",
			mcp.Required(),
			mcp.Description("Service Name."),
		),
	)
}

func ServiceZscalerGetTotal() mcp.Tool {
	return mcp.NewTool("service_zscaler_get_total",
		mcp.WithDescription("Get total number of Zscaler security services"),
	)
}
