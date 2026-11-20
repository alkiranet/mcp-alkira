package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ZtaProfileGetAll() mcp.Tool {
	return mcp.NewTool("zta_profile_get_all",
		mcp.WithDescription("Get all zero trust access profiles of the tenant. A traffic policy names profiles in ztaProfileIds."),
		mcp.WithString("offset",
			mcp.Description("Pagination offset"),
			mcp.DefaultString("0"),
		),
		mcp.WithString("limit",
			mcp.Description("Pagination limit"),
			mcp.DefaultString("50"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func ZtaProfileGetById() mcp.Tool {
	return mcp.NewTool("zta_profile_get_by_id",
		mcp.WithDescription("Get details of one zero trust access profile by its ID (a UUID)."),
		mcp.WithString("profileId",
			mcp.Required(),
			mcp.Description("Zero trust access profile ID."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}
