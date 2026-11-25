package registry

import "github.com/alkiranet/mcp-alkira/writes"

func init() {
	writeFamilies = append(writeFamilies,
		family(writes.ConnectorArubaEdge),
		family(writes.ConnectorCiscoSdwan),
		family(writes.ConnectorFortinetSdwan),
		family(writes.ConnectorVersaSdwan),
		family(writes.ConnectorVmwareSdwan),
	)
}
