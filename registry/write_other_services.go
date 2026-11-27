package registry

import "github.com/alkiranet/mcp-alkira/writes"

func init() {
	writeFamilies = append(writeFamilies,
		family(writes.F5LBService),
		family(writes.ServiceInfoblox),
		family(writes.ServiceZscaler),
		family(writes.ConnectorRemoteAccessTemplate),
	)
}
