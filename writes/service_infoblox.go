package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ServiceInfoblox is the Infoblox service write family (TPS /infoblox-services).
var ServiceInfoblox = handlers.WriteFamily[ak.ServiceInfobloxConfig]{
	Name:         "service_infoblox",
	Noun:         "Infoblox service",
	Descriptor:   ak.ServiceInfobloxDescriptor(),
	ImplicitNote: serviceGroupNote,
	CreateNote: "A service with NIOS instances needs gridMaster with name, gridMasterCredentialId and sharedSecretCredentialId; a NIOS_X-only service takes no gridMaster. " +
		"Its credentials (Infoblox, grid master, instance and join token credentials) are created in the Alkira portal; no tool lists or creates credentials. " +
		"globalCidrListId names a global CIDR list in the service's CXP holding one private /25. One Infoblox service per CXP and segment. Billable once provisioned.",
	UpdateNote: "Instances are matched by id: keep every stored instance in a patch on instances, or remove one with remove.instances. cxp and an instance's platform cannot change. The backend refuses a changed globalCidrListId once the service has left PENDING, and a changed hostName, or type of a NIOS instance, once that instance is provisioned. " +
		"Refused with no request while the service is FAILED, SCHEDULED or IN_PROGRESS, or while the network is FAILED and the service was provisioned before.",
	DeleteNote: "A service whose service group a policy or firewall zone uses, or that is another Infoblox service's anycast backup, is refused by the backend with its reason. A deployed service keeps its credentials until the deletion is provisioned from the Alkira portal.",
}
