// Package schemagen builds JSON Schemas for write tool input from the client-go
// `ak` field rules (tenant.FieldRules). Read-only, secret and expired fields
// are left out, apart from write-only secrets in a create; a patch schema has no required fields (BC-R12) and no
// top-level `set` lists, which change only through list edits.
package schemagen

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	ak "github.com/alkiranet/client-go/tenant"
)

// Generator builds schemas and validates input with the expired rule
// evaluated at one instant. A tool whose schema and validation share a
// Generator advertises exactly the fields its calls accept, however long the
// server runs.
type Generator struct {
	at time.Time
}

// At returns a Generator that evaluates the expired rule at t.
func At(t time.Time) Generator { return Generator{at: t} }

// expired reports a field whose backend expiry (the `expired` tag) has
// passed at g's instant: client-go refuses it on create and patch.
func (g Generator) expired(r ak.FieldRule) bool {
	return r.Expired != 0 && g.at.Unix() > r.Expired
}

// Mode selects the create or patch form of a schema.
type Mode int

const (
	Create Mode = iota
	Patch
)

// Object returns the schema of an object with the given field rules.
func (g Generator) Object(rules []ak.FieldRule, mode Mode) map[string]any {
	return g.objectSchema(rules, mode, "")
}

// objectSchema builds the schema of an object at path ("" at the top level).
func (g Generator) objectSchema(rules []ak.FieldRule, mode Mode, path string) map[string]any {
	props := map[string]any{}
	var required []string
	for _, r := range rules {
		if r.Secret && !(r.WriteOnly && mode == Create) || g.expired(r) || r.ReadOnly && !nestedID(r, mode, path) {
			continue
		}
		props[r.Name] = g.Field(r, mode)
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
func (g Generator) PatchObject(rules []ak.FieldRule) map[string]any {
	s := g.Object(rules, Patch)
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
func (g Generator) ListEdits(rules []ak.FieldRule, remove bool, idMatched []string) map[string]any {
	props := map[string]any{}
	for _, r := range rules {
		switch {
		case r.ReadOnly || r.Immutable || r.Secret || g.expired(r) && !remove:
		case r.Set:
			item := r
			if remove {
				item = ak.FieldRule{Name: r.Name, Type: r.Type, Fields: r.Fields}
			}
			s := g.typeSchema(item, r.Type, Create)
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
func (g Generator) Field(r ak.FieldRule, mode Mode) map[string]any {
	s := g.typeSchema(r, r.Type, mode)
	var notes []string
	if r.WriteOnly {
		notes = append(notes, "Write-only: sent with the create and never shown in any output; an update cannot change it.")
	}
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
	if r.Trimmed {
		notes = append(notes, "No leading or trailing space.")
	}
	if r.MaxItems > 0 {
		notes = append(notes, fmt.Sprintf("At most %d items.", r.MaxItems))
	}
	if r.MinItems > 0 {
		notes = append(notes, fmt.Sprintf("At least %d items.", r.MinItems))
	}
	for _, d := range r.Distinct {
		notes = append(notes, fmt.Sprintf("No two items may have the same %s.", d))
	}
	if notes != nil {
		s["description"] = strings.Join(notes, " ")
	}
	return s
}

// TrimmedPattern matches a string without a leading or trailing space, the
// client-go `trimmed` rule. It avoids lookaround, which RE2 lacks.
const TrimmedPattern = `^([^ ]([\s\S]*[^ ])?)?$`

func (g Generator) typeSchema(r ak.FieldRule, t reflect.Type, mode Mode) map[string]any {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 && t.Kind() == reflect.Slice && t.Name() == "" {
			return map[string]any{} // []byte / json.RawMessage: any JSON value
		}
		elem := r
		elem.MaxItems, elem.MinItems, elem.Distinct = 0, 0, nil
		s := map[string]any{"type": "array", "items": g.typeSchema(elem, t.Elem(), mode)}
		if r.MaxItems > 0 {
			s["maxItems"] = r.MaxItems
		}
		if r.MinItems > 0 {
			s["minItems"] = r.MinItems
		}
		return s
	case reflect.Map:
		return map[string]any{"type": "object", "additionalProperties": g.typeSchema(r, t.Elem(), mode)}
	case reflect.Struct:
		if r.Fields == nil {
			return map[string]any{"type": "object"}
		}
		return g.objectSchema(r.Fields, mode, r.Name+".")
	case reflect.Interface:
		return map[string]any{}
	}
	s := map[string]any{}
	kind := scalarType(t)
	if kind != "" {
		s["type"] = kind
	}
	// A patch leaves an omitted field as stored, so a default there would
	// invite a client to overwrite the stored value.
	if r.HasDefault && mode == Create {
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
	if r.Trimmed {
		if _, ok := s["pattern"]; ok {
			s["allOf"] = []any{map[string]any{"pattern": TrimmedPattern}}
		} else {
			s["pattern"] = TrimmedPattern
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
