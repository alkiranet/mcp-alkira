package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetTrafficPolicies() mcp.Tool {
	return mcp.NewTool("getTrafficPolicies",
		mcp.WithDescription("Get all traffic policies"),
	)
}

func GetTrafficPolicyById() mcp.Tool {
	return mcp.NewTool("getTrafficPolicyById",
		mcp.WithDescription("Get details of one traffic policy by its ID."),
		mcp.WithString("policyId",
			mcp.Required(),
			mcp.Description("Policy ID."),
		),
	)
}

func GetTrafficPolicyByName() mcp.Tool {
	return mcp.NewTool("getTrafficPolicyByName",
		mcp.WithDescription("Get details of one traffic policy by its name."),
		mcp.WithString("policyName",
			mcp.Required(),
			mcp.Description("Policy Name."),
		),
	)
}

func GetTrafficPolicyRules() mcp.Tool {
	return mcp.NewTool("getTrafficPolicyRules",
		mcp.WithDescription("Get all traffic policy rules"),
	)
}

func GetTrafficPolicyRuleById() mcp.Tool {
	return mcp.NewTool("getTrafficPolicyRuleById",
		mcp.WithDescription("Get details of one traffic rule by its ID."),
		mcp.WithString("ruleId",
			mcp.Required(),
			mcp.Description("Rule ID."),
		),
	)
}

func GetTrafficPolicyRuleByName() mcp.Tool {
	return mcp.NewTool("getTrafficPolicyRuleByName",
		mcp.WithDescription("Get details of one traffic rule by its name."),
		mcp.WithString("ruleName",
			mcp.Required(),
			mcp.Description("Rule Name."),
		),
	)
}
