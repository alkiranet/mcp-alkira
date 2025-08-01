package handlers

import (
	"context"
	"fmt"

	"github.com/alkiranet/alkira-client-go/alkira"
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAllInternetApplication(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewInternetApplication(client)

		// Get resources
		applications, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Internet Applications")
		}

		// Return response
		return mcp.NewToolResultText(applications), nil
	}
}
