package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alkiranet/alkira-client-go/alkira"
	"github.com/mark3labs/mcp-go/mcp"
)

// Route query parameters for filtering routes
type RouteQueryParams struct {
	Type                 string `json:"type,omitempty"`                 // required: received, advertised, overlap
	SegmentName          string `json:"segmentName,omitempty"`          // segment name filter
	SegmentNames         string `json:"segmentNames,omitempty"`         // multiple segment names (comma-separated)
	CXP                  string `json:"cxp,omitempty"`                  // Cloud Exchange Point filter
	ConnectorID          string `json:"connectorId,omitempty"`          // connector ID filter
	Offset               int    `json:"offset,omitempty"`               // pagination offset
	Limit                int    `json:"limit,omitempty"`                // pagination limit
	Search               string `json:"search,omitempty"`               // search filter
	PrefixType           string `json:"prefixType,omitempty"`           // prefix type filter
	RouteType            string `json:"routeType,omitempty"`            // route type filter
	SegmentID            string `json:"segmentId,omitempty"`            // segment ID filter
	OverlapType          string `json:"overlapType,omitempty"`          // overlap type filter
	SourceCXP            string `json:"sourceCXP,omitempty"`            // source CXP filter
	EntityInstance       string `json:"entityInstance,omitempty"`       // entity instance filter
	EntityType           string `json:"entityType,omitempty"`           // entity type filter
	Group                string `json:"group,omitempty"`                // group filter
	SegmentResourceShare string `json:"segmentResourceShare,omitempty"` // segment resource share filter
	RouteRecvType        string `json:"routeRecvType,omitempty"`        // route receive type filter
	Prefix               string `json:"prefix,omitempty"`               // prefix filter
	LPMPrefix            string `json:"lpmPrefix,omitempty"`            // LPM prefix filter
}

// Route count query parameters
type RouteCountQueryParams struct {
	Type                 string `json:"type,omitempty"`                 // route type filter
	SegmentName          string `json:"segmentName,omitempty"`          // segment name filter
	SegmentNames         string `json:"segmentNames,omitempty"`         // multiple segment names (comma-separated)
	CXP                  string `json:"cxp,omitempty"`                  // Cloud Exchange Point filter
	ConnectorID          string `json:"connectorId,omitempty"`          // connector ID filter
	SegmentID            string `json:"segmentId,omitempty"`            // segment ID filter
	RouteRecvType        string `json:"routeRecvType,omitempty"`        // route receive type filter
	OverlapType          string `json:"overlapType,omitempty"`          // overlap type filter
	SourceCXP            string `json:"sourceCXP,omitempty"`            // source CXP filter
	EntityInstance       string `json:"entityInstance,omitempty"`       // entity instance filter
	EntityType           string `json:"entityType,omitempty"`           // entity type filter
	Group                string `json:"group,omitempty"`                // group filter
	SegmentResourceShare string `json:"segmentResourceShare,omitempty"` // segment resource share filter
	Search               string `json:"search,omitempty"`               // search filter
	Prefix               string `json:"prefix,omitempty"`               // prefix filter
	LPMPrefix            string `json:"lpmPrefix,omitempty"`            // LPM prefix filter
	PrefixType           string `json:"prefixType,omitempty"`           // prefix type filter
}

// Route UI connector represents a connector in a route
type RouteUIConnector struct {
	Connector struct {
		ConnectorName         string `json:"connectorName"`
		ConnectorInstanceName string `json:"connectorInstanceName"`
		ConnectorType         string `json:"connectorType"`
		ConnectorID           int    `json:"connectorId"`
		ConnectorTag          string `json:"connectorTag"`
		ConnectorCXPID        string `json:"connectorCxpId"`
		ConnectorCXPName      string `json:"connectorCxpName"`
		ConnectorRouteCXP     string `json:"connectorRouteCxp"`
		ConnectorGroup        string `json:"connectorGroup"`
	} `json:"connector"`
	PrefixType          string      `json:"prefixType"`
	ConnRouteType       string      `json:"connRouteType"`
	ConnOriginalPrefix  string      `json:"connOriginalPrefix"`
	ResourceShareName   string      `json:"resourceShareName"`
	RouteSuppressed     bool        `json:"routeSuppressed"`
	OverlappingEntities interface{} `json:"overlappingEntities"`
}

