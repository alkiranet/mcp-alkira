package registry

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/alkiranet/mcp-alkira/handlers"
	"github.com/mark3labs/mcp-go/server"
)

// UX-02: create and update results of a provisionable family say the change
// is pending until it is provisioned from the portal.
func TestWrite_PendingText(t *testing.T) {
	e := newWriteEnv(t)
	r := e.call("widget_create", map[string]any{"body": map[string]any{"name": "w1"}})
	mustOK(t, r)
	if !strings.Contains(r.Text, `Created widget "w1" with id`) || !strings.Contains(r.Text, handlers.PendingCreate) {
		t.Errorf("create text:\n%s", r.Text)
	}
	stored(e)
	r = e.call("widget_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}})
	mustOK(t, r)
	if !strings.Contains(r.Text, `Updated widget "w7" (id 7). `+handlers.PendingChange) {
		t.Errorf("update text:\n%s", r.Text)
	}
}

// UX-08, LIVE-07: the preview names the connectors and services a stored
// segment lists, and
// otherwise says that only the confirmed delete checks for users.
func TestWrite_DeletePreviewInUse(t *testing.T) {
	s := specNamed("segment")
	e := newFamilyEnv(t, s)
	body := s.stored()
	body["connectors"] = []any{23501, 23502}
	body["services"] = []any{4101}
	e.portal.put(7, body)
	r := e.call("segment_delete", map[string]any{"id": "7"})
	previewCode(t, r)
	if !strings.Contains(r.Text, "In use by: connector ids 23501, 23502; service id 4101. The backend refuses this delete while they use it; nothing is deleted then.") {
		t.Errorf("preview:\n%s", r.Text)
	}
	// A segment only a service uses is refused too (SegmentServiceImpl.java:414-426).
	body = s.stored()
	body["services"] = []any{4101}
	e.portal.put(9, body)
	if r := e.call("segment_delete", map[string]any{"id": "9"}); !strings.Contains(r.Text, "In use by: service id 4101.") {
		t.Errorf("preview of a segment a service uses:\n%s", r.Text)
	}
	e.portal.put(8, s.stored())
	r = e.call("segment_delete", map[string]any{"id": "8"})
	if !strings.Contains(r.Text, "In use: only the confirmed delete checks whether another object uses it") {
		t.Errorf("preview without users:\n%s", r.Text)
	}
}

// LIVE-05: a code that does not match says how to get a new preview.
func TestWrite_ConfirmMismatchText(t *testing.T) {
	e := newWriteEnv(t)
	stored(e)
	code := previewCode(t, e.call("widget_delete", map[string]any{"id": "7"}))
	mustRefuse(t, e.call("widget_delete", map[string]any{"id": "999999991", "confirmCode": code}), handlers.RefusalConfirmMismatch,
		"The confirmation code does not match this object. Call widget_delete with id 999999991 and no confirmCode for a new preview and code")
}

// UX-02: a delete of a deployed object that the backend removes at once says
// the network keeps it until the deletion is provisioned.
func TestWrite_DeletePendingText(t *testing.T) {
	s := specNamed("segment")
	e := newFamilyEnv(t, s)
	body := s.stored()
	body["state"] = "SUCCESS"
	e.portal.put(7, body)
	code := previewCode(t, e.call("segment_delete", map[string]any{"id": "7"}))
	r := e.call("segment_delete", map[string]any{"id": "7", "confirmCode": code})
	mustOK(t, r)
	if !strings.Contains(r.Text, `Deleted segment "s7" (id 7). Pending: the deployed segment stays on the network until the deletion is provisioned from the Alkira portal.`) {
		t.Errorf("delete text:\n%s", r.Text)
	}
}

// toolSchemas returns the description and input schema of every tool srv lists.
func toolSchemas(t *testing.T, srv *server.MCPServer) map[string]struct {
	Description string
	Schema      map[string]any
} {
	t.Helper()
	var res struct {
		Tools []struct {
			Name        string         `json:"name"`
			Description string         `json:"description"`
			InputSchema map[string]any `json:"inputSchema"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(call(t, srv, "tools/list", nil), &res); err != nil {
		t.Fatal(err)
	}
	out := map[string]struct {
		Description string
		Schema      map[string]any
	}{}
	for _, tool := range res.Tools {
		out[tool.Name] = struct {
			Description string
			Schema      map[string]any
		}{tool.Description, tool.InputSchema}
	}
	return out
}

// prop returns the schema at a dotted path of properties, with [] for array items.
func prop(s map[string]any, path string) map[string]any {
	for _, p := range strings.Split(path, ".") {
		items := strings.HasSuffix(p, "[]")
		s, _ = s["properties"].(map[string]any)[strings.TrimSuffix(p, "[]")].(map[string]any)
		if items && s != nil {
			s, _ = s["items"].(map[string]any)
		}
		if s == nil {
			return nil
		}
	}
	return s
}

// UX-03, UX-07, UX-16: id and reference fields say where their value comes
// from, and every write family has a read tool that lists or finds ids.
func TestWrite_ReferenceDescriptions(t *testing.T) {
	tools := toolSchemas(t, allWriteServer(t))
	for _, c := range []struct{ tool, path, want string }{
		{"connector_aws_vpc_create", "body.cxp", "CXP name, for example USEAST-AZURE-1; cxp_get_all lists them. At most 128"},
		{"connector_aws_vpc_create", "body.group", "Name of an existing group, not its id; group_get_by_name finds one"},
		{"connector_aws_vpc_create", "body.segments", "Segment names, not ids; segment_get_all lists them."},
		{"connector_aws_vpc_create", "body.credentialId", "No tool lists or creates credentials: create the credential in the portal"},
		{"connector_aws_vpc_update", "patch.vpcRouting.importFromCXPOptions.routeTables[].prefixListIds", "Prefix list ids; list_prefix_get_by_name"},
		{"connector_aws_vpc_delete", "id", `Id of the AWS VPC connector, as a string of digits such as "123" (not a number). The connector_aws_vpc_get_* read tools`},
	} {
		s := prop(tools[c.tool].Schema, c.path)
		if d, _ := s["description"].(string); !strings.Contains(d, c.want) {
			t.Errorf("%s %s: description %q, want %q", c.tool, c.path, d, c.want)
		}
	}
	srv := server.NewMCPServer("mcp-alkira", "test")
	if err := Register(srv, noClient); err != nil {
		t.Fatal(err)
	}
	read := toolSchemas(t, srv)
	for _, f := range writeFamilies {
		_, all := read[f.Name+"_get_all"]
		_, byName := read[f.Name+"_get_by_name"]
		if !all && !byName {
			t.Errorf("%s: no %s_get_all or %s_get_by_name read tool", f.Name, f.Name, f.Name)
		}
	}
}
