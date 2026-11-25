package writes

// sdwanCreateNote is the create fact the five SD-WAN connector families share.
const sdwanCreateNote = "Connector names are unique across every connector type and group, connectors marked for deletion included, and must not start with ALK-. " +
	"Provisioning it deploys billable instances. The VRF mappings name segments by id, and a gateway or customer ASN must differ from the segment's ASN. " +
	"version must name an image the CXP offers for this connector type and size."

// sdwanUpdateNote is the update fact the SD-WAN families share; list is the
// array of instances or edges.
func sdwanUpdateNote(list string) string {
	return "Refused with no request while the connector is FAILED, SCHEDULED or IN_PROGRESS, or while the network is FAILED and the connector was provisioned before. " +
		"An id of another connector type is refused. A patch on " + list + " must list every stored entry by its id; an entry without id is added, and remove." + list +
		" drops a stored one. The VRF mappings cannot change after the connector is provisioned."
}

// sdwanDeleteNote is the delete fact the SD-WAN families share.
const sdwanDeleteNote = "A connector used by segment resources, policies or other objects is refused as in use. " +
	"The backend removes the connector's implicit group, named after the connector, with it."

// sdwanImplicit is the implicit group the backend creates and deletes with a connector.
func sdwanImplicit(name string) []string { return []string{name} }
