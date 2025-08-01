package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAllSegmentResources() mcp.Tool {
	return mcp.NewTool("getAllSegmentResources",
		mcp.WithDescription("Get all segment resources"),
	)
}

func GetAllSegmentResourceShares() mcp.Tool {
	return mcp.NewTool("getAllSegmentResourceShares",
		mcp.WithDescription("Get all segment resource shares"),
	)
}