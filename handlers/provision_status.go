package handlers

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/mark3labs/mcp-go/mcp"
)

// ValidationStateTool is the read tool that reports an async validation state.
const ValidationStateTool = "tenant_network_validation_state_get"

// The write tools name ValidationStateTool after a validation that has not
// finished, since this file serves it.
func init() {
	validationStateCheck = " Check it with " + ValidationStateTool + "."
}

// statusNow is the clock of the "may be stuck" hint.
var statusNow = time.Now

// stuckAfter is how long a request runs before the hint (V3 internal/client/provision/waiter.go:73-131).
const stuckAfter = 30 * time.Minute

var (
	failureStates = []string{ak.ProvisionFailed, ak.ProvisionPartialSuccess}
	// pendingStates: INITIALIZING and a row with no state are pending (Q-A6).
	pendingStates = []string{ak.ProvisionPending, ak.ProvisionScheduled, ak.ProvisionInitializing, ak.ProvisionInProgress, ""}
	runningStates = []string{ak.ProvisionPending, ak.ProvisionScheduled, ak.ProvisionInProgress}
)

var requestID = regexp.MustCompile(`^[0-9A-Za-z-]{1,64}$`)

func requestIDArg(args map[string]any) (string, *mcp.CallToolResult) {
	id, ok := args["id"].(string)
	if !ok || !requestID.MatchString(id) {
		return "", invalidArgs(errors.New("id must be a request id of letters, digits and hyphens"))
	}
	return id, nil
}

// provisionVerdict applies Q-A6 to the request state, the buffered request
// state and the network's last provision state.
func provisionVerdict(pr *ak.TenantNetworkProvisionRequest, buffered, network string, now time.Time) string {
	req := pr.State
	switch {
	case slices.Contains(failureStates, req) || slices.Contains(failureStates, buffered):
		v := "FAILURE: the provision did not complete for every change."
		if pr.ErrorDetails.ContactSupport() {
			v += " Contact support: the backend marks this failure as needing escalation, not a retry."
		} else {
			v += " Fix the cause and provision again."
		}
		if network == ak.ProvisionFailed {
			v += " The network is FAILED, which blocks provisioning for every user until an operator repairs it in the portal."
		}
		return v
	case slices.Contains(pendingStates, req):
		if network == ak.ProvisionFailed && req != ak.ProvisionInProgress {
			return "PARKED: the network's last provision FAILED, so this request will not run until an operator repairs the network."
		}
		v := "PENDING: the provision has not finished; poll again."
		if req == ak.ProvisionInProgress && pr.CreatedAt > 0 && now.Sub(time.UnixMilli(pr.CreatedAt)) > stuckAfter {
			v += fmt.Sprintf(" It has run for more than %d minutes since it was filed and may be stuck; this is not a failure.", int(stuckAfter.Minutes()))
		}
		return v
	case req == ak.ProvisionSuccess && slices.Contains(failureStates, network):
		return fmt.Sprintf("FAILURE: the request finished, but the network's last provision state is %s.", network)
	case req == ak.ProvisionSuccess && slices.Contains(runningStates, network):
		return "SETTLING: the request finished and the network state has not caught up yet; poll again in a few seconds."
	case req == ak.ProvisionSuccess:
		return "SUCCESS: the provision completed."
	}
	return fmt.Sprintf("UNKNOWN: request state %q is not one this tool knows; read it in the portal.", req)
}

// transient reports the load balancer's HTML 403 seen on polls while a
// request runs (RT connector-mutations/RUN-SUMMARY.md:9).
func transient(err error) bool {
	var ae *ak.APIError
	return errors.As(err, &ae) && ae.Status == 403 && strings.HasPrefix(strings.TrimSpace(ae.Body), "<")
}

func readError(what string, err error) *mcp.CallToolResult {
	var ae *ak.APIError
	switch {
	case transient(err):
		return refuse("transient", "Transient error reading "+what+"; poll again. This is not a request failure.")
	case errors.Is(err, ak.ErrNotFound), errors.As(err, &ae) && ae.Status == 404:
		return refuse("not_found", capital(what)+" not found.")
	}
	return mcp.NewToolResultError(capital(what) + " could not be read: " + err.Error())
}

