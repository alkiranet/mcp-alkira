package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// TrafficPolicy is the traffic policy write family (policy-service /policy/policies).
var TrafficPolicy = handlers.WriteFamily[ak.TrafficPolicyConfig]{
	Name:       "policy_traffic",
	Noun:       "traffic policy",
	Descriptor: ak.TrafficPolicyDescriptor(),
	CreateNote: "segmentIds takes exactly one segment id; fromGroups and toGroups take at least one group id each; -1 is the portal's ANY group. ruleListId names an existing traffic policy rule list. " +
		"ztaProfileIds is sent only while RELEASE-ZTNA-POLICY-WITH-PROFILES is enabled, as the portal sends it; zta_profile_get_all lists the profiles. " +
		"The pending change is listed with the network's policy changes.",
	UpdateNote: "ztaProfileIds in a patch is left out while RELEASE-ZTNA-POLICY-WITH-PROFILES is off; stored ids are resent unchanged.",
	DeleteNote: "The delete does not delete the policy's rule list or rules.",
}
