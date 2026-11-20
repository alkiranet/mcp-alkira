package registry

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
	"github.com/alkiranet/mcp-alkira/internal/confirm"
	"github.com/alkiranet/mcp-alkira/tools"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/mark3labs/mcp-go/server"
)

// provPortal fakes the TPS and policy-service endpoints the provisioning
// tools read and the two POSTs they send, and journals every request.
type provPortal struct {
	t   *testing.T
	srv *httptest.Server
	mu  sync.Mutex

	lastState  string // "" answers a body-less 404 (never provisioned)
	tps, pol   string
	states     string // validation-states page data
	user       string // "" answers 500
	flags      []string
	failed     map[string]string // FAILED request id -> GET body
	requests   map[string]string // provision request id -> GET body
	buffered   map[string]string // buffered id -> GET body
	valStates  map[string]string // validation state id -> GET body
	valReqs    map[string]string // validation request id -> GET body
	provision  fault
	validation fault
	filed      []string         // provision request ids, newest first; a 2xx provision POST files one
	fail       map[string]fault // path under /api -> answer in place of the fake's own
	journal    []journalEntry
}

var portalSeq atomic.Int64

func newProvPortal(t *testing.T) *provPortal {
	t.Helper()
	p := &provPortal{t: t, lastState: "SUCCESS", tps: "[]", pol: "[]", states: "[]", user: "u-me",
		failed: map[string]string{}, requests: map[string]string{}, buffered: map[string]string{},
		valStates: map[string]string{}, valReqs: map[string]string{},
		provision:  fault{status: 202, header: map[string]string{"x-provision-request-id": "pr-1"}, body: `{"id":7}`},
		validation: fault{status: 202, header: map[string]string{"x-validation-request-id": "vr-1"}, body: `{"validationRequestId":"vr-1","state":"IN_PROGRESS"}`}}
	p.srv = httptest.NewServer(http.HandlerFunc(p.serve))
	t.Cleanup(p.srv.Close)
	return p
}

// client returns a client for the fake; host names the tenant.
func (p *provPortal) client(host string) *ak.AlkiraClient {
	hc := retryablehttp.NewClient()
	hc.RetryMax = 0
	hc.Logger = nil
	if host == "" {
		// The flag cache is keyed by host, so each portal gets its own.
		host = fmt.Sprintf("tenant-%d", portalSeq.Add(1))
	}
	return &ak.AlkiraClient{Client: hc, URI: p.srv.URL + "/api", TenantNetworkId: "7", Authorization: "api-key test", HostHeader: host}
}

func (p *provPortal) set(f func(p *provPortal)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	f(p)
}

func (p *provPortal) entries() []journalEntry {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]journalEntry(nil), p.journal...)
}

func (p *provPortal) posts() []journalEntry {
	var out []journalEntry
	for _, e := range p.entries() {
		if e.Method != "GET" {
			out = append(out, e)
		}
	}
	return out
}

func (p *provPortal) serve(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	p.mu.Lock()
	defer p.mu.Unlock()
	path := strings.TrimPrefix(r.URL.Path, "/api")
	if path != "/tenantenabledfeatures" {
		p.journal = append(p.journal, journalEntry{r.Method, r.URL.Path, r.URL.RawQuery, string(body)})
	}
	answer := func(status int, b string) {
		w.WriteHeader(status)
		io.WriteString(w, b)
	}
	serveFault := func(f fault) {
		for k, v := range f.header {
			w.Header().Set(k, v)
		}
		answer(f.status, f.body)
	}
	if f, ok := p.fail[path]; ok {
		serveFault(f)
		return
	}
	page := func(data string) string {
		var rows []json.RawMessage
		_ = json.Unmarshal([]byte(data), &rows)
		return fmt.Sprintf(`{"data":%s,"pagination":{"offset":0,"limit":100,"hits":%d}}`, data, len(rows))
	}
	const tn = "/tenantnetworks/7/"
	switch {
	case path == "/tenantenabledfeatures":
		b, _ := json.Marshal(p.flags)
		answer(200, string(b))
	case path == "/user/self":
		if p.user == "" {
			answer(500, `{"message":"Request failed"}`)
			return
		}
		answer(200, fmt.Sprintf(`{"id":%q,"userName":"me@example.com"}`, p.user))
	case path == tn+"last-provision-state":
		if p.lastState == "" {
			w.WriteHeader(404)
			return
		}
		answer(200, fmt.Sprintf(`{"state":%q}`, p.lastState))
	case path == tn+"changes":
		answer(200, p.tps)
	case path == tn+"networkpolicy/changes":
		answer(200, p.pol)
	case path == tn+"validation-states":
		answer(200, page(p.states))
	case strings.HasPrefix(path, tn+"validation-states/"):
		if b, ok := p.valStates[strings.TrimPrefix(path, tn+"validation-states/")]; ok {
			answer(200, b)
			return
		}
		answer(200, ``)
	case path == tn+"provision-requests":
		var rows []string
		if r.URL.Query().Get("state") == "FAILED" {
			for id := range p.failed {
				rows = append(rows, fmt.Sprintf(`{"id":%q,"state":"FAILED","createdAt":1}`, id))
			}
		} else {
			for _, id := range p.filed {
				rows = append(rows, fmt.Sprintf(`{"id":%q,"state":"PENDING"}`, id))
			}
		}
		answer(200, page("["+strings.Join(rows, ",")+"]"))
	case strings.HasPrefix(path, tn+"provision-requests/"):
		id := strings.TrimPrefix(path, tn+"provision-requests/")
		if b, ok := p.failed[id]; ok {
			answer(200, b)
		} else if b, ok := p.requests[id]; ok {
			answer(200, b)
		} else {
			answer(404, `{"code":"AK000003","message":"not found"}`)
		}
	case strings.HasPrefix(path, tn+"buffered-provision-requests/"):
		if b, ok := p.buffered[strings.TrimPrefix(path, tn+"buffered-provision-requests/")]; ok {
			answer(200, b)
			return
		}
		answer(404, `{"code":"AK000003","message":"not found"}`)
	case path == tn+"provision" && r.Method == "POST":
		if p.provision.status/100 == 2 {
			p.filed = append([]string{fmt.Sprintf("pr-%d", len(p.filed)+1)}, p.filed...)
		}
		serveFault(p.provision)
	case path == tn+"validation-requests" && r.Method == "POST":
		serveFault(p.validation)
	case strings.HasPrefix(path, tn+"validation-requests/"):
		if b, ok := p.valReqs[strings.TrimPrefix(path, tn+"validation-requests/")]; ok {
			answer(200, b)
			return
		}
		answer(404, `{"code":"AK000003","message":"not found"}`)
	default:
		answer(501, `{"message":"unexpected request"}`)
	}
}

