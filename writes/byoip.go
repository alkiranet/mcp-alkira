package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// Byoip is the BYOIP prefix write family (TPS /byoips). A prefix has no name;
// tool output names it by prefix.
var Byoip = handlers.WriteFamily[ak.ByoipConfig]{
	Name:       "byoip",
	Noun:       "BYOIP",
	NameKey:    "prefix",
	Descriptor: ak.ByoipDescriptor(),
	CreateNote: "prefix is a /24 you own and is unique across every tenant: the duplicate check reads this network, so a prefix held elsewhere is refused by the backend instead. " +
		"extraAttributes is the signed ROA (message, signature, publicKey); the backend verifies the signature. doNotAdvertise is sent only while the tenant feature AKAMAI-PRO-PHASE-1 is on. " +
		"cloudProvider defaults to AWS and must match the cloud of the cxp; a mismatch comes back from the backend as 403 and is reported as permission_denied with a message naming both providers; it is an argument error, not a permission problem.",
	UpdateNote: "description always changes. extraAttributes changes only while no connector or application uses the prefix, " +
		"and doNotAdvertise only while the tenant feature AKAMAI-PRO-PHASE-1 is on and onboardingStatus is PENDING; otherwise the backend keeps the stored value. " +
		"prefix, cxp and cloudProvider cannot change.",
	DeleteNote: "A prefix used by an internet connector is refused as in use. The backend does not check internet applications: a delete succeeds while an application still holds the prefix's byoipId, " +
		"so check usageCount and the applications that use the prefix before confirming. A prefix used by an Akamai Prolexic connector is refused as backend_error (status 400, constraint.notvalid), not as in_use.",
}
