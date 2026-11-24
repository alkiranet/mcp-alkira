package registry

import (
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/alkiranet/mcp-alkira/handlers"
	"github.com/alkiranet/mcp-alkira/writes"
)

func init() {
	familySpecs = append(familySpecs, serviceSpecs...)
	for name, k := range map[string]string{"service_pan": "K40", "service_check_point": "K41", "service_cisco_ftdv": "K42", "service_fortinet": "K43"} {
		maps.Copy(scenarioRows, map[string]map[string]string{name: {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode TestServices_InstancesByID; " + k + ".1-.6",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go)",
			scImportNonexistent: "TestFamilies_NotFound",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03); TestFamilies_InvalidArgumentsSendNothing",
			scPreflightBlocked:  "TestServices_StateGuard; MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
			scSoftDeleteOff:     "TestFamilies_DocState TestFamilies_DeletePreview",
		}})
	}
}

// serviceNameCases: firewall service names allow letters, digits and hyphens
// only, and the backend accepts a description with a trailing space.
func serviceNameCases(create func(string) map[string]any) map[string]toolCase {
	return map[string]toolCase{
		"charset":        {"_create", map[string]any{"body": create("a/b")}, "outside the hyphen character set"},
		"leading space":  {"_create", map[string]any{"body": create(" a")}, "outside the hyphen character set"},
		"trailing space": {},
	}
}

func withCases(a, b map[string]toolCase) map[string]toolCase {
	maps.Copy(a, b)
	return a
}

func serviceInUse() []inUseCase {
	return []inUseCase{
		{fault{status: 400, body: `{"code":"AK000002","message":"service fw-7 is in use by rule r1"}`}, "service fw-7 is in use by rule r1"},
		{fault{status: 409, body: `{"code":"AK000001","message":"SERVICE with id 7 is still referred externally"}`}, "is still referred externally"},
	}
}

func servicePreviews(noun, creds string) []previewCase {
	also := "Also deleted with it: the credentials it references (" + creds + ")."
	return []previewCase{
		{"PENDING", []string{`Delete preview for ` + noun + ` "fw-7" (id 7). Nothing was deleted.`, "Deployed: no, never deployed (state PENDING).",
			also, "The delete now: removes it at once.", "The next provision: nothing to tear down."}},
		{"SUCCESS", []string{"Deployed: yes (state SUCCESS).", "Also deleted with it: the credentials it references (" + creds +
			"), once a provision removes it; instances not yet deployed lose theirs at once.", "The next provision: tears down the deployed object."}},
	}
}

var serviceNotFound = fault{status: 404, body: `{"code":"AK000003","message":"service entity with id: 404 is not present in the data store"}`}

func panCreate(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "USEAST-AZURE-1", "segments": []any{1207}, "managementSegment": 1207, "licenseType": "PAY_AS_YOU_GO",
		"version": "10.2.10-h9", "credentialId": "c-1", "maxInstanceCount": 1, "instances": []any{map[string]any{"name": "i1", "credentialId": "c-i1"}}}
}

func checkpointCreate(name string) map[string]any {
	return map[string]any{"name": name, "credentialId": "c-1", "cxp": "USEAST-AZURE-1", "licenseType": "PAY_AS_YOU_GO", "version": "R81.20",
		"managementServer": map[string]any{"configurationMode": "MANUAL", "type": "SMS", "reachability": "PUBLIC", "ips": []any{"20.1.1.1"}, "globalCidrListId": 31},
		"segments":         []any{"seg-a"}, "segmentOptions": map[string]any{"seg-a": map[string]any{"zonesToGroups": map[string]any{"DEFAULT": []any{}}}},
		"autoScale": "OFF", "maxInstanceCount": 1, "minInstanceCount": 1, "instances": []any{map[string]any{"name": "i1", "credentialId": "c-i1"}}}
}

