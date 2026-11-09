package handlers

import (
	"context"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/mark3labs/mcp-go/mcp"
)

func PolicyTrafficGetAll(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := ak.NewTrafficPolicy(client)

		// Get resources
		policies, err := api.GetAll("", "")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(policies), nil
	}
}

func PolicyTrafficGetById(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		policyId, err := request.RequireString("policyId")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// INIT
		api := ak.NewTrafficPolicy(client)

		// Get resources
		data, err := api.GetById(policyId)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}

func PolicyTrafficGetByName(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		policyName, err := request.RequireString("policyName")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// INIT
		api := ak.NewTrafficPolicy(client)

		// Get resources
		data, err := api.GetByName(policyName)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}
