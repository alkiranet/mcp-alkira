package registry

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
	"github.com/alkiranet/mcp-alkira/internal/confirm"
	"github.com/alkiranet/mcp-alkira/internal/schemagen"
	"github.com/alkiranet/mcp-alkira/internal/testfamily"
	"github.com/alkiranet/mcp-alkira/writes"
	"github.com/mark3labs/mcp-go/server"
)

type writeEnv struct {
	t      *testing.T
	portal *writePortal
	p      *handlers.WritePipeline
	srv    *server.MCPServer
	now    time.Time
	// key signs the delete codes of p.
	key []byte
	// serve registers the tools under test with p.
	serve func() *server.MCPServer
}

func newWriteEnv(t *testing.T) *writeEnv {
	t.Helper()
	e := &writeEnv{t: t, portal: newWritePortal(t, "widgets"), now: time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC), key: confirm.RandomKey()}
	e.serve = func() *server.MCPServer { return e.server(e.portal.client()) }
	e.setPipeline()
	return e
}

// setPipeline builds e.p with e's key and clock and opts, and serves the tools under test with it.
func (e *writeEnv) setPipeline(opts ...handlers.WriteOption) {
	e.t.Helper()
	base := []handlers.WriteOption{handlers.WithConfirmKey(e.key), handlers.WithClock(func() time.Time { return e.now })}
	e.p = newPipeline(e.t, append(base, opts...)...)
	e.srv = e.serve()
}

// signer returns a signer with the key and clock of e.p.
func (e *writeEnv) signer() *confirm.Signer {
	e.t.Helper()
	s, err := confirm.NewSigner(e.key, confirm.WithClock(func() time.Time { return e.now }))
	if err != nil {
		e.t.Fatal(err)
	}
	return s
}

func newPipeline(t *testing.T, opts ...handlers.WriteOption) *handlers.WritePipeline {
	t.Helper()
	p, err := handlers.NewWritePipeline(opts...)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// server returns a write server for client that shares the pipeline (and so the code key) of e.
func (e *writeEnv) server(client *ak.AlkiraClient) *server.MCPServer {
	srv := server.NewMCPServer("mcp-alkira", "test")
	writes.RegisterType(srv, providerFor(client), e.p, testfamily.Family)
	return srv
}

type writeResult struct {
	Text    string
	IsError bool
	Code    string // refusal code from the result meta
	Refusal string // refusal text from the result meta
}

func callWrite(t *testing.T, srv *server.MCPServer, name string, args map[string]any) writeResult {
	t.Helper()
	raw := call(t, srv, "tools/call", map[string]any{"name": name, "arguments": args})
	var res struct {
		Meta struct {
			Refusal handlers.Refusal `json:"refusal"`
		} `json:"_meta"`
		toolResult
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Content) != 1 {
		t.Fatalf("%s: want one content item, got %s", name, raw)
	}
	if res.IsError && res.Meta.Refusal.Code == "" && !strings.Contains(res.Content[0].Text, "provider") {
		t.Fatalf("%s: error without a typed refusal: %s", name, raw)
	}
	return writeResult{res.Content[0].Text, res.IsError, res.Meta.Refusal.Code, res.Meta.Refusal.Text}
}

func (e *writeEnv) call(name string, args map[string]any) writeResult {
	e.t.Helper()
	return callWrite(e.t, e.srv, name, args)
}

func mustOK(t *testing.T, r writeResult) {
	t.Helper()
	if r.IsError {
		t.Fatalf("unexpected tool error: %s", r.Text)
	}
}

func mustRefuse(t *testing.T, r writeResult, code, text string) {
	t.Helper()
	if !r.IsError || r.Code != code || !strings.Contains(r.Text, text) {
		t.Fatalf("want refusal %s containing %q, got isError=%v code=%q text=%q", code, text, r.IsError, r.Code, r.Text)
	}
}

var codeRE = regexp.MustCompile(`confirmCode: (\S+)`)

func previewCode(t *testing.T, r writeResult) string {
	t.Helper()
	mustOK(t, r)
	m := codeRE.FindStringSubmatch(r.Text)
	if m == nil {
		t.Fatalf("preview without code: %s", r.Text)
	}
	return m[1]
}

func stored(e *writeEnv) {
	e.portal.put(7, map[string]any{"name": "w7", "description": "seven", "psk": "s3cr3t-psk", "state": "PENDING",
		"sites": []any{map[string]any{"id": 1, "name": "a", "key": "site-key-1"}}, "serverOnly": "kept"})
}

// MA-W1, MA-W19, MA-W30: create sends exactly the given fields; unset fields stay absent; explicit zero values are sent.
func TestWrite_CreateRequestShape(t *testing.T) {
	e := newWriteEnv(t)
	r := e.call("widget_create", map[string]any{"body": map[string]any{
		"name": "w1", "enabled": false, "description": "", "count": 0,
	}})
	mustOK(t, r)
	if !strings.Contains(r.Text, `Created widget "w1" with id 101.`) {
		t.Fatalf("text = %s", r.Text)
	}
	reqs := e.portal.requests()
	if len(reqs) != 2 || reqs[0].Method != "GET" || reqs[0].Query != "limit=100&offset=0&paginated=true" || reqs[1].Method != "POST" {
		t.Fatalf("requests = %+v", reqs)
	}
	var sent map[string]any
	json.Unmarshal([]byte(reqs[1].Body), &sent)
	want := map[string]any{"name": "w1", "enabled": false, "description": "", "count": float64(0)}
	if len(sent) != len(want) {
		t.Fatalf("POST body = %s", reqs[1].Body)
	}
	for k, v := range want {
		if sent[k] != v {
			t.Fatalf("POST body = %s", reqs[1].Body)
		}
	}
}

// MA-W1, MA-W2: invalid arguments are refused with zero requests.
func TestWrite_InvalidArgumentsSendNothing(t *testing.T) {
	cases := map[string]struct {
		tool string
		args map[string]any
		want string
	}{
		"missing body":     {"widget_create", map[string]any{}, "body is required"},
		"missing required": {"widget_create", map[string]any{"body": map[string]any{"description": "x"}}, "name is required"},
		"unknown field":    {"widget_create", map[string]any{"body": map[string]any{"name": "a", "colour": "red"}}, "unknown field colour"},
		"wrong type":       {"widget_create", map[string]any{"body": map[string]any{"name": 5}}, "name must be a string"},
		"enum":             {"widget_create", map[string]any{"body": map[string]any{"name": "a", "size": "HUGE"}}, "size must be one of SMALL, LARGE"},
		"maxlen":           {"widget_create", map[string]any{"body": map[string]any{"name": strings.Repeat("a", 21)}}, "at most 20 allowed"},
		"maxlen utf16":     {"widget_create", map[string]any{"body": map[string]any{"name": "a", "description": strings.Repeat("\U0001F600", 21)}}, "42 characters long"},
		"charset":          {"widget_create", map[string]any{"body": map[string]any{"name": "a/b"}}, "outside the name character set"},
		"range":            {"widget_create", map[string]any{"body": map[string]any{"name": "a", "asn": 70000}}, "asn must be in"},
		"max":              {"widget_create", map[string]any{"body": map[string]any{"name": "a", "count": 11}}, "count must be at most 10"},
		"not integer":      {"widget_create", map[string]any{"body": map[string]any{"name": "a", "count": 1.5}}, "count must be an integer"},
		"secret":           {"widget_create", map[string]any{"body": map[string]any{"name": "a", "psk": "x"}}, "psk is a secret field"},
		"nested secret":    {"widget_create", map[string]any{"body": map[string]any{"name": "a", "sites": []any{map[string]any{"name": "s", "key": "k"}}}}, "sites[0].key is a secret field"},
		"nested required":  {"widget_create", map[string]any{"body": map[string]any{"name": "a", "sites": []any{map[string]any{}}}}, "sites[0].name is required"},
		"read-only":        {"widget_create", map[string]any{"body": map[string]any{"name": "a", "state": "SUCCESS"}}, "state is read-only"},
		"extra argument":   {"widget_create", map[string]any{"body": map[string]any{"name": "a"}, "dryRun": true}, "unknown argument dryRun"},
		"patch unknown":    {"widget_update", map[string]any{"id": "7", "patch": map[string]any{"colour": "red"}}, "unknown field colour"},
		"patch wrong type": {"widget_update", map[string]any{"id": "7", "patch": map[string]any{"count": "3"}}, "count must be an integer"},
		"patch remove req": {"widget_update", map[string]any{"id": "7", "patch": map[string]any{"name": nil}}, "name cannot be removed"},
		"non-numeric id":   {"widget_update", map[string]any{"id": "7a", "patch": map[string]any{}}, "id must be a string of digits"},
		"delete bad code":  {"widget_delete", map[string]any{"id": "7", "confirmCode": 5}, "confirmCode must be a non-empty string"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			e := newWriteEnv(t)
			mustRefuse(t, e.call(c.tool, c.args), "invalid_arguments", c.want)
			if n := len(e.portal.requests()); n != 0 {
				t.Fatalf("%d requests sent", n)
			}
		})
	}
}

