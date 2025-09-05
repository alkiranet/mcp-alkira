package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetF5LbServices() mcp.Tool {
	return mcp.NewTool("getF5LbServices",
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

func GetF5LbServiceById() mcp.Tool {
	return mcp.NewTool("getF5LbServiceById",
		mcp.WithDescription("Get details of one F5 Load Balancer service " +
			"by its ID."),
		mcp.WithString("serviceId",
			mcp.Required(),
			mcp.Description("Service ID."),
		),
	)
}

func GetF5LbServiceByName() mcp.Tool {
	return mcp.NewTool("getF5LbServiceByName",
		mcp.WithDescription("Get details of one F5 Load Balancer service " +
			"by its name."),
		mcp.WithString("serviceName",
			mcp.Required(),
			mcp.Description("Service Name."),
		),
	)
}
