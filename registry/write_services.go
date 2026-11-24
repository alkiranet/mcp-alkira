package registry

import "github.com/alkiranet/mcp-alkira/writes"

func init() {
	writeFamilies = append(writeFamilies,
		family(writes.ServicePan),
		family(writes.ServiceCheckPoint),
		family(writes.ServiceCiscoFTDv),
		family(writes.ServiceFortinet),
	)
}
