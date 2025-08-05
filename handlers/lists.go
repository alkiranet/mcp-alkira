package handlers

import (
	"context"

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
			return mcp.NewToolResultError(err.Error()), nil
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
			return mcp.NewToolResultError(err.Error()), nil
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
			return mcp.NewToolResultError(err.Error()), nil
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
			return mcp.NewToolResultError(err.Error()), nil
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
			return mcp.NewToolResultError(err.Error()), nil
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
			return mcp.NewToolResultError(err.Error()), nil
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
			return mcp.NewToolResultError(err.Error()), nil
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
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response
		return mcp.NewToolResultText(lists), nil
	}
}
