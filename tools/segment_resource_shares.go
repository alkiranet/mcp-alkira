package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func SegmentResourceShareGetAll() mcp.Tool {
	return mcp.NewTool("segment_resource_share_get_all",
		mcp.WithDescription("Get all segment resource shares"),
		mcp.WithString("offset",
			mcp.Description("Pagination offset"),
			mcp.DefaultString("0"),
		),
		mcp.WithString("limit",
			mcp.Description("Pagination limit"),
			mcp.DefaultString("20"),
		),
	)
}

func SegmentResourceShareGetById() mcp.Tool {
	return mcp.NewTool("segment_resource_share_get_by_id",
		mcp.WithDescription("Get segment resource share by ID."),
		mcp.WithString("shareId",
			mcp.Required(),
			mcp.Description("Segment Resource Share ID"),
		),
	)
}

func SegmentResourceShareGetByName() mcp.Tool {
	return mcp.NewTool("segment_resource_share_get_by_name",
		mcp.WithDescription("Get segment resource share by name."),
		mcp.WithString("shareName",
			mcp.Required(),
			mcp.Description("Segment Resource Name"),
		),
	)
}

func SegmentResourceShareGetTotal() mcp.Tool {
	return mcp.NewTool("segment_resource_share_get_total",
		mcp.WithDescription("Get total numbers of segment resource shares."),
	)
}
