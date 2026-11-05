package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/alkiranet/mcp-alkira/handlers"
	"github.com/alkiranet/mcp-alkira/prompts"
	"github.com/alkiranet/mcp-alkira/tools"

	ak "github.com/alkiranet/client-go/tenant"

	"github.com/mark3labs/mcp-go/server"
)

// logf a simple log wrapper to log based on ENV var
func logf(level string, message string, v ...interface{}) {
	logLevel := os.Getenv("MCP_LOG")

	if logLevel == level {
		format := fmt.Sprintf("[%s] %s", level, message)
		log.Printf(format, v...)
	}
}

// getEnvOrDefault returns the value of the environment variable or the default value if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {

	//
	// Parse command line flags
	//
	var mode string
	var port string

	var portal string
	var key string
	var maxToken int

	flag.StringVar(&port, "port", getEnvOrDefault("SERVER_PORT", "8081"), "Server port")
	flag.StringVar(&mode, "mode", getEnvOrDefault("SERVER_MODE", "stdio"), "Server mode: 'stdio' or 'sse'")

	flag.StringVar(&portal, "portal", "", "Alkira Portal URL")
	flag.StringVar(&key, "key", "", "Alkira API Key")
	flag.IntVar(&maxToken, "maxToken", 0, "Max Token of payload")
	flag.Parse()

	// Create Alkira Client
	if portal == "" || key == "" {
		logf("ERROR", "Invalid ENV vars, please specify AK_PORTAL and AK_KEY.")
		return
	}

	alkiraClient, err := ak.NewAlkiraClient(
		portal,
		key,
		maxToken,
		"toon",
	)

	if err != nil {
		logf("ERROR", "failed to initialize alkira client, please check your credential and portal URI.")
		return
	}

	getClient := func(context.Context) (*ak.AlkiraClient, error) { return alkiraClient, nil }

	// Create MCP server
	srv := server.NewMCPServer(
		"mcp-alkira",
		"0.2.0",
	)

	// Prompts
	srv.AddPrompt(prompts.Summary(), prompts.SummaryHandler())

	// Add tenant network basic tools
	srv.AddTool(tools.TenantNetworkSummary(), handlers.TenantNetworkSummary(getClient))
	srv.AddTool(tools.TenantNetworkFirewallZones(), handlers.TenantNetworkFirewallZones(getClient))

	// Add billing tag tools
	srv.AddTool(tools.BillingTagGetAll(), handlers.BillingTagGetAll(getClient))
	srv.AddTool(tools.BillingTagGetById(), handlers.BillingTagGetById(getClient))
	srv.AddTool(tools.BillingTagGetByName(), handlers.BillingTagGetByName(getClient))
	srv.AddTool(tools.BillingTagGetTotal(), handlers.BillingTagGetTotal(getClient))

	srv.AddTool(tools.CxpGetAll(), handlers.CxpGetAll(getClient))

	// Add segment tools
	srv.AddTool(tools.SegmentGetAll(), handlers.SegmentGetAll(getClient))
	srv.AddTool(tools.SegmentGetById(), handlers.SegmentGetById(getClient))
	srv.AddTool(tools.SegmentGetByName(), handlers.SegmentGetByName(getClient))

	// Add group tools
	srv.AddTool(tools.GroupGetById(), handlers.GroupGetById(getClient))
	srv.AddTool(tools.GroupGetByName(), handlers.GroupGetByName(getClient))
	srv.AddTool(tools.GroupGetTotal(), handlers.GroupGetTotal(getClient))

	// Add BYOIP tools
	srv.AddTool(tools.ByoipGetAll(), handlers.ByoipGetAll(getClient))
	srv.AddTool(tools.ByoipGetById(), handlers.ByoipGetById(getClient))
	srv.AddTool(tools.ByoipGetByName(), handlers.ByoipGetByName(getClient))
	srv.AddTool(tools.ByoipGetTotal(), handlers.ByoipGetTotal(getClient))

	// Add segment resource tools
	srv.AddTool(tools.SegmentResourceGetAll(), handlers.SegmentResourceGetAll(getClient))
	srv.AddTool(tools.SegmentResourceGetById(), handlers.SegmentResourceGetById(getClient))
	srv.AddTool(tools.SegmentResourceGetByName(), handlers.SegmentResourceGetByName(getClient))
	srv.AddTool(tools.SegmentResourceGetTotal(), handlers.SegmentResourceGetTotal(getClient))

	// Add segment resource share tools
	srv.AddTool(tools.SegmentResourceShareGetAll(), handlers.SegmentResourceShareGetAll(getClient))
	srv.AddTool(tools.SegmentResourceShareGetById(), handlers.SegmentResourceShareGetById(getClient))
	srv.AddTool(tools.SegmentResourceShareGetByName(), handlers.SegmentResourceShareGetByName(getClient))
	srv.AddTool(tools.SegmentResourceShareGetTotal(), handlers.SegmentResourceShareGetTotal(getClient))

	// Add Check Point firewall service tools
	srv.AddTool(tools.ServiceCheckPointGetAll(), handlers.ServiceCheckPointGetAll(getClient))
	srv.AddTool(tools.ServiceCheckPointGetById(), handlers.ServiceCheckPointGetById(getClient))
	srv.AddTool(tools.ServiceCheckPointGetByName(), handlers.ServiceCheckPointGetByName(getClient))
	srv.AddTool(tools.ServiceCheckPointGetTotal(), handlers.ServiceCheckPointGetTotal(getClient))

	// Add Cisco FTDV service tools
	srv.AddTool(tools.ServiceCiscoFTDvGetAll(), handlers.ServiceCiscoFTDvGetAll(getClient))
	srv.AddTool(tools.ServiceCiscoFTDvGetById(), handlers.ServiceCiscoFTDvGetById(getClient))
	srv.AddTool(tools.ServiceCiscoFTDvGetByName(), handlers.ServiceCiscoFTDvGetByName(getClient))
	srv.AddTool(tools.ServiceCiscoFTDvGetTotal(), handlers.ServiceCiscoFTDvGetTotal(getClient))

	// Add F5 LB service tools
	srv.AddTool(tools.ServiceF5LBGetAll(), handlers.ServiceF5LBGetAll(getClient))
	srv.AddTool(tools.ServiceF5LBGetById(), handlers.ServiceF5LBGetById(getClient))
	srv.AddTool(tools.ServiceF5LBGetByName(), handlers.ServiceF5LBGetByName(getClient))
	srv.AddTool(tools.ServiceF5LBGetTotal(), handlers.ServiceF5LBGetTotal(getClient))

	// Add Fortinet service tools
	srv.AddTool(tools.ServiceFortinetGetAll(), handlers.ServiceFortinetGetAll(getClient))
	srv.AddTool(tools.ServiceFortinetGetById(), handlers.ServiceFortinetGetById(getClient))
	srv.AddTool(tools.ServiceFortinetGetByName(), handlers.ServiceFortinetGetByName(getClient))
	srv.AddTool(tools.ServiceFortinetGetTotal(), handlers.ServiceFortinetGetTotal(getClient))

	// Add Infoblox service tools
	srv.AddTool(tools.ServiceInfobloxGetAll(), handlers.ServiceInfobloxGetAll(getClient))
	srv.AddTool(tools.ServiceInfobloxGetById(), handlers.ServiceInfobloxGetById(getClient))
	srv.AddTool(tools.ServiceInfobloxGetByName(), handlers.ServiceInfobloxGetByName(getClient))
	srv.AddTool(tools.ServiceInfobloxGetTotal(), handlers.ServiceInfobloxGetTotal(getClient))

	// Add PAN firewall service tools
	srv.AddTool(tools.ServicePanGetAll(), handlers.ServicePanGetAll(getClient))
	srv.AddTool(tools.ServicePanGetById(), handlers.ServicePanGetById(getClient))
	srv.AddTool(tools.ServicePanGetByName(), handlers.ServicePanGetByName(getClient))
	srv.AddTool(tools.ServicePanGetTotal(), handlers.ServicePanGetTotal(getClient))

	// Add ZScaler service tools
	srv.AddTool(tools.ServiceZscalerGetAll(), handlers.ServiceZscalerGetAll(getClient))
	srv.AddTool(tools.ServiceZscalerGetById(), handlers.ServiceZscalerGetById(getClient))
	srv.AddTool(tools.ServiceZscalerGetByName(), handlers.ServiceZscalerGetByName(getClient))
	srv.AddTool(tools.ServiceZscalerGetTotal(), handlers.ServiceZscalerGetTotal(getClient))

	// Add Aruba Edge Connect SD-WAN connector tools
	srv.AddTool(tools.ConnectorArubaEdgeGetAll(), handlers.ConnectorArubaEdgeGetAll(getClient))
	srv.AddTool(tools.ConnectorArubaEdgeGetById(), handlers.ConnectorArubaEdgeGetById(getClient))
	srv.AddTool(tools.ConnectorArubaEdgeGetByName(), handlers.ConnectorArubaEdgeGetByName(getClient))
	srv.AddTool(tools.ConnectorArubaEdgeGetTotal(), handlers.ConnectorArubaEdgeGetTotal(getClient))

	// Add AWS Direct Connect connector tools
	srv.AddTool(tools.ConnectorAwsDirectConnectGetAll(), handlers.ConnectorAwsDirectConnectGetAll(getClient))
	srv.AddTool(tools.ConnectorAwsDirectConnectGetById(), handlers.ConnectorAwsDirectConnectGetById(getClient))
	srv.AddTool(tools.ConnectorAwsDirectConnectGetByName(), handlers.ConnectorAwsDirectConnectGetByName(getClient))
	srv.AddTool(tools.ConnectorAwsDirectConnectGetTotal(), handlers.ConnectorAwsDirectConnectGetTotal(getClient))

	// Add AWS Transit Gateway connector tools
	srv.AddTool(tools.ConnectorAwsTgwGetAll(), handlers.ConnectorAwsTgwGetAll(getClient))
	srv.AddTool(tools.ConnectorAwsTgwGetById(), handlers.ConnectorAwsTgwGetById(getClient))
	srv.AddTool(tools.ConnectorAwsTgwGetByName(), handlers.ConnectorAwsTgwGetByName(getClient))
	srv.AddTool(tools.ConnectorAwsTgwGetTotal(), handlers.ConnectorAwsTgwGetTotal(getClient))

	// Add Azure ExpressRoute connector tools
	srv.AddTool(tools.ConnectorAzureExpressRouteGetAll(), handlers.ConnectorAzureExpressRouteGetAll(getClient))
	srv.AddTool(tools.ConnectorAzureExpressRouteGetById(), handlers.ConnectorAzureExpressRouteGetById(getClient))
	srv.AddTool(tools.ConnectorAzureExpressRouteGetByName(), handlers.ConnectorAzureExpressRouteGetByName(getClient))
	srv.AddTool(tools.ConnectorAzureExpressRouteGetTotal(), handlers.ConnectorAzureExpressRouteGetTotal(getClient))

	// Add Cisco SD-WAN connector tools
	srv.AddTool(tools.ConnectorCiscoSdwanGetAll(), handlers.ConnectorCiscoSdwanGetAll(getClient))
	srv.AddTool(tools.ConnectorCiscoSdwanGetById(), handlers.ConnectorCiscoSdwanGetById(getClient))
	srv.AddTool(tools.ConnectorCiscoSdwanGetByName(), handlers.ConnectorCiscoSdwanGetByName(getClient))
	srv.AddTool(tools.ConnectorCiscoSdwanGetTotal(), handlers.ConnectorCiscoSdwanGetTotal(getClient))

	// Add Fortinet SD-WAN connector tools
	srv.AddTool(tools.ConnectorFortinetSdwanGetAll(), handlers.ConnectorFortinetSdwanGetAll(getClient))
	srv.AddTool(tools.ConnectorFortinetSdwanGetById(), handlers.ConnectorFortinetSdwanGetById(getClient))
	srv.AddTool(tools.ConnectorFortinetSdwanGetByName(), handlers.ConnectorFortinetSdwanGetByName(getClient))
	srv.AddTool(tools.ConnectorFortinetSdwanGetTotal(), handlers.ConnectorFortinetSdwanGetTotal(getClient))

	// Add Google Cloud Interconnect connector tools
	srv.AddTool(tools.ConnectorGcpInterconnectGetAll(), handlers.ConnectorGcpInterconnectGetAll(getClient))
	srv.AddTool(tools.ConnectorGcpInterconnectGetById(), handlers.ConnectorGcpInterconnectGetById(getClient))
	srv.AddTool(tools.ConnectorGcpInterconnectGetByName(), handlers.ConnectorGcpInterconnectGetByName(getClient))
	srv.AddTool(tools.ConnectorGcpInterconnectGetTotal(), handlers.ConnectorGcpInterconnectGetTotal(getClient))

	// Add Google Cloud VPC connector tools
	srv.AddTool(tools.ConnectorGcpVpcGetAll(), handlers.ConnectorGcpVpcGetAll(getClient))
	srv.AddTool(tools.ConnectorGcpVpcGetById(), handlers.ConnectorGcpVpcGetById(getClient))
	srv.AddTool(tools.ConnectorGcpVpcGetByName(), handlers.ConnectorGcpVpcGetByName(getClient))
	srv.AddTool(tools.ConnectorGcpVpcGetTotal(), handlers.ConnectorGcpVpcGetTotal(getClient))

	// Add Oracle Cloud Infrastructure VCN connector tools
	srv.AddTool(tools.ConnectorOciVcnGetAll(), handlers.ConnectorOciVcnGetAll(getClient))
	srv.AddTool(tools.ConnectorOciVcnGetById(), handlers.ConnectorOciVcnGetById(getClient))
	srv.AddTool(tools.ConnectorOciVcnGetByName(), handlers.ConnectorOciVcnGetByName(getClient))
	srv.AddTool(tools.ConnectorOciVcnGetTotal(), handlers.ConnectorOciVcnGetTotal(getClient))

	// Add Remote Access Template connector tools
	srv.AddTool(tools.ConnectorRemoteAccessGetAll(), handlers.ConnectorRemoteAccessGetAll(getClient))
	srv.AddTool(tools.ConnectorRemoteAccessGetById(), handlers.ConnectorRemoteAccessGetById(getClient))
	srv.AddTool(tools.ConnectorRemoteAccessGetByName(), handlers.ConnectorRemoteAccessGetByName(getClient))
	srv.AddTool(tools.ConnectorRemoteAccessGetTotal(), handlers.ConnectorRemoteAccessGetTotal(getClient))

	// Add Versa SD-WAN connector tools
	srv.AddTool(tools.ConnectorVersaSdwanGetAll(), handlers.ConnectorVersaSdwanGetAll(getClient))
	srv.AddTool(tools.ConnectorVersaSdwanGetById(), handlers.ConnectorVersaSdwanGetById(getClient))
	srv.AddTool(tools.ConnectorVersaSdwanGetByName(), handlers.ConnectorVersaSdwanGetByName(getClient))
	srv.AddTool(tools.ConnectorVersaSdwanGetTotal(), handlers.ConnectorVersaSdwanGetTotal(getClient))

	// Add VMware SD-WAN (VeloCloud) connector tools
	srv.AddTool(tools.ConnectorVmwareSdwanGetAll(), handlers.ConnectorVmwareSdwanGetAll(getClient))
	srv.AddTool(tools.ConnectorVmwareSdwanGetById(), handlers.ConnectorVmwareSdwanGetById(getClient))
	srv.AddTool(tools.ConnectorVmwareSdwanGetByName(), handlers.ConnectorVmwareSdwanGetByName(getClient))
	srv.AddTool(tools.ConnectorVmwareSdwanGetTotal(), handlers.ConnectorVmwareSdwanGetTotal(getClient))

	// Add internet connector tools
	srv.AddTool(tools.ConnectorInternetGetAll(), handlers.ConnectorInternetGetAll(getClient))
	srv.AddTool(tools.ConnectorInternetGetById(), handlers.ConnectorInternetGetById(getClient))
	srv.AddTool(tools.ConnectorInternetGetByName(), handlers.ConnectorInternetGetByName(getClient))
	srv.AddTool(tools.ConnectorInternetGetTotal(), handlers.ConnectorInternetGetTotal(getClient))

	// Add AWS VPC connector tools
	srv.AddTool(tools.ConnectorAwsVpcGetAll(), handlers.ConnectorAwsVpcGetAll(getClient))
	srv.AddTool(tools.ConnectorAwsVpcGetById(), handlers.ConnectorAwsVpcGetById(getClient))
	srv.AddTool(tools.ConnectorAwsVpcGetByName(), handlers.ConnectorAwsVpcGetByName(getClient))
	srv.AddTool(tools.ConnectorAwsVpcGetTotal(), handlers.ConnectorAwsVpcGetTotal(getClient))

	// Add Azure VNET connector tools
	srv.AddTool(tools.ConnectorAzureVnetGetAll(), handlers.ConnectorAzureVnetGetAll(getClient))
	srv.AddTool(tools.ConnectorAzureVnetGetById(), handlers.ConnectorAzureVnetGetById(getClient))
	srv.AddTool(tools.ConnectorAzureVnetGetByName(), handlers.ConnectorAzureVnetGetByName(getClient))
	srv.AddTool(tools.ConnectorAzureVnetGetTotal(), handlers.ConnectorAzureVnetGetTotal(getClient))

	// Add IPSec connector tools
	srv.AddTool(tools.ConnectorIPSecGetAll(), handlers.ConnectorIPSecGetAll(getClient))
	srv.AddTool(tools.ConnectorIPSecGetById(), handlers.ConnectorIPSecGetById(getClient))
	srv.AddTool(tools.ConnectorIPSecGetByName(), handlers.ConnectorIPSecGetByName(getClient))
	srv.AddTool(tools.ConnectorIPSecGetTotal(), handlers.ConnectorIPSecGetTotal(getClient))

	// Add IPSec Adv connector tools
	srv.AddTool(tools.ConnectorIPSecAdvGetAll(), handlers.ConnectorIPSecAdvGetAll(getClient))
	srv.AddTool(tools.ConnectorIPSecAdvGetById(), handlers.ConnectorIPSecAdvGetById(getClient))
	srv.AddTool(tools.ConnectorIPSecAdvGetByName(), handlers.ConnectorIPSecAdvGetByName(getClient))
	srv.AddTool(tools.ConnectorIPSecAdvGetTotal(), handlers.ConnectorIPSecAdvGetTotal(getClient))

	srv.AddTool(tools.ConnectorIPSecTunnelProfileGetAll(), handlers.ConnectorIPSecTunnelProfileGetAll(getClient))

	// Add internet application tools
	srv.AddTool(tools.InternetApplicationGetAll(), handlers.InternetApplicationGetAll(getClient))
	srv.AddTool(tools.InternetApplicationGetById(), handlers.InternetApplicationGetById(getClient))
	srv.AddTool(tools.InternetApplicationGetByName(), handlers.InternetApplicationGetByName(getClient))
	srv.AddTool(tools.InternetApplicationGetTotal(), handlers.InternetApplicationGetTotal(getClient))

	// Add routes tools
	srv.AddTool(tools.GetRoutes(), handlers.RouteGet(getClient))
	srv.AddTool(tools.GetRouteCount(), handlers.RouteGetCount(getClient))
	srv.AddTool(tools.GetRouteSummary(), handlers.RouteSummaryGet(getClient))
	srv.AddTool(tools.GetAllRoutes(), handlers.AllRouteGet(getClient))

	// Add NAT policy tools
	srv.AddTool(tools.PolicyNatGetAll(), handlers.PolicyNatGetAll(getClient))
	srv.AddTool(tools.PolicyNatGetById(), handlers.PolicyNatGetById(getClient))
	srv.AddTool(tools.PolicyNatGetByName(), handlers.PolicyNatGetByName(getClient))
	srv.AddTool(tools.PolicyNatGetTotal(), handlers.PolicyNatGetTotal(getClient))

	srv.AddTool(tools.PolicyNatRuleGetAll(), handlers.PolicyNatRuleGetAll(getClient))
	srv.AddTool(tools.PolicyNatRuleGetById(), handlers.PolicyNatRuleGetById(getClient))
	srv.AddTool(tools.PolicyNatRuleGetByName(), handlers.PolicyNatRuleGetByName(getClient))
	srv.AddTool(tools.PolicyNatRuleGetTotal(), handlers.PolicyNatRuleGetTotal(getClient))

	// Add Route policy tools
	srv.AddTool(tools.PolicyRouteGetAll(), handlers.PolicyRouteGetAll(getClient))
	srv.AddTool(tools.PolicyRouteGetById(), handlers.PolicyRouteGetById(getClient))
	srv.AddTool(tools.PolicyRouteGetByName(), handlers.PolicyRouteGetByName(getClient))
	srv.AddTool(tools.PolicyRouteGetTotal(), handlers.PolicyRouteGetTotal(getClient))

	// Add Traffic policy tools
	srv.AddTool(tools.PolicyTrafficGetAll(), handlers.PolicyTrafficGetAll(getClient))
	srv.AddTool(tools.PolicyTrafficGetById(), handlers.PolicyTrafficGetById(getClient))
	srv.AddTool(tools.PolicyTrafficGetByName(), handlers.PolicyTrafficGetByName(getClient))

	srv.AddTool(tools.PolicyTrafficRuleGetAll(), handlers.PolicyTrafficRuleGetAll(getClient))
	srv.AddTool(tools.PolicyTrafficRuleGetById(), handlers.PolicyTrafficRuleGetById(getClient))
	srv.AddTool(tools.PolicyTrafficRuleGetByName(), handlers.PolicyTrafficRuleGetByName(getClient))

	// Add traffic policy rule list tools
	srv.AddTool(tools.PolicyTrafficRuleListGetAll(), handlers.PolicyTrafficRuleListGetAll(getClient))
	srv.AddTool(tools.PolicyTrafficRuleListGetById(), handlers.PolicyTrafficRuleListGetById(getClient))
	srv.AddTool(tools.PolicyTrafficRuleListGetByName(), handlers.PolicyTrafficRuleListGetByName(getClient))

	// Add AS Path list tools
	srv.AddTool(tools.ListAsPathGetAll(), handlers.ListAsPathGetAll(getClient))
	srv.AddTool(tools.ListAsPathGetById(), handlers.ListAsPathGetById(getClient))
	srv.AddTool(tools.ListAsPathGetByName(), handlers.ListAsPathGetByName(getClient))
	srv.AddTool(tools.ListAsPathGetTotal(), handlers.ListAsPathGetTotal(getClient))

	// Add BGP Community list tools
	srv.AddTool(tools.ListCommunityGetAll(), handlers.ListCommunityGetAll(getClient))
	srv.AddTool(tools.ListCommunityGetById(), handlers.ListCommunityGetById(getClient))
	srv.AddTool(tools.ListCommunityGetByName(), handlers.ListCommunityGetByName(getClient))
	srv.AddTool(tools.ListCommunityGetTotal(), handlers.ListCommunityGetTotal(getClient))

	// Add BGP Extended Community list tools
	srv.AddTool(tools.ListExtendedCommunityGetAll(), handlers.ListExtendedCommunityGetAll(getClient))
	srv.AddTool(tools.ListExtendedCommunityGetById(), handlers.ListExtendedCommunityGetById(getClient))
	srv.AddTool(tools.ListExtendedCommunityGetByName(), handlers.ListExtendedCommunityGetByName(getClient))
	srv.AddTool(tools.ListExtendedCommunityGetTotal(), handlers.ListExtendedCommunityGetTotal(getClient))

	// Add DNS Server list tools
	srv.AddTool(tools.ListDnsServerGetAll(), handlers.ListDnsServerGetAll(getClient))
	srv.AddTool(tools.ListDnsServerGetById(), handlers.ListDnsServerGetById(getClient))
	srv.AddTool(tools.ListDnsServerGetByName(), handlers.ListDnsServerGetByName(getClient))
	srv.AddTool(tools.ListDnsServerGetTotal(), handlers.ListDnsServerGetTotal(getClient))

	// Add Global CIDR list tools
	srv.AddTool(tools.ListGlobalCidrGetAll(), handlers.ListGlobalCidrGetAll(getClient))
	srv.AddTool(tools.ListGlobalCidrGetById(), handlers.ListGlobalCidrGetById(getClient))
	srv.AddTool(tools.ListGlobalCidrGetByName(), handlers.ListGlobalCidrGetByName(getClient))
	srv.AddTool(tools.ListGlobalCidrGetTotal(), handlers.ListGlobalCidrGetTotal(getClient))

	// Add User Defined Route (UDR) list tools
	srv.AddTool(tools.ListUdrGetAll(), handlers.ListUdrGetAll(getClient))
	srv.AddTool(tools.ListUdrGetById(), handlers.ListUdrGetById(getClient))
	srv.AddTool(tools.ListUdrGetByName(), handlers.ListUdrGetByName(getClient))
	srv.AddTool(tools.ListUdrGetTotal(), handlers.ListUdrGetTotal(getClient))

	// Add Policy Prefix list tools
	srv.AddTool(tools.ListPrefixGetById(), handlers.ListPrefixGetById(getClient))
	srv.AddTool(tools.ListPrefixGetByName(), handlers.ListPrefixGetByName(getClient))
	srv.AddTool(tools.ListPrefixGetByPrefix(), handlers.ListPrefixGetByPrefix(getClient))

	// Add Policy FQDN list tools
	srv.AddTool(tools.ListPolicyFqdnGetAll(), handlers.ListPolicyFqdnGetAll(getClient))
	srv.AddTool(tools.ListPolicyFqdnGetById(), handlers.ListPolicyFqdnGetById(getClient))
	srv.AddTool(tools.ListPolicyFqdnGetByName(), handlers.ListPolicyFqdnGetByName(getClient))
	srv.AddTool(tools.ListPolicyFqdnGetTotal(), handlers.ListPolicyFqdnGetTotal(getClient))

	// Add monitoring tools
	srv.AddTool(tools.Alerts(), handlers.Alerts(getClient))
	srv.AddTool(tools.AuditLogs(), handlers.AuditLogs(getClient))
	srv.AddTool(tools.Jobs(), handlers.Jobs(getClient))

	srv.AddTool(tools.TenantResourceUsages(), handlers.TenantResourceUsages(getClient))
	srv.AddTool(tools.TenantResourceLimits(), handlers.TenantResourceLimits(getClient))

	// Add health tools
	srv.AddTool(tools.HealthConnectorGetById(), handlers.HealthConnectorGetById(getClient))
	srv.AddTool(tools.HealthConnectorInstanceGetById(), handlers.HealthConnectorInstanceGetById(getClient))
	srv.AddTool(tools.HealthServiceGetById(), handlers.HealthServiceGetById(getClient))
	srv.AddTool(tools.HealthServiceInstanceGetById(), handlers.HealthServiceInstanceGetById(getClient))

	// Start server based on mode
	switch mode {
	case "stdio":
		logf("INFO", "Starting server in stdio mode...")
		if err := server.ServeStdio(srv); err != nil {
			logf("INFO", "Failed to start stdio server: %v\n", err)
			return
		}
	case "sse":
		logf("INFO", "Starting server in SSE mode on port %s...\n", port)
		sse := server.NewSSEServer(srv)
		if err := sse.Start(":" + port); err != nil {
			logf("ERROR", "Failed to start SSE server: %v\n", err)
			return
		}
		logf("INFO", "SSE server started on port %s\n", port)
	default:
		logf("ERROR", "Unknown server mode: %s. Use 'stdio' or 'sse'.\n", mode)
		return
	}

	logf("INFO", "👋 Server stopped")
}
