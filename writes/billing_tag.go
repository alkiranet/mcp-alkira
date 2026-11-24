package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// BillingTag is the billing tag write family: tenant-level /tags on
// tenant-billing, not provisioned.
var BillingTag = handlers.WriteFamily[ak.BillingTagConfig]{
	Name:       "billing_tag",
	Noun:       "billing tag",
	Descriptor: ak.BillingTagDescriptor(),
	CreateNote: "Billing tags belong to the tenant, not to a network, and need the tenant features BILLING-VIEW and BILLING-TAGS. The change applies at once; no provision is needed. A tenant at its billing tag limit is refused.",
	UpdateNote: "The change applies at once. The backend checks the billing tag limit on every save, so a tenant at its limit cannot update a tag either.",
	DeleteNote: "A billing tag attached to a connector or service is refused as in use.",
}
