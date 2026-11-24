package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// GlobalCidrList is the global CIDR list write family (TPS /global-cidr-lists).
var GlobalCidrList = handlers.WriteFamily[ak.GlobalCidrListConfig]{
	Name:       "list_global_cidr",
	Noun:       "global CIDR list",
	Descriptor: ak.GlobalCidrListDescriptor(),
	// The backend answers every read with the current time as lastConfigUpdatedAt (GlobalCidrList.java:109-112).
	Volatile: []string{"lastConfigUpdatedAt"},
	CreateNote: "cxp names a CXP of the network. tags names the service types that may use the list. " +
		"The number of values is bounded by the tenant's resource limits (one per list by default), not by this schema; the backend also refuses values that overlap another global CIDR list.",
	DeleteNote: "A global CIDR list used by a service, such as Cisco FTDv, F5, Check Point or Infoblox, is refused as in use.",
}
