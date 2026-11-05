// Package registry registers the mcp-alkira tools on an MCP server.
package registry

import (
	"github.com/alkiranet/mcp-alkira/handlers"
	"github.com/alkiranet/mcp-alkira/tools"

	"github.com/mark3labs/mcp-go/server"
)

// Register adds every mcp-alkira tool to srv; each tool call obtains its client from get.
func Register(srv *server.MCPServer, get handlers.ClientProvider) {
	// Add tenant network basic tools
	srv.AddTool(tools.TenantNetworkSummary(), handlers.TenantNetworkSummary(get))
	srv.AddTool(tools.TenantNetworkFirewallZones(), handlers.TenantNetworkFirewallZones(get))

	// Add billing tag tools
	srv.AddTool(tools.BillingTagGetAll(), handlers.BillingTagGetAll(get))
	srv.AddTool(tools.BillingTagGetById(), handlers.BillingTagGetById(get))
	srv.AddTool(tools.BillingTagGetByName(), handlers.BillingTagGetByName(get))
	srv.AddTool(tools.BillingTagGetTotal(), handlers.BillingTagGetTotal(get))

	srv.AddTool(tools.CxpGetAll(), handlers.CxpGetAll(get))

	// Add segment tools
	srv.AddTool(tools.SegmentGetAll(), handlers.SegmentGetAll(get))
	srv.AddTool(tools.SegmentGetById(), handlers.SegmentGetById(get))
	srv.AddTool(tools.SegmentGetByName(), handlers.SegmentGetByName(get))

	// Add group tools
	srv.AddTool(tools.GroupGetById(), handlers.GroupGetById(get))
	srv.AddTool(tools.GroupGetByName(), handlers.GroupGetByName(get))
	srv.AddTool(tools.GroupGetTotal(), handlers.GroupGetTotal(get))

	// Add BYOIP tools
	srv.AddTool(tools.ByoipGetAll(), handlers.ByoipGetAll(get))
	srv.AddTool(tools.ByoipGetById(), handlers.ByoipGetById(get))
	srv.AddTool(tools.ByoipGetByName(), handlers.ByoipGetByName(get))
	srv.AddTool(tools.ByoipGetTotal(), handlers.ByoipGetTotal(get))

	// Add segment resource tools
	srv.AddTool(tools.SegmentResourceGetAll(), handlers.SegmentResourceGetAll(get))
	srv.AddTool(tools.SegmentResourceGetById(), handlers.SegmentResourceGetById(get))
	srv.AddTool(tools.SegmentResourceGetByName(), handlers.SegmentResourceGetByName(get))
	srv.AddTool(tools.SegmentResourceGetTotal(), handlers.SegmentResourceGetTotal(get))

	// Add segment resource share tools
	srv.AddTool(tools.SegmentResourceShareGetAll(), handlers.SegmentResourceShareGetAll(get))
	srv.AddTool(tools.SegmentResourceShareGetById(), handlers.SegmentResourceShareGetById(get))
	srv.AddTool(tools.SegmentResourceShareGetByName(), handlers.SegmentResourceShareGetByName(get))
	srv.AddTool(tools.SegmentResourceShareGetTotal(), handlers.SegmentResourceShareGetTotal(get))

	// Add Check Point firewall service tools
	srv.AddTool(tools.ServiceCheckPointGetAll(), handlers.ServiceCheckPointGetAll(get))
	srv.AddTool(tools.ServiceCheckPointGetById(), handlers.ServiceCheckPointGetById(get))
	srv.AddTool(tools.ServiceCheckPointGetByName(), handlers.ServiceCheckPointGetByName(get))
	srv.AddTool(tools.ServiceCheckPointGetTotal(), handlers.ServiceCheckPointGetTotal(get))

	// Add Cisco FTDV service tools
	srv.AddTool(tools.ServiceCiscoFTDvGetAll(), handlers.ServiceCiscoFTDvGetAll(get))
	srv.AddTool(tools.ServiceCiscoFTDvGetById(), handlers.ServiceCiscoFTDvGetById(get))
	srv.AddTool(tools.ServiceCiscoFTDvGetByName(), handlers.ServiceCiscoFTDvGetByName(get))
	srv.AddTool(tools.ServiceCiscoFTDvGetTotal(), handlers.ServiceCiscoFTDvGetTotal(get))

	// Add F5 LB service tools
	srv.AddTool(tools.ServiceF5LBGetAll(), handlers.ServiceF5LBGetAll(get))
	srv.AddTool(tools.ServiceF5LBGetById(), handlers.ServiceF5LBGetById(get))
	srv.AddTool(tools.ServiceF5LBGetByName(), handlers.ServiceF5LBGetByName(get))
	srv.AddTool(tools.ServiceF5LBGetTotal(), handlers.ServiceF5LBGetTotal(get))

	// Add Fortinet service tools
	srv.AddTool(tools.ServiceFortinetGetAll(), handlers.ServiceFortinetGetAll(get))
	srv.AddTool(tools.ServiceFortinetGetById(), handlers.ServiceFortinetGetById(get))
	srv.AddTool(tools.ServiceFortinetGetByName(), handlers.ServiceFortinetGetByName(get))
	srv.AddTool(tools.ServiceFortinetGetTotal(), handlers.ServiceFortinetGetTotal(get))

	// Add Infoblox service tools
	srv.AddTool(tools.ServiceInfobloxGetAll(), handlers.ServiceInfobloxGetAll(get))
	srv.AddTool(tools.ServiceInfobloxGetById(), handlers.ServiceInfobloxGetById(get))
	srv.AddTool(tools.ServiceInfobloxGetByName(), handlers.ServiceInfobloxGetByName(get))
	srv.AddTool(tools.ServiceInfobloxGetTotal(), handlers.ServiceInfobloxGetTotal(get))

	// Add PAN firewall service tools
	srv.AddTool(tools.ServicePanGetAll(), handlers.ServicePanGetAll(get))
	srv.AddTool(tools.ServicePanGetById(), handlers.ServicePanGetById(get))
	srv.AddTool(tools.ServicePanGetByName(), handlers.ServicePanGetByName(get))
	srv.AddTool(tools.ServicePanGetTotal(), handlers.ServicePanGetTotal(get))

	// Add ZScaler service tools
	srv.AddTool(tools.ServiceZscalerGetAll(), handlers.ServiceZscalerGetAll(get))
	srv.AddTool(tools.ServiceZscalerGetById(), handlers.ServiceZscalerGetById(get))
	srv.AddTool(tools.ServiceZscalerGetByName(), handlers.ServiceZscalerGetByName(get))
	srv.AddTool(tools.ServiceZscalerGetTotal(), handlers.ServiceZscalerGetTotal(get))

	// Add Aruba Edge Connect SD-WAN connector tools
	srv.AddTool(tools.ConnectorArubaEdgeGetAll(), handlers.ConnectorArubaEdgeGetAll(get))
	srv.AddTool(tools.ConnectorArubaEdgeGetById(), handlers.ConnectorArubaEdgeGetById(get))
	srv.AddTool(tools.ConnectorArubaEdgeGetByName(), handlers.ConnectorArubaEdgeGetByName(get))
	srv.AddTool(tools.ConnectorArubaEdgeGetTotal(), handlers.ConnectorArubaEdgeGetTotal(get))

	// Add AWS Direct Connect connector tools
	srv.AddTool(tools.ConnectorAwsDirectConnectGetAll(), handlers.ConnectorAwsDirectConnectGetAll(get))
	srv.AddTool(tools.ConnectorAwsDirectConnectGetById(), handlers.ConnectorAwsDirectConnectGetById(get))
	srv.AddTool(tools.ConnectorAwsDirectConnectGetByName(), handlers.ConnectorAwsDirectConnectGetByName(get))
	srv.AddTool(tools.ConnectorAwsDirectConnectGetTotal(), handlers.ConnectorAwsDirectConnectGetTotal(get))

	// Add AWS Transit Gateway connector tools
	srv.AddTool(tools.ConnectorAwsTgwGetAll(), handlers.ConnectorAwsTgwGetAll(get))
	srv.AddTool(tools.ConnectorAwsTgwGetById(), handlers.ConnectorAwsTgwGetById(get))
	srv.AddTool(tools.ConnectorAwsTgwGetByName(), handlers.ConnectorAwsTgwGetByName(get))
	srv.AddTool(tools.ConnectorAwsTgwGetTotal(), handlers.ConnectorAwsTgwGetTotal(get))

	// Add Azure ExpressRoute connector tools
	srv.AddTool(tools.ConnectorAzureExpressRouteGetAll(), handlers.ConnectorAzureExpressRouteGetAll(get))
	srv.AddTool(tools.ConnectorAzureExpressRouteGetById(), handlers.ConnectorAzureExpressRouteGetById(get))
	srv.AddTool(tools.ConnectorAzureExpressRouteGetByName(), handlers.ConnectorAzureExpressRouteGetByName(get))
	srv.AddTool(tools.ConnectorAzureExpressRouteGetTotal(), handlers.ConnectorAzureExpressRouteGetTotal(get))

	// Add Cisco SD-WAN connector tools
	srv.AddTool(tools.ConnectorCiscoSdwanGetAll(), handlers.ConnectorCiscoSdwanGetAll(get))
	srv.AddTool(tools.ConnectorCiscoSdwanGetById(), handlers.ConnectorCiscoSdwanGetById(get))
	srv.AddTool(tools.ConnectorCiscoSdwanGetByName(), handlers.ConnectorCiscoSdwanGetByName(get))
	srv.AddTool(tools.ConnectorCiscoSdwanGetTotal(), handlers.ConnectorCiscoSdwanGetTotal(get))

	// Add Fortinet SD-WAN connector tools
	srv.AddTool(tools.ConnectorFortinetSdwanGetAll(), handlers.ConnectorFortinetSdwanGetAll(get))
	srv.AddTool(tools.ConnectorFortinetSdwanGetById(), handlers.ConnectorFortinetSdwanGetById(get))
	srv.AddTool(tools.ConnectorFortinetSdwanGetByName(), handlers.ConnectorFortinetSdwanGetByName(get))
	srv.AddTool(tools.ConnectorFortinetSdwanGetTotal(), handlers.ConnectorFortinetSdwanGetTotal(get))

	// Add Google Cloud Interconnect connector tools
	srv.AddTool(tools.ConnectorGcpInterconnectGetAll(), handlers.ConnectorGcpInterconnectGetAll(get))
	srv.AddTool(tools.ConnectorGcpInterconnectGetById(), handlers.ConnectorGcpInterconnectGetById(get))
	srv.AddTool(tools.ConnectorGcpInterconnectGetByName(), handlers.ConnectorGcpInterconnectGetByName(get))
	srv.AddTool(tools.ConnectorGcpInterconnectGetTotal(), handlers.ConnectorGcpInterconnectGetTotal(get))

	// Add Google Cloud VPC connector tools
	srv.AddTool(tools.ConnectorGcpVpcGetAll(), handlers.ConnectorGcpVpcGetAll(get))
	srv.AddTool(tools.ConnectorGcpVpcGetById(), handlers.ConnectorGcpVpcGetById(get))
	srv.AddTool(tools.ConnectorGcpVpcGetByName(), handlers.ConnectorGcpVpcGetByName(get))
	srv.AddTool(tools.ConnectorGcpVpcGetTotal(), handlers.ConnectorGcpVpcGetTotal(get))

	// Add Oracle Cloud Infrastructure VCN connector tools
	srv.AddTool(tools.ConnectorOciVcnGetAll(), handlers.ConnectorOciVcnGetAll(get))
	srv.AddTool(tools.ConnectorOciVcnGetById(), handlers.ConnectorOciVcnGetById(get))
	srv.AddTool(tools.ConnectorOciVcnGetByName(), handlers.ConnectorOciVcnGetByName(get))
	srv.AddTool(tools.ConnectorOciVcnGetTotal(), handlers.ConnectorOciVcnGetTotal(get))

	// Add Remote Access Template connector tools
	srv.AddTool(tools.ConnectorRemoteAccessGetAll(), handlers.ConnectorRemoteAccessGetAll(get))
	srv.AddTool(tools.ConnectorRemoteAccessGetById(), handlers.ConnectorRemoteAccessGetById(get))
	srv.AddTool(tools.ConnectorRemoteAccessGetByName(), handlers.ConnectorRemoteAccessGetByName(get))
	srv.AddTool(tools.ConnectorRemoteAccessGetTotal(), handlers.ConnectorRemoteAccessGetTotal(get))

	// Add Versa SD-WAN connector tools
	srv.AddTool(tools.ConnectorVersaSdwanGetAll(), handlers.ConnectorVersaSdwanGetAll(get))
	srv.AddTool(tools.ConnectorVersaSdwanGetById(), handlers.ConnectorVersaSdwanGetById(get))
	srv.AddTool(tools.ConnectorVersaSdwanGetByName(), handlers.ConnectorVersaSdwanGetByName(get))
	srv.AddTool(tools.ConnectorVersaSdwanGetTotal(), handlers.ConnectorVersaSdwanGetTotal(get))

	// Add VMware SD-WAN (VeloCloud) connector tools
	srv.AddTool(tools.ConnectorVmwareSdwanGetAll(), handlers.ConnectorVmwareSdwanGetAll(get))
	srv.AddTool(tools.ConnectorVmwareSdwanGetById(), handlers.ConnectorVmwareSdwanGetById(get))
	srv.AddTool(tools.ConnectorVmwareSdwanGetByName(), handlers.ConnectorVmwareSdwanGetByName(get))
	srv.AddTool(tools.ConnectorVmwareSdwanGetTotal(), handlers.ConnectorVmwareSdwanGetTotal(get))

	// Add internet connector tools
	srv.AddTool(tools.ConnectorInternetGetAll(), handlers.ConnectorInternetGetAll(get))
	srv.AddTool(tools.ConnectorInternetGetById(), handlers.ConnectorInternetGetById(get))
	srv.AddTool(tools.ConnectorInternetGetByName(), handlers.ConnectorInternetGetByName(get))
	srv.AddTool(tools.ConnectorInternetGetTotal(), handlers.ConnectorInternetGetTotal(get))

	// Add AWS VPC connector tools
	srv.AddTool(tools.ConnectorAwsVpcGetAll(), handlers.ConnectorAwsVpcGetAll(get))
	srv.AddTool(tools.ConnectorAwsVpcGetById(), handlers.ConnectorAwsVpcGetById(get))
	srv.AddTool(tools.ConnectorAwsVpcGetByName(), handlers.ConnectorAwsVpcGetByName(get))
	srv.AddTool(tools.ConnectorAwsVpcGetTotal(), handlers.ConnectorAwsVpcGetTotal(get))

	// Add Azure VNET connector tools
	srv.AddTool(tools.ConnectorAzureVnetGetAll(), handlers.ConnectorAzureVnetGetAll(get))
	srv.AddTool(tools.ConnectorAzureVnetGetById(), handlers.ConnectorAzureVnetGetById(get))
	srv.AddTool(tools.ConnectorAzureVnetGetByName(), handlers.ConnectorAzureVnetGetByName(get))
	srv.AddTool(tools.ConnectorAzureVnetGetTotal(), handlers.ConnectorAzureVnetGetTotal(get))

	// Add IPSec connector tools
	srv.AddTool(tools.ConnectorIPSecGetAll(), handlers.ConnectorIPSecGetAll(get))
	srv.AddTool(tools.ConnectorIPSecGetById(), handlers.ConnectorIPSecGetById(get))
	srv.AddTool(tools.ConnectorIPSecGetByName(), handlers.ConnectorIPSecGetByName(get))
	srv.AddTool(tools.ConnectorIPSecGetTotal(), handlers.ConnectorIPSecGetTotal(get))

	// Add IPSec Adv connector tools
	srv.AddTool(tools.ConnectorIPSecAdvGetAll(), handlers.ConnectorIPSecAdvGetAll(get))
	srv.AddTool(tools.ConnectorIPSecAdvGetById(), handlers.ConnectorIPSecAdvGetById(get))
	srv.AddTool(tools.ConnectorIPSecAdvGetByName(), handlers.ConnectorIPSecAdvGetByName(get))
	srv.AddTool(tools.ConnectorIPSecAdvGetTotal(), handlers.ConnectorIPSecAdvGetTotal(get))

	srv.AddTool(tools.ConnectorIPSecTunnelProfileGetAll(), handlers.ConnectorIPSecTunnelProfileGetAll(get))

	// Add internet application tools
	srv.AddTool(tools.InternetApplicationGetAll(), handlers.InternetApplicationGetAll(get))
	srv.AddTool(tools.InternetApplicationGetById(), handlers.InternetApplicationGetById(get))
	srv.AddTool(tools.InternetApplicationGetByName(), handlers.InternetApplicationGetByName(get))
	srv.AddTool(tools.InternetApplicationGetTotal(), handlers.InternetApplicationGetTotal(get))

	// Add routes tools
	srv.AddTool(tools.GetRoutes(), handlers.RouteGet(get))
	srv.AddTool(tools.GetRouteCount(), handlers.RouteGetCount(get))
	srv.AddTool(tools.GetRouteSummary(), handlers.RouteSummaryGet(get))
	srv.AddTool(tools.GetAllRoutes(), handlers.AllRouteGet(get))

	// Add NAT policy tools
	srv.AddTool(tools.PolicyNatGetAll(), handlers.PolicyNatGetAll(get))
	srv.AddTool(tools.PolicyNatGetById(), handlers.PolicyNatGetById(get))
	srv.AddTool(tools.PolicyNatGetByName(), handlers.PolicyNatGetByName(get))
	srv.AddTool(tools.PolicyNatGetTotal(), handlers.PolicyNatGetTotal(get))

	srv.AddTool(tools.PolicyNatRuleGetAll(), handlers.PolicyNatRuleGetAll(get))
	srv.AddTool(tools.PolicyNatRuleGetById(), handlers.PolicyNatRuleGetById(get))
	srv.AddTool(tools.PolicyNatRuleGetByName(), handlers.PolicyNatRuleGetByName(get))
	srv.AddTool(tools.PolicyNatRuleGetTotal(), handlers.PolicyNatRuleGetTotal(get))

	// Add Route policy tools
	srv.AddTool(tools.PolicyRouteGetAll(), handlers.PolicyRouteGetAll(get))
	srv.AddTool(tools.PolicyRouteGetById(), handlers.PolicyRouteGetById(get))
	srv.AddTool(tools.PolicyRouteGetByName(), handlers.PolicyRouteGetByName(get))
	srv.AddTool(tools.PolicyRouteGetTotal(), handlers.PolicyRouteGetTotal(get))

	// Add Traffic policy tools
	srv.AddTool(tools.PolicyTrafficGetAll(), handlers.PolicyTrafficGetAll(get))
	srv.AddTool(tools.PolicyTrafficGetById(), handlers.PolicyTrafficGetById(get))
	srv.AddTool(tools.PolicyTrafficGetByName(), handlers.PolicyTrafficGetByName(get))

	srv.AddTool(tools.PolicyTrafficRuleGetAll(), handlers.PolicyTrafficRuleGetAll(get))
	srv.AddTool(tools.PolicyTrafficRuleGetById(), handlers.PolicyTrafficRuleGetById(get))
	srv.AddTool(tools.PolicyTrafficRuleGetByName(), handlers.PolicyTrafficRuleGetByName(get))

	// Add traffic policy rule list tools
	srv.AddTool(tools.PolicyTrafficRuleListGetAll(), handlers.PolicyTrafficRuleListGetAll(get))
	srv.AddTool(tools.PolicyTrafficRuleListGetById(), handlers.PolicyTrafficRuleListGetById(get))
	srv.AddTool(tools.PolicyTrafficRuleListGetByName(), handlers.PolicyTrafficRuleListGetByName(get))

	// Add AS Path list tools
	srv.AddTool(tools.ListAsPathGetAll(), handlers.ListAsPathGetAll(get))
	srv.AddTool(tools.ListAsPathGetById(), handlers.ListAsPathGetById(get))
	srv.AddTool(tools.ListAsPathGetByName(), handlers.ListAsPathGetByName(get))
	srv.AddTool(tools.ListAsPathGetTotal(), handlers.ListAsPathGetTotal(get))

	// Add BGP Community list tools
	srv.AddTool(tools.ListCommunityGetAll(), handlers.ListCommunityGetAll(get))
	srv.AddTool(tools.ListCommunityGetById(), handlers.ListCommunityGetById(get))
	srv.AddTool(tools.ListCommunityGetByName(), handlers.ListCommunityGetByName(get))
	srv.AddTool(tools.ListCommunityGetTotal(), handlers.ListCommunityGetTotal(get))

	// Add BGP Extended Community list tools
	srv.AddTool(tools.ListExtendedCommunityGetAll(), handlers.ListExtendedCommunityGetAll(get))
	srv.AddTool(tools.ListExtendedCommunityGetById(), handlers.ListExtendedCommunityGetById(get))
	srv.AddTool(tools.ListExtendedCommunityGetByName(), handlers.ListExtendedCommunityGetByName(get))
	srv.AddTool(tools.ListExtendedCommunityGetTotal(), handlers.ListExtendedCommunityGetTotal(get))

	// Add DNS Server list tools
	srv.AddTool(tools.ListDnsServerGetAll(), handlers.ListDnsServerGetAll(get))
	srv.AddTool(tools.ListDnsServerGetById(), handlers.ListDnsServerGetById(get))
	srv.AddTool(tools.ListDnsServerGetByName(), handlers.ListDnsServerGetByName(get))
	srv.AddTool(tools.ListDnsServerGetTotal(), handlers.ListDnsServerGetTotal(get))

	// Add Global CIDR list tools
	srv.AddTool(tools.ListGlobalCidrGetAll(), handlers.ListGlobalCidrGetAll(get))
	srv.AddTool(tools.ListGlobalCidrGetById(), handlers.ListGlobalCidrGetById(get))
	srv.AddTool(tools.ListGlobalCidrGetByName(), handlers.ListGlobalCidrGetByName(get))
	srv.AddTool(tools.ListGlobalCidrGetTotal(), handlers.ListGlobalCidrGetTotal(get))

	// Add User Defined Route (UDR) list tools
	srv.AddTool(tools.ListUdrGetAll(), handlers.ListUdrGetAll(get))
	srv.AddTool(tools.ListUdrGetById(), handlers.ListUdrGetById(get))
	srv.AddTool(tools.ListUdrGetByName(), handlers.ListUdrGetByName(get))
	srv.AddTool(tools.ListUdrGetTotal(), handlers.ListUdrGetTotal(get))

	// Add Policy Prefix list tools
	srv.AddTool(tools.ListPrefixGetById(), handlers.ListPrefixGetById(get))
	srv.AddTool(tools.ListPrefixGetByName(), handlers.ListPrefixGetByName(get))
	srv.AddTool(tools.ListPrefixGetByPrefix(), handlers.ListPrefixGetByPrefix(get))

	// Add Policy FQDN list tools
	srv.AddTool(tools.ListPolicyFqdnGetAll(), handlers.ListPolicyFqdnGetAll(get))
	srv.AddTool(tools.ListPolicyFqdnGetById(), handlers.ListPolicyFqdnGetById(get))
	srv.AddTool(tools.ListPolicyFqdnGetByName(), handlers.ListPolicyFqdnGetByName(get))
	srv.AddTool(tools.ListPolicyFqdnGetTotal(), handlers.ListPolicyFqdnGetTotal(get))

	// Add monitoring tools
	srv.AddTool(tools.Alerts(), handlers.Alerts(get))
	srv.AddTool(tools.AuditLogs(), handlers.AuditLogs(get))
	srv.AddTool(tools.Jobs(), handlers.Jobs(get))

	srv.AddTool(tools.TenantResourceUsages(), handlers.TenantResourceUsages(get))
	srv.AddTool(tools.TenantResourceLimits(), handlers.TenantResourceLimits(get))

	// Add health tools
	srv.AddTool(tools.HealthConnectorGetById(), handlers.HealthConnectorGetById(get))
	srv.AddTool(tools.HealthConnectorInstanceGetById(), handlers.HealthConnectorInstanceGetById(get))
	srv.AddTool(tools.HealthServiceGetById(), handlers.HealthServiceGetById(get))
	srv.AddTool(tools.HealthServiceInstanceGetById(), handlers.HealthServiceInstanceGetById(get))
}
