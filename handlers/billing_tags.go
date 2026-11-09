package handlers

import (
	"context"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/mark3labs/mcp-go/mcp"
)

func BillingTagGetAll(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		offset, _ := request.RequireString("offset")
		limit, _ := request.RequireString("limit")

		// INIT
		api := ak.NewBillingTag(client)

		// Get resources
		billingTags, err := api.GetAll(offset, limit)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(billingTags), nil
	}
}

func BillingTagGetById(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		tagId, err := request.RequireString("tagId")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// INIT
		api := ak.NewBillingTag(client)

		// Get resources
		data, err := api.GetById(tagId)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}

func BillingTagGetByName(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		tagName, err := request.RequireString("tagName")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// INIT
		api := ak.NewBillingTag(client)

		// Get resources
		data, err := api.GetByName(tagName)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}

func BillingTagGetTotal(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := ak.NewBillingTag(client)

		// Get resources
		data, err := api.GetCount()

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}
