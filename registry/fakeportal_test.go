package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/hashicorp/go-retryablehttp"
)

// faultKind is one of the fault kinds of the v3 mock (V3 internal/mock/faults.go:17-50).
type faultKind int

const (
	faultStatus  faultKind = iota // answer status, headers and body
	faultReset                    // close the connection after reading the request
	faultHang                     // hold the request until the client gives up
	faultBadJSON                  // answer 200 (or status) with malformed JSON
	faultCommit                   // apply the request, then close the connection
	faultRewrite                  // apply the request, then answer status and body instead of the result
)

type fault struct {
	kind   faultKind
	status int
	body   string
	header map[string]string
}

type journalEntry struct {
	Method, Path, Query, Body string
}

// writePortal is a stateful fake of one TPS collection at
// /api/tenantnetworks/7/<collection>: paged list, POST, GET, PUT and DELETE,
// with per-request faults and a journal of every request.
type writePortal struct {
	t          *testing.T
	collection string
	mu         sync.Mutex
	objects    map[int]map[string]any
	nextID     int
	journal    []journalEntry
	faults     map[string][]fault // "METHOD /path" -> faults served in order, one per request
	srv        *httptest.Server
	before     func() // runs before each request is answered
	// flags is the answer of GET /api/tenantenabledfeatures; flagStatus,
	// when set, fails that read. Flag reads are counted in flagReads, not
	// in the journal.
	flags      []string
	flagStatus int
	flagReads  int
	// validation answers GET .../validation-states/{id}; unset ids are IN_PROGRESS.
	validation map[string]string
	// path, when set, replaces /api/tenantnetworks/7/<collection>, as for
	// tenant-level collections.
	path string
	// notFound, when set, answers a missing id instead of 404 AK000003.
	notFound *fault
	// connectors answers GET /api/tenantnetworks/7/connectors (every family,
	// paging ignored) and single GETs of its rows under .../connectors/{id}.
	connectors []map[string]any
	// networkState answers GET /api/tenantnetworks/7; empty means SUCCESS,
	// since the backend always returns a state.
	networkState string
	// groups answers the paged GET /api/tenantnetworks/7/groups when groups
	// are not the portal's own collection.
	groups []map[string]any
	// serviceType, when set, makes the objects services of that type: GET
	// /api/tenantnetworks/7/services lists them after otherServices, GET
	// .../services/{id} answers an object's row, and a plain GET of a row
	// marked for deletion answers not found. As in TPS, rows marked for
	// deletion are listed only with includeMarkedForDeletion=true, and
	// without staleOk=false the list leaves out the rows in fresh.
	// serviceRows overrides fields of an object's row, so a row can carry a
	// state its object body does not.
	serviceType   string
	otherServices []map[string]any
	serviceRows   map[int]map[string]any
	fresh         map[int]bool
}

func newWritePortal(t *testing.T, collection string) *writePortal {
	t.Helper()
	p := &writePortal{t: t, collection: collection, objects: map[int]map[string]any{}, nextID: 100, faults: map[string][]fault{}}
	p.srv = httptest.NewServer(http.HandlerFunc(p.serve))
	t.Cleanup(p.srv.Close)
	return p
}

func (p *writePortal) base() string {
	if p.path != "" {
		return p.path
	}
	return "/api/tenantnetworks/7/" + p.collection
}

// client returns an AlkiraClient for the fake; host distinguishes tenants.
func (p *writePortal) client() *ak.AlkiraClient {
	hc := retryablehttp.NewClient()
	hc.RetryMax = 0
	hc.Logger = nil
	return &ak.AlkiraClient{Client: hc, URI: p.srv.URL + "/api", TenantNetworkId: "7", Authorization: "api-key test"}
}

// put stores an object under id.
func (p *writePortal) put(id int, obj map[string]any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	obj["id"] = id
	p.objects[id] = obj
}

func (p *writePortal) object(id int) map[string]any {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.objects[id]
}

// fail queues f for the next request matching method and path.
func (p *writePortal) fail(method, path string, f fault) {
	p.mu.Lock()
	defer p.mu.Unlock()
	key := method + " " + path
	p.faults[key] = append(p.faults[key], f)
}

func (p *writePortal) requests() []journalEntry {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]journalEntry(nil), p.journal...)
}

