package writes

import "strings"

// credentialRef describes every credential id field: no tool reads or
// writes credentials.
const credentialRef = "Id of a credential stored in the Alkira portal. No tool lists or creates credentials: create the credential in the portal and paste its id here."

// refText returns where the value of a reference field comes from, by field
// name and JSON type ("string", "integer", or "array<item type>"); "" for a
// field that is not a reference.
func refText(name, typ string) string {
	switch {
	case name == "cxp" && typ == "string":
		return "CXP name, for example USEAST-AZURE-1; cxp_get_all lists them."
	case name == "group" && typ == "string":
		return "Name of an existing group, not its id; group_get_by_name finds one and group_create makes one."
	case name == "segments" && typ == "array<string>":
		return "Segment names, not ids; segment_get_all lists them."
	case name == "segments" && typ == "array<integer>", name == "segmentIds" && typ == "array<integer>":
		return "Segment ids, not names; segment_get_all lists them."
	case (name == "segment" || name == "segmentName") && typ == "string":
		return "Segment name, not id; segment_get_all lists them."
	case name == "segmentId" && typ == "integer":
		return "Segment id, not name; segment_get_all lists them."
	case name == "billingTags" && typ == "array<integer>":
		return "Billing tag ids; billing_tag_get_all lists them."
	case strings.HasSuffix(strings.ToLower(name), "prefixlistid") && typ == "integer":
		return "Prefix list id; list_prefix_get_by_name finds one by name and list_prefix_create makes one."
	case strings.HasSuffix(strings.ToLower(name), "prefixlistids") && typ == "array<integer>":
		return "Prefix list ids; list_prefix_get_by_name finds one by name and list_prefix_create makes one."
	case strings.HasSuffix(strings.ToLower(name), "credentialid") && typ == "string":
		return credentialRef
	}
	return ""
}

// describeRefs prefixes the description of every reference field in the
// object schema s, at any depth, with where its value comes from.
func describeRefs(s map[string]any) {
	props, _ := s["properties"].(map[string]any)
	for name, v := range props {
		f, ok := v.(map[string]any)
		if !ok {
			continue
		}
		if text := refText(name, jsonType(f)); text != "" {
			if d, _ := f["description"].(string); d != "" {
				text += " " + d
			}
			f["description"] = text
		}
		describeNested(f)
	}
}

// describeNested walks into the objects a field schema holds.
func describeNested(f map[string]any) {
	if _, ok := f["properties"]; ok {
		describeRefs(f)
	}
	for _, k := range []string{"items", "additionalProperties"} {
		if sub, ok := f[k].(map[string]any); ok {
			describeNested(sub)
		}
	}
}

// jsonType is the type of a field schema: its "type", or "array<item type>".
func jsonType(f map[string]any) string {
	t, _ := f["type"].(string)
	if t != "array" {
		return t
	}
	items, _ := f["items"].(map[string]any)
	it, _ := items["type"].(string)
	return "array<" + it + ">"
}
