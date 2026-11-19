package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// UdrList is the UDR list write family (TPS /udr-lists).
var UdrList = handlers.WriteFamily[ak.UdrListConfig]{
	Name:       "list_udr",
	Noun:       "UDR list",
	Descriptor: ak.UdrListDescriptor(),
	CreateNote: "cloudProvider is AZURE and each route's nextHopType is INTERNET, the only values the backend accepts. " +
		"The number of routes is bounded by the tenant's resource limits (16 by default), not by this schema.",
	UpdateNote: "A patch on udrs replaces the whole route list: send every route to keep, and repeat prefix and nextHopType in each item, or the backend refuses the update.",
	DeleteNote: "A UDR list used by an Azure VNET connector is refused as in use, and stays in use for about four minutes after that connector is deleted; the delete is not retried.",
}
