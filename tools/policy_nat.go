package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetNatPolicies() mcp.Tool {
	return mcp.NewTool("getNatPolicies",
		mcp.WithDescription("Get all NAT policies. There could be too "+
			"many policies that can't be processed. In this case, tool "+
			"getNatPoliciesSummary could be used to get a shorter list. "+
			"Otherwise, user should try to get a single NAT policy "+
			"instead of listing all."),
		mcp.WithString("offset",
			mcp.Description("Offset of paginated data will be returned."),
		),
		mcp.WithString("limit",
			mcp.Description("Limit of paginated data will be returned. If not " +
				"provided, default value is 10."),
		),
	)
}

func GetNatPoliciesSummary() mcp.Tool {
	return mcp.NewTool("getNatPoliciesSummary",
		mcp.WithDescription("Get all NAT policies in summary format. "+
			"This will only return `policy ID` and `policy name`."),
		mcp.WithString("offset",
			mcp.Description("Offset of paginated data will be returned."),
		),
		mcp.WithString("limit",
			mcp.Description("Limit of paginated data will be returned. If not " +
				"provided, default value is 10."),
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
		mcp.WithDescription("Get all NAT policy rules"),
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
