package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func PolicyNatGetAll() mcp.Tool {
	return mcp.NewTool("policy_nat_get_all",
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

func PolicyNatGetById() mcp.Tool {
	return mcp.NewTool("policy_nat_get_by_id",
		mcp.WithDescription("Get details of one NAT policy by its ID."),
		mcp.WithString("policyId",
			mcp.Required(),
			mcp.Description("Policy ID."),
		),
	)
}

func PolicyNatGetByName() mcp.Tool {
	return mcp.NewTool("policy_nat_get_by_name",
		mcp.WithDescription("Get details of one NAT policy by its name."),
		mcp.WithString("policyName",
			mcp.Required(),
			mcp.Description("Policy Name."),
		),
	)
}

func PolicyNatRuleGetAll() mcp.Tool {
	return mcp.NewTool("policy_nat_rule_get_all",
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

func PolicyNatRuleGetById() mcp.Tool {
	return mcp.NewTool("policy_nat_rule_get_by_id",
		mcp.WithDescription("Get details of one NAT rule by its ID."),
		mcp.WithString("ruleId",
			mcp.Required(),
			mcp.Description("Rule ID."),
		),
	)
}

func PolicyNatRuleGetByName() mcp.Tool {
	return mcp.NewTool("policy_nat_rule_get_by_name",
		mcp.WithDescription("Get details of one NAT rule by its name."),
		mcp.WithString("ruleName",
			mcp.Required(),
			mcp.Description("Rule Name."),
		),
	)
}

func PolicyNatGetTotal() mcp.Tool {
	return mcp.NewTool("policy_nat_get_total",
		mcp.WithDescription("Get total numbers of NAT policies."),
	)
}

func PolicyNatRuleGetTotal() mcp.Tool {
	return mcp.NewTool("policy_nat_rule_get_total",
		mcp.WithDescription("Get total numbers of NAT policy rules."),
	)
}
