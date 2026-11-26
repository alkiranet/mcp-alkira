package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/internal/confirm"
	"github.com/alkiranet/mcp-alkira/internal/schemagen"
	"github.com/mark3labs/mcp-go/mcp"
)

// WriteFamily describes one writable resource family for the write pipeline.
type WriteFamily[T any] struct {
	// Name is the tool stem shared with the read tools, for example "segment".
	Name string
	// Noun names the family in tool output, for example "segment".
	Noun       string
	Descriptor ak.Descriptor
	// NameKey is the field that names an object in tool output when the
	// family has no name field, for example "prefix"; empty means "name".
	NameKey string
	// Implicit returns the objects the backend deletes together with an object named name.
	Implicit func(name string) []string
	// ImplicitNote describes, in the delete preview, objects the backend
	// deletes together with the stored object raw named name when their
	// names do not follow from the object name alone; "" means none.
	ImplicitNote func(name string, raw []byte) string
	// ProvisionNote, when set, replaces the provision effect line of the delete preview.
	ProvisionNote string
	// Volatile lists top-level GET fields that change without a user edit; the
	// delete state hash leaves them out.
	Volatile []string
	// CreateNote, UpdateNote and DeleteNote add family facts to the tool descriptions.
	CreateNote, UpdateNote, DeleteNote string
	// StoredRefs lists references the stored object can keep after their
	// target is deleted (BC-R11).
	StoredRefs []StoredRef
}

// StoredRef is a reference an update resends from the stored object. When
// the backend refuses an update on Field and the patch does not set it, the
// update output names the stored IDField value as dangling once Lookup
// answers ak.ErrNotFound for it.
type StoredRef struct {
	Field, IDField, Noun string
	Lookup               func(ctx context.Context, client *ak.AlkiraClient, id string) error
}

// Refusal is the typed reason of a write tool error, carried in the result
// meta under "refusal" so a hosting layer can pass it through unchanged.
type Refusal struct {
	Code string `json:"code"`
	Text string `json:"text"`
}

// Refusal codes of the two-step delete.
const (
	RefusalConfirmExpired  = "confirm_expired"
	RefusalConfirmMismatch = "confirm_mismatch"
	RefusalConfirmStale    = "confirm_stale"
	// RefusalConfirmUnavailable: the pipeline has no signing key, so no delete runs.
	RefusalConfirmUnavailable = "confirm_unavailable"
)

// WritePipeline runs every write tool call. Writes to different objects run
// concurrently; writes to one object run one at a time within the process
// (objectLocks). A delete code is checked by its signature alone, so any
// pipeline built with the same key accepts it.
type WritePipeline struct {
	// confirm signs and checks the pipeline's confirmation codes. It is nil
	// in a pipeline built with WithoutConfirm, in the preview pipeline and in
	// a zero value; the tools that need a code then refuse every call.
	confirm *confirm.Signer
	// timeout bounds a whole tool call: pre-GET, write and confirming GET.
	timeout time.Duration

	// schema evaluates expiry rules at the pipeline's construction, so the
	// schemas its tools advertise and the validation of their calls agree.
	schema schemagen.Generator

	key       []byte
	noConfirm bool
	now       func() time.Time

	locks objectLocks
}

// Schema returns the generator of the pipeline's tool schemas; Validate
// checks calls with the same generator.
func (p *WritePipeline) Schema() schemagen.Generator { return p.schema }

// WriteOption configures NewWritePipeline.
type WriteOption func(*WritePipeline) error

// WithConfirmKey signs the pipeline's confirmation codes, delete codes
// included, with key instead of a random per-process key. Replicas that
// share a key accept each other's codes. A key shorter than
// confirm.MinKeyLen is refused.
func WithConfirmKey(key []byte) WriteOption {
	return func(p *WritePipeline) error {
		if len(key) < confirm.MinKeyLen {
			return confirm.ErrShortKey
		}
		p.key = append([]byte(nil), key...)
		return nil
	}
}

// WithoutConfirm builds the pipeline with no signing key: every delete
// refuses with confirm_unavailable and sends nothing, while create and update
// work. It takes precedence over WithConfirmKey. A replicated server whose
// shared key is missing uses it, since a random per-replica key would issue
// codes the other replicas refuse.
func WithoutConfirm() WriteOption {
	return func(p *WritePipeline) error {
		p.noConfirm = true
		return nil
	}
}

// WithWriteTimeout sets the deadline of one write tool call, in place of
// AK_MCP_WRITE_TIMEOUT.
func WithWriteTimeout(d time.Duration) WriteOption {
	return func(p *WritePipeline) error {
		if d <= 0 {
			return fmt.Errorf("write timeout must be positive, got %s", d)
		}
		p.timeout = d
		return nil
	}
}

// WithClock makes the pipeline read now instead of the wall clock, for the
// expiry of delete codes and the instant field expiry rules are evaluated at.
func WithClock(now func() time.Time) WriteOption {
	return func(p *WritePipeline) error {
		p.now = now
		return nil
	}
}

// NewWritePipeline returns a pipeline configured by opts. Without options,
// delete codes are signed with a random key drawn for this pipeline and the
// call deadline comes from AK_MCP_WRITE_TIMEOUT (default 60 s).
func NewWritePipeline(opts ...WriteOption) (*WritePipeline, error) {
	p := &WritePipeline{timeout: writeTimeout(), now: time.Now}
	for _, o := range opts {
		if err := o(p); err != nil {
			return nil, err
		}
	}
	p.schema = schemagen.At(p.now())
	if p.noConfirm {
		return p, nil
	}
	if p.key == nil {
		p.key = confirm.RandomKey()
	}
	s, err := confirm.NewSigner(p.key, confirm.WithClock(p.now))
	if err != nil {
		return nil, err
	}
	p.confirm = s
	return p, nil
}

func writeTimeout() time.Duration {
	v := strings.TrimSpace(os.Getenv(ak.WriteTimeoutEnv))
	if s, err := strconv.Atoi(v); err == nil && s > 0 {
		return time.Duration(s) * time.Second
	}
	if d, err := time.ParseDuration(v); err == nil && d > 0 {
		return d
	}
	return ak.DefaultWriteTimeout
}

// writeCall holds what every verb of one family needs.
type writeCall[T any] struct {
	p     *WritePipeline
	f     WriteFamily[T]
	rules []ak.FieldRule
}

func newWriteCall[T any](p *WritePipeline, f WriteFamily[T]) writeCall[T] {
	rules, err := ak.FieldRules[T]()
	if err != nil {
		panic(fmt.Sprintf("write family %s: %v", f.Name, err))
	}
	return writeCall[T]{p: p, f: f, rules: rules}
}

func (w writeCall[T]) core(client *ak.AlkiraClient) (*ak.Core[T], error) {
	return ak.NewCore[T](client, w.f.Descriptor)
}

