package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ServiceCiscoFTDv is the Cisco FTDv firewall service write family (TPS /cisco-ftdv-fw-services).
var ServiceCiscoFTDv = handlers.WriteFamily[ak.ServiceCiscoFTDvConfig]{
	Name:       "service_cisco_ftdv",
	Noun:       "Cisco FTDv firewall service",
	Descriptor: ak.ServiceCiscoFTDvDescriptor(),
	CreateNote: "Provisioning it deploys billable firewall instances. " +
		"Each instance credentialId names a Cisco FTDv instance credential; credentialId optionally names a Cisco FTDv FMC credential. Create them in the Alkira portal first, since no tool lists or creates credentials. " +
		"globalCidrListId names a global CIDR list of the same CXP tagged CISCO_FTDV_FW with prefixes of /24 or shorter. " +
		"managementServer.segment must be one of segments, and every segment needs at least one zone in segmentOptions.",
	UpdateNote: "After the first provision name, cxp, managementServer and globalCidrListId cannot change. " +
		serviceInstancesNote,
	DeleteNote:   serviceDeleteNote,
	ImplicitNote: serviceCredentials,
}
