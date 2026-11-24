// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode/utf16"
)

// FieldRule is the parsed `ak` tag of one struct field, with the rules of
// nested struct fields (or of the struct elements of a slice) in Fields.
//
// Tag grammar: `ak:"<key>[=<value>],..."`. "," separates keys and "|"
// separates alternatives inside a value. Keys: required, default=<v>,
// readonly, immutable, maxlen=<n> (UTF-16 code units), charset=<name>
// (name, hyphen, description, noline), enum=<a|b>, min=<n>, max=<n>,
// range=<lo>..<hi>[|<lo>..<hi>], secret, set, flag=<FLAG>, src=<file:line>[|...].
// set marks a scalar list that changes only through add and remove: a merge
// patch cannot set it, because a patch array replaces the stored list whole.
// On a slice, maxlen, charset, enum, min, max and range apply to each element.
// required on a bool or number needs a pointer field, since a zero value
// cannot be told apart from an unset one.
// maxlen, charset, enum, min, max and range need src= on the same tag.
type FieldRule struct {
	Name       string       // JSON name
	GoName     string       // Go field name
	Type       reflect.Type // Go field type
	Required   bool
	ReadOnly   bool
	Immutable  bool
	Secret     bool
	Set        bool
	HasDefault bool
	Default    string
	MaxLen     int // 0: no limit
	Charset    string
	Enum       []string
	Min        *int64
	Max        *int64
	Ranges     []Range
	Flag       string
	Src        []string
	Fields     []FieldRule
}

// Range is one closed interval of a range= rule.
type Range struct {
	Lo, Hi int64
}

// charsets is the closed table of charset= names
// (V3 internal/service/validate/text.go:13-27).
var charsets = map[string]*regexp.Regexp{
	"name":        regexp.MustCompile(`^[a-zA-Z0-9.@_+\-% ]*$`),
	"hyphen":      regexp.MustCompile(`^[a-zA-Z0-9-]+$`),
	"description": regexp.MustCompile(`^[a-zA-Z0-9.:;@_+\-% ]*$`),
	"noline":      regexp.MustCompile("^[^\n\r\u0085  ]*$"),
}

// CharsetPattern returns the regular expression of a charset= name.
func CharsetPattern(name string) (string, bool) {
	re, ok := charsets[name]
	if !ok {
		return "", false
	}
	return re.String(), true
}

type rulesEntry struct {
	rules []FieldRule
	err   error
}

var rulesCache sync.Map // reflect.Type -> rulesEntry

// FieldRules returns the rules of struct type T's JSON fields in declaration
// order. It fails on an unknown tag key, an unknown charset name, range=
// together with min= or max=, a constraint key without src=, or a key that
// does not apply to the field type.
func FieldRules[T any]() ([]FieldRule, error) {
	return rulesOf(reflect.TypeFor[T]())
}

func rulesOf(t reflect.Type) ([]FieldRule, error) {
	if e, ok := rulesCache.Load(t); ok {
		return e.(rulesEntry).rules, e.(rulesEntry).err
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("ak tags: %s is not a struct", t)
	}
	rules, err := parseStruct(t, map[reflect.Type]bool{})
	rulesCache.Store(t, rulesEntry{rules, err})
	return rules, err
}

func parseStruct(t reflect.Type, seen map[reflect.Type]bool) ([]FieldRule, error) {
	if seen[t] {
		return nil, fmt.Errorf("ak tags: recursive type %s", t)
	}
	seen[t] = true
	defer delete(seen, t)

	var rules []FieldRule
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		// encoding/json promotes the exported fields of an embedded struct
		// value even when its type is unexported.
		if !f.IsExported() && !(f.Anonymous && f.Type.Kind() == reflect.Struct) {
			continue
		}
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "-" {
			continue
		}
		if f.Anonymous && name == "" && f.Type.Kind() == reflect.Pointer && indirect(f.Type).Kind() == reflect.Struct {
			// reflect.Value.FieldByName panics through a nil embedded pointer.
			return nil, fmt.Errorf("ak tags: %s.%s: embedded struct pointer; embed the struct by value", t.Name(), f.Name)
		}
		if f.Anonymous && name == "" && indirect(f.Type).Kind() == reflect.Struct {
			inner, err := parseStruct(indirect(f.Type), seen)
			if err != nil {
				return nil, err
			}
			rules = append(rules, inner...)
			continue
		}
		if name == "" {
			name = f.Name
		}
		r, err := parseTag(f.Tag.Get("ak"), f.Type)
		if err != nil {
			return nil, fmt.Errorf("ak tags: %s.%s: %w", t.Name(), f.Name, err)
		}
		r.Name, r.GoName, r.Type = name, f.Name, f.Type
		if st := structOf(f.Type); st != nil {
			if r.Fields, err = parseStruct(st, seen); err != nil {
				return nil, err
			}
		}
		if r.Secret {
			secretKeys.Store(strings.ToLower(name), true)
		}
		rules = append(rules, r)
	}
	return rules, nil
}

