package writes

import (
	"encoding/json"
	"fmt"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// InternetApplication is the internet application write family (TPS /v1/.../internet-applications).
var InternetApplication = handlers.WriteFamily[ak.InternetApplicationConfig]{
	Name:       "internet_application",
	Noun:       "internet application",
	Descriptor: ak.InternetApplicationDescriptor(),
	// TPS deletes the application's inbound internet connector
	// (<cxp>-<segment>-INB-INT) with the last DEFAULT application that uses
	// it, and the ALK-INB-INT-GROUP group with the network's last such
	// connector (InternetApplicationServiceImpl.java:1302-1309, :1364-1378).
	ImplicitNote: func(_ string, raw []byte) string {
		var app struct {
			SegmentName          string `json:"segmentName"`
			InboundConnectorType string `json:"inboundConnectorType"`
		}
		if json.Unmarshal(raw, &app) != nil || app.SegmentName == "" || (app.InboundConnectorType != "" && app.InboundConnectorType != "DEFAULT") {
			return ""
		}
		return fmt.Sprintf("the inbound internet connector <cxp>-%s-INB-INT of segment %q on this application's CXP, if no other internet application uses it, "+
			"and the ALK-INB-INT-GROUP group, if that connector is the network's last inbound internet connector", app.SegmentName, app.SegmentName)
	},
	CreateNote: "connectorId is the connector that hosts the application and segmentName a segment on that connector. fqdnPrefix is required outside China CXPs. " +
		"A target of type INTERNAL_DNS takes policyFqdnListId and no value; other targets take a value. " +
		"The backend creates or reuses one inbound internet connector per segment and CXP for the application (<cxp>-<segment>-INB-INT); it is server-owned.",
	UpdateNote: "The stored object is resent whole. byoipId, publicIps, inboundConnectorType and biDirectionalAvailabilityZone cannot change after the first provision, nor connectorId and connectorType of an application that uses BYOIP. A FAILED application is refused with no request.",
	DeleteNote: "Deleting the last application of a segment and CXP also removes its INB-INT connector, and removing the network's last INB-INT connector removes the ALK-INB-INT-GROUP group. A deployed application is marked for deletion until the next provision.",
}
