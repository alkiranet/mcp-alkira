package main

import (
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

	// Create MCP server
	srv := server.NewMCPServer(
		"mcp-alkira",
		"0.2.0",
	)

	// Prompts
	srv.AddPrompt(prompts.Summary(), prompts.SummaryHandler())

	// Add tenant network basic tools
	srv.AddTool(tools.TenantNetworkSummary(), handlers.TenantNetworkSummary(alkiraClient))
	srv.AddTool(tools.TenantNetworkFirewallZones(), handlers.TenantNetworkFirewallZones(alkiraClient))

	// Add billing tag tools
	srv.AddTool(tools.BillingTagGetAll(), handlers.BillingTagGetAll(alkiraClient))
	srv.AddTool(tools.BillingTagGetById(), handlers.BillingTagGetById(alkiraClient))
	srv.AddTool(tools.BillingTagGetByName(), handlers.BillingTagGetByName(alkiraClient))
	srv.AddTool(tools.BillingTagGetTotal(), handlers.BillingTagGetTotal(alkiraClient))

	srv.AddTool(tools.CxpGetAll(), handlers.CxpGetAll(alkiraClient))

	// Add segment tools
	srv.AddTool(tools.SegmentGetAll(), handlers.SegmentGetAll(alkiraClient))
	srv.AddTool(tools.SegmentGetById(), handlers.SegmentGetById(alkiraClient))
	srv.AddTool(tools.SegmentGetByName(), handlers.SegmentGetByName(alkiraClient))

	// Add group tools
	srv.AddTool(tools.GroupGetById(), handlers.GroupGetById(alkiraClient))
	srv.AddTool(tools.GroupGetByName(), handlers.GroupGetByName(alkiraClient))
	srv.AddTool(tools.GroupGetTotal(), handlers.GroupGetTotal(alkiraClient))

	// Add BYOIP tools
	srv.AddTool(tools.ByoipGetAll(), handlers.ByoipGetAll(alkiraClient))
	srv.AddTool(tools.ByoipGetById(), handlers.ByoipGetById(alkiraClient))
	srv.AddTool(tools.ByoipGetByName(), handlers.ByoipGetByName(alkiraClient))
	srv.AddTool(tools.ByoipGetTotal(), handlers.ByoipGetTotal(alkiraClient))

	// Add segment resource tools
	srv.AddTool(tools.SegmentResourceGetAll(), handlers.SegmentResourceGetAll(alkiraClient))
	srv.AddTool(tools.SegmentResourceGetById(), handlers.SegmentResourceGetById(alkiraClient))
	srv.AddTool(tools.SegmentResourceGetByName(), handlers.SegmentResourceGetByName(alkiraClient))
	srv.AddTool(tools.SegmentResourceGetTotal(), handlers.SegmentResourceGetTotal(alkiraClient))

	// Add segment resource share tools
	srv.AddTool(tools.SegmentResourceShareGetAll(), handlers.SegmentResourceShareGetAll(alkiraClient))
	srv.AddTool(tools.SegmentResourceShareGetById(), handlers.SegmentResourceShareGetById(alkiraClient))
	srv.AddTool(tools.SegmentResourceShareGetByName(), handlers.SegmentResourceShareGetByName(alkiraClient))
	srv.AddTool(tools.SegmentResourceShareGetTotal(), handlers.SegmentResourceShareGetTotal(alkiraClient))

	// Add Check Point firewall service tools
	srv.AddTool(tools.ServiceCheckPointGetAll(), handlers.ServiceCheckPointGetAll(alkiraClient))
	srv.AddTool(tools.ServiceCheckPointGetById(), handlers.ServiceCheckPointGetById(alkiraClient))
	srv.AddTool(tools.ServiceCheckPointGetByName(), handlers.ServiceCheckPointGetByName(alkiraClient))
	srv.AddTool(tools.ServiceCheckPointGetTotal(), handlers.ServiceCheckPointGetTotal(alkiraClient))

	// Add Cisco FTDV service tools
	srv.AddTool(tools.ServiceCiscoFTDvGetAll(), handlers.ServiceCiscoFTDvGetAll(alkiraClient))
	srv.AddTool(tools.ServiceCiscoFTDvGetById(), handlers.ServiceCiscoFTDvGetById(alkiraClient))
	srv.AddTool(tools.ServiceCiscoFTDvGetByName(), handlers.ServiceCiscoFTDvGetByName(alkiraClient))
	srv.AddTool(tools.ServiceCiscoFTDvGetTotal(), handlers.ServiceCiscoFTDvGetTotal(alkiraClient))

	// Add F5 LB service tools
	srv.AddTool(tools.ServiceF5LBGetAll(), handlers.ServiceF5LBGetAll(alkiraClient))
	srv.AddTool(tools.ServiceF5LBGetById(), handlers.ServiceF5LBGetById(alkiraClient))
	srv.AddTool(tools.ServiceF5LBGetByName(), handlers.ServiceF5LBGetByName(alkiraClient))
	srv.AddTool(tools.ServiceF5LBGetTotal(), handlers.ServiceF5LBGetTotal(alkiraClient))

	// Add Fortinet service tools
	srv.AddTool(tools.ServiceFortinetGetAll(), handlers.ServiceFortinetGetAll(alkiraClient))
	srv.AddTool(tools.ServiceFortinetGetById(), handlers.ServiceFortinetGetById(alkiraClient))
	srv.AddTool(tools.ServiceFortinetGetByName(), handlers.ServiceFortinetGetByName(alkiraClient))
	srv.AddTool(tools.ServiceFortinetGetTotal(), handlers.ServiceFortinetGetTotal(alkiraClient))

	// Add Infoblox service tools
	srv.AddTool(tools.ServiceInfobloxGetAll(), handlers.ServiceInfobloxGetAll(alkiraClient))
	srv.AddTool(tools.ServiceInfobloxGetById(), handlers.ServiceInfobloxGetById(alkiraClient))
	srv.AddTool(tools.ServiceInfobloxGetByName(), handlers.ServiceInfobloxGetByName(alkiraClient))
	srv.AddTool(tools.ServiceInfobloxGetTotal(), handlers.ServiceInfobloxGetTotal(alkiraClient))

	// Add PAN firewall service tools
	srv.AddTool(tools.ServicePanGetAll(), handlers.ServicePanGetAll(alkiraClient))
	srv.AddTool(tools.ServicePanGetById(), handlers.ServicePanGetById(alkiraClient))
	srv.AddTool(tools.ServicePanGetByName(), handlers.ServicePanGetByName(alkiraClient))
	srv.AddTool(tools.ServicePanGetTotal(), handlers.ServicePanGetTotal(alkiraClient))

	// Add ZScaler service tools
	srv.AddTool(tools.ServiceZscalerGetAll(), handlers.ServiceZscalerGetAll(alkiraClient))
	srv.AddTool(tools.ServiceZscalerGetById(), handlers.ServiceZscalerGetById(alkiraClient))
	srv.AddTool(tools.ServiceZscalerGetByName(), handlers.ServiceZscalerGetByName(alkiraClient))
	srv.AddTool(tools.ServiceZscalerGetTotal(), handlers.ServiceZscalerGetTotal(alkiraClient))

	// Add Aruba Edge Connect SD-WAN connector tools
	srv.AddTool(tools.ConnectorArubaEdgeGetAll(), handlers.ConnectorArubaEdgeGetAll(alkiraClient))
	srv.AddTool(tools.ConnectorArubaEdgeGetById(), handlers.ConnectorArubaEdgeGetById(alkiraClient))
	srv.AddTool(tools.ConnectorArubaEdgeGetByName(), handlers.ConnectorArubaEdgeGetByName(alkiraClient))
	srv.AddTool(tools.ConnectorArubaEdgeGetTotal(), handlers.ConnectorArubaEdgeGetTotal(alkiraClient))

	// Add AWS Direct Connect connector tools
	srv.AddTool(tools.ConnectorAwsDirectConnectGetAll(), handlers.ConnectorAwsDirectConnectGetAll(alkiraClient))
	srv.AddTool(tools.ConnectorAwsDirectConnectGetById(), handlers.ConnectorAwsDirectConnectGetById(alkiraClient))
	srv.AddTool(tools.ConnectorAwsDirectConnectGetByName(), handlers.ConnectorAwsDirectConnectGetByName(alkiraClient))
	srv.AddTool(tools.ConnectorAwsDirectConnectGetTotal(), handlers.ConnectorAwsDirectConnectGetTotal(alkiraClient))

	// Add AWS Transit Gateway connector tools
	srv.AddTool(tools.ConnectorAwsTgwGetAll(), handlers.ConnectorAwsTgwGetAll(alkiraClient))
	srv.AddTool(tools.ConnectorAwsTgwGetById(), handlers.ConnectorAwsTgwGetById(alkiraClient))
	srv.AddTool(tools.ConnectorAwsTgwGetByName(), handlers.ConnectorAwsTgwGetByName(alkiraClient))
	srv.AddTool(tools.ConnectorAwsTgwGetTotal(), handlers.ConnectorAwsTgwGetTotal(alkiraClient))

	// Add Azure ExpressRoute connector tools
	srv.AddTool(tools.ConnectorAzureExpressRouteGetAll(), handlers.ConnectorAzureExpressRouteGetAll(alkiraClient))
	srv.AddTool(tools.ConnectorAzureExpressRouteGetById(), handlers.ConnectorAzureExpressRouteGetById(alkiraClient))
	srv.AddTool(tools.ConnectorAzureExpressRouteGetByName(), handlers.ConnectorAzureExpressRouteGetByName(alkiraClient))
	srv.AddTool(tools.ConnectorAzureExpressRouteGetTotal(), handlers.ConnectorAzureExpressRouteGetTotal(alkiraClient))

	// Add Cisco SD-WAN connector tools
	srv.AddTool(tools.ConnectorCiscoSdwanGetAll(), handlers.ConnectorCiscoSdwanGetAll(alkiraClient))
	srv.AddTool(tools.ConnectorCiscoSdwanGetById(), handlers.ConnectorCiscoSdwanGetById(alkiraClient))
	srv.AddTool(tools.ConnectorCiscoSdwanGetByName(), handlers.ConnectorCiscoSdwanGetByName(alkiraClient))
	srv.AddTool(tools.ConnectorCiscoSdwanGetTotal(), handlers.ConnectorCiscoSdwanGetTotal(alkiraClient))

	// Add Fortinet SD-WAN connector tools
	srv.AddTool(tools.ConnectorFortinetSdwanGetAll(), handlers.ConnectorFortinetSdwanGetAll(alkiraClient))
	srv.AddTool(tools.ConnectorFortinetSdwanGetById(), handlers.ConnectorFortinetSdwanGetById(alkiraClient))
	srv.AddTool(tools.ConnectorFortinetSdwanGetByName(), handlers.ConnectorFortinetSdwanGetByName(alkiraClient))
	srv.AddTool(tools.ConnectorFortinetSdwanGetTotal(), handlers.ConnectorFortinetSdwanGetTotal(alkiraClient))

	// Add Google Cloud Interconnect connector tools
	srv.AddTool(tools.ConnectorGcpInterconnectGetAll(), handlers.ConnectorGcpInterconnectGetAll(alkiraClient))
	srv.AddTool(tools.ConnectorGcpInterconnectGetById(), handlers.ConnectorGcpInterconnectGetById(alkiraClient))
	srv.AddTool(tools.ConnectorGcpInterconnectGetByName(), handlers.ConnectorGcpInterconnectGetByName(alkiraClient))
	srv.AddTool(tools.ConnectorGcpInterconnectGetTotal(), handlers.ConnectorGcpInterconnectGetTotal(alkiraClient))

	// Add Google Cloud VPC connector tools
	srv.AddTool(tools.ConnectorGcpVpcGetAll(), handlers.ConnectorGcpVpcGetAll(alkiraClient))
	srv.AddTool(tools.ConnectorGcpVpcGetById(), handlers.ConnectorGcpVpcGetById(alkiraClient))
	srv.AddTool(tools.ConnectorGcpVpcGetByName(), handlers.ConnectorGcpVpcGetByName(alkiraClient))
	srv.AddTool(tools.ConnectorGcpVpcGetTotal(), handlers.ConnectorGcpVpcGetTotal(alkiraClient))

	// Add Oracle Cloud Infrastructure VCN connector tools
	srv.AddTool(tools.ConnectorOciVcnGetAll(), handlers.ConnectorOciVcnGetAll(alkiraClient))
	srv.AddTool(tools.ConnectorOciVcnGetById(), handlers.ConnectorOciVcnGetById(alkiraClient))
	srv.AddTool(tools.ConnectorOciVcnGetByName(), handlers.ConnectorOciVcnGetByName(alkiraClient))
	srv.AddTool(tools.ConnectorOciVcnGetTotal(), handlers.ConnectorOciVcnGetTotal(alkiraClient))

	// Add Remote Access Template connector tools
	srv.AddTool(tools.ConnectorRemoteAccessGetAll(), handlers.ConnectorRemoteAccessGetAll(alkiraClient))
	srv.AddTool(tools.ConnectorRemoteAccessGetById(), handlers.ConnectorRemoteAccessGetById(alkiraClient))
	srv.AddTool(tools.ConnectorRemoteAccessGetByName(), handlers.ConnectorRemoteAccessGetByName(alkiraClient))
	srv.AddTool(tools.ConnectorRemoteAccessGetTotal(), handlers.ConnectorRemoteAccessGetTotal(alkiraClient))

	// Add Versa SD-WAN connector tools
	srv.AddTool(tools.ConnectorVersaSdwanGetAll(), handlers.ConnectorVersaSdwanGetAll(alkiraClient))
	srv.AddTool(tools.ConnectorVersaSdwanGetById(), handlers.ConnectorVersaSdwanGetById(alkiraClient))
	srv.AddTool(tools.ConnectorVersaSdwanGetByName(), handlers.ConnectorVersaSdwanGetByName(alkiraClient))
	srv.AddTool(tools.ConnectorVersaSdwanGetTotal(), handlers.ConnectorVersaSdwanGetTotal(alkiraClient))

	// Add VMware SD-WAN (VeloCloud) connector tools
	srv.AddTool(tools.ConnectorVmwareSdwanGetAll(), handlers.ConnectorVmwareSdwanGetAll(alkiraClient))
	srv.AddTool(tools.ConnectorVmwareSdwanGetById(), handlers.ConnectorVmwareSdwanGetById(alkiraClient))
	srv.AddTool(tools.ConnectorVmwareSdwanGetByName(), handlers.ConnectorVmwareSdwanGetByName(alkiraClient))
	srv.AddTool(tools.ConnectorVmwareSdwanGetTotal(), handlers.ConnectorVmwareSdwanGetTotal(alkiraClient))

	// Add internet connector tools
	srv.AddTool(tools.ConnectorInternetGetAll(), handlers.ConnectorInternetGetAll(alkiraClient))
	srv.AddTool(tools.ConnectorInternetGetById(), handlers.ConnectorInternetGetById(alkiraClient))
	srv.AddTool(tools.ConnectorInternetGetByName(), handlers.ConnectorInternetGetByName(alkiraClient))
	srv.AddTool(tools.ConnectorInternetGetTotal(), handlers.ConnectorInternetGetTotal(alkiraClient))

	// Add AWS VPC connector tools
	srv.AddTool(tools.ConnectorAwsVpcGetAll(), handlers.ConnectorAwsVpcGetAll(alkiraClient))
	srv.AddTool(tools.ConnectorAwsVpcGetById(), handlers.ConnectorAwsVpcGetById(alkiraClient))
	srv.AddTool(tools.ConnectorAwsVpcGetByName(), handlers.ConnectorAwsVpcGetByName(alkiraClient))
	srv.AddTool(tools.ConnectorAwsVpcGetTotal(), handlers.ConnectorAwsVpcGetTotal(alkiraClient))

	// Add Azure VNET connector tools
	srv.AddTool(tools.ConnectorAzureVnetGetAll(), handlers.ConnectorAzureVnetGetAll(alkiraClient))
	srv.AddTool(tools.ConnectorAzureVnetGetById(), handlers.ConnectorAzureVnetGetById(alkiraClient))
	srv.AddTool(tools.ConnectorAzureVnetGetByName(), handlers.ConnectorAzureVnetGetByName(alkiraClient))
	srv.AddTool(tools.ConnectorAzureVnetGetTotal(), handlers.ConnectorAzureVnetGetTotal(alkiraClient))

	// Add IPSec connector tools
	srv.AddTool(tools.ConnectorIPSecGetAll(), handlers.ConnectorIPSecGetAll(alkiraClient))
	srv.AddTool(tools.ConnectorIPSecGetById(), handlers.ConnectorIPSecGetById(alkiraClient))
	srv.AddTool(tools.ConnectorIPSecGetByName(), handlers.ConnectorIPSecGetByName(alkiraClient))
	srv.AddTool(tools.ConnectorIPSecGetTotal(), handlers.ConnectorIPSecGetTotal(alkiraClient))

	// Add IPSec Adv connector tools
	srv.AddTool(tools.ConnectorIPSecAdvGetAll(), handlers.ConnectorIPSecAdvGetAll(alkiraClient))
	srv.AddTool(tools.ConnectorIPSecAdvGetById(), handlers.ConnectorIPSecAdvGetById(alkiraClient))
	srv.AddTool(tools.ConnectorIPSecAdvGetByName(), handlers.ConnectorIPSecAdvGetByName(alkiraClient))
	srv.AddTool(tools.ConnectorIPSecAdvGetTotal(), handlers.ConnectorIPSecAdvGetTotal(alkiraClient))

	srv.AddTool(tools.ConnectorIPSecTunnelProfileGetAll(), handlers.ConnectorIPSecTunnelProfileGetAll(alkiraClient))

	// Add internet application tools
	srv.AddTool(tools.InternetApplicationGetAll(), handlers.InternetApplicationGetAll(alkiraClient))
	srv.AddTool(tools.InternetApplicationGetById(), handlers.InternetApplicationGetById(alkiraClient))
	srv.AddTool(tools.InternetApplicationGetByName(), handlers.InternetApplicationGetByName(alkiraClient))
	srv.AddTool(tools.InternetApplicationGetTotal(), handlers.InternetApplicationGetTotal(alkiraClient))

	// Add routes tools
	srv.AddTool(tools.GetRoutes(), handlers.RouteGet(alkiraClient))
	srv.AddTool(tools.GetRouteCount(), handlers.RouteGetCount(alkiraClient))
	srv.AddTool(tools.GetRouteSummary(), handlers.RouteSummaryGet(alkiraClient))
	srv.AddTool(tools.GetAllRoutes(), handlers.AllRouteGet(alkiraClient))

	// Add NAT policy tools
	srv.AddTool(tools.PolicyNatGetAll(), handlers.PolicyNatGetAll(alkiraClient))
	srv.AddTool(tools.PolicyNatGetById(), handlers.PolicyNatGetById(alkiraClient))
	srv.AddTool(tools.PolicyNatGetByName(), handlers.PolicyNatGetByName(alkiraClient))
	srv.AddTool(tools.PolicyNatGetTotal(), handlers.PolicyNatGetTotal(alkiraClient))

	srv.AddTool(tools.PolicyNatRuleGetAll(), handlers.PolicyNatRuleGetAll(alkiraClient))
	srv.AddTool(tools.PolicyNatRuleGetById(), handlers.PolicyNatRuleGetById(alkiraClient))
	srv.AddTool(tools.PolicyNatRuleGetByName(), handlers.PolicyNatRuleGetByName(alkiraClient))
	srv.AddTool(tools.PolicyNatRuleGetTotal(), handlers.PolicyNatRuleGetTotal(alkiraClient))

	// Add Route policy tools
	srv.AddTool(tools.PolicyRouteGetAll(), handlers.PolicyRouteGetAll(alkiraClient))
	srv.AddTool(tools.PolicyRouteGetById(), handlers.PolicyRouteGetById(alkiraClient))
	srv.AddTool(tools.PolicyRouteGetByName(), handlers.PolicyRouteGetByName(alkiraClient))
	srv.AddTool(tools.PolicyRouteGetTotal(), handlers.PolicyRouteGetTotal(alkiraClient))

	// Add Traffic policy tools
	srv.AddTool(tools.PolicyTrafficGetAll(), handlers.PolicyTrafficGetAll(alkiraClient))
	srv.AddTool(tools.PolicyTrafficGetById(), handlers.PolicyTrafficGetById(alkiraClient))
	srv.AddTool(tools.PolicyTrafficGetByName(), handlers.PolicyTrafficGetByName(alkiraClient))

	srv.AddTool(tools.PolicyTrafficRuleGetAll(), handlers.PolicyTrafficRuleGetAll(alkiraClient))
	srv.AddTool(tools.PolicyTrafficRuleGetById(), handlers.PolicyTrafficRuleGetById(alkiraClient))
	srv.AddTool(tools.PolicyTrafficRuleGetByName(), handlers.PolicyTrafficRuleGetByName(alkiraClient))

	// Add traffic policy rule list tools
	srv.AddTool(tools.PolicyTrafficRuleListGetAll(), handlers.PolicyTrafficRuleListGetAll(alkiraClient))
	srv.AddTool(tools.PolicyTrafficRuleListGetById(), handlers.PolicyTrafficRuleListGetById(alkiraClient))
	srv.AddTool(tools.PolicyTrafficRuleListGetByName(), handlers.PolicyTrafficRuleListGetByName(alkiraClient))

	// Add AS Path list tools
	srv.AddTool(tools.ListAsPathGetAll(), handlers.ListAsPathGetAll(alkiraClient))
	srv.AddTool(tools.ListAsPathGetById(), handlers.ListAsPathGetById(alkiraClient))
	srv.AddTool(tools.ListAsPathGetByName(), handlers.ListAsPathGetByName(alkiraClient))
	srv.AddTool(tools.ListAsPathGetTotal(), handlers.ListAsPathGetTotal(alkiraClient))

	// Add BGP Community list tools
	srv.AddTool(tools.ListCommunityGetAll(), handlers.ListCommunityGetAll(alkiraClient))
	srv.AddTool(tools.ListCommunityGetById(), handlers.ListCommunityGetById(alkiraClient))
	srv.AddTool(tools.ListCommunityGetByName(), handlers.ListCommunityGetByName(alkiraClient))
	srv.AddTool(tools.ListCommunityGetTotal(), handlers.ListCommunityGetTotal(alkiraClient))

	// Add BGP Extended Community list tools
	srv.AddTool(tools.ListExtendedCommunityGetAll(), handlers.ListExtendedCommunityGetAll(alkiraClient))
	srv.AddTool(tools.ListExtendedCommunityGetById(), handlers.ListExtendedCommunityGetById(alkiraClient))
	srv.AddTool(tools.ListExtendedCommunityGetByName(), handlers.ListExtendedCommunityGetByName(alkiraClient))
	srv.AddTool(tools.ListExtendedCommunityGetTotal(), handlers.ListExtendedCommunityGetTotal(alkiraClient))

	// Add DNS Server list tools
	srv.AddTool(tools.ListDnsServerGetAll(), handlers.ListDnsServerGetAll(alkiraClient))
	srv.AddTool(tools.ListDnsServerGetById(), handlers.ListDnsServerGetById(alkiraClient))
	srv.AddTool(tools.ListDnsServerGetByName(), handlers.ListDnsServerGetByName(alkiraClient))
	srv.AddTool(tools.ListDnsServerGetTotal(), handlers.ListDnsServerGetTotal(alkiraClient))

	// Add Global CIDR list tools
	srv.AddTool(tools.ListGlobalCidrGetAll(), handlers.ListGlobalCidrGetAll(alkiraClient))
	srv.AddTool(tools.ListGlobalCidrGetById(), handlers.ListGlobalCidrGetById(alkiraClient))
	srv.AddTool(tools.ListGlobalCidrGetByName(), handlers.ListGlobalCidrGetByName(alkiraClient))
	srv.AddTool(tools.ListGlobalCidrGetTotal(), handlers.ListGlobalCidrGetTotal(alkiraClient))

	// Add User Defined Route (UDR) list tools
	srv.AddTool(tools.ListUdrGetAll(), handlers.ListUdrGetAll(alkiraClient))
	srv.AddTool(tools.ListUdrGetById(), handlers.ListUdrGetById(alkiraClient))
	srv.AddTool(tools.ListUdrGetByName(), handlers.ListUdrGetByName(alkiraClient))
	srv.AddTool(tools.ListUdrGetTotal(), handlers.ListUdrGetTotal(alkiraClient))

	// Add Policy Prefix list tools
	srv.AddTool(tools.ListPrefixGetById(), handlers.ListPrefixGetById(alkiraClient))
	srv.AddTool(tools.ListPrefixGetByName(), handlers.ListPrefixGetByName(alkiraClient))
	srv.AddTool(tools.ListPrefixGetByPrefix(), handlers.ListPrefixGetByPrefix(alkiraClient))

	// Add Policy FQDN list tools
	srv.AddTool(tools.ListPolicyFqdnGetAll(), handlers.ListPolicyFqdnGetAll(alkiraClient))
	srv.AddTool(tools.ListPolicyFqdnGetById(), handlers.ListPolicyFqdnGetById(alkiraClient))
	srv.AddTool(tools.ListPolicyFqdnGetByName(), handlers.ListPolicyFqdnGetByName(alkiraClient))
	srv.AddTool(tools.ListPolicyFqdnGetTotal(), handlers.ListPolicyFqdnGetTotal(alkiraClient))

	// Add monitoring tools
	srv.AddTool(tools.Alerts(), handlers.Alerts(alkiraClient))
	srv.AddTool(tools.AuditLogs(), handlers.AuditLogs(alkiraClient))
	srv.AddTool(tools.Jobs(), handlers.Jobs(alkiraClient))

	srv.AddTool(tools.TenantResourceUsages(), handlers.TenantResourceUsages(alkiraClient))
	srv.AddTool(tools.TenantResourceLimits(), handlers.TenantResourceLimits(alkiraClient))

	// Add health tools
	srv.AddTool(tools.HealthConnectorGetById(), handlers.HealthConnectorGetById(alkiraClient))
	srv.AddTool(tools.HealthConnectorInstanceGetById(), handlers.HealthConnectorInstanceGetById(alkiraClient))
	srv.AddTool(tools.HealthServiceGetById(), handlers.HealthServiceGetById(alkiraClient))
	srv.AddTool(tools.HealthServiceInstanceGetById(), handlers.HealthServiceInstanceGetById(alkiraClient))

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
