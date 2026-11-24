package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// CommunityList is the community list write family (TPS /community-lists).
var CommunityList = handlers.WriteFamily[ak.ValueListConfig]{
	Name:       "list_community",
	Noun:       "community list",
	Descriptor: ak.CommunityListDescriptor(),
	DeleteNote: "A community list used by a route policy or an InterCXP route policy is refused as in use.",
}
