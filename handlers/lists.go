package handlers

import (
	"context"
	"fmt"

	"github.com/alkiranet/alkira-client-go/alkira"
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAllListAsPath(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewListAsPath(client)

		// Get resources
		lists, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve AS Path lists")
		}

		// Return response
		return mcp.NewToolResultText(lists), nil
	}
}

func GetAllListCommunity(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewListCommunity(client)

		// Get resources
		lists, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Community lists")
		}

		// Return response
		return mcp.NewToolResultText(lists), nil
	}
}

func GetAllListExtendedCommunity(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewListExtendedCommunity(client)

		// Get resources
		lists, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Extended Community lists")
		}

		// Return response
		return mcp.NewToolResultText(lists), nil
	}
}

func GetAllDnsServerList(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewDnsServerList(client)

		// Get resources
		lists, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve DNS Server lists")
		}

		// Return response
		return mcp.NewToolResultText(lists), nil
	}
}

func GetAllGlobalCidrList(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewGlobalCidrList(client)

		// Get resources
		lists, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Global CIDR lists")
		}

		// Return response
		return mcp.NewToolResultText(lists), nil
	}
}

func GetAllUdrList(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewUdrList(client)

		// Get resources
		lists, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve UDR lists")
		}

		// Return response
		return mcp.NewToolResultText(lists), nil
	}
}

func GetAllPolicyPrefixListIndividual(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewPolicyPrefixList(client)

		// Get resources
		lists, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Policy Prefix lists")
		}

		// Return response
		return mcp.NewToolResultText(lists), nil
	}
}

func GetAllPolicyFqdnListIndividual(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewPolicyFqdnList(client)

		// Get resources
		lists, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Policy FQDN lists")
		}

		// Return response
		return mcp.NewToolResultText(lists), nil
	}
}