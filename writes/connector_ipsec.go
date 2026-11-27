package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ConnectorIPsec is the IPsec connector write family (TPS /ipsecconnectors).
// Pre-shared keys and the BGP key are write-only create arguments, masked in
// every output.
var ConnectorIPsec = handlers.WriteFamily[ak.ConnectorIPsecConfig]{
	Name:       "connector_ipsec",
	Noun:       "IPsec connector",
	Descriptor: ak.ConnectorIPsecDescriptor(),
	// TPS creates a connector's implicit group under the connector's name
	// and removes it with the connector.
	Implicit: func(name string) []string { return []string{name} },
	CreateNote: "Connector names are unique across every connector type and group, connectors marked for deletion included, and must not start with ALK-. segments holds exactly one segment name. " +
		"Each site may set presharedKeys, one key per tunnel (at most two), as the portal's optional key fields do; a site without keys gets keys the backend generates. Keys are write-only: no output shows them, so read a generated key in the Alkira portal. " +
		"routingOptions.dynamicRouting.bgpAuthKeyAlkira is optional and write-only; without it BGP runs without authentication. A site with gatewayIpType DYNAMIC needs advanced.remoteAuthValue. " +
		"The options side must match vpnMode (unset means ROUTE_BASED): routingOptions for ROUTE_BASED, policyOptions for POLICY_BASED. " +
		"enabled and secondaryCXPs are sent only while the tenant feature CXP-FAILOVER-REGION is on, scaleGroupId only while SCALE-GROUPS-FOR-USER-AND-SITES is on. " +
		"routingOptions.dynamicRouting.localAsn takes effect only while the tenant feature RELEASE-EXPLICIT-LOCAL-ASN-FOR-BRANCH-CONNECTOR is on, and sites[].billingTags only while BILLING-TAGS is on; with the feature off the backend drops the value, and the result lists it as not kept.",
	UpdateNote: "Refused with no request while the connector is FAILED, SCHEDULED or IN_PROGRESS, or while the network is FAILED and the connector was provisioned before. An id of another connector type is refused. " +
		"Pre-shared keys and the BGP key cannot be set here; every stored key is resent unchanged, so the tunnels keep their keys. " +
		"A sites patch must list the id of every stored site; a site without an id is added as a new site and the backend generates its keys. A site with gatewayIpType DYNAMIC needs advanced.remoteAuthValue. " +
		"The options side must match vpnMode (unset means ROUTE_BASED): routingOptions for ROUTE_BASED, policyOptions for POLICY_BASED; a patch that leaves them mismatched is refused. " +
		"To switch mode, set vpnMode and the new side in one patch, for example {vpnMode: POLICY_BASED, policyOptions: {...}}; the stored options of the other side, BGP key included, are dropped. " +
		"Setting routingOptions.dynamicRouting to null moves the connector to static routing and drops the stored BGP key, which cannot be set again here. " +
		"enabled and secondaryCXPs are sent only while the tenant feature CXP-FAILOVER-REGION is on, scaleGroupId only while SCALE-GROUPS-FOR-USER-AND-SITES is on. " +
		"routingOptions.dynamicRouting.localAsn takes effect only while the tenant feature RELEASE-EXPLICIT-LOCAL-ASN-FOR-BRANCH-CONNECTOR is on, and sites[].billingTags only while BILLING-TAGS is on; with the feature off the backend drops the value, and the result lists it as not kept.",
	DeleteNote: "A connector used by segment resources, NAT policies or other objects is refused as in use. The backend removes the connector's implicit group, named after the connector, with it.",
}
