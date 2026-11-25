package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ConnectorAzureVnet is the Azure VNet connector write family (TPS /azurevnetconnectors).
var ConnectorAzureVnet = handlers.WriteFamily[ak.ConnectorAzureVnetConfig]{
	Name:       "connector_azure_vnet",
	Noun:       "Azure VNet connector",
	Descriptor: ak.ConnectorAzureVnetDescriptor(),
	Implicit:   func(name string) []string { return []string{name} },
	Volatile:   []string{"lastEnabled", "validationState"},
	CreateNote: "Connector names are unique across every connector type and group, connectors marked for deletion included, and must not start with ALK-. segments holds exactly one segment name. " +
		"vnetId is the Azure resource id of the VNet. resourceGroupName is optional and shown in the portal only; when set, it must be the resource group in vnetId. credentialId is the id of a stored Azure VNet credential for the VNet's subscription; create it in the Alkira portal first, since no tool lists or creates credentials. " +
		"The backend reads the VNet from Azure while it saves; connectionMode defaults to VNET_GATEWAY, as the portal sends; set VNET_PEERING for a peering connector. A VNET_PEERING connector cannot be disabled. " +
		"billingTags is sent only while the tenant feature BILLING-TAGS is on, scaleGroupId only while SCALE-GROUPS-FOR-CLOUD-CONNECTORS is on, cxpPeeringGatewayId only while ATH-ON-AWS-CXP is on.",
	UpdateNote: "Refused with no request while the connector is FAILED, SCHEDULED or IN_PROGRESS, or while the network is FAILED and the connector was provisioned before. An id of another connector type is refused. " +
		"vnetId cannot change. In vnetRouting.importFromCXPOptions, setting cidrs or subnets drops the stored VNet-wide prefixListIds and routeImportMode, and the reverse. " +
		"A 409 asking to retry means the backend lost a race reading the VNet route tables; call the update again.",
	DeleteNote: "A connector used by segment resources, NAT policies, UDR lists or other objects is refused as in use. The backend removes the connector's implicit group, named after the connector, with it.",
}
