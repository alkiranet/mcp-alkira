package handlers

import (
	"context"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/mark3labs/mcp-go/mcp"
)

func ConnectorInternetGetAll(get ClientProvider) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, errResult := resolve(ctx, get)
		if errResult != nil {
			return errResult, nil
		}

		offset, _ := request.RequireString("offset")
		limit, _ := request.RequireString("limit")

		// INIT
		api := ak.NewConnectorInternet(client)

		// Get resources
		connectors, err := api.GetAll(offset, limit)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(connectors), nil
	}
}

func ConnectorInternetGetById(get ClientProvider) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, errResult := resolve(ctx, get)
		if errResult != nil {
			return errResult, nil
		}

		connectorId, err := request.RequireString("connectorId")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// INIT
		api := ak.NewConnectorInternet(client)

		// Get resources
		data, err := api.GetById(connectorId)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}

func ConnectorInternetGetByName(get ClientProvider) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, errResult := resolve(ctx, get)
		if errResult != nil {
			return errResult, nil
		}

		connectorName, err := request.RequireString("connectorName")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// INIT
		api := ak.NewConnectorInternet(client)

		// Get resources
		data, err := api.GetByName(connectorName)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}

func ConnectorInternetGetTotal(get ClientProvider) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, errResult := resolve(ctx, get)
		if errResult != nil {
			return errResult, nil
		}

		// INIT
		api := ak.NewConnectorInternet(client)

		// Get resources
		data, err := api.GetCount()

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}
