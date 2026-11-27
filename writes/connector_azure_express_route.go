package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ConnectorAzureExpressRoute is the Azure ExpressRoute connector write family
// (TPS /v1/tenantnetworks/{id}/azure-express-route-connectors).
var ConnectorAzureExpressRoute = handlers.WriteFamily[ak.ConnectorAzureExpressRouteConfig]{
	Name:       "connector_azure_express_route",
	Noun:       "Azure ExpressRoute connector",
	Descriptor: ak.ConnectorAzureExpressRouteDescriptor(),
	// TPS creates a connector's implicit group under the connector's name
	// and removes it with the connector.
	Implicit: func(name string) []string { return []string{name} },
	Volatile: []string{"lastEnabled"},
	CreateNote: "Connector names are unique across every connector type and group, connectors marked for deletion included, and must not start with ALK-. " +
		"While the tenant feature EXPRESS-ROUTE-CONNECTOR-DISABLE is on, the backend refuses creates and updates with \"not supported for your tenant\"; deletes still run. " +
		"Each instance names an existing ExpressRoute circuit by its Azure resource id and the id of a stored Azure credential that can read it; no tool lists credentials or circuits, so both come from the Alkira portal. The backend reads the circuit from Azure while it saves and refuses a circuit another connector or instance uses. " +
		"vhubPrefix is a free /23 shared by every ExpressRoute connector on the CXP (a second connector on the CXP must use the same one); loopbackSubnet is a free /26 per instance; the backend reserves both. " +
		"tunnelProtocol defaults to VXLAN_GPE. With VXLAN, gatewayMacAddresses holds one address per router (two with redundantRouter) and vnis is optional. With IPSEC, each instance's segmentOptions names the same segments as the connector segmentOptions, with one customer gateway per router, and each tunnel either leaves ikeVersion, profileId, remoteAuthType and remoteAuthValue all unset or sets all four together with initiator. " +
		"Pre-shared keys cannot be set here; the backend generates them. size defaults to LARGE. localAsn is ignored by the backend unless the tenant feature RELEASE-EXPLICIT-LOCAL-ASN-FOR-BRANCH-CONNECTOR is on. billingTags is sent only while BILLING-TAGS is on.",
	UpdateNote: "The backend refuses updates while the tenant feature EXPRESS-ROUTE-CONNECTOR-DISABLE is on. Refused with no request while the connector is FAILED, SCHEDULED or IN_PROGRESS, or while the network is FAILED and the connector was provisioned before. An id of another connector type is refused. " +
		"An instances patch must list the id of every stored instance (remove drops one by id); an instance without an id is new. Inside an instance, segmentOptions merge by segmentName, customerGateways by name and tunnels by id: an entry the patch leaves out is dropped, and a tunnel without an id is new. " +
		"Stored pre-shared keys are resent unchanged and cannot be set; a patch that drops a tunnel holding a stored key is refused. " +
		"Once provisioned, the backend refuses a change of vhubPrefix, of tunnelProtocol to or from VXLAN_GPE, and of a provisioned instance's circuit, loopbackSubnet, gatewayMacAddresses or vnis.",
	DeleteNote: "A connector used by segment resources, NAT policies or other objects is refused as in use. The backend removes the connector's implicit group, named after the connector, with it, and releases its vhubPrefix when no other ExpressRoute connector on the CXP uses it.",
}
