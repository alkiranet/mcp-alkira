package handlers

import (
	"context"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/mark3labs/mcp-go/mcp"
)

// ClientProvider returns the Alkira client a tool call runs with.
type ClientProvider func(ctx context.Context) (*ak.AlkiraClient, error)

// ToolHandler handles one tool call.
type ToolHandler = func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

// PerCall resolves the client from get on every tool call and runs the handler mk builds for it.
func PerCall(get ClientProvider, mk func(*ak.AlkiraClient) ToolHandler) ToolHandler {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := get(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if client == nil {
			return mcp.NewToolResultError("no Alkira client available for this request"), nil
		}
		return mk(client)(ctx, request)
	}
}
