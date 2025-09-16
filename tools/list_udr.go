package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ListUdrGetAll() mcp.Tool {
	return mcp.NewTool("list_udr_get_all",
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

func ListUdrGetById() mcp.Tool {
	return mcp.NewTool("list_udr_get_by_id",
		mcp.WithDescription("Get User Defined Route (UDR) list by ID"),
		mcp.WithString("listId", mcp.Required()),
	)
}

func ListUdrGetByName() mcp.Tool {
	return mcp.NewTool("list_udr_get_by_name",
		mcp.WithDescription("Get User Defined Route (UDR) list by name"),
		mcp.WithString("listName", mcp.Required()),
	)
}

func ListUdrGetTotal() mcp.Tool {
	return mcp.NewTool("list_udr_get_total",
		mcp.WithDescription("Get total numbers of User Defined Route (UDR) lists."),
	)
}
