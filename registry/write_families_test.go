package registry

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
	"github.com/alkiranet/mcp-alkira/internal/confirm"
	"github.com/alkiranet/mcp-alkira/writes"
	"github.com/mark3labs/mcp-go/server"
)

// familySpec is one served write family with its fake-portal setup and its
// cases of the shared suites below. Families in their own files append their
// specs to familySpecs in init.
type familySpec struct {
	name     string
	noun     string // tool-output noun; the name with spaces when empty
	f        writeFamily
	coll     string
	path     string // tenant-level collection path, when set
	notFound fault  // GET answer for a missing id
	flags    []string
	nameKey  string // field that names an object, when not name
	nameMax  int    // name maxlen, 256 when zero
	stored   func() map[string]any
	// create returns a valid create body named name; {nameKey: name} when nil.
	create func(name string) map[string]any
	// connector is the GET /connectors row type of a connector family: each
	// stored object gets a row there, which carries the object's state.
	connector string
	// readOnly lists stored fields of the family an update must not resend.
	readOnly []string
	// updateGets is the number of GETs of the family path an update sends; 1 when zero.
	updateGets int

	creates  []createCase        // TestFamilies_CreateRequestShape
	invalid  map[string]toolCase // TestFamilies_InvalidArgumentsSendNothing; a zero case drops a shared one
	inUse    []inUseCase         // TestFamilies_DeleteInUse
	previews []previewCase       // TestFamilies_DeletePreview
}

// createCase is a create body and the POST body it must send with flags on.
type createCase struct {
	flags      []string
	body, want map[string]any
	// named lists the flags of flags the output names as shaping the request.
	named []string
}

type toolCase struct {
	tool string
	args map[string]any
	want string
}

type inUseCase struct {
	f    fault
	want string
}

// previewCase is a stored state ("" keeps the stored body) and the lines its delete preview holds.
type previewCase struct {
	state string
	want  []string
}

func (s familySpec) nounText() string {
	if s.noun != "" {
		return s.noun
	}
	return strings.ReplaceAll(s.name, "_", " ")
}

func (s familySpec) nameField() string {
	if s.nameKey != "" {
		return s.nameKey
	}
	return "name"
}

func (s familySpec) createBody(name string) map[string]any {
	if s.create != nil {
		return s.create(name)
	}
	return map[string]any{s.nameField(): name}
}

func (s familySpec) maxName() int {
	if s.nameMax != 0 {
		return s.nameMax
	}
	return 256
}

// put stores o under id; for a connector family the state moves to its GET /connectors row.
func (s familySpec) put(e *writeEnv, id int, o map[string]any) {
	if s.connector != "" {
		state, _ := o["state"].(string)
		if state == "" {
			state = "PENDING"
		}
		delete(o, "state")
		e.portal.connectors = append(e.portal.connectors, map[string]any{"id": id, "name": o["name"], "type": s.connector,
			"state": state, "previousState": state})
	}
	e.portal.put(id, o)
}