// provEnv serves the provisioning tools of one pipeline with a fixed clock.
type provEnv struct {
	t      *testing.T
	portal *provPortal
	p      *handlers.WritePipeline
	now    time.Time
	srv    *server.MCPServer
	key    []byte
}

func newProvEnv(t *testing.T) *provEnv {
	t.Helper()
	e := &provEnv{t: t, portal: newProvPortal(t), now: time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC), key: confirm.RandomKey()}
	p, err := handlers.NewWritePipeline(handlers.WithConfirmKey(e.key), handlers.WithClock(func() time.Time { return e.now }))
	if err != nil {
		t.Fatal(err)
	}
	e.p = p
	e.srv = e.server(e.portal.client("tenant-a"))
	return e
}

func (e *provEnv) server(client *ak.AlkiraClient) *server.MCPServer {
	srv := server.NewMCPServer("mcp-alkira", "test")
	get := providerFor(client)
	srv.AddTool(tools.TenantNetworkChangesGet(), perCall(get, handlers.TenantNetworkChangesGet))
	srv.AddTool(tools.TenantNetworkProvisionPreview(), perCall(get, handlers.ProvisionPreview(e.p)))
	srv.AddTool(tools.TenantNetworkProvision(), perCall(get, handlers.Provision(e.p)))
	srv.AddTool(tools.TenantNetworkValidate(), perCall(get, handlers.Validate(e.p)))
	srv.AddTool(tools.TenantNetworkProvisionRequestGet(), perCall(get, handlers.ProvisionRequestGet))
	srv.AddTool(tools.TenantNetworkValidationRequestGet(), perCall(get, handlers.ValidationRequestGet))
	srv.AddTool(tools.TenantNetworkValidationStateGet(), perCall(get, handlers.ValidationStateGet))
	return srv
}

func (e *provEnv) call(name string, args map[string]any) writeResult {
	e.t.Helper()
	return callWrite(e.t, e.srv, name, args)
}

// The MA-W56 fixture: a segment, a connector in it, a NAT policy, an implicit
// group, a remote access connector, a remote access template, a CXP implicit
// reference and a policy-service prefix list.
const (
	chSegment   = `{"entity":{"id":"12","name":"s1","type":"SEGMENT","lastModifiedAt":100,"properties":{"id":12,"name":"s1","lastConfigUpdatedAt":100}},"changeType":"ADDED","lastUpdatedByUserId":"u-other","lastUpdatedBy":"other@example.com"}`
	chConnector = `{"entity":{"id":"1303","name":"c1","type":"CONNECTOR","subType":"IP_SEC","lastModifiedAt":101,"properties":{"id":1303,"name":"c1","type":"IP_SEC","segmentIds":[12],"segments":["s1"],"implicitGroupId":60025,"lastConfigUpdatedAt":101}},"changeType":"ADDED","bestEffortAllowed":true,"lastUpdatedByUserId":"u-me","lastUpdatedBy":"me@example.com"}`
	chNAT       = `{"entity":{"id":"40","name":"n1","type":"NAT_POLICY","lastModifiedAt":102,"properties":{"id":40,"name":"n1","segment":"s9","lastConfigUpdatedAt":102}},"changeType":"MODIFIED","lastUpdatedByUserId":"u-me"}`
	chImplicitG = `{"entity":{"id":"60025","name":"c1-group","type":"GROUP","properties":{"id":60025,"implicit":true}},"changeType":"ADDED","lastUpdatedByUserId":"u-me"}`
	chRemote    = `{"entity":{"id":"1400","name":"ra1","type":"CONNECTOR","subType":"REMOTE_ACCESS","properties":{"id":1400,"type":"REMOTE_ACCESS"}},"changeType":"ADDED","systemGenerated":false}`
	chTemplate  = `{"entity":{"id":"7","name":"ra-template","type":"CONNECTOR_TEMPLATE","subType":"REMOTE_ACCESS","properties":{"id":7,"systemGenerated":true}},"changeType":"MODIFIED"}`
	chCXP       = `{"entity":{"id":"USEAST-AZURE-1","name":"USEAST-AZURE-1","type":"CXP","properties":{"id":"x"},"implicitReference":true},"changeType":"ADDED"}`
	chPrefix    = `{"entity":{"id":"9","name":"p1","type":"PREFIX_LIST","properties":{"id":9,"name":"p1"}},"changeType":"ADDED","lastUpdatedByUserId":"system","lastUpdatedBy":"system"}`
)

func (e *provEnv) fixture() {
	e.portal.set(func(p *provPortal) {
		p.tps = "[" + strings.Join([]string{chSegment, chConnector, chNAT, chImplicitG, chRemote, chTemplate, chCXP}, ",") + "]"
		p.pol = "[" + chPrefix + "]"
	})
}

// MA-W56: preview set and dependencies, zero POST.
func TestProvision_PreviewSetAndDependencies(t *testing.T) {
	e := newProvEnv(t)
	e.fixture()

	r := e.call("tenant_network_provision_preview", map[string]any{"select": []any{"CONNECTOR:1303"}})
	previewCode(t, r)
	for _, want := range []string{
		"Provision preview. Nothing was provisioned.",
		"Network last provision state: SUCCESS.",
		`* CONNECTOR:1303 ADDED "c1" (IPsec), by caller`,
		`  SEGMENT:12 ADDED "s1", by another user other@example.com [OTHER USER]`,
		`Required by the backend, not selected: SEGMENT:12 "s1" (needed by CONNECTOR:1303).`,
		`Request body: [{"changeType":"ADDED","entity":{"id":"1303","type":"CONNECTOR","subType":"IPsec"}}]`,
		"expires: 2026-11-19T12:05:00Z",
	} {
		if !strings.Contains(r.Text, want) {
			t.Errorf("preview lacks %q:\n%s", want, r.Text)
		}
	}
	if strings.Contains(r.Text, "GROUP:60025") {
		t.Errorf("implicit group listed:\n%s", r.Text)
	}
	paths := map[string]bool{}
	for _, j := range e.portal.entries() {
		paths[j.Path] = true
		if j.Method != "GET" {
			t.Errorf("preview sent %s %s", j.Method, j.Path)
		}
	}
	for _, p := range []string{"/api/tenantnetworks/7/changes", "/api/tenantnetworks/7/networkpolicy/changes", "/api/tenantnetworks/7/last-provision-state", "/api/tenantnetworks/7/validation-states"} {
		if !paths[p] {
			t.Errorf("preview did not read %s", p)
		}
	}

	r = e.call("tenant_network_provision_preview", map[string]any{"select": "all"})
	previewCode(t, r)
	want := `Request body: [{"changeType":"ADDED","entity":{"id":"1303","type":"CONNECTOR","subType":"IPsec"}},` +
		`{"changeType":"MODIFIED","entity":{"id":"40","type":"NAT_POLICY"}},` +
		`{"changeType":"MODIFIED","entity":{"id":"7","type":"CONNECTOR_TEMPLATE"},"subType":"REMOTE_ACCESS"},` +
		`{"changeType":"ADDED","entity":{"id":"9","type":"PREFIX_LIST"}},` +
		`{"changeType":"ADDED","entity":{"id":"12","type":"SEGMENT"}}]`
	if !strings.Contains(r.Text, want) {
		t.Errorf("all: body\n%s", r.Text)
	}
	if strings.Contains(r.Text, "Required by the backend") {
		t.Errorf("all: dependencies listed:\n%s", r.Text)
	}
	for _, key := range []string{"* CONNECTOR:1400", "* SEGMENT:12", "* CXP:USEAST-AZURE-1", "* PREFIX_LIST:9"} {
		if !strings.Contains(r.Text, key) {
			t.Errorf("all: %s not selected:\n%s", key, r.Text)
		}
	}

	mustRefuse(t, e.call("tenant_network_provision_preview", map[string]any{"select": []any{"SEGMENT:99"}}), "invalid_selection", "not pending: SEGMENT:99")
	mustRefuse(t, e.call("tenant_network_provision_preview", map[string]any{"select": "some"}), "invalid_arguments", "select must be")
	if n := len(e.portal.posts()); n != 0 {
		t.Errorf("%d POSTs", n)
	}
}

