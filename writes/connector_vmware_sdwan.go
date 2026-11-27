package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ConnectorVmwareSdwan is the VMware SD-WAN connector write family (TPS /vmware-sdwan-connectors).
var ConnectorVmwareSdwan = handlers.WriteFamily[ak.ConnectorVmwareSdwanConfig]{
	Name:       "connector_vmware_sdwan",
	Noun:       "VMware SD-WAN connector",
	Descriptor: ak.ConnectorVmwareSdwanDescriptor(),
	Implicit:   sdwanImplicit,
	// lastEnabled moves without a user edit.
	Volatile: []string{"lastEnabled"},
	CreateNote: sdwanCreateNote + " Each instance credentialId names a VMware SD-WAN instance credential holding the activation code; create it in the Alkira portal first, since no tool lists or creates credentials. " +
		"orchestratorHostAddress is the http or https URL of the orchestrator. vmWareSdWanVRFMappings holds exactly one mapping. tunnelProtocol is IPSEC. " +
		"allowList is sent only while the tenant feature RELEASE-CONNECTOR-VMWARE-SDWAN-ALLOW-LIST is on; the backend refuses it otherwise.",
	UpdateNote: sdwanUpdateNote("instances") + " After an instance is provisioned its credentialId and hostName cannot change.",
	DeleteNote: sdwanDeleteNote,
}
