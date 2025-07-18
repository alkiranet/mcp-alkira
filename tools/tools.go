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

func GetAllLists() mcp.Tool {
	return mcp.NewTool("getAllLists",
		mcp.WithDescription("Get all lists"),
	)
}

func GetAllConnectors() mcp.Tool {
	return mcp.NewTool("getAllConnectors",
		mcp.WithDescription("Get all connectors"),
	)
}
