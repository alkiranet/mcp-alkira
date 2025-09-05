package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAsPathLists() mcp.Tool {
	return mcp.NewTool("getAsPathLists",
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

func GetAsPathListById() mcp.Tool {
	return mcp.NewTool("getAsPathListById",
		mcp.WithDescription("Get AS Path list by ID"),
		mcp.WithString("listId", mcp.Required()),
	)
}

func GetAsPathListByName() mcp.Tool {
	return mcp.NewTool("getAsPathListByName",
		mcp.WithDescription("Get AS Path list by name"),
		mcp.WithString("listName", mcp.Required()),
	)
}
