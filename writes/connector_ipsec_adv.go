package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ConnectorIPsecAdv is the advanced IPsec connector write family (TPS
// /adv-ipsec-connectors). Tunnel pre-shared keys and the BGP key are
// write-only create arguments, masked in every output.
var ConnectorIPsecAdv = handlers.WriteFamily[ak.ConnectorIPsecAdvConfig]{
	Name:       "connector_ipsec_adv",
	Noun:       "advanced IPsec connector",
	Descriptor: ak.ConnectorIPsecAdvDescriptor(),
	// TPS creates a connector's implicit group under the connector's name
	// and removes it with the connector.
	Implicit: func(name string) []string { return []string{name} },
	CreateNote: "Connector names are unique across every connector type and group, connectors marked for deletion included, and must not start with ALK-. segment is one segment name. " +
		"Each gateway has tunnelsPerGateway tunnels; every tunnel carries cxpEnd (public and overlay IP reservation ids) and customerEnd. " +
		"Each tunnel may set preSharedKey, as the portal's optional key field does; a tunnel without a key gets one the backend generates. Keys are write-only: no output shows them, so read a generated key in the Alkira portal. " +
		"routingOptions.dynamicRouting.bgpAuthKeyAlkira is optional and write-only. A tunnel with advanced needs profileId (an IPsec tunnel profile id) and advanced.ikeVersion, remoteAuthType and remoteAuthValue. At least one gateway is ACTIVE (unset haMode means ACTIVE) and at most one STANDBY. " +
		"vpnMode is ROUTE_BASED only; routingOptions holds staticRouting (with prefixListId), dynamicRouting (with customerGwAsn) or both. " +
		"enabled is sent only while the tenant feature CXP-FAILOVER-REGION is on, scaleGroupId only while SCALE-GROUPS-FOR-USER-AND-SITES is on. " +
		"routingOptions.dynamicRouting.localAsn takes effect only while the tenant feature RELEASE-EXPLICIT-LOCAL-ASN-FOR-BRANCH-CONNECTOR is on; with it off the backend drops the value, and the result lists it as not kept.",
	UpdateNote: "Refused with no request while the connector is FAILED, SCHEDULED or IN_PROGRESS, or while the network is FAILED and the connector was provisioned before. An id of another connector type is refused. " +
		"Tunnel pre-shared keys and the BGP key cannot be set here; every stored key is resent unchanged, so the tunnels keep their keys. " +
		"A gateways patch must list the id of every stored gateway, and each gateway's tunnels the id of every stored tunnel; an unknown id is refused. A gateway or tunnel without an id is added as new, and the backend generates a key for each new tunnel. " +
		"Every gateway with tunnels has tunnelsPerGateway tunnels. Every tunnel carries both cxpEnd and customerEnd, or none does. A tunnel with advanced needs profileId (an IPsec tunnel profile id) and advanced.ikeVersion, remoteAuthType and remoteAuthValue. At least one gateway is ACTIVE (unset haMode means ACTIVE) and at most one STANDBY. " +
		"vpnMode is ROUTE_BASED only; routingOptions holds staticRouting (with prefixListId), dynamicRouting (with customerGwAsn) or both. Setting routingOptions.dynamicRouting to null drops the stored BGP key, which cannot be set again here. " +
		"scaleGroupId cannot change once set. enabled is sent only while the tenant feature CXP-FAILOVER-REGION is on, scaleGroupId only while SCALE-GROUPS-FOR-USER-AND-SITES is on. " +
		"routingOptions.dynamicRouting.localAsn takes effect only while the tenant feature RELEASE-EXPLICIT-LOCAL-ASN-FOR-BRANCH-CONNECTOR is on; with it off the backend drops the value, and the result lists it as not kept.",
	DeleteNote: "A connector used by segment resources, NAT policies or other objects is refused as in use. The backend removes the connector's implicit group, named after the connector, with it.",
}
