package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetSegments() mcp.Tool {
	return mcp.NewTool("getSegments",
		mcp.WithDescription("Get all segments."),
	)
}

func GetSegmentById() mcp.Tool {
	return mcp.NewTool("getSegmentById",
		mcp.WithDescription("Get details of one segment by its ID."),
		mcp.WithString("segmentId",
			mcp.Required(),
			mcp.Description("Segment ID."),
		),
	)
}

func GetSegmentByName() mcp.Tool {
	return mcp.NewTool("getSegmentByName",
		mcp.WithDescription("Get details of one segment by its name."),
		mcp.WithString("segmentName",
			mcp.Required(),
			mcp.Description("Segment Name."),
		),
	)
}
