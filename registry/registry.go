// Package registry registers the mcp-alkira tools on an MCP server.
package registry

import (
	"context"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
	"github.com/alkiranet/mcp-alkira/tools"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type handlerFunc = func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

// perCall resolves the client from get on every tool call and runs the handler mk builds for it.
func perCall(get handlers.ClientProvider, mk func(*ak.AlkiraClient) handlerFunc) handlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := get(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if client == nil {
			return mcp.NewToolResultError("no Alkira client available for this request"), nil
		}
		return mk(client)(ctx, request)
	}
}

// Register adds every mcp-alkira tool to srv; each tool call obtains its client from get.
func Register(srv *server.MCPServer, get handlers.ClientProvider) {
	// Add tenant network basic tools
	srv.AddTool(tools.TenantNetworkSummary(), perCall(get, handlers.TenantNetworkSummary))
	srv.AddTool(tools.TenantNetworkFirewallZones(), perCall(get, handlers.TenantNetworkFirewallZones))

	// Add billing tag tools
	srv.AddTool(tools.BillingTagGetAll(), perCall(get, handlers.BillingTagGetAll))
	srv.AddTool(tools.BillingTagGetById(), perCall(get, handlers.BillingTagGetById))
	srv.AddTool(tools.BillingTagGetByName(), perCall(get, handlers.BillingTagGetByName))
	srv.AddTool(tools.BillingTagGetTotal(), perCall(get, handlers.BillingTagGetTotal))

	srv.AddTool(tools.CxpGetAll(), perCall(get, handlers.CxpGetAll))

	// Add segment tools
	srv.AddTool(tools.SegmentGetAll(), perCall(get, handlers.SegmentGetAll))
	srv.AddTool(tools.SegmentGetById(), perCall(get, handlers.SegmentGetById))
	srv.AddTool(tools.SegmentGetByName(), perCall(get, handlers.SegmentGetByName))

	// Add group tools
	srv.AddTool(tools.GroupGetById(), perCall(get, handlers.GroupGetById))
	srv.AddTool(tools.GroupGetByName(), perCall(get, handlers.GroupGetByName))
	srv.AddTool(tools.GroupGetTotal(), perCall(get, handlers.GroupGetTotal))

	// Add BYOIP tools
	srv.AddTool(tools.ByoipGetAll(), perCall(get, handlers.ByoipGetAll))
	srv.AddTool(tools.ByoipGetById(), perCall(get, handlers.ByoipGetById))
	srv.AddTool(tools.ByoipGetByName(), perCall(get, handlers.ByoipGetByName))
	srv.AddTool(tools.ByoipGetTotal(), perCall(get, handlers.ByoipGetTotal))

	// Add segment resource tools
	srv.AddTool(tools.SegmentResourceGetAll(), perCall(get, handlers.SegmentResourceGetAll))
	srv.AddTool(tools.SegmentResourceGetById(), perCall(get, handlers.SegmentResourceGetById))
	srv.AddTool(tools.SegmentResourceGetByName(), perCall(get, handlers.SegmentResourceGetByName))
	srv.AddTool(tools.SegmentResourceGetTotal(), perCall(get, handlers.SegmentResourceGetTotal))

	// Add segment resource share tools
	srv.AddTool(tools.SegmentResourceShareGetAll(), perCall(get, handlers.SegmentResourceShareGetAll))
	srv.AddTool(tools.SegmentResourceShareGetById(), perCall(get, handlers.SegmentResourceShareGetById))
	srv.AddTool(tools.SegmentResourceShareGetByName(), perCall(get, handlers.SegmentResourceShareGetByName))
	srv.AddTool(tools.SegmentResourceShareGetTotal(), perCall(get, handlers.SegmentResourceShareGetTotal))

	// Add Check Point firewall service tools
	srv.AddTool(tools.ServiceCheckPointGetAll(), perCall(get, handlers.ServiceCheckPointGetAll))
	srv.AddTool(tools.ServiceCheckPointGetById(), perCall(get, handlers.ServiceCheckPointGetById))
	srv.AddTool(tools.ServiceCheckPointGetByName(), perCall(get, handlers.ServiceCheckPointGetByName))
	srv.AddTool(tools.ServiceCheckPointGetTotal(), perCall(get, handlers.ServiceCheckPointGetTotal))

	// Add Cisco FTDV service tools
	srv.AddTool(tools.ServiceCiscoFTDvGetAll(), perCall(get, handlers.ServiceCiscoFTDvGetAll))
	srv.AddTool(tools.ServiceCiscoFTDvGetById(), perCall(get, handlers.ServiceCiscoFTDvGetById))
	srv.AddTool(tools.ServiceCiscoFTDvGetByName(), perCall(get, handlers.ServiceCiscoFTDvGetByName))
	srv.AddTool(tools.ServiceCiscoFTDvGetTotal(), perCall(get, handlers.ServiceCiscoFTDvGetTotal))

	// Add F5 LB service tools
	srv.AddTool(tools.ServiceF5LBGetAll(), perCall(get, handlers.ServiceF5LBGetAll))
	srv.AddTool(tools.ServiceF5LBGetById(), perCall(get, handlers.ServiceF5LBGetById))
	srv.AddTool(tools.ServiceF5LBGetByName(), perCall(get, handlers.ServiceF5LBGetByName))
	srv.AddTool(tools.ServiceF5LBGetTotal(), perCall(get, handlers.ServiceF5LBGetTotal))

	// Add Fortinet service tools
	srv.AddTool(tools.ServiceFortinetGetAll(), perCall(get, handlers.ServiceFortinetGetAll))
	srv.AddTool(tools.ServiceFortinetGetById(), perCall(get, handlers.ServiceFortinetGetById))
	srv.AddTool(tools.ServiceFortinetGetByName(), perCall(get, handlers.ServiceFortinetGetByName))
	srv.AddTool(tools.ServiceFortinetGetTotal(), perCall(get, handlers.ServiceFortinetGetTotal))

	// Add Infoblox service tools
	srv.AddTool(tools.ServiceInfobloxGetAll(), perCall(get, handlers.ServiceInfobloxGetAll))
	srv.AddTool(tools.ServiceInfobloxGetById(), perCall(get, handlers.ServiceInfobloxGetById))
	srv.AddTool(tools.ServiceInfobloxGetByName(), perCall(get, handlers.ServiceInfobloxGetByName))
	srv.AddTool(tools.ServiceInfobloxGetTotal(), perCall(get, handlers.ServiceInfobloxGetTotal))

	// Add PAN firewall service tools
	srv.AddTool(tools.ServicePanGetAll(), perCall(get, handlers.ServicePanGetAll))
	srv.AddTool(tools.ServicePanGetById(), perCall(get, handlers.ServicePanGetById))
	srv.AddTool(tools.ServicePanGetByName(), perCall(get, handlers.ServicePanGetByName))
	srv.AddTool(tools.ServicePanGetTotal(), perCall(get, handlers.ServicePanGetTotal))

	// Add ZScaler service tools
	srv.AddTool(tools.ServiceZscalerGetAll(), perCall(get, handlers.ServiceZscalerGetAll))
	srv.AddTool(tools.ServiceZscalerGetById(), perCall(get, handlers.ServiceZscalerGetById))
	srv.AddTool(tools.ServiceZscalerGetByName(), perCall(get, handlers.ServiceZscalerGetByName))
	srv.AddTool(tools.ServiceZscalerGetTotal(), perCall(get, handlers.ServiceZscalerGetTotal))

	// Add Aruba Edge Connect SD-WAN connector tools
	srv.AddTool(tools.ConnectorArubaEdgeGetAll(), perCall(get, handlers.ConnectorArubaEdgeGetAll))
	srv.AddTool(tools.ConnectorArubaEdgeGetById(), perCall(get, handlers.ConnectorArubaEdgeGetById))
	srv.AddTool(tools.ConnectorArubaEdgeGetByName(), perCall(get, handlers.ConnectorArubaEdgeGetByName))
	srv.AddTool(tools.ConnectorArubaEdgeGetTotal(), perCall(get, handlers.ConnectorArubaEdgeGetTotal))

	// Add AWS Direct Connect connector tools
	srv.AddTool(tools.ConnectorAwsDirectConnectGetAll(), perCall(get, handlers.ConnectorAwsDirectConnectGetAll))
	srv.AddTool(tools.ConnectorAwsDirectConnectGetById(), perCall(get, handlers.ConnectorAwsDirectConnectGetById))
	srv.AddTool(tools.ConnectorAwsDirectConnectGetByName(), perCall(get, handlers.ConnectorAwsDirectConnectGetByName))
	srv.AddTool(tools.ConnectorAwsDirectConnectGetTotal(), perCall(get, handlers.ConnectorAwsDirectConnectGetTotal))

	// Add AWS Transit Gateway connector tools
	srv.AddTool(tools.ConnectorAwsTgwGetAll(), perCall(get, handlers.ConnectorAwsTgwGetAll))
	srv.AddTool(tools.ConnectorAwsTgwGetById(), perCall(get, handlers.ConnectorAwsTgwGetById))
	srv.AddTool(tools.ConnectorAwsTgwGetByName(), perCall(get, handlers.ConnectorAwsTgwGetByName))
	srv.AddTool(tools.ConnectorAwsTgwGetTotal(), perCall(get, handlers.ConnectorAwsTgwGetTotal))

	// Add Azure ExpressRoute connector tools
	srv.AddTool(tools.ConnectorAzureExpressRouteGetAll(), perCall(get, handlers.ConnectorAzureExpressRouteGetAll))
	srv.AddTool(tools.ConnectorAzureExpressRouteGetById(), perCall(get, handlers.ConnectorAzureExpressRouteGetById))
	srv.AddTool(tools.ConnectorAzureExpressRouteGetByName(), perCall(get, handlers.ConnectorAzureExpressRouteGetByName))
	srv.AddTool(tools.ConnectorAzureExpressRouteGetTotal(), perCall(get, handlers.ConnectorAzureExpressRouteGetTotal))

	// Add Cisco SD-WAN connector tools
	srv.AddTool(tools.ConnectorCiscoSdwanGetAll(), perCall(get, handlers.ConnectorCiscoSdwanGetAll))
	srv.AddTool(tools.ConnectorCiscoSdwanGetById(), perCall(get, handlers.ConnectorCiscoSdwanGetById))
	srv.AddTool(tools.ConnectorCiscoSdwanGetByName(), perCall(get, handlers.ConnectorCiscoSdwanGetByName))
	srv.AddTool(tools.ConnectorCiscoSdwanGetTotal(), perCall(get, handlers.ConnectorCiscoSdwanGetTotal))

	// Add Fortinet SD-WAN connector tools
	srv.AddTool(tools.ConnectorFortinetSdwanGetAll(), perCall(get, handlers.ConnectorFortinetSdwanGetAll))
	srv.AddTool(tools.ConnectorFortinetSdwanGetById(), perCall(get, handlers.ConnectorFortinetSdwanGetById))
	srv.AddTool(tools.ConnectorFortinetSdwanGetByName(), perCall(get, handlers.ConnectorFortinetSdwanGetByName))
	srv.AddTool(tools.ConnectorFortinetSdwanGetTotal(), perCall(get, handlers.ConnectorFortinetSdwanGetTotal))

	// Add Google Cloud Interconnect connector tools
	srv.AddTool(tools.ConnectorGcpInterconnectGetAll(), perCall(get, handlers.ConnectorGcpInterconnectGetAll))
	srv.AddTool(tools.ConnectorGcpInterconnectGetById(), perCall(get, handlers.ConnectorGcpInterconnectGetById))
	srv.AddTool(tools.ConnectorGcpInterconnectGetByName(), perCall(get, handlers.ConnectorGcpInterconnectGetByName))
	srv.AddTool(tools.ConnectorGcpInterconnectGetTotal(), perCall(get, handlers.ConnectorGcpInterconnectGetTotal))

	// Add Google Cloud VPC connector tools
	srv.AddTool(tools.ConnectorGcpVpcGetAll(), perCall(get, handlers.ConnectorGcpVpcGetAll))
	srv.AddTool(tools.ConnectorGcpVpcGetById(), perCall(get, handlers.ConnectorGcpVpcGetById))
	srv.AddTool(tools.ConnectorGcpVpcGetByName(), perCall(get, handlers.ConnectorGcpVpcGetByName))
	srv.AddTool(tools.ConnectorGcpVpcGetTotal(), perCall(get, handlers.ConnectorGcpVpcGetTotal))

	// Add Oracle Cloud Infrastructure VCN connector tools
	srv.AddTool(tools.ConnectorOciVcnGetAll(), perCall(get, handlers.ConnectorOciVcnGetAll))
	srv.AddTool(tools.ConnectorOciVcnGetById(), perCall(get, handlers.ConnectorOciVcnGetById))
	srv.AddTool(tools.ConnectorOciVcnGetByName(), perCall(get, handlers.ConnectorOciVcnGetByName))
	srv.AddTool(tools.ConnectorOciVcnGetTotal(), perCall(get, handlers.ConnectorOciVcnGetTotal))

	// Add Remote Access Template connector tools
	srv.AddTool(tools.ConnectorRemoteAccessGetAll(), perCall(get, handlers.ConnectorRemoteAccessGetAll))
	srv.AddTool(tools.ConnectorRemoteAccessGetById(), perCall(get, handlers.ConnectorRemoteAccessGetById))
	srv.AddTool(tools.ConnectorRemoteAccessGetByName(), perCall(get, handlers.ConnectorRemoteAccessGetByName))
	srv.AddTool(tools.ConnectorRemoteAccessGetTotal(), perCall(get, handlers.ConnectorRemoteAccessGetTotal))

	// Add Versa SD-WAN connector tools
	srv.AddTool(tools.ConnectorVersaSdwanGetAll(), perCall(get, handlers.ConnectorVersaSdwanGetAll))
	srv.AddTool(tools.ConnectorVersaSdwanGetById(), perCall(get, handlers.ConnectorVersaSdwanGetById))
	srv.AddTool(tools.ConnectorVersaSdwanGetByName(), perCall(get, handlers.ConnectorVersaSdwanGetByName))
	srv.AddTool(tools.ConnectorVersaSdwanGetTotal(), perCall(get, handlers.ConnectorVersaSdwanGetTotal))

	// Add VMware SD-WAN (VeloCloud) connector tools
	srv.AddTool(tools.ConnectorVmwareSdwanGetAll(), perCall(get, handlers.ConnectorVmwareSdwanGetAll))
	srv.AddTool(tools.ConnectorVmwareSdwanGetById(), perCall(get, handlers.ConnectorVmwareSdwanGetById))
	srv.AddTool(tools.ConnectorVmwareSdwanGetByName(), perCall(get, handlers.ConnectorVmwareSdwanGetByName))
	srv.AddTool(tools.ConnectorVmwareSdwanGetTotal(), perCall(get, handlers.ConnectorVmwareSdwanGetTotal))

	// Add internet connector tools
	srv.AddTool(tools.ConnectorInternetGetAll(), perCall(get, handlers.ConnectorInternetGetAll))
	srv.AddTool(tools.ConnectorInternetGetById(), perCall(get, handlers.ConnectorInternetGetById))
	srv.AddTool(tools.ConnectorInternetGetByName(), perCall(get, handlers.ConnectorInternetGetByName))
	srv.AddTool(tools.ConnectorInternetGetTotal(), perCall(get, handlers.ConnectorInternetGetTotal))

	// Add AWS VPC connector tools
	srv.AddTool(tools.ConnectorAwsVpcGetAll(), perCall(get, handlers.ConnectorAwsVpcGetAll))
	srv.AddTool(tools.ConnectorAwsVpcGetById(), perCall(get, handlers.ConnectorAwsVpcGetById))
	srv.AddTool(tools.ConnectorAwsVpcGetByName(), perCall(get, handlers.ConnectorAwsVpcGetByName))
	srv.AddTool(tools.ConnectorAwsVpcGetTotal(), perCall(get, handlers.ConnectorAwsVpcGetTotal))

	// Add Azure VNET connector tools
	srv.AddTool(tools.ConnectorAzureVnetGetAll(), perCall(get, handlers.ConnectorAzureVnetGetAll))
	srv.AddTool(tools.ConnectorAzureVnetGetById(), perCall(get, handlers.ConnectorAzureVnetGetById))
	srv.AddTool(tools.ConnectorAzureVnetGetByName(), perCall(get, handlers.ConnectorAzureVnetGetByName))
	srv.AddTool(tools.ConnectorAzureVnetGetTotal(), perCall(get, handlers.ConnectorAzureVnetGetTotal))

	// Add IPSec connector tools
	srv.AddTool(tools.ConnectorIPSecGetAll(), perCall(get, handlers.ConnectorIPSecGetAll))
	srv.AddTool(tools.ConnectorIPSecGetById(), perCall(get, handlers.ConnectorIPSecGetById))
	srv.AddTool(tools.ConnectorIPSecGetByName(), perCall(get, handlers.ConnectorIPSecGetByName))
	srv.AddTool(tools.ConnectorIPSecGetTotal(), perCall(get, handlers.ConnectorIPSecGetTotal))

	// Add IPSec Adv connector tools
	srv.AddTool(tools.ConnectorIPSecAdvGetAll(), perCall(get, handlers.ConnectorIPSecAdvGetAll))
	srv.AddTool(tools.ConnectorIPSecAdvGetById(), perCall(get, handlers.ConnectorIPSecAdvGetById))
	srv.AddTool(tools.ConnectorIPSecAdvGetByName(), perCall(get, handlers.ConnectorIPSecAdvGetByName))
	srv.AddTool(tools.ConnectorIPSecAdvGetTotal(), perCall(get, handlers.ConnectorIPSecAdvGetTotal))

	srv.AddTool(tools.ConnectorIPSecTunnelProfileGetAll(), perCall(get, handlers.ConnectorIPSecTunnelProfileGetAll))

	// Add internet application tools
	srv.AddTool(tools.InternetApplicationGetAll(), perCall(get, handlers.InternetApplicationGetAll))
	srv.AddTool(tools.InternetApplicationGetById(), perCall(get, handlers.InternetApplicationGetById))
	srv.AddTool(tools.InternetApplicationGetByName(), perCall(get, handlers.InternetApplicationGetByName))
	srv.AddTool(tools.InternetApplicationGetTotal(), perCall(get, handlers.InternetApplicationGetTotal))

	// Add routes tools
	srv.AddTool(tools.GetRoutes(), perCall(get, handlers.RouteGet))
	srv.AddTool(tools.GetRouteCount(), perCall(get, handlers.RouteGetCount))
	srv.AddTool(tools.GetRouteSummary(), perCall(get, handlers.RouteSummaryGet))
	srv.AddTool(tools.GetAllRoutes(), perCall(get, handlers.AllRouteGet))

	// Add NAT policy tools
	srv.AddTool(tools.PolicyNatGetAll(), perCall(get, handlers.PolicyNatGetAll))
	srv.AddTool(tools.PolicyNatGetById(), perCall(get, handlers.PolicyNatGetById))
	srv.AddTool(tools.PolicyNatGetByName(), perCall(get, handlers.PolicyNatGetByName))
	srv.AddTool(tools.PolicyNatGetTotal(), perCall(get, handlers.PolicyNatGetTotal))

	srv.AddTool(tools.PolicyNatRuleGetAll(), perCall(get, handlers.PolicyNatRuleGetAll))
	srv.AddTool(tools.PolicyNatRuleGetById(), perCall(get, handlers.PolicyNatRuleGetById))
	srv.AddTool(tools.PolicyNatRuleGetByName(), perCall(get, handlers.PolicyNatRuleGetByName))
	srv.AddTool(tools.PolicyNatRuleGetTotal(), perCall(get, handlers.PolicyNatRuleGetTotal))

	// Add Route policy tools
	srv.AddTool(tools.PolicyRouteGetAll(), perCall(get, handlers.PolicyRouteGetAll))
	srv.AddTool(tools.PolicyRouteGetById(), perCall(get, handlers.PolicyRouteGetById))
	srv.AddTool(tools.PolicyRouteGetByName(), perCall(get, handlers.PolicyRouteGetByName))
	srv.AddTool(tools.PolicyRouteGetTotal(), perCall(get, handlers.PolicyRouteGetTotal))

	// Add Traffic policy tools
	srv.AddTool(tools.PolicyTrafficGetAll(), perCall(get, handlers.PolicyTrafficGetAll))
	srv.AddTool(tools.PolicyTrafficGetById(), perCall(get, handlers.PolicyTrafficGetById))
	srv.AddTool(tools.PolicyTrafficGetByName(), perCall(get, handlers.PolicyTrafficGetByName))

	srv.AddTool(tools.PolicyTrafficRuleGetAll(), perCall(get, handlers.PolicyTrafficRuleGetAll))
	srv.AddTool(tools.PolicyTrafficRuleGetById(), perCall(get, handlers.PolicyTrafficRuleGetById))
	srv.AddTool(tools.PolicyTrafficRuleGetByName(), perCall(get, handlers.PolicyTrafficRuleGetByName))

	// Add traffic policy rule list tools
	srv.AddTool(tools.PolicyTrafficRuleListGetAll(), perCall(get, handlers.PolicyTrafficRuleListGetAll))
	srv.AddTool(tools.PolicyTrafficRuleListGetById(), perCall(get, handlers.PolicyTrafficRuleListGetById))
	srv.AddTool(tools.PolicyTrafficRuleListGetByName(), perCall(get, handlers.PolicyTrafficRuleListGetByName))

	// Add AS Path list tools
	srv.AddTool(tools.ListAsPathGetAll(), perCall(get, handlers.ListAsPathGetAll))
	srv.AddTool(tools.ListAsPathGetById(), perCall(get, handlers.ListAsPathGetById))
	srv.AddTool(tools.ListAsPathGetByName(), perCall(get, handlers.ListAsPathGetByName))
	srv.AddTool(tools.ListAsPathGetTotal(), perCall(get, handlers.ListAsPathGetTotal))

	// Add BGP Community list tools
	srv.AddTool(tools.ListCommunityGetAll(), perCall(get, handlers.ListCommunityGetAll))
	srv.AddTool(tools.ListCommunityGetById(), perCall(get, handlers.ListCommunityGetById))
	srv.AddTool(tools.ListCommunityGetByName(), perCall(get, handlers.ListCommunityGetByName))
	srv.AddTool(tools.ListCommunityGetTotal(), perCall(get, handlers.ListCommunityGetTotal))

	// Add BGP Extended Community list tools
	srv.AddTool(tools.ListExtendedCommunityGetAll(), perCall(get, handlers.ListExtendedCommunityGetAll))
	srv.AddTool(tools.ListExtendedCommunityGetById(), perCall(get, handlers.ListExtendedCommunityGetById))
	srv.AddTool(tools.ListExtendedCommunityGetByName(), perCall(get, handlers.ListExtendedCommunityGetByName))
	srv.AddTool(tools.ListExtendedCommunityGetTotal(), perCall(get, handlers.ListExtendedCommunityGetTotal))

	// Add DNS Server list tools
	srv.AddTool(tools.ListDnsServerGetAll(), perCall(get, handlers.ListDnsServerGetAll))
	srv.AddTool(tools.ListDnsServerGetById(), perCall(get, handlers.ListDnsServerGetById))
	srv.AddTool(tools.ListDnsServerGetByName(), perCall(get, handlers.ListDnsServerGetByName))
	srv.AddTool(tools.ListDnsServerGetTotal(), perCall(get, handlers.ListDnsServerGetTotal))

	// Add Global CIDR list tools
	srv.AddTool(tools.ListGlobalCidrGetAll(), perCall(get, handlers.ListGlobalCidrGetAll))
	srv.AddTool(tools.ListGlobalCidrGetById(), perCall(get, handlers.ListGlobalCidrGetById))
	srv.AddTool(tools.ListGlobalCidrGetByName(), perCall(get, handlers.ListGlobalCidrGetByName))
	srv.AddTool(tools.ListGlobalCidrGetTotal(), perCall(get, handlers.ListGlobalCidrGetTotal))

	// Add User Defined Route (UDR) list tools
	srv.AddTool(tools.ListUdrGetAll(), perCall(get, handlers.ListUdrGetAll))
	srv.AddTool(tools.ListUdrGetById(), perCall(get, handlers.ListUdrGetById))
	srv.AddTool(tools.ListUdrGetByName(), perCall(get, handlers.ListUdrGetByName))
	srv.AddTool(tools.ListUdrGetTotal(), perCall(get, handlers.ListUdrGetTotal))

	// Add Policy Prefix list tools
	srv.AddTool(tools.ListPrefixGetById(), perCall(get, handlers.ListPrefixGetById))
	srv.AddTool(tools.ListPrefixGetByName(), perCall(get, handlers.ListPrefixGetByName))
	srv.AddTool(tools.ListPrefixGetByPrefix(), perCall(get, handlers.ListPrefixGetByPrefix))

	// Add Policy FQDN list tools
	srv.AddTool(tools.ListPolicyFqdnGetAll(), perCall(get, handlers.ListPolicyFqdnGetAll))
	srv.AddTool(tools.ListPolicyFqdnGetById(), perCall(get, handlers.ListPolicyFqdnGetById))
	srv.AddTool(tools.ListPolicyFqdnGetByName(), perCall(get, handlers.ListPolicyFqdnGetByName))
	srv.AddTool(tools.ListPolicyFqdnGetTotal(), perCall(get, handlers.ListPolicyFqdnGetTotal))

	// Add monitoring tools
	srv.AddTool(tools.Alerts(), perCall(get, handlers.Alerts))
	srv.AddTool(tools.AuditLogs(), perCall(get, handlers.AuditLogs))
	srv.AddTool(tools.Jobs(), perCall(get, handlers.Jobs))

	srv.AddTool(tools.TenantResourceUsages(), perCall(get, handlers.TenantResourceUsages))
	srv.AddTool(tools.TenantResourceLimits(), perCall(get, handlers.TenantResourceLimits))

	// Add health tools
	srv.AddTool(tools.HealthConnectorGetById(), perCall(get, handlers.HealthConnectorGetById))
	srv.AddTool(tools.HealthConnectorInstanceGetById(), perCall(get, handlers.HealthConnectorInstanceGetById))
	srv.AddTool(tools.HealthServiceGetById(), perCall(get, handlers.HealthServiceGetById))
	srv.AddTool(tools.HealthServiceInstanceGetById(), perCall(get, handlers.HealthServiceInstanceGetById))
}
