package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetSegmentResourceShares() mcp.Tool {
	return mcp.NewTool("getSegmentResourceShares",
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

func GetSegmentResourceShareById() mcp.Tool {
	return mcp.NewTool("getSegmentResourceShareById",
		mcp.WithDescription("Get segment resource share by ID."),
		mcp.WithString("shareId",
			mcp.Required(),
			mcp.Description("Segment Resource Share ID"),
		),
	)
}

func GetSegmentResourceShareByName() mcp.Tool {
	return mcp.NewTool("getSegmentResourceShareByName",
		mcp.WithDescription("Get segment resource share by name."),
		mcp.WithString("shareName",
			mcp.Required(),
			mcp.Description("Segment Resource Name"),
		),
	)
}
