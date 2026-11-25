package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ConnectorArubaEdge is the Aruba EdgeConnect connector write family (TPS /aruba-edge-connectors).
var ConnectorArubaEdge = handlers.WriteFamily[ak.ConnectorArubaEdgeConfig]{
	Name:       "connector_aruba_edge",
	Noun:       "Aruba EdgeConnect connector",
	Descriptor: ak.ConnectorArubaEdgeDescriptor(),
	Implicit:   sdwanImplicit,
	// lastEnabled moves without a user edit.
	Volatile: []string{"lastEnabled"},
	CreateNote: sdwanCreateNote + " Each instance credentialId names an Aruba EdgeConnect instance credential holding the account key; create it in the Alkira portal first, since no tool lists or creates credentials. " +
		"tunnelProtocol defaults to GRE, which an Azure CXP refuses: set IPSEC there. More than one arubaEdgeVRFMappings entry needs version 9 or later. " +
		"allowList is sent only while the tenant feature RELEASE-CONNECTOR-ARUBA-EDGE-CONNECT-ALLOW-LIST is on, scaleGroupId only while SCALE-GROUPS-FOR-USER-AND-SITES is on.",
	UpdateNote: sdwanUpdateNote("instances") + " After an instance is provisioned its accountName, credentialId, hostName and siteTag cannot change; tunnelProtocol cannot change once the connector is provisioned.",
	DeleteNote: sdwanDeleteNote,
}
