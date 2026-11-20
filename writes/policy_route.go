package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// RoutePolicy is the route policy write family (TPS /route-policies).
var RoutePolicy = handlers.WriteFamily[ak.RoutePolicyConfig]{
	Name:       "policy_route",
	Noun:       "route policy",
	Descriptor: ak.RoutePolicyDescriptor(),
	UpdateNote: "Refused with no request while the route policy is FAILED, SCHEDULED or IN_PROGRESS, or while the network is FAILED and the policy is not PENDING.",
	DeleteNote: "A route policy still referenced is refused as in use. " +
		"A deployed route policy stays live until a provision that includes the delete.",
}
