package registry

import "github.com/alkiranet/mcp-alkira/writes"

func init() {
	writeFamilies = append(writeFamilies,
		family(writes.SegmentResource),
		family(writes.SegmentResourceShare),
		family(writes.ConnectorInternet),
		family(writes.InternetApplication),
		family(writes.Byoip),
	)
}
