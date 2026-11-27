package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ConnectorAwsVpc is the AWS VPC connector write family (TPS /awsvpcconnectors).
var ConnectorAwsVpc = handlers.WriteFamily[ak.ConnectorAwsVpcConfig]{
	Name:       "connector_aws_vpc",
	Noun:       "AWS VPC connector",
	Descriptor: ak.ConnectorAwsVpcDescriptor(),
	Implicit:   func(name string) []string { return []string{name} },
	// lastEnabled moves without a user edit.
	Volatile: []string{"lastEnabled"},
	CreateNote: "Connector names are unique across every connector type and group, connectors marked for deletion included, and must not start with ALK-. segments holds exactly one segment name. " +
		"credentialId is the id of a stored AWS credential; create it in the Alkira portal first, since no tool lists or creates credentials. The backend reads the VPC from AWS while it saves, so vpcId and customerRegion must name a VPC that credential can see. " +
		"Without vpcRouting the backend exports the VPC CIDRs and imports the default route into every route table. " +
		"billingTags is sent only while the tenant feature BILLING-TAGS is on, enabled only while CXP-FAILOVER-REGION is on (the backend stores true otherwise), scaleGroupId only while SCALE-GROUPS-FOR-CLOUD-CONNECTORS is on.",
	UpdateNote: "Refused with no request while the connector is FAILED, SCHEDULED or IN_PROGRESS, or while the network is FAILED and the connector was provisioned before. An id of another connector type is refused. " +
		"vpcId and tgwConnectEnabled cannot change. A patch without vpcRouting leaves routing the backend built to the backend. A patch that sets vpcRouting merges into the stored routing key by key: exportToCXPOptions or importFromCXPOptions left out of the patch is kept and resent, while userInputPrefixes or routeTables set in the patch replace the stored list. Set exportToCXPOptions or importFromCXPOptions to null to drop it. Backend-built entries kept by the merge are resent as user input, so the backend no longer rebuilds them from the VPC.",
	DeleteNote: "A connector used by segment resources, NAT policies or other objects is refused as in use. The backend removes the connector's implicit group, named after the connector, with it.",
}
