package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ConnectorArubaEdgeGetAll() mcp.Tool {
	return mcp.NewTool("connector_aruba_edge_get_all",
		mcp.WithDescription("Get all Aruba Edge Connect SD-WAN connectors"),
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

func ConnectorArubaEdgeGetById() mcp.Tool {
	return mcp.NewTool("connector_aruba_edge_get_by_id",
		mcp.WithDescription("Get details of one Aruba Edge Connect SD-WAN connector by its ID."),
		mcp.WithString("connectorId",
			mcp.Required(),
			mcp.Description("Connector ID."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func ConnectorArubaEdgeGetByName() mcp.Tool {
	return mcp.NewTool("connector_aruba_edge_get_by_name",
		mcp.WithDescription("Get details of one Aruba Edge Connect SD-WAN connector by its name."),
		mcp.WithString("connectorName",
			mcp.Required(),
			mcp.Description("Connector Name."),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}

func ConnectorArubaEdgeGetTotal() mcp.Tool {
	return mcp.NewTool("connector_aruba_edge_get_total",
		mcp.WithDescription("Get total numbers of Aruba Edge Connect SD-WAN connectors."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)
}
