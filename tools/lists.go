package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAsPathLists() mcp.Tool {
	return mcp.NewTool("getAsPathLists",
		mcp.WithDescription("Get all AS Path lists for BGP routing."),
	)
}

func GetCommunityLists() mcp.Tool {
	return mcp.NewTool("getCommunityLists",
		mcp.WithDescription("Get all BGP Community lists."),
	)
}

func GetExtendedCommunityLists() mcp.Tool {
	return mcp.NewTool("getExtendedCommunityLists",
		mcp.WithDescription("Get all BGP Extended Community lists."),
	)
}

func GetDnsServerLists() mcp.Tool {
	return mcp.NewTool("getDnsServerLists",
		mcp.WithDescription("Get all DNS Server lists."),
	)
}

func GetGlobalCidrLists() mcp.Tool {
	return mcp.NewTool("getGlobalCidrLists",
		mcp.WithDescription("Get all Global CIDR lists."),
	)
}

func GetUdrLists() mcp.Tool {
	return mcp.NewTool("getUdrLists",
		mcp.WithDescription("Get all User Defined Route (UDR) lists."),
	)
}

func GetPolicyRuleLists() mcp.Tool {
	return mcp.NewTool("getPolicyRuleList",
		mcp.WithDescription("Get all Policy Rule lists."),
	)
}

func GetPolicyPrefixLists() mcp.Tool {
	return mcp.NewTool("getPolicyPrefixList",
		mcp.WithDescription("Get all Policy Prefix lists."),
	)
}

func GetPolicyFqdnLists() mcp.Tool {
	return mcp.NewTool("getPolicyFqdnList",
		mcp.WithDescription("Get all Policy FQDN lists."),
	)
}
