package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alkiranet/alkira-client-go/alkira"
	"github.com/mark3labs/mcp-go/mcp"
)

// Route summary structures for enhanced analysis
type RouteSummaryByConnectorType struct {
	ConnectorType    string                 `json:"connectorType"`
	Count            int                    `json:"count"`
	UniqueConnectors int                    `json:"uniqueConnectors"`
	Segments         []string               `json:"segments"`
	CXPs             []string               `json:"cxps"`
	PrefixRanges     map[string]int         `json:"prefixRanges,omitempty"`
	Details          []RouteConnectorDetail `json:"details,omitempty"`
}

type RouteConnectorDetail struct {
	ConnectorName string   `json:"connectorName"`
	ConnectorID   int      `json:"connectorId"`
	SegmentName   string   `json:"segmentName"`
	CXP           string   `json:"cxp"`
	Prefixes      []string `json:"prefixes"`
	RouteCount    int      `json:"routeCount"`
}

type RouteSummaryResponse struct {
	TotalRoutes      int                           `json:"totalRoutes"`
	GroupBy          string                        `json:"groupBy"`
	GeneratedAt      time.Time                     `json:"generatedAt"`
	SummaryByType    []RouteSummaryByConnectorType `json:"summaryByType,omitempty"`
	SummaryBySegment map[string]int                `json:"summaryBySegment,omitempty"`
	SummaryByCXP     map[string]int                `json:"summaryByCXP,omitempty"`
	PrefixAnalysis   *PrefixAnalysis               `json:"prefixAnalysis,omitempty"`
}

type PrefixAnalysis struct {
	TotalPrefixes    int            `json:"totalPrefixes"`
	DefaultRoutes    int            `json:"defaultRoutes"`
	HostRoutes       int            `json:"hostRoutes"`
	PrivateRanges    map[string]int `json:"privateRanges"`
	PublicRanges     int            `json:"publicRanges"`
	TranslatedRoutes int            `json:"translatedRoutes"`
	SharedRoutes     int            `json:"sharedRoutes"`
}

// Enhanced query parameters
type EnhancedRouteQueryParams struct {
	alkira.RouteQueryParams
	OutputFormat            string `json:"outputFormat,omitempty"`
	PrefixRange             string `json:"prefixRange,omitempty"`
	ConnectorTypes          string `json:"connectorTypes,omitempty"`
	IncludeSharedRoutes     bool   `json:"includeSharedRoutes"`
	IncludeTranslatedRoutes bool   `json:"includeTranslatedRoutes"`
	RouteStatus             string `json:"routeStatus,omitempty"`
}

// Pagination helper for large result sets
type PaginationHelper struct {
	TotalRoutes     int                    `json:"totalRoutes"`
	BatchSize       int                    `json:"batchSize"`
	TotalBatches    int                    `json:"totalBatches"`
	ProcessedRoutes int                    `json:"processedRoutes"`
	Routes          []alkira.RouteUIResult `json:"routes,omitempty"`
	Summary         *RouteSummaryResponse  `json:"summary,omitempty"`
}

// getRoutesEnhanced retrieves routes with enhanced filtering and output formatting
func getRoutesEnhanced(client *alkira.AlkiraClient, params EnhancedRouteQueryParams) (interface{}, error) {
	// Push as many filters as possible to the API level
	optimizedParams := optimizeRouteQueryParams(params)

	// Get routes with API-level filtering
	routes, err := client.GetRoutes(optimizedParams)
	if err != nil {
		return nil, err
	}

	// Apply only the remaining client-side filtering that can't be done by API
	filteredRoutes := applyRemainingFiltering(routes.Data, params)
	routes.Data = filteredRoutes

	// Format output based on outputFormat parameter
	switch strings.ToLower(params.OutputFormat) {
	case "summary":
		return generateRouteSummary(routes.Data, "connectorType", false, false), nil
	case "table":
		return formatAsTable(routes.Data), nil
	case "csv":
		return formatAsCSV(routes.Data), nil
	default:
		return routes, nil
	}
}

