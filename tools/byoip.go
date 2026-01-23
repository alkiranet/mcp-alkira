package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ByoipGetAll() mcp.Tool {
	return mcp.NewTool("byoip_get_all",
		mcp.WithDescription("Get all BYOIPs (Bring Your Own IP)."),
		mcp.WithString("offset",
			mcp.Description("Pagination offset"),
			mcp.DefaultString("0"),
		),
		mcp.WithString("limit",
			mcp.Description("Pagination limit"),
			mcp.DefaultString("50"),
		),
	)
}

func ByoipGetById() mcp.Tool {
	return mcp.NewTool("byoip_get_by_id",
		mcp.WithDescription("Get details of one BYOIP by its ID."),
		mcp.WithString("byoipId",
			mcp.Required(),
			mcp.Description("BYOIP ID."),
		),
	)
}

func ByoipGetByName() mcp.Tool {
	return mcp.NewTool("byoip_get_by_name",
		mcp.WithDescription("Get details of one BYOIP by its name."),
		mcp.WithString("byoipName",
			mcp.Required(),
			mcp.Description("BYOIP Name."),
		),
	)
}

func ByoipGetTotal() mcp.Tool {
	return mcp.NewTool("byoip_get_total",
		mcp.WithDescription("Get total numbers of BYOIPs."),
	)
}