// Route UI result represents a single route result
type RouteUIResult struct {
	TenantNetworkID        int                `json:"tenantNetworkId"`
	Prefix                 string             `json:"prefix"`
	SegmentName            string             `json:"segmentName"`
	VrfName                string             `json:"vrfName"`
	CxpName                string             `json:"cxpName"`
	RouteType              string             `json:"routeType"`
	OriginalPrefix         string             `json:"originalPrefix"`
	NatMetadata            interface{}        `json:"natMetadata"`
	OverlapCxps            []string           `json:"overlapCxps"`
	OverlapInvalidateNodes interface{}        `json:"overlapInvalidateNodes"`
	OverlappedRSMetadata   interface{}        `json:"overlappedRsMetadata"`
	CXPToOverlapEntities   interface{}        `json:"cxpToOverlapEntities"`
	OverlapEntityToCXPs    interface{}        `json:"overlapEntityToCXPs"`
	Connectors             []RouteUIConnector `json:"connectors"`
	SegShareOverlapping    interface{}        `json:"segShareOverlapping"`
}

// Pagination data for route responses
type PaginationData struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Hits   int `json:"hits"`
}

// Routes UI response
type RoutesUIResponse struct {
	Data                 []RouteUIResult `json:"data"`
	Pagination           PaginationData  `json:"pagination"`
	LatestRouteTimeStamp int             `json:"latestRouteTimestamp"`
}

// Route count response
type RouteCountResponse struct {
	Count int `json:"count"`
}

// Route count UI result
type RouteCountUIResult struct {
	SegmentName string `json:"segment"`
	CxpName     string `json:"cxp"`
	Count       int    `json:"count"`
}

// Route summary structures for enhanced analysis
type RouteSummaryByConnectorType struct {
	ConnectorType  string                    `json:"connectorType"`
	Count          int                       `json:"count"`
	UniqueConnectors int                     `json:"uniqueConnectors"`
	Segments       []string                  `json:"segments"`
	CXPs           []string                  `json:"cxps"`
	PrefixRanges   map[string]int           `json:"prefixRanges,omitempty"`
	Details        []RouteConnectorDetail   `json:"details,omitempty"`
}

type RouteConnectorDetail struct {
	ConnectorName   string   `json:"connectorName"`
	ConnectorID     int      `json:"connectorId"`
	SegmentName     string   `json:"segmentName"`
	CXP             string   `json:"cxp"`
	Prefixes        []string `json:"prefixes"`
	RouteCount      int      `json:"routeCount"`
}

type RouteSummaryResponse struct {
	TotalRoutes        int                          `json:"totalRoutes"`
	GroupBy            string                       `json:"groupBy"`
	GeneratedAt        time.Time                    `json:"generatedAt"`
	SummaryByType      []RouteSummaryByConnectorType `json:"summaryByType,omitempty"`
	SummaryBySegment   map[string]int               `json:"summaryBySegment,omitempty"`
	SummaryByCXP       map[string]int               `json:"summaryByCXP,omitempty"`
	PrefixAnalysis     *PrefixAnalysis              `json:"prefixAnalysis,omitempty"`
}

type PrefixAnalysis struct {
	TotalPrefixes     int            `json:"totalPrefixes"`
	DefaultRoutes     int            `json:"defaultRoutes"`
	HostRoutes        int            `json:"hostRoutes"`
	PrivateRanges     map[string]int `json:"privateRanges"`
	PublicRanges      int            `json:"publicRanges"`
	TranslatedRoutes  int            `json:"translatedRoutes"`
	SharedRoutes      int            `json:"sharedRoutes"`
}