func GetRoutes(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Validate required parameters
		_, err := request.RequireString("tenantNetworkId")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		routeType := request.GetString("type", "received")
		if routeType != "received" && routeType != "advertised" && routeType != "overlap" {
			return mcp.NewToolResultError("type must be 'received', 'advertised', or 'overlap'"), nil
		}

		// Auto-optimize limit for better performance
		limit := request.GetInt("limit", 0)
		if limit == 0 || limit > 1000 {
			limit = 100 // Optimal default
		}

		// Build enhanced query parameters from request
		params := EnhancedRouteQueryParams{
			RouteQueryParams: alkira.RouteQueryParams{
				Type:                 routeType,
				SegmentName:          request.GetString("segmentName", ""),
				SegmentNames:         request.GetString("segmentNames", ""),
				CXP:                  request.GetString("cxp", ""),
				ConnectorID:          request.GetString("connectorId", ""),
				Offset:               request.GetInt("offset", 0),
				Limit:                limit,
				Search:               request.GetString("search", ""),
				PrefixType:           request.GetString("prefixType", ""),
				RouteType:            request.GetString("routeType", ""),
				SegmentID:            request.GetString("segmentId", ""),
				OverlapType:          request.GetString("overlapType", ""),
				SourceCXP:            request.GetString("sourceCXP", ""),
				EntityInstance:       request.GetString("entityInstance", ""),
				EntityType:           request.GetString("entityType", ""),
				Group:                request.GetString("group", ""),
				SegmentResourceShare: request.GetString("segmentResourceShare", ""),
				RouteRecvType:        request.GetString("routeRecvType", ""),
				Prefix:               request.GetString("prefix", ""),
				LPMPrefix:            request.GetString("lpmPrefix", ""),
			},
			// Enhanced parameters
			OutputFormat:            request.GetString("outputFormat", "json"),
			PrefixRange:             request.GetString("prefixRange", ""),
			ConnectorTypes:          request.GetString("connectorTypes", ""),
			IncludeSharedRoutes:     request.GetBool("includeSharedRoutes", true),
			IncludeTranslatedRoutes: request.GetBool("includeTranslatedRoutes", true),
			RouteStatus:             request.GetString("routeStatus", ""),
		}

		// Get routes with enhanced filtering
		routes, err := getRoutesEnhanced(client, params)
		if err != nil {
			// Provide helpful error messages for common issues
			if strings.Contains(err.Error(), "404") {
				return mcp.NewToolResultError("Tenant network not found. Verify tenantNetworkId is correct"), nil
			}
			if strings.Contains(err.Error(), "segment") && routeType == "advertised" {
				return mcp.NewToolResultError("For advertised routes, specify either segmentName or cxp parameter"), nil
			}
			return mcp.NewToolResultError("Route query failed: " + err.Error()), nil
		}

		// Return response as JSON
		routesJSON, err := json.Marshal(routes)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal routes response: " + err.Error()), nil
		}
		return mcp.NewToolResultText(string(routesJSON)), nil
	}
}

func GetRouteCount(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Validate required parameters
		_, err := request.RequireString("tenantNetworkId")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		routeType := request.GetString("type", "received")
		if routeType != "received" && routeType != "advertised" && routeType != "overlap" {
			return mcp.NewToolResultError("type must be 'received', 'advertised', or 'overlap'"), nil
		}

		// Build query parameters from request
		params := alkira.RouteCountQueryParams{
			Type:                 routeType,
			SegmentName:          request.GetString("segmentName", ""),
			SegmentNames:         request.GetString("segmentNames", ""),
			CXP:                  request.GetString("cxp", ""),
			ConnectorID:          request.GetString("connectorId", ""),
			SegmentID:            request.GetString("segmentId", ""),
			RouteRecvType:        request.GetString("routeRecvType", ""),
			OverlapType:          request.GetString("overlapType", ""),
			SourceCXP:            request.GetString("sourceCXP", ""),
			EntityInstance:       request.GetString("entityInstance", ""),
			EntityType:           request.GetString("entityType", ""),
			Group:                request.GetString("group", ""),
			SegmentResourceShare: request.GetString("segmentResourceShare", ""),
			Search:               request.GetString("search", ""),
			Prefix:               request.GetString("prefix", ""),
			LPMPrefix:            request.GetString("lpmPrefix", ""),
			PrefixType:           request.GetString("prefixType", ""),
		}

		// Get route count using alkira client directly
		routeCount, err := client.GetRouteCount(params)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Return response as JSON
		routeCountJSON, err := json.Marshal(routeCount)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal route count response: " + err.Error()), nil
		}
		return mcp.NewToolResultText(string(routeCountJSON)), nil
	}
}

