package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ConnectorFortinetSdwan is the Fortinet SD-WAN connector write family (TPS /ftnt-sdwan-connectors).
var ConnectorFortinetSdwan = handlers.WriteFamily[ak.ConnectorFortinetSdwanConfig]{
	Name:       "connector_fortinet_sdwan",
	Noun:       "Fortinet SD-WAN connector",
	Descriptor: ak.ConnectorFortinetSdwanDescriptor(),
	Implicit:   sdwanImplicit,
	CreateNote: sdwanCreateNote + " Each edge credentialId names a Fortinet SD-WAN instance credential, created in the Alkira portal, holding userName (admin), password, licenseType and, for BRING_YOUR_OWN, licenseKey; " +
		"create one per edge first: it is single-use. The edge licenseType must match the credential's. " +
		"A BRING_YOUR_OWN edge needs serialNumber; the backend drops serialNumber on a PAY_AS_YOU_GO edge. An Azure CXP refuses GRE.",
	UpdateNote: sdwanUpdateNote("instances") + " remove.instances drops an edge; a new hostName is a new edge with a new id and needs a new credential.",
	DeleteNote: sdwanDeleteNote + " Edge credentials are single-use: a new connector needs new ones, and the credentials of a never-deployed connector stay until deleted in the Alkira portal.",
}
