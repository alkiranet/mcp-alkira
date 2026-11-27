package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// Group is the group write family (TPS /groups). Group and connector names
// share one name space.
var Group = handlers.WriteFamily[ak.GroupConfig]{
	Name:       "group",
	Noun:       "group",
	Descriptor: ak.GroupDescriptor(),
	CreateNote: "Group names and connector names share one name space, so a name used by any connector is refused too.",
	UpdateNote: "Only groups with type EXPLICIT can be changed; the groups the backend creates (connector groups, <segment>-ALL-Group) are refused with no request.",
	DeleteNote: "Only groups with type EXPLICIT can be deleted; the groups the backend creates (connector groups, <segment>-ALL-Group) are refused with no request and no confirmation code. " +
		"A group used by a connector, policy, segment resource or firewall zone is refused as in use.",
}
