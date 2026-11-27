package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ConnectorVersaSdwan is the Versa SD-WAN connector write family (TPS
// /versa-sdwan-connectors). The shared keys are required write-only create
// arguments: a key left out or null is stored as the backend's fixed default.
var ConnectorVersaSdwan = handlers.WriteFamily[ak.ConnectorVersaSdwanConfig]{
	Name:       "connector_versa_sdwan",
	Noun:       "Versa SD-WAN connector",
	Descriptor: ak.ConnectorVersaSdwanDescriptor(),
	Implicit:   sdwanImplicit,
	CreateNote: sdwanCreateNote + " localPublicSharedKey and remotePublicSharedKey are required and must not be blank: without them the backend stores its fixed default key. They are write-only: no output shows them, and an update resends the stored keys unchanged. " +
		"remoteId and localId are IKE identities in email form; versaControllerHost (required) names the Versa controller.",
	UpdateNote: sdwanUpdateNote("instances") + " The stored shared keys are resent unchanged; the output masks them.",
	DeleteNote: sdwanDeleteNote,
}
