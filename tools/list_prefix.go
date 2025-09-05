package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetPrefixListById() mcp.Tool {
	return mcp.NewTool("getPrefixListById",
		mcp.WithDescription("Get Policy Prefix list by ID"),
		mcp.WithString("listId", mcp.Required()),
	)
}

func GetPrefixListByName() mcp.Tool {
	return mcp.NewTool("getPrefixListByName",
		mcp.WithDescription("Get Policy Prefix list by name"),
		mcp.WithString("listName", mcp.Required()),
	)
}
