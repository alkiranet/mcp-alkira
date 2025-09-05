package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetRoutePolicies() mcp.Tool {
	return mcp.NewTool("getRoutePoliciesAll",
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

func GetRoutePolicyById() mcp.Tool {
	return mcp.NewTool("getRoutePolicyById",
		mcp.WithDescription("Get details of one route policy by its ID."),
		mcp.WithString("policyId",
			mcp.Required(),
			mcp.Description("Policy ID."),
		),
	)
}

func GetRoutePolicyByName() mcp.Tool {
	return mcp.NewTool("getRoutePolicyByName",
		mcp.WithDescription("Get details of one route policy by its name."),
		mcp.WithString("policyName",
			mcp.Required(),
			mcp.Description("Policy Name."),
		),
	)
}
