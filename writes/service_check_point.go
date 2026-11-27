package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ServiceCheckPoint is the Check Point firewall service write family (TPS /chkp-fw-services).
var ServiceCheckPoint = handlers.WriteFamily[ak.ServiceCheckpointConfig]{
	Name:       "service_check_point",
	Noun:       "Check Point firewall service",
	Descriptor: ak.ServiceCheckpointDescriptor(),
	CreateNote: "Provisioning it deploys billable firewall instances. " +
		"credentialId names a Check Point firewall credential, managementServer.credentialId a Check Point management server credential (AUTOMATED mode) and each instance credentialId a Check Point instance credential; create them in the Alkira portal first, since no tool lists or creates credentials. " +
		"managementServer.globalCidrListId names a global CIDR list of the same CXP holding /24 prefixes. " +
		"segmentOptions holds one entry, keyed by the service segment, with the single zone DEFAULT.",
	UpdateNote: "cxp cannot change after create. Check Point credentials cannot be updated in place: a new secret needs a new credential id. " +
		serviceInstancesNote + " " +
		"Without the tenant feature CHKP-FW-AUTOSCALE the instance count cannot change.",
	DeleteNote:   serviceDeleteNote,
	ImplicitNote: serviceCredentials,
}
