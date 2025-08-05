package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAllNatPolicy() mcp.Tool {
	return mcp.NewTool("getAllNatPolicy",
		mcp.WithDescription("Get all NAT policies"),
	)
}

func GetAllNatRule() mcp.Tool {
	return mcp.NewTool("getAllNatRule",
		mcp.WithDescription("Get all NAT policy rules"),
		mcp.WithString("limit",
			mcp.Description("Maximum number of items to return (optional)"),
		),
		mcp.WithString("offset",
			mcp.Description("Number of items to skip (optional)"),
		),
		mcp.WithString("page",
			mcp.Description("Page number (1-indexed, alternative to offset) (optional)"),
		),
	)
}

func GetAllRoutePolicy() mcp.Tool {
	return mcp.NewTool("getAllRoutePolicy",
		mcp.WithDescription("Get all route policies"),
	)
}

func GetAllTrafficPolicy() mcp.Tool {
	return mcp.NewTool("getAllTrafficPolicy",
		mcp.WithDescription("Get all traffic policies"),
	)
}

func GetAllTrafficPolicyRule() mcp.Tool {
	return mcp.NewTool("getAllTrafficPolicyRule",
		mcp.WithDescription("Get all traffic policy rules"),
		mcp.WithString("limit",
			mcp.Description("Maximum number of items to return (optional)"),
		),
		mcp.WithString("offset",
			mcp.Description("Number of items to skip (optional)"),
		),
		mcp.WithString("page",
			mcp.Description("Page number (1-indexed, alternative to offset) (optional)"),
		),
	)
}

func GetAllPolicyRuleList() mcp.Tool {
	return mcp.NewTool("getAllPolicyRuleList",
		mcp.WithDescription("Get all policy rule lists"),
	)
}

func GetAllPolicyPrefixList() mcp.Tool {
	return mcp.NewTool("getAllPolicyPrefixList",
		mcp.WithDescription("Get all policy prefix lists"),
	)
}

func GetAllPolicyFqdnList() mcp.Tool {
	return mcp.NewTool("getAllPolicyFqdnList",
		mcp.WithDescription("Get all policy FQDN lists"),
	)
}