// MA-W59: attribution of each change; mine selects the caller's rows only.
func TestProvision_Attribution(t *testing.T) {
	e := newProvEnv(t)
	e.portal.set(func(p *provPortal) {
		p.tps = `[{"entity":{"id":"1","name":"mine","type":"SEGMENT","properties":{}},"changeType":"ADDED","lastUpdatedByUserId":"u-me"},` +
			`{"entity":{"id":"2","name":"theirs","type":"SEGMENT","properties":{}},"changeType":"ADDED","lastUpdatedByUserId":"u-2","lastUpdatedBy":"two@example.com"},` +
			`{"entity":{"id":"3","name":"sys","type":"SEGMENT","properties":{"systemGenerated":true}},"changeType":"ADDED"},` +
			`{"entity":{"id":"4","name":"none","type":"SEGMENT","properties":{}},"changeType":"ADDED"}]`
	})
	r := e.call("tenant_network_changes_get", map[string]any{})
	mustOK(t, r)
	for _, want := range []string{
		`SEGMENT:1 ADDED "mine", by caller`,
		`SEGMENT:2 ADDED "theirs", by another user two@example.com [OTHER USER]`,
		`SEGMENT:3 ADDED "sys", by system`,
		`SEGMENT:4 ADDED "none", by untracked`,
		"Changes by other users: 1.",
	} {
		if !strings.Contains(r.Text, want) {
			t.Errorf("changes lack %q:\n%s", want, r.Text)
		}
	}
	r = e.call("tenant_network_provision_preview", map[string]any{"select": "mine"})
	previewCode(t, r)
	if !strings.Contains(r.Text, `Request body: [{"changeType":"ADDED","entity":{"id":"1","type":"SEGMENT"}}]`) {
		t.Errorf("mine body:\n%s", r.Text)
	}
	e.portal.set(func(p *provPortal) { p.user = "" })
	mustRefuse(t, e.call("tenant_network_provision_preview", map[string]any{"select": "mine"}), "caller_unknown", `select "mine" needs the caller's user id`)
	if n := len(e.portal.posts()); n != 0 {
		t.Errorf("%d POSTs", n)
	}
}

// MA-W13: the provision re-reads the state and both lists and sends one
// POST whose body equals the portal's for the previewed set.
func TestProvision_SendsPortalBody(t *testing.T) {
	e := newProvEnv(t)
	e.fixture()
	sel := []any{"CONNECTOR:1303", "SEGMENT:12", "CONNECTOR_TEMPLATE:7", "CXP:USEAST-AZURE-1", "CONNECTOR:1400"}
	code := previewCode(t, e.call("tenant_network_provision_preview", map[string]any{"select": sel}))
	before := len(e.portal.entries())

	r := e.call("tenant_network_provision", map[string]any{"select": sel, "confirmCode": code, "description": "mcp change 1"})
	mustOK(t, r)
	if !strings.Contains(r.Text, "Provision requested for 3 changes; request id pr-1.") {
		t.Errorf("result: %s", r.Text)
	}
	var reads []string
	for _, j := range e.portal.entries()[before:] {
		if j.Method == "GET" {
			reads = append(reads, j.Path)
		}
	}
	for _, p := range []string{"/api/tenantnetworks/7/last-provision-state", "/api/tenantnetworks/7/changes", "/api/tenantnetworks/7/networkpolicy/changes", "/api/tenantnetworks/7/validation-states"} {
		if !strings.Contains(strings.Join(reads, " "), p) {
			t.Errorf("provision did not re-read %s: %v", p, reads)
		}
	}
	posts := e.portal.posts()
	if len(posts) != 1 {
		t.Fatalf("%d POSTs", len(posts))
	}
	post := posts[0]
	if post.Path != "/api/tenantnetworks/7/provision" || post.Query != "description=mcp+change+1&selectedOnly=true" {
		t.Errorf("POST %s?%s", post.Path, post.Query)
	}
	want := `[{"changeType":"ADDED","entity":{"id":"1303","type":"CONNECTOR","subType":"IPsec"}},` +
		`{"changeType":"MODIFIED","entity":{"id":"7","type":"CONNECTOR_TEMPLATE"},"subType":"REMOTE_ACCESS"},` +
		`{"changeType":"ADDED","entity":{"id":"12","type":"SEGMENT"}}]`
	if post.Body != want {
		t.Errorf("body\n got %s\nwant %s", post.Body, want)
	}

	// RELEASE-PROV-BEST-EFFORT-CLEANUP adds bestEffortAllowed and the STANDARD cleanup type where allowed.
	e.portal.set(func(p *provPortal) { p.flags = []string{"RELEASE-PROV-BEST-EFFORT-CLEANUP"} })
	e.srv = e.server(e.portal.client("tenant-b"))
	code = previewCode(t, e.call("tenant_network_provision_preview", map[string]any{"select": []any{"CONNECTOR:1303", "SEGMENT:12"}}))
	mustOK(t, e.call("tenant_network_provision", map[string]any{"select": []any{"SEGMENT:12", "CONNECTOR:1303"}, "confirmCode": code}))
	posts = e.portal.posts()
	if got := posts[len(posts)-1]; got.Query != "selectedOnly=true" ||
		got.Body != `[{"changeType":"ADDED","entity":{"id":"1303","type":"CONNECTOR","subType":"IPsec"},"bestEffortAllowed":true,"cleanupType":"REQUIRED"},{"changeType":"ADDED","entity":{"id":"12","type":"SEGMENT"}}]` {
		t.Errorf("best effort: ?%s %s", got.Query, got.Body)
	}

	for _, j := range e.portal.entries() {
		if strings.Contains(j.Query, "force") || strings.Contains(j.Query, "includeChanges") || strings.Contains(j.Query, "provision=true") {
			t.Errorf("%s %s?%s", j.Method, j.Path, j.Query)
		}
		if j.Method == "POST" && strings.TrimSpace(j.Body) == "" {
			t.Errorf("POST %s with an empty body", j.Path)
		}
	}
}

