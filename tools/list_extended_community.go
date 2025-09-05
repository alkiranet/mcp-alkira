package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetExtendedCommunityLists() mcp.Tool {
	return mcp.NewTool("getExtendedCommunityLists",
		mcp.WithDescription("Get all BGP Extended Community lists."),
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

func GetExtendedCommunityListById() mcp.Tool {
	return mcp.NewTool("getExtendedCommunityListById",
		mcp.WithDescription("Get BGP Extended Community list by ID"),
		mcp.WithString("listId", mcp.Required()),
	)
}

func GetExtendedCommunityListByName() mcp.Tool {
	return mcp.NewTool("getExtendedCommunityListByName",
		mcp.WithDescription("Get BGP Extended Community list by name"),
		mcp.WithString("listName", mcp.Required()),
	)
}
