package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetCiscoFTDvServices() mcp.Tool {
	return mcp.NewTool("getCiscoFTDvServices",
		mcp.WithDescription("Get all Cisco Firepower Threat Defense virtual " +
			"services"),
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

func GetCiscoFTDvServiceById() mcp.Tool {
	return mcp.NewTool("getCiscoFTDvServiceById",
		mcp.WithDescription("Get details of one Cisco Firepower Threat " +
			"Defense virtual service by its ID."),
		mcp.WithString("serviceId",
			mcp.Required(),
			mcp.Description("Service ID."),
		),
	)
}

func GetCiscoFTDvServiceByName() mcp.Tool {
	return mcp.NewTool("getCiscoFTDvServiceByName",
		mcp.WithDescription("Get details of one Cisco Firepower Threat " +
			"Defense virtual service by its name."),
		mcp.WithString("serviceName",
			mcp.Required(),
			mcp.Description("Service Name."),
		),
	)
}