// MA-W13: refusals and backend answers of the provision.
func TestProvision_Refusals(t *testing.T) {
	e := newProvEnv(t)
	e.fixture()
	sel := []any{"CONNECTOR:1303"}
	code := previewCode(t, e.call("tenant_network_provision_preview", map[string]any{"select": sel}))

	mustRefuse(t, e.call("tenant_network_provision", map[string]any{"select": sel}), "invalid_arguments", "confirmCode is required")
	mustRefuse(t, e.call("tenant_network_provision", map[string]any{"select": sel, "confirmCode": code, "description": strings.Repeat("d", 129)}),
		"invalid_arguments", "at most 128 allowed")

	// A selected change whose validation is not finished refuses with no POST.
	e.portal.set(func(p *provPortal) {
		p.states = `[{"id":3,"networkEntityType":"CONNECTOR","networkEntitySubType":"IP_SEC","networkEntityId":"1303","state":"IN_PROGRESS"}]`
	})
	mustRefuse(t, e.call("tenant_network_provision", map[string]any{"select": sel, "confirmCode": code}), "validation_pending",
		"CONNECTOR:1303 is IN_PROGRESS (validation state id CONNECTOR:IP_SEC:1303)")
	r := e.call("tenant_network_provision_preview", map[string]any{"select": sel})
	mustOK(t, r)
	if !strings.Contains(r.Text, "No confirmCode: selected changes are not validated: CONNECTOR:1303 is IN_PROGRESS") || strings.Contains(r.Text, "\nconfirmCode: ") {
		t.Errorf("preview with a running validation:\n%s", r.Text)
	}
	e.portal.set(func(p *provPortal) { p.states = "[]" })
	if n := len(e.portal.posts()); n != 0 {
		t.Fatalf("%d POSTs", n)
	}

	for _, tc := range []struct {
		status     int
		body, code string
		text       string
	}{
		{400, `{"code":"AK000001","message":"cannot include CONNECTOR - c1 in provisioning as the dependency SEGMENT - s1 is not included"}`, "dependency_not_selected", "the dependency SEGMENT - s1 is not included. Select the changes it names as well"},
		{400, `{"code":"AK000001","message":"cannot include SEGMENT - s1 in de-provisioning as it is referenced by [CONNECTOR - c1]"}`, "dependency_not_selected", "referenced by [CONNECTOR - c1]. Select the changes it names as well"},
		{400, `{"code":"AK000001","message":"De-provision 'CONNECTOR' with name 'c1' first, before provisioning a new one with same name"}`, "dependency_not_selected", "deprovision the object it names first"},
		{400, `{"code":"AK000001","message":"cannot include CONNECTOR - c1 in provisioning as its validation status is FAILED"}`, "validation_pending", "its validation status is FAILED. Check it with tenant_network_validation_state_get."},
		{403, `{"code":"AK000004","message":"Provisioning is disabled"}`, "permission_denied", "Provisioning is not enabled for this tenant"},
		{500, `{"message":"Request failed"}`, "outcome_unknown", "may have been filed"},
	} {
		e.portal.set(func(p *provPortal) { p.provision = fault{status: tc.status, body: tc.body} })
		code := previewCode(t, e.call("tenant_network_provision_preview", map[string]any{"select": sel}))
		mustRefuse(t, e.call("tenant_network_provision", map[string]any{"select": sel, "confirmCode": code}), tc.code, tc.text)
	}
	if n := len(e.portal.posts()); n != 6 {
		t.Errorf("%d POSTs, want one per backend answer", n)
	}

	// A provision without a request id header.
	e.portal.set(func(p *provPortal) { p.provision = fault{status: 202, body: `{}`} })
	code = previewCode(t, e.call("tenant_network_provision_preview", map[string]any{"select": sel}))
	r = e.call("tenant_network_provision", map[string]any{"select": sel, "confirmCode": code})
	mustOK(t, r)
	if !strings.Contains(r.Text, "The provision was accepted (status 202) but no provision request id was returned") || strings.Contains(r.Text, "request id pr-") {
		t.Errorf("no request id: %s", r.Text)
	}

	// A running provision refuses the preview and the provision, as the portal disables Provision.
	e.portal.set(func(p *provPortal) {
		p.provision = fault{status: 202, header: map[string]string{"x-provision-request-id": "pr-2"}}
	})
	code = previewCode(t, e.call("tenant_network_provision_preview", map[string]any{"select": sel}))
	e.portal.set(func(p *provPortal) { p.lastState = "IN_PROGRESS" })
	posts := len(e.portal.posts())
	mustRefuse(t, e.call("tenant_network_provision", map[string]any{"select": sel, "confirmCode": code}), "provision_running", "A provision is running on this network")
	mustRefuse(t, e.call("tenant_network_provision_preview", map[string]any{"select": sel}), "provision_running", "nothing was previewed")
	if len(e.portal.posts()) != posts {
		t.Error("POST while a provision runs")
	}

	// A never-provisioned network proceeds.
	e.portal.set(func(p *provPortal) { p.lastState = "" })
	r = e.call("tenant_network_provision_preview", map[string]any{"select": sel})
	previewCode(t, r)
	if !strings.Contains(r.Text, "Network last provision state: never provisioned.") {
		t.Errorf("never provisioned: %s", r.Text)
	}

	// The kill switch refuses before any change list is read.
	e.portal.set(func(p *provPortal) { p.flags = []string{ak.WriteDisabledFlag} })
	e.srv = e.server(e.portal.client("tenant-killed"))
	code = previewCode(t, e.call("tenant_network_provision_preview", map[string]any{"select": sel}))
	posts = len(e.portal.posts())
	mustRefuse(t, e.call("tenant_network_provision", map[string]any{"select": sel, "confirmCode": code}), "writes_disabled", "disabled for this tenant")
	if len(e.portal.posts()) != posts {
		t.Error("POST with the kill switch on")
	}
}

