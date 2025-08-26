package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetGroups() mcp.Tool {
	return mcp.NewTool("getGroups",
		mcp.WithDescription("Get all groups. The returned data will be " +
			"always paginated."),
		mcp.WithString("offset",
			mcp.Description("Offset of paginated data will be returned."),
		),
		mcp.WithString("limit",
			mcp.Description("Limit of paginated data will be returned. If " +
				"not provided, default value is 10."),
		),
	)
}

func GetGroupById() mcp.Tool {
	return mcp.NewTool("getGroupById",
		mcp.WithDescription("Get details of one group by its ID."),
		mcp.WithString("groupId",
			mcp.Required(),
			mcp.Description("Group ID."),
		),
	)
}

func GetGroupByName() mcp.Tool {
	return mcp.NewTool("getGroupByName",
		mcp.WithDescription("Get details of one group by its name."),
		mcp.WithString("groupName",
			mcp.Required(),
			mcp.Description("Group Name."),
		),
	)
}