// WriteCreate returns the {family}_create handler.
func WriteCreate[T any](p *WritePipeline, f WriteFamily[T]) func(*ak.AlkiraClient) ToolHandler {
	w := newWriteCall(p, f)
	return func(client *ak.AlkiraClient) ToolHandler {
		return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, res := toolArgs(req, []string{"body"}, []string{"body"})
			if res != nil {
				return res, nil
			}
			raw, _ := json.Marshal(args["body"])
			if err := w.p.schema.Validate(w.rules, raw, schemagen.Create); err != nil {
				return invalidArgs(err), nil
			}
			var obj T
			if err := json.Unmarshal(raw, &obj); err != nil {
				return invalidArgs(err), nil
			}
			core, err := w.core(client)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			ctx, cancel := context.WithTimeout(ctx, w.p.timeout)
			defer cancel()
			if key := fieldOf(raw, w.f.Descriptor.Uniqueness.Key); w.f.Descriptor.Uniqueness.Key != "" && key != "" {
				unlock, err := w.lockObject(ctx, client, "create\x00"+key)
				if err != nil {
					return w.errorResult("create", "", key, err, redaction{}), nil
				}
				defer unlock()
			}

			created, wr, err := core.Create(ctx, &obj, ak.WriteOptions{})
			name := w.nameOf(raw)
			var oe *ak.OutcomeError
			if errors.As(err, &oe) {
				return w.outcomeResult(oe, fieldOf(raw, w.f.Descriptor.Uniqueness.Key)), nil
			}
			if err != nil {
				return w.errorResult("create", "", fieldOf(raw, w.f.Descriptor.Uniqueness.Key), err, w.redaction(raw, sent(obj))), nil
			}
			var out strings.Builder
			if wr.ValidationState == ak.ValidationFailed {
				id := wr.ID
				if id == "" {
					// The state id is TYPE:SUBTYPE:ID.
					id = wr.ValidationStateID[strings.LastIndex(wr.ValidationStateID, ":")+1:]
				}
				w.validationFailed(&out, fmt.Sprintf("Saved %s%s as id %s", w.f.Noun, quoted(name), orUnknown(id)), wr, sent(obj))
			} else {
				fmt.Fprintf(&out, "Created %s%s with id %s.", w.f.Noun, quoted(name), orUnknown(wr.ID))
				if w.f.Descriptor.Provisionable {
					out.WriteString(" It is saved, not provisioned: it takes effect after a provision.")
				}
				out.WriteString("\n")
			}
			validationRunning(&out, wr)
			w.flagEffects(&out, wr)
			w.writeTail(&out, wr, created, sent(obj), nil)
			return mcp.NewToolResultText(out.String()), nil
		}
	}
}

// WriteUpdate returns the {family}_update handler.
func WriteUpdate[T any](p *WritePipeline, f WriteFamily[T]) func(*ak.AlkiraClient) ToolHandler {
	w := newWriteCall(p, f)
	return func(client *ak.AlkiraClient) ToolHandler {
		return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, res := toolArgs(req, []string{"id", "patch", "add", "remove"}, []string{"id"})
			if res != nil {
				return res, nil
			}
			id, res := idArg(args)
			if res != nil {
				return res, nil
			}
			if args["patch"] == nil && args["add"] == nil && args["remove"] == nil {
				return invalidArgs(errors.New("patch, add or remove is required")), nil
			}
			var patch []byte
			if args["patch"] != nil {
				patch, _ = json.Marshal(args["patch"])
				if err := w.p.schema.Validate(w.rules, patch, schemagen.Patch); err != nil {
					return invalidArgs(err), nil
				}
			}
			var edits ak.ListEdits
			var err error
			if edits.Add, err = listEditArg(args, "add", w.editFields(false)); err != nil {
				return invalidArgs(err), nil
			}
			if edits.Remove, err = listEditArg(args, "remove", w.editFields(true)); err != nil {
				return invalidArgs(err), nil
			}
			core, err := w.core(client)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			ctx, cancel := context.WithTimeout(ctx, w.p.timeout)
			defer cancel()
			unlock, err := w.lockObject(ctx, client, "id\x00"+id)
			if err != nil {
				return w.errorResult("update", id, w.nameOf(patch), err, redaction{}), nil
			}
			defer unlock()

			updated, wr, err := core.Update(ctx, id, patch, edits)
			var oe *ak.OutcomeError
			if errors.As(err, &oe) {
				return w.outcomeResult(oe, w.nameOf(patch)), nil
			}
			if err != nil {
				res := w.errorResult("update", id, w.nameOf(patch), err, w.updateRedaction(ctx, core, id, patch, err))
				if note := w.danglingRef(ctx, client, core, id, patch, err); note != "" && len(res.Content) > 0 {
					if tc, ok := res.Content[0].(mcp.TextContent); ok {
						tc.Text += " " + note
						res.Content[0] = tc
					}
					if rf, ok := res.Meta["refusal"].(Refusal); ok {
						rf.Text += " " + note
						res.Meta["refusal"] = rf
					}
				}
				return res, nil
			}
			var out strings.Builder
			switch {
			case wr.Unchanged && len(wr.FieldsDropped) > 0:
				fmt.Fprintf(&out, "No change to %s id %s: %s left out because its tenant feature is off; nothing sent.\n", w.f.Noun, id, strings.Join(wr.FieldsDropped, ", "))
			case wr.Unchanged:
				fmt.Fprintf(&out, "No change to %s id %s: the stored lists already match the edits; nothing sent.\n", w.f.Noun, id)
			case wr.ValidationState == ak.ValidationFailed:
				w.validationFailed(&out, fmt.Sprintf("Saved the update of %s id %s", w.f.Noun, id), wr, patch)
			default:
				fmt.Fprintf(&out, "Updated %s id %s.", w.f.Noun, id)
				if w.f.Descriptor.Provisionable {
					out.WriteString(" The change takes effect after a provision.")
				}
				out.WriteString("\n")
			}
			validationRunning(&out, wr)
			absentText(&out, wr.Absent)
			if wr.Unchanged {
				w.flagEffects(&out, wr)
				out.WriteString("Object:\n")
				out.WriteString(w.render(wr.Body))
				return mcp.NewToolResultText(out.String()), nil
			}
			w.flagEffects(&out, wr)
			w.writeTail(&out, wr, updated, patch, nullKeys(patch))
			if diff := editsDiff(edits, wr.Body, wr.Absent, w.f.Descriptor.IDMatchedArrays); len(diff) > 0 {
				fmt.Fprintf(&out, "The backend did not keep the list edits as sent: %s.\n", strings.Join(diff, "; "))
			}
			out.WriteString("Previous object:\n")
			out.WriteString(w.render(wr.Previous))
			return mcp.NewToolResultText(out.String()), nil
		}
	}
}

// WriteDelete returns the {family}_delete handler. Without confirmCode it
// sends only GETs and returns a preview with a one-time code; with the code
// it re-reads the object, checks the code against its current state and
// sends one DELETE (D17).
func WriteDelete[T any](p *WritePipeline, f WriteFamily[T]) func(*ak.AlkiraClient) ToolHandler {
	w := newWriteCall(p, f)
	return func(client *ak.AlkiraClient) ToolHandler {
		return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args, res := toolArgs(req, []string{"id", "confirmCode"}, []string{"id"})
			if res != nil {
				return res, nil
			}
			id, res := idArg(args)
			if res != nil {
				return res, nil
			}
			code, isString := args["confirmCode"].(string)
			if _, set := args["confirmCode"]; set && (!isString || code == "") {
				return invalidArgs(errors.New("confirmCode must be a non-empty string")), nil
			}
			if w.p.confirm == nil {
				return refuse(RefusalConfirmUnavailable, "Deletes are unavailable: no confirmation key is configured, so no confirmation code can be issued or checked; nothing was sent."), nil
			}
			core, err := w.core(client)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			tenant := tenantOf(client)
			if code == "" {
				ctx, cancel := context.WithTimeout(ctx, w.p.timeout)
				defer cancel()
				if res := w.typeAvailable(ctx, client); res != nil {
					return res, nil
				}
				return w.preview(ctx, core, tenant, id), nil
			}
			if err := w.p.confirm.CheckSignature(code, tenant, w.f.Name, id); err != nil {
				return w.confirmRefusal(err, id), nil
			}
			ctx, cancel := context.WithTimeout(ctx, w.p.timeout)
			defer cancel()
			unlock, err := w.lockObject(ctx, client, "id\x00"+id)
			if err != nil {
				return w.errorResult("delete", id, "", err, redaction{}), nil
			}
			defer unlock()

			// The code is checked on the core's own pre-GET body, so the
			// DELETE goes out only for the state the preview showed.
			var raw []byte
			wr, err := core.DeleteWith(ctx, id, ak.DeleteOptions{Check: func(stored []byte) error {
				raw = stored
				state, err := confirm.StateHash(stored, w.f.Volatile)
				if err != nil {
					return err
				}
				if err := w.p.confirm.Check(code, tenant, w.f.Name, id, state); err != nil {
					return codeError{err}
				}
				return nil
			}})
			var ce codeError
			if errors.As(err, &ce) {
				return w.confirmRefusal(ce.err, id), nil
			}
			if err != nil {
				return w.errorResult("delete", id, w.nameOf(raw), err, w.redaction(raw)), nil
			}
			raw = wr.Previous
			var out strings.Builder
			// client-go returns a DELETE answered with the family's not-found
			// signature as a result: the object went between the pre-GET and the DELETE.
			if wr.Status >= 400 {
				fmt.Fprintf(&out, "%s%s (id %s) was already deleted: the backend answered the DELETE with not found (status %d).\n", capital(w.f.Noun), quoted(w.nameOf(raw)), id, wr.Status)
			} else if marked, _ := w.marksRow(raw); marked {
				fmt.Fprintf(&out, "Marked %s%s (id %s) for deletion. It stays deployed and keeps its name and references until the next provision.\n", w.f.Noun, quoted(w.nameOf(raw)), id)
			} else {
				fmt.Fprintf(&out, "Deleted %s%s (id %s).\n", w.f.Noun, quoted(w.nameOf(raw)), id)
			}
			out.WriteString("Previous object:\n")
			out.WriteString(w.render(wr.Previous))
			return mcp.NewToolResultText(out.String()), nil
		}
	}
}

