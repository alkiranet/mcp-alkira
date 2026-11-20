package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ConnectorInternet is the internet connector write family (TPS /internetconnectors).
var ConnectorInternet = handlers.WriteFamily[ak.ConnectorInternetConfig]{
	Name:       "connector_internet",
	Noun:       "internet connector",
	Descriptor: ak.ConnectorInternetDescriptor(),
	// TPS creates a connector's implicit group under the connector's name
	// (ConnectorServiceImpl.java:957-962) and removes it with the connector.
	Implicit: func(name string) []string { return []string{name} },
	CreateNote: "Connector names are unique across every connector type and group, connectors marked for deletion included, and must not start with ALK-. " +
		"segments holds exactly one segment name. billingTags is sent only while the tenant feature BILLING-TAGS is on.",
	UpdateNote: "Refused with no request while the connector is FAILED, SCHEDULED or IN_PROGRESS, or while the network is FAILED and the connector was provisioned before. " +
		"An id of another connector type is refused.",
	DeleteNote: "A connector used by internet applications, segment resources or NAT policies is refused as in use. The backend removes the connector's implicit group, named after the connector, with it. " +
		"A deployed connector is marked for deletion and keeps its name until the next provision.",
}
