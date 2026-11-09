package handlers

import (
	"context"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/mark3labs/mcp-go/mcp"
)

func ByoipGetAll(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		offset, _ := request.RequireString("offset")
		limit, _ := request.RequireString("limit")

		// INIT
		api := ak.NewByoip(client)

		// Get resources
		data, err := api.GetAll(offset, limit)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}

func ByoipGetById(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		byoipId, err := request.RequireString("byoipId")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// INIT
		api := ak.NewByoip(client)

		// Get resources
		data, err := api.GetById(byoipId)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}

func ByoipGetByName(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		byoipName, err := request.RequireString("byoipName")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// INIT
		api := ak.NewByoip(client)

		// Get resources
		data, err := api.GetByName(byoipName)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}

func ByoipGetTotal(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := ak.NewByoip(client)

		// Get resources
		data, err := api.GetCount()

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}
