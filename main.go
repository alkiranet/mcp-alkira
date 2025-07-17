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
	var portal string
	var username string
	var password string

	flag.StringVar(&portal, "portal", getEnvOrDefault("AK_PORTAL", ""), "Alkira Portal URL")
	flag.StringVar(&username, "username", getEnvOrDefault("AK_USERNAME", ""), "Alkira Portal Username")
	flag.StringVar(&password, "password", getEnvOrDefault("AK_PASSWORD", ""), "Alkira Portal Password")
	flag.Parse()

	// Create Alkira Client
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

	// Start the stdio server
	fmt.Println("🚀 mcp-alkira server started")

	if err := server.ServeStdio(srv); err != nil {
		fmt.Printf("😡 Server error: %v\n", err)
	}

	fmt.Println("👋 Server stopped")
}
