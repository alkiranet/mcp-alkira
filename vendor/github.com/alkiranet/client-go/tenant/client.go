// Copyright (C) 2020-2025 Alkira Inc. All Rights Reserved.

package tenant

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
)

// Default variables for retrying
const defaultProvTimeout time.Duration = 240 * time.Minute
const defaultRetryInterval time.Duration = 5 * time.Second
const defaultRetryTimeout time.Duration = 10 * time.Second

// This structure defines the basic elements of an Alkira Client.
//
// Client an retryable HTTP client
// URI The URI to make to API request to
type AlkiraClient struct {
	Client          *retryablehttp.Client
	URI             string
	ApiKey          string
	TenantNetworkId string
	Authorization   string
	MaxToken        int
}

// NewAlkiraClient creates a new alkira client
//
// This function create a new Alkira Client that could be used to make
// interaction with Alkira Portal. It will authenticate with portal with the
// API key and retrieve Tenant Network ID automically that will be needed for
// any further API calls.
//
// uri Alkira portal URI
// apiKey Alkira API key to be used to authenticate
// maxToken Max token size if response payload is too big (for AI agent)
func NewAlkiraClient(uri string, apiKey string, maxToken int) (*AlkiraClient, error) {

	logf("DEBUG", "Creating new Alkira Client")

	// Construct portal URLs
	url := "https://" + uri
	apiUrl := url + "/api"

	// Generate Authorization header string
	auth := "api-key " + base64.StdEncoding.EncodeToString([]byte(apiKey))

	// Create retry-able HTTP client
	tr := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{},
	}

	// Config retry client
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient.Transport = tr
	retryClient.RetryMax = 5
	retryClient.Backoff = func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
		// Respect server-specified Retry-After header
		if resp != nil && resp.StatusCode == 429 {
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, err := strconv.Atoi(retryAfter); err == nil {
					logf("DEBUG", "Retry-After: %v\n", seconds)
					return time.Duration(seconds) * time.Second
				}
			}
		}
		return retryablehttp.LinearJitterBackoff(min, max, attemptNum, resp)
	}
	retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		shouldRetry, e := retryablehttp.DefaultRetryPolicy(ctx, resp, err)

		// Always retry on 429 and 500 status codes
		if resp != nil {
			switch resp.StatusCode {
			case 429, 500:
				return true, fmt.Errorf("retryable status code: %d", resp.StatusCode)
			}
		}
		return shouldRetry, e
	}

	// Get the tenant network ID
	var result []TenantNetworkId
	tenantNetworkUrl := apiUrl + "/tenantnetworksummaries"

	tenantNetworkRequest, _ := retryablehttp.NewRequest("GET", tenantNetworkUrl, nil)
	tenantNetworkRequest.Header.Set("Content-Type", "application/json")
	tenantNetworkRequest.Header.Set("Authorization", auth)
	tenantNetworkResponse, err := retryClient.Do(tenantNetworkRequest)

	if err != nil {
		return nil, fmt.Errorf("failed to make tenant network request, %v", err)
	}

	defer tenantNetworkResponse.Body.Close()

	data, _ := ioutil.ReadAll(tenantNetworkResponse.Body)
	logf("TRACE", "Tenant Network Summary: %s\n", string(data))

	if tenantNetworkResponse.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get tenant network (%d)", tenantNetworkResponse.StatusCode)
	}

	json.Unmarshal([]byte(data), &result)

	tenantNetworkId := 0

	if len(result) > 0 {
		tenantNetworkId = result[0].Id
	} else {
		return nil, fmt.Errorf("failed to get tenant network ID")
	}

	// Construct our client with all information
	client := &AlkiraClient{
		Client:          retryClient,
		URI:             apiUrl,
		ApiKey:          apiKey,
		TenantNetworkId: strconv.Itoa(tenantNetworkId),
		Authorization:   auth,
		MaxToken:        maxToken,
	}

	return client, nil
}

// Get retrieve a resource by sending a GET request
func (ac *AlkiraClient) Get(uri string) ([]byte, error) {
	logf("DEBUG", "client-get URI: %s\n", uri)

	requestId := "client-" + uuid.New().String()
	request, _ := retryablehttp.NewRequest("GET", uri, nil)

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", ac.Authorization)
	request.Header.Set("x-ak-request-id", requestId)

	response, err := ac.Client.Do(request)

	if err != nil {
		return nil, fmt.Errorf("client-get(%s) failed to send request, %v", requestId, err)
	}

	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)
	logf("DEBUG", "client-get(%s) %d RSP: %s", requestId, response.StatusCode, string(data))

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("client-get(%s): %d %s", requestId, response.StatusCode, string(data))
	}

	return data, nil
}

// create send a POST request to create resource
func (ac *AlkiraClient) Create(uri string, body []byte) ([]byte, error) {

	logf("DEBUG", "client-create REQ: %s", string(body))

	requestId := "client-" + uuid.New().String()
	request, _ := retryablehttp.NewRequest("POST", uri, bytes.NewBuffer(body))

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", ac.Authorization)
	request.Header.Set("x-ak-request-id", requestId)

	response, err := ac.Client.Do(request)

	if err != nil {
		return nil, fmt.Errorf("client-create(%s): failed to send request, %v", requestId, err)
	}

	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)

	logf("DEBUG", "client-create(%s) %d RSP: %s", requestId, response.StatusCode, string(data))

	if response.StatusCode != 201 && response.StatusCode != 200 {
		return nil, fmt.Errorf("client-create(%s): %d %s.", requestId, response.StatusCode, string(data))
	}

	return data, nil
}

// delete send a DELETE request to delete a resource
func (ac *AlkiraClient) Delete(uri string) error {
	logf("DEBUG", "client-delete: URI %s\n", uri)

	requestId := "client-" + uuid.New().String()
	request, _ := retryablehttp.NewRequest("DELETE", uri, nil)

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", ac.Authorization)
	request.Header.Set("x-ak-request-id", requestId)

	response, err := ac.Client.Do(request)

	if err != nil {
		return fmt.Errorf("client-delete(%s): failed to send request, %v", requestId, err)
	}

	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)

	logf("DEBUG", "client-delete(%s): %d RSP: %s\n", requestId, response.StatusCode, string(data))

	if response.StatusCode < 200 || response.StatusCode > 299 {
		if response.StatusCode == 404 {
			logf("INFO", "client-delete(%s): %d resource was already deleted.\n", requestId, response.StatusCode)
			return nil
		}

		return fmt.Errorf("client-delete(%s): %d %s", requestId, response.StatusCode, string(data))
	}

	return nil
}

// Update send a PUT API request to update a resource
func (ac *AlkiraClient) Update(uri string, body []byte) (string, error) {

	logf("DEBUG", "client-update: REQ: %s\n", string(body))

	requestId := "client-" + uuid.New().String()
	request, _ := retryablehttp.NewRequest("PUT", uri, bytes.NewBuffer(body))

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", ac.Authorization)
	request.Header.Set("x-ak-request-id", requestId)

	response, err := ac.Client.Do(request)

	if err != nil {
		return "", fmt.Errorf("client-update(%s): failed to send request, %v", requestId, err)
	}

	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)

	logf("DEBUG", "client-update(%s): %d RSP: %v\n", requestId, response.StatusCode, data)

	if response.StatusCode != 200 && response.StatusCode != 202 {
		return "", fmt.Errorf("client-update(%s): %d %s", requestId, response.StatusCode, string(data))
	}

	return "", nil
}
