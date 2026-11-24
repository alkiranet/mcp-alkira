package handlers

import (
	"context"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/mark3labs/mcp-go/mcp"
)

func ZtaProfileGetAll(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		offset, limit, err := pagingArgs(request, "0", "50")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, err := ak.NewZtaProfile(client).GetAll(offset, limit)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(data), nil
	}
}

func ZtaProfileGetById(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		profileId, err := request.RequireString("profileId")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, err := ak.NewZtaProfile(client).GetById(profileId)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(data), nil
	}
}
