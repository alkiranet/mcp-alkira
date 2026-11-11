// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
)

const (
	maxErrorBody  = 4 << 10  // bytes kept in HTTPError.Body
	maxLookupBody = 1 << 20  // larger 2xx lookup bodies are rejected
	maxDrain      = 64 << 10 // bytes read before Close so the connection can be reused
)

// ErrNoTenantNetwork is returned by LookupTenantNetworkId when the tenant has no tenant network.
var ErrNoTenantNetwork = errors.New("tenant has no tenant network")

// ErrEmptyToken is returned, without sending a request, by every call on a
// client built by NewAlkiraClientWithToken with an empty or blank token.
var ErrEmptyToken = errors.New("empty bearer token")

// ErrInvalidTenantNetworkId is returned, without sending a request, by every
// call on a client built by NewAlkiraClientWithToken with a tenantNetworkId
// that is neither empty nor a positive integer.
var ErrInvalidTenantNetworkId = errors.New("invalid tenant network id")

// ErrNilHTTPClient is returned, without sending a request, by every call on a
// client built by NewAlkiraClientWithToken with a nil httpClient.
var ErrNilHTTPClient = errors.New("nil http client")

// ErrInvalidHostHeader is returned, without sending a request, by every call on
// a client built by NewAlkiraClientWithToken with a hostHeader that is not a
// valid Host header value.
var ErrInvalidHostHeader = errors.New("invalid host header")

// ErrInvalidTenantNetworkResponse is wrapped by LookupTenantNetworkId errors for
// a 2xx response that is not valid JSON, exceeds 1 MiB, or whose first entry has
// a missing or non-positive id.
var ErrInvalidTenantNetworkResponse = errors.New("invalid tenant network response")

// HTTPError reports a response with an unexpected status. LookupTenantNetworkId
// returns it for a non-2xx status, and Get, Create, Update and Delete for any
// status they treat as a failure.
type HTTPError struct {
	Op     string // "tenant network lookup", or "client-<method>(<request id>)"
	Status int
	Body   string // at most the first 4 KiB of the response body
	Err    error  // error returned together with the response, if any
	text   string // full message for Get, Create, Update and Delete
}

// Error returns the message. For LookupTenantNetworkId it omits Body; for Get,
// Create, Update and Delete it is the v0.5.5 message, which includes the whole body.
func (e *HTTPError) Error() string {
	if e.text != "" {
		return e.text
	}
	return fmt.Sprintf("%s: HTTP %d", e.Op, e.Status)
}

// Unwrap returns Err.
func (e *HTTPError) Unwrap() error {
	return e.Err
}

func newHTTPError(op string, status int, body []byte, err error, text string) *HTTPError {
	if len(body) > maxErrorBody {
		body = body[:maxErrorBody]
	}
	return &HTTPError{Op: op, Status: status, Body: string(body), Err: err, text: text}
}

// responseError builds an *HTTPError from a response that retryablehttp
// returned together with err, and closes the response body.
func responseError(op string, resp *http.Response, err error, text string) *HTTPError {
	defer drainClose(resp.Body)
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
	return newHTTPError(op, resp.StatusCode, body, err, text)
}

// NewAlkiraClientWithToken creates an Alkira client that authenticates with a bearer token.
//
// It performs no I/O. baseURL is the API base including "/api". hostHeader is
// sent as the Host header when non-empty. bearerToken is the raw token without
// the "bearer " scheme; surrounding whitespace is trimmed, and a blank token
// makes every call return ErrEmptyToken. tenantNetworkId is empty or a positive
// integer; any other value makes every call return ErrInvalidTenantNetworkId.
// A nil httpClient makes every call return ErrNilHTTPClient, and a hostHeader
// with characters not allowed in a Host header makes every call return
// ErrInvalidHostHeader.
//
// Get, Create, Update and Delete take no context, so httpClient's transport
// must enforce request deadlines. Set httpClient.ErrorHandler to
// retryablehttp.PassthroughErrorHandler to receive *HTTPError after retries
// are exhausted. Callers that share one http.Transport across clients must wrap
// it in a RoundTripper without CloseIdleConnections: retryablehttp calls it
// after every failed request, which would empty the shared pool.
func NewAlkiraClientWithToken(baseURL, hostHeader, bearerToken, tenantNetworkId string,
	maxToken int, compression string, httpClient *retryablehttp.Client) *AlkiraClient {

	bearerToken = strings.TrimSpace(bearerToken)

	ac := &AlkiraClient{
		Client:          httpClient,
		URI:             strings.TrimRight(baseURL, "/"),
		TenantNetworkId: tenantNetworkId,
		Authorization:   "bearer " + bearerToken,
		MaxToken:        maxToken,
		Compression:     compression,
		HostHeader:      hostHeader,
	}
	switch {
	case httpClient == nil:
		ac.configErr = ErrNilHTTPClient
	case bearerToken == "":
		ac.configErr = ErrEmptyToken
	case !validTenantNetworkId(tenantNetworkId):
		ac.configErr = ErrInvalidTenantNetworkId
	case !validHostHeader(hostHeader):
		ac.configErr = ErrInvalidHostHeader
	}
	return ac
}

