package handlers

import (
	"context"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/mark3labs/mcp-go/mcp"
)

func HealthConnectorGetById(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

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

func HealthConnectorInstanceGetById(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

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

func HealthServiceGetById(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

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

func HealthServiceInstanceGetById(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

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