// codeError carries a refused confirm code out of the delete check.
type codeError struct{ err error }

func (e codeError) Error() string { return e.err.Error() }

// writeFeatures reads the flags a write uses; *ak.AlkiraClient implements it.
type writeFeatures interface {
	WriteFeatures(ctx context.Context) ([]string, error)
}

// typeAvailable refuses a family whose tenant feature is off before a delete
// preview, which sends no write and so does not pass the core's flag check.
// It reads the flags through WriteFeatures, as the confirmed delete does, so
// during a failed flag read both use the cached flags of any age.
func (w writeCall[T]) typeAvailable(ctx context.Context, client writeFeatures) *mcp.CallToolResult {
	var gate []ak.FlagRule
	for _, r := range w.f.Descriptor.Flags {
		if r.Effect == ak.FlagTypeAvailable {
			gate = append(gate, r)
		}
	}
	if gate == nil {
		return nil
	}
	flags, err := client.WriteFeatures(ctx)
	if err != nil {
		return w.errorResult("delete", "", "", err, redaction{})
	}
	for _, r := range gate {
		needs := r.Flag
		if len(r.AnyOf) > 0 {
			needs += " and one of " + strings.Join(r.AnyOf, ", ")
		}
		on := slices.Contains(flags, r.Flag) && (len(r.AnyOf) == 0 || slices.ContainsFunc(r.AnyOf, func(f string) bool { return slices.Contains(flags, f) }))
		if !on {
			return w.errorResult("delete", "", "", fmt.Errorf("%s: %w (needs %s)", w.f.Descriptor.Collection, ak.ErrNotEnabled, needs), redaction{})
		}
	}
	return nil
}

func (w writeCall[T]) preview(ctx context.Context, core *ak.Core[T], tenant, id string) *mcp.CallToolResult {
	_, raw, err := core.Get(ctx, id)
	if err != nil {
		return w.errorResult("delete", id, "", err, redaction{})
	}
	// Get returns rows the family's writes refuse (FamilyType.Value), such as
	// implicit groups; the preview refuses them as the delete would.
	if f := w.f.Descriptor.FamilyType; f.Value != "" {
		var m map[string]any
		_ = json.Unmarshal(raw, &m)
		if v := m[f.Field]; fmt.Sprint(v) != f.Value {
			return w.errorResult("delete", id, "", &ak.FamilyError{ID: id, Collection: w.f.Descriptor.Collection, Field: f.Field, Got: v, Want: f.Value}, w.redaction(raw))
		}
	}
	state, err := confirm.StateHash(raw, w.f.Volatile)
	if err != nil {
		return w.errorResult("delete", id, "", err, w.redaction(raw))
	}
	code, exp := w.p.confirm.Issue(tenant, w.f.Name, id, state)
	name := w.nameOf(raw)
	deployed, err := ak.Deployed(w.f.Descriptor, raw)
	if err != nil {
		return w.errorResult("delete", id, name, err, w.redaction(raw))
	}
	entityState := stateOf(raw)

	var out strings.Builder
	fmt.Fprintf(&out, "Delete preview for %s%s (id %s). Nothing was deleted.\n", w.f.Noun, quoted(name), id)
	stateless := !slices.ContainsFunc(w.rules, func(r ak.FieldRule) bool { return r.Name == "state" })
	switch {
	case !w.f.Descriptor.Provisionable:
		out.WriteString("Deployed: not applicable; this family is not provisioned.\n")
	case stateless:
		fmt.Fprintf(&out, "Deployed: unknown; %ss carry no state. If it was provisioned, it stays live until a provision that includes the delete.\n", w.f.Noun)
	case deployed:
		fmt.Fprintf(&out, "Deployed: yes (state %s).\n", entityState)
	default:
		fmt.Fprintf(&out, "Deployed: no, never deployed (state %s).\n", orNone(entityState))
	}
	if w.f.Implicit != nil && name != "" {
		if implicit := w.f.Implicit(name); len(implicit) > 0 {
			fmt.Fprintf(&out, "Also deleted with it: %s.\n", quoteAll(implicit))
		}
	}
	if w.f.ImplicitNote != nil {
		if note := w.f.ImplicitNote(name, raw); note != "" {
			fmt.Fprintf(&out, "Also deleted with it: %s.\n", note)
		}
	}
	if w.f.Descriptor.MarkedRowLifecycle && deployed {
		out.WriteString("The delete now: marks it for deletion; it keeps its name and references until the next provision.\n")
	} else {
		out.WriteString("The delete now: removes it at once.\n")
	}
	switch {
	case w.f.ProvisionNote != "":
		fmt.Fprintf(&out, "The next provision: %s\n", w.f.ProvisionNote)
	case !w.f.Descriptor.Provisionable:
		out.WriteString("The next provision: no effect.\n")
	case stateless:
		out.WriteString("The next provision: tears it down if it was deployed; otherwise nothing to tear down.\n")
	case deployed:
		out.WriteString("The next provision: tears down the deployed object.\n")
	default:
		out.WriteString("The next provision: nothing to tear down.\n")
	}
	fmt.Fprintf(&out, "To delete it, call %s_delete with id %s and confirmCode %s before %s.\n",
		w.f.Name, id, code, exp.UTC().Format(time.RFC3339))
	fmt.Fprintf(&out, "confirmCode: %s\nexpires: %s\n", code, exp.UTC().Format(time.RFC3339))
	out.WriteString("Object:\n")
	out.WriteString(w.render(raw))
	return mcp.NewToolResultText(out.String())
}

// marksRow reports whether a delete leaves a marked row until the next provision.
func (w writeCall[T]) marksRow(raw []byte) (bool, error) {
	deployed, err := ak.Deployed(w.f.Descriptor, raw)
	return w.f.Descriptor.MarkedRowLifecycle && deployed, err
}

func stateOf(raw []byte) string {
	var s struct {
		State string `json:"state"`
	}
	_ = json.Unmarshal(raw, &s)
	return s.State
}

