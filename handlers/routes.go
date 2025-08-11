package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

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

// Routes API client for handlers
type RoutesClientHandler struct {
	Client *alkira.AlkiraClient
}

// NewRoutesClientHandler creates a new Routes client for handlers
func NewRoutesClientHandler(client *alkira.AlkiraClient) *RoutesClientHandler {
	return &RoutesClientHandler{
		Client: client,
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

		// Build query parameters from request
		params := RouteQueryParams{
			Type:                 request.GetString("type", ""),
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
		}

		// Get routes
		routes, err := api.GetRoutes(tenantNetworkID, params)
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
			Type:                 request.GetString("type", ""),
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

