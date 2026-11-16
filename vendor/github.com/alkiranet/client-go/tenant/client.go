// Copyright (C) 2020-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
)

// Default variables for retrying
const defaultRetryInterval time.Duration = 5 * time.Second
const defaultRetryTimeout time.Duration = 10 * time.Second

// portalScheme prefixes the uri passed to NewAlkiraClient; tests point it at plain HTTP.
var portalScheme = "https://"

// This structure defines an Alkira Client.
//
// Client        - An retryable HTTP client.
// URI           - Alkira Portal URI.
// ApiKey        - Alkira API Key to authenticate.
// Authorization - Authorization header.
// MaxToken      - Max tokens for LLM usages.
// Compression   - Compression method for optimizting output for LLMs.
// HostHeader    - Host header sent when non-empty; TLS SNI and verification are unchanged.
//
// Hold it by pointer. A value copy stored in an unexported field of another
// struct bypasses the String and GoString redaction under %+v.
type AlkiraClient struct {
	Client          *retryablehttp.Client
	URI             string
	ApiKey          string
	TenantNetworkId string
	Authorization   string
	MaxToken        int
	Compression     string
	HostHeader      string
	extraHeaders    map[string]string
	configErr       error
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
func NewAlkiraClient(uri string, apiKey string, maxToken int, compression string) (*AlkiraClient, error) {

	logf("DEBUG", "Creating new Alkira Client")

	// Construct portal URLs
	url := portalScheme + uri
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

	client := &AlkiraClient{
		Client:        retryClient,
		URI:           apiUrl,
		ApiKey:        apiKey,
		Authorization: auth,
		MaxToken:      maxToken,
		Compression:   compression,
	}

	tenantNetworkId, err := client.LookupTenantNetworkId(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant network ID: %w", err)
	}
	client.TenantNetworkId = tenantNetworkId

	return client, nil
}

// Get retrieve resources by sending a HTTP GET request
func (ac *AlkiraClient) Get(uri string) ([]byte, error) {
	return ac.get(uri, 0)
}

// get is Get that, when maxBody > 0, reads at most maxBody+1 bytes of a 200
// body, so the caller can reject a larger one without buffering it.
func (ac *AlkiraClient) get(uri string, maxBody int64) ([]byte, error) {
	if err := ac.check(uri); err != nil {
		return nil, err
	}
	logf("DEBUG", "client-get URI: %s\n", uri)

	requestId := "client-" + uuid.New().String()
	request, err := retryablehttp.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("client-get: invalid request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", ac.Authorization)
	request.Header.Set("x-ak-request-id", requestId)
	ac.decorate(request)
	requestId = request.Header.Get("x-ak-request-id")

	response, err := ac.Client.Do(request)

	if err != nil {
		if response != nil {
			return nil, responseError("client-get("+requestId+")", response, err,
				fmt.Sprintf("client-get(%s) failed to send request, %v", requestId, err))
		}
		return nil, fmt.Errorf("client-get(%s) failed to send request, %w", requestId, err)
	}

	defer response.Body.Close()
	body := io.Reader(response.Body)
	if maxBody > 0 && response.StatusCode == 200 {
		body = io.LimitReader(body, maxBody+1)
	}
	data, _ := ioutil.ReadAll(body)
	logf("DEBUG", "client-get(%s) %d RSP: %s", requestId, response.StatusCode, string(data))

	if response.StatusCode != 200 {
		return nil, newHTTPError("client-get("+requestId+")", response.StatusCode, data, nil, fmt.Sprintf("client-get(%s): %d %s", requestId, response.StatusCode, string(data)))
	}

	return data, nil
}

// Create send a HTTP POST request to create resource
func (ac *AlkiraClient) Create(uri string, body []byte) ([]byte, error) {
	if err := ac.check(uri); err != nil {
		return nil, err
	}

	logf("DEBUG", "client-create REQ: %s", string(body))

	requestId := "client-" + uuid.New().String()
	request, err := retryablehttp.NewRequest("POST", uri, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("client-create: invalid request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", ac.Authorization)
	request.Header.Set("x-ak-request-id", requestId)
	ac.decorate(request)
	requestId = request.Header.Get("x-ak-request-id")

	response, err := ac.Client.Do(request)

	if err != nil {
		if response != nil {
			return nil, responseError("client-create("+requestId+")", response, err,
				fmt.Sprintf("client-create(%s): failed to send request, %v", requestId, err))
		}
		return nil, fmt.Errorf("client-create(%s): failed to send request, %w", requestId, err)
	}

	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)

	logf("DEBUG", "client-create(%s) %d RSP: %s", requestId, response.StatusCode, string(data))

	if response.StatusCode != 201 && response.StatusCode != 200 {
		return nil, newHTTPError("client-create("+requestId+")", response.StatusCode, data, nil, fmt.Sprintf("client-create(%s): %d %s.", requestId, response.StatusCode, string(data)))
	}

	return data, nil
}

// delete send a DELETE request to delete a resource
func (ac *AlkiraClient) Delete(uri string) error {
	if err := ac.check(uri); err != nil {
		return err
	}
	logf("DEBUG", "client-delete: URI %s\n", uri)

	requestId := "client-" + uuid.New().String()
	request, err := retryablehttp.NewRequest("DELETE", uri, nil)
	if err != nil {
		return fmt.Errorf("client-delete: invalid request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", ac.Authorization)
	request.Header.Set("x-ak-request-id", requestId)
	ac.decorate(request)
	requestId = request.Header.Get("x-ak-request-id")

	response, err := ac.Client.Do(request)

	if err != nil {
		if response != nil {
			return responseError("client-delete("+requestId+")", response, err,
				fmt.Sprintf("client-delete(%s): failed to send request, %v", requestId, err))
		}
		return fmt.Errorf("client-delete(%s): failed to send request, %w", requestId, err)
	}

	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)

	logf("DEBUG", "client-delete(%s): %d RSP: %s\n", requestId, response.StatusCode, string(data))

	if response.StatusCode < 200 || response.StatusCode > 299 {
		if response.StatusCode == 404 {
			logf("INFO", "client-delete(%s): %d resource was already deleted.\n", requestId, response.StatusCode)
			return nil
		}

		return newHTTPError("client-delete("+requestId+")", response.StatusCode, data, nil, fmt.Sprintf("client-delete(%s): %d %s", requestId, response.StatusCode, string(data)))
	}

	return nil
}

// Update send a PUT API request to update a resource
func (ac *AlkiraClient) Update(uri string, body []byte) (string, error) {
	if err := ac.check(uri); err != nil {
		return "", err
	}

	logf("DEBUG", "client-update: REQ: %s\n", string(body))

	requestId := "client-" + uuid.New().String()
	request, err := retryablehttp.NewRequest("PUT", uri, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("client-update: invalid request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", ac.Authorization)
	request.Header.Set("x-ak-request-id", requestId)
	ac.decorate(request)
	requestId = request.Header.Get("x-ak-request-id")

	response, err := ac.Client.Do(request)

	if err != nil {
		if response != nil {
			return "", responseError("client-update("+requestId+")", response, err,
				fmt.Sprintf("client-update(%s): failed to send request, %v", requestId, err))
		}
		return "", fmt.Errorf("client-update(%s): failed to send request, %w", requestId, err)
	}

	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)

	logf("DEBUG", "client-update(%s): %d RSP: %v\n", requestId, response.StatusCode, data)

	if response.StatusCode != 200 && response.StatusCode != 202 {
		return "", newHTTPError("client-update("+requestId+")", response.StatusCode, data, nil, fmt.Sprintf("client-update(%s): %d %s", requestId, response.StatusCode, string(data)))
	}

	return "", nil
}
