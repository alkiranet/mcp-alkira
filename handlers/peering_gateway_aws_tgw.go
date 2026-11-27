package handlers

import (
	"context"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/mark3labs/mcp-go/mcp"
)

func PeeringGatewayAwsTgwGetAll(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		offset, limit, err := pagingArgs(request, "0", "50")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, err := ak.NewPeeringGatewayAwsTgw(client).GetAll(offset, limit)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(data), nil
	}
}

func PeeringGatewayAwsTgwGetById(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("peeringGatewayId")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, err := ak.NewPeeringGatewayAwsTgw(client).GetById(id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(data), nil
	}
}

func PeeringGatewayAwsTgwAttachmentGetAll(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		offset, limit, err := pagingArgs(request, "0", "50")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, err := ak.NewPeeringGatewayAwsTgwAttachment(client).GetAll(offset, limit)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(data), nil
	}
}

func PeeringGatewayAwsTgwAttachmentGetById(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("attachmentId")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, err := ak.NewPeeringGatewayAwsTgwAttachment(client).GetById(id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(data), nil
	}
}