// MA-W57: the code is bound to tenant, select and expiry.
func TestProvision_CodeBinding(t *testing.T) {
	e := newProvEnv(t)
	e.fixture()
	setA := []any{"CONNECTOR:1303", "SEGMENT:12"}
	codeA := previewCode(t, e.call("tenant_network_provision_preview", map[string]any{"select": setA}))

	tamper := func(code string, f func(b []byte)) string {
		b, err := base64.RawURLEncoding.DecodeString(code)
		if err != nil {
			t.Fatal(err)
		}
		f(b)
		return base64.RawURLEncoding.EncodeToString(b)
	}
	later := tamper(codeA, func(b []byte) { binary.BigEndian.PutUint64(b[:8], binary.BigEndian.Uint64(b[:8])+3600) })
	flipped := tamper(codeA, func(b []byte) { b[len(b)-1] ^= 1 })

	mustRefuse(t, e.call("tenant_network_provision", map[string]any{"select": []any{"CONNECTOR:1303"}, "confirmCode": codeA}), "confirm_mismatch", "does not match")
	mustRefuse(t, e.call("tenant_network_provision", map[string]any{"select": "all", "confirmCode": codeA}), "confirm_mismatch", "does not match")
	mustRefuse(t, e.call("tenant_network_provision", map[string]any{"select": setA, "confirmCode": flipped}), "confirm_mismatch", "does not match")
	mustRefuse(t, e.call("tenant_network_provision", map[string]any{"select": setA, "confirmCode": later}), "confirm_mismatch", "does not match")
	mustRefuse(t, e.call("tenant_network_provision", map[string]any{"select": setA, "confirmCode": "not-a-code"}), "confirm_mismatch", "does not match")

	other := e.server(e.portal.client("tenant-other"))
	mustRefuse(t, callWrite(t, other, "tenant_network_provision", map[string]any{"select": setA, "confirmCode": codeA}), "confirm_mismatch", "does not match")

	e.now = e.now.Add(confirm.Lifetime + time.Second)
	mustRefuse(t, e.call("tenant_network_provision", map[string]any{"select": setA, "confirmCode": codeA}), "confirm_expired", "expired")
	e.now = e.now.Add(-confirm.Lifetime)
	if n := len(e.portal.posts()); n != 0 {
		t.Fatalf("%d POSTs after refusals", n)
	}

	// Item order and case do not change the selection.
	mustOK(t, e.call("tenant_network_provision", map[string]any{"select": []any{"segment:12", "CONNECTOR:1303"}, "confirmCode": codeA}))
	if n := len(e.portal.posts()); n != 1 {
		t.Errorf("%d POSTs, want 1", n)
	}

	// A delete code never passes as a provision code.
	sig, err := confirm.NewSigner(e.key, confirm.WithClock(func() time.Time { return e.now }))
	if err != nil {
		t.Fatal(err)
	}
	del, _ := sig.Issue("tenant-a#7", "segment", "12", [32]byte{})
	mustRefuse(t, e.call("tenant_network_provision", map[string]any{"select": setA, "confirmCode": del}), "confirm_mismatch", "does not match")
}

// MA-W58: a changed pending list voids the code.
func TestProvision_PendingListChanged(t *testing.T) {
	sel := []any{"CONNECTOR:1303"}
	for name, change := range map[string]func(p *provPortal){
		"added": func(p *provPortal) {
			p.pol = "[" + chPrefix + `,{"entity":{"id":"10","name":"p2","type":"PREFIX_LIST","properties":{}},"changeType":"ADDED"}]`
		},
		"removed": func(p *provPortal) { p.pol = "[]" },
		"edited": func(p *provPortal) {
			p.tps = strings.Replace(p.tps, `"lastConfigUpdatedAt":100`, `"lastConfigUpdatedAt":200`, 1)
		},
		"selected change gone": func(p *provPortal) { p.tps = "[" + chSegment + "]" },
	} {
		t.Run(name, func(t *testing.T) {
			e := newProvEnv(t)
			e.fixture()
			code := previewCode(t, e.call("tenant_network_provision_preview", map[string]any{"select": sel}))
			e.portal.set(change)
			mustRefuse(t, e.call("tenant_network_provision", map[string]any{"select": sel, "confirmCode": code}), "confirm_stale",
				"pending changes changed since the preview; run the preview again")
			if n := len(e.portal.posts()); n != 0 {
				t.Errorf("%d POSTs", n)
			}
		})
	}
}

// MA-W60: a FAILED network refuses the preview (no code) and the provision (no POST).
func TestProvision_FailedNetwork(t *testing.T) {
	e := newProvEnv(t)
	e.fixture()
	e.portal.set(func(p *provPortal) {
		p.lastState = "FAILED"
		p.failed["pr-f1"] = `{"id":"pr-f1","state":"FAILED","errorDetails":{"code":"AK000009","message":"RESERVATION_FAILED","metadata":{"contactSupport":true}}}`
	})
	r := e.call("tenant_network_provision_preview", map[string]any{"select": "all"})
	mustRefuse(t, r, "network_failed", `Request pr-f1: "RESERVATION_FAILED" (code AK000009) [contact support`)
	if strings.Contains(r.Text, "confirmCode") {
		t.Errorf("code on a FAILED network: %s", r.Text)
	}
	var lists []string
	for _, j := range e.portal.entries() {
		if j.Path == "/api/tenantnetworks/7/provision-requests" && strings.Contains(j.Query, "state=FAILED") {
			lists = append(lists, j.Query)
		}
	}
	if len(lists) != 1 || lists[0] != "limit=3&offset=0&paginated=true&state=FAILED" {
		t.Errorf("failed requests: want one page of 3, got %v", lists)
	}

	e.portal.set(func(p *provPortal) { p.lastState = "SUCCESS" })
	code := previewCode(t, e.call("tenant_network_provision_preview", map[string]any{"select": "all"}))
	e.portal.set(func(p *provPortal) { p.lastState = "FAILED" })
	mustRefuse(t, e.call("tenant_network_provision", map[string]any{"select": "all", "confirmCode": code}), "network_failed", "nothing was sent")
	if n := len(e.portal.posts()); n != 0 {
		t.Errorf("%d POSTs", n)
	}
}

