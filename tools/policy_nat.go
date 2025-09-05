package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetNatPolicies() mcp.Tool {
	return mcp.NewTool("getNatPolicies",
		mcp.WithDescription("Get all NAT policies. By default, max 50 policies " +
			"will be returned once."),
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

func GetNatPolicyById() mcp.Tool {
	return mcp.NewTool("getNatPolicyById",
		mcp.WithDescription("Get details of one NAT policy by its ID."),
		mcp.WithString("policyId",
			mcp.Required(),
			mcp.Description("Policy ID."),
		),
	)
}

func GetNatPolicyByName() mcp.Tool {
	return mcp.NewTool("getNatPolicyByName",
		mcp.WithDescription("Get details of one NAT policy by its name."),
		mcp.WithString("policyName",
			mcp.Required(),
			mcp.Description("Policy Name."),
		),
	)
}

func GetNatPolicyRules() mcp.Tool {
	return mcp.NewTool("getNatPolicyRules",
		mcp.WithDescription("Get all NAT policy rules. By default, max 50 " +
			"rules will be returned once."),
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

func GetNatPolicyRuleById() mcp.Tool {
	return mcp.NewTool("getNatPolicyRuleById",
		mcp.WithDescription("Get details of one NAT rule by its ID."),
		mcp.WithString("ruleId",
			mcp.Required(),
			mcp.Description("Rule ID."),
		),
	)
}

func GetNatPolicyRuleByName() mcp.Tool {
	return mcp.NewTool("getNatPolicyRuleByName",
		mcp.WithDescription("Get details of one NAT rule by its name."),
		mcp.WithString("ruleName",
			mcp.Required(),
			mcp.Description("Rule Name."),
		),
	)
}
