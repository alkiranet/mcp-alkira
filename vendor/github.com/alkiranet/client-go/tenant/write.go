// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
)

// Classified write errors. Test with errors.Is; *APIError, *StateError,
// *DuplicateError and *OutcomeError carry the details.
var (
	ErrNotFound       = errors.New("not found")
	ErrInUse          = errors.New("still in use")
	ErrDuplicate      = errors.New("already exists")
	ErrStateRefused   = errors.New("refused in the current state")
	ErrFamilyMismatch = errors.New("id belongs to another family")
	ErrOutcomeUnknown = errors.New("outcome unknown")
	ErrRateLimited    = errors.New("rate limited")
	ErrLimitExceeded  = errors.New("tenant limit reached")
	ErrInvalidInput   = errors.New("invalid input")
	// ErrValidationPending: TPS refuses a change while an async validation
	// of the entity is still running.
	ErrValidationPending = errors.New("validation still running")
	// ErrPostSend wraps a transport failure after the request was sent
	// (connection closed, reset or EOF); the write may have been applied.
	ErrPostSend = errors.New("connection closed after the request was sent")
	// ErrNotSent wraps a failure before the request left the client (no
	// connection, TLS handshake, deadline already passed); nothing was applied.
	ErrNotSent = errors.New("request not sent")

	// errDeadlineInFlight marks a context that ended while a write was in
	// flight; the write may have been applied.
	errDeadlineInFlight = errors.New("deadline reached while the request was in flight")
)

// WriteUserAgent is the User-Agent of write-path requests (Core and its
// confirming GETs). Set it to "mcp-alkira/<version>" before the first call.
var WriteUserAgent = "mcp-alkira/dev"

// WriteTimeoutEnv names the variable holding the default per-call write
// deadline, in seconds or as a Go duration.
const WriteTimeoutEnv = "AK_MCP_WRITE_TIMEOUT"

// DefaultWriteTimeout applies when the context has no deadline and
// AK_MCP_WRITE_TIMEOUT is unset or invalid.
const DefaultWriteTimeout = 60 * time.Second

// maxWriteRetryAfter is the longest Retry-After a write waits for before its one retry.
const maxWriteRetryAfter = 30 * time.Second

// writeDialBackoff is the wait before retrying a write whose connection was
// never established.
var writeDialBackoff = time.Second

const maxWriteBody = 16 << 20

var (
	requestRun = strings.SplitN(uuid.New().String(), "-", 2)[0]
	requestSeq atomic.Uint64
)

// nextRequestId returns mcp-<run>-<seq>: run is fixed per process, seq counts calls.
func nextRequestId() string {
	return fmt.Sprintf("mcp-%s-%d", requestRun, requestSeq.Add(1))
}

// ConfigStatus is the configStatus block of a save response. A save with
// ConfigValid false is stored; Messages say what the backend flagged.
type ConfigStatus struct {
	ConfigValid *bool    `json:"configValid,omitempty"`
	Messages    []string `json:"messages,omitempty"`
}

// WriteResult is the outcome of a successful write.
type WriteResult struct {
	Body         []byte        // response body
	Status       int           // response status
	ID           string        // created object id (Create only)
	RequestID    string        // x-ak-request-id sent with the request
	ConfigStatus *ConfigStatus // from the response body, when present
	// Previous is the stored object read before an update or delete.
	Previous []byte
	// Absent maps each list field to the removed items it did not hold (Update only).
	Absent map[string][]any
	// Unchanged reports an update whose list edits left every list as stored;
	// no PUT was sent and Body is the stored object.
	Unchanged bool
	// Saved reports a create or update the backend stored, including one
	// whose async validation FAILED.
	Saved bool
	// FlagsApplied names the feature flags that changed the request.
	FlagsApplied []string
	// FieldsDropped names the fields left out because their flag is off.
	FieldsDropped []string
	// FieldsCleared names the stored fields an update sent empty because of
	// a FlagClearField or ClearStored rule.
	FieldsCleared []string
	// ValidationStateID is the x-ak-validation-state-id of an async write.
	ValidationStateID string
	// ValidationState is ValidationRunning while the async validation has
	// not finished within the poll, ValidationUnknown when the validation
	// state id was lost with the response, else its terminal state.
	ValidationState string
	// ValidationErrors holds the errors of a FAILED async validation.
	ValidationErrors *ValidationErrorDetails
	// ValidationReadError is the validation state read error that ended the
	// poll early because a retry would not change it (not found, or a 4xx
	// other than 408 and 429). ValidationState then stays ValidationRunning;
	// the save stands.
	ValidationReadError error
	// ProvisionRequestID is the x-provision-request-id of a provision.
	ProvisionRequestID string
	// ValidationRequestID is the x-validation-request-id of a dry-run validation request.
	ValidationRequestID string
}

