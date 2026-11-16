package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func BillingTagGetAll() mcp.Tool {
	return mcp.NewTool("billing_tag_get_all",
		mcp.WithDescription("Get all billing tags"),
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

func BillingTagGetById() mcp.Tool {
	return mcp.NewTool("billing_tag_get_by_id",
		mcp.WithDescription("Get billing tag by ID"),
		mcp.WithString("tagId", mcp.Required()),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func BillingTagGetByName() mcp.Tool {
	return mcp.NewTool("billing_tag_get_by_name",
		mcp.WithDescription("Get billing tag by name"),
		mcp.WithString("tagName", mcp.Required()),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func BillingTagGetTotal() mcp.Tool {
	return mcp.NewTool("billing_tag_get_total",
		mcp.WithDescription("Get total numbers of billing tags."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}
