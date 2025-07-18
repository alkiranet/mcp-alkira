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

func GetAllLists(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT
		dnsServerList := alkira.NewDnsServerList(client)
		globalCidrList := alkira.NewGlobalCidrList(client)
		prefixList := alkira.NewPolicyPrefixList(client)

		// Get resources
		dnsServerLists, err := dnsServerList.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve DNS Server List")
		}

		globalCidrLists, err := globalCidrList.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Global CIDR List")
		}

		prefixLists, err := prefixList.GetAll()

		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve Prefix List")
		}

		lists := dnsServerLists + globalCidrLists + prefixLists

		// Return response
		return mcp.NewToolResultText(lists), nil
	}
}


func GetAllConnectors(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		// INIT - Cloud Provider Connectors
		awsVpcConnector := alkira.NewConnectorAwsVpc(client)
		awsTgwConnector := alkira.NewConnectorAwsTgw(client)
		awsDirectConnectConnector := alkira.NewConnectorAwsDirectConnect(client)
		azureVnetConnector := alkira.NewConnectorAzureVnet(client)
		azureExpressRouteConnector := alkira.NewConnectorAzureExpressRoute(client)
		gcpVpcConnector := alkira.NewConnectorGcpVpc(client)
		gcpInterconnectConnector := alkira.NewConnectorGcpInterconnect(client)
		ociVcnConnector := alkira.NewConnectorOciVcn(client)

		// INIT - SD-WAN Connectors
		ciscoSdwanConnector := alkira.NewConnectorCiscoSdwan(client)
		versaSdwanConnector := alkira.NewConnectorVersaSdwan(client)
		fortinetSdwanConnector := alkira.NewConnectorFortinetSdwan(client)
		vmwareSdwanConnector := alkira.NewConnectorVmwareSdwan(client)
		arubaEdgeConnector := alkira.NewConnectorArubaEdge(client)

		// INIT - VPN/Security Connectors
		ipsecConnector := alkira.NewConnectorIPSec(client)
		advIpsecConnector := alkira.NewConnectorAdvIPSec(client)
		ipsecTunnelProfileConnector := alkira.NewConnectorIPSecTunnelProfile(client)

		// INIT - Internet/DDoS Protection
		inetConnector := alkira.NewConnectorInternet(client)
		//akamaiProlexicConnector := alkira.NewConnectorAkamaiProlexic(client)

		// INIT - Remote Access
		remoteAccessTemplateConnector := alkira.NewConnectorRemoteAccessTemplate(client)

		// Get resources - Cloud Provider Connectors
		awsVpcConnectors, err := awsVpcConnector.GetAll()
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve connector-aws-vpc")
		}

		awsTgwConnectors, err := awsTgwConnector.GetAll()
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve connector-aws-tgw")
		}

		awsDirectConnectConnectors, err := awsDirectConnectConnector.GetAll()
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve connector-aws-direct-connect")
		}

		azureVnetConnectors, err := azureVnetConnector.GetAll()
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve connector-azure-vnet")
		}

		azureExpressRouteConnectors, err := azureExpressRouteConnector.GetAll()
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve connector-azure-express-route")
		}

		gcpVpcConnectors, err := gcpVpcConnector.GetAll()
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve connector-gcp-vpc")
		}

		gcpInterconnectConnectors, err := gcpInterconnectConnector.GetAll()
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve connector-gcp-interconnect")
		}

		ociVcnConnectors, err := ociVcnConnector.GetAll()
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve connector-oci-vcn")
		}

		// Get resources - SD-WAN Connectors
		ciscoSdwanConnectors, err := ciscoSdwanConnector.GetAll()
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve connector-cisco-sdwan")
		}

		versaSdwanConnectors, err := versaSdwanConnector.GetAll()
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve connector-versa-sdwan")
		}

		fortinetSdwanConnectors, err := fortinetSdwanConnector.GetAll()
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve connector-fortinet-sdwan")
		}

		vmwareSdwanConnectors, err := vmwareSdwanConnector.GetAll()
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve connector-vmware-sdwan")
		}

		arubaEdgeConnectors, err := arubaEdgeConnector.GetAll()
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve connector-aruba-edge")
		}

		// Get resources - VPN/Security Connectors
		ipsecConnectors, err := ipsecConnector.GetAll()
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve connector-ipsec")
		}

		advIpsecConnectors, err := advIpsecConnector.GetAll()
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve connector-adv-ipsec")
		}

		ipsecTunnelProfileConnectors, err := ipsecTunnelProfileConnector.GetAll()
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve connector-ipsec-tunnel-profile")
		}

		// Get resources - Internet/DDoS Protection
		inetConnectors, err := inetConnector.GetAll()
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve connector-internet")
		}

		// akamaiProlexicConnectors, err := akamaiProlexicConnector.GetAll()
		// if err != nil {
		// 	return nil, fmt.Errorf("Failed to retrieve connector-akamai-prolexic")
		// }

		// Get resources - Remote Access
		remoteAccessTemplateConnectors, err := remoteAccessTemplateConnector.GetAll()
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve connector-remote-access-template")
		}

		// Combine all connectors
		connectors := awsVpcConnectors + awsTgwConnectors + awsDirectConnectConnectors +
			azureVnetConnectors + azureExpressRouteConnectors +
			gcpVpcConnectors + gcpInterconnectConnectors + ociVcnConnectors +
			ciscoSdwanConnectors + versaSdwanConnectors + fortinetSdwanConnectors +
			vmwareSdwanConnectors + arubaEdgeConnectors +
			ipsecConnectors + advIpsecConnectors + ipsecTunnelProfileConnectors +
			inetConnectors + //akamaiProlexicConnectors +
			remoteAccessTemplateConnectors

		// Return response
		return mcp.NewToolResultText(connectors), nil
	}
}
