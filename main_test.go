package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/registry"
	"github.com/mark3labs/mcp-go/server"
)

// list sends a JSON-RPC list request through srv and returns the names of the items under key.
func list(t *testing.T, srv *server.MCPServer, method, key string) []string {
	t.Helper()
	body, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": method})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(srv.HandleMessage(context.Background(), body))
	if err != nil {
		t.Fatal(err)
	}
	var env struct {
		Result map[string][]struct {
			Name string `json:"name"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, item := range env.Result[key] {
		names = append(names, item.Name)
	}
	return names
}

func noClient(context.Context) (*ak.AlkiraClient, error) {
	return nil, errors.New("no client in this test")
}

// readToolCount is the number of tools registry.Register serves.
func readToolCount(t *testing.T) int {
	srv := server.NewMCPServer("mcp-alkira", "test")
	registry.Register(srv, noClient)
	return len(list(t, srv, "tools/list", "tools"))
}

func TestNewServer_ServesRegistry(t *testing.T) {
	srv := newServer(noClient, false)
	// registry.TestRegister_ToolNames pins the names of the read tools.
	if got, want := len(list(t, srv, "tools/list", "tools")), readToolCount(t); got != want {
		t.Fatalf("tools/list returned %d tools, want the %d read tools", got, want)
	}
	if got := list(t, srv, "prompts/list", "prompts"); len(got) != 1 {
		t.Fatalf("prompts/list returned %v, want one prompt", got)
	}
}

// writeTools returns the tools srv lists without readOnlyHint: every read
// tool sets it, every write tool clears it.
func writeTools(t *testing.T, srv *server.MCPServer) []string {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "tools/list"})
	raw, err := json.Marshal(srv.HandleMessage(context.Background(), body))
	if err != nil {
		t.Fatal(err)
	}
	var env struct {
		Result struct {
			Tools []struct {
				Name        string `json:"name"`
				Annotations struct {
					ReadOnlyHint *bool `json:"readOnlyHint"`
				} `json:"annotations"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, tool := range env.Result.Tools {
		if ro := tool.Annotations.ReadOnlyHint; ro == nil || !*ro {
			names = append(names, tool.Name)
		}
	}
	return names
}

// Without -allowWrite only read tools are served; with it, the write tools as well.
func TestNewServer_AllowWrite(t *testing.T) {
	srv := newServer(noClient, true)
	names, writes := list(t, srv, "tools/list", "tools"), writeTools(t, srv)
	if len(names) != readToolCount(t)+len(writes) {
		t.Fatalf("%d tools, %d of them write tools", len(names), len(writes))
	}
	if w := writeTools(t, newServer(noClient, false)); w != nil {
		t.Fatalf("write tools %v served without -allowWrite", w)
	}
}

func TestCheckWriteMode(t *testing.T) {
	if err := checkWriteMode(true, "sse"); err == nil {
		t.Fatal("-allowWrite accepted with -mode sse")
	}
	for _, c := range []struct {
		allow bool
		mode  string
	}{{true, "stdio"}, {false, "sse"}, {false, "stdio"}} {
		if err := checkWriteMode(c.allow, c.mode); err != nil {
			t.Fatalf("%v %s: %v", c.allow, c.mode, err)
		}
	}
}
