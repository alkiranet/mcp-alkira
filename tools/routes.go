package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetRoutes() mcp.Tool {
	return mcp.NewTool("getRoutes",
		mcp.WithDescription("Get routes for a tenant network with optional filtering"),
		mcp.WithString("tenantNetworkId", mcp.Required(), mcp.Description("The tenant network ID")),
		mcp.WithString("type", mcp.Description("Route type filter (received, advertised, overlap)")),
		mcp.WithString("segmentName", mcp.Description("Segment name filter")),
		mcp.WithString("segmentNames", mcp.Description("Multiple segment names (comma-separated)")),
		mcp.WithString("cxp", mcp.Description("Cloud Exchange Point filter")),
		mcp.WithString("connectorId", mcp.Description("Connector ID filter")),
		mcp.WithNumber("offset", mcp.Description("Pagination offset")),
		mcp.WithNumber("limit", mcp.Description("Pagination limit")),
		mcp.WithString("search", mcp.Description("Search filter")),
		mcp.WithString("prefixType", mcp.Description("Prefix type filter")),
		mcp.WithString("routeType", mcp.Description("Route type filter")),
		mcp.WithString("segmentId", mcp.Description("Segment ID filter")),
		mcp.WithString("overlapType", mcp.Description("Overlap type filter")),
		mcp.WithString("sourceCXP", mcp.Description("Source CXP filter")),
		mcp.WithString("entityInstance", mcp.Description("Entity instance filter")),
		mcp.WithString("entityType", mcp.Description("Entity type filter")),
		mcp.WithString("group", mcp.Description("Group filter")),
		mcp.WithString("segmentResourceShare", mcp.Description("Segment resource share filter")),
		mcp.WithString("routeRecvType", mcp.Description("Route receive type filter")),
		mcp.WithString("prefix", mcp.Description("Prefix filter")),
		mcp.WithString("lpmPrefix", mcp.Description("LPM prefix filter")),
	)
}

func GetRouteCount() mcp.Tool {
	return mcp.NewTool("getRouteCount",
		mcp.WithDescription("Get route count for a tenant network with optional filtering"),
		mcp.WithString("tenantNetworkId", mcp.Required(), mcp.Description("The tenant network ID")),
		mcp.WithString("type", mcp.Description("Route type filter")),
		mcp.WithString("segmentName", mcp.Description("Segment name filter")),
		mcp.WithString("segmentNames", mcp.Description("Multiple segment names (comma-separated)")),
		mcp.WithString("cxp", mcp.Description("Cloud Exchange Point filter")),
		mcp.WithString("connectorId", mcp.Description("Connector ID filter")),
		mcp.WithString("segmentId", mcp.Description("Segment ID filter")),
		mcp.WithString("routeRecvType", mcp.Description("Route receive type filter")),
		mcp.WithString("overlapType", mcp.Description("Overlap type filter")),
		mcp.WithString("sourceCXP", mcp.Description("Source CXP filter")),
		mcp.WithString("entityInstance", mcp.Description("Entity instance filter")),
		mcp.WithString("entityType", mcp.Description("Entity type filter")),
		mcp.WithString("group", mcp.Description("Group filter")),
		mcp.WithString("segmentResourceShare", mcp.Description("Segment resource share filter")),
		mcp.WithString("search", mcp.Description("Search filter")),
		mcp.WithString("prefix", mcp.Description("Prefix filter")),
		mcp.WithString("lpmPrefix", mcp.Description("LPM prefix filter")),
		mcp.WithString("prefixType", mcp.Description("Prefix type filter")),
	)
}

