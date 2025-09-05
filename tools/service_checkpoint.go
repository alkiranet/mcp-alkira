package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetCheckPointServices() mcp.Tool {
	return mcp.NewTool("getCheckPointServices",
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

func GetCheckPointServiceById() mcp.Tool {
	return mcp.NewTool("getCheckPointServiceById",
		mcp.WithDescription("Get details of one Check Point firewall service " +
			"by its ID."),
		mcp.WithString("serviceId",
			mcp.Required(),
			mcp.Description("Service ID."),
		),
	)
}

func GetCheckPointServiceByName() mcp.Tool {
	return mcp.NewTool("getCheckPointServiceByName",
		mcp.WithDescription("Get details of one Check Point firewall service " +
			"by its name."),
		mcp.WithString("serviceName",
			mcp.Required(),
			mcp.Description("Service Name."),
		),
	)
}
