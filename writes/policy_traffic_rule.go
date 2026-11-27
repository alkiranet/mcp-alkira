package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// TrafficRule is the traffic policy rule write family (policy-service /policy/rules).
var TrafficRule = handlers.WriteFamily[ak.TrafficPolicyRuleConfig]{
	Name:       "policy_traffic_rule",
	Noun:       "traffic policy rule",
	Descriptor: ak.TrafficPolicyRuleDescriptor(),
	CreateNote: "matchCondition needs exactly one source, srcIp (\"any\" matches every address) or srcPrefixListId, and exactly one destination, dstIp, dstPrefixListId or internetApplicationId; the portal sends \"any\" for an unset srcIp or dstIp. " +
		"srcPortList is required, [\"any\"] for every port; dstPortList is required unless internetApplicationId is set; protocol any allows only the port \"any\". " +
		"A rule has no effect until a traffic policy rule list holds it. The pending change is listed with the network's policy changes.",
	UpdateNote: "To switch the source between srcIp and srcPrefixListId, or the destination between dstIp, dstPrefixListId and internetApplicationId, set the old key to null in the same patch, " +
		"for example {\"matchCondition\": {\"srcIp\": \"10.1.0.0/16\", \"srcPrefixListId\": null}}: the backend refuses a rule with two sources (rule.00000006) or two destinations (rule.00000007). " +
		"While the tenant feature DEPRECATE-QOSMOS-APPLICATION-FAMILY is on, a create or patch that sets a non-empty applicationFamilyList is refused with no request, and an update sends a stored list empty, as the portal does, and the output says so. " +
		"The backend returns srcPortList and dstPortList in its own order; a reordered list is not reported as a change.",
	DeleteNote: "A rule held by a rule list is refused as in use; remove it from the rule list first.",
}
