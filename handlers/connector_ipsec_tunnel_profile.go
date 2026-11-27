package handlers

import (
	"context"
	"encoding/json"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/mark3labs/mcp-go/mcp"
)

func ConnectorIPSecTunnelProfileGetAll(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		offset, limit, err := pagingArgs(request, "0", "50")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// INIT
		api := ak.NewConnectorIPSecTunnelProfile(client)

		// Get resources
		connectors, err := api.GetAll(offset, limit)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(connectors), nil
	}
}

func ConnectorIPSecTunnelProfileGetById(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		profileId, err := request.RequireString("profileId")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		_, raw, err := client.GetIPsecTunnelProfile(ctx, profileId)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(string(raw)), nil
	}
}

// ConnectorIPSecTunnelProfileGetByName pages the profile collection and
// returns every profile with the name; no ?name= query is sent.
func ConnectorIPSecTunnelProfileGetByName(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		profileName, err := request.RequireString("profileName")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		profiles, err := client.GetIPsecTunnelProfileByName(ctx, profileName)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, err := json.Marshal(profiles)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}
