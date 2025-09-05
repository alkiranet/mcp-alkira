package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/alkiranet/mcp-alkira/handlers"
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
	)

	if err != nil {
		logf("ERROR", "failed to initialize alkira client, please check your credential and portal URI.")
		return
	}

	// Create MCP server
	srv := server.NewMCPServer(
		"mcp-alkira",
		"0.1.0",
	)

	// Add tenant network basic tools
	srv.AddTool(tools.GetTenantNetworkSummary(), handlers.GetTenantNetworkSummary(alkiraClient))

	// Add billing tag tools
	srv.AddTool(tools.GetBillingTags(), handlers.GetBillingTags(alkiraClient))
	srv.AddTool(tools.GetBillingTagById(), handlers.GetBillingTagById(alkiraClient))
	srv.AddTool(tools.GetBillingTagByName(), handlers.GetBillingTagByName(alkiraClient))

	srv.AddTool(tools.GetCxps(), handlers.GetCxps(alkiraClient))

	// Add segment tools
	srv.AddTool(tools.GetSegments(), handlers.GetSegments(alkiraClient))
	srv.AddTool(tools.GetSegmentById(), handlers.GetSegmentById(alkiraClient))
	srv.AddTool(tools.GetSegmentByName(), handlers.GetSegmentByName(alkiraClient))

	// Add group tools
	srv.AddTool(tools.GetGroups(), handlers.GetGroups(alkiraClient))
	srv.AddTool(tools.GetGroupById(), handlers.GetGroupById(alkiraClient))
	srv.AddTool(tools.GetGroupByName(), handlers.GetGroupByName(alkiraClient))

	// Add segment resource tools
	srv.AddTool(tools.GetSegmentResources(), handlers.GetSegmentResources(alkiraClient))
	srv.AddTool(tools.GetSegmentResourceById(), handlers.GetSegmentResourceById(alkiraClient))
	srv.AddTool(tools.GetSegmentResourceByName(), handlers.GetSegmentResourceByName(alkiraClient))

	// Add segment resource share tools
	srv.AddTool(tools.GetSegmentResourceShares(), handlers.GetSegmentResourceShares(alkiraClient))
	srv.AddTool(tools.GetSegmentResourceShareById(), handlers.GetSegmentResourceShareById(alkiraClient))
	srv.AddTool(tools.GetSegmentResourceShareByName(), handlers.GetSegmentResourceShareByName(alkiraClient))

	// Add Check Point firewall service tools
	srv.AddTool(tools.GetCheckPointServices(), handlers.GetCheckPointServices(alkiraClient))
	srv.AddTool(tools.GetCheckPointServiceById(), handlers.GetCheckPointServiceById(alkiraClient))
	srv.AddTool(tools.GetCheckPointServiceByName(), handlers.GetCheckPointServiceByName(alkiraClient))

	// Add Cisco FTDV service tools
	srv.AddTool(tools.GetCiscoFTDvServices(), handlers.GetCiscoFTDvServices(alkiraClient))
	srv.AddTool(tools.GetCiscoFTDvServiceById(), handlers.GetCiscoFTDvServiceById(alkiraClient))
	srv.AddTool(tools.GetCiscoFTDvServiceByName(), handlers.GetCiscoFTDvServiceByName(alkiraClient))

	// Add F5 LB service tools
	srv.AddTool(tools.GetF5LbServices(), handlers.GetF5LbServices(alkiraClient))
	srv.AddTool(tools.GetF5LbServiceById(), handlers.GetF5LbServiceById(alkiraClient))
	srv.AddTool(tools.GetF5LbServiceByName(), handlers.GetF5LbServiceByName(alkiraClient))

	// Add Fortinet service tools
	srv.AddTool(tools.GetFortinetServices(), handlers.GetFortinetServices(alkiraClient))
	srv.AddTool(tools.GetFortinetServiceById(), handlers.GetFortinetServiceById(alkiraClient))
	srv.AddTool(tools.GetFortinetServiceByName(), handlers.GetFortinetServiceByName(alkiraClient))

	// Add Infoblox service tools
	srv.AddTool(tools.GetInfobloxServices(), handlers.GetInfobloxServices(alkiraClient))
	srv.AddTool(tools.GetInfobloxServiceById(), handlers.GetInfobloxServiceById(alkiraClient))
	srv.AddTool(tools.GetInfobloxServiceByName(), handlers.GetInfobloxServiceByName(alkiraClient))

	// Add PAN firewall service tools
	srv.AddTool(tools.GetPanServices(), handlers.GetPanServices(alkiraClient))
	srv.AddTool(tools.GetPanServiceById(), handlers.GetPanServiceById(alkiraClient))
	srv.AddTool(tools.GetPanServiceByName(), handlers.GetPanServiceByName(alkiraClient))

	// Add ZScaler service tools
	srv.AddTool(tools.GetZscalerServices(), handlers.GetZscalerServices(alkiraClient))
	srv.AddTool(tools.GetZscalerServiceById(), handlers.GetZscalerServiceById(alkiraClient))
	srv.AddTool(tools.GetZscalerServiceByName(), handlers.GetZscalerServiceByName(alkiraClient))

	// Add Aruba Edge Connect SD-WAN connector tools
	srv.AddTool(tools.GetArubaEdgeConnectors(), handlers.GetArubaEdgeConnectors(alkiraClient))
	srv.AddTool(tools.GetArubaEdgeConnectorById(), handlers.GetArubaEdgeConnectorById(alkiraClient))
	srv.AddTool(tools.GetArubaEdgeConnectorByName(), handlers.GetArubaEdgeConnectorByName(alkiraClient))

	// Add AWS Direct Connect connector tools
	srv.AddTool(tools.GetAwsDirectConnectConnectors(), handlers.GetAwsDirectConnectConnectors(alkiraClient))
	srv.AddTool(tools.GetAwsDirectConnectConnectorById(), handlers.GetAwsDirectConnectConnectorById(alkiraClient))
	srv.AddTool(tools.GetAwsDirectConnectConnectorByName(), handlers.GetAwsDirectConnectConnectorByName(alkiraClient))

	// Add AWS Transit Gateway connector tools
	srv.AddTool(tools.GetAwsTgwConnectors(), handlers.GetAwsTgwConnectors(alkiraClient))
	srv.AddTool(tools.GetAwsTgwConnectorById(), handlers.GetAwsTgwConnectorById(alkiraClient))
	srv.AddTool(tools.GetAwsTgwConnectorByName(), handlers.GetAwsTgwConnectorByName(alkiraClient))

	// Add Azure ExpressRoute connector tools
	srv.AddTool(tools.GetAzureExpressRouteConnectors(), handlers.GetAzureExpressRouteConnectors(alkiraClient))
	srv.AddTool(tools.GetAzureExpressRouteConnectorById(), handlers.GetAzureExpressRouteConnectorById(alkiraClient))
	srv.AddTool(tools.GetAzureExpressRouteConnectorByName(), handlers.GetAzureExpressRouteConnectorByName(alkiraClient))

	// Add Cisco SD-WAN connector tools
	srv.AddTool(tools.GetCiscoSdwanConnectors(), handlers.GetCiscoSdwanConnectors(alkiraClient))
	srv.AddTool(tools.GetCiscoSdwanConnectorById(), handlers.GetCiscoSdwanConnectorById(alkiraClient))
	srv.AddTool(tools.GetCiscoSdwanConnectorByName(), handlers.GetCiscoSdwanConnectorByName(alkiraClient))

	// Add Fortinet SD-WAN connector tools
	srv.AddTool(tools.GetFortinetSdwanConnectors(), handlers.GetFortinetSdwanConnectors(alkiraClient))
	srv.AddTool(tools.GetFortinetSdwanConnectorById(), handlers.GetFortinetSdwanConnectorById(alkiraClient))
	srv.AddTool(tools.GetFortinetSdwanConnectorByName(), handlers.GetFortinetSdwanConnectorByName(alkiraClient))

	// Add Google Cloud Interconnect connector tools
	srv.AddTool(tools.GetGcpInterconnectConnectors(), handlers.GetGcpInterconnectConnectors(alkiraClient))
	srv.AddTool(tools.GetGcpInterconnectConnectorById(), handlers.GetGcpInterconnectConnectorById(alkiraClient))
	srv.AddTool(tools.GetGcpInterconnectConnectorByName(), handlers.GetGcpInterconnectConnectorByName(alkiraClient))

	// Add Google Cloud VPC connector tools
	srv.AddTool(tools.GetGcpVpcConnectors(), handlers.GetGcpVpcConnectors(alkiraClient))
	srv.AddTool(tools.GetGcpVpcConnectorById(), handlers.GetGcpVpcConnectorById(alkiraClient))
	srv.AddTool(tools.GetGcpVpcConnectorByName(), handlers.GetGcpVpcConnectorByName(alkiraClient))

	// Add Oracle Cloud Infrastructure VCN connector tools
	srv.AddTool(tools.GetOciVcnConnectors(), handlers.GetOciVcnConnectors(alkiraClient))
	srv.AddTool(tools.GetOciVcnConnectorById(), handlers.GetOciVcnConnectorById(alkiraClient))
	srv.AddTool(tools.GetOciVcnConnectorByName(), handlers.GetOciVcnConnectorByName(alkiraClient))

	// Add Remote Access Template connector tools
	srv.AddTool(tools.GetRemoteAccessConnectors(), handlers.GetRemoteAccessConnectors(alkiraClient))
	srv.AddTool(tools.GetRemoteAccessConnectorById(), handlers.GetRemoteAccessConnectorById(alkiraClient))
	srv.AddTool(tools.GetRemoteAccessConnectorByName(), handlers.GetRemoteAccessConnectorByName(alkiraClient))

	// Add Versa SD-WAN connector tools
	srv.AddTool(tools.GetVersaSdwanConnectors(), handlers.GetVersaSdwanConnectors(alkiraClient))
	srv.AddTool(tools.GetVersaSdwanConnectorById(), handlers.GetVersaSdwanConnectorById(alkiraClient))
	srv.AddTool(tools.GetVersaSdwanConnectorByName(), handlers.GetVersaSdwanConnectorByName(alkiraClient))

	// Add VMware SD-WAN (VeloCloud) connector tools
	srv.AddTool(tools.GetVmwareSdwanConnectors(), handlers.GetVmwareSdwanConnectors(alkiraClient))
	srv.AddTool(tools.GetVmwareSdwanConnectorById(), handlers.GetVmwareSdwanConnectorById(alkiraClient))
	srv.AddTool(tools.GetVmwareSdwanConnectorByName(), handlers.GetVmwareSdwanConnectorByName(alkiraClient))

	// Add internet connector tools
	srv.AddTool(tools.GetInternetConnectors(), handlers.GetInternetConnectors(alkiraClient))
	srv.AddTool(tools.GetInternetConnectorById(), handlers.GetInternetConnectorById(alkiraClient))
	srv.AddTool(tools.GetInternetConnectorByName(), handlers.GetInternetConnectorByName(alkiraClient))

	// Add AWS VPC connector tools
	srv.AddTool(tools.GetAwsVpcConnectors(), handlers.GetAwsVpcConnectors(alkiraClient))
	srv.AddTool(tools.GetAwsVpcConnectorById(), handlers.GetAwsVpcConnectorById(alkiraClient))
	srv.AddTool(tools.GetAwsVpcConnectorByName(), handlers.GetAwsVpcConnectorByName(alkiraClient))

	// Add Azure VNET connector tools
	srv.AddTool(tools.GetAzureVnetConnectors(), handlers.GetAzureVnetConnectors(alkiraClient))
	srv.AddTool(tools.GetAzureVnetConnectorById(), handlers.GetAzureVnetConnectorById(alkiraClient))
	srv.AddTool(tools.GetAzureVnetConnectorByName(), handlers.GetAzureVnetConnectorByName(alkiraClient))

	// Add IPSec connector tools
	srv.AddTool(tools.GetIPSecConnectors(), handlers.GetIPSecConnectors(alkiraClient))
	srv.AddTool(tools.GetIPSecConnectorById(), handlers.GetIPSecConnectorById(alkiraClient))
	srv.AddTool(tools.GetIPSecConnectorByName(), handlers.GetIPSecConnectorByName(alkiraClient))

	// Add IPSec Adv connector tools
	srv.AddTool(tools.GetIPSecAdvConnectors(), handlers.GetIPSecAdvConnectors(alkiraClient))
	srv.AddTool(tools.GetIPSecAdvConnectorById(), handlers.GetIPSecAdvConnectorById(alkiraClient))
	srv.AddTool(tools.GetIPSecAdvConnectorByName(), handlers.GetIPSecAdvConnectorByName(alkiraClient))
	srv.AddTool(tools.GetIPSecConnectorTunnelProfile(), handlers.GetIPSecConnectorTunnelProfile(alkiraClient))

	// Add internet application tools
	srv.AddTool(tools.GetInternetApplications(), handlers.GetInternetApplications(alkiraClient))
	srv.AddTool(tools.GetInternetApplicationById(), handlers.GetInternetApplicationById(alkiraClient))
	srv.AddTool(tools.GetInternetApplicationByName(), handlers.GetInternetApplicationByName(alkiraClient))

	// Add routes tools
	srv.AddTool(tools.GetRoutes(), handlers.GetRoutes(alkiraClient))
	srv.AddTool(tools.GetRouteCount(), handlers.GetRouteCount(alkiraClient))
	srv.AddTool(tools.GetRouteSummary(), handlers.GetRouteSummary(alkiraClient))
	srv.AddTool(tools.GetAllRoutes(), handlers.GetAllRoutes(alkiraClient))

	// Add NAT policy tools
	srv.AddTool(tools.GetNatPolicies(), handlers.GetNatPolicies(alkiraClient))
	srv.AddTool(tools.GetNatPolicyById(), handlers.GetNatPolicyById(alkiraClient))
	srv.AddTool(tools.GetNatPolicyByName(), handlers.GetNatPolicyByName(alkiraClient))

	srv.AddTool(tools.GetNatPolicyRules(), handlers.GetNatPolicyRules(alkiraClient))
	srv.AddTool(tools.GetNatPolicyRuleById(), handlers.GetNatPolicyRuleById(alkiraClient))
	srv.AddTool(tools.GetNatPolicyRuleByName(), handlers.GetNatPolicyRuleByName(alkiraClient))

	// Add Route policy tools
	srv.AddTool(tools.GetRoutePolicies(), handlers.GetRoutePolicies(alkiraClient))
	srv.AddTool(tools.GetRoutePolicyById(), handlers.GetRoutePolicyById(alkiraClient))
	srv.AddTool(tools.GetRoutePolicyByName(), handlers.GetRoutePolicyByName(alkiraClient))

	// Add Traffic policy tools
	srv.AddTool(tools.GetTrafficPolicies(), handlers.GetTrafficPolicies(alkiraClient))
	srv.AddTool(tools.GetTrafficPolicyById(), handlers.GetTrafficPolicyById(alkiraClient))
	srv.AddTool(tools.GetTrafficPolicyByName(), handlers.GetTrafficPolicyByName(alkiraClient))

	srv.AddTool(tools.GetTrafficPolicyRules(), handlers.GetTrafficPolicyRules(alkiraClient))
	srv.AddTool(tools.GetTrafficPolicyRuleById(), handlers.GetTrafficPolicyRuleById(alkiraClient))
	srv.AddTool(tools.GetTrafficPolicyRuleByName(), handlers.GetTrafficPolicyRuleByName(alkiraClient))

	// Add traffic policy rule list tools
	srv.AddTool(tools.GetTrafficPolicyRuleLists(), handlers.GetTrafficPolicyRuleLists(alkiraClient))
	srv.AddTool(tools.GetTrafficPolicyRuleListById(), handlers.GetTrafficPolicyRuleListById(alkiraClient))
	srv.AddTool(tools.GetTrafficPolicyRuleListByName(), handlers.GetTrafficPolicyRuleListByName(alkiraClient))

	// Add AS Path list tools
	srv.AddTool(tools.GetAsPathLists(), handlers.GetAsPathLists(alkiraClient))
	srv.AddTool(tools.GetAsPathListById(), handlers.GetAsPathListById(alkiraClient))
	srv.AddTool(tools.GetAsPathListByName(), handlers.GetAsPathListByName(alkiraClient))

	// Add BGP Community list tools
	srv.AddTool(tools.GetCommunityLists(), handlers.GetCommunityLists(alkiraClient))
	srv.AddTool(tools.GetCommunityListById(), handlers.GetCommunityListById(alkiraClient))
	srv.AddTool(tools.GetCommunityListByName(), handlers.GetCommunityListByName(alkiraClient))

	// Add BGP Extended Community list tools
	srv.AddTool(tools.GetExtendedCommunityLists(), handlers.GetExtendedCommunityLists(alkiraClient))
	srv.AddTool(tools.GetExtendedCommunityListById(), handlers.GetExtendedCommunityListById(alkiraClient))
	srv.AddTool(tools.GetExtendedCommunityListByName(), handlers.GetExtendedCommunityListByName(alkiraClient))

	// Add DNS Server list tools
	srv.AddTool(tools.GetDnsServerLists(), handlers.GetDnsServerLists(alkiraClient))
	srv.AddTool(tools.GetDnsServerListById(), handlers.GetDnsServerListById(alkiraClient))
	srv.AddTool(tools.GetDnsServerListByName(), handlers.GetDnsServerListByName(alkiraClient))

	// Add Global CIDR list tools
	srv.AddTool(tools.GetGlobalCidrLists(), handlers.GetGlobalCidrLists(alkiraClient))
	srv.AddTool(tools.GetGlobalCidrListById(), handlers.GetGlobalCidrListById(alkiraClient))
	srv.AddTool(tools.GetGlobalCidrListByName(), handlers.GetGlobalCidrListByName(alkiraClient))

	// Add User Defined Route (UDR) list tools
	srv.AddTool(tools.GetUdrLists(), handlers.GetUdrLists(alkiraClient))
	srv.AddTool(tools.GetUdrListById(), handlers.GetUdrListById(alkiraClient))
	srv.AddTool(tools.GetUdrListByName(), handlers.GetUdrListByName(alkiraClient))

	// Add Policy Prefix list tools
	srv.AddTool(tools.GetPrefixListById(), handlers.GetPrefixListById(alkiraClient))
	srv.AddTool(tools.GetPrefixListByName(), handlers.GetPrefixListByName(alkiraClient))

	// Add Policy FQDN list tools
	srv.AddTool(tools.GetPolicyFqdnLists(), handlers.GetPolicyFqdnLists(alkiraClient))
	srv.AddTool(tools.GetPolicyFqdnListById(), handlers.GetPolicyFqdnListById(alkiraClient))
	srv.AddTool(tools.GetPolicyFqdnListByName(), handlers.GetPolicyFqdnListByName(alkiraClient))

	// Add monitoring tools
	srv.AddTool(tools.GetAlerts(), handlers.GetAlerts(alkiraClient))
	srv.AddTool(tools.GetAuditLogs(), handlers.GetAuditLogs(alkiraClient))
	srv.AddTool(tools.GetJobs(), handlers.GetJobs(alkiraClient))

	srv.AddTool(tools.GetResourceUsages(), handlers.GetResourceUsages(alkiraClient))
	srv.AddTool(tools.GetResourceLimits(), handlers.GetResourceLimits(alkiraClient))

	// Add health tools
	srv.AddTool(tools.GetHealths(), handlers.GetHealths(alkiraClient))
	srv.AddTool(tools.GetConnectorHealthById(), handlers.GetConnectorHealthById(alkiraClient))
	srv.AddTool(tools.GetConnectorInstanceHealthById(), handlers.GetConnectorInstanceHealthById(alkiraClient))
	srv.AddTool(tools.GetServiceHealthById(), handlers.GetServiceHealthById(alkiraClient))
	srv.AddTool(tools.GetServiceInstanceHealthById(), handlers.GetServiceInstanceHealthById(alkiraClient))

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