// optimizeRouteQueryParams pushes as many filters as possible to the API level
func optimizeRouteQueryParams(params EnhancedRouteQueryParams) alkira.RouteQueryParams {
	optimized := params.RouteQueryParams

	// Push connector type filtering to API using EntityType
	if params.ConnectorTypes != "" {
		// Take the first connector type for API filtering
		connectorTypes := strings.Split(params.ConnectorTypes, ",")
		if len(connectorTypes) > 0 {
			optimized.EntityType = strings.TrimSpace(connectorTypes[0])
		}
	}

	// Push shared route filtering to API using PrefixType
	if !params.IncludeSharedRoutes {
		optimized.PrefixType = "LOCAL"
	}

	// Push translated route filtering to API using RouteType
	if !params.IncludeTranslatedRoutes {
		optimized.RouteType = "ORIGINAL"
	}

	// Push route status filtering to API using RouteRecvType
	if params.RouteStatus != "" {
		switch strings.ToLower(params.RouteStatus) {
		case "suppressed":
			optimized.RouteRecvType = "HIGH_CHURN"
		case "overlap":
			optimized.RouteRecvType = "OVERLAP"
		case "active":
			optimized.RouteRecvType = "ORIGINAL"
		}
	}

	return optimized
}

// applyRemainingFiltering applies only the filtering that cannot be done by the API
func applyRemainingFiltering(routes []alkira.RouteUIResult, params EnhancedRouteQueryParams) []alkira.RouteUIResult {
	filtered := routes

	// Handle multiple connector types (API can only handle one)
	if params.ConnectorTypes != "" {
		connectorTypes := strings.Split(params.ConnectorTypes, ",")
		if len(connectorTypes) > 1 {
			// API already filtered by first type, now filter by the remaining types
			filtered = filterByConnectorTypes(filtered, connectorTypes)
		}
	}

	// Filter by prefix range (custom logic that API doesn't support)
	if params.PrefixRange != "" {
		filtered = filterByPrefixRange(filtered, params.PrefixRange)
	}

	return filtered
}

func filterByConnectorTypes(routes []alkira.RouteUIResult, connectorTypes []string) []alkira.RouteUIResult {
	var filtered []alkira.RouteUIResult
	for _, route := range routes {
		for _, connector := range route.Connectors {
			for _, targetType := range connectorTypes {
				if strings.EqualFold(strings.TrimSpace(targetType), connector.Connector.ConnectorType) {
					filtered = append(filtered, route)
					goto nextRoute
				}
			}
		}
	nextRoute:
	}
	return filtered
}

func filterByPrefixRange(routes []alkira.RouteUIResult, prefixRange string) []alkira.RouteUIResult {
	_, targetNet, err := net.ParseCIDR(prefixRange)
	if err != nil {
		return routes // Invalid CIDR, return all
	}

	var filtered []alkira.RouteUIResult
	for _, route := range routes {
		if prefixInRange(route.Prefix, targetNet) {
			filtered = append(filtered, route)
		}
	}
	return filtered
}

func prefixInRange(prefix string, targetNet *net.IPNet) bool {
	if prefix == "0.0.0.0/0" {
		return false // Skip default routes
	}

	ip, _, err := net.ParseCIDR(prefix)
	if err != nil {
		// Try parsing as IP
		ip = net.ParseIP(prefix)
		if ip == nil {
			return false
		}
	}

	return targetNet.Contains(ip)
}

// Output formatting methods
func formatAsTable(routes []alkira.RouteUIResult) string {
	if len(routes) == 0 {
		return "No routes found."
	}

	var result strings.Builder
	result.WriteString("| Prefix | Segment | Connector Name | Connector Type | CXP | Route Type |\n")
	result.WriteString("|--------|---------|----------------|----------------|-----|------------|\n")

	for _, route := range routes {
		for _, connector := range route.Connectors {
			result.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
				route.Prefix,
				route.SegmentName,
				connector.Connector.ConnectorName,
				connector.Connector.ConnectorType,
				connector.Connector.ConnectorCXPName,
				route.RouteType))
		}
	}

	return result.String()
}

func formatAsCSV(routes []alkira.RouteUIResult) string {
	var result strings.Builder
	w := csv.NewWriter(&result)

	// Write header
	w.Write([]string{"Prefix", "Segment", "ConnectorName", "ConnectorType", "ConnectorID", "CXP", "RouteType", "PrefixType", "ResourceShare"})

	// Write data
	for _, route := range routes {
		for _, connector := range route.Connectors {
			w.Write([]string{
				route.Prefix,
				route.SegmentName,
				connector.Connector.ConnectorName,
				connector.Connector.ConnectorType,
				strconv.Itoa(connector.Connector.ConnectorID),
				connector.Connector.ConnectorCXPName,
				route.RouteType,
				connector.PrefixType,
				connector.ResourceShareName,
			})
		}
	}

	w.Flush()
	return result.String()
}

