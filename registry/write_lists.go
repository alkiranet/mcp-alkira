package registry

import "github.com/alkiranet/mcp-alkira/writes"

func init() {
	writeFamilies = append(writeFamilies,
		family(writes.AsPathList),
		family(writes.CommunityList),
		family(writes.ExtendedCommunityList),
		family(writes.GlobalCidrList),
		family(writes.UdrList),
		family(writes.DnsServerList),
		family(writes.PolicyFqdnList),
	)
}
