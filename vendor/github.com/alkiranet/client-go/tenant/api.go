// Copyright (C) 2023-2025 Alkira Inc. All Rights Reserved.

// This file implment common functions for resource API by using
// generics. Each resource defined by type T should be able to use all
// those common functions.
//
// NOTE: for some special resources that needs some handling. The raw
// client functions in client.go could be still used directly in that
// case.
package tenant

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Pagination Support
//
// Most APIs follow this standard pagination structure. Please check
// the specific API carefully and there may be cases that some APIs
// don't follow it.
const PaginationOn bool = true
const PaginationOff bool = false

type Pagination struct {
	Paginated bool `json:"paginated,omitempty"`
	Offset    int  `json:"Offset"`
	Limit     int  `json:"limit"`
	Hits      int  `json:"hits,omitempty"`
}

// For each API defined as T, GET result with pagination On will have
// a wrapper layer around.
type DataWithPagination[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Struct defines a Alkira API by resource T
type AlkiraApi[T any] struct {
	Client     *AlkiraClient
	Uri        string
	Pagination bool
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
//
// This function may return big payload with many resources,
// especially in scaled tenant. Even with pagination support, this
// function should be avoided to use in most cases.
func (a *AlkiraApi[T]) GetAll(offset string, limit string) (string, error) {

	uri, err := url.Parse(a.Uri)

	if err != nil {
		return "", fmt.Errorf("GetAll: failed to parse URI %s: %v", a.Uri, err)
	}

	//
	// Query parameters for pagination
	//
	q := uri.Query()

	if a.Pagination == PaginationOn {
		q.Add("paginated", "true")
	}
	if offset != "" {
		q.Add("offset", offset)
	}
	if limit != "" {
		q.Add("limit", limit)
	}
	uri.RawQuery = q.Encode()

	data, err := a.Client.Get(uri.String())

	logf("DEBUG", "GetAll: payload size %d", len(data))

	if a.Client.MaxToken != 0 && len(data) > a.Client.MaxToken {
		return "{Payload too large. Please try other APIs}", nil
	}

	return string(data), err
}

// GetSummary get resource summary by stripping fields
//
// A special hacky function made to handle the super large payload
// that may exceed the max token (mainly used with AI agent).
func (a *AlkiraApi[T]) GetSummary(offset string, limit string) (string, error) {

	uri, err := url.Parse(a.Uri)

	if err != nil {
		return "", fmt.Errorf("GetSummary: failed to parse URI %s: %v", a.Uri, err)
	}

	//
	// Query parameters for pagination
	//
	q := uri.Query()

	if a.Pagination == PaginationOn {
		q.Add("paginated", "true")
	}
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
	if a.Pagination == PaginationOn {
		err = json.Unmarshal([]byte(data), &resultPaginated)
	} else {
		err = json.Unmarshal([]byte(data), &result)
	}

	if err != nil {
		return "", fmt.Errorf("GetSummary: failed to unmarshal: %v", err)
	}

	// Marshal the summary data
	var summary []byte

	if a.Pagination == PaginationOn {
		summary, err = json.Marshal(resultPaginated)
	} else {
		summary, err = json.Marshal(result)
	}

	if err != nil {
		return "", fmt.Errorf("GetSummary: failed to marshal: %v", err)
	}

	return string(summary), nil
}

// GetById get a resource by its ID
func (a *AlkiraApi[T]) GetById(id string) (string, error) {

	if len(id) == 0 {
		return "", fmt.Errorf("GetById: Invalid resource ID")
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
		return "", fmt.Errorf("GetByName: Invalid resource name")
	}

	// Construct single resource URI
	uri := fmt.Sprintf("%s?name=%s&paginated=false", a.Uri, name)

	data, err := a.Client.Get(uri)

	if err != nil {
		return "", err
	}

	return string(data), nil
}

type ResourceCount struct {
	Total int `json:"total"`
}

// GetCount
//
// A special hacky function to use the pagination block retrieve the
// total count of the given resource.
func (a *AlkiraApi[T]) GetCount() (string, error) {

	uri, err := url.Parse(a.Uri)

	if err != nil {
		return "", fmt.Errorf("GetCount: failed to parse URI %s: %v", a.Uri, err)
	}

	//
	// Query parameters for pagination
	//
	q := uri.Query()

	if a.Pagination == PaginationOff {
		return "", fmt.Errorf("GetCount: couldn't get resource count since pagination is not supported.")
	}

	q.Add("paginated", "true")
	q.Add("offset", "0")
	q.Add("limit", "1")

	uri.RawQuery = q.Encode()

	data, err := a.Client.Get(uri.String())

	if err != nil {
		return "", fmt.Errorf("GetCount: failed to GET data: %v", err)
	}

	var resultPaginated DataWithPagination[T]
	err = json.Unmarshal([]byte(data), &resultPaginated)

	if err != nil {
		return "", fmt.Errorf("GetCount: failed to unmarshal: %v", err)
	}

	count, err := json.Marshal(ResourceCount{Total: resultPaginated.Pagination.Hits})

	if err != nil {
		return "", fmt.Errorf("GetCount: failed to marshal: %v", err)
	}

	return string(count), nil
}
