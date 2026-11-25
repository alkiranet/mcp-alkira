package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ConnectorIPsecTunnelProfile is the IPsec tunnel profile write family (TPS
// /ipsec-tunnel-profiles).
var ConnectorIPsecTunnelProfile = handlers.WriteFamily[ak.IPsecTunnelProfileConfig]{
	Name:       "connector_ipsec_tunnel_profile",
	Noun:       "IPsec tunnel profile",
	Descriptor: ak.IPsecTunnelProfileDescriptor(),
	CreateNote: "ipSecConfiguration holds the ESP algorithms and ikeConfiguration the IKE algorithms; algorithm names are upper case as the backend lists them, for example AES256CBC, SHA256, MODP2048. " +
		"An AES GCM ESP algorithm carries its own integrity, so integrityAlgorithm may be left out, and it cannot be paired with dhGroup NONE. " +
		"Profiles are not provisioned: an advanced IPsec connector tunnel refers to one by profileId.",
	UpdateNote: "The same algorithm rules as on create apply to the merged profile.",
	DeleteNote: "A profile used by an advanced IPsec connector or an ExpressRoute connector tunnel is refused as in use.",
}
