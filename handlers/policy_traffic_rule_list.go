package handlers

import (
	"context"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/mark3labs/mcp-go/mcp"
)

func PolicyTrafficRuleListGetAll(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := ak.NewTrafficPolicyRuleList(client)

		// Get resources
		lists, err := api.GetAll("", "")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(lists), nil
	}
}

func PolicyTrafficRuleListGetById(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		ruleListId, err := request.RequireString("ruleListId")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// INIT
		api := ak.NewTrafficPolicyRuleList(client)

		// Get resources
		data, err := api.GetById(ruleListId)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}

func PolicyTrafficRuleListGetByName(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		ruleListName, err := request.RequireString("ruleListName")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// INIT
		api := ak.NewTrafficPolicyRuleList(client)

		// Get resources
		data, err := api.GetByName(ruleListName)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}