var familySpecs = []familySpec{
	{
		name: "segment", f: family(writes.Segment), coll: "segments",
		notFound: fault{status: 400, body: `{"code":"AK000002","message":"Segment with id: '404' was not found"}`},
		stored: func() map[string]any {
			return map[string]any{"name": "s7", "description": "seven", "asn": 65514, "ipBlocks": map[string]any{"values": []any{"10.1.0.0/24"}},
				"state": "PENDING", "segmentNo": 3, "internalName": "seg-7", "enableIpv6ToIpv4Translation": false,
				"serviceTrafficDistribution": map[string]any{"algorithm": "HASHING", "algorithmAttributes": map[string]any{"keys": "DEFAULT"}}}
		},
		readOnly: []string{"state", "segmentNo", "internalName"},
		creates: []createCase{
			{nil, map[string]any{"name": "s1", "ipBlocks": map[string]any{"values": []any{"10.2.0.0/24"}}},
				map[string]any{"name": "s1", "ipBlocks": map[string]any{"values": []any{"10.2.0.0/24"}}, "asn": 65514,
					"serviceTrafficDistribution": segmentHashing()}, nil},
			{[]string{"OVERLAP-ROUTE-INVALIDATION-ON-SEGMENT"}, map[string]any{"name": "s1", "description": ""},
				map[string]any{"name": "s1", "description": "", "asn": 65514,
					"serviceTrafficDistribution": segmentHashing(), "overlappingRouteValidationEnabled": true}, []string{"OVERLAP-ROUTE-INVALIDATION-ON-SEGMENT"}},
			{[]string{"IFA-IPV6-SUPPORT"}, map[string]any{"name": "s1", "asn": 65001, "enableIpv6ToIpv4Translation": false},
				map[string]any{"name": "s1", "asn": 65001, "enableIpv6ToIpv4Translation": false, "serviceTrafficDistribution": segmentHashing()}, []string{"IFA-IPV6-SUPPORT"}},
		},
		invalid: map[string]toolCase{
			"asn range":       {"_create", map[string]any{"body": map[string]any{"name": "a", "asn": 64512}}, "asn must be in"},
			"asn zero":        {"_update", map[string]any{"id": "7", "patch": map[string]any{"asn": 0}}, "asn must be in"},
			"read-only state": {"_update", map[string]any{"id": "7", "patch": map[string]any{"state": "SUCCESS"}}, "state is read-only"},
		},
		inUse: []inUseCase{
			{fault{status: 400, body: `{"code":"AK000002","message":"segment with id : 7 is associated with user created segment resources"}`}, "is associated with user created segment resources"},
			{fault{status: 409, body: `{"code":"AK000001","message":"Segment is used by policy 4"}`}, "Segment is used by policy 4"},
		},
		previews: []previewCase{
			{"PENDING", []string{`Delete preview for segment "s7" (id 7). Nothing was deleted.`, "Deployed: no, never deployed (state PENDING).",
				`Also deleted with it: "s7-ALL", "s7-ALL-Group".`, "The delete now: removes it at once.", "The next provision: nothing to tear down."}},
			{"SUCCESS", []string{"Deployed: yes (state SUCCESS).", `Also deleted with it: "s7-ALL", "s7-ALL-Group".`, "The delete now: removes it at once.",
				"The next provision: tears down the deployed object."}},
		},
	},
	{
		name: "group", f: family(writes.Group), coll: "groups",
		notFound: fault{status: 404, body: `{"code":"AK000001","message":"Group not found"}`},
		stored: func() map[string]any {
			return map[string]any{"name": "g7", "description": "seven", "internalName": "grp-7", "implicit": false, "tags": []any{}, "type": "EXPLICIT"}
		},
		readOnly: []string{"internalName", "implicit", "type"},
		creates:  []createCase{{nil, map[string]any{"name": "g1", "description": ""}, map[string]any{"name": "g1", "description": ""}, nil}},
		inUse: []inUseCase{
			{fault{status: 400, body: `{"code":"AK000002","message":"Validation failed","errors":[{"field":"group","code":"group.00000010","message":"used by connector 3"}]}`}, "used by connector 3"},
			{fault{status: 409, body: `{"code":"AK000001","message":"Group is used by a segment resource"}`}, "Group is used by a segment resource"},
		},
		previews: []previewCase{
			{"", []string{`Delete preview for group "g7" (id 7).`, "Deployed: unknown; groups carry no state. If it was provisioned, it stays live until a provision that includes the delete.",
				"The delete now: removes it at once.", "The next provision: tears it down if it was deployed; otherwise nothing to tear down."}},
		},
	},
	{
		name: "billing_tag", f: family(writes.BillingTag), coll: "tags", path: "/api/tags", flags: []string{"BILLING-VIEW", "BILLING-TAGS"},
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"Tag not found"}`},
		stored: func() map[string]any {
			return map[string]any{"name": "t7", "description": "seven", "tenantId": 9, "instances": []any{}}
		},
		readOnly: []string{"tenantId", "instances"},
		creates:  []createCase{{[]string{"BILLING-VIEW", "BILLING-TAGS"}, map[string]any{"name": "t1"}, map[string]any{"name": "t1"}, nil}},
		inUse: []inUseCase{
			{fault{status: 400, body: `{"code":"AK000002","message":"Validation failed","errors":[{"field":"tag","code":"tag.00000002","message":"tag is in use"}]}`}, "tag is in use"},
		},
		previews: []previewCase{
			{"", []string{`Delete preview for billing tag "t7" (id 7).`, "Deployed: not applicable; this family is not provisioned.", "The next provision: no effect."}},
		},
	},
}

func segmentHashing() map[string]any {
	return map[string]any{"algorithm": "HASHING", "algorithmAttributes": map[string]any{"keys": "DEFAULT"}}
}

// withFields returns a copy of m with extra set.
func withFields(m, extra map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range m {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

// jsonRoundTrip returns v as a JSON decode gives it back, numbers as float64.
func jsonRoundTrip(v map[string]any) map[string]any {
	b, _ := json.Marshal(v)
	var out map[string]any
	json.Unmarshal(b, &out)
	return out
}

func newFamilyEnv(t *testing.T, s familySpec) *writeEnv {
	t.Helper()
	e := &writeEnv{t: t, portal: newWritePortal(t, s.coll), now: time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC), key: confirm.RandomKey()}
	e.portal.path, e.portal.notFound, e.portal.flags = s.path, &s.notFound, s.flags
	e.serve = func() *server.MCPServer {
		srv := server.NewMCPServer("mcp-alkira", "test")
		s.f.register(srv, providerFor(e.portal.client()), e.p)
		return srv
	}
	e.setPipeline()
	return e
}

func specNamed(name string) familySpec {
	for _, s := range familySpecs {
		if s.name == name {
			return s
		}
	}
	panic(name)
}

func sentBody(t *testing.T, e *writeEnv, method string) map[string]any {
	t.Helper()
	for _, r := range e.portal.requests() {
		if r.Method == method {
			var m map[string]any
			if err := json.Unmarshal([]byte(r.Body), &m); err != nil {
				t.Fatalf("%s body %q: %v", method, r.Body, err)
			}
			return m
		}
	}
	t.Fatalf("no %s in %+v", method, e.portal.requests())
	return nil
}

// MA-W1, MA-W19, MA-W30: each create sends the given fields plus the declared
// defaults to the family's path, with no provision or async query.
func TestFamilies_CreateRequestShape(t *testing.T) {
	for _, s := range familySpecs {
		for i, c := range s.creates {
			t.Run(fmt.Sprintf("%s-%d", s.name, i), func(t *testing.T) {
				s.flags = c.flags
				e := newFamilyEnv(t, s)
				r := e.call(s.name+"_create", map[string]any{"body": c.body})
				mustOK(t, r)
				if !strings.Contains(r.Text, "with id 101.") {
					t.Fatalf("text = %s", r.Text)
				}
				if got := sentBody(t, e, "POST"); !reflect.DeepEqual(got, jsonRoundTrip(c.want)) {
					t.Fatalf("POST body = %v, want %v", got, c.want)
				}
				// D16: a default sent into a flag-gated field names its flag.
				for _, flag := range c.flags {
					named := strings.Contains(r.Text, "Tenant feature flags that shaped the request: "+flag+".")
					if named != slices.Contains(c.named, flag) {
						t.Fatalf("flag %s named %v:\n%s", flag, named, r.Text)
					}
				}
				if c.flags == nil && strings.Contains(r.Text, "Tenant feature flags that shaped the request") {
					t.Fatalf("flags named with every flag off:\n%s", r.Text)
				}
				for _, req := range e.portal.requests() {
					if req.Method != "GET" && req.Path != e.portal.base() {
						t.Fatalf("write outside the family path: %+v", req)
					}
					if strings.Contains(req.Query, "provision") || strings.Contains(req.Query, "async") || strings.Contains(req.Query, "name=") {
						t.Fatalf("unexpected query: %+v", req)
					}
				}
			})
		}
	}
}

// MA-W19: schema defaults equal the backend and portal defaults (Segment.java:52, :136; createSegments/index.js:403-408).
func TestFamilies_SchemaDefaults(t *testing.T) {
	var res struct {
		Tools []struct {
			Name        string `json:"name"`
			InputSchema struct {
				Properties struct {
					Body struct {
						Properties map[string]json.RawMessage `json:"properties"`
					} `json:"body"`
				} `json:"properties"`
			} `json:"inputSchema"`
		} `json:"tools"`
	}
	e := newFamilyEnv(t, specNamed("segment"))
	json.Unmarshal(call(t, e.srv, "tools/list", nil), &res)
	for _, tool := range res.Tools {
		if tool.Name != "segment_create" {
			continue
		}
		props := tool.InputSchema.Properties.Body.Properties
		for field, want := range map[string]string{"asn": "65514", "overlappingRouteValidationEnabled": "true", "enableIpv6ToIpv4Translation": "false"} {
			var p struct {
				Default json.RawMessage `json:"default"`
			}
			json.Unmarshal(props[field], &p)
			if string(p.Default) != want {
				t.Errorf("%s default = %s, want %s", field, p.Default, want)
			}
		}
		var std struct {
			Properties map[string]struct {
				Default json.RawMessage `json:"default"`
			} `json:"properties"`
		}
		json.Unmarshal(props["serviceTrafficDistribution"], &std)
		if string(std.Properties["algorithm"].Default) != `"HASHING"` {
			t.Errorf("serviceTrafficDistribution.algorithm default = %s", std.Properties["algorithm"].Default)
		}
		return
	}
	t.Fatal("segment_create not listed")
}

// MA-W2: invalid arguments are refused with zero requests, per family.
func TestFamilies_InvalidArgumentsSendNothing(t *testing.T) {
	for _, s := range familySpecs {
		cases := map[string]toolCase{
			"trailing space":   {"_create", map[string]any{"body": withFields(s.createBody("a"), map[string]any{"description": "x "})}, "description has a leading or trailing space"},
			"read-only id":     {"_create", map[string]any{"body": withFields(s.createBody("a"), map[string]any{"id": "5"})}, "id is read-only"},
			"unknown":          {"_create", map[string]any{"body": withFields(s.createBody("a"), map[string]any{"colour": "red"})}, "unknown field colour"},
			"patch wrong type": {"_update", map[string]any{"id": "7", "patch": map[string]any{"description": 5}}, "description must be a string"},
		}
		if s.nameKey == "" {
			cases["missing name"] = toolCase{"_create", map[string]any{"body": map[string]any{"description": "x"}}, "name is required"}
			cases["leading space"] = toolCase{"_create", map[string]any{"body": s.createBody(" a")}, "name has a leading or trailing space"}
			cases["charset"] = toolCase{"_create", map[string]any{"body": s.createBody("a/b")}, "outside the name character set"}
			cases["maxlen"] = toolCase{"_create", map[string]any{"body": s.createBody(strings.Repeat("a", s.maxName()+1))}, fmt.Sprintf("at most %d allowed", s.maxName())}
			cases["patch remove req"] = toolCase{"_update", map[string]any{"id": "7", "patch": map[string]any{"name": nil}}, "name cannot be removed"}
		}
		for name, c := range s.invalid {
			if c.tool == "" {
				delete(cases, name)
				continue
			}
			cases[name] = c
		}
		for name, c := range cases {
			t.Run(s.name+"/"+name, func(t *testing.T) {
				e := newFamilyEnv(t, s)
				mustRefuse(t, e.call(s.name+c.tool, c.args), "invalid_arguments", c.want)
				if n := len(e.portal.requests()); n != 0 {
					t.Fatalf("%d requests sent", n)
				}
			})
		}
	}
}

// MA-W4, MA-W38, MA-W30: a one-field patch resends the stored body with the change, read-only fields left out.
func TestFamilies_UpdateMerge(t *testing.T) {
	for _, s := range familySpecs {
		t.Run(s.name, func(t *testing.T) {
			e := newFamilyEnv(t, s)
			s.put(e, 7, s.stored())
			r := e.call(s.name+"_update", map[string]any{"id": "7", "patch": map[string]any{"description": ""}})
			mustOK(t, r)
			gets, want := 0, max(s.updateGets, 1)
			for _, req := range e.portal.requests() {
				if req.Method == "GET" && strings.HasPrefix(req.Path, e.portal.base()) {
					gets++
				}
			}
			if gets != want || e.portal.count("PUT") != 1 {
				t.Fatalf("requests = %+v", e.portal.requests())
			}
			body := sentBody(t, e, "PUT")
			if body["description"] != "" || body[s.nameField()] == nil {
				t.Fatalf("PUT body = %v", body)
			}
			for _, ro := range s.readOnly {
				if _, ok := body[ro]; ok {
					t.Fatalf("read-only %s resent: %v", ro, body)
				}
			}
			if !strings.Contains(r.Text, "Previous object:") || !strings.Contains(r.Text, `"description": "seven"`) {
				t.Fatalf("no previous object: %s", r.Text)
			}
			if s.name == "segment" {
				if body["asn"] != float64(65514) || body["serviceTrafficDistribution"] == nil || body["ipBlocks"] == nil {
					t.Fatalf("stored segment fields not resent: %v", body)
				}
			}
		})
	}
}

// MA-W4: a stored segment asn outside today's ranges does not block a description patch.
func TestSegment_LegacyASNUpdate(t *testing.T) {
	e := newFamilyEnv(t, specNamed("segment"))
	o := specNamed("segment").stored()
	o["asn"] = 64512
	e.portal.put(7, o)
	mustOK(t, e.call("segment_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}))
	if body := sentBody(t, e, "PUT"); body["asn"] != float64(64512) {
		t.Fatalf("PUT body = %v", body)
	}
}

// MA-W31: reservePublicIPsForUserAndSiteConnectivity is resent unchanged before its expiry (the
// after-expiry drop runs in client-go, whose clock this package cannot move).
func TestSegment_ReservePublicIPsResent(t *testing.T) {
	e := newFamilyEnv(t, specNamed("segment"))
	o := specNamed("segment").stored()
	o["reservePublicIPsForUserAndSiteConnectivity"] = true
	e.portal.put(7, o)
	mustOK(t, e.call("segment_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}))
	if body := sentBody(t, e, "PUT"); body["reservePublicIPsForUserAndSiteConnectivity"] != true {
		t.Fatalf("PUT body = %v", body)
	}
}

// MA-W5: a FAILED segment is refused with no PUT; a segment on a FAILED network is not checked.
// Groups carry no state in TPS (model/Group.java), so a group has no state guard.
func TestFamilies_StateGuard(t *testing.T) {
	e := newFamilyEnv(t, specNamed("segment"))
	o := specNamed("segment").stored()
	o["state"] = "FAILED"
	e.portal.put(7, o)
	mustRefuse(t, e.call("segment_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}), "state_refused",
		"segment id 7 cannot change while its state is FAILED")
	if e.portal.count("PUT") != 0 {
		t.Fatal("PUT sent")
	}
	o = specNamed("segment").stored()
	o["state"] = "SUCCESS"
	e.portal.put(8, o)
	mustOK(t, e.call("segment_update", map[string]any{"id": "8", "patch": map[string]any{"description": "x"}}))

	g := newFamilyEnv(t, specNamed("group"))
	g.portal.put(7, specNamed("group").stored())
	mustOK(t, g.call("group_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}))
	if ak.GroupDescriptor().StateGuard != 0 {
		t.Fatalf("group descriptor has a state guard: %v", ak.GroupDescriptor().StateGuard)
	}
}

// MA-W16: a row marked for deletion is refused before any write. A connector
// family finds such a row through its GET /connectors row; its own tests cover it.
func TestFamilies_DocState(t *testing.T) {
	for _, s := range familySpecs {
		if s.connector != "" {
			continue
		}
		t.Run(s.name, func(t *testing.T) {
			e := newFamilyEnv(t, s)
			o := s.stored()
			o["docState"] = "MARKED_FOR_DELETION"
			s.put(e, 7, o)
			mustRefuse(t, e.call(s.name+"_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}), "state_refused", "deletion pending provisioning")
			if e.portal.count("PUT") != 0 {
				t.Fatal("PUT sent")
			}
		})
	}
}

// MA-W10, MA-W39: each family's not-found signature reads as not found; a bare 404 does not.
func TestFamilies_NotFound(t *testing.T) {
	for _, s := range familySpecs {
		t.Run(s.name, func(t *testing.T) {
			e := newFamilyEnv(t, s)
			noun := s.nounText()
			mustRefuse(t, e.call(s.name+"_update", map[string]any{"id": "404", "patch": map[string]any{"description": "x"}}), "not_found", noun+" id 404 not found.")
			mustRefuse(t, e.call(s.name+"_delete", map[string]any{"id": "404"}), "not_found", noun+" id 404 not found.")
			if e.portal.count("PUT")+e.portal.count("DELETE") != 0 {
				t.Fatal("write sent")
			}
			for _, body := range []string{``, `<html>Not Found</html>`, `{"message":"Not Found"}`} {
				e.portal.notFound = &fault{status: 404, body: body}
				r := e.call(s.name+"_update", map[string]any{"id": "404", "patch": map[string]any{"description": "x"}})
				if !r.IsError || r.Code == "not_found" {
					t.Fatalf("bare 404 %q read as %q: %s", body, r.Code, r.Text)
				}
			}
		})
	}
}

// A segment or group gone between the pre-GET and the DELETE: the DELETE's not-found answer reads as already deleted.
func TestFamilies_DeleteAlreadyGone(t *testing.T) {
	for name, body := range map[string]string{
		"segment": `{"code":"AK000002","message":"Segment with id: '7' was not found"}`,
		"group":   `{"code":"AK000002","message":"Group with id: '7' was not found"}`,
	} {
		t.Run(name, func(t *testing.T) {
			s := specNamed(name)
			e := newFamilyEnv(t, s)
			e.portal.put(7, s.stored())
			code := previewCode(t, e.call(name+"_delete", map[string]any{"id": "7"}))
			e.portal.fail("DELETE", e.portal.base()+"/7", fault{status: 400, body: body})
			r := e.call(name+"_delete", map[string]any{"id": "7", "confirmCode": code})
			mustOK(t, r)
			if !strings.Contains(r.Text, "(id 7) was already deleted: the backend answered the DELETE with not found (status 400).") {
				t.Fatalf("text = %s", r.Text)
			}
			if e.portal.count("DELETE") != 1 {
				t.Fatal("want one DELETE")
			}
		})
	}
}

// MA-W12, MA-W32: in-use answers keep the backend text and add the staged-referrer note; one DELETE each.
func TestFamilies_DeleteInUse(t *testing.T) {
	for _, s := range familySpecs {
		for i, c := range s.inUse {
			t.Run(fmt.Sprintf("%s-%d", s.name, i), func(t *testing.T) {
				e := newFamilyEnv(t, s)
				s.put(e, 7, s.stored())
				code := previewCode(t, e.call(s.name+"_delete", map[string]any{"id": "7"}))
				e.portal.fail("DELETE", e.portal.base()+"/7", c.f)
				r := e.call(s.name+"_delete", map[string]any{"id": "7", "confirmCode": code})
				mustRefuse(t, r, "in_use", c.want)
				mustRefuse(t, r, "in_use", "A referrer deleted without a provision keeps the reference until the next provision.")
				if e.portal.count("DELETE") != 1 {
					t.Fatal("want one DELETE, no retry")
				}
			})
		}
	}
}

// MA-W11: the duplicate guard pages the family list (hit on page 2) and sends
// no POST. A connector family's guard reads GET /connectors and the groups
// instead; its own tests cover it.
func TestFamilies_DuplicateGuard(t *testing.T) {
	for _, s := range familySpecs {
		if s.connector != "" {
			continue
		}
		t.Run(s.name, func(t *testing.T) {
			e := newFamilyEnv(t, s)
			key := s.nameField()
			for i := 1; i <= 130; i++ {
				o := s.stored()
				o[key] = fmt.Sprintf("n%03d", i)
				s.put(e, i, o)
			}
			want := fmt.Sprintf(`%s named "n120" already exists (id 120).`, handlers.Article(s.nounText()))
			if key != "name" {
				want = fmt.Sprintf(`%s with %s "n120" already exists (id 120).`, handlers.Article(s.nounText()), key)
			}
			mustRefuse(t, e.call(s.name+"_create", map[string]any{"body": s.createBody("n120")}), "duplicate", want)
			if e.portal.count("POST") != 0 {
				t.Fatal("POST sent")
			}
			var queries []string
			for _, r := range e.portal.requests() {
				if r.Path == e.portal.base() {
					queries = append(queries, r.Query)
				}
				if strings.Contains(r.Query, "name=") {
					t.Fatalf("guard sent ?name=: %+v", r)
				}
			}
			if len(queries) != 2 || queries[1] != "limit=100&offset=100&paginated=true" {
				t.Fatalf("guard reads = %v", queries)
			}
		})
	}
}

// MA-W11: a group name equal to a connector name is refused; a connector row whose single GET is 404 is skipped.
func TestGroup_ConnectorNameSpace(t *testing.T) {
	e := newFamilyEnv(t, specNamed("group"))
	e.portal.connectors = []map[string]any{
		{"id": 31, "name": "edge", "type": "IPSEC"},
		{"id": 32, "name": "gone", "type": "AWS_VPC", "gone": true},
	}
	mustRefuse(t, e.call("group_create", map[string]any{"body": map[string]any{"name": "edge"}}), "duplicate",
		`an ipsec connector named "edge" already exists (id 31); connector names are unique across every connector type and group.`)
	if e.portal.count("POST") != 0 {
		t.Fatal("POST sent")
	}
	mustOK(t, e.call("group_create", map[string]any{"body": map[string]any{"name": "gone"}}))
	if e.portal.count("POST") != 1 {
		t.Fatal("want one POST")
	}
}

// MA-W23: a field the create echo drops is named.
func TestFamilies_EchoDiff(t *testing.T) {
	e := newFamilyEnv(t, specNamed("segment"))
	e.portal.fail("POST", e.portal.base(), fault{status: 201, body: `{"id":101,"name":"s1","asn":65514}`})
	r := e.call("segment_create", map[string]any{"body": map[string]any{"name": "s1", "ipBlocks": map[string]any{"values": []any{"10.2.0.0/24"}}}})
	mustOK(t, r)
	if !strings.Contains(r.Text, "ipBlocks dropped") {
		t.Fatalf("text = %s", r.Text)
	}
}

// MA-W51: the delete preview lists implicit objects, deployed state and the provision effect per family.
func TestFamilies_DeletePreview(t *testing.T) {
	for _, s := range familySpecs {
		for i, c := range s.previews {
			t.Run(fmt.Sprintf("%s-%d", s.name, i), func(t *testing.T) {
				e := newFamilyEnv(t, s)
				o := s.stored()
				if c.state != "" {
					o["state"] = c.state
				}
				s.put(e, 7, o)
				r := e.call(s.name+"_delete", map[string]any{"id": "7"})
				previewCode(t, r)
				implicit := false
				for _, w := range append(c.want, "expires: 2026-11-17T12:05:00Z") {
					implicit = implicit || strings.HasPrefix(w, "Also deleted with it")
					if !strings.Contains(r.Text, w) {
						t.Errorf("preview lacks %q:\n%s", w, r.Text)
					}
				}
				if !implicit && strings.Contains(r.Text, "Also deleted with it") {
					t.Errorf("unexpected implicit objects:\n%s", r.Text)
				}
				if e.portal.count("DELETE") != 0 {
					t.Fatal("DELETE sent")
				}
			})
		}
	}
}

// MA-W10, MA-W52 to MA-W55 per family: a valid code sends one DELETE; a code for another family or id is refused.
func TestFamilies_DeleteWithCode(t *testing.T) {
	for _, s := range familySpecs {
		t.Run(s.name, func(t *testing.T) {
			e := newFamilyEnv(t, s)
			s.put(e, 7, s.stored())
			s.put(e, 8, s.stored())
			code := previewCode(t, e.call(s.name+"_delete", map[string]any{"id": "7"}))
			mustRefuse(t, e.call(s.name+"_delete", map[string]any{"id": "8", "confirmCode": code}), handlers.RefusalConfirmMismatch, "confirmation code does not match this object")
			o := e.portal.object(7)
			o["description"] = "changed"
			mustRefuse(t, e.call(s.name+"_delete", map[string]any{"id": "7", "confirmCode": code}), handlers.RefusalConfirmStale, "the object changed since the preview")
			if e.portal.count("DELETE") != 0 {
				t.Fatal("DELETE sent on a refused code")
			}
			code = previewCode(t, e.call(s.name+"_delete", map[string]any{"id": "7"}))
			r := e.call(s.name+"_delete", map[string]any{"id": "7", "confirmCode": code})
			mustOK(t, r)
			if !strings.Contains(r.Text, "(id 7).") || e.portal.count("DELETE") != 1 || e.portal.object(7) != nil {
				t.Fatalf("text = %s", r.Text)
			}
			for _, req := range e.portal.requests() {
				if req.Method == "DELETE" && req.Query != "" {
					t.Fatalf("DELETE query = %q", req.Query)
				}
			}
		})
	}
}

// MA-W12, MA-W49: the billing tag limit is a limit on every verb, never not found; create and the
// TPS create limit map the same way.
func TestFamilies_Limit(t *testing.T) {
	limit := fault{status: 404, body: `{"code":"AK000003","message":"Cannot create more than 10 billing tags"}`}
	e := newFamilyEnv(t, specNamed("billing_tag"))
	e.portal.put(7, specNamed("billing_tag").stored())
	e.portal.fail("PUT", "/api/tags/7", limit)
	mustRefuse(t, e.call("billing_tag_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}), "limit_exceeded", "Cannot create more than 10 billing tags")
	e.portal.fail("POST", "/api/tags", limit)
	mustRefuse(t, e.call("billing_tag_create", map[string]any{"body": map[string]any{"name": "t8"}}), "limit_exceeded", "Cannot create more than 10 billing tags")

	s := newFamilyEnv(t, specNamed("segment"))
	s.portal.fail("POST", s.portal.base(), fault{status: 400, body: `{"code":"AK000002","message":"Cannot create a new segment, limit = 20"}`})
	mustRefuse(t, s.call("segment_create", map[string]any{"body": map[string]any{"name": "s9"}}), "limit_exceeded", "Cannot create a new segment, limit = 20")
	if s.portal.count("POST") != 1 {
		t.Fatal("want one POST")
	}
}

// Billing tags need BILLING-VIEW and BILLING-TAGS; with either off every verb, the delete preview included, is
// refused before any write and the preview issues no code.
func TestBillingTag_NotEnabled(t *testing.T) {
	for _, off := range []string{"BILLING-VIEW", "BILLING-TAGS"} {
		s := specNamed("billing_tag")
		s.flags = slices.DeleteFunc(slices.Clone(s.flags), func(f string) bool { return f == off })
		e := newFamilyEnv(t, s)
		e.portal.put(7, s.stored())
		for tool, args := range map[string]map[string]any{
			"billing_tag_create": {"body": map[string]any{"name": "t1"}},
			"billing_tag_update": {"id": "7", "patch": map[string]any{"description": "x"}},
			"billing_tag_delete": {"id": "7"},
		} {
			r := e.call(tool, args)
			mustRefuse(t, r, "not_enabled", "Billing tag is not enabled for this tenant (tenant feature "+off+" is off); nothing was sent.")
			if strings.Contains(r.Text, "confirmCode") {
				t.Fatalf("%s: code issued: %s", tool, r.Text)
			}
		}
		if e.portal.count("POST")+e.portal.count("PUT")+e.portal.count("DELETE") != 0 {
			t.Fatalf("%s off: write sent", off)
		}
	}
	s := specNamed("billing_tag")
	s.flags = []string{}
	e := newFamilyEnv(t, s)
	e.portal.put(7, s.stored())
	// A code signed for this object while the flag was on is still refused.
	raw, _ := json.Marshal(e.portal.object(7))
	state, err := confirm.StateHash(raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	code, _ := e.signer().Issue(e.portal.client().URI+"#7", "billing_tag", "7", state)
	mustRefuse(t, e.call("billing_tag_delete", map[string]any{"id": "7", "confirmCode": code}), "not_enabled", "BILLING-VIEW is off")
	for _, r := range e.portal.requests() {
		if r.Method != "GET" {
			t.Fatalf("write sent: %+v", r)
		}
	}
	if e.portal.count("DELETE") != 0 {
		t.Fatal("DELETE sent")
	}
}

// A group whose type is not EXPLICIT (a connector group, a <segment>-ALL-Group)
// is refused by the update and by the delete preview before any write: the
// preview issues no confirmation code, and a code signed for it is refused too.
func TestGroup_ImplicitRefused(t *testing.T) {
	s := specNamed("group")
	e := newFamilyEnv(t, s)
	g := s.stored()
	g["type"] = "IMPLICIT"
	e.portal.put(7, g)
	const want = "Group id 7 has type IMPLICIT; only groups with type EXPLICIT can be changed or deleted. Nothing was sent."
	for tool, args := range map[string]map[string]any{
		"group_update": {"id": "7", "patch": map[string]any{"description": "x"}},
		"group_delete": {"id": "7"},
	} {
		r := e.call(tool, args)
		mustRefuse(t, r, "family_mismatch", want)
		if strings.Contains(r.Text, "confirmCode") {
			t.Fatalf("%s: code issued: %s", tool, r.Text)
		}
	}
	raw, _ := json.Marshal(e.portal.object(7))
	state, err := confirm.StateHash(raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	code, _ := e.signer().Issue(e.portal.client().URI+"#7", "group", "7", state)
	mustRefuse(t, e.call("group_delete", map[string]any{"id": "7", "confirmCode": code}), "family_mismatch", want)
	if n := e.portal.count("PUT") + e.portal.count("DELETE"); n != 0 {
		t.Fatalf("%d writes sent", n)
	}
}

// With IFA-IPV6-SUPPORT off, a segment update sends a stored source pool and
// translation as [] and false, and the output names the fields and the flag.
// reservePublicIPsForUserAndSiteConnectivityForCXPs needs an Akamai flag as
// well as the per-CXP flag; without one it is left out.
func TestSegment_FlagShaping(t *testing.T) {
	s := specNamed("segment")
	e := newFamilyEnv(t, s)
	stored := s.stored()
	stored["enableIpv6ToIpv4Translation"] = true
	stored["srcIpv4PoolList"] = []any{map[string]any{"startIp": "100.64.0.1", "endIp": "100.64.0.10"}}
	e.portal.put(7, stored)
	r := e.call("segment_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}})
	if r.IsError {
		t.Fatalf("update: %s", r.Text)
	}
	put := sentBody(t, e, "PUT")
	if fmt.Sprint(put["srcIpv4PoolList"]) != "[]" || put["enableIpv6ToIpv4Translation"] != false {
		t.Errorf("PUT body %v", put)
	}
	for _, line := range []string{
		"Tenant feature flags that shaped the request: IFA-IPV6-SUPPORT.",
		"Stored values sent empty or false, as the portal does, because the backend refuses them under the current tenant features: " +
			"srcIpv4PoolList (tenant feature IFA-IPV6-SUPPORT is off), enableIpv6ToIpv4Translation (tenant feature IFA-IPV6-SUPPORT is off).",
	} {
		if !strings.Contains(r.Text, line) {
			t.Errorf("output lacks %q:\n%s", line, r.Text)
		}
	}

	s.flags = []string{"ALLOW-USERS-AND-SITES-CONNECTIVITY-PER-CXP"}
	e = newFamilyEnv(t, s)
	r = e.call("segment_create", map[string]any{"body": map[string]any{"name": "s1", "reservePublicIPsForUserAndSiteConnectivityForCXPs": []any{"US-WEST"}}})
	if r.IsError {
		t.Fatalf("create: %s", r.Text)
	}
	if _, sent := sentBody(t, e, "POST")["reservePublicIPsForUserAndSiteConnectivityForCXPs"]; sent ||
		!strings.Contains(r.Text, "Left out because their tenant feature is off: reservePublicIPsForUserAndSiteConnectivityForCXPs.") {
		t.Errorf("per-CXP flag alone: %s", r.Text)
	}
}