// FieldError is one errors[] entry of a backend ErrorDetails body.
type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// APIError is a non-2xx response on the write path.
type APIError struct {
	Method    string
	URI       string
	Status    int
	Code      string       // ErrorDetails.code
	BackendID string       // ErrorDetails.id, the id support asks for
	Message   string       // ErrorDetails.message, or the raw body when it is not ErrorDetails
	Errors    []FieldError // ErrorDetails.errors[]
	RequestID string       // x-ak-request-id sent with the request
	Body      string       // first 4 KiB of the response body
	// RetryAfter is the parsed Retry-After header, zero when absent.
	RetryAfter time.Duration
	// detailed reports that the body is ErrorDetails JSON with a code or message.
	detailed bool
	kind     error
}

func (e *APIError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s: %d", e.Method, e.URI, e.Status)
	if e.Code != "" {
		fmt.Fprintf(&b, " %s", e.Code)
	}
	if e.Message != "" {
		fmt.Fprintf(&b, ": %s", e.Message)
	}
	for _, fe := range e.Errors {
		fmt.Fprintf(&b, "; %s: %s", fe.Field, fe.Message)
	}
	if e.BackendID != "" {
		fmt.Fprintf(&b, " (error id %s)", e.BackendID)
	}
	fmt.Fprintf(&b, " (request id %s)", e.RequestID)
	return b.String()
}

// Is reports whether the error was classified as target.
func (e *APIError) Is(target error) bool {
	return e.kind != nil && e.kind == target
}

// matches reports whether the response matches one of sigs.
func (e *APIError) matches(sigs []Signature) bool {
	if !e.detailed {
		return false
	}
	for _, s := range sigs {
		if s.Status != e.Status {
			continue
		}
		if s.Code != "" && s.Code != e.Code {
			continue
		}
		if s.FieldCode != "" && !e.hasFieldCode(s.FieldCode) {
			continue
		}
		if s.Message != "" && !strings.Contains(e.Message, s.Message) {
			continue
		}
		return true
	}
	return false
}

func (e *APIError) hasFieldCode(code string) bool {
	for _, fe := range e.Errors {
		if fe.Code == code {
			return true
		}
	}
	return false
}

func newAPIError(method, uri, requestId string, resp *http.Response, body []byte) *APIError {
	e := &APIError{Method: method, URI: uri, Status: resp.StatusCode, RequestID: requestId}
	if len(body) > maxErrorBody {
		e.Body = string(body[:maxErrorBody])
	} else {
		e.Body = string(body)
	}
	e.RetryAfter, _ = parseRetryAfter(resp.Header.Get("Retry-After"))

	var details struct {
		Code    string       `json:"code"`
		ID      string       `json:"id"`
		Message string       `json:"message"`
		Errors  []FieldError `json:"errors"`
	}
	if err := json.Unmarshal(body, &details); err == nil {
		e.Code, e.BackendID, e.Message, e.Errors = details.Code, details.ID, details.Message, details.Errors
		e.detailed = details.Code != "" || details.Message != ""
	}
	if !e.detailed && e.Message == "" {
		e.Message = strings.TrimSpace(e.Body)
	}

	switch {
	// AbstractConnectorService, AbstractBaseService, ConnectorServiceImpl,
	// ServiceLogicImpl: "... cannot be updated|deleted until the validation completes".
	case e.Status == 403 && strings.Contains(e.Message, "until the validation completes"):
		e.kind = ErrValidationPending
	case e.Status == 429:
		e.kind = ErrRateLimited
	// ProvisionedEntityValidatorImpl: "Cannot modify|delete <type> ... at this time (network state ...)".
	case e.Status == 400 && strings.Contains(e.Message, " at this time") &&
		(strings.HasPrefix(e.Message, "Cannot modify") || strings.HasPrefix(e.Message, "Cannot delete")):
		e.kind = ErrStateRefused
	// tenant-billing answers its tag-count limit with 404 on create and update (TagServiceImpl.java:95-108).
	case e.Status == 404 && strings.HasPrefix(e.Message, "Cannot create more than"):
		e.kind = ErrLimitExceeded
	// TPS "Cannot create a new <type>, limit = N": 400 from ResourceLimitsValidatorImpl.java:157,
	// 403 from TenantLimitsValidatorImpl.java:84-145.
	case (e.Status == 400 || e.Status == 403) && strings.HasPrefix(e.Message, "Cannot create a new ") && strings.Contains(e.Message, ", limit = "):
		e.kind = ErrLimitExceeded
	}
	return e
}

