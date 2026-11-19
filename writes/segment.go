package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// Segment is the segment write family (TPS /segments).
var Segment = handlers.WriteFamily[ak.SegmentConfig]{
	Name:       "segment",
	Noun:       "segment",
	Descriptor: ak.SegmentDescriptor(),
	// TPS creates the -ALL segment resource and its -ALL-Group group with the
	// segment and deletes them with it.
	Implicit: func(name string) []string {
		return []string{name + "-ALL", name + "-ALL-Group"}
	},
	CreateNote: "reservePublicIPsForUserAndSiteConnectivityForCXPs also needs the tenant feature AKAMAI-PRO-PHASE-1 or ADVANCED-IPSEC-CONNECTOR; without one it is left out. " +
		"ipBlock must be one of ipBlocks.values when both are set, and ipBlocks.values must not repeat a block.",
	UpdateNote: "A FAILED segment is refused with no request; fix it in the portal or with a successful provision first. " +
		"reservePublicIPsForUserAndSiteConnectivityForCXPs also needs the tenant feature AKAMAI-PRO-PHASE-1 or ADVANCED-IPSEC-CONNECTOR. " +
		"With the tenant feature IFA-IPV6-SUPPORT off, a stored srcIpv4PoolList and enableIpv6ToIpv4Translation are sent as [] and false, as the portal does.",
	DeleteNote: "The delete also removes the segment's -ALL segment resource and -ALL-Group group. " +
		"It is not checked against NAT policies or DNS server lists, which keep a segment id that no longer exists: delete or change them first. " +
		"The backend deletes a segment at once even while a provision is deploying it, so do not delete during a provision. " +
		"A deployed segment stays live until a provision that includes the delete.",
}
