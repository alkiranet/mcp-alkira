package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ListDnsServerGetAll() mcp.Tool {
	return mcp.NewTool("list_dns_server_get_all",
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

func ListDnsServerGetById() mcp.Tool {
	return mcp.NewTool("list_dns_server_get_by_id",
		mcp.WithDescription("Get DNS Server list by ID"),
		mcp.WithString("listId", mcp.Required()),
	)
}

func ListDnsServerGetByName() mcp.Tool {
	return mcp.NewTool("list_dns_server_get_by_name",
		mcp.WithDescription("Get DNS Server list by name"),
		mcp.WithString("listName", mcp.Required()),
	)
}

func ListDnsServerGetTotal() mcp.Tool {
	return mcp.NewTool("list_dns_server_get_total",
		mcp.WithDescription("Get total numbers of DNS Server lists."),
	)
}