// Enhanced query parameters
type EnhancedRouteQueryParams struct {
	RouteQueryParams
	OutputFormat           string `json:"outputFormat,omitempty"`
	PrefixRange            string `json:"prefixRange,omitempty"`
	ConnectorTypes         string `json:"connectorTypes,omitempty"`
	IncludeSharedRoutes    bool   `json:"includeSharedRoutes"`
	IncludeTranslatedRoutes bool  `json:"includeTranslatedRoutes"`
	RouteStatus            string `json:"routeStatus,omitempty"`
}

// Routes API client for handlers
type RoutesClientHandler struct {
	Client *alkira.AlkiraClient
}

// Pagination helper for large result sets
type PaginationHelper struct {
	TotalRoutes    int                `json:"totalRoutes"`
	BatchSize      int                `json:"batchSize"`
	TotalBatches   int                `json:"totalBatches"`
	ProcessedRoutes int               `json:"processedRoutes"`
	Routes         []RouteUIResult    `json:"routes,omitempty"`
	Summary        *RouteSummaryResponse `json:"summary,omitempty"`
}

// NewRoutesClientHandler creates a new Routes client for handlers
func NewRoutesClientHandler(client *alkira.AlkiraClient) *RoutesClientHandler {
	return &RoutesClientHandler{
		Client: client,
	}
}

// GetRoutesEnhanced retrieves routes with enhanced filtering and output formatting
func (r *RoutesClientHandler) GetRoutesEnhanced(tenantNetworkID string, params EnhancedRouteQueryParams) (interface{}, error) {
	// First get the basic routes
	routes, err := r.GetRoutes(tenantNetworkID, params.RouteQueryParams)
	if err != nil {
		return nil, err
	}

	// Apply enhanced filtering
	filteredRoutes := r.applyEnhancedFiltering(routes.Data, params)
	routes.Data = filteredRoutes

	// Format output based on outputFormat parameter
	switch strings.ToLower(params.OutputFormat) {
	case "summary":
		return r.generateRouteSummary(routes.Data, "connectorType", false, false), nil
	case "table":
		return r.formatAsTable(routes.Data), nil
	case "csv":
		return r.formatAsCSV(routes.Data), nil
	default:
		return routes, nil
	}
}

// GetRoutes retrieves routes for a tenant network with optional filtering
func (r *RoutesClientHandler) GetRoutes(tenantNetworkID string, params RouteQueryParams) (*RoutesUIResponse, error) {
	// Construct the URI
	uri := fmt.Sprintf("%s/tenantnetworks/%s/routes", r.Client.URI, tenantNetworkID)

	// Build query parameters
	queryParams := url.Values{}
	if params.Type != "" {
		queryParams.Set("type", params.Type)
	}
	if params.SegmentName != "" {
		queryParams.Set("segmentName", params.SegmentName)
	}
	if params.SegmentNames != "" {
		queryParams.Set("segmentNames", params.SegmentNames)
	}
	if params.CXP != "" {
		queryParams.Set("cxp", params.CXP)
	}
	if params.ConnectorID != "" {
		queryParams.Set("connectorId", params.ConnectorID)
	}
	if params.Offset > 0 {
		queryParams.Set("offset", strconv.Itoa(params.Offset))
	}
	if params.Limit > 0 {
		queryParams.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Search != "" {
		queryParams.Set("search", params.Search)
	}
	if params.PrefixType != "" {
		queryParams.Set("prefixType", params.PrefixType)
	}
	if params.RouteType != "" {
		queryParams.Set("routeType", params.RouteType)
	}
	if params.SegmentID != "" {
		queryParams.Set("segmentId", params.SegmentID)
	}
	if params.OverlapType != "" {
		queryParams.Set("overlapType", params.OverlapType)
	}
	if params.SourceCXP != "" {
		queryParams.Set("sourceCXP", params.SourceCXP)
	}
	if params.EntityInstance != "" {
		queryParams.Set("entityInstance", params.EntityInstance)
	}
	if params.EntityType != "" {
		queryParams.Set("entityType", params.EntityType)
	}
	if params.Group != "" {
		queryParams.Set("group", params.Group)
	}
	if params.SegmentResourceShare != "" {
		queryParams.Set("segmentResourceShare", params.SegmentResourceShare)
	}
	if params.RouteRecvType != "" {
		queryParams.Set("routeRecvType", params.RouteRecvType)
	}
	if params.Prefix != "" {
		queryParams.Set("prefix", params.Prefix)
	}
	if params.LPMPrefix != "" {
		queryParams.Set("lpmPrefix", params.LPMPrefix)
	}

	// Add query parameters to URI if any exist
	if len(queryParams) > 0 {
		uri += "?" + queryParams.Encode()
	}

	// Create a temporary API to use the get method
	tempAPI := &alkira.AlkiraAPI[RoutesUIResponse]{
		Client: r.Client,
		Uri:    uri,
	}

	// Validate required parameters
	if params.Type == "" {
		return nil, fmt.Errorf("type parameter is required (must be 'received', 'advertised', or 'overlap')")
	}
	if params.Type == "advertised" && params.SegmentName == "" && params.CXP == "" {
		return nil, fmt.Errorf("advertised routes require either segmentName or cxp parameter")
	}

	// Make the request
	data, err := tempAPI.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get routes: %v", err)
	}

	// Parse the response
	var result RoutesUIResponse
	err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal routes response: %v", err)
	}

	return &result, nil
}

