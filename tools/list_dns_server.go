package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetDnsServerLists() mcp.Tool {
	return mcp.NewTool("getDnsServerLists",
		mcp.WithDescription("Get all DNS Server lists."),
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

func GetDnsServerListById() mcp.Tool {
	return mcp.NewTool("getDnsServerListById",
		mcp.WithDescription("Get DNS Server list by ID"),
		mcp.WithString("listId", mcp.Required()),
	)
}

func GetDnsServerListByName() mcp.Tool {
	return mcp.NewTool("getDnsServerListByName",
		mcp.WithDescription("Get DNS Server list by name"),
		mcp.WithString("listName", mcp.Required()),
	)
}