// ProvisionRequestGet returns the tenant_network_provision_request_get handler.
func ProvisionRequestGet(client *ak.AlkiraClient) ToolHandler {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, res := toolArgs(req, []string{"id"}, []string{"id"})
		if res != nil {
			return res, nil
		}
		id, res := requestIDArg(args)
		if res != nil {
			return res, nil
		}
		pr, err := client.GetProvisionRequest(ctx, id)
		if err != nil {
			return readError("provision request "+id, err), nil
		}
		var out strings.Builder
		fmt.Fprintf(&out, "Provision request %s: state %s.\n", id, orNone(pr.State))
		buffered := ""
		if b := pr.BufferedProvisionRequestId; b != nil {
			bp, err := client.GetBufferedProvisionRequest(ctx, *b)
			if err != nil {
				fmt.Fprintf(&out, "Buffered request %d: state unknown (%v).\n", *b, err)
			} else {
				buffered = bp.State
				fmt.Fprintf(&out, "Buffered request %d: state %s.\n", *b, orNone(bp.State))
			}
		} else {
			out.WriteString("Buffered request: none yet.\n")
		}
		network, provisioned, err := client.GetLastProvisionState(ctx)
		switch {
		case err != nil:
			fmt.Fprintf(&out, "Network last provision state: unknown (%v).\n", err)
		case !provisioned:
			out.WriteString("Network last provision state: never provisioned.\n")
		default:
			fmt.Fprintf(&out, "Network last provision state: %s.\n", network)
		}
		if pr.ErrorDetails != nil {
			fmt.Fprintf(&out, "Error details: %s.\n", errorDetailsText(pr.ErrorDetails))
		}
		if pr.Description != "" {
			fmt.Fprintf(&out, "Description: %s\n", strconv.Quote(pr.Description))
		}
		out.WriteString("Verdict: " + provisionVerdict(pr, buffered, network, statusNow()) + "\n")
		return mcp.NewToolResultText(out.String()), nil
	}
}

// ValidationRequestGet returns the tenant_network_validation_request_get handler.
func ValidationRequestGet(client *ak.AlkiraClient) ToolHandler {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, res := toolArgs(req, []string{"id"}, []string{"id"})
		if res != nil {
			return res, nil
		}
		id, res := requestIDArg(args)
		if res != nil {
			return res, nil
		}
		vr, err := client.GetValidationRequest(ctx, id)
		if err != nil {
			return readError("validation request "+id, err), nil
		}
		var out strings.Builder
		fmt.Fprintf(&out, "Validation request %s: state %s.\n", id, orNone(vr.State))
		switch vr.State {
		case "SUCCESS":
			out.WriteString("Verdict: the changes passed the dry run; nothing was saved or provisioned.\n")
		case "FAILED":
			out.WriteString("Verdict: the dry run FAILED; provisioning these changes would fail the same way.\n")
		default:
			out.WriteString("Verdict: still running; poll again.\n")
		}
		for _, c := range vr.Changes {
			fmt.Fprintf(&out, "Change: %s:%s %s %q\n", c.EntityType, c.EntityID, c.ChangeType, c.EntityName)
		}
		for _, e := range vr.Errors {
			fmt.Fprintf(&out, "Error: %s\n", strconv.Quote(e.Message))
		}
		return mcp.NewToolResultText(out.String()), nil
	}
}

// ValidationStateGet returns the tenant_network_validation_state_get handler.
// A final state is reported as a synchronous save would report it.
func ValidationStateGet(client *ak.AlkiraClient) ToolHandler {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, res := toolArgs(req, []string{"id"}, []string{"id"})
		if res != nil {
			return res, nil
		}
		id, ok := args["id"].(string)
		if !ok || id == "" || len(id) > 200 {
			return invalidArgs(errors.New("id must be a validation state id TYPE:SUBTYPE:ID")), nil
		}
		vs, err := client.GetValidationState(ctx, id)
		if err != nil {
			return readError("validation state "+id, err), nil
		}
		subject := strings.ToLower(vs.NetworkEntityType)
		if vs.NetworkEntitySubType != "" {
			subject += " (" + vs.NetworkEntitySubType + ")"
		}
		subject += " id " + vs.NetworkEntityID
		switch vs.State {
		case ak.ValidationSucceeded:
			return mcp.NewToolResultText(fmt.Sprintf("Validation SUCCESS: saved %s is valid (validation state id %s). The change is saved, not provisioned; it takes effect after a provision.\n", subject, id)), nil
		case ak.ValidationFailed:
			msg := "no details"
			if e := vs.ErrorDetails; e != nil {
				msg = errorDetailsText(&ak.ProvisionErrorDetails{Code: e.Code, ID: e.ID, Message: e.Message, Errors: e.Errors})
			}
			return mcp.NewToolResultText(fmt.Sprintf("Saved %s (validation state id %s), but validation FAILED: %s; fix it with its _update tool or remove it with its _delete tool. It blocks provisioning until it is fixed or deleted.\n", subject, id, msg)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Validation of %s is still running (state %s, validation state id %s); poll again.\n", subject, orNone(vs.State), id)), nil
	}
}
