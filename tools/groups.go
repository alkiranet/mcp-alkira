package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetGroups() mcp.Tool {
	return mcp.NewTool("getGroups",
		mcp.WithDescription("Get all groups."),
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