func ftdvCreate(name string) map[string]any {
	return map[string]any{"name": name, "globalCidrListId": 32, "cxp": "USEAST-AZURE-1", "managementServer": map[string]any{"segment": "seg-a"},
		"segments": []any{"seg-a"}, "segmentOptions": map[string]any{"seg-a": map[string]any{"zonesToGroups": map[string]any{"z1": []any{}}}},
		"maxInstanceCount": 1, "minInstanceCount": 1,
		"instances": []any{map[string]any{"hostName": "i1", "licenseType": "PAY_AS_YOU_GO", "version": "7.4.2-172", "credentialId": "c-i1"}}}
}

func fortinetCreate(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "USEAST-AZURE-1", "version": "7.4.3", "credentialId": "c-1", "licenseType": "PAY_AS_YOU_GO",
		"managementServer": map[string]any{"segment": "seg-a"}, "segments": []any{"seg-a"}, "maxInstanceCount": 1, "minInstanceCount": 1,
		"instances": []any{map[string]any{"name": "i1", "hostName": "i1"}}}
}

// serviceStored returns a stored body: create(name) plus the backend fields of a GET.
func serviceStored(create func(string) map[string]any, extra map[string]any) func() map[string]any {
	return func() map[string]any {
		return withFields(jsonRoundTrip(create("fw-7")), withFields(map[string]any{"description": "seven", "internalName": "fw-7-int", "billingTags": []any{}}, extra))
	}
}

var serviceSpecs = []familySpec{
	{
		name: "service_pan", noun: "PAN firewall service", f: family(writes.ServicePan), coll: "panfwservices", service: "PAN", nameMax: 60,
		notFound: serviceNotFound, create: panCreate, invalid: withCases(serviceNameCases(panCreate), map[string]toolCase{
			"managementSegment 0": {"_create", map[string]any{"body": withFields(panCreate("a"), map[string]any{"managementSegment": 0})}, "managementSegment must be at least 1"},
			"licenseKey":          {"_create", map[string]any{"body": withFields(panCreate("a"), map[string]any{"licenseKey": "k"})}, "unknown field licenseKey"},
			"panWarmBootEnabled":  {"_create", map[string]any{"body": withFields(panCreate("a"), map[string]any{"panWarmBootEnabled": true})}, "panWarmBootEnabled is read-only"},
		}),
		stored:   serviceStored(panCreate, map[string]any{"panWarmBootEnabled": false, "size": "SMALL", "tunnelProtocol": "IPSEC"}),
		readOnly: []string{"internalName", "panWarmBootEnabled", "state"},
		creates: []createCase{
			{nil, panCreate("p1"), withFields(panCreate("p1"), map[string]any{"size": "SMALL", "tunnelProtocol": "IPSEC", "minInstanceCount": 1}), nil},
			{[]string{"BILLING-TAGS"}, withFields(panCreate("p1"), map[string]any{"billingTags": []any{3}}),
				withFields(panCreate("p1"), map[string]any{"billingTags": []any{3}, "size": "SMALL", "tunnelProtocol": "IPSEC", "minInstanceCount": 1}), []string{"BILLING-TAGS"}},
		},
		inUse:    serviceInUse(),
		previews: servicePreviews("PAN firewall service", "c-1, c-i1"),
	},
	{
		name: "service_check_point", noun: "Check Point firewall service", f: family(writes.ServiceCheckPoint), coll: "chkp-fw-services", service: "CHKPFW", nameMax: 60,
		notFound: serviceNotFound, create: checkpointCreate, invalid: serviceNameCases(checkpointCreate),
		stored:   serviceStored(checkpointCreate, map[string]any{"size": "SMALL", "tunnelProtocol": "IPSEC"}),
		readOnly: []string{"internalName", "state"},
		creates: []createCase{{nil, withFields(checkpointCreate("c1"), map[string]any{"billingTags": []any{3}}),
			withFields(checkpointCreate("c1"), map[string]any{"billingTags": []any{3}, "size": "SMALL", "tunnelProtocol": "IPSEC"}), nil}},
		inUse:    serviceInUse(),
		previews: servicePreviews("Check Point firewall service", "c-1, c-i1"),
	},
	{
		name: "service_cisco_ftdv", noun: "Cisco FTDv firewall service", f: family(writes.ServiceCiscoFTDv), coll: "cisco-ftdv-fw-services", service: "CISCO_FTDV_FW", nameMax: 60,
		notFound: serviceNotFound, create: ftdvCreate, invalid: serviceNameCases(ftdvCreate),
		stored:   serviceStored(ftdvCreate, map[string]any{"size": "SMALL", "tunnelProtocol": "IPSEC", "autoScale": "OFF"}),
		readOnly: []string{"internalName", "state"},
		creates: []createCase{{nil, ftdvCreate("f1"),
			withFields(ftdvCreate("f1"), map[string]any{"size": "SMALL", "tunnelProtocol": "IPSEC", "autoScale": "OFF"}), nil}},
		inUse:    serviceInUse(),
		previews: servicePreviews("Cisco FTDv firewall service", "c-i1"),
	},
	{
		name: "service_fortinet", noun: "Fortinet firewall service", f: family(writes.ServiceFortinet), coll: "ftnt-fw-services", service: "FTNTFW", nameMax: 22,
		notFound: serviceNotFound, create: fortinetCreate, invalid: serviceNameCases(fortinetCreate),
		stored:   serviceStored(fortinetCreate, map[string]any{"size": "SMALL", "tunnelProtocol": "IPSEC", "autoScale": "OFF", "scheme": "TERM_BASED"}),
		readOnly: []string{"internalName", "state"},
		// TPS accepts a description with a trailing space (FortinetFWService.java:84-86).
		creates: []createCase{{nil, withFields(fortinetCreate("t1"), map[string]any{"description": "x "}),
			withFields(fortinetCreate("t1"), map[string]any{"description": "x ", "size": "SMALL", "tunnelProtocol": "IPSEC", "autoScale": "OFF", "scheme": "TERM_BASED"}), nil}},
		inUse:    serviceInUse(),
		previews: servicePreviews("Fortinet firewall service", "c-1"),
	},
}

