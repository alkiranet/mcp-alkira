package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ServiceFortinet is the Fortinet firewall service write family (TPS /ftnt-fw-services).
var ServiceFortinet = handlers.WriteFamily[ak.ServiceFortinetConfig]{
	Name:       "service_fortinet",
	Noun:       "Fortinet firewall service",
	Descriptor: ak.ServiceFortinetDescriptor(),
	CreateNote: "Provisioning it deploys billable firewall instances. " +
		"credentialId names a Fortinet firewall credential; with BRING_YOUR_OWN each instance credentialId names a Fortinet instance credential of the same licence type. Create them in the Alkira portal first, since no tool lists or creates credentials. " +
		"managementServer.segment must be one of segments. The backend stores each instance under its hostName.",
	UpdateNote: "managementServer.segment keeps its stored value. After the first provision name, cxp, licenseType, tunnelProtocol, version and credentialId cannot change. " +
		serviceInstancesNote,
	DeleteNote:   serviceDeleteNote,
	ImplicitNote: serviceCredentials,
}
