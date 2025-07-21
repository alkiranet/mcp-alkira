package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAllConnectorArubaEdge() mcp.Tool {
	return mcp.NewTool("getAllConnectorArubaEdge",
		mcp.WithDescription("Get all Aruba Edge Connect SD-WAN connectors"),
	)
}

func GetAllConnectorAwsDirectConnect() mcp.Tool {
	return mcp.NewTool("getAllConnectorAwsDirectConnect",
		mcp.WithDescription("Get all AWS Direct Connect connectors"),
	)
}

func GetAllConnectorAwsTgw() mcp.Tool {
	return mcp.NewTool("getAllConnectorAwsTgw",
		mcp.WithDescription("Get all AWS Transit Gateway connectors"),
	)
}

func GetAllConnectorAwsVpc() mcp.Tool {
	return mcp.NewTool("getAllConnectorAwsVpc",
		mcp.WithDescription("Get all AWS VPC connectors"),
	)
}

func GetAllConnectorAzureExpressRoute() mcp.Tool {
	return mcp.NewTool("getAllConnectorAzureExpressRoute",
		mcp.WithDescription("Get all Azure ExpressRoute connectors"),
	)
}

func GetAllConnectorAzureVnet() mcp.Tool {
	return mcp.NewTool("getAllConnectorAzureVnet",
		mcp.WithDescription("Get all Azure Virtual Network connectors"),
	)
}

func GetAllConnectorCiscoSdwan() mcp.Tool {
	return mcp.NewTool("getAllConnectorCiscoSdwan",
		mcp.WithDescription("Get all Cisco SD-WAN connectors"),
	)
}

func GetAllConnectorFortinetSdwan() mcp.Tool {
	return mcp.NewTool("getAllConnectorFortinetSdwan",
		mcp.WithDescription("Get all Fortinet SD-WAN connectors"),
	)
}

func GetAllConnectorGcpInterconnect() mcp.Tool {
	return mcp.NewTool("getAllConnectorGcpInterconnect",
		mcp.WithDescription("Get all Google Cloud Interconnect connectors"),
	)
}

func GetAllConnectorGcpVpc() mcp.Tool {
	return mcp.NewTool("getAllConnectorGcpVpc",
		mcp.WithDescription("Get all Google Cloud VPC connectors"),
	)
}

func GetAllConnectorInternet() mcp.Tool {
	return mcp.NewTool("getAllConnectorInternet",
		mcp.WithDescription("Get all Internet Exit connectors"),
	)
}

func GetAllConnectorIPSec() mcp.Tool {
	return mcp.NewTool("getAllConnectorIPSec",
		mcp.WithDescription("Get all IPSec connectors"),
	)
}

func GetAllConnectorAdvIPSec() mcp.Tool {
	return mcp.NewTool("getAllConnectorAdvIPSec",
		mcp.WithDescription("Get all Advanced IPSec connectors"),
	)
}

func GetAllConnectorIPSecTunnelProfile() mcp.Tool {
	return mcp.NewTool("getAllConnectorIPSecTunnelProfile",
		mcp.WithDescription("Get all IPSec Tunnel Profile connectors"),
	)
}

func GetAllConnectorOciVcn() mcp.Tool {
	return mcp.NewTool("getAllConnectorOciVcn",
		mcp.WithDescription("Get all Oracle Cloud Infrastructure VCN connectors"),
	)
}

func GetAllConnectorRemoteAccessTemplate() mcp.Tool {
	return mcp.NewTool("getAllConnectorRemoteAccessTemplate",
		mcp.WithDescription("Get all Remote Access Template connectors"),
	)
}

func GetAllConnectorVersaSdwan() mcp.Tool {
	return mcp.NewTool("getAllConnectorVersaSdwan",
		mcp.WithDescription("Get all Versa SD-WAN connectors"),
	)
}

func GetAllConnectorVmwareSdwan() mcp.Tool {
	return mcp.NewTool("getAllConnectorVmwareSdwan",
		mcp.WithDescription("Get all VMware SD-WAN (VeloCloud) connectors"),
	)
}
