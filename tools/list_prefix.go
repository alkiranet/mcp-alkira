package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ListPrefixGetById() mcp.Tool {
	return mcp.NewTool("list_prefix_get_by_id",
		mcp.WithDescription("Get Policy Prefix list by ID"),
		mcp.WithString("listId", mcp.Required()),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func ListPrefixGetByName() mcp.Tool {
	return mcp.NewTool("list_prefix_get_by_name",
		mcp.WithDescription("Get Policy Prefix list by name"),
		mcp.WithString("listName", mcp.Required()),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func ListPrefixGetByPrefix() mcp.Tool {
	return mcp.NewTool("list_prefix_get_by_prefix",
		mcp.WithDescription("Get Policy Prefix lists by a prefix"),
		mcp.WithString("prefix", mcp.Required()),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}
