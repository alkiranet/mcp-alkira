package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ListCommunityGetAll() mcp.Tool {
	return mcp.NewTool("list_community_get_all",
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

func ListCommunityGetById() mcp.Tool {
	return mcp.NewTool("list_community_get_by_id",
		mcp.WithDescription("Get BGP Community list by ID"),
		mcp.WithString("listId", mcp.Required()),
	)
}

func ListCommunityGetByName() mcp.Tool {
	return mcp.NewTool("list_community_get_by_name",
		mcp.WithDescription("Get BGP Community list by name"),
		mcp.WithString("listName", mcp.Required()),
	)
}

func ListCommunityGetTotal() mcp.Tool {
	return mcp.NewTool("list_community_get_total",
		mcp.WithDescription("Get total numbers of BGP Community lists."),
	)
}
