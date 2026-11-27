package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ConnectorAwsTgw is the AWS TGW connector write family (TPS /aws-tgw-connectors).
var ConnectorAwsTgw = handlers.WriteFamily[ak.ConnectorAwsTgwConfig]{
	Name:       "connector_aws_tgw",
	Noun:       "AWS TGW connector",
	Descriptor: ak.ConnectorAwsTgwDescriptor(),
	Implicit:   func(name string) []string { return []string{name} },
	CreateNote: "Connector names are unique across every connector type and group, connectors marked for deletion included, must not start with ALK- and hold letters, digits, - and _ only. segments holds exactly one segment name. " +
		"awsTgwPeeringAttachmentId names an ACTIVE AWS TGW peering attachment (peering_gateway_aws_tgw_attachment_get_all lists them) of an AWS TGW on the connector's CXP whose ASN differs from the segment ASN; an attachment serves one connector. " +
		"staticRoutes holds at most one prefix list id. Provisioning it deploys billable resources. " +
		"billingTags is sent only while the tenant feature BILLING-TAGS is on, scaleGroupId only while RELEASE-SCALE-GROUP-FOR-TGW-CONNECTOR is on.",
	UpdateNote: "Refused with no request while the connector is FAILED, SCHEDULED or IN_PROGRESS, or while the network is FAILED and the connector was provisioned before. An id of another connector type is refused. " +
		"awsTgwPeeringAttachmentId cannot change; after the first provision cxp and segments cannot change either. The backend refuses every update once the attachment is deleted.",
	DeleteNote: "A connector used by NAT policies or other objects is refused as in use. The backend removes the connector's implicit group, named after the connector, with it. " +
		"A deployed connector keeps its attachment until the deletion is provisioned from the Alkira portal.",
}
