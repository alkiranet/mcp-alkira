package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAllInternetApplication() mcp.Tool {
	return mcp.NewTool("getAllInternetApplication",
		mcp.WithDescription("Get all Internet Applications"),
	)
}
