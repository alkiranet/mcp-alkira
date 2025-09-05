package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetSegmentResources() mcp.Tool {
	return mcp.NewTool("getSegmentResources",
		mcp.WithDescription("Get all segment resources"),
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

func GetSegmentResourceById() mcp.Tool {
	return mcp.NewTool("getSegmentResourceById",
		mcp.WithDescription("Get segment resource by ID."),
		mcp.WithString("resourceId",
			mcp.Required(),
			mcp.Description("Segment Resource ID"),
		),
	)
}

func GetSegmentResourceByName() mcp.Tool {
	return mcp.NewTool("getSegmentResourceByName",
		mcp.WithDescription("Get segment resource by name."),
		mcp.WithString("resourceName",
			mcp.Required(),
			mcp.Description("Segment Resource Name"),
		),
	)
}
