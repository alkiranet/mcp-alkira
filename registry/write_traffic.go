package registry

import "github.com/alkiranet/mcp-alkira/writes"

func init() {
	writeFamilies = append(writeFamilies,
		family(writes.TrafficRule),
		family(writes.TrafficRuleList),
		family(writes.TrafficPolicy),
	)
}