// count returns the number of requests with method.
func (p *writePortal) count(method string) int {
	n := 0
	for _, e := range p.requests() {
		if e.Method == method {
			n++
		}
	}
	return n
}

func (p *writePortal) serve(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	p.mu.Lock()
	if r.URL.Path == "/api/tenantenabledfeatures" {
		p.flagReads++
		flags, status := p.flags, p.flagStatus
		p.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if status != 0 {
			w.WriteHeader(status)
			fmt.Fprint(w, `{"code":"AK000500","message":"unavailable"}`)
			return
		}
		if flags == nil {
			flags = []string{}
		}
		json.NewEncoder(w).Encode(flags)
		return
	}
	p.journal = append(p.journal, journalEntry{r.Method, r.URL.Path, r.URL.RawQuery, string(body)})
	key := r.Method + " " + r.URL.Path
	var f *fault
	if q := p.faults[key]; len(q) > 0 {
		f = &q[0]
		p.faults[key] = q[1:]
	}
	before := p.before
	p.mu.Unlock()

	if before != nil {
		before()
	}
	if f != nil {
		switch f.kind {
		case faultReset:
			hijackClose(w)
			return
		case faultHang:
			<-r.Context().Done()
			return
		case faultBadJSON:
			status := f.status
			if status == 0 {
				status = 200
			}
			w.WriteHeader(status)
			fmt.Fprint(w, `{"id": 1,`)
			return
		case faultStatus:
			for k, v := range f.header {
				w.Header().Set(k, v)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(f.status)
			fmt.Fprint(w, f.body)
			return
		case faultCommit:
			p.handle(httptest.NewRecorder(), r, body)
			hijackClose(w)
			return
		case faultRewrite:
			p.handle(httptest.NewRecorder(), r, body)
			for k, v := range f.header {
				w.Header().Set(k, v)
			}
			w.WriteHeader(f.status)
			fmt.Fprint(w, f.body)
			return
		}
	}
	p.handle(w, r, body)
}

func hijackClose(w http.ResponseWriter) {
	conn, _, err := w.(http.Hijacker).Hijack()
	if err == nil {
		conn.Close()
	}
}

func (p *writePortal) handle(w http.ResponseWriter, r *http.Request, body []byte) {
	p.mu.Lock()
	defer p.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	notFound := func() {
		if p.notFound != nil {
			w.WriteHeader(p.notFound.status)
			fmt.Fprint(w, p.notFound.body)
			return
		}
		w.WriteHeader(404)
		fmt.Fprint(w, `{"code":"AK000003","message":"Entity not found","id":"ALK-404"}`)
	}
	if rest, ok := strings.CutPrefix(r.URL.Path, "/api/tenantnetworks/7/connectors"); ok && r.Method == "GET" {
		if rest == "" {
			rows := p.connectors
			if rows == nil {
				rows = []map[string]any{}
			}
			// TPS answers ?name= from the database: only the live row of that name.
			if name := r.URL.Query().Get("name"); name != "" {
				rows = []map[string]any{}
				for _, c := range p.connectors {
					if c["name"] == name && c["gone"] == nil {
						rows = append(rows, c)
					}
				}
			}
			json.NewEncoder(w).Encode(rows)
			return
		}
		for _, c := range p.connectors {
			if "/"+fmt.Sprint(c["id"]) == rest && c["gone"] == nil {
				json.NewEncoder(w).Encode(c)
				return
			}
		}
		w.WriteHeader(404)
		fmt.Fprint(w, `{"code":"AK000003","message":"Entity not found","id":"ALK-404"}`)
		return
	}
	if rest, ok := strings.CutPrefix(r.URL.Path, "/api/tenantnetworks/7/services"); ok && r.Method == "GET" && p.serviceType != "" {
		q := r.URL.Query()
		marked, stale := q.Get("includeMarkedForDeletion") == "true", q.Get("staleOk") != "false"
		rows := append([]map[string]any{}, p.otherServices...)
		ids := make([]int, 0, len(p.objects))
		for id := range p.objects {
			ids = append(ids, id)
		}
		sort.Ints(ids)
		for _, id := range ids {
			o := p.objects[id]
			if o["docState"] == "MARKED_FOR_DELETION" && !marked || rest == "" && stale && p.fresh[id] {
				continue
			}
			prev := o["previousState"]
			if prev == nil {
				prev = o["state"]
			}
			row := map[string]any{"id": id, "name": o["name"], "type": p.serviceType, "state": o["state"], "previousState": prev}
			for k, v := range p.serviceRows[id] {
				row[k] = v
			}
			rows = append(rows, row)
		}
		if rest == "" {
			json.NewEncoder(w).Encode(rows)
			return
		}
		for _, row := range rows {
			if "/"+fmt.Sprint(row["id"]) == rest {
				json.NewEncoder(w).Encode(row)
				return
			}
		}
		notFound()
		return
	}
	if r.Method == "GET" && r.URL.Path == "/api/tenantnetworks/7" {
		state := p.networkState
		if state == "" {
			state = "SUCCESS"
		}
		fmt.Fprintf(w, `{"id":7,"state":%q}`, state)
		return
	}
	if r.Method == "GET" && r.URL.Path == "/api/tenantnetworks/7/groups" && p.base() != r.URL.Path {
		rows := p.groups
		if rows == nil {
			rows = []map[string]any{}
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		page := []map[string]any{}
		for i := offset; i < len(rows) && (limit == 0 || i < offset+limit); i++ {
			page = append(page, rows[i])
		}
		json.NewEncoder(w).Encode(map[string]any{"data": page, "pagination": map[string]any{"hits": len(rows)}})
		return
	}
	if vid, ok := strings.CutPrefix(r.URL.Path, "/api/tenantnetworks/7/validation-states/"); ok {
		body, ok := p.validation[vid]
		if !ok {
			body = fmt.Sprintf(`{"id":1,"networkEntityId":%q,"state":"IN_PROGRESS"}`, vid)
		}
		fmt.Fprint(w, body)
		return
	}
	// With ?async=true a create or update answers 202 with the validation state id.
	async := func(id int, status int) int {
		if r.URL.Query().Get("async") != "true" {
			return status
		}
		w.Header().Set("x-ak-validation-state-id", fmt.Sprintf("WIDGET:DEFAULT:%d", id))
		return 202
	}
	if r.URL.Path == p.base() {
		switch r.Method {
		case "GET":
			ids := make([]int, 0, len(p.objects))
			for id, o := range p.objects {
				// Live policy-service and AS-path list rows carry docState ADDED.
				if ds := o["docState"]; ds == nil || ds == "ADDED" {
					ids = append(ids, id)
				}
			}
			sort.Ints(ids)
			offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
			limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
			if limit == 0 {
				limit = len(ids)
			}
			page := []map[string]any{}
			for i := offset; i < len(ids) && i < offset+limit; i++ {
				page = append(page, p.objects[ids[i]])
			}
			json.NewEncoder(w).Encode(map[string]any{"data": page, "pagination": map[string]any{"hits": len(ids)}})
		case "POST":
			var obj map[string]any
			json.Unmarshal(body, &obj)
			p.nextID++
			obj["id"], obj["state"] = p.nextID, "PENDING"
			p.objects[p.nextID] = obj
			w.WriteHeader(async(p.nextID, 201))
			json.NewEncoder(w).Encode(obj)
		default:
			w.WriteHeader(405)
		}
		return
	}
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, p.base()+"/"))
	obj, ok := p.objects[id]
	if err != nil || !ok {
		notFound()
		return
	}
	switch r.Method {
	case "GET":
		if p.serviceType != "" && obj["docState"] == "MARKED_FOR_DELETION" && r.URL.Query().Get("includeMarkedForDeletion") != "true" {
			notFound()
			return
		}
		json.NewEncoder(w).Encode(obj)
	case "PUT":
		var next map[string]any
		json.Unmarshal(body, &next)
		next["id"] = id
		if s, ok := obj["state"]; ok {
			next["state"] = s
		}
		p.objects[id] = next
		w.WriteHeader(async(id, 200))
		json.NewEncoder(w).Encode(next)
	case "DELETE":
		// TPS ProvisionedState.isProvisioned: neither state nor previousState is PENDING.
		s, _ := obj["state"].(string)
		prev, _ := obj["previousState"].(string)
		if s != "" && s != "PENDING" && prev != "PENDING" {
			obj["docState"] = "MARKED_FOR_DELETION"
		} else {
			delete(p.objects, id)
		}
		w.WriteHeader(200)
	}
}
