// Package schemagen builds JSON Schemas for write tool input from the client-go
// `ak` field rules (tenant.FieldRules). Read-only and secret fields are left
// out; a patch schema has no required fields (BC-R12) and no top-level `set`
// lists, which change only through list edits.
package schemagen

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	ak "github.com/alkiranet/client-go/tenant"
)

// Mode selects the create or patch form of a schema.
type Mode int

const (
	Create Mode = iota
	Patch
)

// Object returns the schema of an object with the given field rules.
func Object(rules []ak.FieldRule, mode Mode) map[string]any {
	return objectSchema(rules, mode, "")
}

// objectSchema builds the schema of an object at path ("" at the top level).
func objectSchema(rules []ak.FieldRule, mode Mode, path string) map[string]any {
	props := map[string]any{}
	var required []string
	for _, r := range rules {
		if r.Secret || r.ReadOnly && !nestedID(r, mode, path) {
			continue
		}
		props[r.Name] = Field(r, mode)
		if r.Required && mode == Create {
			required = append(required, r.Name)
		}
	}
	s := map[string]any{"type": "object", "properties": props, "additionalProperties": false}
	if required != nil {
		s["required"] = required
	}
	return s
}

// PatchObject returns the schema of an update patch: the patch form of Object
// without the top-level `set` fields.
func PatchObject(rules []ak.FieldRule) map[string]any {
	s := Object(rules, Patch)
	for _, r := range rules {
		if r.Set {
			delete(s["properties"].(map[string]any), r.Name)
		}
	}
	return s
}

// ListEdits returns the schema of the add or remove argument of an update:
// one array per top-level `set` field; remove also takes element ids of the
// top-level idMatched arrays. Removed items skip the element rules, so a
// stored value that breaks today's rules can still be removed. It returns
// nil when no field takes the edit.
func ListEdits(rules []ak.FieldRule, remove bool, idMatched []string) map[string]any {
	props := map[string]any{}
	for _, r := range rules {
		switch {
		case r.ReadOnly || r.Immutable || r.Secret:
		case r.Set:
			item := r
			if remove {
				item = ak.FieldRule{Name: r.Name, Type: r.Type, Fields: r.Fields}
			}
			s := typeSchema(item, r.Type, Create)
			s["minItems"] = 1
			props[r.Name] = s
		case remove && slices.Contains(idMatched, r.Name):
			props[r.Name] = map[string]any{
				"type": "array", "minItems": 1, "items": map[string]any{"type": []string{"string", "integer"}},
				"description": "Ids of the elements to remove.",
			}
		}
	}
	if len(props) == 0 {
		return nil
	}
	return map[string]any{"type": "object", "properties": props, "additionalProperties": false}
}

// Field returns the schema of one field.
func Field(r ak.FieldRule, mode Mode) map[string]any {
	s := typeSchema(r, r.Type, mode)
	var notes []string
	if r.Immutable {
		notes = append(notes, "Cannot be changed after create.")
	}
	if r.Flag != "" {
		notes = append(notes, fmt.Sprintf("Applies only when the tenant feature %s is enabled.", r.Flag))
	}
	if r.Ranges != nil {
		notes = append(notes, "Allowed ranges: "+rangesText(r.Ranges)+".")
	}
	if r.MaxLen > 0 {
		notes = append(notes, fmt.Sprintf("At most %d UTF-16 code units.", r.MaxLen))
	}
	if notes != nil {
		s["description"] = strings.Join(notes, " ")
	}
	return s
}

func typeSchema(r ak.FieldRule, t reflect.Type, mode Mode) map[string]any {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 && t.Kind() == reflect.Slice && t.Name() == "" {
			return map[string]any{} // []byte / json.RawMessage: any JSON value
		}
		return map[string]any{"type": "array", "items": typeSchema(r, t.Elem(), mode)}
	case reflect.Map:
		return map[string]any{"type": "object", "additionalProperties": typeSchema(r, t.Elem(), mode)}
	case reflect.Struct:
		if r.Fields == nil {
			return map[string]any{"type": "object"}
		}
		return objectSchema(r.Fields, mode, r.Name+".")
	case reflect.Interface:
		return map[string]any{}
	}
	s := map[string]any{}
	kind := scalarType(t)
	if kind != "" {
		s["type"] = kind
	}
	if r.HasDefault {
		s["default"] = scalar(r.Default, kind)
	}
	if r.Enum != nil {
		vals := make([]any, len(r.Enum))
		for i, v := range r.Enum {
			vals[i] = scalar(v, kind)
		}
		s["enum"] = vals
	}
	if r.MaxLen > 0 {
		s["maxLength"] = r.MaxLen
	}
	if r.Charset != "" {
		if p, ok := ak.CharsetPattern(r.Charset); ok {
			s["pattern"] = p
		}
	}
	if r.Min != nil {
		s["minimum"] = *r.Min
	}
	if r.Max != nil {
		s["maximum"] = *r.Max
	}
	if r.Ranges != nil {
		lo, hi := r.Ranges[0].Lo, r.Ranges[0].Hi
		for _, rg := range r.Ranges[1:] {
			lo, hi = min(lo, rg.Lo), max(hi, rg.Hi)
		}
		s["minimum"], s["maximum"] = lo, hi
	}
	return s
}

func scalarType(t reflect.Type) string {
	if t == reflect.TypeFor[json.Number]() {
		return "number"
	}
	switch k := t.Kind(); {
	case k == reflect.String:
		return "string"
	case k == reflect.Bool:
		return "boolean"
	case k >= reflect.Int && k <= reflect.Uint64:
		return "integer"
	case k == reflect.Float32 || k == reflect.Float64:
		return "number"
	}
	return ""
}

// scalar converts a tag value to the JSON type of its field; client-go has
// already checked that it parses.
func scalar(v, kind string) any {
	switch kind {
	case "integer":
		n, _ := strconv.ParseInt(v, 10, 64)
		return n
	case "number":
		f, _ := strconv.ParseFloat(v, 64)
		return f
	case "boolean":
		b, _ := strconv.ParseBool(v)
		return b
	}
	return v
}

func rangesText(rs []ak.Range) string {
	parts := make([]string, len(rs))
	for i, r := range rs {
		parts[i] = fmt.Sprintf("%d..%d", r.Lo, r.Hi)
	}
	return strings.Join(parts, ", ")
}
