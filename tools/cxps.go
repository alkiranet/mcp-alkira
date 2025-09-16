package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func CxpGetAll() mcp.Tool {
	return mcp.NewTool("cxp_get_all",
		mcp.WithDescription("Get all CXPs (Cloud Exchange Point). "+
			"Each CXP is uniquely defined by cloud provider (`provider` field) " +
			"and cloud provider region (`providerRegion` field)."),
	)
}
