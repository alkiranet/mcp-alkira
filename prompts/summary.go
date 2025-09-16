package prompts

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)


func Summary() mcp.Prompt {

	return mcp.NewPrompt("summary",
		mcp.WithPromptDescription("Prompt to gather summary of essential informations of the tenant"),
	)
}

func SummaryHandler() func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {

	return func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return mcp.NewGetPromptResult(
			"Summary",
			[]mcp.PromptMessage{
				mcp.NewPromptMessage(
					mcp.RoleUser,
					mcp.NewTextContent("Firsly, get tenant summary and " +
						"total number of resources and CXPs, combine " +
						"the result together, then summarize alerts and " +
						"audit logs, summarize the result and print " +
						"in a fancy table by the categories of " +
						"resources (connector, service, list, etc)."),
				),
			},
		), nil
	}
}
