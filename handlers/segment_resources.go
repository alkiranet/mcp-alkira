package handlers

import (
	"context"
	"fmt"

	"github.com/alkiranet/alkira-client-go/alkira"
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAllSegmentResources(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewSegmentResource(client)

		// Get resources
		segmentResources, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve segment resources")
		}

		// Return response
		return mcp.NewToolResultText(segmentResources), nil
	}
}

func GetAllSegmentResourceShares(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewSegmentResourceShare(client)

		// Get resources
		segmentResourceShares, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve segment resource shares")
		}

		// Return response
		return mcp.NewToolResultText(segmentResourceShares), nil
	}
}