package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GroupGetAll() mcp.Tool {
	return mcp.NewTool("group_get_all",
		mcp.WithDescription("Get all groups. There are usually lots of " +
			"groups. Getting all is not an efficient way to work with groups."),
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

func GroupGetById() mcp.Tool {
	return mcp.NewTool("group_get_by_id",
		mcp.WithDescription("Get details of one group by its ID."),
		mcp.WithString("groupId",
			mcp.Required(),
			mcp.Description("Group ID."),
		),
	)
}

func GroupGetByName() mcp.Tool {
	return mcp.NewTool("group_get_by_name",
		mcp.WithDescription("Get details of one group by its name."),
		mcp.WithString("groupName",
			mcp.Required(),
			mcp.Description("Group Name."),
		),
	)
}

func GroupGetTotal() mcp.Tool {
	return mcp.NewTool("group_get_total",
		mcp.WithDescription("Get total numbers of groups."),
	)
}