// GetRouteCount retrieves route count for a tenant network with optional filtering
func (r *RoutesClientHandler) GetRouteCount(tenantNetworkID string, params RouteCountQueryParams) (*RouteCountResponse, error) {
	// Construct the URI
	uri := fmt.Sprintf("%s/tenantnetworks/%s/route-count", r.Client.URI, tenantNetworkID)

	// Build query parameters
	queryParams := url.Values{}
	if params.Type != "" {
		queryParams.Set("type", params.Type)
	}
	if params.SegmentName != "" {
		queryParams.Set("segmentName", params.SegmentName)
	}
	if params.SegmentNames != "" {
		queryParams.Set("segmentNames", params.SegmentNames)
	}
	if params.CXP != "" {
		queryParams.Set("cxp", params.CXP)
	}
	if params.ConnectorID != "" {
		queryParams.Set("connectorId", params.ConnectorID)
	}
	if params.SegmentID != "" {
		queryParams.Set("segmentId", params.SegmentID)
	}
	if params.RouteRecvType != "" {
		queryParams.Set("routeRecvType", params.RouteRecvType)
	}
	if params.OverlapType != "" {
		queryParams.Set("overlapType", params.OverlapType)
	}
	if params.SourceCXP != "" {
		queryParams.Set("sourceCXP", params.SourceCXP)
	}
	if params.EntityInstance != "" {
		queryParams.Set("entityInstance", params.EntityInstance)
	}
	if params.EntityType != "" {
		queryParams.Set("entityType", params.EntityType)
	}
	if params.Group != "" {
		queryParams.Set("group", params.Group)
	}
	if params.SegmentResourceShare != "" {
		queryParams.Set("segmentResourceShare", params.SegmentResourceShare)
	}
	if params.Search != "" {
		queryParams.Set("search", params.Search)
	}
	if params.Prefix != "" {
		queryParams.Set("prefix", params.Prefix)
	}
	if params.LPMPrefix != "" {
		queryParams.Set("lpmPrefix", params.LPMPrefix)
	}
	if params.PrefixType != "" {
		queryParams.Set("prefixType", params.PrefixType)
	}

	// Add query parameters to URI if any exist
	if len(queryParams) > 0 {
		uri += "?" + queryParams.Encode()
	}

	// Create a temporary API to use the get method
	tempAPI := &alkira.AlkiraAPI[RouteCountResponse]{
		Client: r.Client,
		Uri:    uri,
	}

	// Validate required parameters
	if params.Type == "" {
		return nil, fmt.Errorf("type parameter is required (must be 'received', 'advertised', or 'overlap')")
	}

	// Make the request
	data, err := tempAPI.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get route count: %v", err)
	}

	// Parse the response
	var result RouteCountResponse
	err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal route count response: %v", err)
	}

	return &result, nil
}