// Route summary generation
func generateRouteSummary(routes []alkira.RouteUIResult, groupBy string, includeDetails bool, includePrefixAnalysis bool) *RouteSummaryResponse {
	summary := &RouteSummaryResponse{
		TotalRoutes: len(routes),
		GroupBy:     groupBy,
		GeneratedAt: time.Now(),
	}

	switch groupBy {
	case "connectorType":
		summary.SummaryByType = summarizeByConnectorType(routes, includeDetails)
	case "segment":
		summary.SummaryBySegment = summarizeBySegment(routes)
	case "cxp":
		summary.SummaryByCXP = summarizeByCXP(routes)
	}

	if includePrefixAnalysis {
		summary.PrefixAnalysis = analyzePrefixes(routes)
	}

	return summary
}

func summarizeByConnectorType(routes []alkira.RouteUIResult, includeDetails bool) []RouteSummaryByConnectorType {
	typeMap := make(map[string]*RouteSummaryByConnectorType)
	connectorMap := make(map[string]map[string]bool) // connectorType -> connectorName -> exists

	for _, route := range routes {
		for _, connector := range route.Connectors {
			connType := connector.Connector.ConnectorType

			if typeMap[connType] == nil {
				typeMap[connType] = &RouteSummaryByConnectorType{
					ConnectorType: connType,
					Segments:      make([]string, 0),
					CXPs:          make([]string, 0),
				}
				connectorMap[connType] = make(map[string]bool)
			}

			typeMap[connType].Count++

			// Track unique connectors
			connectorMap[connType][connector.Connector.ConnectorName] = true

			// Track segments
			if !contains(typeMap[connType].Segments, route.SegmentName) {
				typeMap[connType].Segments = append(typeMap[connType].Segments, route.SegmentName)
			}

			// Track CXPs
			if !contains(typeMap[connType].CXPs, connector.Connector.ConnectorCXPName) {
				typeMap[connType].CXPs = append(typeMap[connType].CXPs, connector.Connector.ConnectorCXPName)
			}

			if includeDetails {
				if typeMap[connType].Details == nil {
					typeMap[connType].Details = make([]RouteConnectorDetail, 0)
				}
				// Add connector detail (simplified for now)
				detail := RouteConnectorDetail{
					ConnectorName: connector.Connector.ConnectorName,
					ConnectorID:   connector.Connector.ConnectorID,
					SegmentName:   route.SegmentName,
					CXP:           connector.Connector.ConnectorCXPName,
					Prefixes:      []string{route.Prefix},
					RouteCount:    1,
				}
				typeMap[connType].Details = append(typeMap[connType].Details, detail)
			}
		}
	}

	// Set unique connector counts
	for connType, summary := range typeMap {
		summary.UniqueConnectors = len(connectorMap[connType])
	}

	// Convert map to slice and sort
	var result []RouteSummaryByConnectorType
	for _, summary := range typeMap {
		result = append(result, *summary)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	return result
}

func summarizeBySegment(routes []alkira.RouteUIResult) map[string]int {
	segmentCount := make(map[string]int)
	for _, route := range routes {
		segmentCount[route.SegmentName]++
	}
	return segmentCount
}

func summarizeByCXP(routes []alkira.RouteUIResult) map[string]int {
	cxpCount := make(map[string]int)
	for _, route := range routes {
		for _, connector := range route.Connectors {
			cxpCount[connector.Connector.ConnectorCXPName]++
		}
	}
	return cxpCount
}

func analyzePrefixes(routes []alkira.RouteUIResult) *PrefixAnalysis {
	analysis := &PrefixAnalysis{
		PrivateRanges: make(map[string]int),
	}

	for _, route := range routes {
		analysis.TotalPrefixes++

		if route.Prefix == "0.0.0.0/0" {
			analysis.DefaultRoutes++
			continue
		}

		if strings.HasSuffix(route.Prefix, "/32") {
			analysis.HostRoutes++
		}

		if route.RouteType == "TRANSLATED" {
			analysis.TranslatedRoutes++
		}

		for _, connector := range route.Connectors {
			if connector.PrefixType == "SHARED" || connector.ResourceShareName != "" {
				analysis.SharedRoutes++
				break
			}
		}

		// Analyze IP ranges
		ip, _, err := net.ParseCIDR(route.Prefix)
		if err != nil {
			ip = net.ParseIP(route.Prefix)
		}

		if ip != nil {
			if ip.IsPrivate() {
				if ip.To4() != nil {
					firstOctet := ip.To4()[0]
					switch {
					case firstOctet == 10:
						analysis.PrivateRanges["10.x.x.x"]++
					case firstOctet == 172:
						analysis.PrivateRanges["172.16-31.x.x"]++
					case firstOctet == 192:
						analysis.PrivateRanges["192.168.x.x"]++
					}
				}
			} else {
				analysis.PublicRanges++
			}
		}
	}

	return analysis
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetRouteSummary handler
func GetRouteSummary(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract tenant network ID (required)
		_, err := request.RequireString("tenantNetworkId")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Extract parameters
		routeType := request.GetString("type", "received")
		groupBy := request.GetString("groupBy", "connectorType")
		includeDetails := request.GetBool("includeDetails", false)
		includePrefixAnalysis := request.GetBool("includePrefixAnalysis", false)

		// Get routes with efficient API-level filtering (avoid fetching ALL routes)
		allRoutes, err := getAllRoutesWithPaginationOptimized(client, routeType)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Generate summary
		summary := generateRouteSummary(allRoutes, groupBy, includeDetails, includePrefixAnalysis)

		summaryJSON, err := json.Marshal(summary)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal summary: " + err.Error()), nil
		}

		return mcp.NewToolResultText(string(summaryJSON)), nil
	}
}

