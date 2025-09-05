# CRUSH.md - Alkira MCP Server

## Build/Lint/Test Commands
- `make build` - Build binary with go fmt
- `make fmt` - Format all Go code
- `make vendor` - Update dependencies
- `go build -o bin/mcp-alkira` - Manual build
- `go test ./...` - Run all tests
- `go vet ./...` - Static analysis
- `go mod verify` - Verify dependencies

## Code Style Guidelines
- **Package**: All files in `handlers/` use `package handlers`
- **Imports**: Standard libs first, then third-party, then local
- **Formatting**: Use `gofmt` standard formatting
- **Naming**: PascalCase for functions, camelCase for variables
- **Error Handling**: Return `mcp.NewToolResultError(err.Error())` for errors
- **Function Pattern**: Handler functions return `func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)`
- **Comments**: Use single-line `//` comments for implementation notes
- **Structure**: Consistent pattern: INIT section, API call, error check, return

## Project Structure
- `handlers/` - MCP tool implementations
- `tools/` - Mirror of handlers (likely generated)
- `main.go` - Server entry point with MCP registration
- Use `github.com/alkiranet/client-go` for Alkira API
- Use `github.com/mark3labs/mcp-go` for MCP framework

## Testing
- No existing test files found
- Follow Go testing conventions when adding tests
- Use table-driven tests for handler functions