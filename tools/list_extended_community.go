package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ListExtendedCommunityGetAll() mcp.Tool {
	return mcp.NewTool("list_extended_community_get_all",
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

func ListExtendedCommunityGetById() mcp.Tool {
	return mcp.NewTool("list_extended_community_get_by_id",
		mcp.WithDescription("Get BGP Extended Community list by ID"),
		mcp.WithString("listId", mcp.Required()),
	)
}

func ListExtendedCommunityGetByName() mcp.Tool {
	return mcp.NewTool("list_extended_community_get_by_name",
		mcp.WithDescription("Get BGP Extended Community list by name"),
		mcp.WithString("listName", mcp.Required()),
	)
}

func ListExtendedCommunityGetTotal() mcp.Tool {
	return mcp.NewTool("list_extended_community_get_total",
		mcp.WithDescription("Get total numbers of BGP Extended Community lists."),
	)
}
