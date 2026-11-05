package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func PolicyTrafficGetAll() mcp.Tool {
	return mcp.NewTool("policy_traffic_get_all",
		mcp.WithDescription("Get all traffic policies"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func PolicyTrafficGetById() mcp.Tool {
	return mcp.NewTool("policy_traffic_get_by_id",
		mcp.WithDescription("Get details of one traffic policy by its ID."),
		mcp.WithString("policyId",
			mcp.Required(),
			mcp.Description("Policy ID."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func PolicyTrafficGetByName() mcp.Tool {
	return mcp.NewTool("policy_traffic_get_by_name",
		mcp.WithDescription("Get details of one traffic policy by its name."),
		mcp.WithString("policyName",
			mcp.Required(),
			mcp.Description("Policy Name."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func PolicyTrafficRuleGetAll() mcp.Tool {
	return mcp.NewTool("policy_traffic_rule_get_all",
		mcp.WithDescription("Get all traffic policy rules"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func PolicyTrafficRuleGetById() mcp.Tool {
	return mcp.NewTool("policy_traffic_rule_get_by_id",
		mcp.WithDescription("Get details of one traffic rule by its ID."),
		mcp.WithString("ruleId",
			mcp.Required(),
			mcp.Description("Rule ID."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func PolicyTrafficRuleGetByName() mcp.Tool {
	return mcp.NewTool("policy_traffic_rule_get_by_name",
		mcp.WithDescription("Get details of one traffic rule by its name."),
		mcp.WithString("ruleName",
			mcp.Required(),
			mcp.Description("Rule Name."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func PolicyTrafficRuleListGetAll() mcp.Tool {
	return mcp.NewTool("policy_traffic_rule_list_get_all",
		mcp.WithDescription("Get all Policy Rule lists."),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func PolicyTrafficRuleListGetById() mcp.Tool {
	return mcp.NewTool("policy_traffic_rule_list_get_by_id",
		mcp.WithDescription("Get traffic policy rule list by ID"),
		mcp.WithString("ruleListId", mcp.Required()),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}

func PolicyTrafficRuleListGetByName() mcp.Tool {
	return mcp.NewTool("policy_traffic_rule_list_get_by_name",
		mcp.WithDescription("Get traffic policy rule list by name"),
		mcp.WithString("ruleListName", mcp.Required()),
		mcp.WithReadOnlyHintAnnotation(true),
	)
}