// MA-W4, MA-W38, MA-W27, MA-W20: update merges the patch into the stored body.
func TestWrite_UpdateMerge(t *testing.T) {
	e := newWriteEnv(t)
	stored(e)
	r := e.call("widget_update", map[string]any{"id": "7", "patch": map[string]any{"description": "new", "tags": []any{}, "serverOnly": nil}})
	// serverOnly is not a field of the type, so the patch is refused before any request.
	mustRefuse(t, r, "invalid_arguments", "unknown field serverOnly")

	r = e.call("widget_update", map[string]any{"id": "7", "patch": map[string]any{"description": "new", "tags": []any{}}})
	mustOK(t, r)
	if e.portal.count("GET") != 1 || e.portal.count("PUT") != 1 {
		t.Fatalf("requests = %+v", e.portal.requests())
	}
	put := e.portal.requests()[1]
	var body map[string]any
	json.Unmarshal([]byte(put.Body), &body)
	if body["description"] != "new" || body["psk"] != "s3cr3t-psk" || body["serverOnly"] != "kept" || body["name"] != "w7" {
		t.Fatalf("PUT body = %s", put.Body)
	}
	if _, ok := body["state"]; ok {
		t.Fatalf("read-only state resent: %s", put.Body)
	}
	if tags, ok := body["tags"].([]any); !ok || len(tags) != 0 {
		t.Fatalf("explicit [] not sent: %s", put.Body)
	}
	if strings.Contains(r.Text, "s3cr3t-psk") || strings.Contains(r.Text, "site-key-1") {
		t.Fatalf("secret in output: %s", r.Text)
	}
	if !strings.Contains(r.Text, "Previous object:") || !strings.Contains(r.Text, `"description": "seven"`) {
		t.Fatalf("no previous object: %s", r.Text)
	}
}

// MA-W4: null removes a key; a backend that keeps it is reported (BC-R3).
func TestWrite_UpdateNullKeptByBackend(t *testing.T) {
	e := newWriteEnv(t)
	stored(e)
	e.portal.fail("PUT", e.portal.base()+"/7", fault{status: 200, body: `{"id":7,"name":"w7","description":"seven"}`})
	r := e.call("widget_update", map[string]any{"id": "7", "patch": map[string]any{"description": nil}})
	mustOK(t, r)
	var body map[string]any
	json.Unmarshal([]byte(e.portal.requests()[1].Body), &body)
	if _, ok := body["description"]; ok {
		t.Fatalf("null did not remove the key: %v", body)
	}
	if !strings.Contains(r.Text, "description kept by the backend") {
		t.Fatalf("text = %s", r.Text)
	}
}

// MA-W4: id-matched arrays and immutable fields refuse with no PUT.
func TestWrite_UpdateRefusals(t *testing.T) {
	cases := map[string]struct {
		patch map[string]any
		want  string
	}{
		"dropped site id": {map[string]any{"sites": []any{map[string]any{"name": "b"}}}, "id 1 is missing"},
		"unknown site id": {map[string]any{"sites": []any{map[string]any{"id": 1, "name": "a"}, map[string]any{"id": 9, "name": "z"}}}, "unknown id 9"},
		"immutable":       {map[string]any{"region": "us-east-2"}, "region cannot be changed"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			e := newWriteEnv(t)
			stored(e)
			o := e.portal.object(7)
			o["region"] = "us-east-1"
			mustRefuse(t, e.call("widget_update", map[string]any{"id": "7", "patch": c.patch}), "invalid_arguments", c.want)
			if e.portal.count("PUT") != 0 {
				t.Fatal("PUT sent")
			}
		})
	}
}

