package handlers

import (
	"context"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/mark3labs/mcp-go/mcp"
)

func TenantResourceUsages(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		resourceCategory, err := request.RequireString("category")
		if err != nil {
			resourceCategory = ""
		}

		resourceType, err := request.RequireString("type")
		if err != nil {
			resourceType = ""
		}

		resourceScope, err := request.RequireString("scope")
		if err != nil {
			resourceScope = ""
		}

		// Get resources
		data, err := client.GetResourceUsages(
			resourceCategory, resourceType, resourceScope)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}

func TenantResourceLimits(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// Get resources
		data, err := client.GetResourceLimits()

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}