// Q-A15: selected FAILED objects are warned about.
func TestProvision_FailedEntityWarning(t *testing.T) {
	e := newProvEnv(t)
	e.portal.set(func(p *provPortal) {
		p.tps = `[{"entity":{"id":"5","name":"c5","type":"CONNECTOR","provisionedState":"FAILED","properties":{"type":"AWS_VPC"}},"changeType":"MODIFIED","lastUpdatedByUserId":"u-me"}]`
	})
	r := e.call("tenant_network_provision_preview", map[string]any{"select": "all"})
	code := previewCode(t, r)
	if !strings.Contains(r.Text, "Warning: selected changes whose objects are FAILED: CONNECTOR:5.") || !strings.Contains(r.Text, `"subType":"AWS VPC"`) {
		t.Errorf("preview: %s", r.Text)
	}
	r = e.call("tenant_network_provision", map[string]any{"select": "all", "confirmCode": code})
	mustOK(t, r)
	if !strings.Contains(r.Text, "Warning: selected changes whose objects were FAILED: CONNECTOR:5.") {
		t.Errorf("provision: %s", r.Text)
	}
}

// Dry-run validation: flag gate (HS14 text), body without cleanup options, conflicts.
func TestProvision_Validate(t *testing.T) {
	e := newProvEnv(t)
	e.fixture()
	sel := []any{"CONNECTOR:1303", "SEGMENT:12"}
	mustRefuse(t, e.call("tenant_network_validate", map[string]any{"select": sel}), "not_enabled",
		"Dry-run validation is not enabled for this tenant (tenant feature RELEASE-DRY-RUN-NETWORK-VALIDATION is off); nothing was sent.")
	if n := len(e.portal.posts()); n != 0 {
		t.Fatalf("%d POSTs", n)
	}

	e.portal.set(func(p *provPortal) {
		p.flags = []string{"RELEASE-DRY-RUN-NETWORK-VALIDATION", "RELEASE-PROV-BEST-EFFORT-CLEANUP"}
	})
	e.srv = e.server(e.portal.client("tenant-v"))
	r := e.call("tenant_network_validate", map[string]any{"select": sel})
	mustOK(t, r)
	if !strings.Contains(r.Text, "validation request id vr-1") {
		t.Errorf("%s", r.Text)
	}
	posts := e.portal.posts()
	if len(posts) != 1 || posts[0].Path != "/api/tenantnetworks/7/validation-requests" || posts[0].Query != "" ||
		posts[0].Body != `[{"changeType":"ADDED","entity":{"id":"1303","type":"CONNECTOR","subType":"IPsec"}},{"changeType":"ADDED","entity":{"id":"12","type":"SEGMENT"}}]` {
		t.Errorf("validation POST: %+v", posts)
	}
	e.portal.set(func(p *provPortal) {
		p.validation = fault{status: 409, body: `{"code":"AK000001","message":"already running"}`}
	})
	mustRefuse(t, e.call("tenant_network_validate", map[string]any{"select": sel}), "conflict", "A dry-run validation is already running for this tenant.")
	e.portal.set(func(p *provPortal) { p.validation = fault{status: 503, body: `{"message":"no worker"}`} })
	mustRefuse(t, e.call("tenant_network_validate", map[string]any{"select": sel}), "unavailable", "unavailable now")

	e.portal.set(func(p *provPortal) {
		p.valReqs["vr-1"] = `{"validationRequestId":"vr-1","state":"FAILED","changes":[{"changeType":"ADDED","entityType":"CONNECTOR","entityId":"1303","entityName":"c1"}],` +
			`"errors":[{"message":"cannot include CONNECTOR - c1 in provisioning as the dependency SEGMENT - s1 is not included"}]}`
	})
	r = e.call("tenant_network_validation_request_get", map[string]any{"id": "vr-1"})
	mustOK(t, r)
	for _, want := range []string{"state FAILED", "dry run FAILED", `CONNECTOR:1303 ADDED "c1"`, "dependency SEGMENT - s1 is not included"} {
		if !strings.Contains(r.Text, want) {
			t.Errorf("validation request lacks %q: %s", want, r.Text)
		}
	}
	mustRefuse(t, e.call("tenant_network_validation_request_get", map[string]any{"id": "missing"}), "not_found", "not found")
}

// MA-W18: provision request verdicts (Q-A6).
func TestProvision_RequestVerdicts(t *testing.T) {
	stuck := time.Now().Add(-31 * time.Minute).UnixMilli()
	for _, tc := range []struct {
		name, request, buffered, network string
		want                             []string
	}{
		{"success", `{"id":"r","state":"SUCCESS","bufferedProvisionRequestId":1}`, `{"id":1,"state":"SUCCESS"}`, "SUCCESS", []string{"Verdict: SUCCESS"}},
		{"failed", `{"id":"r","state":"FAILED","bufferedProvisionRequestId":1,"errorDetails":{"message":"boom"}}`, `{"id":1,"state":"FAILED"}`, "FAILED", []string{"Verdict: FAILURE", `Error details: "boom"`}},
		{"request partial", `{"id":"r","state":"PARTIAL_SUCCESS","bufferedProvisionRequestId":1}`, `{"id":1,"state":"PARTIAL_SUCCESS"}`, "PARTIAL_SUCCESS", []string{"Verdict: FAILURE"}},
		{"failed, buffered partial", `{"id":"r","state":"FAILED","bufferedProvisionRequestId":1}`, `{"id":1,"state":"PARTIAL_SUCCESS"}`, "SUCCESS", []string{"Verdict: FAILURE", "Buffered request 1: state PARTIAL_SUCCESS."}},
		{"pending, buffered partial", `{"id":"r","state":"PENDING","bufferedProvisionRequestId":1}`, `{"id":1,"state":"PARTIAL_SUCCESS"}`, "SUCCESS", []string{"Verdict: FAILURE"}},
		{"network partial", `{"id":"r","state":"SUCCESS","bufferedProvisionRequestId":1}`, `{"id":1,"state":"SUCCESS"}`, "PARTIAL_SUCCESS", []string{"Verdict: FAILURE: the request finished, but the network's last provision state is PARTIAL_SUCCESS."}},
		{"in progress", `{"id":"r","state":"IN_PROGRESS","bufferedProvisionRequestId":1}`, `{"id":1,"state":"IN_PROGRESS"}`, "IN_PROGRESS", []string{"Verdict: PENDING"}},
		{"initializing", `{"id":"r","state":"INITIALIZING"}`, ``, "SUCCESS", []string{"Verdict: PENDING", "Buffered request: none yet."}},
		{"state-less", `{"id":"r"}`, ``, "SUCCESS", []string{"Provision request r: state none.", "Verdict: PENDING"}},
		{"settling", `{"id":"r","state":"SUCCESS","bufferedProvisionRequestId":1}`, `{"id":1,"state":"SUCCESS"}`, "IN_PROGRESS", []string{"Verdict: SETTLING"}},
		{"contact support", `{"id":"r","state":"FAILED","errorDetails":{"message":"hard","metadata":{"contactSupport":true}}}`, ``, "FAILED", []string{"Verdict: FAILURE", "Contact support"}},
		{"parked", `{"id":"r","state":"PENDING"}`, ``, "FAILED", []string{"Verdict: PARKED"}},
		{"stuck", fmt.Sprintf(`{"id":"r","state":"IN_PROGRESS","createdAt":%d}`, stuck), ``, "IN_PROGRESS", []string{"Verdict: PENDING", "may be stuck; this is not a failure"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := newProvEnv(t)
			e.portal.set(func(p *provPortal) {
				p.requests["r"] = tc.request
				if tc.buffered != "" {
					p.buffered["1"] = tc.buffered
				}
				p.lastState = tc.network
			})
			r := e.call("tenant_network_provision_request_get", map[string]any{"id": "r"})
			mustOK(t, r)
			if tc.network != "" && !strings.Contains(r.Text, "Network last provision state: "+tc.network+".") {
				t.Errorf("network state missing: %s", r.Text)
			}
			for _, w := range tc.want {
				if !strings.Contains(r.Text, w) {
					t.Errorf("lacks %q:\n%s", w, r.Text)
				}
			}
			for _, j := range e.portal.entries() {
				if j.Method != "GET" {
					t.Errorf("%s %s", j.Method, j.Path)
				}
			}
		})
	}
	e := newProvEnv(t)
	mustRefuse(t, e.call("tenant_network_provision_request_get", map[string]any{"id": "nope"}), "not_found", "not found")
	mustRefuse(t, e.call("tenant_network_provision_request_get", map[string]any{"id": "../x"}), "invalid_arguments", "id must be")
}

