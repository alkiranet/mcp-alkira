package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// TrafficRuleList is the traffic policy rule list write family (policy-service /policy/rulelists).
var TrafficRuleList = handlers.WriteFamily[ak.TrafficPolicyRuleListConfig]{
	Name:       "policy_traffic_rule_list",
	Noun:       "traffic policy rule list",
	Descriptor: ak.TrafficPolicyRuleListDescriptor(),
	CreateNote: "Each rules entry names a traffic policy rule by ruleId with a priority of 1 or more; priorities and rule ids must each differ, and the portal numbers priorities 1, 2, 3 in list order. " +
		"A rule list has no effect until a traffic policy uses it.",
	UpdateNote: "A patch that sets rules replaces the whole list; send every entry to keep.",
	DeleteNote: "A rule list used by a traffic policy is refused as in use; delete the policy first. The rules it holds are not deleted.",
}
