package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ConnectorOciVcn is the OCI VCN connector write family (TPS /oci-vcn-connectors).
var ConnectorOciVcn = handlers.WriteFamily[ak.ConnectorOciVcnConfig]{
	Name:       "connector_oci_vcn",
	Noun:       "OCI VCN connector",
	Descriptor: ak.ConnectorOciVcnDescriptor(),
	Implicit:   func(name string) []string { return []string{name} },
	// lastEnabled moves without a user edit.
	Volatile: []string{"lastEnabled"},
	CreateNote: "Connector names are unique across every connector type and group, connectors marked for deletion included, and must not start with ALK-. segments holds exactly one segment name. " +
		"credentialId is the id of a stored OCI credential; create it in the Alkira portal first, since no tool lists or creates credentials. vcnId is the OCI id of the VCN and customerRegion its OCI region; one connector per VCN in a network. " +
		"The backend reads the VCN from OCI while it saves, so the VCN must exist and the credential must see it. " +
		"Without vcnRouting the backend exports the VCN CIDRs and imports the default route into every route table. " +
		"billingTags is sent only while the tenant feature BILLING-TAGS is on.",
	UpdateNote: "Refused with no request while the connector is FAILED, SCHEDULED or IN_PROGRESS, or while the network is FAILED and the connector was provisioned before. An id of another connector type is refused. " +
		"A patch without vcnRouting leaves routing the backend built to the backend. A patch that sets vcnRouting merges into the stored routing key by key: exportToCXPOptions or importFromCXPOptions left out of the patch is kept and resent, while userInputPrefixes or routeTables set in the patch replace the stored list. Set exportToCXPOptions or importFromCXPOptions to null to drop it. Backend-built entries kept by the merge are resent as user input, so the backend no longer rebuilds them from the VCN.",
	DeleteNote: "A connector used by segment resources, NAT policies or other objects is refused as in use. The backend removes the connector's implicit group, named after the connector, with it.",
}
