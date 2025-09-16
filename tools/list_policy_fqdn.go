package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ListPolicyFqdnGetAll() mcp.Tool {
	return mcp.NewTool("list_policy_fqdn_get_all",
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

func ListPolicyFqdnGetById() mcp.Tool {
	return mcp.NewTool("list_policy_fqdn_get_by_id",
		mcp.WithDescription("Get Policy FQDN list by ID"),
		mcp.WithString("listId", mcp.Required()),
	)
}

func ListPolicyFqdnGetByName() mcp.Tool {
	return mcp.NewTool("list_policy_fqdn_get_by_name",
		mcp.WithDescription("Get Policy FQDN list by name"),
		mcp.WithString("listName", mcp.Required()),
	)
}

func ListPolicyFqdnGetTotal() mcp.Tool {
	return mcp.NewTool("list_policy_fqdn_get_total",
		mcp.WithDescription("Get total numbers of Policy FQDN lists."),
	)
}
