package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetFortinetServices() mcp.Tool {
	return mcp.NewTool("getFortinetServices",
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

func GetFortinetServiceById() mcp.Tool {
	return mcp.NewTool("getFortinetServiceById",
		mcp.WithDescription("Get details of one Fortinet firewall service " +
			"by its ID."),
		mcp.WithString("serviceId",
			mcp.Required(),
			mcp.Description("Service ID."),
		),
	)
}

func GetFortinetServiceByName() mcp.Tool {
	return mcp.NewTool("getFortinetServiceByName",
		mcp.WithDescription("Get details of one Fortinet firewall service " +
			"by its name."),
		mcp.WithString("serviceName",
			mcp.Required(),
			mcp.Description("Service Name."),
		),
	)
}
