package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ConnectorGcpVpc is the GCP VPC connector write family (TPS /gcpvpcconnectors).
var ConnectorGcpVpc = handlers.WriteFamily[ak.ConnectorGcpVpcConfig]{
	Name:       "connector_gcp_vpc",
	Noun:       "GCP VPC connector",
	Descriptor: ak.ConnectorGcpVpcDescriptor(),
	Implicit:   func(name string) []string { return []string{name} },
	// lastEnabled moves without a user edit.
	Volatile: []string{"lastEnabled"},
	CreateNote: "Connector names are unique across every connector type and group, connectors marked for deletion included, and must not start with ALK-. segments holds exactly one segment name. " +
		"credentialId is the id of a stored GCP credential; create it in the Alkira portal first, since no tool lists or creates credentials. vpcName is the GCP network name and customerRegion its GCP region; projectId defaults to the credential's project. " +
		"The backend reads the VPC from GCP while it saves, so the network must exist and the credential must see it; it stamps vpcId and routingMode from it. " +
		"Without gcpRouting the backend imports the default route and exports nothing. In gcpRouting.exportToCXPOptions, exportAllSubnets true takes no userInputPrefixes and false needs them; each entry is a SUBNET with the GCP subnet id or fqId. prefixListIds go with routeImportMode ADVERTISE_CUSTOM_PREFIX only. customerAsn defaults to 64522. " +
		"billingTags is sent only while the tenant feature BILLING-TAGS is on, scaleGroupId only while SCALE-GROUPS-FOR-CLOUD-CONNECTORS is on.",
	UpdateNote: "Refused with no request while the connector is FAILED, SCHEDULED or IN_PROGRESS, or while the network is FAILED and the connector was provisioned before. An id of another connector type is refused. " +
		"vpcName and projectId cannot change. vpcId and routingMode are re-read from the VPC on every save. A patch that sets gcpRouting merges into the stored routing key by key, nested objects included: a key left out of the patch keeps its stored value and is resent, while userInputPrefixes or prefixListIds set in the patch replace the stored list. Set a key to null to drop it, for example exportToCXPOptions.userInputPrefixes when setting exportAllSubnets to true.",
	DeleteNote: "A connector used by segment resources, NAT policies or other objects is refused as in use. The backend removes the connector's implicit group, named after the connector, with it.",
}
