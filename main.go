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
	var username string
	var password string

	flag.StringVar(&port, "port", getEnvOrDefault("SERVER_PORT", "8081"), "Server port")
	flag.StringVar(&mode, "mode", getEnvOrDefault("SERVER_MODE", "stdio"), "Server mode: 'stdio' or 'sse'")

	flag.StringVar(&portal, "portal", getEnvOrDefault("AK_PORTAL", "terraform.preprod.alkira3.net"), "Alkira Portal URL")
	flag.StringVar(&username, "username", getEnvOrDefault("AK_USERNAME", "terraform@alkira.com"), "Alkira Portal Username")
	flag.StringVar(&password, "password", getEnvOrDefault("AK_PASSWORD", "Alkira2023!"), "Alkira Portal Password")
	flag.Parse()

	// Create Alkira Client
	if portal == "" || username == "" || password == "" {
		log.Printf("[ERROR] Invalid ENV vars, please specify AK_PORTAL, AK_USERNAME and AK_PASSWORD.")
		return
	}

	alkiraClient, err := alkira.NewAlkiraClient(
		portal,
		username,
		password,
		"",
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
	srv.AddTool(tools.GetAllGroups(), handlers.GetAllGroups(alkiraClient))
	srv.AddTool(tools.GetAllBillingTags(), handlers.GetAllBillingTags(alkiraClient))
	srv.AddTool(tools.GetAllLists(), handlers.GetAllLists(alkiraClient))
	srv.AddTool(tools.GetAllConnectors(), handlers.GetAllConnectors(alkiraClient))

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
