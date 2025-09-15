package handlers

import (
	"context"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAlerts(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		s, _ := request.RequireString("status")
		t, _ := request.RequireString("type")
		p, _ := request.RequireString("priority")

		offset, _ := request.RequireString("offset")
		limit, _ := request.RequireString("limit")

		// Get resources
		alerts, err := client.GetAlertsSummary(s, t, p, offset, limit)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(alerts), nil
	}
}

func GetAuditLogs(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		s, _ := request.RequireString("status")
		t, _ := request.RequireString("type")

		offset, _ := request.RequireString("offset")
		limit, _ := request.RequireString("limit")

		// Get resources
		auditLogs, err := client.GetAuditLogSummary(s, t, offset, limit)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(auditLogs), nil
	}
}

func GetJobs(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		s, err := request.RequireString("status")
		if err != nil {
			s = ""
		}

		t, err := request.RequireString("type")
		if err != nil {
			t = ""
		}

		offset, _ := request.RequireString("offset")
		limit, _ := request.RequireString("limit")

		// Get resources
		jobs, err := client.GetJobs(s, t, offset, limit)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(jobs), nil
	}
}

func GetResourceUsages(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		resourceCategory, err := request.RequireString("category")
		if err != nil {
			resourceCategory = ""
		}

		resourceType, err := request.RequireString("type")
		if err != nil {
			resourceType = ""
		}

		resourceScope, err := request.RequireString("scope")
		if err != nil {
			resourceScope = ""
		}

		// Get resources
		data, err := client.GetResourceUsages(
			resourceCategory, resourceType, resourceScope)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}

func GetResourceLimits(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// Get resources
		data, err := client.GetResourceLimits()

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(data), nil
	}
}
