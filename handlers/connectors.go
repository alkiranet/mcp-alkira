package handlers

import (
	"context"

	"github.com/alkiranet/alkira-client-go/alkira"
	"github.com/mark3labs/mcp-go/mcp"
)

type ConnectorAPI interface {
	GetAll() (string, error)
}

func createConnectorHandler[T ConnectorAPI](newConnectorFunc func(*alkira.AlkiraClient) T) func(*alkira.AlkiraClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(client *alkira.AlkiraClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			api := newConnectorFunc(client)
			connectors, err := api.GetAll()
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(connectors), nil
		}
	}
}

var (
	GetAllConnectorArubaEdge               = createConnectorHandler(alkira.NewConnectorArubaEdge)
	GetAllConnectorAwsDirectConnect        = createConnectorHandler(alkira.NewConnectorAwsDirectConnect)
	GetAllConnectorAwsTgw                  = createConnectorHandler(alkira.NewConnectorAwsTgw)
	GetAllConnectorAwsVpc                  = createConnectorHandler(alkira.NewConnectorAwsVpc)
	GetAllConnectorAzureExpressRoute       = createConnectorHandler(alkira.NewConnectorAzureExpressRoute)
	GetAllConnectorAzureVnet               = createConnectorHandler(alkira.NewConnectorAzureVnet)
	GetAllConnectorCiscoSdwan              = createConnectorHandler(alkira.NewConnectorCiscoSdwan)
	GetAllConnectorFortinetSdwan           = createConnectorHandler(alkira.NewConnectorFortinetSdwan)
	GetAllConnectorGcpInterconnect         = createConnectorHandler(alkira.NewConnectorGcpInterconnect)
	GetAllConnectorGcpVpc                  = createConnectorHandler(alkira.NewConnectorGcpVpc)
	GetAllConnectorInternet                = createConnectorHandler(alkira.NewConnectorInternet)
	GetAllConnectorIPSec                   = createConnectorHandler(alkira.NewConnectorIPSec)
	GetAllConnectorAdvIPSec                = createConnectorHandler(alkira.NewConnectorAdvIPSec)
	GetAllConnectorIPSecTunnelProfile      = createConnectorHandler(alkira.NewConnectorIPSecTunnelProfile)
	GetAllConnectorOciVcn                  = createConnectorHandler(alkira.NewConnectorOciVcn)
	GetAllConnectorRemoteAccessTemplate    = CreatePaginatedGetAllHandler(func(client *alkira.AlkiraClient) GetAllAPI {
		return alkira.NewConnectorRemoteAccessTemplate(client)
	})
	GetAllConnectorVersaSdwan              = createConnectorHandler(alkira.NewConnectorVersaSdwan)
	GetAllConnectorVmwareSdwan             = createConnectorHandler(alkira.NewConnectorVmwareSdwan)
)
