package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetCommunityLists() mcp.Tool {
	return mcp.NewTool("getCommunityLists",
		mcp.WithDescription("Get all BGP Community lists."),
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

func GetCommunityListById() mcp.Tool {
	return mcp.NewTool("getCommunityListById",
		mcp.WithDescription("Get BGP Community list by ID"),
		mcp.WithString("listId", mcp.Required()),
	)
}

func GetCommunityListByName() mcp.Tool {
	return mcp.NewTool("getCommunityListByName",
		mcp.WithDescription("Get BGP Community list by name"),
		mcp.WithString("listName", mcp.Required()),
	)
}
