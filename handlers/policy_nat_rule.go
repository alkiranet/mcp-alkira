package handlers

import (
	"context"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/mark3labs/mcp-go/mcp"
)

func PolicyNatRuleGetAll(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		offset, err := request.RequireString("offset")
		limit, err := request.RequireString("limit")

		// INIT
		api := ak.NewNATRule(client)

		// Get resources
		rules, err := api.GetSummary(offset, limit)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(rules), nil
	}
}

func PolicyNatRuleGetById(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		ruleId, err := request.RequireString("ruleId")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// INIT
		api := ak.NewNATRule(client)

		// Get resources
		data, err := api.GetById(ruleId)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}

func PolicyNatRuleGetByName(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		ruleName, err := request.RequireString("ruleName")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// INIT
		api := ak.NewNATRule(client)

		// Get resources
		data, err := api.GetByName(ruleName)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}

func PolicyNatRuleGetTotal(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := ak.NewNATRule(client)

		// Get resources
		data, err := api.GetCount()

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}
