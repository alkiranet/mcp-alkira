package registry

import "github.com/alkiranet/mcp-alkira/writes"

func init() {
	writeFamilies = append(writeFamilies,
		family(writes.PrefixList),
		family(writes.NatRule),
		family(writes.NatPolicy),
		family(writes.RoutePolicy),
	)
}
