package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/alkiranet/mcp-alkira/handlers"
	"github.com/alkiranet/mcp-alkira/tools"

	"github.com/alkiranet/alkira-client-go/alkira"

	"github.com/mark3labs/mcp-go/server"
)

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

	flag.StringVar(&port, "port", getEnvOrDefault("SERVER_PORT", "8081"), "Server port")
	flag.StringVar(&mode, "mode", getEnvOrDefault("SERVER_MODE", "stdio"), "Server mode: 'stdio' or 'sse'")

	flag.StringVar(&portal, "portal", getEnvOrDefault("AK_PORTAL", ""), "Alkira Portal URL")
	flag.StringVar(&key, "key", getEnvOrDefault("AK_KEY", ""), "Alkira API Key")
	flag.Parse()

	// Create Alkira Client
	if portal == "" || key == "" {
		log.Printf("[ERROR] Invalid ENV vars, please specify AK_PORTAL and AK_KEY.")
		return
	}

	alkiraClient, err := alkira.NewAlkiraClient(
		portal,
		"",
		"",
		key,
		false,
		"header",
	)

	if err != nil {
		log.Printf("[ERROR] failed to initialize alkira client, please check your credential and portal URI.")
		return
	}

	// Create MCP server
	srv := server.NewMCPServer(
		"mcp-alkira",
		"0.1.0",
	)

	// Add tools
	srv.AddTool(tools.GetAllSegments(), handlers.GetAllSegments(alkiraClient))
	srv.AddTool(tools.GetAllSegmentResources(), handlers.GetAllSegmentResources(alkiraClient))
	srv.AddTool(tools.GetAllSegmentResourceShares(), handlers.GetAllSegmentResourceShares(alkiraClient))
	srv.AddTool(tools.GetAllGroups(), handlers.GetAllGroups(alkiraClient))
	srv.AddTool(tools.GetAllBillingTags(), handlers.GetAllBillingTags(alkiraClient))
	srv.AddTool(tools.GetAllCxps(), handlers.GetAllCxps(alkiraClient))

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
	srv.AddTool(tools.GetAllConnectorAwsVpc(), handlers.GetAllConnectorAwsVpc(alkiraClient))
	srv.AddTool(tools.GetAllConnectorAzureExpressRoute(), handlers.GetAllConnectorAzureExpressRoute(alkiraClient))
	srv.AddTool(tools.GetAllConnectorAzureVnet(), handlers.GetAllConnectorAzureVnet(alkiraClient))
	srv.AddTool(tools.GetAllConnectorCiscoSdwan(), handlers.GetAllConnectorCiscoSdwan(alkiraClient))
	srv.AddTool(tools.GetAllConnectorFortinetSdwan(), handlers.GetAllConnectorFortinetSdwan(alkiraClient))
	srv.AddTool(tools.GetAllConnectorGcpInterconnect(), handlers.GetAllConnectorGcpInterconnect(alkiraClient))
	srv.AddTool(tools.GetAllConnectorGcpVpc(), handlers.GetAllConnectorGcpVpc(alkiraClient))
	srv.AddTool(tools.GetAllConnectorInternet(), handlers.GetAllConnectorInternet(alkiraClient))
	srv.AddTool(tools.GetAllInternetApplication(), handlers.GetAllInternetApplication(alkiraClient))
	srv.AddTool(tools.GetAllConnectorIPSec(), handlers.GetAllConnectorIPSec(alkiraClient))
	srv.AddTool(tools.GetAllConnectorAdvIPSec(), handlers.GetAllConnectorAdvIPSec(alkiraClient))
	srv.AddTool(tools.GetAllConnectorIPSecTunnelProfile(), handlers.GetAllConnectorIPSecTunnelProfile(alkiraClient))
	srv.AddTool(tools.GetAllConnectorOciVcn(), handlers.GetAllConnectorOciVcn(alkiraClient))
	srv.AddTool(tools.GetAllConnectorRemoteAccessTemplate(), handlers.GetAllConnectorRemoteAccessTemplate(alkiraClient))
	srv.AddTool(tools.GetAllConnectorVersaSdwan(), handlers.GetAllConnectorVersaSdwan(alkiraClient))
	srv.AddTool(tools.GetAllConnectorVmwareSdwan(), handlers.GetAllConnectorVmwareSdwan(alkiraClient))

	// Add policy tools
	srv.AddTool(tools.GetAllNatPolicy(), handlers.GetAllNatPolicy(alkiraClient))
	srv.AddTool(tools.GetAllNatRule(), handlers.GetAllNatRule(alkiraClient))
	srv.AddTool(tools.GetAllRoutePolicy(), handlers.GetAllRoutePolicy(alkiraClient))
	srv.AddTool(tools.GetAllTrafficPolicy(), handlers.GetAllTrafficPolicy(alkiraClient))
	srv.AddTool(tools.GetAllTrafficPolicyRule(), handlers.GetAllTrafficPolicyRule(alkiraClient))
	srv.AddTool(tools.GetAllPolicyRuleList(), handlers.GetAllPolicyRuleList(alkiraClient))
	srv.AddTool(tools.GetAllPolicyPrefixList(), handlers.GetAllPolicyPrefixList(alkiraClient))
	srv.AddTool(tools.GetAllPolicyFqdnList(), handlers.GetAllPolicyFqdnList(alkiraClient))

	// Add list tools
	srv.AddTool(tools.GetAllListAsPath(), handlers.GetAllListAsPath(alkiraClient))
	srv.AddTool(tools.GetAllListCommunity(), handlers.GetAllListCommunity(alkiraClient))
	srv.AddTool(tools.GetAllListExtendedCommunity(), handlers.GetAllListExtendedCommunity(alkiraClient))
	srv.AddTool(tools.GetAllDnsServerList(), handlers.GetAllDnsServerList(alkiraClient))
	srv.AddTool(tools.GetAllGlobalCidrList(), handlers.GetAllGlobalCidrList(alkiraClient))
	srv.AddTool(tools.GetAllUdrList(), handlers.GetAllUdrList(alkiraClient))
	srv.AddTool(tools.GetAllPolicyPrefixListIndividual(), handlers.GetAllPolicyPrefixListIndividual(alkiraClient))
	srv.AddTool(tools.GetAllPolicyFqdnListIndividual(), handlers.GetAllPolicyFqdnListIndividual(alkiraClient))

	// Add monitoring tools
	srv.AddTool(tools.GetAlerts(), handlers.GetAlerts(alkiraClient))
	srv.AddTool(tools.GetAuditLogs(), handlers.GetAuditLogs(alkiraClient))
	srv.AddTool(tools.GetJobs(), handlers.GetJobs(alkiraClient))
	srv.AddTool(tools.GetAllHealth(), handlers.GetAllHealth(alkiraClient))

	// Start server based on mode
	switch mode {
	case "stdio":
		fmt.Println("Starting server in stdio mode...")
		if err := server.ServeStdio(srv); err != nil {
			fmt.Printf("Failed to start stdio server: %v\n", err)
			return
		}
	case "sse":
		fmt.Printf("Starting server in SSE mode on port %s...\n", port)
		sse := server.NewSSEServer(srv)
		if err := sse.Start(":" + port); err != nil {
			fmt.Printf("Failed to start SSE server: %v\n", err)
			return
		}
		fmt.Printf("SSE server started on port %s\n", port)
	default:
		fmt.Printf("Unknown server mode: %s. Use 'stdio' or 'sse'.\n", mode)
		return
	}

	fmt.Println("👋 Server stopped")
}