// MA-W44: the validation state read tool; final states read like a save.
func TestProvision_ValidationStateGet(t *testing.T) {
	e := newProvEnv(t)
	e.portal.set(func(p *provPortal) {
		p.valStates["CONNECTOR:IP_SEC:1303"] = `{"id":3,"networkEntityType":"CONNECTOR","networkEntitySubType":"IP_SEC","networkEntityId":"1303","state":"FAILED",` +
			`"errorDetails":{"code":"AK000001","message":"invalid tunnel","errors":[{"field":"sites[0].tunnels","code":"constraint.invalid","message":"overlap"}]}}`
		p.valStates["CONNECTOR:IP_SEC:1"] = `{"id":4,"networkEntityType":"CONNECTOR","networkEntitySubType":"IP_SEC","networkEntityId":"1","state":"SUCCESS"}`
		p.valStates["CONNECTOR:IP_SEC:2"] = `{"id":5,"networkEntityType":"CONNECTOR","networkEntitySubType":"IP_SEC","networkEntityId":"2","state":"PENDING"}`
		p.valStates["CONNECTOR:IP_SEC:3"] = `{"id":6,"networkEntityType":"CONNECTOR","networkEntitySubType":"IP_SEC","networkEntityId":"3","state":"IN_PROGRESS"}`
	})
	for id, want := range map[string]string{
		"CONNECTOR:IP_SEC:1303": `Saved connector (IP_SEC) id 1303 (validation state id CONNECTOR:IP_SEC:1303), but validation FAILED: "invalid tunnel"; sites[0].tunnels: overlap (code AK000001); fix it with its _update tool or remove it with its _delete tool. It blocks provisioning until it is fixed or deleted.`,
		"CONNECTOR:IP_SEC:1":    "Validation SUCCESS: saved connector (IP_SEC) id 1 is valid",
		"CONNECTOR:IP_SEC:2":    "is still running (state PENDING",
		"CONNECTOR:IP_SEC:3":    "is still running (state IN_PROGRESS",
	} {
		r := e.call("tenant_network_validation_state_get", map[string]any{"id": id})
		mustOK(t, r)
		if !strings.Contains(r.Text, want) {
			t.Errorf("%s: %s", id, r.Text)
		}
	}
	mustRefuse(t, e.call("tenant_network_validation_state_get", map[string]any{"id": "CONNECTOR:IP_SEC:9"}), "not_found", "not found")
}

// tenant_network_changes_get reads only the flags, both change lists and the
// caller: failing provision request and validation state reads do not reach it.
func TestProvision_ChangesGetReadsOnlyChanges(t *testing.T) {
	e := newProvEnv(t)
	e.fixture()
	broken := fault{status: 500, body: `{"message":"Request failed"}`}
	e.portal.set(func(p *provPortal) {
		p.fail = map[string]fault{"/tenantnetworks/7/provision-requests": broken, "/tenantnetworks/7/validation-states": broken}
	})
	r := e.call("tenant_network_changes_get", map[string]any{})
	mustOK(t, r)
	if !strings.Contains(r.Text, "Pending changes: ") {
		t.Fatalf("text = %s", r.Text)
	}
	for _, j := range e.portal.entries() {
		if strings.HasSuffix(j.Path, "/provision-requests") || strings.HasSuffix(j.Path, "/validation-states") {
			t.Errorf("changes_get read %s", j.Path)
		}
	}
}

// A pending-change read that fails is the same typed refusal in all four
// tools that read the pending changes; no POST goes out.
func TestProvision_PlanFailure(t *testing.T) {
	e := newProvEnv(t)
	e.fixture()
	e.portal.set(func(p *provPortal) { p.flags = []string{"RELEASE-DRY-RUN-NETWORK-VALIDATION"} })
	e.srv = e.server(e.portal.client("tenant-plan"))
	sel := []any{"CONNECTOR:1303", "SEGMENT:12"}
	code := previewCode(t, e.call("tenant_network_provision_preview", map[string]any{"select": sel}))
	e.portal.set(func(p *provPortal) {
		p.fail = map[string]fault{"/tenantnetworks/7/networkpolicy/changes": {status: 500, body: `{"message":"Request failed"}`}}
	})
	for tool, c := range map[string]struct {
		args    map[string]any
		nothing string
	}{
		"tenant_network_changes_get":       {map[string]any{}, "listed"},
		"tenant_network_provision_preview": {map[string]any{"select": sel}, "previewed"},
		"tenant_network_provision":         {map[string]any{"select": sel, "confirmCode": code}, "sent"},
		"tenant_network_validate":          {map[string]any{"select": sel}, "sent"},
	} {
		mustRefuse(t, e.call(tool, c.args), "backend_error", "The pending changes could not be read, so nothing was "+c.nothing+": policy changes:")
	}
	if n := len(e.portal.posts()); n != 0 {
		t.Fatalf("%d POSTs", n)
	}
}

