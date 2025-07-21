package handlers

import (
	"context"
	"fmt"

	"github.com/alkiranet/alkira-client-go/alkira"
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAllServiceCheckpoint(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewServiceCheckpoint(client)

		// Get resources
		services, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve checkpoint services")
		}

		// Return response
		return mcp.NewToolResultText(services), nil
	}
}

func GetAllServiceCiscoFTDv(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewServiceCiscoFTDv(client)

		// Get resources
		services, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Cisco FTDv services")
		}

		// Return response
		return mcp.NewToolResultText(services), nil
	}
}

func GetAllServiceF5Lb(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewServiceF5Lb(client)

		// Get resources
		services, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve F5 Load Balancer services")
		}

		// Return response
		return mcp.NewToolResultText(services), nil
	}
}

func GetAllServiceFortinet(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewServiceFortinet(client)

		// Get resources
		services, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Fortinet services")
		}

		// Return response
		return mcp.NewToolResultText(services), nil
	}
}

func GetAllServiceInfoblox(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewServiceInfoblox(client)

		// Get resources
		services, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Infoblox services")
		}

		// Return response
		return mcp.NewToolResultText(services), nil
	}
}

func GetAllServicePan(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewServicePan(client)

		// Get resources
		services, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Palo Alto Networks services")
		}

		// Return response
		return mcp.NewToolResultText(services), nil
	}
}

func GetAllServiceZscaler(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewServiceZscaler(client)

		// Get resources
		services, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Zscaler services")
		}

		// Return response
		return mcp.NewToolResultText(services), nil
	}
}
