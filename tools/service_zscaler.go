package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetZscalerServices() mcp.Tool {
	return mcp.NewTool("getZscalerServices",
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

func GetZscalerServiceById() mcp.Tool {
	return mcp.NewTool("getZscalerServiceById",
		mcp.WithDescription("Get details of one Zscaler security service " +
			"by its ID."),
		mcp.WithString("serviceId",
			mcp.Required(),
			mcp.Description("Service ID."),
		),
	)
}

func GetZscalerServiceByName() mcp.Tool {
	return mcp.NewTool("getZscalerServiceByName",
		mcp.WithDescription("Get details of one Zscaler security service " +
			"by its name."),
		mcp.WithString("serviceName",
			mcp.Required(),
			mcp.Description("Service Name."),
		),
	)
}