func (w writeCall[T]) confirmRefusal(err error, id string) *mcp.CallToolResult {
	switch {
	case errors.Is(err, confirm.ErrExpired):
		return refuse(RefusalConfirmExpired, fmt.Sprintf("confirmation code expired; call %s_delete again with id %s and no confirmCode for a new preview", w.f.Name, id))
	case errors.Is(err, confirm.ErrStale):
		return refuse(RefusalConfirmStale, "the object changed since the preview; review the new preview: call "+w.f.Name+"_delete again with id "+id+" and no confirmCode")
	}
	return refuse(RefusalConfirmMismatch, "confirmation code does not match this object")
}

// writeTail adds the config status, the echo diff and the stored object.
func (w writeCall[T]) writeTail(out *strings.Builder, wr *ak.WriteResult, obj *T, sent []byte, removed []string) {
	if cs := wr.ConfigStatus; cs != nil && cs.ConfigValid != nil && !*cs.ConfigValid {
		r := w.redaction(sent, wr.Previous, wr.Body)
		msgs := make([]string, len(cs.Messages))
		for i, m := range cs.Messages {
			msgs[i] = r.mask(m)
		}
		fmt.Fprintf(out, "Saved with warnings: %s\n", quoteAll(msgs))
	}
	if diff := echoDiff(withoutKeys(sent, wr.FieldsDropped), wr.Body, removed); len(diff) > 0 {
		fmt.Fprintf(out, "The backend did not keep the request as sent: %s.\n", strings.Join(diff, "; "))
	}
	if obj != nil {
		out.WriteString("Object:\n")
		out.WriteString(w.render(wr.Body))
	}
}

// render prints a stored body through the typed struct with every secret
// field and every key of secretKeys masked; keys outside the struct are not shown.
func (w writeCall[T]) render(raw []byte) string {
	var obj T
	if len(bytes.TrimSpace(raw)) == 0 || json.Unmarshal(raw, &obj) != nil {
		return "(no object body)\n"
	}
	typed, _ := json.Marshal(obj)
	var v any
	dec := json.NewDecoder(bytes.NewReader(typed))
	dec.UseNumber()
	if dec.Decode(&v) != nil {
		return "(no object body)\n"
	}
	v = mask(w.rules, v)
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b) + "\n"
}

// Masked replaces a secret value in write tool output.
const Masked = "***"

// secretKeys lists JSON keys masked in write output wherever they appear: the
// credential and inline secrets of 01-api-surface.md section 4.
var secretKeys = []string{
	"ec2SecretKey", "ec2ExternalId", "secretKey", "private_key", "private_key_id", "key",
	"adminPassword", "sicKey", "password", "fmcRegistrationKey", "alkiraAdminPassword",
	"licenseKey", "ionToken", "ionSecret", "authToken", "sharedSecret", "joinToken",
	"activationKey", "configData", "privateKey", "bindPassword", "authKey", "authCode",
	"registrationKey", "masterKey", "registrationPinValue", "activationCode",
	"bgpAuthenticationKey", "accountKey", "apiKey",
	"presharedKeys", "preSharedKey", "localAuthValue", "remoteAuthValue",
	"bgpAuthKeyAlkira", "bgpAuthKey", "localPublicSharedKey", "remotePublicSharedKey",
}

// SecretKeys returns the JSON keys masked in write output wherever they
// appear, in addition to the fields the family's rules mark secret.
func SecretKeys() []string { return slices.Clone(secretKeys) }

// secretKeySet holds secretKeys lower-cased; keys match without regard to
// case, since the backend spells some of them in more than one case.
var secretKeySet = func() map[string]bool {
	m := make(map[string]bool, len(secretKeys))
	for _, k := range secretKeys {
		m[strings.ToLower(k)] = true
	}
	return m
}()

func isSecretKey(k string) bool { return secretKeySet[strings.ToLower(k)] }

// secretField reports whether key k of an object with field rules rules holds
// a secret, and returns the rules of its nested fields; nil rules below a
// field without struct rules leave only secretKeys to apply.
func secretField(rules []ak.FieldRule, k string) (secret bool, sub []ak.FieldRule) {
	for i := range rules {
		if rules[i].Name == k {
			return rules[i].Secret || isSecretKey(k), rules[i].Fields
		}
	}
	return isSecretKey(k), nil
}

// mask walks v with the field rules of its type and replaces every secret value.
func mask(rules []ak.FieldRule, v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, e := range t {
			secret, sub := secretField(rules, k)
			if e != nil && secret {
				t[k] = Masked
				continue
			}
			t[k] = mask(sub, e)
		}
	case []any:
		for i, e := range t {
			t[i] = mask(rules, e)
		}
	}
	return v
}

// echoDiff names the fields of the request the response does not carry as
// sent. Arrays compare without order and strings without case, because the
// backend sorts some arrays and lower-cases some ids. removed lists fields a
// patch set to null; one still present in the echo was kept by the backend.
func echoDiff(sent, echo []byte, removed []string) []string {
	want, err1 := decodeNumbers(sent)
	got, err2 := decodeNumbers(echo)
	wm, ok1 := want.(map[string]any)
	gm, ok2 := got.(map[string]any)
	if err1 != nil || err2 != nil || !ok1 || !ok2 {
		return nil
	}
	var diff []string
	diffObject(wm, gm, "", &diff)
	for _, k := range removed {
		if v, ok := gm[k]; ok && v != nil {
			diff = append(diff, k+" kept by the backend")
		}
	}
	sort.Strings(diff)
	return diff
}

func diffObject(want, got map[string]any, path string, diff *[]string) {
	for k, wv := range want {
		if wv == nil {
			continue
		}
		gv, ok := got[k]
		switch {
		case !ok || gv == nil:
			*diff = append(*diff, path+k+" dropped")
		default:
			wsub, wIsMap := wv.(map[string]any)
			gsub, gIsMap := gv.(map[string]any)
			if wIsMap && gIsMap {
				diffObject(wsub, gsub, path+k+".", diff)
				continue
			}
			wl, wIsList := wv.([]any)
			gl, gIsList := gv.([]any)
			if wIsList && gIsList && len(gl) < len(wl) {
				*diff = append(*diff, fmt.Sprintf("%s%s: %d of %d elements dropped", path, k, len(wl)-len(gl), len(wl)))
				continue
			}
			if !echoes(gv, wv) {
				*diff = append(*diff, path+k+" changed by the backend")
			}
		}
	}
}

// echoes reports whether the response value got carries the sent value want:
// strings compare without case, an object holds every non-null key sent
// (the backend adds ids and derived fields), and an array holds as many
// elements as were sent, each sent element held by its own response element
// in any order, the subset rule holds() applies to list edits.
func echoes(got, want any) bool {
	switch wv := want.(type) {
	case map[string]any:
		gm, ok := got.(map[string]any)
		if !ok {
			return false
		}
		for k, v := range wv {
			if v != nil && !echoes(gm[k], v) {
				return false
			}
		}
		return true
	case []any:
		gl, ok := got.([]any)
		if !ok || len(gl) != len(wv) {
			return false
		}
		// Pair sent and response elements by augmenting paths, so a loose
		// element cannot take the only response element a stricter one fits.
		owner := make([]int, len(gl))
		for i := range owner {
			owner[i] = -1
		}
		var assign func(w int, seen []bool) bool
		assign = func(w int, seen []bool) bool {
			for g := range gl {
				if seen[g] || !echoes(gl[g], wv[w]) {
					continue
				}
				seen[g] = true
				if owner[g] < 0 || assign(owner[g], seen) {
					owner[g] = w
					return true
				}
			}
			return false
		}
		for w := range wv {
			if !assign(w, make([]bool, len(gl))) {
				return false
			}
		}
		return true
	}
	return canonical(got) == canonical(want)
}

