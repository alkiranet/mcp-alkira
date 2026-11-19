package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// NatRule is the NAT rule write family (TPS /nat-rules).
var NatRule = handlers.WriteFamily[ak.NatRuleConfig]{
	Name:       "policy_nat_rule",
	Noun:       "NAT rule",
	Descriptor: ak.NatRuleDescriptor(),
	CreateNote: "A NAT rule has no effect until a NAT policy lists it in natRuleIds.",
	UpdateNote: "direction cannot change; a patch that sets another direction is refused with no request, with or without a rename.",
	DeleteNote: "A NAT rule used by a NAT policy is refused as in use; remove it from the policy's natRuleIds first. " +
		"A rule held only by NAT policies marked for deletion is itself marked and stays until the next provision.",
}
