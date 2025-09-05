package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetBillingTags() mcp.Tool {
	return mcp.NewTool("getBillingTags",
		mcp.WithDescription("Get all billing tags"),
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

func GetBillingTagById() mcp.Tool {
	return mcp.NewTool("getBillingTagById",
		mcp.WithDescription("Get billing tag by ID"),
		mcp.WithString("tagId", mcp.Required()),
	)
}

func GetBillingTagByName() mcp.Tool {
	return mcp.NewTool("getBillingTagByName",
		mcp.WithDescription("Get billing tag by name"),
		mcp.WithString("tagName", mcp.Required()),
	)
}
