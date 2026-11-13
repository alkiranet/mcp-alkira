package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func SegmentGetAll() mcp.Tool {
	return mcp.NewTool("segment_get_all",
		mcp.WithDescription("Get all segments."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func SegmentGetById() mcp.Tool {
	return mcp.NewTool("segment_get_by_id",
		mcp.WithDescription("Get details of one segment by its ID."),
		mcp.WithString("segmentId",
			mcp.Required(),
			mcp.Description("Segment ID."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func SegmentGetByName() mcp.Tool {
	return mcp.NewTool("segment_get_by_name",
		mcp.WithDescription("Get details of one segment by its name."),
		mcp.WithString("segmentName",
			mcp.Required(),
			mcp.Description("Segment Name."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}
