package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetInternetApplications() mcp.Tool {
	return mcp.NewTool("getInternetApplications",
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

func GetInternetApplicationById() mcp.Tool {
	return mcp.NewTool("getInternetApplicationById",
		mcp.WithDescription("Get Internet Application by ID"),
		mcp.WithString("applicationId", mcp.Required()),
	)
}

func GetInternetApplicationByName() mcp.Tool {
	return mcp.NewTool("getInternetApplicationByName",
		mcp.WithDescription("Get Internet Application by name"),
		mcp.WithString("applicationName", mcp.Required()),
	)
}