// MA-W5 plumbing, MA-W16: state and docState refusals send no write.
func TestWrite_StateRefusals(t *testing.T) {
	cases := map[string]struct {
		obj  map[string]any
		want string
	}{
		"failed":  {map[string]any{"name": "w", "state": "FAILED"}, "cannot change while its state is FAILED"},
		"marked":  {map[string]any{"name": "w", "docState": "MARKED_FOR_DELETION"}, "deletion pending provisioning"},
		"deleted": {map[string]any{"name": "w", "docState": "DELETED"}, "already deleted; the backend keeps a residue row"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			e := newWriteEnv(t)
			e.portal.put(7, c.obj)
			mustRefuse(t, e.call("widget_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}), "state_refused", c.want)
			if e.portal.count("PUT") != 0 {
				t.Fatal("PUT sent")
			}
			if q := e.portal.requests()[0].Query; q != "includeMarkedForDeletion=true" {
				t.Fatalf("pre-GET query = %q", q)
			}
		})
	}
}

// MA-W10, MA-W51: delete without a code previews with GETs only.
func TestWrite_DeletePreview(t *testing.T) {
	e := newWriteEnv(t)
	stored(e)
	r := e.call("widget_delete", map[string]any{"id": "7"})
	previewCode(t, r)
	for _, want := range []string{
		`Delete preview for widget "w7" (id 7). Nothing was deleted.`,
		"Deployed: no, never deployed (state PENDING).",
		`Also deleted with it: "w7-Group".`,
		"The delete now: removes it at once.",
		"The next provision: nothing to tear down.",
		"expires: 2026-11-17T12:05:00Z",
		`"psk": "***"`,
		`"key": "***"`,
	} {
		if !strings.Contains(r.Text, want) {
			t.Errorf("preview lacks %q:\n%s", want, r.Text)
		}
	}
	if strings.Contains(r.Text, "s3cr3t-psk") || strings.Contains(r.Text, "site-key-1") || strings.Contains(r.Text, "serverOnly") {
		t.Fatalf("preview leaks a secret or an unmodelled key: %s", r.Text)
	}
	if n := e.portal.count("DELETE"); n != 0 {
		t.Fatalf("%d DELETE sent", n)
	}

	// A first provision still SCHEDULED keeps previousState PENDING: not deployed yet.
	e.portal.put(9, map[string]any{"name": "w9", "state": "SCHEDULED", "previousState": "PENDING"})
	r = e.call("widget_delete", map[string]any{"id": "9"})
	if !strings.Contains(r.Text, "Deployed: no, never deployed (state SCHEDULED).") || !strings.Contains(r.Text, "removes it at once") {
		t.Errorf("scheduled first provision:\n%s", r.Text)
	}

	e.portal.put(8, map[string]any{"name": "w8", "state": "SUCCESS"})
	r = e.call("widget_delete", map[string]any{"id": "8"})
	for _, want := range []string{"Deployed: yes (state SUCCESS).", "marks it for deletion; it keeps its name and references until the next provision", "tears down the deployed object"} {
		if !strings.Contains(r.Text, want) {
			t.Errorf("deployed preview lacks %q:\n%s", want, r.Text)
		}
	}
}

// MA-W10, MA-W35: a valid code is checked on the core pre-GET, then one DELETE.
func TestWrite_DeleteWithCode(t *testing.T) {
	e := newWriteEnv(t)
	stored(e)
	code := previewCode(t, e.call("widget_delete", map[string]any{"id": "7"}))
	before := len(e.portal.requests())
	r := e.call("widget_delete", map[string]any{"id": "7", "confirmCode": code})
	mustOK(t, r)
	if !strings.Contains(r.Text, `Deleted widget "w7" (id 7).`) || strings.Contains(r.Text, "s3cr3t-psk") {
		t.Fatalf("text = %s", r.Text)
	}
	// One pre-GET, the DELETE, and one flag read on a cold cache.
	reqs := e.portal.requests()[before:]
	if len(reqs) != 2 || reqs[0].Method != "GET" || reqs[1].Method != "DELETE" || e.portal.flagReads != 1 {
		t.Fatalf("requests = %+v, flag reads %d", reqs, e.portal.flagReads)
	}

	// A deployed object with marked rows is reported as marked.
	e.portal.put(8, map[string]any{"name": "w8", "state": "SUCCESS"})
	code = previewCode(t, e.call("widget_delete", map[string]any{"id": "8"}))
	r = e.call("widget_delete", map[string]any{"id": "8", "confirmCode": code})
	if !strings.Contains(r.Text, "Marked widget \"w8\" (id 8) for deletion") {
		t.Fatalf("text = %s", r.Text)
	}
}

// MA-W10: no delete cap on the desktop.
func TestWrite_TwentyDeletes(t *testing.T) {
	e := newWriteEnv(t)
	for i := 1; i <= 20; i++ {
		e.portal.put(i, map[string]any{"name": "w"})
		id := strconv.Itoa(i)
		code := previewCode(t, e.call("widget_delete", map[string]any{"id": id}))
		mustOK(t, e.call("widget_delete", map[string]any{"id": id, "confirmCode": code}))
	}
	if n := e.portal.count("DELETE"); n != 20 {
		t.Fatalf("%d DELETE sent", n)
	}
}

// MA-W10: delete of a missing object reports not found and sends nothing else.
func TestWrite_DeleteNotFound(t *testing.T) {
	e := newWriteEnv(t)
	mustRefuse(t, e.call("widget_delete", map[string]any{"id": "404"}), "not_found", "widget id 404 not found.")
	if len(e.portal.requests()) != 1 {
		t.Fatalf("requests = %+v", e.portal.requests())
	}
}

// MA-W52: an expired code is refused; a fresh preview then deletes.
func TestWrite_DeleteCodeExpired(t *testing.T) {
	e := newWriteEnv(t)
	stored(e)
	code := previewCode(t, e.call("widget_delete", map[string]any{"id": "7"}))
	e.now = e.now.Add(confirm.Lifetime + time.Second)
	mustRefuse(t, e.call("widget_delete", map[string]any{"id": "7", "confirmCode": code}), handlers.RefusalConfirmExpired,
		"confirmation code expired; call widget_delete again with id 7 and no confirmCode for a new preview")
	if e.portal.count("DELETE") != 0 {
		t.Fatal("DELETE sent")
	}
	code = previewCode(t, e.call("widget_delete", map[string]any{"id": "7"}))
	mustOK(t, e.call("widget_delete", map[string]any{"id": "7", "confirmCode": code}))
	if e.portal.count("DELETE") != 1 {
		t.Fatal("want exactly one DELETE")
	}
}

// MA-W53: tampered codes and codes of another id or family are refused.
func TestWrite_DeleteCodeTampered(t *testing.T) {
	e := newWriteEnv(t)
	stored(e)
	e.portal.put(8, map[string]any{"name": "w7", "description": "seven", "psk": "s3cr3t-psk", "state": "PENDING",
		"sites": []any{map[string]any{"id": 1, "name": "a", "key": "site-key-1"}}, "serverOnly": "kept"})
	code := previewCode(t, e.call("widget_delete", map[string]any{"id": "7"}))
	buf, _ := base64.RawURLEncoding.DecodeString(code)

	flipMAC := append([]byte(nil), buf...)
	flipMAC[len(flipMAC)-1] ^= 1
	later := append([]byte(nil), buf...)
	later[7] += 60

	sig := e.signer()
	var hash [32]byte
	copy(hash[:], buf[8:40])
	otherFamily, _ := sig.Issue(e.portal.srv.URL+"/api#7", "gadget", "7", hash)

	for name, c := range map[string]struct{ id, code string }{
		"mac byte":     {"7", base64.RawURLEncoding.EncodeToString(flipMAC)},
		"expiry later": {"7", base64.RawURLEncoding.EncodeToString(later)},
		"other id":     {"8", code},
		"other family": {"7", otherFamily},
		"garbage":      {"7", "not-a-code"},
	} {
		t.Run(name, func(t *testing.T) {
			before := len(e.portal.requests())
			mustRefuse(t, e.call("widget_delete", map[string]any{"id": c.id, "confirmCode": c.code}), handlers.RefusalConfirmMismatch, "confirmation code does not match this object")
			if n := len(e.portal.requests()) - before; n != 0 {
				t.Fatalf("%d requests sent", n)
			}
		})
	}
	if e.portal.count("DELETE") != 0 {
		t.Fatal("DELETE sent")
	}
}

// MA-W54: a code is void once the object changed or is gone.
func TestWrite_DeleteCodeStale(t *testing.T) {
	e := newWriteEnv(t)
	stored(e)
	code := previewCode(t, e.call("widget_delete", map[string]any{"id": "7"}))
	e.portal.object(7)["description"] = "changed in the portal"
	mustRefuse(t, e.call("widget_delete", map[string]any{"id": "7", "confirmCode": code}), handlers.RefusalConfirmStale, "the object changed since the preview; review the new preview")
	if e.portal.count("DELETE") != 0 {
		t.Fatal("DELETE sent")
	}

	// A volatile field does not void the code.
	code = previewCode(t, e.call("widget_delete", map[string]any{"id": "7"}))
	e.portal.object(7)["lastSeenAt"] = "2026-11-17T12:01:00Z"
	mustOK(t, e.call("widget_delete", map[string]any{"id": "7", "confirmCode": code}))

	// Reusing the code after the delete: the object is gone.
	mustRefuse(t, e.call("widget_delete", map[string]any{"id": "7", "confirmCode": code}), "not_found", "not found")

	// A deployed object is marked by the delete; reusing its code reads the marked row.
	e.portal.put(8, map[string]any{"name": "w8", "state": "SUCCESS"})
	code = previewCode(t, e.call("widget_delete", map[string]any{"id": "8"}))
	mustOK(t, e.call("widget_delete", map[string]any{"id": "8", "confirmCode": code}))
	mustRefuse(t, e.call("widget_delete", map[string]any{"id": "8", "confirmCode": code}), "state_refused", "deletion pending provisioning")
	if n := e.portal.count("DELETE"); n != 2 {
		t.Fatalf("%d DELETE sent, want 2", n)
	}
}

// MA-W55: a code issued in one tenant's session is refused in another's.
func TestWrite_DeleteCodeOtherTenant(t *testing.T) {
	e := newWriteEnv(t)
	stored(e)
	code := previewCode(t, e.call("widget_delete", map[string]any{"id": "7"}))

	other := newWritePortal(t, "widgets")
	other.put(7, e.portal.object(7))
	srvB := e.server(other.client())
	mustRefuse(t, callWrite(t, srvB, "widget_delete", map[string]any{"id": "7", "confirmCode": code}), handlers.RefusalConfirmMismatch, "does not match this object")
	if other.count("DELETE") != 0 || len(other.requests()) != 0 {
		t.Fatalf("requests in tenant B = %+v", other.requests())
	}
}

// A code issued by one pipeline is accepted by another built with the same
// key, as on another replica, and refused by one with a different key.
func TestWrite_DeleteCodeSharedKey(t *testing.T) {
	key := confirm.RandomKey()
	serve := func(portal *writePortal, key []byte) *server.MCPServer {
		srv := server.NewMCPServer("mcp-alkira", "test")
		writes.RegisterType(srv, providerFor(portal.client()), newPipeline(t, handlers.WithConfirmKey(key)), testfamily.Family)
		return srv
	}
	portal := newWritePortal(t, "widgets")
	portal.put(7, map[string]any{"name": "w7", "state": "PENDING"})
	portal.put(8, map[string]any{"name": "w8", "state": "PENDING"})
	replicaA, replicaB := serve(portal, key), serve(portal, key)

	code := previewCode(t, callWrite(t, replicaA, "widget_delete", map[string]any{"id": "7"}))
	mustOK(t, callWrite(t, replicaB, "widget_delete", map[string]any{"id": "7", "confirmCode": code}))

	code = previewCode(t, callWrite(t, replicaA, "widget_delete", map[string]any{"id": "8"}))
	otherKey := serve(portal, confirm.RandomKey())
	mustRefuse(t, callWrite(t, otherKey, "widget_delete", map[string]any{"id": "8", "confirmCode": code}), handlers.RefusalConfirmMismatch, "does not match this object")
	if n := portal.count("DELETE"); n != 1 {
		t.Fatalf("%d DELETE sent, want 1", n)
	}
}

// MA-W14: concurrent writes are not serialised; the backend's locks decide
// contention, as for the portal.
func TestWrite_Concurrent(t *testing.T) {
	e := newWriteEnv(t)
	stored(e)
	e.portal.put(8, map[string]any{"name": "w8", "state": "PENDING"})
	var (
		mu        sync.Mutex
		active    int
		maxActive int
		closeOnce sync.Once
	)
	// The hook runs for every request, so the GETs and the PUTs of the two
	// updates can each bring two requests in flight; the channel closes once.
	both := make(chan struct{})
	e.portal.mu.Lock()
	e.portal.before = func() {
		mu.Lock()
		active++
		maxActive = max(maxActive, active)
		if active == 2 {
			closeOnce.Do(func() { close(both) })
		}
		mu.Unlock()
		select {
		case <-both:
		case <-time.After(5 * time.Second):
		}
		mu.Lock()
		active--
		mu.Unlock()
	}
	e.portal.mu.Unlock()
	var wg sync.WaitGroup
	for _, id := range []string{"7", "8"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mustOK(t, callWrite(t, e.srv, "widget_update", map[string]any{"id": id, "patch": map[string]any{"description": "x"}}))
		}()
	}
	wg.Wait()
	if maxActive != 2 {
		t.Fatalf("at most %d request in flight, want 2", maxActive)
	}
	if e.portal.count("PUT") != 2 {
		t.Fatalf("requests = %+v", e.portal.requests())
	}
}

// MA-W14: two updates of one object run one after the other, so the second
// merges onto the stored object the first one wrote.
func TestWrite_SameObjectSerialized(t *testing.T) {
	e := newWriteEnv(t)
	stored(e)
	var (
		mu        sync.Mutex
		active    int
		maxActive int
	)
	// Each request waits briefly for another one, so overlapping calls would
	// both read the stored object before either PUT.
	e.portal.mu.Lock()
	e.portal.before = func() {
		mu.Lock()
		active++
		maxActive = max(maxActive, active)
		mu.Unlock()
		time.Sleep(50 * time.Millisecond)
		mu.Lock()
		active--
		mu.Unlock()
	}
	e.portal.mu.Unlock()
	var wg sync.WaitGroup
	for _, m := range []string{"m1", "m2"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mustOK(t, callWrite(t, e.srv, "widget_update", map[string]any{"id": "7", "add": map[string]any{"members": []any{m}}}))
		}()
	}
	wg.Wait()
	if maxActive != 1 {
		t.Fatalf("%d requests in flight, want 1", maxActive)
	}
	var methods []string
	for _, r := range e.portal.requests() {
		methods = append(methods, r.Method)
	}
	if got := strings.Join(methods, " "); got != "GET PUT GET PUT" {
		t.Fatalf("requests = %s", got)
	}
	members, _ := e.portal.object(7)["members"].([]any)
	if len(members) != 2 {
		t.Fatalf("stored members = %v, want both m1 and m2", members)
	}
}