// parseRetryAfter parses delay-seconds or an HTTP date.
func parseRetryAfter(v string) (time.Duration, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, false
	}
	if s, err := strconv.Atoi(v); err == nil {
		if s < 0 {
			return 0, false
		}
		return time.Duration(s) * time.Second, true
	}
	if t, err := http.ParseTime(v); err == nil {
		d := time.Until(t)
		if d < 0 {
			d = 0
		}
		return d, true
	}
	return 0, false
}

// writeTimeout returns the default per-call write deadline.
func writeTimeout() time.Duration {
	v := strings.TrimSpace(os.Getenv(WriteTimeoutEnv))
	if v == "" {
		return DefaultWriteTimeout
	}
	if s, err := strconv.Atoi(v); err == nil && s > 0 {
		return time.Duration(s) * time.Second
	}
	if d, err := time.ParseDuration(v); err == nil && d > 0 {
		return d
	}
	return DefaultWriteTimeout
}

// withWriteDeadline applies the default write deadline when ctx has none.
func withWriteDeadline(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, writeTimeout())
}

// writeRetryPolicy allows one retry of a POST, PUT or DELETE: on 429 with
// Retry-After of at most 30 s, and on a connection that was never
// established (dial or proxyconnect). retryablehttp passes resp == nil on a
// transport error and never the method, so this policy only suits the write
// client.
//
// A wait that would not end before the context deadline is not started:
// retryablehttp would return only ctx.Err() mid-wait, losing the response
// that says the write was not applied.
func writeRetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if err != nil {
		var op *net.OpError
		if errors.As(err, &op) && (op.Op == "dial" || op.Op == "proxyconnect") {
			return fitsDeadline(ctx, writeDialBackoff), nil
		}
		return false, nil
	}
	if resp.StatusCode == 429 {
		if d, ok := parseRetryAfter(resp.Header.Get("Retry-After")); ok && d <= maxWriteRetryAfter {
			return fitsDeadline(ctx, d), nil
		}
	}
	return false, nil
}

// fitsDeadline reports whether a wait of d ends before ctx's deadline.
func fitsDeadline(ctx context.Context, d time.Duration) bool {
	deadline, ok := ctx.Deadline()
	return !ok || d < time.Until(deadline)
}

func writeBackoff(_, _ time.Duration, _ int, resp *http.Response) time.Duration {
	if resp != nil {
		if d, ok := parseRetryAfter(resp.Header.Get("Retry-After")); ok {
			return d
		}
	}
	return writeDialBackoff
}

// writer is the retry client for POST, PUT and DELETE. It shares the HTTP
// client, and so the transport, of ac.Client; the read client's policy
// retries every 5xx and post-send failures, which a write must not.
func (ac *AlkiraClient) writer() *retryablehttp.Client {
	return &retryablehttp.Client{
		HTTPClient:   ac.Client.HTTPClient,
		Logger:       ac.Client.Logger,
		RetryMax:     1,
		CheckRetry:   writeRetryPolicy,
		Backoff:      writeBackoff,
		ErrorHandler: retryablehttp.PassthroughErrorHandler,
	}
}

// reader is the read client's policy with typed errors kept after the last retry.
func (ac *AlkiraClient) reader() *retryablehttp.Client {
	return &retryablehttp.Client{
		HTTPClient:   ac.Client.HTTPClient,
		Logger:       ac.Client.Logger,
		RetryWaitMin: ac.Client.RetryWaitMin,
		RetryWaitMax: ac.Client.RetryWaitMax,
		RetryMax:     ac.Client.RetryMax,
		CheckRetry:   ac.Client.CheckRetry,
		Backoff:      ac.Client.Backoff,
		ErrorHandler: retryablehttp.PassthroughErrorHandler,
	}
}

// isPreSend reports whether err is a transport failure before any request
// byte reached the server: no connection, or a failed TLS handshake.
func isPreSend(err error) bool {
	var op *net.OpError
	if errors.As(err, &op) && (op.Op == "dial" || op.Op == "proxyconnect") {
		return true
	}
	// Certificate errors arise only in the handshake. tls.RecordHeaderError is
	// left out: it also arises on an established connection after the body
	// was sent.
	var (
		cv *tls.CertificateVerificationError
		ua x509.UnknownAuthorityError
		hn x509.HostnameError
		ci x509.CertificateInvalidError
	)
	return errors.As(err, &cv) || errors.As(err, &ua) || errors.As(err, &hn) || errors.As(err, &ci)
}