// canonical renders v with strings lower-cased and arrays sorted.
func canonical(v any) string {
	switch t := v.(type) {
	case string:
		b, _ := json.Marshal(strings.ToLower(t))
		return string(b)
	case []any:
		parts := make([]string, len(t))
		for i, e := range t {
			parts[i] = canonical(e)
		}
		sort.Strings(parts)
		return "[" + strings.Join(parts, ",") + "]"
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, len(keys))
		for i, k := range keys {
			parts[i] = strconv.Quote(k) + ":" + canonical(t[k])
		}
		return "{" + strings.Join(parts, ",") + "}"
	}
	b, _ := json.Marshal(v)
	return string(b)
}

// errorResult maps a pipeline or client-go error to its tool text (01-api-surface.md section 3).
// r holds the secret values of the call's request and stored bodies; those
// of the response body are added here.
func (w writeCall[T]) errorResult(verb, id, name string, err error, r redaction) *mcp.CallToolResult {
	noun := w.f.Noun
	subject := noun
	if id != "" {
		subject = noun + " id " + id
	}
	var (
		se *ak.StateError
		de *ak.DuplicateError
		oe *ak.OutcomeError
		ae *ak.APIError
		fe *ak.FamilyError
	)
	if errors.As(err, &ae) {
		r = r.add(w.redaction([]byte(ae.Body)))
	}
	switch {
	case errors.As(err, &oe):
		return w.outcomeResult(oe, name)
	case errors.Is(err, ak.ErrValidationPending):
		check := " Validation still running for this object." + validationStateCheck
		target := w.f.Noun + quoted(name)
		switch {
		case id != "" && name != "":
			target += " (id " + id + ")"
		case id != "":
			target += " id " + id
		}
		switch verb {
		case "update":
			return refuse("validation_pending", fmt.Sprintf("Validation of the previous change to %s is still running; retry the update after it finishes.%s%s", target, check, ids(ae)))
		case "delete":
			return refuse("validation_pending", fmt.Sprintf("Validation of %s is still running; retry the delete after it finishes.%s%s", target, check, ids(ae)))
		}
		return refuse("validation_pending", strings.TrimSpace(check)+ids(ae))
	case errors.Is(err, ak.ErrNotEnabled):
		what, flag := w.flagRefusal(err)
		if strings.Contains(flag, " and one of ") {
			return refuse("not_enabled", fmt.Sprintf("%s is not enabled for this tenant (it needs the tenant features %s); nothing was sent.", what, flag))
		}
		return refuse("not_enabled", fmt.Sprintf("%s is not enabled for this tenant (tenant feature %s is off); nothing was sent.", what, flag))
	case errors.Is(err, ak.ErrFlagsUnavailable):
		// No flags could be read and none are cached, so the kill switch is unknown.
		return refuse("flags_unavailable", capital(ak.ErrFlagsUnavailable.Error())+". Nothing was sent.")
	case errors.Is(err, ak.ErrWritesDisabled):
		return refuse("writes_disabled", "Changes through MCP are disabled for this tenant; nothing was sent. Make the change in the portal, or ask a tenant administrator to turn MCP changes back on.")
	case errors.Is(err, ak.ErrNotSent):
		return refuse("not_sent", fmt.Sprintf("The %s of %s was not sent, nothing changed: %v", verb, subject, err))
	case errors.As(err, &se):
		switch {
		case se.DocState == "DELETED":
			return refuse("state_refused", fmt.Sprintf("%s is already deleted; the backend keeps a residue row.", subject))
		case se.DocState != "":
			return refuse("state_refused", fmt.Sprintf("%s is marked for deletion; deletion pending provisioning.", subject))
		case se.State == "FAILED" || se.NetworkState == "FAILED":
			return refuse("state_refused", fmt.Sprintf("%s cannot change while %s. A FAILED object or network must be fixed in the portal or by a successful provision first.", subject, stateText(se)))
		}
		return refuse("state_refused", fmt.Sprintf("%s cannot change while %s; wait for the running provision to finish.", subject, stateText(se)))
	case errors.As(err, &de):
		if de.Group {
			return refuse("duplicate", fmt.Sprintf("a group named %q already exists (id %s); the backend creates a group of that name with the %s, so the name cannot be used.", de.Value, de.ID, w.f.Noun))
		}
		if de.Type != "" {
			return refuse("duplicate", fmt.Sprintf("%s connector named %q already exists (id %s); connector names are unique across every connector type and group.", Article(strings.ToLower(de.Type)), de.Value, de.ID))
		}
		return refuse("duplicate", fmt.Sprintf("%s already exists (id %s).", w.keyed(de.Value), de.ID))
	case errors.Is(err, ak.ErrDuplicate):
		// The backend message can carry raw database constraint text; it is never passed through.
		return refuse("duplicate", w.keyed(name)+" already exists.")
	case errors.Is(err, ak.ErrNotFound):
		return refuse("not_found", fmt.Sprintf("%s not found.", subject))
	case errors.As(err, &fe) && fe.Want != "" && fe.Got != nil:
		return refuse("family_mismatch", fmt.Sprintf("%s has %s %v; only %ss with %s %s can be changed or deleted. Nothing was sent.", capital(subject), fe.Field, fe.Got, noun, fe.Field, fe.Want))
	case errors.Is(err, ak.ErrFamilyMismatch):
		return refuse("family_mismatch", fmt.Sprintf("id %s belongs to another family, not %s; nothing was changed.", id, Article(noun)))
	case errors.Is(err, ak.ErrInUse):
		return refuse("in_use", fmt.Sprintf("%s is still referenced: %s. A referrer deleted without a provision keeps the reference until the next provision.", subject, backendText(ae, r)))
	case errors.Is(err, ak.ErrRateLimited):
		if ae != nil && ae.RetryAfter > 0 {
			return refuse("rate_limited", fmt.Sprintf("The backend is busy or rate limiting writes; retry after %d s.%s", int(ae.RetryAfter.Seconds()), ids(ae)))
		}
		return refuse("rate_limited", "Rate limit reached; retry later."+ids(ae))
	case errors.Is(err, ak.ErrLimitExceeded):
		return refuse("limit_exceeded", fmt.Sprintf("The tenant is at its limit: %s", backendText(ae, r)))
	case errors.Is(err, ak.ErrStateRefused) && ae == nil:
		// The object's state could not be read from its GET body; nothing was sent.
		reason := err.Error()
		if i := strings.Index(reason, ak.ErrStateRefused.Error()+": "); i >= 0 {
			reason = reason[i+len(ak.ErrStateRefused.Error())+2:]
		}
		return refuse("state_refused", fmt.Sprintf("%s was not changed: its state could not be checked (%s); nothing was sent. Read it with %s_get_by_id.", capital(subject), reason, w.f.Name))
	case errors.Is(err, ak.ErrStateRefused):
		return refuse("state_refused", fmt.Sprintf("%s: %s Wait for the running provision to finish; a FAILED object or network must be fixed first.", subject, backendText(ae, r)))
	case errors.Is(err, ak.ErrInvalidInput):
		return refuse("invalid_arguments", "Invalid input: "+strings.TrimPrefix(err.Error(), ak.ErrInvalidInput.Error()+": "))
	case errors.Is(err, context.DeadlineExceeded):
		// client-go reports a deadline after the write went out as an
		// OutcomeError, so a bare deadline came before the write was sent.
		return refuse("timeout", fmt.Sprintf("The %s of %s timed out after %s before the write was sent; nothing changed. Retry the %s.", verb, subject, w.p.timeout, verb))
	case ae != nil && ae.Status == 403:
		return refuse("permission_denied", fmt.Sprintf("Permission denied for %s %s: %s", verb, noun, backendText(ae, r)))
	case ae != nil && ae.Status >= 500:
		// A 5xx message can carry raw database or stack text; only the ids are passed on.
		return refuse("backend_error", fmt.Sprintf("%s %s failed with status %d%s.", capital(verb), noun, ae.Status, ids(ae)))
	case ae != nil:
		return refuse("backend_error", fmt.Sprintf("%s %s failed with status %d: %s", capital(verb), noun, ae.Status, backendText(ae, r)))
	}
	return refuse("error", fmt.Sprintf("%s %s failed: %v", capital(verb), noun, err))
}