// GetAllRoutes handler with automatic pagination
func GetAllRoutes(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract tenant network ID (required)
		_, err := request.RequireString("tenantNetworkId")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Extract parameters
		routeType := request.GetString("type", "received")
		batchSize := request.GetInt("batchSize", 50)
		outputFormat := request.GetString("outputFormat", "json")
		segmentName := request.GetString("segmentName", "")
		connectorType := request.GetString("connectorType", "")
		cxp := request.GetString("cxp", "")

		if batchSize > 100 {
			batchSize = 100
		}

		// Get all routes with pagination
		allRoutes, err := getAllRoutesWithPaginationAndFiltering(client, routeType, batchSize, segmentName, connectorType, cxp)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Format output
		switch strings.ToLower(outputFormat) {
		case "summary":
			summary := generateRouteSummary(allRoutes, "connectorType", false, true)
			summaryJSON, err := json.Marshal(summary)
			if err != nil {
				return mcp.NewToolResultError("Failed to marshal summary: " + err.Error()), nil
			}
			return mcp.NewToolResultText(string(summaryJSON)), nil
		case "count":
			count := map[string]int{"totalRoutes": len(allRoutes)}
			countJSON, _ := json.Marshal(count)
			return mcp.NewToolResultText(string(countJSON)), nil
		default:
			result := alkira.RoutesUIResponse{
				Data: allRoutes,
				Pagination: alkira.PaginationData{
					Offset: 0,
					Limit:  len(allRoutes),
					Hits:   len(allRoutes),
				},
			}
			routesJSON, err := json.Marshal(result)
			if err != nil {
				return mcp.NewToolResultError("Failed to marshal routes: " + err.Error()), nil
			}
			return mcp.NewToolResultText(string(routesJSON)), nil
		}
	}
}

// GetRoutesByConnectorType functionality merged into getRoutes with connectorTypes parameter

// Helper functions for pagination
// getAllRoutesWithPaginationOptimized uses smarter pagination with reasonable limits
func getAllRoutesWithPaginationOptimized(client *alkira.AlkiraClient, routeType string) ([]alkira.RouteUIResult, error) {
	var allRoutes []alkira.RouteUIResult
	offset := 0
	limit := 100       // Use optimal batch size
	maxRoutes := 10000 // Reasonable safety limit to prevent memory issues

	for {
		params := alkira.RouteQueryParams{
			Type:   routeType,
			Offset: offset,
			Limit:  limit,
		}

		response, err := client.GetRoutes(params)
		if err != nil {
			return nil, err
		}

		allRoutes = append(allRoutes, response.Data...)

		// Safety check to prevent infinite loops and memory issues
		if len(allRoutes) >= maxRoutes {
			break
		}

		if len(response.Data) < limit {
			break
		}

		offset += limit
	}

	return allRoutes, nil
}

func getAllRoutesWithPaginationAndFiltering(client *alkira.AlkiraClient, routeType string, batchSize int, segmentName, connectorType, cxp string) ([]alkira.RouteUIResult, error) {
	var allRoutes []alkira.RouteUIResult
	offset := 0
	maxRoutes := 10000 // Safety limit

	for {
		params := alkira.RouteQueryParams{
			Type:        routeType,
			Offset:      offset,
			Limit:       batchSize,
			SegmentName: segmentName,
			EntityType:  connectorType,
			CXP:         cxp,
		}

		response, err := client.GetRoutes(params)
		if err != nil {
			return nil, err
		}

		allRoutes = append(allRoutes, response.Data...)

		// Safety check to prevent memory issues
		if len(allRoutes) >= maxRoutes {
			break
		}

		if len(response.Data) < batchSize {
			break
		}

		offset += batchSize
	}

	return allRoutes, nil
}
