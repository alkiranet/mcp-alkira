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
	srv.AddTool(tools.GetBillingTags(), handlers.GetBillingTags(alkiraClient))
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
	srv.AddTool(tools.GetAllSegmentResources(), handlers.GetAllSegmentResources(alkiraClient))
	srv.AddTool(tools.GetAllSegmentResourceShares(), handlers.GetAllSegmentResourceShares(alkiraClient))

	// Add service tools
	srv.AddTool(tools.GetAllServiceCheckpoint(), handlers.GetAllServiceCheckpoint(alkiraClient))
	srv.AddTool(tools.GetAllServiceCiscoFTDv(), handlers.GetAllServiceCiscoFTDv(alkiraClient))
	srv.AddTool(tools.GetAllServiceF5Lb(), handlers.GetAllServiceF5Lb(alkiraClient))
	srv.AddTool(tools.GetAllServiceFortinet(), handlers.GetAllServiceFortinet(alkiraClient))
	srv.AddTool(tools.GetAllServiceInfoblox(), handlers.GetAllServiceInfoblox(alkiraClient))
	srv.AddTool(tools.GetAllServicePan(), handlers.GetAllServicePan(alkiraClient))
	srv.AddTool(tools.GetAllServiceZscaler(), handlers.GetAllServiceZscaler(alkiraClient))

	// Add connector tools
	srv.AddTool(tools.GetAllConnectorArubaEdge(), handlers.GetAllConnectorArubaEdge(alkiraClient))
	srv.AddTool(tools.GetAllConnectorAwsDirectConnect(), handlers.GetAllConnectorAwsDirectConnect(alkiraClient))
	srv.AddTool(tools.GetAllConnectorAwsTgw(), handlers.GetAllConnectorAwsTgw(alkiraClient))
	srv.AddTool(tools.GetAllConnectorAzureExpressRoute(), handlers.GetAllConnectorAzureExpressRoute(alkiraClient))
	srv.AddTool(tools.GetAllConnectorCiscoSdwan(), handlers.GetAllConnectorCiscoSdwan(alkiraClient))
	srv.AddTool(tools.GetAllConnectorFortinetSdwan(), handlers.GetAllConnectorFortinetSdwan(alkiraClient))
	srv.AddTool(tools.GetAllConnectorGcpInterconnect(), handlers.GetAllConnectorGcpInterconnect(alkiraClient))
	srv.AddTool(tools.GetAllConnectorGcpVpc(), handlers.GetAllConnectorGcpVpc(alkiraClient))
	srv.AddTool(tools.GetAllConnectorOciVcn(), handlers.GetAllConnectorOciVcn(alkiraClient))
	srv.AddTool(tools.GetAllConnectorRemoteAccessTemplate(), handlers.GetAllConnectorRemoteAccessTemplate(alkiraClient))
	srv.AddTool(tools.GetAllConnectorVersaSdwan(), handlers.GetAllConnectorVersaSdwan(alkiraClient))
	srv.AddTool(tools.GetAllConnectorVmwareSdwan(), handlers.GetAllConnectorVmwareSdwan(alkiraClient))

	// Add internet connector tools
	srv.AddTool(tools.GetInternetConnectors(), handlers.GetInternetConnectors(alkiraClient))
	srv.AddTool(tools.GetInternetConnectorById(), handlers.GetInternetConnectorById(alkiraClient))
	srv.AddTool(tools.GetInternetConnectorByName(), handlers.GetInternetConnectorByName(alkiraClient))

	// Add connector-aws-vpc tools
	srv.AddTool(tools.GetAwsVpcConnectors(), handlers.GetAwsVpcConnectors(alkiraClient))
	srv.AddTool(tools.GetAwsVpcConnectorById(), handlers.GetAwsVpcConnectorById(alkiraClient))
	srv.AddTool(tools.GetAwsVpcConnectorByName(), handlers.GetAwsVpcConnectorByName(alkiraClient))

	// Add connector-azure-vnet tools
	srv.AddTool(tools.GetAzureVnetConnectors(), handlers.GetAzureVnetConnectors(alkiraClient))
	srv.AddTool(tools.GetAzureVnetConnectorById(), handlers.GetAzureVnetConnectorById(alkiraClient))
	srv.AddTool(tools.GetAzureVnetConnectorByName(), handlers.GetAzureVnetConnectorByName(alkiraClient))

	// Add connector-ipsec tools
	srv.AddTool(tools.GetIPSecConnectors(), handlers.GetIPSecConnectors(alkiraClient))
	srv.AddTool(tools.GetIPSecConnectorsSummary(), handlers.GetIPSecConnectorsSummary(alkiraClient))
	srv.AddTool(tools.GetIPSecConnectorById(), handlers.GetIPSecConnectorById(alkiraClient))
	srv.AddTool(tools.GetIPSecConnectorByName(), handlers.GetIPSecConnectorByName(alkiraClient))

	srv.AddTool(tools.GetIPSecAdvConnectors(), handlers.GetIPSecAdvConnectors(alkiraClient))
	srv.AddTool(tools.GetIPSecAdvConnectorById(), handlers.GetIPSecAdvConnectorById(alkiraClient))
	srv.AddTool(tools.GetIPSecAdvConnectorByName(), handlers.GetIPSecAdvConnectorByName(alkiraClient))
	srv.AddTool(tools.GetIPSecConnectorTunnelProfile(), handlers.GetIPSecConnectorTunnelProfile(alkiraClient))

	// Add internet application tools
	srv.AddTool(tools.GetAllInternetApplication(), handlers.GetAllInternetApplication(alkiraClient))

	// Add routes tools
	srv.AddTool(tools.GetRoutes(), handlers.GetRoutes(alkiraClient))
	srv.AddTool(tools.GetRouteCount(), handlers.GetRouteCount(alkiraClient))
	srv.AddTool(tools.GetRouteSummary(), handlers.GetRouteSummary(alkiraClient))
	srv.AddTool(tools.GetAllRoutes(), handlers.GetAllRoutes(alkiraClient))

	// Add NAT policy tools
	srv.AddTool(tools.GetNatPolicies(), handlers.GetNatPolicies(alkiraClient))
	srv.AddTool(tools.GetNatPoliciesSummary(), handlers.GetNatPoliciesSummary(alkiraClient))
	srv.AddTool(tools.GetNatPolicyById(), handlers.GetNatPolicyById(alkiraClient))
	srv.AddTool(tools.GetNatPolicyByName(), handlers.GetNatPolicyByName(alkiraClient))

	srv.AddTool(tools.GetNatPolicyRules(), handlers.GetNatPolicyRules(alkiraClient))
	srv.AddTool(tools.GetNatPolicyRuleById(), handlers.GetNatPolicyRuleById(alkiraClient))
	srv.AddTool(tools.GetNatPolicyRuleByName(), handlers.GetNatPolicyRuleByName(alkiraClient))

	// Add route policy tools
	srv.AddTool(tools.GetRoutePolicies(), handlers.GetRoutePolicies(alkiraClient))
	srv.AddTool(tools.GetRoutePolicyById(), handlers.GetRoutePolicyById(alkiraClient))
	srv.AddTool(tools.GetRoutePolicyByName(), handlers.GetRoutePolicyByName(alkiraClient))

	// Add traffic policy tools
	srv.AddTool(tools.GetTrafficPolicies(), handlers.GetTrafficPolicies(alkiraClient))
	srv.AddTool(tools.GetTrafficPolicyById(), handlers.GetTrafficPolicyById(alkiraClient))
	srv.AddTool(tools.GetTrafficPolicyByName(), handlers.GetTrafficPolicyByName(alkiraClient))

	srv.AddTool(tools.GetTrafficPolicyRules(), handlers.GetTrafficPolicyRules(alkiraClient))
	srv.AddTool(tools.GetTrafficPolicyRuleById(), handlers.GetTrafficPolicyRuleById(alkiraClient))
	srv.AddTool(tools.GetTrafficPolicyRuleByName(), handlers.GetTrafficPolicyRuleByName(alkiraClient))

	// Add list tools
	srv.AddTool(tools.GetAsPathLists(), handlers.GetAsPathLists(alkiraClient))
	srv.AddTool(tools.GetCommunityLists(), handlers.GetCommunityLists(alkiraClient))
	srv.AddTool(tools.GetExtendedCommunityLists(), handlers.GetExtendedCommunityLists(alkiraClient))
	srv.AddTool(tools.GetDnsServerLists(), handlers.GetDnsServerLists(alkiraClient))
	srv.AddTool(tools.GetGlobalCidrLists(), handlers.GetGlobalCidrLists(alkiraClient))
	srv.AddTool(tools.GetUdrLists(), handlers.GetUdrLists(alkiraClient))

	srv.AddTool(tools.GetPolicyRuleLists(), handlers.GetPolicyRuleLists(alkiraClient))
	srv.AddTool(tools.GetPolicyPrefixLists(), handlers.GetPolicyPrefixLists(alkiraClient))
	srv.AddTool(tools.GetPolicyFqdnLists(), handlers.GetPolicyFqdnLists(alkiraClient))

	// Add monitoring tools
	srv.AddTool(tools.GetAlerts(), handlers.GetAlerts(alkiraClient))
	srv.AddTool(tools.GetAuditLogs(), handlers.GetAuditLogs(alkiraClient))
	srv.AddTool(tools.GetJobs(), handlers.GetJobs(alkiraClient))

	srv.AddTool(tools.GetResourceUsages(), handlers.GetResourceUsages(alkiraClient))
	srv.AddTool(tools.GetResourceLimits(), handlers.GetResourceLimits(alkiraClient))

	// Add health tools
	srv.AddTool(tools.GetResourceHealths(), handlers.GetResourceHealths(alkiraClient))
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