// outcomeResult reports a lost response with the flag effects of the
// request that was sent.
func (w writeCall[T]) outcomeResult(oe *ak.OutcomeError, name string) *mcp.CallToolResult {
	const lost = "the response was lost"
	key := w.f.Descriptor.Uniqueness.Key
	object := fmt.Sprintf("%s with %s %q", w.f.Noun, key, name)
	async := slices.Contains(oe.FlagsApplied, ak.AsyncValidationFlag)
	text := func(t string) string {
		var out strings.Builder
		out.WriteString(t)
		if oe.FlagsApplied != nil || oe.FieldsDropped != nil {
			out.WriteString("\n")
		}
		w.flagEffects(&out, &ak.WriteResult{FlagsApplied: oe.FlagsApplied, FieldsDropped: oe.FieldsDropped, FieldsCleared: oe.FieldsCleared})
		switch {
		case async && oe.Outcome == ak.OutcomeCreated:
			// A lost async create found by lookup has no validation state id (ak.ValidationUnknown).
			fmt.Fprintf(&out, "Validation state unknown: the validation state id was lost with the response, so validation may still be running or may have FAILED. Check the object with %s_get_by_id before provisioning.\n", w.f.Name)
		case async && oe.Outcome != ak.OutcomeMayStillBeCreated:
			fmt.Fprintf(&out, "The write was sent for async validation: validation may still be running; check the object with %s_get_by_id.\n", w.f.Name)
		}
		return strings.TrimSuffix(out.String(), "\n")
	}
	switch oe.Outcome {
	case ak.OutcomeApplied:
		return mcp.NewToolResultText(text(fmt.Sprintf("Updated %s id %s: applied (%s; a confirming read shows every change).", w.f.Noun, oe.ID, lost)))
	case ak.OutcomeMayStillApply:
		return refuse("may_still_apply", text(fmt.Sprintf("Update of %s id %s may still apply (%s; a confirming read does not show the change yet, and the backend can still commit it). Read it with %s_get_by_id before retrying.", w.f.Noun, oe.ID, lost, w.f.Name)))
	case ak.OutcomeCreated:
		return mcp.NewToolResultText(text(fmt.Sprintf("Created %s, id %s (%s; a lookup by %s found the object).", object, oe.ID, lost, key)))
	case ak.OutcomeMayStillBeCreated:
		return refuse("may_still_be_created", text(fmt.Sprintf("%s may still be created (%s; a lookup by %s found no object yet). List the %ss and look for %s %q before creating it again.", capital(object), lost, key, w.f.Noun, key, name)))
	case ak.OutcomeDeleted:
		return mcp.NewToolResultText(text(fmt.Sprintf("Deleted %s id %s (%s; a confirming read finds no object).", w.f.Noun, oe.ID, lost)))
	}
	if oe.Method == "POST" && (w.f.Descriptor.Uniqueness.Scope == ak.ScopeNone || key == "" || name == "") {
		// client-go looks a lost create up only by a uniqueness key value.
		return refuse("outcome_unknown", text(fmt.Sprintf("Create of %s: outcome unknown (%s: %v). List the %ss and check for the new object before creating it again.",
			w.f.Noun, lost, oe.Err, w.f.Noun)))
	}
	if oe.Method == "POST" {
		return refuse("outcome_unknown", text(fmt.Sprintf("Create of %s: outcome unknown (%s and the lookup by %s did not finish: %v). List the %ss and look for %s %q before creating it again.",
			object, lost, key, oe.Err, w.f.Noun, key, name)))
	}
	what := "outcome unknown"
	if oe.Method == "DELETE" {
		what = "delete outcome unknown"
		if oe.Current != nil {
			what += ", a confirming read still finds the object"
		}
	}
	return refuse("outcome_unknown", text(fmt.Sprintf("%s %s: %s (%v). Read it with %s_get_by_id before retrying.", capital(w.f.Noun), oe.ID, what, oe.Err, w.f.Name)))
}

// flagEffects names the flags that changed the request, the fields left out
// and the stored values sent empty (D16).
func (w writeCall[T]) flagEffects(out *strings.Builder, wr *ak.WriteResult) {
	if len(wr.FlagsApplied) > 0 {
		fmt.Fprintf(out, "Tenant feature flags that shaped the request: %s.\n", strings.Join(wr.FlagsApplied, ", "))
	}
	if len(wr.FieldsDropped) > 0 {
		fmt.Fprintf(out, "Left out because their tenant feature is off: %s.\n", strings.Join(wr.FieldsDropped, ", "))
	}
	var cleared []string
	for _, f := range wr.FieldsCleared {
		cleared = append(cleared, f+w.clearReason(f))
	}
	if len(cleared) > 0 {
		fmt.Fprintf(out, "Stored values sent empty or false, as the portal does, because the backend refuses them under the current tenant features: %s.\n", strings.Join(cleared, ", "))
	}
}

// clearReason names the tenant feature behind a field client-go sent empty.
func (w writeCall[T]) clearReason(field string) string {
	for _, r := range w.f.Descriptor.Flags {
		switch {
		case r.Field != field:
		case r.Effect == ak.FlagClearField:
			return fmt.Sprintf(" (tenant feature %s is on)", r.Flag)
		case r.ClearStored:
			return fmt.Sprintf(" (tenant feature %s is off)", r.Flag)
		}
	}
	return ""
}

// validationRunning reports an async validation that had not finished (Q-A2).
func validationRunning(out *strings.Builder, wr *ak.WriteResult) {
	switch {
	case wr.ValidationState == ak.ValidationRunning && wr.ValidationReadError != nil:
		fmt.Fprintf(out, "Saved; the validation state (id %s) could not be read, so its result is unknown: %v.%s\n", wr.ValidationStateID, wr.ValidationReadError, validationStateCheck)
	case wr.ValidationState == ak.ValidationRunning:
		fmt.Fprintf(out, "Saved; validation is still running (state id %s).%s\n", wr.ValidationStateID, validationStateCheck)
	}
}

// validationFailed reports a save whose async validation FAILED. The save
// stays, so the text says how to fix or remove it, never "not saved".
// Secret values the request sent, or the stored object held, are masked in
// the echoed validation message.
func (w writeCall[T]) validationFailed(out *strings.Builder, saved string, wr *ak.WriteResult, sent []byte) {
	msg := "no details"
	if ve := wr.ValidationErrors; ve != nil {
		msg = backendText(&ak.APIError{Message: ve.Message, Errors: ve.Errors, Code: ve.Code, BackendID: ve.ID}, w.redaction(sent, wr.Previous, wr.Body))
	}
	fmt.Fprintf(out, "%s (validation state id %s), but validation FAILED: %s; fix it with %s_update or remove it with %s_delete. It blocks provisioning until it is fixed or deleted.\n",
		saved, wr.ValidationStateID, msg, w.f.Name, w.f.Name)
}

// secretValues appends every non-empty string held in a secret field or a
// secretKeys key of v, at any depth.
func secretValues(rules []ak.FieldRule, v any, acc []string) []string {
	var strs func(any)
	strs = func(v any) {
		switch t := v.(type) {
		case string:
			if t != "" && !slices.Contains(acc, t) {
				acc = append(acc, t)
			}
		case map[string]any:
			for _, e := range t {
				strs(e)
			}
		case []any:
			for _, e := range t {
				strs(e)
			}
		}
	}
	switch t := v.(type) {
	case map[string]any:
		for k, e := range t {
			secret, sub := secretField(rules, k)
			if secret {
				strs(e)
				continue
			}
			acc = secretValues(sub, e, acc)
		}
	case []any:
		for _, e := range t {
			acc = secretValues(rules, e, acc)
		}
	}
	return acc
}

