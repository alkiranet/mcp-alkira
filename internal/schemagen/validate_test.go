package schemagen

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	ak "github.com/alkiranet/client-go/tenant"
)

type site struct {
	ID    *int64  `json:"id,omitempty" ak:"readonly"`
	Name  *string `json:"name,omitempty"`
	State *string `json:"state,omitempty" ak:"readonly"`
}

type family struct {
	ID    *int64  `json:"id,omitempty" ak:"readonly"`
	Name  *string `json:"name,omitempty"`
	Sites []site  `json:"sites,omitempty"`
}

// A nested read-only id is accepted in a patch, as client-go accepts it: an
// id-matched array pairs elements by it. Other read-only fields, the top-level
// id and a nested id on create stay refused.
func TestValidate_NestedReadOnlyID(t *testing.T) {
	rules, err := ak.FieldRules[family]()
	if err != nil {
		t.Fatal(err)
	}
	if err := At(time.Now()).Validate(rules, json.RawMessage(`{"sites":[{"id":1,"name":"a"}]}`), Patch); err != nil {
		t.Errorf("nested id in a patch: %v", err)
	}
	for _, c := range []struct {
		raw  string
		mode Mode
		want string
	}{
		{`{"id":1}`, Patch, "id is read-only"},
		{`{"sites":[{"id":1,"state":"X"}]}`, Patch, "sites[0].state is read-only"},
		{`{"name":"n","sites":[{"id":1}]}`, Create, "sites[0].id is read-only"},
	} {
		if err := At(time.Now()).Validate(rules, json.RawMessage(c.raw), c.mode); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v, want %q", c.raw, err, c.want)
		}
	}
	patch := At(time.Now()).PatchObject(rules)["properties"].(map[string]any)
	if _, ok := patch["id"]; ok {
		t.Error("patch schema has the top-level id")
	}
	elem := patch["sites"].(map[string]any)["items"].(map[string]any)["properties"].(map[string]any)
	if _, ok := elem["id"]; !ok {
		t.Error("patch schema lacks the element id")
	}
	if _, ok := elem["state"]; ok {
		t.Error("patch schema has the element state")
	}
	create := At(time.Now()).Object(rules, Create)["properties"].(map[string]any)["sites"].(map[string]any)["items"].(map[string]any)["properties"].(map[string]any)
	if _, ok := create["id"]; ok {
		t.Error("create schema has the element id")
	}
}

type expiring struct {
	Name *string `json:"name,omitempty" ak:"required"`
	Old  *int64  `json:"old,omitempty" ak:"expired=1780272000,src=Test.java:1"`
	Late *int64  `json:"late,omitempty" ak:"expired=4102444800,src=Test.java:2"`
}

// A field whose expired date has passed is left out of the create and patch
// schemas and refused, as client-go refuses it; one not yet expired stays.
func TestExpiredField(t *testing.T) {
	g := At(time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC))
	rules, err := ak.FieldRules[expiring]()
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []map[string]any{g.Object(rules, Create), g.PatchObject(rules)} {
		props := s["properties"].(map[string]any)
		if _, ok := props["old"]; ok {
			t.Error("schema has the expired field")
		}
		if _, ok := props["late"]; !ok {
			t.Error("schema lacks the field not yet expired")
		}
	}
	for _, mode := range []Mode{Create, Patch} {
		if err := g.Validate(rules, json.RawMessage(`{"name":"n","old":1}`), mode); err == nil || !strings.Contains(err.Error(), "old is no longer accepted by the backend") {
			t.Errorf("mode %d: %v", mode, err)
		}
		if err := g.Validate(rules, json.RawMessage(`{"name":"n","late":1}`), mode); err != nil {
			t.Errorf("mode %d: %v", mode, err)
		}
	}
	if _, ok := At(time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)).Object(rules, Create)["properties"].(map[string]any)["old"]; !ok {
		t.Error("field left out before its expiry")
	}
}
