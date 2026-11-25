package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ConnectorGcpInterconnect is the GCP Interconnect connector write family
// (TPS /gcp-interconnect-connectors). The instance BGP key is not a tool
// argument: an update resends each stored key unchanged.
var ConnectorGcpInterconnect = handlers.WriteFamily[ak.ConnectorGcpInterconnectConfig]{
	Name:       "connector_gcp_interconnect",
	Noun:       "GCP Interconnect connector",
	Descriptor: ak.ConnectorGcpInterconnectDescriptor(),
	Implicit:   func(name string) []string { return []string{name} },
	// lastEnabled moves without a user edit.
	Volatile: []string{"lastEnabled"},
	CreateNote: "Connector names are unique across every connector type and group, connectors marked for deletion included, and must not start with ALK-. " +
		"Each instance is a VLAN attachment; every instance lists the same segments in segmentOptions, each segment once, with at least one customer gateway (tunnelCount 2) and disableInternetExit and advertiseOnPremRoutes set. " +
		"loopbackPrefixes are /23 to /26 blocks the backend reserves. tunnelProtocol defaults to GRE; IPSEC, VXLAN and VXLAN_GPE need their tenant feature on the CXP. gatewayMacAddress is set exactly for VXLAN, vni exactly for VXLAN and VXLAN_GPE. " +
		"The BGP key cannot be set here. size defaults to LARGE. " +
		"billingTags is sent only while the tenant feature BILLING-TAGS is on, scaleGroupId only while SUPPORT-SCALE-GROUP-FOR-DX-CONNECTOR is on. A customer gateway loopbackIp needs GCP-INTERCONNECT-CGW-LOOPBACK-IP, and instance localAsn is kept only while RELEASE-EXPLICIT-LOCAL-ASN-FOR-BRANCH-CONNECTOR is on.",
	UpdateNote: "Refused with no request while the connector is FAILED, SCHEDULED or IN_PROGRESS, or while the network is FAILED and the connector was provisioned before. An id of another connector type is refused. " +
		"An instances patch must list the id of every stored instance; an instance without an id is added as a new one. Each stored BGP key is resent unchanged and cannot be set here. " +
		"Once the connector is provisioned, tunnelProtocol and an instance's gcpEdgeAvailabilityDomain cannot change and a loopback prefix cannot be removed.",
	DeleteNote: "A connector used by segment resources or other objects is refused as in use. The backend removes the connector's implicit group, named after the connector, with it.",
}