// MA-W11, MA-W50: a service of any type holding the name refuses the create
// with no POST; the guard reads GET /services with marked rows and
// staleOk=false, never the family list or ?name=.
func TestServices_DuplicateGuard(t *testing.T) {
	for _, s := range serviceSpecs {
		t.Run(s.name, func(t *testing.T) {
			e := newFamilyEnv(t, s)
			e.portal.otherServices = []map[string]any{{"id": 40, "name": "dup", "type": "ZIA", "state": "SUCCESS"}}
			mustRefuse(t, e.call(s.name+"_create", map[string]any{"body": s.createBody("dup")}), "duplicate",
				`a service named "dup" already exists (id 40, type ZIA); service names are unique across every service type in the network.`)
			if e.portal.count("POST") != 0 {
				t.Fatal("POST sent")
			}
			for _, r := range e.portal.requests() {
				if r.Path != "/api/tenantnetworks/7/services" || r.Query != "includeMarkedForDeletion=true&staleOk=false" {
					t.Fatalf("guard read %+v", r)
				}
			}
		})
	}
}

// BC-R2, MA-W5: an id of another service type is refused before the family
// GET, and the D13 guard reads the state from the GET /services row, which
// here differs from the object body.
func TestServices_StateGuard(t *testing.T) {
	update := func(e *writeEnv, s familySpec, id string) writeResult {
		return e.call(s.name+"_update", map[string]any{"id": id, "patch": map[string]any{"description": "x"}})
	}
	for _, s := range serviceSpecs {
		t.Run(s.name, func(t *testing.T) {
			e := newFamilyEnv(t, s)
			e.portal.otherServices = []map[string]any{{"id": 40, "name": "zia", "type": "ZIA", "state": "PENDING"}}
			mustRefuse(t, update(e, s, "40"), "family_mismatch", "type ZIA")
			s.put(e, 7, s.stored())
			e.portal.serviceRows = map[int]map[string]any{}
			for _, state := range []string{"FAILED", "IN_PROGRESS", "SCHEDULED"} {
				e.portal.serviceRows[7] = map[string]any{"state": state, "previousState": state}
				mustRefuse(t, update(e, s, "7"), "state_refused", "its state is "+state)
			}
			if e.portal.count("PUT") != 0 {
				t.Fatal("PUT sent on a refused update")
			}
			o := s.stored()
			o["state"] = "FAILED"
			s.put(e, 8, o)
			e.portal.serviceRows[8] = map[string]any{"state": "PENDING", "previousState": "PENDING"}
			mustOK(t, update(e, s, "8"))
			e.portal.networkState = "FAILED"
			mustOK(t, update(e, s, "8"))
			e.portal.serviceRows[8] = map[string]any{"state": "SUCCESS", "previousState": "SUCCESS"}
			mustRefuse(t, update(e, s, "8"), "state_refused", "FAILED")
			if e.portal.count("PUT") != 2 {
				t.Fatalf("PUTs = %d, want 2", e.portal.count("PUT"))
			}
		})
	}
}

