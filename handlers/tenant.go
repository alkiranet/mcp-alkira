package handlers

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

func TenantResourceUsages(get ClientProvider) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, errResult := resolve(ctx, get)
		if errResult != nil {
			return errResult, nil
		}

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

func TenantResourceLimits(get ClientProvider) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, errResult := resolve(ctx, get)
		if errResult != nil {
			return errResult, nil
		}

		// Get resources
		data, err := client.GetResourceLimits()

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}