func parseTag(tag string, ft reflect.Type) (FieldRule, error) {
	var r FieldRule
	if tag == "" {
		return r, nil
	}
	seen := map[string]bool{}
	constraint := false
	for _, item := range strings.Split(tag, ",") {
		key, val, hasVal := strings.Cut(item, "=")
		if seen[key] {
			return r, fmt.Errorf("duplicate key %q", key)
		}
		seen[key] = true
		flagKey := key == "required" || key == "readonly" || key == "immutable" || key == "secret" || key == "set"
		if flagKey && hasVal || !flagKey && (!hasVal || val == "") {
			return r, fmt.Errorf("malformed key %q", item)
		}
		elem := elemKind(ft)
		switch key {
		case "required":
			r.Required = true
		case "readonly":
			r.ReadOnly = true
		case "immutable":
			r.Immutable = true
		case "secret":
			r.Secret = true
		case "set":
			if indirect(ft).Kind() != reflect.Slice || structOf(ft) != nil {
				return r, errors.New("set on a field that is not a list of scalars")
			}
			r.Set = true
		case "default":
			if err := checkScalar(val, elem); err != nil {
				return r, fmt.Errorf("default=%s: %w", val, err)
			}
			r.HasDefault, r.Default = true, val
		case "maxlen":
			n, err := strconv.Atoi(val)
			if err != nil || n <= 0 {
				return r, fmt.Errorf("maxlen=%s: not a positive integer", val)
			}
			if elem != reflect.String {
				return r, fmt.Errorf("maxlen on a non-string field")
			}
			r.MaxLen, constraint = n, true
		case "charset":
			if _, ok := charsets[val]; !ok {
				return r, fmt.Errorf("unknown charset %q", val)
			}
			if elem != reflect.String {
				return r, fmt.Errorf("charset on a non-string field")
			}
			r.Charset, constraint = val, true
		case "enum":
			r.Enum = strings.Split(val, "|")
			for _, v := range r.Enum {
				if err := checkScalar(v, elem); err != nil {
					return r, fmt.Errorf("enum value %q: %w", v, err)
				}
			}
			constraint = true
		case "min", "max":
			n, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return r, fmt.Errorf("%s=%s: not an integer", key, val)
			}
			if !isInt(elem) {
				return r, fmt.Errorf("%s on a non-integer field", key)
			}
			if key == "min" {
				r.Min = &n
			} else {
				r.Max = &n
			}
			constraint = true
		case "range":
			if !isInt(elem) {
				return r, fmt.Errorf("range on a non-integer field")
			}
			for _, part := range strings.Split(val, "|") {
				lo, hi, ok := strings.Cut(part, "..")
				l, err1 := strconv.ParseInt(lo, 10, 64)
				h, err2 := strconv.ParseInt(hi, 10, 64)
				if !ok || err1 != nil || err2 != nil || l > h {
					return r, fmt.Errorf("range=%s: malformed interval %q", val, part)
				}
				r.Ranges = append(r.Ranges, Range{l, h})
			}
			constraint = true
		case "flag":
			r.Flag = val
		case "src":
			r.Src = strings.Split(val, "|")
		default:
			return r, fmt.Errorf("unknown key %q", key)
		}
	}
	if r.Ranges != nil && (r.Min != nil || r.Max != nil) {
		return r, errors.New("range with min or max")
	}
	if r.Min != nil && r.Max != nil && *r.Min > *r.Max {
		return r, errors.New("min above max")
	}
	// A zero bool or number cannot be told apart from an unset one, so a
	// required one must be a pointer.
	if r.Required && (ft.Kind() == reflect.Bool || isInt(ft.Kind()) || ft.Kind() == reflect.Float32 || ft.Kind() == reflect.Float64) {
		return r, errors.New("required on a non-pointer bool or number; use a pointer")
	}
	if constraint && len(r.Src) == 0 {
		return r, errors.New("constraint key without src=")
	}
	return r, nil
}

func indirect(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

// elemKind is the kind a scalar rule checks: the field's, or its elements' for a slice.
func elemKind(t reflect.Type) reflect.Kind {
	t = indirect(t)
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		t = indirect(t.Elem())
	}
	return t.Kind()
}

