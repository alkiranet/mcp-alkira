package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/internal/confirm"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestEditsDiff(t *testing.T) {
	edits := ak.ListEdits{
		Add: map[string]json.RawMessage{
			"objs":  json.RawMessage(`[{"name":"x"},{"name":"y"}]`),
			"names": json.RawMessage(`["Web"]`),
		},
		Remove: map[string]json.RawMessage{
			"names": json.RawMessage(`["db","OLD"]`),
			"sites": json.RawMessage(`[1,"2"]`),
		},
	}
	echo := []byte(`{"objs":[{"id":5,"name":"x","enabled":true}],"names":["web","db"],"sites":[{"id":2,"name":"b"}]}`)
	absent := map[string][]any{"names": {"OLD"}}
	got := editsDiff(edits, echo, absent, []string{"sites"})
	want := []string{
		`names: "Web" not added`,
		`names: "db" not removed`,
		`objs: {"name":"y"} not added`,
		`sites: element id 2 not removed`,
	}
	if !slices.Equal(got, want) {
		t.Fatalf("diff = %q, want %q", got, want)
	}
}

// Options replace the defaults; a short key or a non-positive deadline is refused.
func TestNewWritePipeline_Options(t *testing.T) {
	t.Setenv(ak.WriteTimeoutEnv, "")
	p, err := NewWritePipeline()
	if err != nil || p.confirm == nil || p.timeout != ak.DefaultWriteTimeout {
		t.Fatalf("defaults: %+v, %v", p, err)
	}
	t.Setenv(ak.WriteTimeoutEnv, "90")
	if p, _ := NewWritePipeline(); p.timeout != 90*time.Second {
		t.Fatalf("env timeout = %s", p.timeout)
	}
	p, err = NewWritePipeline(WithConfirmKey(bytes.Repeat([]byte("k"), 32)), WithWriteTimeout(5*time.Second))
	if err != nil || p.confirm == nil || p.timeout != 5*time.Second {
		t.Fatalf("options: %+v, %v", p, err)
	}
	for _, key := range [][]byte{nil, bytes.Repeat([]byte("k"), 31)} {
		if p, err := NewWritePipeline(WithConfirmKey(key)); !errors.Is(err, confirm.ErrShortKey) || p != nil {
			t.Errorf("key of %d bytes: %v, %v", len(key), p, err)
		}
	}
	for _, d := range []time.Duration{0, -time.Second} {
		if p, err := NewWritePipeline(WithWriteTimeout(d)); err == nil || p != nil {
			t.Errorf("timeout %s: %v, %v", d, p, err)
		}
	}
	for _, opts := range [][]WriteOption{{WithoutConfirm()}, {WithConfirmKey(bytes.Repeat([]byte("k"), 32)), WithoutConfirm()}, {WithoutConfirm(), WithConfirmKey(bytes.Repeat([]byte("k"), 32))}} {
		if p, err := NewWritePipeline(opts...); err != nil || p.confirm != nil {
			t.Errorf("WithoutConfirm: %+v, %v", p, err)
		}
	}
}

// A pipeline built WithoutConfirm refuses every delete with confirm_unavailable and sends nothing.
func TestWriteDelete_WithoutConfirm(t *testing.T) {
	p, err := NewWritePipeline(WithoutConfirm())
	if err != nil {
		t.Fatal(err)
	}
	h := WriteDelete(p, WriteFamily[thing]{Name: "thing", Noun: "thing", Descriptor: ak.Descriptor{Collection: "things"}})(nil)
	for _, args := range []map[string]any{{"id": "7"}, {"id": "7", "confirmCode": "x"}} {
		req := mcp.CallToolRequest{}
		req.Params.Arguments = args
		res, err := h(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		if r := refusalOf(t, res); r.Code != RefusalConfirmUnavailable {
			t.Errorf("%v: %+v", args, r)
		}
	}
}

// WithClock sets the clock of the delete codes.
func TestNewWritePipeline_Clock(t *testing.T) {
	at := time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)
	p, err := NewWritePipeline(WithClock(func() time.Time { return at }))
	if err != nil {
		t.Fatal(err)
	}
	if _, exp := p.confirm.Issue("t", "f", "1", [32]byte{}); !exp.Equal(at.Add(confirm.Lifetime)) {
		t.Fatalf("expiry = %s", exp)
	}
}

