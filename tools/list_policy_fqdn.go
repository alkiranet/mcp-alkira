package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetPolicyFqdnLists() mcp.Tool {
	return mcp.NewTool("getPolicyFqdnList",
		mcp.WithDescription("Get all Policy FQDN lists."),
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

func GetPolicyFqdnListById() mcp.Tool {
	return mcp.NewTool("getPolicyFqdnListById",
		mcp.WithDescription("Get Policy FQDN list by ID"),
		mcp.WithString("listId", mcp.Required()),
	)
}

func GetPolicyFqdnListByName() mcp.Tool {
	return mcp.NewTool("getPolicyFqdnListByName",
		mcp.WithDescription("Get Policy FQDN list by name"),
		mcp.WithString("listName", mcp.Required()),
	)
}