// doWrite sends one POST, PUT or DELETE. Any 2xx is success. A non-2xx
// returns *APIError. A transport failure is ErrNotSent when no request
// byte reached the server and ErrPostSend otherwise. The request id is fixed across the retry.
func (ac *AlkiraClient) doWrite(ctx context.Context, method, uri string, body []byte) (*WriteResult, error) {
	if err := ac.check(uri); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNotSent, err)
	}
	if ac.Client == nil {
		return nil, fmt.Errorf("%w: %w", ErrNotSent, ErrNilHTTPClient)
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("%s %s: %w: %w", method, uri, ErrNotSent, err)
	}
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := retryablehttp.NewRequestWithContext(ctx, method, uri, reader)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w: invalid request: %w", method, uri, ErrNotSent, err)
	}
	requestId := ac.setWriteHeaders(req)
	logf("DEBUG", "client-write(%s) %s %s REQ: %s", requestId, method, uri, redactBody(body))

	resp, err := ac.writer().Do(req)
	if resp == nil {
		switch {
		case err == nil:
			err = errors.New("no response")
		case ctx.Err() != nil:
			return nil, fmt.Errorf("%s %s (request id %s): %w: %w", method, uri, requestId, errDeadlineInFlight, ctx.Err())
		case isPreSend(err):
			return nil, fmt.Errorf("%s %s (request id %s): %w: %w", method, uri, requestId, ErrNotSent, err)
		}
		// Any other transport failure may have reached the server.
		return nil, fmt.Errorf("%s %s (request id %s): %w: %w", method, uri, requestId, ErrPostSend, err)
	}
	defer drainClose(resp.Body)
	data, rerr := io.ReadAll(io.LimitReader(resp.Body, maxWriteBody))
	logf("DEBUG", "client-write(%s) %d RSP: %s", requestId, resp.StatusCode, redactBody(data))

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, newAPIError(method, uri, requestId, resp, data)
	}
	if rerr != nil {
		return nil, fmt.Errorf("%s %s (request id %s): %d response: %w: %w", method, uri, requestId, resp.StatusCode, ErrPostSend, rerr)
	}
	res := &WriteResult{Body: data, Status: resp.StatusCode, RequestID: requestId,
		ValidationStateID:   resp.Header.Get("x-ak-validation-state-id"),
		ProvisionRequestID:  resp.Header.Get("x-provision-request-id"),
		ValidationRequestID: resp.Header.Get("x-validation-request-id")}
	var cs struct {
		ConfigStatus *ConfigStatus `json:"configStatus"`
	}
	if json.Unmarshal(data, &cs) == nil {
		res.ConfigStatus = cs.ConfigStatus
	}
	return res, nil
}

// readerNo5xx is reader without retries on a 5xx answer, for a read whose
// caller handles a 5xx itself.
func (ac *AlkiraClient) readerNo5xx() *retryablehttp.Client {
	rc := ac.reader()
	check := rc.CheckRetry
	if check == nil {
		check = retryablehttp.DefaultRetryPolicy
	}
	rc.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if err == nil && resp != nil && resp.StatusCode >= 500 {
			return false, nil
		}
		return check(ctx, resp, err)
	}
	return rc
}

// doGet sends one GET on the write path (pre-GET, guard reads, confirming
// GET) under ctx. A non-2xx returns *APIError.
func (ac *AlkiraClient) doGet(ctx context.Context, uri string) ([]byte, error) {
	return ac.doGetWith(ctx, uri, ac.reader)
}

// doGetNo5xxRetry is doGet that returns a 5xx at once instead of retrying it.
func (ac *AlkiraClient) doGetNo5xxRetry(ctx context.Context, uri string) ([]byte, error) {
	return ac.doGetWith(ctx, uri, ac.readerNo5xx)
}

func (ac *AlkiraClient) doGetWith(ctx context.Context, uri string, client func() *retryablehttp.Client) ([]byte, error) {
	if err := ac.check(uri); err != nil {
		return nil, err
	}
	if ac.Client == nil {
		return nil, ErrNilHTTPClient
	}
	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("GET %s: invalid request: %w", uri, err)
	}
	requestId := ac.setWriteHeaders(req)

	resp, err := client().Do(req)
	if resp == nil {
		if err == nil {
			err = errors.New("no response")
		}
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return nil, fmt.Errorf("GET %s (request id %s): %w", uri, requestId, err)
	}
	defer drainClose(resp.Body)
	data, rerr := io.ReadAll(io.LimitReader(resp.Body, maxWriteBody))
	logf("DEBUG", "client-get(%s) %d RSP: %s", requestId, resp.StatusCode, redactBody(data))

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, newAPIError("GET", uri, requestId, resp, data)
	}
	if rerr != nil {
		return nil, fmt.Errorf("GET %s (request id %s): %w", uri, requestId, rerr)
	}
	return data, nil
}

func (ac *AlkiraClient) setWriteHeaders(req *retryablehttp.Request) string {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", ac.Authorization)
	req.Header.Set("x-ak-request-id", nextRequestId())
	req.Header.Set("User-Agent", WriteUserAgent)
	ac.decorate(req)
	return req.Header.Get("x-ak-request-id")
}
