package registry

import (
	"github.com/alkiranet/mcp-alkira/handlers"
	"github.com/alkiranet/mcp-alkira/tools"
	"github.com/mark3labs/mcp-go/server"
)

// The provision preview, validation and status tools. Register serves the
// read tools with a preview pipeline built from its options. No tool
// provisions: changes are provisioned from the portal.
func init() {
	readToolSets = append(readToolSets, func(srv *server.MCPServer, get handlers.ClientProvider, opts []handlers.WriteOption) error {
		p, err := handlers.NewPreviewPipeline(opts...)
		if err != nil {
			return err
		}
		srv.AddTool(tools.TenantNetworkChangesGet(), perCall(get, handlers.TenantNetworkChangesGet))
		srv.AddTool(tools.TenantNetworkProvisionPreview(), perCall(get, handlers.ProvisionPreview(p)))
		srv.AddTool(tools.TenantNetworkProvisionRequestGet(), perCall(get, handlers.ProvisionRequestGet))
		srv.AddTool(tools.TenantNetworkValidationRequestGet(), perCall(get, handlers.ValidationRequestGet))
		srv.AddTool(tools.TenantNetworkValidationStateGet(), perCall(get, handlers.ValidationStateGet))
		return nil
	})
	writeToolSets = append(writeToolSets, func(srv *server.MCPServer, get handlers.ClientProvider, p *handlers.WritePipeline) {
		srv.AddTool(tools.TenantNetworkValidate(), perCall(get, handlers.Validate(p)))
	})
}
