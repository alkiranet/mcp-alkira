package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetRoutes() mcp.Tool {
	return mcp.NewTool("getRoutes",
		mcp.WithDescription("Get routes for a tenant network with optional filtering. The 'type' parameter is required and must be 'received', 'advertised', or 'overlap'. For advertised routes, either segmentName or cxp must be specified."),
		mcp.WithString("tenantNetworkId", mcp.Required(), mcp.Description("The tenant network ID")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Route type filter: 'received', 'advertised', or 'overlap' (required)")),
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
		// Enhanced filtering options
		mcp.WithString("outputFormat", mcp.Description("Output format: json, table, csv, summary (default: json)")),
		mcp.WithString("prefixRange", mcp.Description("Filter by IP prefix range (e.g., '10.0.0.0/8', '192.168.0.0/16')")),
		mcp.WithString("connectorTypes", mcp.Description("Comma-separated list of connector types to include")),
		mcp.WithBoolean("includeSharedRoutes", mcp.Description("Include segment resource shared routes (default: true)")),
		mcp.WithBoolean("includeTranslatedRoutes", mcp.Description("Include NAT translated routes (default: true)")),
		mcp.WithString("routeStatus", mcp.Description("Filter by route status: active, suppressed, overlap")),
	)
}

func GetRouteCount() mcp.Tool {
	return mcp.NewTool("getRouteCount",
		mcp.WithDescription("Get route count for a tenant network with optional filtering. The 'type' parameter is required and must be 'received', 'advertised', or 'overlap'."),
		mcp.WithString("tenantNetworkId", mcp.Required(), mcp.Description("The tenant network ID")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Route type filter: 'received', 'advertised', or 'overlap' (required)")),
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
		// Enhanced filtering options
		mcp.WithString("outputFormat", mcp.Description("Output format: json, table, csv, summary (default: json)")),
		mcp.WithString("prefixRange", mcp.Description("Filter by IP prefix range (e.g., '10.0.0.0/8', '192.168.0.0/16')")),
		mcp.WithString("connectorTypes", mcp.Description("Comma-separated list of connector types to include")),
		mcp.WithBoolean("includeSharedRoutes", mcp.Description("Include segment resource shared routes (default: true)")),
		mcp.WithBoolean("includeTranslatedRoutes", mcp.Description("Include NAT translated routes (default: true)")),
	)
}

// GetRouteSummary provides aggregated route statistics
func GetRouteSummary() mcp.Tool {
	return mcp.NewTool("getRouteSummary",
		mcp.WithDescription("Get aggregated route statistics by connector type, segment, CXP, etc."),
		mcp.WithString("tenantNetworkId", mcp.Required(), mcp.Description("The tenant network ID")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Route type: 'received', 'advertised', or 'overlap' (required)")),
		mcp.WithString("groupBy", mcp.Description("Group routes by: connectorType, segment, cxp, connectorName, prefixRange (default: connectorType)")),
		mcp.WithBoolean("includeDetails", mcp.Description("Include detailed breakdown (default: false)")),
		mcp.WithBoolean("includePrefixAnalysis", mcp.Description("Include IP prefix range analysis (default: false)")),
	)
}

// GetAllRoutes provides automatic pagination for large result sets
func GetAllRoutes() mcp.Tool {
	return mcp.NewTool("getAllRoutes",
		mcp.WithDescription("Get ALL routes with automatic pagination (handles large result sets efficiently)"),
		mcp.WithString("tenantNetworkId", mcp.Required(), mcp.Description("The tenant network ID")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Route type: 'received', 'advertised', or 'overlap' (required)")),
		mcp.WithNumber("batchSize", mcp.Description("Routes per batch (default: 50, max: 100)")),
		mcp.WithBoolean("showProgress", mcp.Description("Show pagination progress (default: false)")),
		mcp.WithString("outputFormat", mcp.Description("Output format: json, summary, count (default: json)")),
		// Filtering options
		mcp.WithString("segmentName", mcp.Description("Filter by segment name")),
		mcp.WithString("connectorType", mcp.Description("Filter by connector type (SAAS, REMOTE_ACCESS, AWS_VPC, etc.)")),
		mcp.WithString("cxp", mcp.Description("Filter by Cloud Exchange Point")),
	)
}

// GetRoutesByConnectorType provides optimized filtering by connector type
func GetRoutesByConnectorType() mcp.Tool {
	return mcp.NewTool("getRoutesByConnectorType",
		mcp.WithDescription("Get all routes for specific connector types with optimized filtering"),
		mcp.WithString("tenantNetworkId", mcp.Required(), mcp.Description("The tenant network ID")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Route type: 'received', 'advertised', or 'overlap' (required)")),
		mcp.WithString("connectorType", mcp.Required(), mcp.Description("Connector type: SAAS, REMOTE_ACCESS, AWS_VPC, GCP_VPC, AZURE_VNET, INB_INT, DIRECT_CONNECT, IP_SEC, ADV_IP_SEC, OCI_VCN")),
		mcp.WithBoolean("includePrefixSummary", mcp.Description("Include prefix range analysis (default: false)")),
		mcp.WithString("outputFormat", mcp.Description("Output format: json, table, summary (default: json)")),
		mcp.WithString("segmentName", mcp.Description("Filter by specific segment")),
	)
}