// validHostHeader reports whether h contains only the bytes net/http accepts
// in a Host header (the httpguts.ValidHostHeader set). net/http sends an
// invalid Host as an empty Host instead of failing the request.
func validHostHeader(h string) bool {
	for i := 0; i < len(h); i++ {
		c := h[i]
		if !('a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' ||
			strings.IndexByte("!$%&'()*+,-.:;=[]_~", c) >= 0) {
			return false
		}
	}
	return true
}

// validTenantNetworkId reports whether id is empty or a positive decimal
// integer, the only values safe to place in a URL path unescaped.
func validTenantNetworkId(id string) bool {
	if id == "" {
		return true
	}
	for _, c := range id {
		if c < '0' || c > '9' {
			return false
		}
	}
	n, err := strconv.Atoi(id)
	return err == nil && n > 0
}

// WithHeader sets a header sent on every request, overriding generated headers
// such as x-ak-request-id, and returns ac.
//
// Names are canonicalized. Authorization and Host cannot be overridden; set
// them through the constructor. The header map is copied on each call, so a
// copy of the struct can be configured independently. Not safe to call
// concurrently with requests on the same client.
func (ac *AlkiraClient) WithHeader(name, value string) *AlkiraClient {
	headers := make(map[string]string, len(ac.extraHeaders)+1)
	for k, v := range ac.extraHeaders {
		headers[k] = v
	}
	headers[http.CanonicalHeaderKey(name)] = value
	ac.extraHeaders = headers
	return ac
}

// decorate applies HostHeader and the WithHeader headers. It must run after
// all other headers are set so caller-supplied values win.
func (ac *AlkiraClient) decorate(req *retryablehttp.Request) {
	if ac.HostHeader != "" {
		req.Host = ac.HostHeader
	}
	for k, v := range ac.extraHeaders {
		if k == "Authorization" {
			continue
		}
		req.Header.Set(k, v)
	}
}

// LookupTenantNetworkId returns the id of the tenant's first tenant network.
// The result is not stored. Errors: *HTTPError for a non-2xx response,
// ErrNoTenantNetwork for an empty list, ErrInvalidTenantNetworkResponse for a
// malformed 2xx response, otherwise the constructor or transport error.
func (ac *AlkiraClient) LookupTenantNetworkId(ctx context.Context) (string, error) {
	if ac.configErr != nil {
		return "", ac.configErr
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", ac.URI+"/tenantnetworksummaries", nil)
	if err != nil {
		return "", fmt.Errorf("tenant network lookup: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", ac.Authorization)
	req.Header.Set("x-ak-request-id", "client-"+uuid.New().String())
	ac.decorate(req)

	resp, err := ac.Client.Do(req)
	if resp != nil {
		defer drainClose(resp.Body)
	}

	// PassthroughErrorHandler can return a response together with an error;
	// the status is the more useful signal.
	if resp != nil && (resp.StatusCode < 200 || resp.StatusCode > 299) {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		return "", newHTTPError("tenant network lookup", resp.StatusCode, body, err, "")
	}

	if err != nil {
		return "", fmt.Errorf("tenant network lookup: %w", err)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxLookupBody+1))
	if err != nil {
		return "", fmt.Errorf("tenant network lookup: reading body: %w", err)
	}

	if len(data) > maxLookupBody {
		return "", fmt.Errorf("%w: body exceeds %d bytes", ErrInvalidTenantNetworkResponse, maxLookupBody)
	}

	var result []TenantNetworkId
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidTenantNetworkResponse, err)
	}

	if len(result) == 0 {
		return "", ErrNoTenantNetwork
	}

	if result[0].Id <= 0 {
		return "", fmt.Errorf("%w: id %d", ErrInvalidTenantNetworkResponse, result[0].Id)
	}

	return strconv.Itoa(result[0].Id), nil
}

// drainClose reads up to maxDrain bytes so the connection can return to the
// pool, then closes body. Larger bodies close the connection instead.
func drainClose(body io.ReadCloser) {
	io.Copy(io.Discard, io.LimitReader(body, maxDrain))
	body.Close()
}