// Secret values are masked in backend text also as they appear inside a JSON string.
func TestRedaction_MasksJSONEscaped(t *testing.T) {
	r := redaction{values: []string{`p"a\ss<1>`}}
	for _, text := range []string{
		`plain p"a\ss<1> end`,
		`json {"psk":"p\"a\\ss<1>"} end`,
		`html {"psk":"p\"a\\ss\u003c1\u003e"} end`,
	} {
		if got := r.mask(text); strings.Contains(got, "ss") || !strings.Contains(got, Masked) {
			t.Errorf("mask(%s) = %s", text, got)
		}
	}
}

// The timeout refusal of a deadline before the write says nothing changed.
func TestErrorResult_Timeout(t *testing.T) {
	w := newWriteCall(&WritePipeline{timeout: time.Minute}, WriteFamily[thing]{Name: "thing", Noun: "thing", Descriptor: ak.Descriptor{Collection: "things"}})
	r := refusalOf(t, w.errorResult("update", "7", "", context.DeadlineExceeded, redaction{}))
	if want := "The update of thing id 7 timed out after 1m0s before the write was sent; nothing changed. Retry the update."; r.Code != "timeout" || r.Text != want {
		t.Errorf("refusal = %+v", r)
	}
}

func TestKeyed(t *testing.T) {
	call := func(key string) writeCall[struct{}] {
		return writeCall[struct{}]{f: WriteFamily[struct{}]{Noun: "widget", Descriptor: ak.Descriptor{Uniqueness: ak.Uniqueness{Key: key}}}}
	}
	for _, c := range []struct{ key, value, want string }{
		{"name", "w1", `a widget named "w1"`},
		{"", "w1", `a widget named "w1"`},
		{"name", "", "a widget with this name"},
		{"prefix", "10.0.0.0/8", `a widget with prefix "10.0.0.0/8"`},
		{"prefix", "", "a widget with this prefix"},
	} {
		if got := call(c.key).keyed(c.value); got != c.want {
			t.Errorf("keyed(%q, %q) = %q, want %q", c.key, c.value, got, c.want)
		}
	}
}

// secretKeys match without regard to case, below fields without struct rules too.
func TestMask_SecretKeyCase(t *testing.T) {
	v := map[string]any{"extra": map[string]any{"PresharedKey": "psk-1", "PRESHAREDKEYS": []any{"psk-2"}, "name": "n"}}
	got := secretValues(nil, v, nil)
	slices.Sort(got)
	if want := []string{"psk-1", "psk-2"}; !slices.Equal(got, want) {
		t.Fatalf("secretValues = %q, want %q", got, want)
	}
	m := mask(nil, v).(map[string]any)["extra"].(map[string]any)
	if m["PresharedKey"] != Masked || m["PRESHAREDKEYS"] != Masked || m["name"] != "n" {
		t.Fatalf("mask = %v", m)
	}
}

// Object array elements compare by the subset rule: the backend adds ids
// and derived fields to each element.
func TestEchoDiff_ObjectArrays(t *testing.T) {
	sent := []byte(`{"sites":[{"name":"s1"},{"name":"S2","tags":["a","b"]}],"names":["x","y"]}`)
	for _, c := range []struct {
		echo string
		want []string
	}{
		{`{"sites":[{"id":2,"name":"s2","tags":["b","a"]},{"id":1,"name":"s1"}],"names":["Y","x"]}`, nil},
		{`{"sites":[{"id":1,"name":"s1"},{"id":2,"name":"s3","tags":["a","b"]}],"names":["x","y"]}`, []string{"sites changed by the backend"}},
		{`{"sites":[{"id":1,"name":"s1"},{"id":2,"name":"s2","tags":["a"]}],"names":["x","y"]}`, []string{"sites changed by the backend"}},
		{`{"sites":[{"id":1,"name":"s1"},{"id":2,"name":"s2","tags":["a","b"]},{"id":3,"name":"s4"}],"names":["x","y","z"]}`,
			[]string{"names changed by the backend", "sites changed by the backend"}},
	} {
		if got := echoDiff(sent, []byte(c.echo), nil); !slices.Equal(got, c.want) {
			t.Errorf("echo %s: diff = %q, want %q", c.echo, got, c.want)
		}
	}
	// A loose element must not take the only response element a stricter one fits.
	sent = []byte(`{"sites":[{"name":"a"},{"name":"a","x":1}]}`)
	if got := echoDiff(sent, []byte(`{"sites":[{"name":"a","x":1},{"name":"a"}]}`), nil); got != nil {
		t.Errorf("pairing: diff = %q", got)
	}
}

