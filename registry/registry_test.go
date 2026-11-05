package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/mark3labs/mcp-go/server"
)

// wantToolCount is the number of tools Register adds.
const wantToolCount = 191

var update = flag.Bool("update", false, "rewrite testdata/tools_list.golden.json from the current registry")

var goldenPath = filepath.Join("testdata", "tools_list.golden.json")

func newServer(get func(context.Context) (*ak.AlkiraClient, error)) *server.MCPServer {
	srv := server.NewMCPServer("mcp-alkira", "test")
	Register(srv, get)
	return srv
}

func noClient(context.Context) (*ak.AlkiraClient, error) {
	return nil, errors.New("no client in this test")
}

// call sends one JSON-RPC request through srv and returns the raw "result" member.
func call(t *testing.T, srv *server.MCPServer, method string, params any) json.RawMessage {
	t.Helper()
	req := map[string]any{"jsonrpc": "2.0", "id": 1, "method": method}
	if params != nil {
		req["params"] = params
	}
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(srv.HandleMessage(context.Background(), body))
	if err != nil {
		t.Fatal(err)
	}
	var env struct {
		Result json.RawMessage `json:"result"`
		Error  json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatal(err)
	}
	if env.Error != nil {
		t.Fatalf("%s returned a JSON-RPC error: %s", method, env.Error)
	}
	return env.Result
}

// toolsListJSON returns the tools/list result as indented JSON with a trailing newline.
func toolsListJSON(t *testing.T, srv *server.MCPServer) []byte {
	t.Helper()
	var out bytes.Buffer
	if err := json.Indent(&out, call(t, srv, "tools/list", nil), "", "  "); err != nil {
		t.Fatal(err)
	}
	out.WriteByte('\n')
	return out.Bytes()
}

func toolNames(t *testing.T, srv *server.MCPServer) []string {
	t.Helper()
	var res struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(call(t, srv, "tools/list", nil), &res); err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(res.Tools))
	for _, tool := range res.Tools {
		names = append(names, tool.Name)
	}
	return names
}

type toolResult struct {
	IsError bool `json:"isError"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

func callTool(t *testing.T, srv *server.MCPServer, name string, args map[string]any) toolResult {
	t.Helper()
	var res toolResult
	raw := call(t, srv, "tools/call", map[string]any{"name": name, "arguments": args})
	if err := json.Unmarshal(raw, &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Content) != 1 || res.Content[0].Type != "text" {
		t.Fatalf("%s: want one text content item, got %s", name, raw)
	}
	return res
}

func TestRegister_ToolsListGolden(t *testing.T) {
	got := toolsListJSON(t, newServer(noClient))
	if *update {
		if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("tools/list differs from %s; run with -update to regenerate after an intended change", goldenPath)
	}
}

func TestRegister_Count191(t *testing.T) {
	if got := len(toolNames(t, newServer(noClient))); got != wantToolCount {
		t.Fatalf("registered tools = %d, want %d", got, wantToolCount)
	}
}

func TestRegister_AllToolsUseProvider(t *testing.T) {
	var calls atomic.Int64
	srv := newServer(func(ctx context.Context) (*ak.AlkiraClient, error) {
		calls.Add(1)
		return nil, errors.New("provider failed")
	})
	names := toolNames(t, srv)

	t.Run("tools", func(t *testing.T) {
		for _, name := range names {
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				res := callTool(t, srv, name, map[string]any{})
				if !res.IsError || res.Content[0].Text != "provider failed" {
					t.Fatalf("want isError with provider error, got isError=%v text=%q", res.IsError, res.Content[0].Text)
				}
			})
		}
	})

	if got := calls.Load(); got != int64(len(names)) {
		t.Fatalf("provider calls = %d, want %d", got, len(names))
	}
}

type requestLog struct {
	mu    sync.Mutex
	paths []string
}

func (l *requestLog) add(path string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.paths = append(l.paths, path)
}

func (l *requestLog) get() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.paths...)
}

// fakePortal serves canned JSON for the given paths and records each request path.
func fakePortal(t *testing.T, routes map[string]string) (*ak.AlkiraClient, *requestLog) {
	t.Helper()
	seen := &requestLog{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.add(r.URL.Path)
		if got := r.Header.Get("Authorization"); got != "api-key test" {
			t.Errorf("Authorization = %q", got)
		}
		body, ok := routes[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}))
	t.Cleanup(ts.Close)

	httpClient := retryablehttp.NewClient()
	httpClient.RetryMax = 0
	httpClient.Logger = nil
	return &ak.AlkiraClient{
		Client:          httpClient,
		URI:             ts.URL + "/api",
		TenantNetworkId: "7",
		Authorization:   "api-key test",
	}, seen
}

func providerFor(client *ak.AlkiraClient) func(context.Context) (*ak.AlkiraClient, error) {
	return func(context.Context) (*ak.AlkiraClient, error) { return client, nil }
}

func TestRegister_SegmentGetAll_HTTP(t *testing.T) {
	client, seen := fakePortal(t, map[string]string{
		"/api/tenantnetworks/7/segments": `[{"id":1,"name":"seg-a","asn":65000},{"id":2,"name":"seg-b","asn":65001}]`,
	})
	res := callTool(t, newServer(providerFor(client)), "segment_get_all", map[string]any{})
	if res.IsError {
		t.Fatalf("unexpected tool error: %s", res.Content[0].Text)
	}
	if text := res.Content[0].Text; !strings.Contains(text, "seg-a") || !strings.Contains(text, "seg-b") {
		t.Fatalf("summary missing segments: %s", text)
	}
	if paths := seen.get(); len(paths) != 1 || paths[0] != "/api/tenantnetworks/7/segments" {
		t.Fatalf("requests = %v", paths)
	}
}

func TestRegister_SegmentGetById_HTTP(t *testing.T) {
	client, seen := fakePortal(t, map[string]string{
		"/api/tenantnetworks/7/segments/42": `{"id":42,"name":"seg-42"}`,
	})
	res := callTool(t, newServer(providerFor(client)), "segment_get_by_id", map[string]any{"segmentId": "42"})
	if res.IsError {
		t.Fatalf("unexpected tool error: %s", res.Content[0].Text)
	}
	if !strings.Contains(res.Content[0].Text, "seg-42") {
		t.Fatalf("unexpected body: %s", res.Content[0].Text)
	}
	if paths := seen.get(); len(paths) != 1 || paths[0] != "/api/tenantnetworks/7/segments/42" {
		t.Fatalf("requests = %v", paths)
	}
}

func TestRegister_TenantNetworkSummary_HTTP(t *testing.T) {
	client, seen := fakePortal(t, map[string]string{
		"/api/tenantnetworksummaries": `[{"id":7,"name":"tn-7"}]`,
	})
	res := callTool(t, newServer(providerFor(client)), "tenant_network_summary", map[string]any{})
	if res.IsError {
		t.Fatalf("unexpected tool error: %s", res.Content[0].Text)
	}
	if !strings.Contains(res.Content[0].Text, "tn-7") {
		t.Fatalf("unexpected body: %s", res.Content[0].Text)
	}
	if paths := seen.get(); len(paths) != 1 || paths[0] != "/api/tenantnetworksummaries" {
		t.Fatalf("requests = %v", paths)
	}
}

func TestDesktopMain_UsesRegistry(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(src, []byte("AddTool(")) {
		t.Fatal("main.go registers tools directly; tools belong in registry.Register")
	}
	if !bytes.Contains(src, []byte("registry.Register(srv,")) {
		t.Fatal("main.go does not call registry.Register")
	}
}
