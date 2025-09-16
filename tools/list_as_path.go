package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ListAsPathGetAll() mcp.Tool {
	return mcp.NewTool("list_as_path_get_all",
		mcp.WithDescription("Get all AS Path lists for BGP routing."),
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

func ListAsPathGetById() mcp.Tool {
	return mcp.NewTool("list_as_path_get_by_id",
		mcp.WithDescription("Get AS Path list by ID"),
		mcp.WithString("listId", mcp.Required()),
	)
}

func ListAsPathGetByName() mcp.Tool {
	return mcp.NewTool("list_as_path_get_by_name",
		mcp.WithDescription("Get AS Path list by name"),
		mcp.WithString("listName", mcp.Required()),
	)
}

func ListAsPathGetTotal() mcp.Tool {
	return mcp.NewTool("list_as_path_get_total",
		mcp.WithDescription("Get total numbers of AS Path lists."),
	)
}
