package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/alkiranet/mcp-alkira/handlers"
	"github.com/alkiranet/mcp-alkira/internal/confirm"
	"github.com/alkiranet/mcp-alkira/prompts"
	"github.com/alkiranet/mcp-alkira/registry"

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

// version is the server version reported to MCP clients, set at build time with -ldflags "-X main.version=...".
var version = "dev"

// getEnvOrDefault returns the value of the environment variable or the default value if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// newServer builds the MCP server with the summary prompt and the read tools,
// plus the write tools when allowWrite is set.
func newServer(get handlers.ClientProvider, allowWrite bool) *server.MCPServer {
	srv := server.NewMCPServer("mcp-alkira", version)
	srv.AddPrompt(prompts.Summary(), prompts.SummaryHandler())
	// One key per process signs every confirmation code, so the write tools
	// accept the codes a read tool issues.
	key := handlers.WithConfirmKey(confirm.RandomKey())
	// A 32-byte key cannot fail.
	if err := registry.Register(srv, get, key); err != nil {
		panic(err)
	}
	if allowWrite {
		if err := registry.RegisterWrite(srv, get, key); err != nil {
			panic(err)
		}
	}
	return srv
}

// checkWriteMode refuses write tools outside stdio: the SSE server listens
// on every interface without authentication.
func checkWriteMode(allowWrite bool, mode string) error {
	if allowWrite && mode != "stdio" {
		return fmt.Errorf("-allowWrite requires -mode stdio; mode %q serves without authentication", mode)
	}
	return nil
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
	var allowWrite bool

	flag.StringVar(&port, "port", getEnvOrDefault("SERVER_PORT", "8081"), "Server port")
	flag.StringVar(&mode, "mode", getEnvOrDefault("SERVER_MODE", "stdio"), "Server mode: 'stdio' or 'sse'")

	flag.StringVar(&portal, "portal", "", "Alkira Portal URL")
	flag.StringVar(&key, "key", "", "Alkira API Key")
	flag.IntVar(&maxToken, "maxToken", 0, "Max Token of payload")
	flag.BoolVar(&allowWrite, "allowWrite", os.Getenv("AK_MCP_ALLOW_WRITE") == "true", "Serve create, update and delete tools (stdio only)")
	flag.Parse()

	if err := checkWriteMode(allowWrite, mode); err != nil {
		fmt.Fprintln(os.Stderr, "mcp-alkira:", err)
		os.Exit(2)
	}
	ak.WriteUserAgent = "mcp-alkira/" + version

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

	srv := newServer(func(context.Context) (*ak.AlkiraClient, error) {
		return alkiraClient, nil
	}, allowWrite)
	if allowWrite {
		logf("INFO", "Serving read and write tools")
	} else {
		logf("INFO", "Serving read tools only; start with -allowWrite to add write tools")
	}

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