func GetRoutes(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract tenant network ID (required)
		tenantNetworkID, err := request.RequireString("tenantNetworkId")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Initialize routes client
		api := NewRoutesClientHandler(client)

		// Build enhanced query parameters from request
		params := EnhancedRouteQueryParams{
			RouteQueryParams: RouteQueryParams{
				Type:                 request.GetString("type", "received"), // Default to 'received' if not specified
			SegmentName:          request.GetString("segmentName", ""),
			SegmentNames:         request.GetString("segmentNames", ""),
			CXP:                  request.GetString("cxp", ""),
			ConnectorID:          request.GetString("connectorId", ""),
			Offset:               request.GetInt("offset", 0),
			Limit:                request.GetInt("limit", 0),
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
			OutputFormat:           request.GetString("outputFormat", "json"),
			PrefixRange:            request.GetString("prefixRange", ""),
			ConnectorTypes:         request.GetString("connectorTypes", ""),
			IncludeSharedRoutes:    request.GetBool("includeSharedRoutes", true),
			IncludeTranslatedRoutes: request.GetBool("includeTranslatedRoutes", true),
			RouteStatus:            request.GetString("routeStatus", ""),
		}

		// Get routes with enhanced filtering
		routes, err := api.GetRoutesEnhanced(tenantNetworkID, params)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
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
		// Extract tenant network ID (required)
		tenantNetworkID, err := request.RequireString("tenantNetworkId")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Initialize routes client
		api := NewRoutesClientHandler(client)

		// Build query parameters from request
		params := RouteCountQueryParams{
			Type:                 request.GetString("type", "received"), // Default to 'received' if not specified
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

		// Get route count
		routeCount, err := api.GetRouteCount(tenantNetworkID, params)
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

// Enhanced filtering methods
func (r *RoutesClientHandler) applyEnhancedFiltering(routes []RouteUIResult, params EnhancedRouteQueryParams) []RouteUIResult {
	filtered := routes

	// Filter by connector types
	if params.ConnectorTypes != "" {
		connectorTypes := strings.Split(params.ConnectorTypes, ",")
		filtered = r.filterByConnectorTypes(filtered, connectorTypes)
	}

	// Filter by prefix range
	if params.PrefixRange != "" {
		filtered = r.filterByPrefixRange(filtered, params.PrefixRange)
	}

	// Filter shared routes
	if !params.IncludeSharedRoutes {
		filtered = r.filterOutSharedRoutes(filtered)
	}

	// Filter translated routes
	if !params.IncludeTranslatedRoutes {
		filtered = r.filterOutTranslatedRoutes(filtered)
	}

	// Filter by route status
	if params.RouteStatus != "" {
		filtered = r.filterByRouteStatus(filtered, params.RouteStatus)
	}

	return filtered
}

func (r *RoutesClientHandler) filterByConnectorTypes(routes []RouteUIResult, connectorTypes []string) []RouteUIResult {
	var filtered []RouteUIResult
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

func (r *RoutesClientHandler) filterByPrefixRange(routes []RouteUIResult, prefixRange string) []RouteUIResult {
	_, targetNet, err := net.ParseCIDR(prefixRange)
	if err != nil {
		return routes // Invalid CIDR, return all
	}

	var filtered []RouteUIResult
	for _, route := range routes {
		if r.prefixInRange(route.Prefix, targetNet) {
			filtered = append(filtered, route)
		}
	}
	return filtered
}

func (r *RoutesClientHandler) prefixInRange(prefix string, targetNet *net.IPNet) bool {
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

func (r *RoutesClientHandler) filterOutSharedRoutes(routes []RouteUIResult) []RouteUIResult {
	var filtered []RouteUIResult
	for _, route := range routes {
		hasSharedConnector := false
		for _, connector := range route.Connectors {
			if connector.PrefixType == "SHARED" || connector.ResourceShareName != "" {
				hasSharedConnector = true
				break
			}
		}
		if !hasSharedConnector {
			filtered = append(filtered, route)
		}
	}
	return filtered
}

func (r *RoutesClientHandler) filterOutTranslatedRoutes(routes []RouteUIResult) []RouteUIResult {
	var filtered []RouteUIResult
	for _, route := range routes {
		if route.RouteType != "TRANSLATED" {
			filtered = append(filtered, route)
		}
	}
	return filtered
}

func (r *RoutesClientHandler) filterByRouteStatus(routes []RouteUIResult, status string) []RouteUIResult {
	var filtered []RouteUIResult
	switch strings.ToLower(status) {
	case "active":
		for _, route := range routes {
			hasActiveConnector := false
			for _, connector := range route.Connectors {
				if !connector.RouteSuppressed {
					hasActiveConnector = true
					break
				}
			}
			if hasActiveConnector {
				filtered = append(filtered, route)
			}
		}
	case "suppressed":
		for _, route := range routes {
			hasSuppressedConnector := false
			for _, connector := range route.Connectors {
				if connector.RouteSuppressed {
					hasSuppressedConnector = true
					break
				}
			}
			if hasSuppressedConnector {
				filtered = append(filtered, route)
			}
		}
	case "overlap":
		for _, route := range routes {
			if strings.Contains(strings.ToUpper(route.RouteType), "OVERLAP") {
				filtered = append(filtered, route)
			}
		}
	default:
		filtered = routes
	}
	return filtered
}

// Output formatting methods
func (r *RoutesClientHandler) formatAsTable(routes []RouteUIResult) string {
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

func (r *RoutesClientHandler) formatAsCSV(routes []RouteUIResult) string {
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
func (r *RoutesClientHandler) generateRouteSummary(routes []RouteUIResult, groupBy string, includeDetails bool, includePrefixAnalysis bool) *RouteSummaryResponse {
	summary := &RouteSummaryResponse{
		TotalRoutes: len(routes),
		GroupBy:     groupBy,
		GeneratedAt: time.Now(),
	}

	switch groupBy {
	case "connectorType":
		summary.SummaryByType = r.summarizeByConnectorType(routes, includeDetails)
	case "segment":
		summary.SummaryBySegment = r.summarizeBySegment(routes)
	case "cxp":
		summary.SummaryByCXP = r.summarizeByCXP(routes)
	}

	if includePrefixAnalysis {
		summary.PrefixAnalysis = r.analyzePrefixes(routes)
	}

	return summary
}

func (r *RoutesClientHandler) summarizeByConnectorType(routes []RouteUIResult, includeDetails bool) []RouteSummaryByConnectorType {
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
			if !r.contains(typeMap[connType].Segments, route.SegmentName) {
				typeMap[connType].Segments = append(typeMap[connType].Segments, route.SegmentName)
			}

			// Track CXPs
			if !r.contains(typeMap[connType].CXPs, connector.Connector.ConnectorCXPName) {
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

func (r *RoutesClientHandler) summarizeBySegment(routes []RouteUIResult) map[string]int {
	segmentCount := make(map[string]int)
	for _, route := range routes {
		segmentCount[route.SegmentName]++
	}
	return segmentCount
}

func (r *RoutesClientHandler) summarizeByCXP(routes []RouteUIResult) map[string]int {
	cxpCount := make(map[string]int)
	for _, route := range routes {
		for _, connector := range route.Connectors {
			cxpCount[connector.Connector.ConnectorCXPName]++
		}
	}
	return cxpCount
}

func (r *RoutesClientHandler) analyzePrefixes(routes []RouteUIResult) *PrefixAnalysis {
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

func (r *RoutesClientHandler) contains(slice []string, item string) bool {
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
		tenantNetworkID, err := request.RequireString("tenantNetworkId")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Extract parameters
		routeType := request.GetString("type", "received")
		groupBy := request.GetString("groupBy", "connectorType")
		includeDetails := request.GetBool("includeDetails", false)
		includePrefixAnalysis := request.GetBool("includePrefixAnalysis", false)

		// Initialize routes client
		api := NewRoutesClientHandler(client)

		// Get all routes
		allRoutes, err := api.getAllRoutesWithPagination(api, tenantNetworkID, routeType)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Generate summary
		summary := api.generateRouteSummary(allRoutes, groupBy, includeDetails, includePrefixAnalysis)

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
		tenantNetworkID, err := request.RequireString("tenantNetworkId")
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

		// Initialize routes client
		api := NewRoutesClientHandler(client)

		// Get all routes with pagination
		allRoutes, err := api.getAllRoutesWithPaginationAndFiltering(api, tenantNetworkID, routeType, batchSize, segmentName, connectorType, cxp)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Format output
		switch strings.ToLower(outputFormat) {
		case "summary":
			summary := api.generateRouteSummary(allRoutes, "connectorType", false, true)
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
			result := RoutesUIResponse{
				Data: allRoutes,
				Pagination: PaginationData{
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

// GetRoutesByConnectorType handler
func GetRoutesByConnectorType(client *alkira.AlkiraClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract parameters
		tenantNetworkID, err := request.RequireString("tenantNetworkId")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		connectorType, err := request.RequireString("connectorType")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		routeType := request.GetString("type", "received")
		includePrefixSummary := request.GetBool("includePrefixSummary", false)
		outputFormat := request.GetString("outputFormat", "json")
		segmentName := request.GetString("segmentName", "")

		// Initialize routes client
		api := NewRoutesClientHandler(client)

		// Build enhanced parameters
		params := EnhancedRouteQueryParams{
			RouteQueryParams: RouteQueryParams{
				Type:        routeType,
				SegmentName: segmentName,
				Limit:       1000,
			},
			OutputFormat:   outputFormat,
			ConnectorTypes: connectorType,
		}

		// Get filtered routes
		result, err := api.GetRoutesEnhanced(tenantNetworkID, params)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Add prefix summary if requested
		if includePrefixSummary {
			if routes, ok := result.(*RoutesUIResponse); ok {
				summary := api.generateRouteSummary(routes.Data, "connectorType", false, true)
				enhancedResult := map[string]interface{}{
					"routes": routes,
					"summary": summary,
				}
				resultJSON, err := json.Marshal(enhancedResult)
				if err != nil {
					return mcp.NewToolResultError("Failed to marshal enhanced result: " + err.Error()), nil
				}
				return mcp.NewToolResultText(string(resultJSON)), nil
			}
		}

		// Return result
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal result: " + err.Error()), nil
		}

		return mcp.NewToolResultText(string(resultJSON)), nil
	}
}

// Helper functions for pagination
func (r *RoutesClientHandler) getAllRoutesWithPagination(api *RoutesClientHandler, tenantNetworkID, routeType string) ([]RouteUIResult, error) {
	var allRoutes []RouteUIResult
	offset := 0
	limit := 50

	for {
		params := RouteQueryParams{
			Type:   routeType,
			Offset: offset,
			Limit:  limit,
		}

		response, err := api.GetRoutes(tenantNetworkID, params)
		if err != nil {
			return nil, err
		}

		allRoutes = append(allRoutes, response.Data...)

		if len(response.Data) < limit {
			break
		}

		offset += limit
	}

	return allRoutes, nil
}

func (r *RoutesClientHandler) getAllRoutesWithPaginationAndFiltering(api *RoutesClientHandler, tenantNetworkID, routeType string, batchSize int, segmentName, connectorType, cxp string) ([]RouteUIResult, error) {
	var allRoutes []RouteUIResult
	offset := 0

	for {
		params := RouteQueryParams{
			Type:        routeType,
			Offset:      offset,
			Limit:       batchSize,
			SegmentName: segmentName,
			EntityType:  connectorType,
			CXP:         cxp,
		}

		response, err := api.GetRoutes(tenantNetworkID, params)
		if err != nil {
			return nil, err
		}

		allRoutes = append(allRoutes, response.Data...)


		if len(response.Data) < batchSize {
			break
		}

		offset += batchSize
	}

	return allRoutes, nil
}

