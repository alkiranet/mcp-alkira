package writes

import (
	"context"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// NatPolicy is the NAT policy write family (TPS /nat-policies).
var NatPolicy = handlers.WriteFamily[ak.NatPolicyConfig]{
	Name:       "policy_nat",
	Noun:       "NAT policy",
	Descriptor: ak.NatPolicyDescriptor(),
	CreateNote: "The backend itself accepts a second NAT policy with the same name; this tool reads every page of the list to refuse it. " +
		"natRuleIds that name no NAT rule are dropped by the backend without an error; the output names each dropped id. " +
		"allowOverlappingOutboundDestinationPrefixes is sent only while RELEASE-NAT-POLICY-OUTBOUND and RELEASE-NAT-POLICY-OUTBOUND-DNAT are both enabled, as the portal sends it.",
	UpdateNote: "natRuleIds that name no NAT rule are dropped by the backend without an error; the output names each dropped id.",
	DeleteNote: "A NAT policy still referenced is refused as in use. " +
		"A segment delete is not checked against NAT policies, so a policy can hold a segment id that no longer exists.",
	// A segment delete leaves the policy's segmentId behind and the policy reads back without segment.
	StoredRefs: []handlers.StoredRef{{Field: "segment", IDField: "segmentId", Noun: "segment",
		Lookup: func(ctx context.Context, c *ak.AlkiraClient, id string) error {
			_, _, err := c.GetSegment(ctx, id)
			return err
		}}},
}
