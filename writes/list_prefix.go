package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// PrefixList is the prefix list write family (policy-service /policy/prefixlists).
var PrefixList = handlers.WriteFamily[ak.PrefixListConfig]{
	Name:       "list_prefix",
	Noun:       "prefix list",
	Descriptor: ak.PrefixListDescriptor(),
	CreateNote: "The backend derives the list type from prefixRanges on every save, so type is not an input. " +
		"The number of prefixes and prefix ranges is bounded by the tenant's resource limits, not by this schema. " +
		"The pending change is listed with the network's policy changes.",
	DeleteNote: "A prefix list used by a policy, a rule or another object is refused as in use. " +
		"A deployed prefix list stays live until a provision that includes the delete.",
}