type thing struct {
	ID   *int64  `json:"id,omitempty" ak:"readonly"`
	Name *string `json:"name,omitempty"`
}

func refusalOf(t *testing.T, res *mcp.CallToolResult) Refusal {
	t.Helper()
	r, ok := res.Meta["refusal"].(Refusal)
	if !ok {
		t.Fatalf("no refusal: %+v", res)
	}
	return r
}

func TestErrorResult_Texts(t *testing.T) {
	w := newWriteCall(&WritePipeline{timeout: time.Minute}, WriteFamily[thing]{Name: "thing", Noun: "thing",
		Descriptor: ak.Descriptor{Collection: "things"}})

	// A state refusal without a backend answer shows client-go's reason.
	err := fmt.Errorf("%w: GET id 7 has no docState", ak.ErrStateRefused)
	r := refusalOf(t, w.errorResult("update", "7", "", err, redaction{}))
	if r.Code != "state_refused" || !strings.Contains(r.Text, "(GET id 7 has no docState)") || strings.Contains(r.Text, "provision") {
		t.Errorf("state refused without API error: %+v", r)
	}

	// A family without a uniqueness key gets no lookup text on a lost create.
	oe := &ak.OutcomeError{Method: "POST", Outcome: ak.OutcomeUnknown, Err: errors.New("EOF")}
	r = refusalOf(t, w.outcomeResult(oe, ""))
	if r.Code != "outcome_unknown" || strings.Contains(r.Text, "lookup") || !strings.HasPrefix(r.Text, "Create of thing: outcome unknown") {
		t.Errorf("no uniqueness key: %+v", r)
	}
}

func TestObjectLocks(t *testing.T) {
	w := newWriteCall(&WritePipeline{timeout: time.Minute}, WriteFamily[thing]{Name: "thing", Noun: "thing",
		Descriptor: ak.Descriptor{Collection: "things"}})
	client := &ak.AlkiraClient{URI: "https://a/api", TenantNetworkId: "7"}
	ctx := context.Background()

	unlock, err := w.lockObject(ctx, client, "id\x007")
	if err != nil {
		t.Fatal(err)
	}
	// Another object of the same family, and the same id on another tenant, do not wait.
	other, err := w.lockObject(ctx, client, "id\x008")
	if err != nil {
		t.Fatal(err)
	}
	other()
	tenant2, err := w.lockObject(ctx, &ak.AlkiraClient{URI: "https://a/api", TenantNetworkId: "8"}, "id\x007")
	if err != nil {
		t.Fatal(err)
	}
	tenant2()

	// The same object waits until the call's deadline and is reported as not sent.
	short, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()
	if _, err := w.lockObject(short, client, "id\x007"); !errors.Is(err, ak.ErrNotSent) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v", err)
	} else if r := refusalOf(t, w.errorResult("update", "7", "", err, redaction{})); r.Code != "not_sent" || !strings.Contains(r.Text, "another write to the same thing was still running") {
		t.Fatalf("refusal = %+v", r)
	}

	got := make(chan struct{})
	go func() {
		release, err := w.lockObject(ctx, client, "id\x007")
		if err != nil {
			t.Error(err)
			return
		}
		release()
		close(got)
	}()
	unlock()
	<-got
	w.p.locks.mu.Lock()
	defer w.p.locks.mu.Unlock()
	if len(w.p.locks.m) != 0 {
		t.Fatalf("locks left after release: %v", w.p.locks.m)
	}
}