// MA-W12, MA-W24: classified errors map to their texts with one request each.
func TestWrite_ErrorMapping(t *testing.T) {
	coll := "/api/tenantnetworks/7/widgets"
	cases := map[string]struct {
		f    fault
		code string
		want []string
	}{
		"429 with Retry-After": {fault{status: 429, header: map[string]string{"Retry-After": "31"}, body: `{"code":"AK000429","message":"busy"}`}, "rate_limited", []string{"retry after 31 s"}},
		"429 without":          {fault{status: 429, body: `{"message":"Too many requests"}`}, "rate_limited", []string{"Rate limit reached; retry later."}},
		"409 duplicate db": {fault{status: 409, body: `{"code":"AK000001","id":"ALK-1","message":"duplicate key value violates unique constraint \"udr_list_tenant_id_tenant_network_id_name\""}`},
			"duplicate", []string{`a widget named "w1" already exists.`}},
		"403 permission": {fault{status: 403, body: `{"code":"AK000403","id":"ALK-77","message":"Forbidden"}`}, "permission_denied", []string{"Permission denied", "ALK-77", "AK000403", "request id mcp-"}},
		"500 other":      {fault{status: 500, body: `{"code":"AK000500","id":"ALK-9","message":"Internal error"}`}, "backend_error", []string{"status 500", "ALK-9"}},
		"500 db text": {fault{status: 500, body: `{"code":"AK000500","id":"ALK-10","message":"duplicate key value violates unique constraint \"udr_list_tenant_id_tenant_network_id_name\""}`},
			"backend_error", []string{"Create widget failed with status 500 (code AK000500, error id ALK-10"}},
		"limit": {fault{status: 404, body: `{"code":"AK000404","message":"Cannot create more than 10 billing tags"}`}, "limit_exceeded", []string{`The tenant is at its limit: "Cannot create more than 10 billing tags" (code AK000404`}},
		"validation pending": {fault{status: 403, body: `{"code":"AK000403","id":"ALK-8","message":"Cannot create until the validation completes"}`}, "validation_pending",
			[]string{"Validation still running for this object.", "ALK-8"}},
		"busy state": {fault{status: 400, body: `{"code":"AK000002","message":"Cannot modify connector with name: 'x', id: 7 at this time (network state : FAILED, connector state : SCHEDULED)"}`}, "state_refused", []string{"Cannot modify connector", "Wait for the running provision"}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			e := newWriteEnv(t)
			e.portal.fail("POST", coll, c.f)
			r := e.call("widget_create", map[string]any{"body": map[string]any{"name": "w1"}})
			for _, w := range c.want {
				mustRefuse(t, r, c.code, w)
			}
			if strings.Contains(r.Text, "udr_list_tenant") {
				t.Fatalf("raw constraint text passed through: %s", r.Text)
			}
			if n := e.portal.count("POST"); n != 1 {
				t.Fatalf("%d POST sent", n)
			}
		})
	}
}

// MA-W12: an in-use delete answer keeps the backend text.
func TestWrite_DeleteInUse(t *testing.T) {
	e := newWriteEnv(t)
	stored(e)
	code := previewCode(t, e.call("widget_delete", map[string]any{"id": "7"}))
	e.portal.fail("DELETE", e.portal.base()+"/7", fault{status: 409, body: `{"code":"AK000409","message":"Widget is in use","errors":[{"field":"widget","code":"widget.00000010","message":"used by policy 3"}]}`})
	mustRefuse(t, e.call("widget_delete", map[string]any{"id": "7", "confirmCode": code}), "in_use", `widget id 7 is still referenced: "Widget is in use; widget: used by policy 3" (code AK000409`)
	if e.portal.count("DELETE") != 1 {
		t.Fatal("want one DELETE, no retry")
	}
}

// MA-W11 plumbing: the duplicate guard refuses with no POST.
func TestWrite_DuplicateGuard(t *testing.T) {
	e := newWriteEnv(t)
	stored(e)
	mustRefuse(t, e.call("widget_create", map[string]any{"body": map[string]any{"name": "w7"}}), "duplicate", `a widget named "w7" already exists (id 7).`)
	if e.portal.count("POST") != 0 {
		t.Fatal("POST sent")
	}
}

