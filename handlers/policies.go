package handlers

import (
	"context"
	"fmt"

	"github.com/alkiranet/alkira-client-go/alkira"
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAllNatPolicy(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewNatPolicy(client)

		// Get resources
		policies, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve NAT policies")
		}

		// Return response
		return mcp.NewToolResultText(policies), nil
	}
}

func GetAllNatRule(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewNatRule(client)

		// Get resources
		rules, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve NAT policy rules")
		}

		// Return response
		return mcp.NewToolResultText(rules), nil
	}
}

func GetAllRoutePolicy(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewRoutePolicy(client)

		// Get resources
		policies, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve route policies")
		}

		// Return response
		return mcp.NewToolResultText(policies), nil
	}
}

func GetAllTrafficPolicy(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewTrafficPolicy(client)

		// Get resources
		policies, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve traffic policies")
		}

		// Return response
		return mcp.NewToolResultText(policies), nil
	}
}

func GetAllTrafficPolicyRule(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewTrafficPolicyRule(client)

		// Get resources
		rules, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve traffic policy rules")
		}

		// Return response
		return mcp.NewToolResultText(rules), nil
	}
}

func GetAllPolicyRuleList(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewPolicyRuleList(client)

		// Get resources
		ruleLists, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve policy rule lists")
		}

		// Return response
		return mcp.NewToolResultText(ruleLists), nil
	}
}

func GetAllPolicyPrefixList(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewPolicyPrefixList(client)

		// Get resources
		prefixLists, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve policy prefix lists")
		}

		// Return response
		return mcp.NewToolResultText(prefixLists), nil
	}
}

func GetAllPolicyFqdnList(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewPolicyFqdnList(client)

		// Get resources
		fqdnLists, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve policy FQDN lists")
		}

		// Return response
		return mcp.NewToolResultText(fqdnLists), nil
	}
}