// MA-W6, MA-W7: annotations and registration of the provisioning tools.
func TestProvision_Registration(t *testing.T) {
	read := map[string]bool{"tenant_network_changes_get": true, "tenant_network_provision_preview": true, "tenant_network_provision_request_get": true,
		"tenant_network_validation_request_get": true, "tenant_network_validation_state_get": true}
	want := map[string][3]bool{"tenant_network_provision": {false, true, false}, "tenant_network_validate": {false, false, false}}
	for n := range read {
		want[n] = [3]bool{true, false, true}
	}
	srv := server.NewMCPServer("mcp-alkira", "test")
	if err := Register(srv, noClient); err != nil {
		t.Fatal(err)
	}
	readNames := toolNames(t, srv)
	if err := RegisterWrite(srv, noClient); err != nil {
		t.Fatal(err)
	}
	var res struct {
		Tools []struct {
			Name        string `json:"name"`
			Annotations struct {
				ReadOnlyHint, DestructiveHint, IdempotentHint, OpenWorldHint *bool
			} `json:"annotations"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(call(t, srv, "tools/list", nil), &res); err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, tool := range res.Tools {
		w, ok := want[tool.Name]
		if !ok {
			continue
		}
		found++
		isRead := false
		for _, n := range readNames {
			isRead = isRead || n == tool.Name
		}
		if isRead != read[tool.Name] {
			t.Errorf("%s: registered through Register = %v", tool.Name, isRead)
		}
		a := tool.Annotations
		if a.ReadOnlyHint == nil || a.DestructiveHint == nil || a.IdempotentHint == nil || a.OpenWorldHint == nil {
			t.Fatalf("%s: nil hint", tool.Name)
		}
		if *a.ReadOnlyHint != w[0] || *a.DestructiveHint != w[1] || *a.IdempotentHint != w[2] || !*a.OpenWorldHint {
			t.Errorf("%s: hints %v %v %v %v", tool.Name, *a.ReadOnlyHint, *a.DestructiveHint, *a.IdempotentHint, *a.OpenWorldHint)
		}
	}
	if found != len(want) {
		t.Errorf("%d of %d provisioning tools served", found, len(want))
	}
}

// Register and RegisterWrite given the same WithConfirmKey share the code
// key: a preview served by Register on one replica issues a code that
// tenant_network_provision served by RegisterWrite on another accepts. A
// different key refuses it, and a short key registers nothing.
func TestProvision_SharedKey(t *testing.T) {
	portal := newProvPortal(t)
	portal.set(func(p *provPortal) { p.tps = "[" + chSegment + "]" })
	get := providerFor(portal.client(""))
	key := confirm.RandomKey()
	replica := func(k []byte) *server.MCPServer {
		srv := server.NewMCPServer("mcp-alkira", "test")
		if err := Register(srv, get, handlers.WithConfirmKey(k)); err != nil {
			t.Fatal(err)
		}
		if err := RegisterWrite(srv, get, handlers.WithConfirmKey(k)); err != nil {
			t.Fatal(err)
		}
		return srv
	}
	a, b, other := replica(key), replica(key), replica(confirm.RandomKey())
	code := previewCode(t, callWrite(t, a, "tenant_network_provision_preview", map[string]any{"select": "all"}))
	mustRefuse(t, callWrite(t, other, "tenant_network_provision", map[string]any{"select": "all", "confirmCode": code}), "confirm_mismatch", "does not match")
	if n := len(portal.posts()); n != 0 {
		t.Fatalf("%d POSTs with another key", n)
	}
	mustOK(t, callWrite(t, b, "tenant_network_provision", map[string]any{"select": "all", "confirmCode": code}))
	if n := len(portal.posts()); n != 1 {
		t.Errorf("%d POSTs", n)
	}

	srv := server.NewMCPServer("mcp-alkira", "test")
	if err := Register(srv, get, handlers.WithConfirmKey([]byte("short"))); err == nil {
		t.Error("short key accepted")
	}
	// No tool is registered, so the server offers no tools capability.
	var res struct {
		Error *struct{ Code int } `json:"error"`
	}
	body, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "tools/list"})
	raw, _ := json.Marshal(srv.HandleMessage(context.Background(), body))
	if json.Unmarshal(raw, &res); res.Error == nil {
		t.Errorf("tools registered after the key error: %s", raw)
	}
}

// Register without WithConfirmKey serves the preview, which refuses with
// confirm_unavailable and sends no request: a code signed with a key no
// other replica holds would be refused by tenant_network_provision.
func TestProvision_PreviewWithoutKey(t *testing.T) {
	portal := newProvPortal(t)
	portal.set(func(p *provPortal) { p.tps = "[" + chSegment + "]" })
	srv := server.NewMCPServer("mcp-alkira", "test")
	if err := Register(srv, providerFor(portal.client(""))); err != nil {
		t.Fatal(err)
	}
	mustRefuse(t, callWrite(t, srv, "tenant_network_provision_preview", map[string]any{"select": "all"}), "confirm_unavailable", "no confirmation key is configured")
	if n := len(portal.entries()); n != 0 {
		t.Errorf("%d requests without a key", n)
	}
}

// A code files one provision: the pending list is unchanged until a run
// succeeds, but the request the first call filed voids the code.
func TestProvision_CodeReplay(t *testing.T) {
	e := newProvEnv(t)
	e.fixture()
	sel := []any{"CONNECTOR:1303"}
	code := previewCode(t, e.call("tenant_network_provision_preview", map[string]any{"select": sel}))
	mustOK(t, e.call("tenant_network_provision", map[string]any{"select": sel, "confirmCode": code}))
	mustRefuse(t, e.call("tenant_network_provision", map[string]any{"select": sel, "confirmCode": code}), "confirm_stale", "pending changes changed since the preview")
	if n := len(e.portal.posts()); n != 1 {
		t.Errorf("%d POSTs, want 1", n)
	}
	// Another actor's provision filed after the preview voids it too.
	code = previewCode(t, e.call("tenant_network_provision_preview", map[string]any{"select": sel}))
	e.portal.set(func(p *provPortal) { p.filed = append([]string{"pr-other"}, p.filed...) })
	mustRefuse(t, e.call("tenant_network_provision", map[string]any{"select": sel, "confirmCode": code}), "confirm_stale", "pending changes changed since the preview")
	if n := len(e.portal.posts()); n != 1 {
		t.Errorf("%d POSTs, want 1", n)
	}
}

// The write tools name tenant_network_validation_state_get after a
// validation that has not finished, since this change serves it.
func TestProvision_ValidationStatePointer(t *testing.T) {
	e := newWriteEnv(t)
	e.portal.flags = []string{asyncFlag}
	e.setPipeline(handlers.WithWriteTimeout(5 * time.Second)) // too short for a 10 s poll
	r := e.call("widget_create", map[string]any{"body": map[string]any{"name": "w1"}})
	mustOK(t, r)
	if !strings.Contains(r.Text, "Saved; validation is still running (state id WIDGET:DEFAULT:101). Check it with tenant_network_validation_state_get.") {
		t.Fatalf("text = %s", r.Text)
	}
}