// MA-W22, MA-W39: lost responses report the outcome of the confirming read.
func TestWrite_Outcomes(t *testing.T) {
	t.Run("PUT reset, applied", func(t *testing.T) {
		e := newWriteEnv(t)
		stored(e)
		e.portal.fail("PUT", e.portal.base()+"/7", fault{kind: faultCommit})
		r := e.call("widget_update", map[string]any{"id": "7", "patch": map[string]any{"description": "new"}})
		mustOK(t, r)
		if !strings.Contains(r.Text, "applied") || e.portal.count("PUT") != 1 {
			t.Fatalf("text = %s", r.Text)
		}
	})
	t.Run("PUT reset, may still apply", func(t *testing.T) {
		e := newWriteEnv(t)
		stored(e)
		e.portal.fail("PUT", e.portal.base()+"/7", fault{kind: faultReset})
		mustRefuse(t, e.call("widget_update", map[string]any{"id": "7", "patch": map[string]any{"description": "new"}}), "may_still_apply",
			"Update of widget id 7 may still apply (the response was lost; a confirming read does not show the change yet")
	})
	t.Run("PUT reset with list edits, applied", func(t *testing.T) {
		e := newWriteEnv(t)
		stored(e)
		e.portal.fail("PUT", e.portal.base()+"/7", fault{kind: faultCommit})
		r := e.call("widget_update", map[string]any{"id": "7", "add": map[string]any{"members": []any{"m1"}}})
		mustOK(t, r)
		if !strings.Contains(r.Text, "Updated widget id 7: applied") {
			t.Fatalf("text = %s", r.Text)
		}
	})
	t.Run("DELETE 500 then not found", func(t *testing.T) {
		e := newWriteEnv(t)
		stored(e)
		code := previewCode(t, e.call("widget_delete", map[string]any{"id": "7"}))
		e.portal.fail("DELETE", e.portal.base()+"/7", fault{kind: faultCommit})
		r := e.call("widget_delete", map[string]any{"id": "7", "confirmCode": code})
		mustOK(t, r)
		if !strings.Contains(r.Text, "Deleted widget id 7") || e.portal.count("DELETE") != 1 {
			t.Fatalf("text = %s", r.Text)
		}
	})
	for name, body := range map[string]string{"empty": "", "html": "<html>Not Found</html>", "gateway": `{"message":"Not Found"}`} {
		t.Run("DELETE 500 then 404 "+name, func(t *testing.T) {
			e := newWriteEnv(t)
			stored(e)
			code := previewCode(t, e.call("widget_delete", map[string]any{"id": "7"}))
			e.portal.fail("DELETE", e.portal.base()+"/7", fault{status: 500, body: `{"message":"Request failed"}`})
			// The pre-GET answers the object; the confirming GET answers a 404 that is not ErrorDetails.
			e.portal.mu.Lock()
			e.portal.faults["GET "+e.portal.base()+"/7"] = []fault{{status: 200, body: mustJSON(e.portal.objects[7])}, {status: 404, body: body}}
			e.portal.mu.Unlock()
			r := e.call("widget_delete", map[string]any{"id": "7", "confirmCode": code})
			mustRefuse(t, r, "outcome_unknown", "delete outcome unknown")
			if strings.Contains(r.Text, "still finds the object") {
				t.Fatalf("text = %s", r.Text)
			}
		})
	}
	t.Run("DELETE 500 then object still there", func(t *testing.T) {
		e := newWriteEnv(t)
		stored(e)
		code := previewCode(t, e.call("widget_delete", map[string]any{"id": "7"}))
		e.portal.fail("DELETE", e.portal.base()+"/7", fault{status: 500, body: `{"message":"Request failed"}`})
		mustRefuse(t, e.call("widget_delete", map[string]any{"id": "7", "confirmCode": code}), "outcome_unknown",
			"Widget 7: delete outcome unknown, a confirming read still finds the object")
		if e.portal.count("DELETE") != 1 {
			t.Fatal("DELETE retried")
		}
	})
	t.Run("POST Request failed, created", func(t *testing.T) {
		e := newWriteEnv(t)
		e.portal.fail("POST", e.portal.base(), fault{kind: faultCommit})
		r := e.call("widget_create", map[string]any{"body": map[string]any{"name": "w1"}})
		mustOK(t, r)
		if !strings.Contains(r.Text, `Created widget with name "w1", id 101`) || e.portal.count("POST") != 1 {
			t.Fatalf("text = %s", r.Text)
		}
	})
	t.Run("POST Request failed, may still be created", func(t *testing.T) {
		e := newWriteEnv(t)
		e.portal.fail("POST", e.portal.base(), fault{status: 500, body: `{"message":"Request failed"}`})
		mustRefuse(t, e.call("widget_create", map[string]any{"body": map[string]any{"name": "w1"}}), "may_still_be_created",
			`Widget with name "w1" may still be created (the response was lost; a lookup by name found no object yet). List the widgets and look for name "w1" before creating it again.`)
		if e.portal.count("POST") != 1 {
			t.Fatal("POST retried")
		}
	})
	t.Run("deadline", func(t *testing.T) {
		e := newWriteEnv(t)
		stored(e)
		e.setPipeline(handlers.WithWriteTimeout(200 * time.Millisecond))
		e.portal.fail("PUT", e.portal.base()+"/7", fault{kind: faultHang})
		r := e.call("widget_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}})
		if !r.IsError || e.portal.count("PUT") != 1 {
			t.Fatalf("result = %+v", r)
		}
	})
}

func mustJSON(v any) string { b, _ := json.Marshal(v); return string(b) }

// List edits: add and remove on a set field, absent items, no-op edits and refusals.
func TestWrite_ListEdits(t *testing.T) {
	e := newWriteEnv(t)
	stored(e)
	e.portal.object(7)["members"] = []any{"a", "b"}

	r := e.call("widget_update", map[string]any{"id": "7", "add": map[string]any{"members": []any{"c"}}, "remove": map[string]any{"members": []any{"a", "zz"}}})
	mustOK(t, r)
	if e.portal.count("GET") != 1 || e.portal.count("PUT") != 1 {
		t.Fatalf("requests = %+v", e.portal.requests())
	}
	var body map[string]any
	json.Unmarshal([]byte(e.portal.requests()[1].Body), &body)
	if got := mustJSON(body["members"]); got != `["b","c"]` {
		t.Fatalf("PUT members = %s", got)
	}
	if !strings.Contains(r.Text, `Not in the stored members, nothing to remove: "zz".`) || !strings.Contains(r.Text, "Previous object:") {
		t.Fatalf("text = %s", r.Text)
	}

	// Edits that change nothing send no PUT.
	before := len(e.portal.requests())
	r = e.call("widget_update", map[string]any{"id": "7", "add": map[string]any{"members": []any{"b"}}, "remove": map[string]any{"members": []any{"a"}}})
	mustOK(t, r)
	if !strings.Contains(r.Text, "No change to widget id 7: the stored lists already match the edits; nothing sent.") ||
		!strings.Contains(r.Text, `Not in the stored members, nothing to remove: "a".`) {
		t.Fatalf("text = %s", r.Text)
	}
	if reqs := e.portal.requests()[before:]; len(reqs) != 1 || reqs[0].Method != "GET" {
		t.Fatalf("requests = %+v", reqs)
	}

	// Removes on an id-matched array match element ids: an absent id is
	// reported, and an element the response still holds is named by its id.
	e.portal.fail("PUT", e.portal.base()+"/7", fault{status: 200, body: `{"id":7,"name":"w7","sites":[{"id":1,"name":"a"}]}`})
	r = e.call("widget_update", map[string]any{"id": "7", "remove": map[string]any{"sites": []any{1, 9}}})
	mustOK(t, r)
	if !strings.Contains(r.Text, "Not in the stored sites, nothing to remove: 9.") ||
		!strings.Contains(r.Text, "The backend did not keep the list edits as sent: sites: element id 1 not removed.") {
		t.Fatalf("text = %s", r.Text)
	}

	// An element of an id-matched array is removed by id.
	mustOK(t, e.call("widget_update", map[string]any{"id": "7", "remove": map[string]any{"sites": []any{1}}}))
	if sites, _ := e.portal.object(7)["sites"].([]any); len(sites) != 0 {
		t.Fatalf("sites = %v", sites)
	}

	// The backend not keeping an added item is reported.
	e.portal.fail("PUT", e.portal.base()+"/7", fault{status: 200, body: `{"id":7,"name":"w7","members":["b"]}`})
	r = e.call("widget_update", map[string]any{"id": "7", "add": map[string]any{"members": []any{"d"}}})
	mustOK(t, r)
	if !strings.Contains(r.Text, `The backend did not keep the list edits as sent: members: "d" not added.`) {
		t.Fatalf("text = %s", r.Text)
	}

	// Items compare exactly: a remove that differs only in case is absent, not "not removed".
	e.portal.object(7)["members"] = []any{"a", "b"}
	r = e.call("widget_update", map[string]any{"id": "7", "add": map[string]any{"members": []any{"c"}}, "remove": map[string]any{"members": []any{"A"}}})
	mustOK(t, r)
	if !strings.Contains(r.Text, `Not in the stored members, nothing to remove: "A".`) || strings.Contains(r.Text, "not removed") || strings.Contains(r.Text, "did not keep the list edits") {
		t.Fatalf("text = %s", r.Text)
	}
	if got := mustJSON(e.portal.object(7)["members"]); got != `["a","b","c"]` {
		t.Fatalf("members = %s", got)
	}

	for name, c := range map[string]struct {
		args map[string]any
		want string
	}{
		"patch sets a set field":  {map[string]any{"id": "7", "patch": map[string]any{"members": []any{"x"}}}, "members cannot be set by a patch; use add or remove"},
		"patch nulls a set field": {map[string]any{"id": "7", "patch": map[string]any{"members": nil}}, "members cannot be set by a patch; use add or remove"},
		"nothing to change":       {map[string]any{"id": "7"}, "patch, add or remove is required"},
		"add not an object":       {map[string]any{"id": "7", "add": []any{"x"}}, "add must be a non-empty object"},
		"empty items":             {map[string]any{"id": "7", "remove": map[string]any{"members": []any{}}}, "remove.members must be a non-empty array"},
		"add breaks element rule": {map[string]any{"id": "7", "add": map[string]any{"members": []any{"a/b"}}}, "add.members"},
		"add on a plain list":     {map[string]any{"id": "7", "add": map[string]any{"tags": []any{"x"}}}, "tags is not a list that add applies to"},
		"add on a secret field":   {map[string]any{"id": "7", "add": map[string]any{"psk": []any{"hunter2"}}}, "add.psk: psk is not a list that add applies to"},
		"remove on a read-only":   {map[string]any{"id": "7", "remove": map[string]any{"state": []any{"x"}}}, "remove.state: state is not a list that remove applies to"},
		"add on an id-matched":    {map[string]any{"id": "7", "add": map[string]any{"sites": []any{1}}}, "add.sites: sites is not a list that add applies to"},
		"both add and remove":     {map[string]any{"id": "7", "add": map[string]any{"members": []any{"x"}}, "remove": map[string]any{"members": []any{"x"}}}, "in both add and remove"},
	} {
		t.Run(name, func(t *testing.T) {
			before := len(e.portal.requests())
			reads := e.portal.flagReads
			r := e.call("widget_update", c.args)
			mustRefuse(t, r, "invalid_arguments", c.want)
			if strings.Contains(r.Text, "hunter2") {
				t.Fatalf("secret value echoed: %s", r.Text)
			}
			if n := len(e.portal.requests()) - before; n != 0 || e.portal.flagReads != reads {
				t.Fatalf("%d requests sent", n)
			}
		})
	}
}

// Kill switch: BACKEND-MCP-WRITE-DISABLED refuses every write with no write request; previews still run.
func TestWrite_KillSwitch(t *testing.T) {
	e := newWriteEnv(t)
	stored(e)
	code := previewCode(t, e.call("widget_delete", map[string]any{"id": "7"}))
	e.portal.flags = []string{"SELECTIVE-PROV", ak.WriteDisabledFlag}
	const text = "Changes through MCP are disabled for this tenant; nothing was sent."
	mustRefuse(t, e.call("widget_create", map[string]any{"body": map[string]any{"name": "w1"}}), "writes_disabled", text)
	mustRefuse(t, e.call("widget_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}), "writes_disabled", text)
	mustRefuse(t, e.call("widget_delete", map[string]any{"id": "7", "confirmCode": code}), "writes_disabled", text)
	previewCode(t, e.call("widget_delete", map[string]any{"id": "7"}))
	if e.portal.count("POST")+e.portal.count("PUT")+e.portal.count("DELETE") != 0 {
		t.Fatalf("requests = %+v", e.portal.requests())
	}
	if e.portal.flagReads != 1 {
		t.Fatalf("%d flag reads, want 1 (cached)", e.portal.flagReads)
	}
}

// Flags that cannot be read, with none cached, refuse every write with no
// write request; a delete preview sends only GETs and still runs.
func TestWrite_FlagsUnavailable(t *testing.T) {
	e := newWriteEnv(t)
	stored(e)
	e.portal.flagStatus = 500
	const text = "Cannot check whether changes through MCP are allowed for this tenant; try again. Nothing was sent."
	mustRefuse(t, e.call("widget_create", map[string]any{"body": map[string]any{"name": "w1"}}), "flags_unavailable", text)
	mustRefuse(t, e.call("widget_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}), "flags_unavailable", text)
	code := previewCode(t, e.call("widget_delete", map[string]any{"id": "7"}))
	mustRefuse(t, e.call("widget_delete", map[string]any{"id": "7", "confirmCode": code}), "flags_unavailable", text)
	if e.portal.count("POST")+e.portal.count("PUT")+e.portal.count("DELETE") != 0 {
		t.Fatalf("requests = %+v", e.portal.requests())
	}
}

// dialFailure fails every write request before it is sent.
type dialFailure struct{ next http.RoundTripper }

func (d dialFailure) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Method != "GET" {
		return nil, &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}
	}
	return d.next.RoundTrip(r)
}

// A write that fails before any byte is sent is reported as not sent, with no confirming read.
func TestWrite_NotSent(t *testing.T) {
	e := newWriteEnv(t)
	stored(e)
	client := e.portal.client()
	client.Client.HTTPClient.Transport = dialFailure{http.DefaultTransport}
	mustRefuse(t, callWrite(t, e.server(client), "widget_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}), "not_sent",
		"The update of widget id 7 was not sent, nothing changed")
	if reqs := e.portal.requests(); len(reqs) != 1 || reqs[0].Method != "GET" {
		t.Fatalf("requests = %+v", reqs)
	}
}

// MA-W29: a malformed pre-GET refuses; a 2xx create answer without a usable id reports the id the lookup finds.
func TestWrite_MalformedBodies(t *testing.T) {
	e := newWriteEnv(t)
	stored(e)
	e.portal.fail("GET", e.portal.base()+"/7", fault{kind: faultBadJSON})
	r := e.call("widget_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}})
	if !r.IsError || e.portal.count("PUT") != 0 {
		t.Fatalf("result = %+v", r)
	}
	for i, body := range []string{`{"id": 1,`, `{"name":"w"}`, ``} {
		name := "w" + strconv.Itoa(i)
		e.portal.fail("POST", e.portal.base(), fault{kind: faultRewrite, status: 201, body: body})
		r = e.call("widget_create", map[string]any{"body": map[string]any{"name": name}})
		mustOK(t, r)
		want := `Created widget "` + name + `" with id ` + strconv.Itoa(101+i) + "."
		if !strings.Contains(r.Text, want) {
			t.Fatalf("body %q: want %q in %s", body, want, r.Text)
		}
	}
	if e.portal.count("POST") != 3 {
		t.Fatalf("%d POST sent", e.portal.count("POST"))
	}
}

// A POST whose answer never comes and whose lookup fails reports the object by its key, not by id.
func TestWrite_CreateLostUnknown(t *testing.T) {
	e := newWriteEnv(t)
	e.setPipeline(handlers.WithWriteTimeout(300 * time.Millisecond))
	e.portal.fail("POST", e.portal.base(), fault{kind: faultHang})
	e.portal.fail("GET", e.portal.base(), fault{status: 200, body: `{"data":[],"pagination":{"hits":0}}`}) // duplicate guard
	e.portal.fail("GET", e.portal.base(), fault{status: 503, body: `{"code":"AK000503","message":"unavailable"}`})
	r := e.call("widget_create", map[string]any{"body": map[string]any{"name": "w1"}})
	mustRefuse(t, r, "outcome_unknown", `Create of widget with name "w1": outcome unknown`)
	if !strings.Contains(r.Text, `look for name "w1" before creating it again`) || strings.Contains(r.Text, "get_by_id") {
		t.Fatalf("text = %s", r.Text)
	}
	if e.portal.count("POST") != 1 {
		t.Fatal("POST retried")
	}
}

// MA-W23, MA-W48: echo diff and config status in the create output.
func TestWrite_EchoDiffAndConfigStatus(t *testing.T) {
	e := newWriteEnv(t)
	e.portal.fail("POST", e.portal.base(), fault{status: 201, body: `{"id":5,"name":"W1","tags":["b","a"],"sites":[{"id":1,"name":"s1"}],"configStatus":{"configValid":false,"messages":["prefix overlaps 10.0.0.0/8"]}}`})
	r := e.call("widget_create", map[string]any{"body": map[string]any{
		"name": "w1", "tags": []any{"a", "b"}, "description": "d",
		"sites": []any{map[string]any{"name": "s1"}, map[string]any{"name": "s2"}},
	}})
	mustOK(t, r)
	for _, want := range []string{`Saved with warnings: "prefix overlaps 10.0.0.0/8"`, "description dropped", "sites: 1 of 2 elements dropped"} {
		if !strings.Contains(r.Text, want) {
			t.Errorf("output lacks %q:\n%s", want, r.Text)
		}
	}
	if strings.Contains(r.Text, "name changed") || strings.Contains(r.Text, "tags changed") || strings.Contains(r.Text, "tags:") {
		t.Fatalf("case or order difference reported: %s", r.Text)
	}
}

// A family whose only list edit is remove by id says how remove works.
func TestWrite_RemoveOnlyDescription(t *testing.T) {
	type gadget struct {
		Name  *string                 `json:"name,omitempty" ak:"required"`
		Sites []testfamily.WidgetSite `json:"sites,omitempty"`
	}
	f := handlers.WriteFamily[gadget]{Name: "gadget", Noun: "gadget", Descriptor: ak.Descriptor{Collection: "gadgets", IDMatchedArrays: []string{"sites"}}}
	for _, tool := range writes.Tools(schemagen.At(time.Now()), f) {
		if tool.Name != "gadget_update" {
			continue
		}
		if !strings.Contains(tool.Description, "list edits") || !strings.Contains(tool.Description, "The remove argument drops elements of the listed arrays by id") ||
			strings.Contains(tool.Description, "Lists under add") || !strings.Contains(string(tool.RawInputSchema), `"remove"`) || strings.Contains(string(tool.RawInputSchema), `"add"`) {
			t.Fatalf("description = %s\nschema = %s", tool.Description, tool.RawInputSchema)
		}
	}
}

// MA-W6: write annotations by verb, every hint set.
func TestWrite_Annotations(t *testing.T) {
	want := map[string][3]bool{"create": {false, false, false}, "update": {false, true, true}, "delete": {false, true, true}}
	var res struct {
		Tools []struct {
			Name        string `json:"name"`
			Annotations struct {
				ReadOnlyHint, DestructiveHint, IdempotentHint, OpenWorldHint *bool
			} `json:"annotations"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(call(t, allWriteServer(t), "tools/list", nil), &res); err != nil {
		t.Fatal(err)
	}
	if want := 3 * (1 + len(writeFamilies)); len(res.Tools) != want && len(res.Tools) != want-3 {
		t.Fatalf("%d tools", len(res.Tools))
	}
	for _, tool := range res.Tools {
		a := tool.Annotations
		if a.ReadOnlyHint == nil || a.DestructiveHint == nil || a.IdempotentHint == nil || a.OpenWorldHint == nil {
			t.Fatalf("%s: nil hint", tool.Name)
		}
		w := want[tool.Name[strings.LastIndex(tool.Name, "_")+1:]]
		if *a.ReadOnlyHint != w[0] || *a.DestructiveHint != w[1] || *a.IdempotentHint != w[2] || !*a.OpenWorldHint {
			t.Errorf("%s: hints %v %v %v %v", tool.Name, *a.ReadOnlyHint, *a.DestructiveHint, *a.IdempotentHint, *a.OpenWorldHint)
		}
	}
}

// writeServer serves the read tools and the test family's write tools.
func writeServer(get func(context.Context) (*ak.AlkiraClient, error)) *server.MCPServer {
	srv := newServer(get)
	p, err := handlers.NewWritePipeline()
	if err != nil {
		panic(err)
	}
	writes.RegisterType(srv, get, p, testfamily.Family)
	return srv
}

// MA-W8: write tools resolve their client through the provider.
func TestWrite_ToolsUseProvider(t *testing.T) {
	srv := writeServer(func(context.Context) (*ak.AlkiraClient, error) { return nil, errorString("provider failed") })
	for _, name := range []string{"widget_create", "widget_update", "widget_delete"} {
		r := callWrite(t, srv, name, map[string]any{})
		if !r.IsError || r.Text != "provider failed" {
			t.Fatalf("%s: %+v", name, r)
		}
	}
}

type errorString string

func (e errorString) Error() string { return string(e) }

// MA-W3: no generated schema holds a secret field; the scan catches a seeded one.
func TestWrite_NoSecretInSchemas(t *testing.T) {
	var res struct {
		Tools []struct {
			Name        string          `json:"name"`
			InputSchema json.RawMessage `json:"inputSchema"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(call(t, allWriteServer(t), "tools/list", nil), &res); err != nil {
		t.Fatal(err)
	}
	for _, tool := range res.Tools {
		if found := secretProperties(tool.InputSchema); found != nil {
			t.Errorf("%s: secret properties %v", tool.Name, found)
		}
	}
	seeded := `{"type":"object","properties":{"body":{"type":"object","properties":{"sites":{"type":"array","items":{"type":"object","properties":{"preSharedKey":{"type":"string"}}}}}}}}`
	if found := secretProperties([]byte(seeded)); !slices.Equal(found, []string{"preSharedKey"}) {
		t.Fatalf("seeded nested secret not found: %v", found)
	}
}

// secretProperties lists every property name in schema that is a known secret key or "psk"/"key" of the test family.
func secretProperties(schema []byte) []string {
	var v any
	json.Unmarshal(schema, &v)
	var found []string
	var walk func(any)
	walk = func(v any) {
		switch t := v.(type) {
		case map[string]any:
			if props, ok := t["properties"].(map[string]any); ok {
				for k := range props {
					if slices.Contains(handlers.SecretKeys(), k) || k == "psk" {
						found = append(found, k)
					}
				}
			}
			for _, e := range t {
				walk(e)
			}
		case []any:
			for _, e := range t {
				walk(e)
			}
		}
	}
	walk(v)
	return found
}

// MA-W33: a constraint tag without src= fails registration.
func TestWrite_TagCitationLint(t *testing.T) {
	type uncited struct {
		Name *string `json:"name,omitempty" ak:"maxlen=10"`
	}
	if _, err := ak.FieldRules[uncited](); err == nil {
		t.Fatal("uncited constraint accepted")
	}
	if _, err := ak.FieldRules[testfamily.Widget](); err != nil {
		t.Fatal(err)
	}
}

// writeGoldenDir holds one golden file per write tool, so a change that adds a
// family adds its own files and touches no shared golden.
const writeGoldenDir = "testdata/tools_list_write"

// allWriteServer serves the write tools of the test family and of every
// served family, with schemas generated at goldenClock.
func allWriteServer(t *testing.T) *server.MCPServer {
	srv := server.NewMCPServer("mcp-alkira", "test")
	p := newPipeline(t, handlers.WithClock(goldenClock))
	family(testfamily.Family).register(srv, noClient, p)
	for _, f := range writeFamilies {
		f.register(srv, noClient, p)
	}
	return srv
}

// goldenClock is the fixed date the golden tests generate schemas at, so an
// expired= date that passes never changes a golden.
func goldenClock() time.Time { return time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC) }

// MA-W7: the write tools generated for the test family and every served family are pinned.
func TestWrite_ToolsListGolden(t *testing.T) {
	var res struct {
		Tools json.RawMessage `json:"tools"`
	}
	if err := json.Unmarshal(call(t, allWriteServer(t), "tools/list", nil), &res); err != nil {
		t.Fatal(err)
	}
	var tools []json.RawMessage
	if err := json.Unmarshal(res.Tools, &tools); err != nil {
		t.Fatal(err)
	}
	got := map[string][]byte{}
	for _, raw := range tools {
		var tool struct{ Name string }
		if err := json.Unmarshal(raw, &tool); err != nil {
			t.Fatal(err)
		}
		var b bytes.Buffer
		if err := json.Indent(&b, raw, "", "  "); err != nil {
			t.Fatal(err)
		}
		b.WriteByte('\n')
		got[tool.Name+".json"] = b.Bytes()
	}
	if *update {
		if err := os.RemoveAll(writeGoldenDir); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(writeGoldenDir, 0o755); err != nil {
			t.Fatal(err)
		}
		for name, b := range got {
			if err := os.WriteFile(filepath.Join(writeGoldenDir, name), b, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	entries, err := os.ReadDir(writeGoldenDir)
	if err != nil {
		t.Fatalf("%v (run with -update to create it)", err)
	}
	for _, e := range entries {
		if _, ok := got[e.Name()]; !ok {
			t.Errorf("%s/%s has no served tool; run with -update if the change is intended", writeGoldenDir, e.Name())
		}
	}
	for name, b := range got {
		want, err := os.ReadFile(filepath.Join(writeGoldenDir, name))
		if err != nil || !bytes.Equal(b, want) {
			t.Errorf("write tool %s differs from %s/%s; run with -update if the change is intended", strings.TrimSuffix(name, ".json"), writeGoldenDir, name)
		}
	}
}

// MA-W7: registering the write tools leaves every read tool entry unchanged.
func TestWrite_ReadEntriesUnchanged(t *testing.T) {
	var read, both struct {
		Tools []json.RawMessage `json:"tools"`
	}
	json.Unmarshal(call(t, newServer(noClient), "tools/list", nil), &read)
	json.Unmarshal(call(t, writeServer(noClient), "tools/list", nil), &both)
	if len(both.Tools) != len(read.Tools)+3 {
		t.Fatalf("read %d, with writes %d", len(read.Tools), len(both.Tools))
	}
	set := map[string]bool{}
	for _, r := range both.Tools {
		set[string(r)] = true
	}
	for _, r := range read.Tools {
		if !set[string(r)] {
			t.Fatalf("read entry changed: %s", r)
		}
	}
}

const asyncFlag = ak.AsyncValidationFlag

// MA-W42: an async save whose validation is still running names the state id and the read tool.
func TestWrite_AsyncValidating(t *testing.T) {
	e := newWriteEnv(t)
	stored(e)
	e.portal.flags = []string{asyncFlag}
	e.setPipeline(handlers.WithWriteTimeout(5 * time.Second)) // too short for a 10 s poll: the result is "validating" at once
	r := e.call("widget_create", map[string]any{"body": map[string]any{"name": "w1"}})
	mustOK(t, r)
	for _, want := range []string{`Created widget "w1" with id 101.`, "Saved; validation is still running (state id WIDGET:DEFAULT:101).",
		"Tenant feature flags that shaped the request: " + asyncFlag + "."} {
		if !strings.Contains(r.Text, want) {
			t.Fatalf("create lacks %q:\n%s", want, r.Text)
		}
	}
	if q := e.portal.requests()[1].Query; q != "async=true" {
		t.Fatalf("POST query = %q", q)
	}
	r = e.call("widget_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}})
	mustOK(t, r)
	if !strings.Contains(r.Text, "Updated widget id 7.") || !strings.Contains(r.Text, "state id WIDGET:DEFAULT:7") {
		t.Fatalf("update text = %s", r.Text)
	}
}

// MA-W42: a FAILED async validation keeps the save and says how to fix or remove it.
func TestWrite_AsyncFailed(t *testing.T) {
	failedWith := func(id, msg string) string {
		return `{"id":3,"networkEntityId":"` + id + `","state":"FAILED","errorDetails":{"code":"AK000400","id":"ALK-5","message":"` + msg + `","errors":[{"field":"cidr","code":"x","message":"overlaps 10.0.0.0/8"}]}}`
	}
	failed := func(id string) string { return failedWith(id, "Validation failed") }
	t.Run("create", func(t *testing.T) {
		t.Parallel()
		e := newWriteEnv(t)
		e.portal.flags = []string{asyncFlag}
		e.portal.validation = map[string]string{"WIDGET:DEFAULT:101": failed("101")}
		r := e.call("widget_create", map[string]any{"body": map[string]any{"name": "w1"}})
		mustOK(t, r)
		want := `Saved widget "w1" as id 101 (validation state id WIDGET:DEFAULT:101), but validation FAILED: "Validation failed; cidr: overlaps 10.0.0.0/8" (code AK000400, error id ALK-5); fix it with widget_update or remove it with widget_delete.`
		if !strings.Contains(r.Text, want) || strings.Contains(r.Text, "not saved") || strings.Contains(r.Text, "Created widget") {
			t.Fatalf("text = %s", r.Text)
		}
	})
	t.Run("update", func(t *testing.T) {
		t.Parallel()
		e := newWriteEnv(t)
		stored(e)
		e.portal.flags = []string{asyncFlag}
		// The stored psk is resent with the update; the backend message echoes it and a nested site key.
		e.portal.validation = map[string]string{"WIDGET:DEFAULT:7": failedWith("7", "psk s3cr3t-psk and key site-key-1 rejected")}
		r := e.call("widget_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}})
		mustOK(t, r)
		if !strings.Contains(r.Text, "Saved the update of widget id 7 (validation state id WIDGET:DEFAULT:7), but validation FAILED: \"psk *** and key *** rejected; cidr: overlaps 10.0.0.0/8\" (code AK000400") ||
			strings.Contains(r.Text, "not saved") || strings.Contains(r.Text, "s3cr3t-psk") || strings.Contains(r.Text, "site-key-1") {
			t.Fatalf("text = %s", r.Text)
		}
	})
	t.Run("create without an id in the answer", func(t *testing.T) {
		t.Parallel()
		e := newWriteEnv(t)
		e.portal.flags = []string{asyncFlag}
		e.portal.validation = map[string]string{"WIDGET:DEFAULT:101": failed("101")}
		empty := fault{status: 200, body: `{"data":[],"pagination":{"hits":0}}`}
		e.portal.fail("GET", e.portal.base(), empty) // duplicate guard
		e.portal.fail("GET", e.portal.base(), empty) // id lookup after an answer without id
		e.portal.fail("POST", e.portal.base(), fault{kind: faultRewrite, status: 202, body: `{}`, header: map[string]string{"x-ak-validation-state-id": "WIDGET:DEFAULT:101"}})
		r := e.call("widget_create", map[string]any{"body": map[string]any{"name": "w1"}})
		mustOK(t, r)
		if !strings.Contains(r.Text, `Saved widget "w1" as id 101 (validation state id WIDGET:DEFAULT:101), but validation FAILED`) {
			t.Fatalf("text = %s", r.Text)
		}
	})
}

// A patch whose only field a flag left out sends nothing and says why.
func TestWrite_UpdateOnlyDroppedField(t *testing.T) {
	e := newWriteEnv(t)
	stored(e)
	r := e.call("widget_update", map[string]any{"id": "7", "patch": map[string]any{"fast": true}})
	mustOK(t, r)
	for _, want := range []string{"No change to widget id 7: fast left out because its tenant feature is off; nothing sent.", "Left out because their tenant feature is off: fast."} {
		if !strings.Contains(r.Text, want) {
			t.Fatalf("text lacks %q:\n%s", want, r.Text)
		}
	}
	if strings.Contains(r.Text, "stored lists") || e.portal.count("PUT") != 0 {
		t.Fatalf("text = %s, requests = %+v", r.Text, e.portal.requests())
	}
}

// Lost responses keep the flag effects of the request that was sent.
func TestWrite_LostResponseFlagEffects(t *testing.T) {
	t.Run("applied with a dropped field", func(t *testing.T) {
		e := newWriteEnv(t)
		stored(e)
		e.portal.flags = []string{"TEST-SIZE"}
		e.portal.fail("PUT", e.portal.base()+"/7", fault{kind: faultCommit})
		r := e.call("widget_update", map[string]any{"id": "7", "patch": map[string]any{"description": "new", "fast": true, "size": "LARGE"}})
		mustOK(t, r)
		for _, want := range []string{"Updated widget id 7: applied", "Tenant feature flags that shaped the request: TEST-FAST-PATH, TEST-SIZE.", "Left out because their tenant feature is off: fast."} {
			if !strings.Contains(r.Text, want) {
				t.Fatalf("text lacks %q:\n%s", want, r.Text)
			}
		}
		if e.portal.flagReads != 1 {
			t.Fatalf("%d flag reads, want 1", e.portal.flagReads)
		}
	})
	t.Run("async update may still apply", func(t *testing.T) {
		e := newWriteEnv(t)
		stored(e)
		e.portal.flags = []string{asyncFlag}
		e.portal.fail("PUT", e.portal.base()+"/7", fault{kind: faultReset})
		r := e.call("widget_update", map[string]any{"id": "7", "patch": map[string]any{"description": "new"}})
		mustRefuse(t, r, "may_still_apply", "validation may still be running; check the object with widget_get_by_id.")
		if !strings.Contains(r.Text, "shaped the request: "+asyncFlag) {
			t.Fatalf("text = %s", r.Text)
		}
	})
	t.Run("async create found by lookup", func(t *testing.T) {
		e := newWriteEnv(t)
		e.portal.flags = []string{asyncFlag}
		e.portal.fail("POST", e.portal.base(), fault{kind: faultCommit})
		r := e.call("widget_create", map[string]any{"body": map[string]any{"name": "w1", "fast": true}})
		mustOK(t, r)
		for _, want := range []string{`Created widget with name "w1", id 101`, "Left out because their tenant feature is off: fast.",
			"Validation state unknown", "may have FAILED. Check the object with widget_get_by_id before provisioning."} {
			if !strings.Contains(r.Text, want) {
				t.Fatalf("text lacks %q:\n%s", want, r.Text)
			}
		}
	})
}

// MA-W43: a 403 while the previous validation runs gives the validation text, never the permission text, with no retry.
func TestWrite_ValidationPending(t *testing.T) {
	const pending = `{"code":"AK000403","id":"ALK-8","message":"Connector cannot be modified until the validation completes"}`
	e := newWriteEnv(t)
	stored(e)
	e.portal.fail("PUT", e.portal.base()+"/7", fault{status: 403, body: pending})
	r := e.call("widget_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}})
	mustRefuse(t, r, "validation_pending", "Validation of the previous change to widget id 7 is still running; retry the update after it finishes. Validation still running for this object.")
	code := previewCode(t, e.call("widget_delete", map[string]any{"id": "7"}))
	e.portal.fail("DELETE", e.portal.base()+"/7", fault{status: 403, body: pending})
	r = e.call("widget_delete", map[string]any{"id": "7", "confirmCode": code})
	mustRefuse(t, r, "validation_pending", `Validation of widget "w7" (id 7) is still running; retry the delete after it finishes.`)
	if strings.Contains(r.Text, "Permission denied") || e.portal.count("PUT") != 1 || e.portal.count("DELETE") != 1 {
		t.Fatalf("text = %s, requests = %+v", r.Text, e.portal.requests())
	}
}

// MA-W45: output names the flags that shaped the request and the fields left out; refusals by flag.
func TestWrite_FlagEffects(t *testing.T) {
	t.Run("two flags on", func(t *testing.T) {
		e := newWriteEnv(t)
		e.portal.flags = []string{"TEST-FAST-PATH", "TEST-SIZE"}
		r := e.call("widget_create", map[string]any{"body": map[string]any{"name": "w1", "fast": true, "size": "LARGE"}})
		mustOK(t, r)
		if !strings.Contains(r.Text, "Tenant feature flags that shaped the request: TEST-FAST-PATH, TEST-SIZE.") {
			t.Fatalf("text = %s", r.Text)
		}
		var body map[string]any
		json.Unmarshal([]byte(e.portal.requests()[1].Body), &body)
		if body["fast"] != true || body["size"] != "LARGE" {
			t.Fatalf("POST body = %v", body)
		}
	})
	t.Run("flag off drops the field", func(t *testing.T) {
		e := newWriteEnv(t)
		r := e.call("widget_create", map[string]any{"body": map[string]any{"name": "w1", "fast": true}})
		mustOK(t, r)
		if !strings.Contains(r.Text, "Left out because their tenant feature is off: fast.") || strings.Contains(r.Text, "fast dropped") {
			t.Fatalf("text = %s", r.Text)
		}
		if strings.Contains(e.portal.requests()[1].Body, "fast") {
			t.Fatalf("POST body = %s", e.portal.requests()[1].Body)
		}
	})
	t.Run("flag read failure", func(t *testing.T) {
		e := newWriteEnv(t)
		e.portal.flagStatus = 500
		mustRefuse(t, e.call("widget_create", map[string]any{"body": map[string]any{"name": "w1", "fast": true}}), "flags_unavailable",
			"Cannot check whether changes through MCP are allowed for this tenant; try again. Nothing was sent.")
		if e.portal.count("POST") != 0 {
			t.Fatalf("requests = %+v", e.portal.requests())
		}
	})
	t.Run("refused field", func(t *testing.T) {
		e := newWriteEnv(t)
		mustRefuse(t, e.call("widget_create", map[string]any{"body": map[string]any{"name": "w1", "asn": 65001}}), "not_enabled",
			"Field asn is not enabled for this tenant (tenant feature TEST-ASN is off); nothing was sent.")
		if e.portal.count("POST") != 0 {
			t.Fatal("POST sent")
		}
	})
	t.Run("type not enabled", func(t *testing.T) {
		e := newWriteEnv(t)
		type gadget struct {
			Name *string `json:"name,omitempty" ak:"required"`
		}
		f := handlers.WriteFamily[gadget]{Name: "gadget", Noun: "gadget", Descriptor: ak.Descriptor{Collection: "gadgets",
			Flags: []ak.FlagRule{{Flag: "TEST-GADGETS", Effect: ak.FlagTypeAvailable, UISource: "testfamily.go:1"}}}}
		srv := server.NewMCPServer("mcp-alkira", "test")
		writes.RegisterType(srv, providerFor(e.portal.client()), e.p, f)
		mustRefuse(t, callWrite(t, srv, "gadget_create", map[string]any{"body": map[string]any{"name": "g"}}), "not_enabled",
			"Gadget is not enabled for this tenant (tenant feature TEST-GADGETS is off); nothing was sent.")
		if n := e.portal.count("POST"); n != 0 {
			t.Fatal("POST sent")
		}
	})
}

// A 4xx answer to an update that echoes a stored secret, which the PUT
// resends, has it masked.
func TestWrite_BackendTextMasksSecrets(t *testing.T) {
	t.Run("update", func(t *testing.T) {
		e := newWriteEnv(t)
		stored(e)
		e.portal.fail("PUT", e.portal.base()+"/7", fault{status: 400, body: `{"code":"AK000400","message":"invalid psk s3cr3t-psk","errors":[{"field":"sites","message":"key site-key-1 too short"}]}`})
		r := e.call("widget_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}})
		mustRefuse(t, r, "backend_error", `Update widget failed with status 400: "invalid psk ***; sites: key *** too short" (code AK000400`)
		if strings.Contains(r.Text, "s3cr3t-psk") || strings.Contains(r.Text, "site-key-1") {
			t.Fatalf("secret in output: %s", r.Text)
		}
	})
	t.Run("update with a raw body", func(t *testing.T) {
		e := newWriteEnv(t)
		stored(e)
		e.portal.fail("PUT", e.portal.base()+"/7", fault{status: 400, body: `bad request: {"psk":"s3cr3t-psk"}`})
		r := e.call("widget_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}})
		if !r.IsError || strings.Contains(r.Text, "s3cr3t-psk") {
			t.Fatalf("secret in output: %s", r.Text)
		}
	})
}

// HS19: without a signing key the delete tool refuses and sends no request.
func TestWrite_DeleteWithoutSigner(t *testing.T) {
	portal := newWritePortal(t, "widgets")
	portal.put(7, map[string]any{"name": "w7", "state": "PENDING"})
	p := newPipeline(t, handlers.WithoutConfirm(), handlers.WithWriteTimeout(time.Minute))
	srv := server.NewMCPServer("mcp-alkira", "test")
	writes.RegisterType(srv, providerFor(portal.client()), p, testfamily.Family)
	for _, args := range []map[string]any{{"id": "7"}, {"id": "7", "confirmCode": "c29tZS1jb2Rl"}} {
		mustRefuse(t, callWrite(t, srv, "widget_delete", args), handlers.RefusalConfirmUnavailable, "nothing was sent")
	}
	if n := len(portal.requests()); n != 0 {
		t.Fatalf("%d requests sent", n)
	}
}

// A family without a duplicate guard does not promise one in its create description.
func TestWrite_CreateDescriptionWithoutKey(t *testing.T) {
	type gadget struct {
		Name *string `json:"name,omitempty"`
	}
	f := handlers.WriteFamily[gadget]{Name: "gadget", Noun: "gadget", Descriptor: ak.Descriptor{Collection: "gadgets"}}
	desc := writes.Tools(schemagen.At(time.Now()), f)[0].Description
	if strings.Contains(desc, "Refused with no request") || strings.Contains(desc, "  ") {
		t.Fatalf("description = %q", desc)
	}
}

// MA-W42: a validation state read that a retry cannot change ends the poll;
// the save stands and the output says the result is unknown.
func TestWrite_AsyncValidationReadError(t *testing.T) {
	t.Parallel()
	e := newWriteEnv(t)
	e.portal.flags = []string{asyncFlag}
	e.portal.fail("GET", "/api/tenantnetworks/7/validation-states/WIDGET:DEFAULT:101", fault{status: 404, body: `{"code":"AK000003","message":"Entity not found"}`})
	r := e.call("widget_create", map[string]any{"body": map[string]any{"name": "w1"}})
	mustOK(t, r)
	if !strings.Contains(r.Text, `Created widget "w1" with id 101.`) ||
		!strings.Contains(r.Text, "Saved; the validation state (id WIDGET:DEFAULT:101) could not be read, so its result is unknown: ") ||
		strings.Contains(r.Text, "validation is still running") {
		t.Fatalf("text = %s", r.Text)
	}
}
