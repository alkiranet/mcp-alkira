package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ConnectorCiscoSdwan is the Cisco SD-WAN connector write family (TPS /ciscosdwaningresses).
var ConnectorCiscoSdwan = handlers.WriteFamily[ak.ConnectorCiscoSdwanConfig]{
	Name:       "connector_cisco_sdwan",
	Noun:       "Cisco SD-WAN connector",
	Descriptor: ak.ConnectorCiscoSdwanDescriptor(),
	Implicit:   sdwanImplicit,
	// lastEnabled and the validation state move without a user edit.
	Volatile: []string{"lastEnabled", "validationState"},
	CreateNote: sdwanCreateNote + " Each edge carries its cloud-init content in cloudInitFile; CSR and CAT8000V need org, vbond, uuid and otp in it. " +
		"credentialId names a Cisco SD-WAN credential (user name other than admin, and password) and sshKeyPairCredentialId an SSH key pair credential, VEDGE only; create them in the Alkira portal first, since no tool lists or creates credentials. " +
		"type defaults to VEDGE and cannot change after create. GRE and allowList need type CAT8000V, and an Azure CXP refuses GRE. " +
		"billingTags is sent only while the tenant feature BILLING-TAGS is on, allowList only while RELEASE-CONNECTOR-CISCO-SDWAN-ALLOW-LIST is on.",
	UpdateNote: sdwanUpdateNote("ciscoEdgeInfo") + " After the connector is provisioned the edges' cloud-init file, host and SSH key pair and the tunnelProtocol cannot change.",
	DeleteNote: sdwanDeleteNote,
}