// sent is the JSON of a typed create body.
func sent(obj any) []byte {
	b, _ := json.Marshal(obj)
	return b
}

// withoutKeys returns the JSON object sent without keys.
func withoutKeys(sent []byte, keys []string) []byte {
	if len(keys) == 0 {
		return sent
	}
	v, err := decodeNumbers(sent)
	m, ok := v.(map[string]any)
	if err != nil || !ok {
		return sent
	}
	// A key may be a dotted path to a nested field.
	for _, k := range keys {
		path := strings.Split(k, ".")
		parent := m
		for _, p := range path[:len(path)-1] {
			parent, _ = parent[p].(map[string]any)
		}
		delete(parent, path[len(path)-1])
	}
	b, _ := json.Marshal(m)
	return b
}

// absentText reports removed items the stored lists did not hold.
func absentText(out *strings.Builder, absent map[string][]any) {
	fields := make([]string, 0, len(absent))
	for k, items := range absent {
		if len(items) > 0 {
			fields = append(fields, k)
		}
	}
	sort.Strings(fields)
	for _, k := range fields {
		items := make([]string, len(absent[k]))
		for i, it := range absent[k] {
			b, _ := json.Marshal(it)
			items[i] = string(b)
		}
		fmt.Fprintf(out, "Not in the stored %s, nothing to remove: %s.\n", k, strings.Join(items, ", "))
	}
}

// editsDiff names added items the response does not hold and removed items
// it still holds, with the comparison client-go uses: exact JSON values, an
// added object held by any element carrying every key it sent, and removes
// on id-matched arrays matched by element id. Removed items reported absent
// are skipped.
func editsDiff(edits ak.ListEdits, echo []byte, absent map[string][]any, idMatched []string) []string {
	got, err := decodeNumbers(echo)
	gm, ok := got.(map[string]any)
	if err != nil || !ok {
		return nil
	}
	var diff []string
	for k, raw := range edits.Add {
		items, _ := decodeNumbers(raw)
		list, _ := items.([]any)
		stored, _ := gm[k].([]any)
		for _, it := range list {
			if !slices.ContainsFunc(stored, func(e any) bool { return holds(e, it) }) {
				diff = append(diff, fmt.Sprintf("%s: %s not added", k, jsonText(it)))
			}
		}
	}
	for k, raw := range edits.Remove {
		items, _ := decodeNumbers(raw)
		list, _ := items.([]any)
		stored, _ := gm[k].([]any)
		byID := slices.Contains(idMatched, k)
		for _, it := range list {
			if slices.ContainsFunc(absent[k], func(a any) bool { return jsonText(a) == jsonText(it) }) {
				continue
			}
			still := slices.ContainsFunc(stored, func(e any) bool {
				if byID {
					m, ok := e.(map[string]any)
					return ok && m["id"] != nil && fmt.Sprint(m["id"]) == fmt.Sprint(it)
				}
				return jsonText(e) == jsonText(it)
			})
			if still {
				what := jsonText(it)
				if byID {
					what = "element id " + fmt.Sprint(it)
				}
				diff = append(diff, fmt.Sprintf("%s: %s not removed", k, what))
			}
		}
	}
	sort.Strings(diff)
	return diff
}

// holds reports whether stored carries sent: every key of a sent object with
// a value it holds, any other value equal as JSON.
func holds(stored, sent any) bool {
	sm, ok := sent.(map[string]any)
	if !ok {
		return jsonText(stored) == jsonText(sent)
	}
	cm, ok := stored.(map[string]any)
	if !ok {
		return false
	}
	for k, v := range sm {
		if !holds(cm[k], v) {
			return false
		}
	}
	return true
}

// jsonText renders v as JSON; map keys come out sorted.
func jsonText(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// editFields lists the fields the add or remove argument takes, from the
// same generated schema the tool advertises.
func (w writeCall[T]) editFields(remove bool) []string {
	s := w.p.schema.ListEdits(w.rules, remove, w.f.Descriptor.IDMatchedArrays)
	if s == nil {
		return nil
	}
	var names []string
	for k := range s["properties"].(map[string]any) {
		names = append(names, k)
	}
	return names
}

// listEditArg returns the add or remove argument as field -> JSON array. A
// field outside allowed is refused, so secret, read-only and plain fields
// never reach the request or the output.
func listEditArg(args map[string]any, key string, allowed []string) (map[string]json.RawMessage, error) {
	v, set := args[key]
	if !set || v == nil {
		return nil, nil
	}
	m, ok := v.(map[string]any)
	if !ok || len(m) == 0 {
		return nil, fmt.Errorf("%s must be a non-empty object of field name to list of items", key)
	}
	var errs []string
	out := make(map[string]json.RawMessage, len(m))
	for k, items := range m {
		if !slices.Contains(allowed, k) {
			errs = append(errs, fmt.Sprintf("%s.%s: %s is not a list that %s applies to", key, k, k, key))
			continue
		}
		if l, ok := items.([]any); !ok || len(l) == 0 {
			errs = append(errs, fmt.Sprintf("%s.%s must be a non-empty array", key, k))
			continue
		}
		out[k], _ = json.Marshal(items)
	}
	if errs != nil {
		sort.Strings(errs)
		return nil, errors.New(strings.Join(errs, "; "))
	}
	return out, nil
}

// flagRefusal returns what a flag refusal names (the family, or a field) and
// the flag, from client-go's "<what>: <error> (needs <flag>)".
func (w writeCall[T]) flagRefusal(err error) (what, flag string) {
	msg := err.Error()
	what, _, _ = strings.Cut(msg, ": ")
	if i := strings.LastIndex(msg, "(needs "); i >= 0 {
		flag = strings.TrimSuffix(msg[i+len("(needs "):], ")")
	}
	if what == w.f.Descriptor.Collection || what == msg {
		return capital(w.f.Noun), flag
	}
	return "Field " + what, flag
}

func stateText(se *ak.StateError) string {
	if se.NetworkState != "" {
		return fmt.Sprintf("its state is %s and the network state is %s", orNone(se.State), se.NetworkState)
	}
	return "its state is " + se.State
}

// keyed names an object by its uniqueness key value, for example
// `a widget named "w1"`, or `a widget with <key> "v"` when the key is not name.
func (w writeCall[T]) keyed(value string) string {
	key := w.f.Descriptor.Uniqueness.Key
	switch {
	case value == "" && (key == "" || key == "name"):
		return Article(w.f.Noun) + " with this name"
	case value == "":
		return fmt.Sprintf("%s with this %s", Article(w.f.Noun), key)
	case key == "" || key == "name":
		return fmt.Sprintf("%s named %q", Article(w.f.Noun), value)
	}
	return fmt.Sprintf("%s with %s %q", Article(w.f.Noun), key, value)
}

// danglingRef names a stored reference whose target is gone when the backend
// refused the update on a field the patch does not set.
func (w writeCall[T]) danglingRef(ctx context.Context, client *ak.AlkiraClient, core *ak.Core[T], id string, patch []byte, err error) string {
	var ae *ak.APIError
	if len(w.f.StoredRefs) == 0 || !errors.As(err, &ae) {
		return ""
	}
	var set map[string]json.RawMessage
	_ = json.Unmarshal(patch, &set)
	for _, ref := range w.f.StoredRefs {
		if _, ok := set[ref.Field]; ok || !slices.ContainsFunc(ae.Errors, func(fe ak.FieldError) bool { return fe.Field == ref.Field }) {
			continue
		}
		_, raw, gerr := core.Get(ctx, id)
		if gerr != nil {
			return ""
		}
		var stored map[string]json.RawMessage
		if json.Unmarshal(raw, &stored) != nil || len(stored[ref.IDField]) == 0 {
			return ""
		}
		refID := strings.Trim(string(stored[ref.IDField]), `"`)
		if !errors.Is(ref.Lookup(ctx, client, refID), ak.ErrNotFound) {
			continue
		}
		return fmt.Sprintf("The stored %s %s names a %s that no longer exists, so the backend refused the stored reference, not the patch. "+
			"Set %s to an existing %s in the patch, or delete this %s.", ref.IDField, refID, ref.Noun, ref.Field, ref.Noun, w.f.Noun)
	}
	return ""
}

// backendText is the backend message, masked with r and quoted, with its
// error code, error id and request id.
func backendText(ae *ak.APIError, r redaction) string {
	if ae == nil {
		return ""
	}
	if r.withheld {
		return "(backend message withheld: the stored object could not be read to mask its secrets)" + ids(ae)
	}
	msg := ae.Message
	for _, fe := range ae.Errors {
		msg += fmt.Sprintf("; %s: %s", fe.Field, fe.Message)
	}
	return strconv.Quote(r.mask(strings.TrimSpace(msg))) + ids(ae)
}

// redaction holds the secret values that backend text of one call may echo.
type redaction struct {
	values []string
	// withheld reports that the secrets are not all known, so backend text is not shown.
	withheld bool
}

// redaction collects the secret values of JSON bodies: those of a field the
// rules mark secret, or of a secretKeys key, at any depth.
func (w writeCall[T]) redaction(bodies ...[]byte) redaction {
	var r redaction
	for _, b := range bodies {
		if len(bytes.TrimSpace(b)) == 0 {
			continue
		}
		if v, err := decodeNumbers(b); err == nil {
			r.values = secretValues(w.rules, v, r.values)
		}
	}
	return r
}

func (r redaction) add(o redaction) redaction {
	for _, v := range o.values {
		if !slices.Contains(r.values, v) {
			r.values = append(r.values, v)
		}
	}
	r.withheld = r.withheld || o.withheld
	return r
}

// mask replaces every secret value in s, longest first, so a secret that
// contains another is masked whole. Backend text can carry a value as it
// appears inside a JSON string, so the JSON-escaped forms are masked too.
func (r redaction) mask(s string) string {
	var forms []string
	for _, v := range r.values {
		for _, f := range []string{v, jsonEscaped(v, true), jsonEscaped(v, false)} {
			if !slices.Contains(forms, f) {
				forms = append(forms, f)
			}
		}
	}
	sort.Slice(forms, func(i, j int) bool { return len(forms[i]) > len(forms[j]) })
	for _, f := range forms {
		s = strings.ReplaceAll(s, f, Masked)
	}
	return s
}

// jsonEscaped returns v as it appears between the quotes of a JSON string,
// with or without HTML escaping of <, > and &.
func jsonEscaped(v string, html bool) string {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(html)
	_ = enc.Encode(v)
	out := strings.TrimSuffix(b.String(), "\n")
	return out[1 : len(out)-1]
}

// storedReadTimeout bounds the read of the stored object after a failed update.
const storedReadTimeout = 10 * time.Second

// updateRedaction returns the secrets that backend text of a failed update
// may echo. The PUT resends the stored object with its secrets, and client-go
// returns no stored body with an error, so the object is read again when the
// error carries backend text; when that read fails, the text is withheld.
func (w writeCall[T]) updateRedaction(ctx context.Context, core *ak.Core[T], id string, patch []byte, err error) redaction {
	r := w.redaction(patch)
	var ae *ak.APIError
	if !errors.As(err, &ae) || ae.Message == "" && len(ae.Errors) == 0 {
		return r
	}
	rctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), storedReadTimeout)
	defer cancel()
	_, stored, gerr := core.Get(rctx, id)
	if gerr != nil {
		r.withheld = true
		return r
	}
	return r.add(w.redaction(stored))
}

