package handlers

import (
	"context"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/mark3labs/mcp-go/mcp"
)

func CxpGetAll(get ClientProvider) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, errResult := resolve(ctx, get)
		if errResult != nil {
			return errResult, nil
		}

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
