package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetCxps() mcp.Tool {
	return mcp.NewTool("getCxps",
		mcp.WithDescription("Get all CXPs (Cloud Exchange Point). "+
			"Each CXP is unqiuly defined by cloud provider (`provider`) and "+
			"could provider region (`providerRegion`)."),
	)
}
