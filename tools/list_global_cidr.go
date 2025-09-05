package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetGlobalCidrLists() mcp.Tool {
	return mcp.NewTool("getGlobalCidrLists",
		mcp.WithDescription("Get all Global CIDR lists."),
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

func GetGlobalCidrListById() mcp.Tool {
	return mcp.NewTool("getGlobalCidrListById",
		mcp.WithDescription("Get Global CIDR list by ID"),
		mcp.WithString("listId", mcp.Required()),
	)
}

func GetGlobalCidrListByName() mcp.Tool {
	return mcp.NewTool("getGlobalCidrListByName",
		mcp.WithDescription("Get Global CIDR list by name"),
		mcp.WithString("listName", mcp.Required()),
	)
}