// quoteAll quotes each of texts and joins them with ", ".
func quoteAll(texts []string) string {
	q := make([]string, len(texts))
	for i, t := range texts {
		q[i] = strconv.Quote(t)
	}
	return strings.Join(q, ", ")
}

func ids(ae *ak.APIError) string {
	if ae == nil {
		return ""
	}
	var parts []string
	if ae.Code != "" {
		parts = append(parts, "code "+ae.Code)
	}
	if ae.BackendID != "" {
		parts = append(parts, "error id "+ae.BackendID)
	}
	if ae.RequestID != "" {
		parts = append(parts, "request id "+ae.RequestID)
	}
	if parts == nil {
		return ""
	}
	return " (" + strings.Join(parts, ", ") + ")"
}

func refuse(code, text string) *mcp.CallToolResult {
	res := mcp.NewToolResultError(text)
	res.Meta = map[string]any{"refusal": Refusal{Code: code, Text: text}}
	return res
}

func invalidArgs(err error) *mcp.CallToolResult {
	return refuse("invalid_arguments", "Invalid arguments: "+strings.ReplaceAll(err.Error(), "\n", "; "))
}

// toolArgs checks the top-level tool arguments.
func toolArgs(req mcp.CallToolRequest, allowed, required []string) (map[string]any, *mcp.CallToolResult) {
	args := req.GetArguments()
	var errs []string
	for k := range args {
		if !slices.Contains(allowed, k) {
			errs = append(errs, "unknown argument "+k)
		}
	}
	for _, k := range required {
		if args[k] == nil {
			errs = append(errs, k+" is required")
		}
	}
	if errs != nil {
		sort.Strings(errs)
		return nil, invalidArgs(errors.New(strings.Join(errs, "; ")))
	}
	return args, nil
}

var numericID = regexp.MustCompile(`^[0-9]+$`)

// idArg returns the id argument; a non-numeric path id makes the backend answer 500.
func idArg(args map[string]any) (string, *mcp.CallToolResult) {
	id, ok := args["id"].(string)
	if !ok || !numericID.MatchString(id) {
		return "", invalidArgs(errors.New("id must be a string of digits"))
	}
	return id, nil
}

// tenantOf identifies the tenant a client serves, for binding delete codes.
func tenantOf(c *ak.AlkiraClient) string {
	host := c.HostHeader
	if host == "" {
		host = c.URI
	}
	return host + "#" + c.TenantNetworkId
}

func decodeNumbers(b []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var v any
	err := dec.Decode(&v)
	return v, err
}

// nameOf returns the field that names an object of the family in tool output.
func (w writeCall[T]) nameOf(body []byte) string {
	if w.f.NameKey != "" {
		return fieldOf(body, w.f.NameKey)
	}
	return fieldOf(body, "name")
}

// fieldOf returns a top-level scalar field of a JSON object as text, or "".
func fieldOf(body []byte, key string) string {
	var m map[string]any
	if key == "" || json.Unmarshal(body, &m) != nil || m[key] == nil {
		return ""
	}
	return fmt.Sprint(m[key])
}

func orUnknown(id string) string {
	if id == "" {
		return "unknown"
	}
	return id
}

func nullKeys(patch []byte) []string {
	var m map[string]json.RawMessage
	_ = json.Unmarshal(patch, &m)
	var keys []string
	for k, v := range m {
		if string(bytes.TrimSpace(v)) == "null" {
			keys = append(keys, k)
		}
	}
	return keys
}

func quoted(name string) string {
	if name == "" {
		return ""
	}
	return " " + strconv.Quote(name)
}

func orNone(s string) string {
	if s == "" {
		return "none"
	}
	return s
}

// Article returns noun with its indefinite article: "an" when the noun
// starts with a vowel letter, else "a". A first word that starts with a
// vowel letter but is spoken with a consonant sound takes "a", as in
// "a UDR list".
func Article(noun string) string {
	first, _, _ := strings.Cut(noun, " ")
	if noun == "" || first == "UDR" || !strings.ContainsRune("aeiouAEIOU", rune(noun[0])) {
		return "a " + noun
	}
	return "an " + noun
}

func capital(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
