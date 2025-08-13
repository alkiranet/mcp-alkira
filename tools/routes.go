package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func GetRoutes() mcp.Tool {
	return mcp.NewTool("getRoutes",
		mcp.WithDescription("Retrieve detailed route information from the routes cache. Examples: Get received routes for segment 'Seg1', Find routes for connector ID 60, Search routes containing 'customer'. Returns route objects with connector details, BGP attributes, segment assignments, and inter-CXP relationships. Requires pagination for large datasets (use limit=100). For advertised routes, specify segmentName or cxp. Common combinations: type+segmentName, type+connectorId+segmentName, type+cxp."),
		mcp.WithString("tenantNetworkId", mcp.Required(), mcp.Description("Tenant network ID to query routes for")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Routes type cache to pull: 'received' (routes learned by Alkira) or 'advertised' (routes Alkira advertises to connectors)")),
		mcp.WithString("segmentName", mcp.Description("Filter by network segment - isolated routing domains that provide security boundaries and traffic control (e.g., 'Corporate', 'DMZ')")),
		mcp.WithString("segmentNames", mcp.Description("Comma-separated list of segment names to match against (e.g., 'seg1,seg2')")),
		mcp.WithString("cxp", mcp.Description("Filter by CXP (Cloud Exchange Point) - Alkira's PoPs in different cloud regions where connectors terminate (e.g., 'US-WEST', 'EU-CENTRAL')")),
		mcp.WithString("connectorId", mcp.Description("Filter by connector ID - connectors link customer networks (AWS VPCs, Azure VNets, on-premises) to Alkira (e.g., 60)")),
		mcp.WithNumber("offset", mcp.Description("Starting offset for route retrieval (pagination required for large datasets)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of routes to return per page (recommended: 100 for optimal performance)")),
		mcp.WithString("search", mcp.Description("Search across all fields, returns routes matching the term (e.g., 'Ipsec', 'customer')")),
		mcp.WithString("prefixType", mcp.Description("Filter by route scope: 'LOCAL' (segment-only), 'SHARED' (cross-segment via segment resource shares)")),
		mcp.WithString("segmentId", mcp.Description("Filter by segment ID (numeric identifier like 12)")),
		mcp.WithString("overlapType", mcp.Description("Filter by overlap type - routes with conflicting prefixes: 'OVERLAP_INVALIDATED', 'SEGSHARE_OVERLAP', 'REMOTE_CXP_OVERLAP_INVALIDATED'")),
		mcp.WithString("sourceCXP", mcp.Description("Filter by source CXP for inter-CXP routes - routes redistributed between different Alkira regions")),
		mcp.WithString("entityInstance", mcp.Description("Exact match for entity instance name (e.g., 'Site1', 'Joshua-VPC-1')")),
		mcp.WithString("entityType", mcp.Description("Filter by entity type: 'IP-SEC', 'VPC', etc.")),
		mcp.WithString("group", mcp.Description("Filter by group routes belong to (e.g., 'routes-group1')")),
		mcp.WithString("segmentResourceShare", mcp.Description("Filter by segment resource share - allows selective route sharing between network segments (e.g., 'corp-dmz-share')")),
		mcp.WithString("routeRecvType", mcp.Description("Filter by route processing status: 'HIGH_CHURN' (suppressed/flapping), 'OVERLAP' (conflicting prefixes), 'ORIGINAL' (clean routes)")),
		mcp.WithString("prefix", mcp.Description("Exact match prefix filter. Use URL encoding for '/' (e.g., '10.1.1.0%2F24' for '10.1.1.0/24')")),
		mcp.WithString("lpmPrefix", mcp.Description("Longest Prefix Match search for IP address. If connectorId provided, searches on that connector; otherwise searches across segments (e.g., '0.0.0.0', '10.1.0.0')")),
		// Enhanced filtering options
		mcp.WithString("outputFormat", mcp.Description("Output format: json, table, csv, summary (default: json)")),
		mcp.WithString("prefixRange", mcp.Description("Filter by IP prefix range (e.g., '10.0.0.0/8', '192.168.0.0/16')")),
		mcp.WithString("connectorTypes", mcp.Description("Comma-separated list of connector types (e.g., 'AWS_VPC,AZURE_VNET,IP_SEC')")),
		mcp.WithBoolean("includeSharedRoutes", mcp.Description("Include segment resource shared routes (default: true)")),
		mcp.WithBoolean("includeTranslatedRoutes", mcp.Description("Include NAT translated routes (default: true)")),
		mcp.WithString("routeStatus", mcp.Description("Filter by route status: active, suppressed, overlap")),
	)
}

func GetRouteCount() mcp.Tool {
	return mcp.NewTool("getRouteCount",
		mcp.WithDescription("Get route count matching specified criteria. More efficient than getRoutes for count-only queries. Useful for: checking route table sizes before pagination, monitoring route growth, validating filter effectiveness before full data retrieval."),
		mcp.WithString("tenantNetworkId", mcp.Required(), mcp.Description("Tenant network ID to count routes for")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Routes type cache to count: 'received' (routes learned by Alkira) or 'advertised' (routes Alkira advertises to connectors)")),
		mcp.WithString("segmentName", mcp.Description("Filter by network segment - isolated routing domains that provide security boundaries and traffic control (e.g., 'Corporate', 'DMZ')")),
		mcp.WithString("segmentNames", mcp.Description("Comma-separated list of segment names to match against (e.g., 'seg1,seg2')")),
		mcp.WithString("cxp", mcp.Description("Filter by CXP (Cloud Exchange Point) - Alkira's PoPs in different cloud regions where connectors terminate (e.g., 'US-WEST', 'EU-CENTRAL')")),
		mcp.WithString("connectorId", mcp.Description("Filter by connector ID - connectors link customer networks (AWS VPCs, Azure VNets, on-premises) to Alkira (e.g., 60)")),
		mcp.WithString("segmentId", mcp.Description("Filter by segment ID (numeric identifier like 12)")),
		mcp.WithString("routeRecvType", mcp.Description("Filter by route processing status: 'HIGH_CHURN' (suppressed/flapping), 'OVERLAP' (conflicting prefixes), 'ORIGINAL' (clean routes)")),
		mcp.WithString("overlapType", mcp.Description("Filter by overlap type - routes with conflicting prefixes: 'OVERLAP_INVALIDATED', 'SEGSHARE_OVERLAP', 'REMOTE_CXP_OVERLAP_INVALIDATED'")),
		mcp.WithString("sourceCXP", mcp.Description("Filter by source CXP for inter-CXP routes - routes redistributed between different Alkira regions")),
		mcp.WithString("entityInstance", mcp.Description("Exact match for entity instance name (e.g., 'Site1', 'Joshua-VPC-1')")),
		mcp.WithString("entityType", mcp.Description("Filter by entity type: 'IP-SEC', 'VPC', etc.")),
		mcp.WithString("group", mcp.Description("Filter by group routes belong to (e.g., 'routes-group1')")),
		mcp.WithString("segmentResourceShare", mcp.Description("Filter by segment resource share - allows selective route sharing between network segments (e.g., 'corp-dmz-share')")),
		mcp.WithString("search", mcp.Description("Search across all fields, returns routes matching the term (e.g., 'Ipsec', 'customer')")),
		mcp.WithString("prefix", mcp.Description("Exact match prefix filter. Use URL encoding for '/' (e.g., '10.1.1.0%2F24' for '10.1.1.0/24')")),
		mcp.WithString("lpmPrefix", mcp.Description("Longest Prefix Match search for IP address (e.g., '0.0.0.0', '10.1.0.0')")),
		mcp.WithString("prefixType", mcp.Description("Filter by route scope: 'LOCAL' (segment-only), 'SHARED' (cross-segment via segment resource shares)")),
		// Enhanced filtering options
		mcp.WithString("outputFormat", mcp.Description("Output format: json, table, csv, summary (default: json)")),
		mcp.WithString("prefixRange", mcp.Description("Filter by IP prefix range (e.g., '10.0.0.0/8', '192.168.0.0/16')")),
		mcp.WithString("connectorTypes", mcp.Description("Comma-separated list of connector types (e.g., 'AWS_VPC,AZURE_VNET,IP_SEC')")),
		mcp.WithBoolean("includeSharedRoutes", mcp.Description("Include segment resource shared routes (default: true)")),
		mcp.WithBoolean("includeTranslatedRoutes", mcp.Description("Include NAT translated routes (default: true)")),
	)
}

// GetRouteSummary provides aggregated route statistics
func GetRouteSummary() mcp.Tool {
	return mcp.NewTool("getRouteSummary",
		mcp.WithDescription("Get aggregated route analytics grouped by connector type, segment, or CXP. Provides route distribution insights, connector utilization, and network topology understanding. Use includeDetails for per-connector breakdowns."),
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
		mcp.WithString("outputFormat", mcp.Description("Output format: json, summary, count (default: json)")),
		// Filtering options
		mcp.WithString("segmentName", mcp.Description("Filter by segment name")),
		mcp.WithString("connectorType", mcp.Description("Filter by connector type (SAAS, REMOTE_ACCESS, AWS_VPC, etc.)")),
		mcp.WithString("cxp", mcp.Description("Filter by Cloud Exchange Point")),
	)
}

// Note: GetRoutesByConnectorType functionality merged into getRoutes with connectorTypes parameter