// structOf returns the struct type of t, or of t's elements for a slice or map.
func structOf(t reflect.Type) reflect.Type {
	t = indirect(t)
	switch t.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		t = indirect(t.Elem())
	}
	if t.Kind() == reflect.Struct && t.PkgPath() != "encoding/json" {
		return t
	}
	return nil
}

func isInt(k reflect.Kind) bool {
	return k >= reflect.Int && k <= reflect.Uint64
}

func checkScalar(v string, k reflect.Kind) error {
	var err error
	switch {
	case k == reflect.String:
	case isInt(k):
		_, err = strconv.ParseInt(v, 10, 64)
	case k == reflect.Bool:
		_, err = strconv.ParseBool(v)
	case k == reflect.Float32 || k == reflect.Float64:
		_, err = strconv.ParseFloat(v, 64)
	default:
		err = fmt.Errorf("unsupported kind %s", k)
	}
	return err
}

// validateValue checks struct value v against rules. With patch set,
// required is not checked (BC-R12).
func validateValue(rules []FieldRule, v reflect.Value, patch bool, path string) error {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	var errs []error
	for _, r := range rules {
		fv := v.FieldByName(r.GoName)
		if !fv.IsValid() {
			continue
		}
		errs = append(errs, validateField(r, fv, patch, path+r.Name))
	}
	return errors.Join(errs...)
}

func validateField(r FieldRule, fv reflect.Value, patch bool, path string) error {
	if absent(fv) {
		if r.Required && !patch {
			return fmt.Errorf("%w: %s is required", ErrInvalidInput, path)
		}
		return nil
	}
	for fv.Kind() == reflect.Pointer || fv.Kind() == reflect.Interface {
		fv = fv.Elem()
	}
	switch fv.Kind() {
	case reflect.Slice, reflect.Array:
		var errs []error
		for i := 0; i < fv.Len(); i++ {
			ev := fv.Index(i)
			ep := fmt.Sprintf("%s[%d]", path, i)
			if r.Fields != nil {
				errs = append(errs, validateValue(r.Fields, ev, patch, ep+"."))
				continue
			}
			errs = append(errs, checkScalarRules(r, ev, ep))
		}
		return errors.Join(errs...)
	case reflect.Map:
		if r.Fields == nil {
			return nil
		}
		var errs []error
		iter := fv.MapRange()
		for iter.Next() {
			errs = append(errs, validateValue(r.Fields, iter.Value(), patch, fmt.Sprintf("%s[%v].", path, iter.Key())))
		}
		return errors.Join(errs...)
	case reflect.Struct:
		if r.Fields == nil {
			return nil
		}
		return validateValue(r.Fields, fv, patch, path+".")
	}
	return checkScalarRules(r, fv, path)
}

// absent reports whether v is a nil pointer, slice, map or interface, or a zero scalar.
func absent(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice:
		return v.IsNil()
	}
	return v.IsZero()
}

