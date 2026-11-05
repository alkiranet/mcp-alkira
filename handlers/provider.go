package handlers

import (
	"context"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/mark3labs/mcp-go/mcp"
)

// ClientProvider returns the Alkira client a tool call runs with.
type ClientProvider func(ctx context.Context) (*ak.AlkiraClient, error)

// resolve returns the client from get, or a tool error result when get fails.
func resolve(ctx context.Context, get ClientProvider) (*ak.AlkiraClient, *mcp.CallToolResult) {
	client, err := get(ctx)
	if err != nil {
		return nil, mcp.NewToolResultError(err.Error())
	}
	return client, nil
}
