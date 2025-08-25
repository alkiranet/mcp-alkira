package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetBillingTags() mcp.Tool {
	return mcp.NewTool("getBillingTags",
		mcp.WithDescription("Get all billing tags"),
	)
}
