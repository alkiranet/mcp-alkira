package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func SegmentResourceGetAll() mcp.Tool {
	return mcp.NewTool("segment_resource_get_all",
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

func SegmentResourceGetById() mcp.Tool {
	return mcp.NewTool("segment_resource_get_by_id",
		mcp.WithDescription("Get segment resource by ID."),
		mcp.WithString("resourceId",
			mcp.Required(),
			mcp.Description("Segment Resource ID"),
		),
	)
}

func SegmentResourceGetByName() mcp.Tool {
	return mcp.NewTool("segment_resource_get_by_name",
		mcp.WithDescription("Get segment resource by name."),
		mcp.WithString("resourceName",
			mcp.Required(),
			mcp.Description("Segment Resource Name"),
		),
	)
}

func SegmentResourceGetTotal() mcp.Tool {
	return mcp.NewTool("segment_resource_get_total",
		mcp.WithDescription("Get total numbers of segment resources."),
	)
}