func checkScalarRules(r FieldRule, v reflect.Value, path string) error {
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	switch {
	case v.Kind() == reflect.String:
		s := v.String()
		if r.MaxLen > 0 {
			if n := len(utf16.Encode([]rune(s))); n > r.MaxLen {
				return fmt.Errorf("%w: %s is %d characters long, at most %d allowed", ErrInvalidInput, path, n, r.MaxLen)
			}
		}
		if r.Charset != "" && !charsets[r.Charset].MatchString(s) {
			return fmt.Errorf("%w: %s has characters outside the %s character set", ErrInvalidInput, path, r.Charset)
		}
		if r.Enum != nil && !contains(r.Enum, s) {
			return fmt.Errorf("%w: %s must be one of %s", ErrInvalidInput, path, strings.Join(r.Enum, ", "))
		}
	case isInt(v.Kind()):
		var n int64
		if v.CanInt() {
			n = v.Int()
		} else {
			n = int64(v.Uint())
		}
		if r.Min != nil && n < *r.Min {
			return fmt.Errorf("%w: %s must be at least %d", ErrInvalidInput, path, *r.Min)
		}
		if r.Max != nil && n > *r.Max {
			return fmt.Errorf("%w: %s must be at most %d", ErrInvalidInput, path, *r.Max)
		}
		if r.Ranges != nil && !inRanges(r.Ranges, n) {
			return fmt.Errorf("%w: %s must be in %s", ErrInvalidInput, path, rangesText(r.Ranges))
		}
		if r.Enum != nil && !contains(r.Enum, strconv.FormatInt(n, 10)) {
			return fmt.Errorf("%w: %s must be one of %s", ErrInvalidInput, path, strings.Join(r.Enum, ", "))
		}
	}
	return nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func inRanges(rs []Range, n int64) bool {
	for _, r := range rs {
		if n >= r.Lo && n <= r.Hi {
			return true
		}
	}
	return false
}

func rangesText(rs []Range) string {
	parts := make([]string, len(rs))
	for i, r := range rs {
		parts[i] = fmt.Sprintf("%d..%d", r.Lo, r.Hi)
	}
	return strings.Join(parts, " or ")
}

// validatePatch checks a merge patch against rules: unknown and readonly
// fields are refused, every set field must decode into its Go type with no
// unknown nested fields and pass its rules except required. It returns the
// decoded top-level keys.
func validatePatch(rules []FieldRule, patch []byte) (map[string]json.RawMessage, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(patch, &m); err != nil || m == nil {
		return nil, fmt.Errorf("%w: patch must be a JSON object", ErrInvalidInput)
	}
	byName := make(map[string]FieldRule, len(rules))
	for _, r := range rules {
		byName[r.Name] = r
	}
	var errs []error
	for k, raw := range m {
		r, ok := byName[k]
		switch {
		case !ok:
			errs = append(errs, fmt.Errorf("%w: unknown field %s", ErrInvalidInput, k))
			continue
		case r.ReadOnly:
			errs = append(errs, fmt.Errorf("%w: %s is read-only", ErrInvalidInput, k))
			continue
		case r.Secret:
			errs = append(errs, fmt.Errorf("%w: %s is a secret; the stored value is kept and cannot be set", ErrInvalidInput, k))
			continue
		case r.Set:
			errs = append(errs, fmt.Errorf("%w: %s cannot be set by a patch; use add or remove", ErrInvalidInput, k))
			continue
		case isNull(raw) && (r.Required || r.Immutable):
			errs = append(errs, fmt.Errorf("%w: %s cannot be removed", ErrInvalidInput, k))
			continue
		case isNull(raw):
			continue
		}
		pv := reflect.New(r.Type)
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		if err := dec.Decode(pv.Interface()); err != nil {
			errs = append(errs, fmt.Errorf("%w: %s: %v", ErrInvalidInput, k, err))
			continue
		}
		if v, err := decodeAny(raw); err == nil {
			errs = append(errs, nestedRefused(r, v, k))
		}
		errs = append(errs, validateField(r, pv.Elem(), true, k))
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return m, nil
}

// nestedRefused refuses a secret or read-only field set anywhere inside the
// patch value v of field r. A nested id is accepted: id-matched arrays pair
// elements by it.
func nestedRefused(r FieldRule, v any, path string) error {
	if r.Fields == nil {
		return nil
	}
	var errs []error
	for _, c := range children(r, v, path) {
		for _, f := range r.Fields {
			ev, ok := c.obj[f.Name]
			if !ok {
				continue
			}
			fp := c.path + "." + f.Name
			switch {
			case f.Secret:
				errs = append(errs, fmt.Errorf("%w: %s is a secret; the stored value is kept and cannot be set", ErrInvalidInput, fp))
			case f.ReadOnly && f.Name != "id":
				errs = append(errs, fmt.Errorf("%w: %s is read-only", ErrInvalidInput, fp))
			default:
				errs = append(errs, nestedRefused(f, ev, fp))
			}
		}
	}
	return errors.Join(errs...)
}

type child struct {
	obj  map[string]any
	path string
}

// children returns the objects r.Fields describe inside v: v itself for a
// struct field, its elements for a slice, its values for a map.
func children(r FieldRule, v any, path string) []child {
	var out []child
	switch x := v.(type) {
	case []any:
		for i, e := range x {
			if m, ok := e.(map[string]any); ok {
				out = append(out, child{m, fmt.Sprintf("%s[%d]", path, i)})
			}
		}
	case map[string]any:
		if indirect(r.Type).Kind() != reflect.Map {
			return []child{{x, path}}
		}
		for k, e := range x {
			if m, ok := e.(map[string]any); ok {
				out = append(out, child{m, fmt.Sprintf("%s[%s]", path, k)})
			}
		}
	}
	return out
}

// stripReadOnly deletes the read-only fields of obj at every level, in
// place. A nested id stays: id-matched arrays pair elements by it.
func stripReadOnly(rules []FieldRule, obj map[string]any, nested bool) {
	for _, r := range rules {
		v, ok := obj[r.Name]
		switch {
		case !ok:
		case r.ReadOnly && !(nested && r.Name == "id"):
			delete(obj, r.Name)
		case r.Fields != nil:
			for _, c := range children(r, v, "") {
				stripReadOnly(r.Fields, c.obj, true)
			}
		}
	}
}

func isNull(raw json.RawMessage) bool {
	return string(bytes.TrimSpace(raw)) == "null"
}
