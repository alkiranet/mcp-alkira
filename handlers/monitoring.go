package handlers

import (
	"context"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAlerts(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		s, err := request.RequireString("status")
        if err != nil {
			s = ""
        }

		t, err := request.RequireString("type")
        if err != nil {
			t = ""
        }

		p, err := request.RequireString("priority")
        if err != nil {
			p = ""
        }

		// Get resources
		alerts, err := client.GetAlerts(s, t, p)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(alerts), nil
	}
}

func GetAuditLogs(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		s, err := request.RequireString("status")
        if err != nil {
			s = ""
        }

		t, err := request.RequireString("type")
        if err != nil {
			t = ""
        }

		// Get resources
		auditLogs, err := client.GetAuditLogs(s, t)

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

		// Get resources
		jobs, err := client.GetJobs(s, t)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(jobs), nil
	}
}

func GetAllHealth(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// Get resources
		health, err := client.GetHealthAll()

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(health), nil
	}
}

func GetHealthOfConnector(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		id, err := request.RequireString("id")

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

func GetHealthOfConnectorInstance(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		id, err := request.RequireString("id")

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

func GetHealthOfService(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		id, err := request.RequireString("id")

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

func GetHealthOfServiceInstance(client *ak.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		id, err := request.RequireString("id")

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
