package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	ak "github.com/alkiranet/client-go/tenant"
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

func TestNewServer_ServesRegistry(t *testing.T) {
	srv := newServer(func(context.Context) (*ak.AlkiraClient, error) {
		return nil, errors.New("no client in this test")
	})
	if got := len(list(t, srv, "tools/list", "tools")); got != 191 {
		t.Fatalf("tools/list returned %d tools, want 191", got)
	}
	if got := list(t, srv, "prompts/list", "prompts"); len(got) != 1 {
		t.Fatalf("prompts/list returned %v, want one prompt", got)
	}
}
