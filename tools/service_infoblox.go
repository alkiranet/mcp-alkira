package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetInfobloxServices() mcp.Tool {
	return mcp.NewTool("getInfobloxServices",
		mcp.WithDescription("Get all Infoblox DNS/DHCP services"),
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

func GetInfobloxServiceById() mcp.Tool {
	return mcp.NewTool("getInfobloxServiceById",
		mcp.WithDescription("Get details of one Infoblox DNS/DHCP service " +
			"by its ID."),
		mcp.WithString("serviceId",
			mcp.Required(),
			mcp.Description("Service ID."),
		),
	)
}

func GetInfobloxServiceByName() mcp.Tool {
	return mcp.NewTool("getInfobloxServiceByName",
		mcp.WithDescription("Get details of one Infoblox DNS/DHCP service " +
			"by its name."),
		mcp.WithString("serviceName",
			mcp.Required(),
			mcp.Description("Service Name."),
		),
	)
}
