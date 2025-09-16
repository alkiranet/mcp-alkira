package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func InternetApplicationGetAll() mcp.Tool {
	return mcp.NewTool("internet_application_get_all",
		mcp.WithDescription("Get all Internet Applications"),
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

func InternetApplicationGetById() mcp.Tool {
	return mcp.NewTool("internet_application_get_by_id",
		mcp.WithDescription("Get Internet Application by ID"),
		mcp.WithString("applicationId", mcp.Required()),
	)
}

func InternetApplicationGetByName() mcp.Tool {
	return mcp.NewTool("internet_application_get_by_name",
		mcp.WithDescription("Get Internet Application by name"),
		mcp.WithString("applicationName", mcp.Required()),
	)
}

func InternetApplicationGetTotal() mcp.Tool {
	return mcp.NewTool("internet_application_get_total",
		mcp.WithDescription("Get total number of Internet Application"),
	)
}
