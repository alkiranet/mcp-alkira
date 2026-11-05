package handlers

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

func HealthConnectorGetById(get ClientProvider) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, errResult := resolve(ctx, get)
		if errResult != nil {
			return errResult, nil
		}

		id, err := request.RequireString("connectorId")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get resources
		health, err := client.GetHealthOfConnector(id)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(health), nil
	}
}

func HealthConnectorInstanceGetById(get ClientProvider) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, errResult := resolve(ctx, get)
		if errResult != nil {
			return errResult, nil
		}

		id, err := request.RequireString("connectorId")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		instanceId, err := request.RequireString("instanceId")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get resources
		health, err := client.GetHealthOfConnectorInstance(id, instanceId)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(health), nil
	}
}

func HealthServiceGetById(get ClientProvider) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, errResult := resolve(ctx, get)
		if errResult != nil {
			return errResult, nil
		}

		id, err := request.RequireString("serviceId")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get resources
		health, err := client.GetHealthOfService(id)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(health), nil
	}
}

func HealthServiceInstanceGetById(get ClientProvider) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, errResult := resolve(ctx, get)
		if errResult != nil {
			return errResult, nil
		}

		id, err := request.RequireString("serviceId")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		instanceId, err := request.RequireString("instanceId")

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get resources
		health, err := client.GetHealthOfServiceInstance(id, instanceId)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(health), nil
	}
}
