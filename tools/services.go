package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAllServiceCheckpoint() mcp.Tool {
	return mcp.NewTool("getAllServiceCheckpoint",
		mcp.WithDescription("Get all Check Point firewall services"),
	)
}

func GetAllServiceCiscoFTDv() mcp.Tool {
	return mcp.NewTool("getAllServiceCiscoFTDv",
		mcp.WithDescription("Get all Cisco Firepower Threat Defense virtual services"),
	)
}

func GetAllServiceF5Lb() mcp.Tool {
	return mcp.NewTool("getAllServiceF5Lb",
		mcp.WithDescription("Get all F5 Load Balancer services"),
	)
}

func GetAllServiceFortinet() mcp.Tool {
	return mcp.NewTool("getAllServiceFortinet",
		mcp.WithDescription("Get all Fortinet firewall services"),
	)
}

func GetAllServiceInfoblox() mcp.Tool {
	return mcp.NewTool("getAllServiceInfoblox",
		mcp.WithDescription("Get all Infoblox DNS/DHCP services"),
	)
}

func GetAllServicePan() mcp.Tool {
	return mcp.NewTool("getAllServicePan",
		mcp.WithDescription("Get all Palo Alto Networks firewall services"),
	)
}

func GetAllServiceZscaler() mcp.Tool {
	return mcp.NewTool("getAllServiceZscaler",
		mcp.WithDescription("Get all Zscaler security services"),
	)
}
