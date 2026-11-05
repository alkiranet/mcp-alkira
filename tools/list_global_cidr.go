package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ListGlobalCidrGetAll() mcp.Tool {
	return mcp.NewTool("list_global_cidr_get_all",
		mcp.WithDescription("Get all Global CIDR lists."),
		mcp.WithString("offset",
			mcp.Description("Pagination offset"),
			mcp.DefaultString("0"),
		),
		mcp.WithString("limit",
			mcp.Description("Pagination limit"),
			mcp.DefaultString("50"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func ListGlobalCidrGetById() mcp.Tool {
	return mcp.NewTool("list_global_cidr_get_by_id",
		mcp.WithDescription("Get Global CIDR list by ID"),
		mcp.WithString("listId", mcp.Required()),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func ListGlobalCidrGetByName() mcp.Tool {
	return mcp.NewTool("list_global_cidr_get_by_name",
		mcp.WithDescription("Get Global CIDR list by name"),
		mcp.WithString("listName", mcp.Required()),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func ListGlobalCidrGetTotal() mcp.Tool {
	return mcp.NewTool("list_global_cidr_get_total",
		mcp.WithDescription("Get total numbers of Global CIDR lists."),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}
