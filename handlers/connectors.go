package handlers

import (
	"context"
	"fmt"

	"github.com/alkiranet/alkira-client-go/alkira"
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAllConnectorArubaEdge(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewConnectorArubaEdge(client)

		// Get resources
		connectors, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Aruba Edge Connect connectors")
		}

		// Return response
		return mcp.NewToolResultText(connectors), nil
	}
}

func GetAllConnectorAwsDirectConnect(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewConnectorAwsDirectConnect(client)

		// Get resources
		connectors, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve AWS Direct Connect connectors")
		}

		// Return response
		return mcp.NewToolResultText(connectors), nil
	}
}

func GetAllConnectorAwsTgw(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewConnectorAwsTgw(client)

		// Get resources
		connectors, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve AWS Transit Gateway connectors")
		}

		// Return response
		return mcp.NewToolResultText(connectors), nil
	}
}

func GetAllConnectorAwsVpc(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewConnectorAwsVpc(client)

		// Get resources
		connectors, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve AWS VPC connectors")
		}

		// Return response
		return mcp.NewToolResultText(connectors), nil
	}
}

func GetAllConnectorAzureExpressRoute(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewConnectorAzureExpressRoute(client)

		// Get resources
		connectors, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Azure ExpressRoute connectors")
		}

		// Return response
		return mcp.NewToolResultText(connectors), nil
	}
}

func GetAllConnectorAzureVnet(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewConnectorAzureVnet(client)

		// Get resources
		connectors, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Azure Virtual Network connectors")
		}

		// Return response
		return mcp.NewToolResultText(connectors), nil
	}
}

func GetAllConnectorCiscoSdwan(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewConnectorCiscoSdwan(client)

		// Get resources
		connectors, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Cisco SD-WAN connectors")
		}

		// Return response
		return mcp.NewToolResultText(connectors), nil
	}
}

func GetAllConnectorFortinetSdwan(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewConnectorFortinetSdwan(client)

		// Get resources
		connectors, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Fortinet SD-WAN connectors")
		}

		// Return response
		return mcp.NewToolResultText(connectors), nil
	}
}

func GetAllConnectorGcpInterconnect(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewConnectorGcpInterconnect(client)

		// Get resources
		connectors, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve GCP Interconnect connectors")
		}

		// Return response
		return mcp.NewToolResultText(connectors), nil
	}
}

func GetAllConnectorGcpVpc(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewConnectorGcpVpc(client)

		// Get resources
		connectors, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve GCP VPC connectors")
		}

		// Return response
		return mcp.NewToolResultText(connectors), nil
	}
}

func GetAllConnectorInternet(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewConnectorInternet(client)

		// Get resources
		connectors, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Internet Exit connectors")
		}

		// Return response
		return mcp.NewToolResultText(connectors), nil
	}
}

func GetAllConnectorIPSec(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewConnectorIPSec(client)

		// Get resources
		connectors, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve IPSec connectors")
		}

		// Return response
		return mcp.NewToolResultText(connectors), nil
	}
}

func GetAllConnectorAdvIPSec(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewConnectorAdvIPSec(client)

		// Get resources
		connectors, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Advanced IPSec connectors")
		}

		// Return response
		return mcp.NewToolResultText(connectors), nil
	}
}

func GetAllConnectorIPSecTunnelProfile(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewConnectorIPSecTunnelProfile(client)

		// Get resources
		connectors, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve IPSec Tunnel Profile connectors")
		}

		// Return response
		return mcp.NewToolResultText(connectors), nil
	}
}

func GetAllConnectorOciVcn(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewConnectorOciVcn(client)

		// Get resources
		connectors, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Oracle Cloud VCN connectors")
		}

		// Return response
		return mcp.NewToolResultText(connectors), nil
	}
}

func GetAllConnectorRemoteAccessTemplate(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewConnectorRemoteAccessTemplate(client)

		// Get resources
		connectors, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Remote Access Template connectors")
		}

		// Return response
		return mcp.NewToolResultText(connectors), nil
	}
}

func GetAllConnectorVersaSdwan(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewConnectorVersaSdwan(client)

		// Get resources
		connectors, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Versa SD-WAN connectors")
		}

		// Return response
		return mcp.NewToolResultText(connectors), nil
	}
}

func GetAllConnectorVmwareSdwan(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		api := alkira.NewConnectorVmwareSdwan(client)

		// Get resources
		connectors, err := api.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve VMware SD-WAN connectors")
		}

		// Return response
		return mcp.NewToolResultText(connectors), nil
	}
}
