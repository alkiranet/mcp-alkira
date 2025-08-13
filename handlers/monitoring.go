package handlers

import (
	"context"
	"strconv"

	"github.com/alkiranet/alkira-client-go/alkira"
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAlerts(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

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

func GetAuditLogs(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

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

func GetJobs(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

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

func GetAllHealth(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

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


func GetHealthOfConnectorInstance(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		id, err := request.RequireString("id")

        if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
        }

		instanceId, err := request.RequireString("instanceId")

        if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
        }

		// Convert string IDs to integers as required by API
		connectorIdInt, err := strconv.Atoi(id)
		if err != nil {
			return mcp.NewToolResultError("Invalid connector ID: " + err.Error()), nil
		}

		instanceIdInt, err := strconv.Atoi(instanceId)
		if err != nil {
			return mcp.NewToolResultError("Invalid instance ID: " + err.Error()), nil
		}

		// Get resources using the renamed method
		health, err := client.GetHealthConnectorInstance(connectorIdInt, instanceIdInt)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(health), nil
	}
}

func GetHealthOfService(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		id, err := request.RequireString("id")

        if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
        }

		// Convert string ID to integer as required by API
		serviceIdInt, err := strconv.Atoi(id)
		if err != nil {
			return mcp.NewToolResultError("Invalid service ID: " + err.Error()), nil
		}

		// Get resources
		health, err := client.GetHealthOfService(serviceIdInt)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(health), nil
	}
}

func GetHealthOfServiceInstance(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		id, err := request.RequireString("id")

        if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
        }

		instanceId, err := request.RequireString("instanceId")

        if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
        }

		// Convert string IDs to integers as required by API
		serviceIdInt, err := strconv.Atoi(id)
		if err != nil {
			return mcp.NewToolResultError("Invalid service ID: " + err.Error()), nil
		}

		instanceIdInt, err := strconv.Atoi(instanceId)
		if err != nil {
			return mcp.NewToolResultError("Invalid instance ID: " + err.Error()), nil
		}

		// Get resources
		health, err := client.GetHealthOfServiceInstance(serviceIdInt, instanceIdInt)

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(health), nil
	}
}
