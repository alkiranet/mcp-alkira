package handlers

import (
	"context"
	"fmt"

	"github.com/alkiranet/alkira-client-go/alkira"
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAllSegments(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewSegment(client)

		// Get resources
		segments, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve segments")
		}

		// Return response
		return mcp.NewToolResultText(segments), nil
	}
}

func GetAllGroups(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewGroup(client)

		// Get resources
		groups, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve groups")
		}

		// Return response
		return mcp.NewToolResultText(groups), nil
	}
}

func GetAllBillingTags(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewBillingTag(client)

		// Get resources
		billingTags, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve billingTags")
		}

		// Return response
		return mcp.NewToolResultText(billingTags), nil
	}
}
