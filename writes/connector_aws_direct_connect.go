package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ConnectorAwsDirectConnect is the AWS Direct Connect connector write family
// (TPS /directconnectconnectors).
var ConnectorAwsDirectConnect = handlers.WriteFamily[ak.ConnectorAwsDirectConnectConfig]{
	Name:       "connector_aws_direct_connect",
	Noun:       "AWS Direct Connect connector",
	Descriptor: ak.ConnectorAwsDirectConnectDescriptor(),
	Implicit:   func(name string) []string { return []string{name} },
	// lastEnabled moves without a user edit.
	Volatile: []string{"lastEnabled"},
	CreateNote: "Connector names are unique across every connector type and group, connectors marked for deletion included, and must not start with ALK-. " +
		"While it saves, the backend reads each instance's connectionId from AWS in customerRegion with the instance credentialId (the id of an AWS credential created in the Alkira portal; no tool lists or creates credentials), so the Direct Connect connection must exist and be visible to it; vlan is required unless the AWS connection has a VLAN (a hosted connection), which is then used. " +
		"dcGatewayAsn is 64800-65000 or 4200000000-4294967294, the same in every instance and different from underlayAsn. " +
		"Each instance lists its segments in segmentOptions with a /24 loopbackSubnet: with tunnelProtocol GRE (the default), VXLAN or VXLAN_GPE, customerLoopbackIp, alkLoopbackIp1 and alkLoopbackIp2 are set; with IPSEC the backend allocates them from numOfCustomerLoopbackIps and tunnelCountPerCustomerLoopbackIp. " +
		"VXLAN needs vni and gatewayMacAddress on every instance and one segment per instance; 20LARGE needs VXLAN or VXLAN_GPE. " +
		"The BGP keys (bgpAuthKey, bgpAuthKeyAlkira) cannot be set through this tool. Provisioning it deploys billable resources. " +
		"billingTags is sent only while the tenant feature BILLING-TAGS is on, scaleGroupId only while SUPPORT-SCALE-GROUP-FOR-DX-CONNECTOR is on (the backend also needs SCALE-GROUPS-FOR-USER-AND-SITES), loopbackPrefixes only while RELEASE-CUSTOM-SCALE-OPTIONS is on.",
	UpdateNote: "Refused with no request while the connector is FAILED, SCHEDULED or IN_PROGRESS, or while the network is FAILED and the connector was provisioned before. An id of another connector type is refused. " +
		"A patch on instances must list every stored instance by its id; an instance without id is added, and remove.instances drops a stored instance. " +
		"The stored BGP keys of each kept instance are resent; a patch cannot set or change them. After the first provision tunnelProtocol cannot change. The backend reads every connection from AWS again on each save.",
	DeleteNote: "A connector used by other objects is refused as in use. The backend removes the connector's implicit group, named after the connector, and releases its loopback IPs with it.",
}
