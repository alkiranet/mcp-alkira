package schemagen

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode/utf16"

	ak "github.com/alkiranet/client-go/tenant"
)

// Validate checks a tool argument object against rules, the same rules the
// schema is built from: unknown, read-only and secret fields are refused,
// every value must have its field's JSON type and pass the tag constraints,
// and in Create mode required fields must be set. In Patch mode null removes
// a field (RFC 7386) and required is not checked. mcp-go does not validate
// arguments against a raw input schema, so this runs before any request.
func Validate(rules []ak.FieldRule, raw json.RawMessage, mode Mode) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return fmt.Errorf("not valid JSON: %v", err)
	}
	errs := object(rules, v, mode, "")
	slices.SortFunc(errs, func(a, b error) int { return strings.Compare(a.Error(), b.Error()) })
	return errors.Join(errs...)
}

func object(rules []ak.FieldRule, v any, mode Mode, path string) []error {
	m, ok := v.(map[string]any)
	if !ok {
		return []error{fmt.Errorf("%s must be an object", orRoot(path))}
	}
	byName := make(map[string]ak.FieldRule, len(rules))
	for _, r := range rules {
		byName[r.Name] = r
	}
	var errs []error
	for k, val := range m {
		r, ok := byName[k]
		switch {
		case !ok:
			errs = append(errs, fmt.Errorf("unknown field %s", path+k))
		// A nested id stays patchable, as in client-go: id-matched arrays pair
		// elements by it.
		case r.ReadOnly && !nestedID(r, mode, path):
			errs = append(errs, fmt.Errorf("%s is read-only", path+k))
		case r.Secret:
			errs = append(errs, fmt.Errorf("%s is a secret field and is not accepted", path+k))
		case r.Set && mode == Patch && path == "":
			errs = append(errs, fmt.Errorf("%s cannot be set by a patch; use add or remove", k))
		case val == nil && mode == Patch && (r.Required || r.Immutable):
			errs = append(errs, fmt.Errorf("%s cannot be removed", path+k))
		case val == nil:
		default:
			errs = append(errs, value(r, r.Type, val, mode, path+k)...)
		}
	}
	if mode == Create {
		for _, r := range rules {
			if r.Required && !r.ReadOnly && !r.Secret && m[r.Name] == nil {
				errs = append(errs, fmt.Errorf("%s is required", path+r.Name))
			}
		}
	}
	return errs
}

// nestedID reports a read-only id below the top level of a patch.
func nestedID(r ak.FieldRule, mode Mode, path string) bool {
	return mode == Patch && path != "" && r.Name == "id"
}

func value(r ak.FieldRule, t reflect.Type, v any, mode Mode, path string) []error {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		if t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 && t.Name() == "" {
			return nil
		}
		list, ok := v.([]any)
		if !ok {
			return []error{fmt.Errorf("%s must be an array", path)}
		}
		var errs []error
		for i, e := range list {
			if e == nil {
				errs = append(errs, fmt.Errorf("%s[%d] must not be null", path, i))
				continue
			}
			errs = append(errs, value(r, t.Elem(), e, mode, fmt.Sprintf("%s[%d]", path, i))...)
		}
		return errs
	case reflect.Map:
		m, ok := v.(map[string]any)
		if !ok {
			return []error{fmt.Errorf("%s must be an object", path)}
		}
		var errs []error
		for k, e := range m {
			if e != nil {
				errs = append(errs, value(r, t.Elem(), e, mode, path+"."+k)...)
			}
		}
		return errs
	case reflect.Struct:
		if r.Fields == nil {
			if _, ok := v.(map[string]any); !ok {
				return []error{fmt.Errorf("%s must be an object", path)}
			}
			return nil
		}
		return object(r.Fields, v, mode, path+".")
	case reflect.Interface:
		return nil
	}
	return scalarValue(r, scalarType(t), t.Kind(), v, path)
}

func scalarValue(r ak.FieldRule, kind string, k reflect.Kind, v any, path string) []error {
	var text string
	switch kind {
	case "string":
		s, ok := v.(string)
		if !ok {
			return []error{fmt.Errorf("%s must be a string", path)}
		}
		text = s
		if r.MaxLen > 0 {
			if n := len(utf16.Encode([]rune(s))); n > r.MaxLen {
				return []error{fmt.Errorf("%s is %d characters long, at most %d allowed", path, n, r.MaxLen)}
			}
		}
		if r.Charset != "" && !charset(r.Charset).MatchString(s) {
			return []error{fmt.Errorf("%s has characters outside the %s character set", path, r.Charset)}
		}
	case "boolean":
		b, ok := v.(bool)
		if !ok {
			return []error{fmt.Errorf("%s must be a boolean", path)}
		}
		text = strconv.FormatBool(b)
	case "integer":
		num, ok := v.(json.Number)
		n, err := num.Int64()
		if !ok || err != nil || (k >= reflect.Uint && n < 0) {
			return []error{fmt.Errorf("%s must be an integer", path)}
		}
		text = num.String()
		if r.Min != nil && n < *r.Min {
			return []error{fmt.Errorf("%s must be at least %d", path, *r.Min)}
		}
		if r.Max != nil && n > *r.Max {
			return []error{fmt.Errorf("%s must be at most %d", path, *r.Max)}
		}
		if r.Ranges != nil && !inRanges(r.Ranges, n) {
			return []error{fmt.Errorf("%s must be in %s", path, rangesText(r.Ranges))}
		}
	case "number":
		num, ok := v.(json.Number)
		if _, err := num.Float64(); !ok || err != nil {
			return []error{fmt.Errorf("%s must be a number", path)}
		}
		text = num.String()
	default:
		return []error{fmt.Errorf("%s has an unsupported type", path)}
	}
	if r.Enum != nil && !slices.Contains(r.Enum, text) {
		return []error{fmt.Errorf("%s must be one of %s", path, strings.Join(r.Enum, ", "))}
	}
	return nil
}

var charsetCache sync.Map // name -> *regexp.Regexp

func charset(name string) *regexp.Regexp {
	if re, ok := charsetCache.Load(name); ok {
		return re.(*regexp.Regexp)
	}
	p, _ := ak.CharsetPattern(name)
	re := regexp.MustCompile(p)
	charsetCache.Store(name, re)
	return re
}

func inRanges(rs []ak.Range, n int64) bool {
	for _, r := range rs {
		if n >= r.Lo && n <= r.Hi {
			return true
		}
	}
	return false
}

func orRoot(path string) string {
	if path == "" {
		return "the value"
	}
	return strings.TrimSuffix(path, ".")
}
