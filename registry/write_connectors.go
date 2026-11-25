package registry

import "github.com/alkiranet/mcp-alkira/writes"

func init() {
	writeFamilies = append(writeFamilies,
		family(writes.ConnectorIPsec),
		family(writes.ConnectorAwsVpc),
		family(writes.ConnectorAzureVnet),
	)
}
