package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ExtendedCommunityList is the extended community list write family (TPS /extended-community-lists).
var ExtendedCommunityList = handlers.WriteFamily[ak.ValueListConfig]{
	Name:       "list_extended_community",
	Noun:       "extended community list",
	Descriptor: ak.ExtendedCommunityListDescriptor(),
	CreateNote: "values holds either rt and soo values only, or regex values only; a list that mixes them is refused with no request.",
	UpdateNote: "The values after the update must still be all rt and soo values or all regex values; a mix is refused with no request.",
	DeleteNote: "An extended community list used by a route policy or an InterCXP route policy is refused as in use.",
}