// MA-W11, MA-W50: a name held by a row marked for deletion or by a service
// created seconds ago, which the cached GET /services leaves out, refuses the create.
func TestServices_MarkedAndFreshNames(t *testing.T) {
	for _, s := range serviceSpecs {
		t.Run(s.name, func(t *testing.T) {
			e := newFamilyEnv(t, s)
			o := s.stored()
			o["name"], o["docState"] = "gone", "MARKED_FOR_DELETION"
			s.put(e, 7, o)
			o = s.stored()
			o["name"] = "new"
			s.put(e, 8, o)
			e.portal.fresh = map[int]bool{8: true}
			for _, name := range []string{"gone", "new"} {
				mustRefuse(t, e.call(s.name+"_create", map[string]any{"body": s.createBody(name)}), "duplicate", `a service named "`+name+`" already exists`)
			}
			if e.portal.count("POST") != 0 {
				t.Fatal("POST sent")
			}
		})
	}
}

// MA-W51, MA-W53: a delete marks a service for deletion only when neither its
// state nor its previous state is PENDING (TPS ProvisionedState.isProvisioned,
// Defaults.java:546-547); a service edited after a provision (PENDING, previous
// SUCCESS) is removed at once, as TPS does.
func TestServices_DeployedDelete(t *testing.T) {
	for _, c := range []struct {
		state, prev string
		marked      bool
	}{{"SUCCESS", "SUCCESS", true}, {"PENDING", "SUCCESS", false}, {"SUCCESS", "PENDING", false}} {
		s := specNamed("service_fortinet")
		e := newFamilyEnv(t, s)
		o := s.stored()
		o["state"], o["previousState"] = c.state, c.prev
		s.put(e, 7, o)
		preview := e.call(s.name+"_delete", map[string]any{"id": "7"})
		r := e.call(s.name+"_delete", map[string]any{"id": "7", "confirmCode": previewCode(t, preview)})
		mustOK(t, r)
		deployed := strings.Contains(preview.Text, "Deployed: yes")
		marked := strings.Contains(r.Text, "Marked Fortinet firewall service") && e.portal.object(7) != nil
		late := strings.Contains(preview.Text, "once a provision removes it")
		if deployed != c.marked || marked != c.marked || late != c.marked {
			t.Fatalf("%s/%s: preview\n%s\ndelete\n%s", c.state, c.prev, preview.Text, r.Text)
		}
	}
}

