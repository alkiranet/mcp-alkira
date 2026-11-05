package handlers

import (
	"context"
	"testing"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestResolve_NilClientIsToolError(t *testing.T) {
	_, res := resolve(context.Background(), func(context.Context) (*ak.AlkiraClient, error) {
		return nil, nil
	})
	if res == nil || !res.IsError {
		t.Fatalf("resolve with a nil client returned %+v, want a tool error", res)
	}
}

func TestConnectorIPSecTunnelProfileGetAll_MissingArgs(t *testing.T) {
	get := func(context.Context) (*ak.AlkiraClient, error) { return &ak.AlkiraClient{}, nil }
	for _, args := range []map[string]any{{"limit": "10"}, {"offset": "0"}} {
		req := mcp.CallToolRequest{}
		req.Params.Arguments = args
		res, err := ConnectorIPSecTunnelProfileGetAll(get)(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		if !res.IsError {
			t.Fatalf("args %v: want a tool error, got %+v", args, res)
		}
	}
}
