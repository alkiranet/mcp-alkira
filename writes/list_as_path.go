package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// AsPathList is the AS-path list write family (TPS /as-path-lists).
var AsPathList = handlers.WriteFamily[ak.ValueListConfig]{
	Name:       "list_as_path",
	Noun:       "AS-path list",
	Descriptor: ak.AsPathListDescriptor(),
	DeleteNote: "An AS-path list used by a route policy or an InterCXP route policy is refused as in use. " +
		"A list still held by rules of route policies pending deletion is marked for deletion instead of removed.",
	// A marked AS-path list is outside every provision sweep and every
	// selected-changes provision (V3 ADR-2:36, :318).
	ProvisionNote: "tears it down if it was deployed. A provision never sweeps an AS-path list delete, and the delete cannot be selected in a provision of chosen changes, so a list left marked for deletion stays marked.",
}
