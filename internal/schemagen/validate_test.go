package schemagen

import (
	"encoding/json"
	"strings"
	"testing"

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
	if err := Validate(rules, json.RawMessage(`{"sites":[{"id":1,"name":"a"}]}`), Patch); err != nil {
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
		if err := Validate(rules, json.RawMessage(c.raw), c.mode); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v, want %q", c.raw, err, c.want)
		}
	}
	patch := PatchObject(rules)["properties"].(map[string]any)
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
	create := Object(rules, Create)["properties"].(map[string]any)["sites"].(map[string]any)["items"].(map[string]any)["properties"].(map[string]any)
	if _, ok := create["id"]; ok {
		t.Error("create schema has the element id")
	}
}
