package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetAllSegments() mcp.Tool {
	return mcp.NewTool("getAllSegments",
		mcp.WithDescription("Get all segments"),
	)
}

func GetAllGroups() mcp.Tool {
	return mcp.NewTool("getAllGroups",
		mcp.WithDescription("Get all groups"),
	)
}

func GetAllBillingTags() mcp.Tool {
	return mcp.NewTool("getAllBillingTags",
		mcp.WithDescription("Get all billing tags"),
	)
}

func GetAllCxps() mcp.Tool {
	return mcp.NewTool("getAllCxps",
		mcp.WithDescription("Get all available CXP (Cloud Exchange Point). " +
			"Each CXP is unqiuly defined by cloud provider (provider) and " +
			"could provider region (providerRegion)."),
	)
}

