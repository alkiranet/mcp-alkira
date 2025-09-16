package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func PolicyRouteGetAll() mcp.Tool {
	return mcp.NewTool("policy_route_get_all",
		mcp.WithDescription("Get all route policies."),
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

func PolicyRouteGetById() mcp.Tool {
	return mcp.NewTool("policy_route_get_by_id",
		mcp.WithDescription("Get details of one route policy by its ID."),
		mcp.WithString("policyId",
			mcp.Required(),
			mcp.Description("Policy ID."),
		),
	)
}

func PolicyRouteGetByName() mcp.Tool {
	return mcp.NewTool("policy_route_get_by_name",
		mcp.WithDescription("Get details of one route policy by its name."),
		mcp.WithString("policyName",
			mcp.Required(),
			mcp.Description("Policy Name."),
		),
	)
}

func PolicyRouteGetTotal() mcp.Tool {
	return mcp.NewTool("policy_route_get_total",
		mcp.WithDescription("Get total numbers of route policies."),
	)
}
