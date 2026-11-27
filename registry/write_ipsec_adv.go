package registry

import "github.com/alkiranet/mcp-alkira/writes"

func init() {
	writeFamilies = append(writeFamilies,
		family(writes.ConnectorIPsecAdv),
		family(writes.ConnectorIPsecTunnelProfile),
	)
}
