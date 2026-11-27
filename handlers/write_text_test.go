package handlers

import (
	"context"
	"fmt"
	"testing"
	"time"

	ak "github.com/alkiranet/client-go/tenant"
)

// A deadline during the delete preview is reported as a preview timeout,
// and a flag read cut short by the deadline as a timeout, not as missing flags.
func TestErrorResult_TimeoutWording(t *testing.T) {
	w := newWriteCall(&WritePipeline{timeout: time.Millisecond}, WriteFamily[thing]{Name: "thing", Noun: "thing", Descriptor: ak.Descriptor{Collection: "things"}})
	r := refusalOf(t, w.errorResult(verbPreview, "7", "", context.DeadlineExceeded, redaction{}))
	if want := "The delete preview of thing id 7 timed out after 1ms; nothing was deleted. Retry the delete."; r.Code != "timeout" || r.Text != want {
		t.Errorf("preview refusal = %+v", r)
	}
	// client-go returns the caller's context error alone when the deadline
	// passes during the flag read (features.go:103-105).
	flags := fmt.Errorf("read tenant features: %w", context.DeadlineExceeded)
	r = refusalOf(t, w.errorResult("update", "7", "", flags, redaction{}))
	if want := "The update of thing id 7 timed out after 1ms before the write was sent; nothing changed. Retry the update."; r.Code != "timeout" || r.Text != want {
		t.Errorf("flag read refusal = %+v", r)
	}
	if r := refusalOf(t, w.errorResult("update", "7", "", ak.ErrFlagsUnavailable, redaction{})); r.Code != "flags_unavailable" {
		t.Errorf("flags refusal = %+v", r)
	}
}

// backendText drops the generic "Validation failed" head of field errors and
// the ": " of a field error without a field.
func TestBackendText(t *testing.T) {
	for _, c := range []struct {
		ae   ak.APIError
		want string
	}{
		{ak.APIError{Message: "Validation failed", Errors: []ak.FieldError{{Message: "Cannot delete group 'g' because it is used by some connectors"}}, Code: "AK000001"},
			`"Cannot delete group 'g' because it is used by some connectors" (code AK000001)`},
		{ak.APIError{Message: "Validation failed", Errors: []ak.FieldError{{Field: "cidr", Message: "overlaps"}}}, `"cidr: overlaps"`},
		{ak.APIError{Message: "Widget is in use", Errors: []ak.FieldError{{Field: "widget", Message: "used by policy 3"}}}, `"Widget is in use; widget: used by policy 3"`},
		{ak.APIError{Message: "Validation failed"}, `"Validation failed"`},
	} {
		if got := backendText(&c.ae, redaction{}); got != c.want {
			t.Errorf("backendText(%+v) = %s, want %s", c.ae, got, c.want)
		}
	}
}

// WithWriteFamilies limits Serves to the named families.
func TestWithWriteFamilies(t *testing.T) {
	p, err := NewWritePipeline(WithWriteFamilies(ParseWriteFamilies(" segment, connector_aws_vpc ,,")...))
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]bool{"segment": true, "connector_aws_vpc": true, "segment_resource": false, "group": false} {
		if p.Serves(name) != want {
			t.Errorf("Serves(%s) = %v", name, !want)
		}
	}
	if p, _ := NewWritePipeline(); !p.Serves("group") || p.WriteFamilies() != nil {
		t.Error("a pipeline without the option serves every family")
	}
}

// D19: an update refused because the object changed since it was read says
// nothing was sent and to retry.
func TestErrorResult_ChangedSinceRead(t *testing.T) {
	w := newWriteCall(&WritePipeline{timeout: time.Minute}, WriteFamily[thing]{Name: "thing", Noun: "thing", Descriptor: ak.Descriptor{Collection: "things"}})
	err := fmt.Errorf("%w: id 7 %w; read it again and retry the update", ak.ErrNotSent, ak.ErrChangedSinceRead)
	r := refusalOf(t, w.errorResult("update", "7", "", err, redaction{}))
	if want := "The update of thing id 7 was not sent, nothing changed: another change saved the object after this update read it. Retry the update; it reads the current object first, so check that the patch still fits it."; r.Code != "not_sent" || r.Text != want {
		t.Errorf("refusal = %+v", r)
	}
}
