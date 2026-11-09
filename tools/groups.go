package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GroupGetById() mcp.Tool {
	return mcp.NewTool("group_get_by_id",
		mcp.WithDescription("Get details of one group by its ID."),
		mcp.WithString("groupId",
			mcp.Required(),
			mcp.Description("Group ID."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func GroupGetByName() mcp.Tool {
	return mcp.NewTool("group_get_by_name",
		mcp.WithDescription("Get details of one group by its name."),
		mcp.WithString("groupName",
			mcp.Required(),
			mcp.Description("Group Name."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func GroupGetTotal() mcp.Tool {
	return mcp.NewTool("group_get_total",
		mcp.WithDescription("Get total numbers of groups."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}
