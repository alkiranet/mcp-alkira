package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetPanServices() mcp.Tool {
	return mcp.NewTool("getPanServices",
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

func GetPanServiceById() mcp.Tool {
	return mcp.NewTool("getPanServiceById",
		mcp.WithDescription("Get details of one Palo Alto Networks firewall " +
			"service by its ID."),
		mcp.WithString("serviceId",
			mcp.Required(),
			mcp.Description("Service ID."),
		),
	)
}

func GetPanServiceByName() mcp.Tool {
	return mcp.NewTool("getPanServiceByName",
		mcp.WithDescription("Get details of one Palo Alto Networks firewall " +
			"service by its name."),
		mcp.WithString("serviceName",
			mcp.Required(),
			mcp.Description("Service Name."),
		),
	)
}
