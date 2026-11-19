// Package testfamily is a synthetic resource family that exercises the write
// framework in tests and in binaries built with -tags writetest. Production
// builds do not register it.
package testfamily

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// Widget is the write type of the test family.
type Widget struct {
	ID          *int64       `json:"id,omitempty" ak:"readonly"`
	Name        *string      `json:"name,omitempty" ak:"required,maxlen=20,charset=name,src=testfamily.go:15"`
	Description *string      `json:"description,omitempty" ak:"maxlen=40,src=testfamily.go:16"`
	Size        *string      `json:"size,omitempty" ak:"enum=SMALL|LARGE,default=SMALL,flag=TEST-SIZE,src=testfamily.go:17"`
	Asn         *int64       `json:"asn,omitempty" ak:"range=65001..65514|2100000000..2147483647,default=65514,flag=TEST-ASN,src=testfamily.go:18"`
	Count       *int64       `json:"count,omitempty" ak:"min=0,max=10,src=testfamily.go:19"`
	Enabled     *bool        `json:"enabled,omitempty" ak:"default=true"`
	Region      *string      `json:"region,omitempty" ak:"immutable"`
	Fast        *bool        `json:"fast,omitempty" ak:"flag=TEST-FAST-PATH"`
	Psk         *string      `json:"psk,omitempty" ak:"secret"`
	Tags        []string     `json:"tags,omitempty"`
	Sites       []WidgetSite `json:"sites,omitempty"`
	Members     []string     `json:"members,omitempty" ak:"set,charset=name,src=testfamily.go:25"`
	State       *string      `json:"state,omitempty" ak:"readonly"`
}

// WidgetSite is a nested element with its own secret.
type WidgetSite struct {
	ID   *int64  `json:"id,omitempty"`
	Name *string `json:"name,omitempty" ak:"required"`
	Key  *string `json:"key,omitempty" ak:"secret"`
}

// Family is the test family: provisionable, marked rows once deployed, an
// implicit group per widget.
var Family = handlers.WriteFamily[Widget]{
	Name: "widget",
	Noun: "widget",
	Descriptor: ak.Descriptor{
		Collection:         "widgets",
		Provisionable:      true,
		MarkedRowLifecycle: true,
		DocState:           ak.DocStateMissingIsAdded,
		Uniqueness:         ak.Uniqueness{Key: "name", Scope: ak.ScopeCollection},
		StateGuard:         ak.StateGuardFailedOnly,
		IDMatchedArrays:    []string{"sites"},
		Duplicate:          []ak.Signature{{Status: 409, Code: "AK000001"}},
		InUse:              []ak.Signature{{Status: 409, FieldCode: "widget.00000010"}},
		// fast and size are left out while their flag is off; asn is refused.
		Flags: []ak.FlagRule{
			{Flag: "TEST-FAST-PATH", Effect: ak.FlagIncludeField, Field: "fast", UISource: "testfamily.go:21"},
			{Flag: "TEST-SIZE", Effect: ak.FlagIncludeField, Field: "size", UISource: "testfamily.go:16"},
			{Flag: "TEST-ASN", Effect: ak.FlagIncludeField, Field: "asn", Refuse: true, UISource: "testfamily.go:17"},
		},
		AsyncCapable: true,
	},
	Implicit: func(name string) []string { return []string{name + "-Group"} },
	Volatile: []string{"lastSeenAt"},
}
