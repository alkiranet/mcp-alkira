package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetUdrLists() mcp.Tool {
	return mcp.NewTool("getUdrLists",
		mcp.WithDescription("Get all User Defined Route (UDR) lists."),
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

func GetUdrListById() mcp.Tool {
	return mcp.NewTool("getUdrListById",
		mcp.WithDescription("Get User Defined Route (UDR) list by ID"),
		mcp.WithString("listId", mcp.Required()),
	)
}

func GetUdrListByName() mcp.Tool {
	return mcp.NewTool("getUdrListByName",
		mcp.WithDescription("Get User Defined Route (UDR) list by name"),
		mcp.WithString("listName", mcp.Required()),
	)
}
