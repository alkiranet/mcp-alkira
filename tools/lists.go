package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAllListAsPath() mcp.Tool {
	return mcp.NewTool("getAllListAsPath",
		mcp.WithDescription("Get all AS Path lists for BGP routing"),
	)
}

func GetAllListCommunity() mcp.Tool {
	return mcp.NewTool("getAllListCommunity",
		mcp.WithDescription("Get all BGP Community lists"),
	)
}

func GetAllListExtendedCommunity() mcp.Tool {
	return mcp.NewTool("getAllListExtendedCommunity",
		mcp.WithDescription("Get all BGP Extended Community lists"),
	)
}

func GetAllDnsServerList() mcp.Tool {
	return mcp.NewTool("getAllDnsServerList",
		mcp.WithDescription("Get all DNS Server lists"),
	)
}

func GetAllGlobalCidrList() mcp.Tool {
	return mcp.NewTool("getAllGlobalCidrList",
		mcp.WithDescription("Get all Global CIDR lists"),
	)
}

func GetAllUdrList() mcp.Tool {
	return mcp.NewTool("getAllUdrList",
		mcp.WithDescription("Get all User Defined Route (UDR) lists"),
	)
}

func GetAllPolicyPrefixListIndividual() mcp.Tool {
	return mcp.NewTool("getAllPolicyPrefixListIndividual",
		mcp.WithDescription("Get all Policy Prefix lists individually"),
	)
}

func GetAllPolicyFqdnListIndividual() mcp.Tool {
	return mcp.NewTool("getAllPolicyFqdnListIndividual",
		mcp.WithDescription("Get all Policy FQDN lists individually"),
	)
}