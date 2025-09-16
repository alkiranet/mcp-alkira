package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ServicePanGetAll() mcp.Tool {
	return mcp.NewTool("service_pan_get_all",
		mcp.WithDescription("Get all Palo Alto Networks (PAN) firewall services"),
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

func ServicePanGetById() mcp.Tool {
	return mcp.NewTool("service_pan_get_by_id",
		mcp.WithDescription("Get details of one Palo Alto Networks firewall " +
			"service by its ID."),
		mcp.WithString("serviceId",
			mcp.Required(),
			mcp.Description("Service ID."),
		),
	)
}

func ServicePanGetByName() mcp.Tool {
	return mcp.NewTool("service_pan_get_by_name",
		mcp.WithDescription("Get details of one Palo Alto Networks firewall " +
			"service by its name."),
		mcp.WithString("serviceName",
			mcp.Required(),
			mcp.Description("Service Name."),
		),
	)
}

func ServicePanGetTotal() mcp.Tool {
	return mcp.NewTool("service_pan_get_total",
		mcp.WithDescription("Get total number of Palo Alto Networks firewall services"),
	)
}
