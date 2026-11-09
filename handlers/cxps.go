package handlers

import (
	"context"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/mark3labs/mcp-go/mcp"
)

func CxpGetAll(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := ak.NewCXP(client)

		// Get resources
		cxps, err := api.GetSummary("", "")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(cxps), nil
	}
}
