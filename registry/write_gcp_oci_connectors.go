package registry

import "github.com/alkiranet/mcp-alkira/writes"

func init() {
	writeFamilies = append(writeFamilies,
		family(writes.ConnectorGcpVpc),
		family(writes.ConnectorGcpInterconnect),
		family(writes.ConnectorOciVcn),
	)
}
