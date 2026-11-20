package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// PolicyFqdnList is the policy FQDN list write family (TPS /policy-fqdn-lists).
var PolicyFqdnList = handlers.WriteFamily[ak.PolicyFqdnListConfig]{
	Name:       "list_policy_fqdn",
	Noun:       "policy FQDN list",
	Descriptor: ak.PolicyFqdnListDescriptor(),
	CreateNote: "dnsServerListId is the id of an existing DNS server list. fqdns keeps its order. " +
		"The number of fqdns is bounded by the tenant's resource limits (one by default), not by this schema.",
	DeleteNote: "A policy FQDN list used by a NAT rule or an internet application is refused as in use.",
}
