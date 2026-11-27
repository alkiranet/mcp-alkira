package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ServicePan is the PAN firewall service write family (TPS /panfwservices).
var ServicePan = handlers.WriteFamily[ak.ServicePanConfig]{
	Name:       "service_pan",
	Noun:       "PAN firewall service",
	Descriptor: ak.ServicePanDescriptor(),
	CreateNote: "Provisioning it deploys billable firewall instances. " +
		"credentialId names a PAN firewall credential and each instance credentialId a PAN instance credential; create them in the Alkira portal first, since no tool lists or creates credentials. " +
		"segments and managementSegment are segment ids, and segments must include managementSegment. " +
		"Version 9.1 and later needs registrationCredentialId. A BRING_YOUR_OWN licence is checked with Palo Alto Networks when the save is validated. " +
		"With the tenant feature RELEASE-ASYNC-NETWORK-ENTITY-VALIDATION-PHASE-1 the save is validated after it is stored; the output reports the validation state.",
	UpdateNote: "cxp, licenseType, subLicenseType, managementSegment, credentialId, panoramaEnabled, panoramaDeviceGroup, panoramaTemplate and the SCM fields cannot change after create. " +
		serviceInstancesNote,
	DeleteNote:   serviceDeleteNote,
	ImplicitNote: serviceCredentials,
}
