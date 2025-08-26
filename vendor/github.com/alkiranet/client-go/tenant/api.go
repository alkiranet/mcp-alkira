// Copyright (C) 2023-2025 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Pagination Support
const PaginationOn bool = true
const PaginationOff bool = false

// For each API defined as T, GET result with pagination On will have
// a wrapper layer around.
type Pagination struct {
	Paginated bool `json:"paginated"`
	Offset    int  `json:"Offset"`
	Limit     int  `json:"limit"`
}

type DataWithPagination[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Struct defines a Alkira API by resource T
type AlkiraApi[T any] struct {
	Client *AlkiraClient
	Uri    string
}

// Create create a resource by making a POST request
func (a *AlkiraApi[T]) Create(resource *T) (*T, error) {

	// Construct the request
	body, err := json.Marshal(resource)

	if err != nil {
		return nil, fmt.Errorf("api-create: failed to marshal: %v", err)
	}

	data, err := a.Client.Create(a.Uri, body)

	if err != nil {
		return nil, err
	}

	var result T
	err = json.Unmarshal([]byte(data), &result)

	if err != nil {
		return nil, fmt.Errorf("API-Create: failed to unmarshal: %v", err)
	}

	return &result, nil
}

// Delete delete a resource by its ID
func (a *AlkiraApi[T]) Delete(id string) error {
	uri := fmt.Sprintf("%s/%s", a.Uri, id)
	return a.Client.Delete(uri)
}

// Update update a resource by its ID
func (a *AlkiraApi[T]) Update(id string, resource *T) (string, error) {

	// Construct single resource URI
	uri := fmt.Sprintf("%s/%s", a.Uri, id)

	// Construct the request
	body, err := json.Marshal(resource)

	if err != nil {
		return "", fmt.Errorf("API-Update: failed to marshal: %v", err)
	}

	return a.Client.Update(uri, body)
}

// GetAll get all resources
func (a *AlkiraApi[T]) GetAll() (string, error) {
	data, err := a.Client.Get(a.Uri)

	if a.Client.MaxToken != 0 && len(data) > a.Client.MaxToken {
		return "{Payload too large. Please try other APIs}", nil
	}

	return string(data), err
}

// GetSummary get resource summary by stripping fields
//
// A special hacky function made to handle the super large payload
// that may exceed the max token (mainly used with AI agent).
func (a *AlkiraApi[T]) GetSummary(paginated string, offset string, limit string) (string, error) {

	uri, err := url.Parse(a.Uri)

	if err != nil {
		return "", fmt.Errorf("GetPaginated: failed to parse URI %s: %v", a.Uri, err)
	}

	//
	// Query parameters for pagination
	//
	q := uri.Query()
	q.Add("paginated", paginated)

	if offset != "" {
		q.Add("offset", offset)
	}

	if limit != "" {
		q.Add("limit", limit)
	}
	uri.RawQuery = q.Encode()

	data, err := a.Client.Get(uri.String())

	if err != nil {
		return "", fmt.Errorf("GetSummary: failed to GET data: %v", err)
	}

	logf("DEBUG", "GetSummary: payload size %d", len(data))

	if a.Client.MaxToken != 0 && len(data) > a.Client.MaxToken {
		return "{Payload too large. Please try other APIs}", nil
	}

	var result []T
	var resultPaginated DataWithPagination[T]

	// Pagination Support
	if paginated == "true" {
		err = json.Unmarshal([]byte(data), &resultPaginated)
	} else {
		err = json.Unmarshal([]byte(data), &result)
	}

	if err != nil {
		return "", fmt.Errorf("GetSummary: failed to unmarshal: %v", err)
	}

	// Marshal the summary data
	var summary []byte

	if paginated == "true" {
		summary, err = json.Marshal(resultPaginated)
	} else {
		summary, err = json.Marshal(result)
	}

	if err != nil {
		return "", fmt.Errorf("GetSummary: failed to marshal: %v", err)
	}

	return string(summary), nil
}

// GetAllPaginated get all resources
func (a *AlkiraApi[T]) GetAllPaginated(offset string, limit string) (string, error) {

	uri, err := url.Parse(a.Uri)

	if err != nil {
		return "", fmt.Errorf("GetPaginated: failed to parse URI %s: %v", a.Uri, err)
	}

	//
	// Query parameters for pagination
	//
	q := uri.Query()
	q.Add("paginated", "true")

	if offset != "" {
		q.Add("offset", offset)
	}

	if limit != "" {
		q.Add("limit", limit)
	}
	uri.RawQuery = q.Encode()

	data, err := a.Client.Get(uri.String())

	logf("DEBUG", "GetSummary: payload size %d", len(data))

	if a.Client.MaxToken != 0 && len(data) > a.Client.MaxToken {
		return "{Payload too large. Please try other APIs}", nil
	}

	return string(data), err
}

// GetById get a resource by its ID
func (a *AlkiraApi[T]) GetById(id string) (string, error) {

	if len(id) == 0 {
		return "", fmt.Errorf("API-GetByName: Invalid resource ID")
	}

	// Construct single resource URI
	uri := fmt.Sprintf("%s/%s?includeMarkedForDeletion=true", a.Uri, id)

	data, err := a.Client.Get(uri)

	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetByName get a resource by its name
func (a *AlkiraApi[T]) GetByName(name string) (string, error) {

	if len(name) == 0 {
		return "", fmt.Errorf("API-GetByName: Invalid resource name")
	}

	// Construct single resource URI
	uri := fmt.Sprintf("%s?name=%s&paginated=false", a.Uri, name)

	data, err := a.Client.Get(uri)

	if err != nil {
		return "", err
	}

	return string(data), nil
}