// An update keeps an instance by id and adds one without id; a patch that
// leaves out a stored instance id is refused before any write.
func TestServices_InstancesByID(t *testing.T) {
	s := specNamed("service_fortinet")
	e := newFamilyEnv(t, s)
	o := s.stored()
	o["instances"] = []any{map[string]any{"id": 51, "name": "i1", "hostName": "i1", "internalName": "x"}}
	s.put(e, 7, o)
	mustRefuse(t, e.call("service_fortinet_update", map[string]any{"id": "7", "patch": map[string]any{
		"instances": []any{map[string]any{"name": "i2", "hostName": "i2"}}}}), "invalid_arguments", "51")
	mustOK(t, e.call("service_fortinet_update", map[string]any{"id": "7", "patch": map[string]any{"maxInstanceCount": 2, "minInstanceCount": 2,
		"instances": []any{map[string]any{"id": 51}, map[string]any{"name": "i2", "hostName": "i2"}}}}))
	got := sentBody(t, e, "PUT")["instances"]
	want := []any{map[string]any{"id": float64(51), "name": "i1", "hostName": "i1"}, map[string]any{"name": "i2", "hostName": "i2"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("PUT instances = %v", got)
	}
}

// MA-W12: a duplicate the guard cannot see (the 409 of TPS ServiceLogicImpl.java:2174-2184) reads as duplicate.
func TestServices_BackendDuplicate(t *testing.T) {
	for _, s := range serviceSpecs {
		t.Run(s.name, func(t *testing.T) {
			e := newFamilyEnv(t, s)
			e.portal.fail("POST", e.portal.base(), fault{status: 409, body: `{"code":"AK000001","message":"Service with name : dup already exists in network : 7"}`})
			mustRefuse(t, e.call(s.name+"_create", map[string]any{"body": s.createBody("dup")}), "duplicate", "already exists")
			if e.portal.count("POST") != 1 {
				t.Fatal("want one POST, no retry")
			}
		})
	}
}

// MA-W42 to MA-W44: with the async validation flag on, PAN creates and
// updates send ?async=true and wait for the validation state
// (PANFWServiceLogicImpl.java:594-597): a finished poll and one still
// running when the write deadline nears; with the flag off nothing is async.
// The other three families never send it.
func TestServices_AsyncValidation(t *testing.T) {
	queries := func(e *writeEnv) (out []string) {
		for _, req := range e.portal.requests() {
			if req.Method == "POST" || req.Method == "PUT" {
				out = append(out, req.Query)
			}
		}
		return out
	}
	pan := specNamed("service_pan")
	t.Run("finished", func(t *testing.T) {
		pan.flags = []string{asyncFlag}
		e := newFamilyEnv(t, pan)
		e.portal.validation = map[string]string{"WIDGET:DEFAULT:101": `{"id":3,"networkEntityId":"101","state":"SUCCESS"}`,
			"WIDGET:DEFAULT:7": `{"id":4,"networkEntityId":"7","state":"SUCCESS"}`}
		pan.put(e, 7, pan.stored())
		r := e.call("service_pan_create", map[string]any{"body": pan.createBody("a1")})
		u := e.call("service_pan_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}})
		mustOK(t, r)
		mustOK(t, u)
		if q := queries(e); !slices.Equal(q, []string{"async=true", "async=true"}) || !strings.Contains(r.Text, asyncFlag) || strings.Contains(r.Text, "still running") {
			t.Fatalf("queries %v:\n%s", q, r.Text)
		}
	})
	t.Run("timed out", func(t *testing.T) {
		pan.flags = []string{asyncFlag}
		e := newFamilyEnv(t, pan)
		e.setPipeline(handlers.WithWriteTimeout(5 * time.Second))
		r := e.call("service_pan_create", map[string]any{"body": pan.createBody("a1")})
		mustOK(t, r)
		if !strings.Contains(r.Text, "Saved; validation is still running (state id WIDGET:DEFAULT:101).") || !slices.Equal(queries(e), []string{"async=true"}) {
			t.Fatalf("text = %s", r.Text)
		}
	})
	t.Run("flag off", func(t *testing.T) {
		pan.flags = nil
		e := newFamilyEnv(t, pan)
		mustOK(t, e.call("service_pan_create", map[string]any{"body": pan.createBody("a1")}))
		if q := queries(e); !slices.Equal(q, []string{""}) {
			t.Fatalf("queries %v", q)
		}
	})
	for _, s := range serviceSpecs[1:] {
		t.Run(s.name, func(t *testing.T) {
			s.flags = []string{asyncFlag}
			e := newFamilyEnv(t, s)
			r := e.call(s.name+"_create", map[string]any{"body": s.createBody("a1")})
			mustOK(t, r)
			if q := queries(e); !slices.Equal(q, []string{""}) || strings.Contains(r.Text, asyncFlag) {
				t.Fatalf("queries %v:\n%s", q, r.Text)
			}
		})
	}
}
