package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ServiceZscaler is the Zscaler internet access service write family (TPS
// /zscaler-internet-access-services). The pre-shared key is a required
// write-only create argument, masked in every output.
var ServiceZscaler = handlers.WriteFamily[ak.ServiceZscalerConfig]{
	Name:       "service_zscaler",
	Noun:       "Zscaler internet access service",
	Descriptor: ak.ServiceZscalerDescriptor(),
	// TPS deletes the service's firewall zones before the service
	// (ZscalerInternetAccessServiceLogicImpl.java:124-127).
	ImplicitNote: func(string, []byte) string { return "its firewall zones" },
	CreateNote: "ipsecConfiguration.preSharedKey is required, must not be blank and must match the key configured for this location in ZIA, as the portal form requires; it is write-only: no output shows it, and an update resends the stored key unchanged. " +
		"segments holds exactly one segment name. localFqdnId is the IKE local identity of the ZIA location's VPN credential. primaryPublicEdgeIp and secondaryPublicEdgeIp are the ZIA public service edges.",
	UpdateNote: "The stored pre-shared key is resent unchanged and cannot be set here; change it in the portal. cxp cannot change. " +
		"Changing segments deletes the service's firewall zones and creates a default zone. " +
		"Refused with no request while the service is FAILED, SCHEDULED or IN_PROGRESS, or while the network is FAILED and the service was provisioned before.",
	DeleteNote: "A service a policy uses is refused by the backend with its reason.",
}
