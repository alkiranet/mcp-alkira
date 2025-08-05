package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/alkiranet/alkira-client-go/alkira"
	"github.com/mark3labs/mcp-go/mcp"
)

type GetAllAPI interface {
	GetAll() (string, error)
}

type PaginationParams struct {
	Limit  int
	Offset int
	Page   int
}

func CreateGetAllHandler(newAPI func(*alkira.AlkiraClient) GetAllAPI) func(*alkira.AlkiraClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(client *alkira.AlkiraClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			api := newAPI(client)
			resources, err := api.GetAll()
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(resources), nil
		}
	}
}

func CreatePaginatedGetAllHandler(newAPI func(*alkira.AlkiraClient) GetAllAPI) func(*alkira.AlkiraClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(client *alkira.AlkiraClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Parse pagination parameters
			pagination, err := parsePaginationParams(request)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Pagination parameter error: %s", err.Error())), nil
			}

			api := newAPI(client)
			resources, err := api.GetAll()
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Apply pagination if requested
			if pagination != nil {
				paginatedResult, err := applyPagination(resources, *pagination)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("Pagination error: %s", err.Error())), nil
				}
				return mcp.NewToolResultText(paginatedResult), nil
			}

			// Check if response might be too large
			if len(resources) > 50000 { // Rough estimate for token limit
				return mcp.NewToolResultError("Response too large. Please use pagination parameters: limit, offset, or page. Example: --limit 100"), nil
			}

			return mcp.NewToolResultText(resources), nil
		}
	}
}

func parsePaginationParams(request mcp.CallToolRequest) (*PaginationParams, error) {
	var pagination *PaginationParams

	// Check for limit parameter
	if limitStr, err := request.RequireString("limit"); err == nil && limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, fmt.Errorf("invalid limit parameter: %s", limitStr)
		}
		if limit <= 0 {
			return nil, fmt.Errorf("limit must be greater than 0")
		}
		
		if pagination == nil {
			pagination = &PaginationParams{}
		}
		pagination.Limit = limit
	}

	// Check for offset parameter
	if offsetStr, err := request.RequireString("offset"); err == nil && offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return nil, fmt.Errorf("invalid offset parameter: %s", offsetStr)
		}
		if offset < 0 {
			return nil, fmt.Errorf("offset must be non-negative")
		}
		
		if pagination == nil {
			pagination = &PaginationParams{}
		}
		pagination.Offset = offset
	}

	// Check for page parameter
	if pageStr, err := request.RequireString("page"); err == nil && pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return nil, fmt.Errorf("invalid page parameter: %s", pageStr)
		}
		if page <= 0 {
			return nil, fmt.Errorf("page must be greater than 0")
		}
		
		if pagination == nil {
			pagination = &PaginationParams{}
		}
		pagination.Page = page
	}

	return pagination, nil
}

func applyPagination(jsonData string, pagination PaginationParams) (string, error) {
	// Parse JSON array
	var items []interface{}
	if err := json.Unmarshal([]byte(jsonData), &items); err != nil {
		return "", fmt.Errorf("failed to parse JSON data: %w", err)
	}

	totalItems := len(items)
	
	// Set default limit if not specified
	limit := pagination.Limit
	if limit == 0 {
		limit = 100 // Default limit
	}

	// Calculate offset
	offset := pagination.Offset
	if pagination.Page > 0 {
		// Page-based pagination (1-indexed)
		offset = (pagination.Page - 1) * limit
	}

	// Validate offset
	if offset >= totalItems {
		// Return empty array if offset is beyond data
		emptyResult, _ := json.Marshal([]interface{}{})
		return string(emptyResult), nil
	}

	// Calculate end index
	end := offset + limit
	if end > totalItems {
		end = totalItems
	}

	// Slice the data
	paginatedItems := items[offset:end]

	// Create response with metadata
	response := map[string]interface{}{
		"data": paginatedItems,
		"pagination": map[string]interface{}{
			"total":  totalItems,
			"limit":  limit,
			"offset": offset,
			"count":  len(paginatedItems),
		},
	}

	if pagination.Page > 0 {
		response["pagination"].(map[string]interface{})["page"] = pagination.Page
		response["pagination"].(map[string]interface{})["totalPages"] = (totalItems + limit - 1) / limit
	}

	result, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal paginated result: %w", err)
	}

	return string(result), nil
}

func CreateHandlerWithOptionalParams(handler func(*alkira.AlkiraClient, map[string]string) (string, error), paramNames ...string) func(*alkira.AlkiraClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(client *alkira.AlkiraClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			params := make(map[string]string)
			for _, paramName := range paramNames {
				value, err := request.RequireString(paramName)
				if err != nil {
					value = ""
				}
				params[paramName] = value
			}

			result, err := handler(client, params)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		}
	}
}

func CreateHandlerWithRequiredParams(handler func(*alkira.AlkiraClient, ...string) (string, error), paramNames ...string) func(*alkira.AlkiraClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(client *alkira.AlkiraClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var params []string
			for _, paramName := range paramNames {
				value, err := request.RequireString(paramName)
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				params = append(params, value)
			}

			result, err := handler(client, params...)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		}
	}
}

func CreateSimpleHandler(handler func(*alkira.AlkiraClient) (string, error)) func(*alkira.AlkiraClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(client *alkira.AlkiraClient) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			result, err := handler(client)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		}
	}
}