package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// SegmentResourceShare is the segment resource share write family (TPS /segment-resource-shares).
var SegmentResourceShare = handlers.WriteFamily[ak.SegmentResourceShareConfig]{
	Name:       "segment_resource_share",
	Noun:       "segment resource share",
	Descriptor: ak.SegmentResourceShareDescriptor(),
	CreateNote: "designatedSegment is a segment name. The endAResources and endBResources segment resource ids sit in two different segments, one of them the designated segment. " +
		"serviceList [0] means no firewall and [-1] any firewall; set serviceList or ruleListId, not both. fromEnd applies to a UNIDIRECTIONAL share only. " +
		"The backend creates an inter-segment policy for the share; it is not shown.",
	UpdateNote: "designatedSegment cannot change after the share was provisioned. A FAILED share is refused with no request.",
	DeleteNote: "The backend removes the share's inter-segment policy with it. A deployed share stays live until a provision that includes the delete.",
}
