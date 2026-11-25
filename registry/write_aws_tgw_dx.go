package registry

import "github.com/alkiranet/mcp-alkira/writes"

func init() {
	writeFamilies = append(writeFamilies,
		family(writes.ConnectorAwsTgw),
		family(writes.ConnectorAwsDirectConnect),
	)
}
