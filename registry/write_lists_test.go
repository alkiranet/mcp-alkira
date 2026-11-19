package registry

import (
	"fmt"
	"maps"
	"reflect"
	"strings"
	"testing"

	"github.com/alkiranet/mcp-alkira/handlers"
	"github.com/alkiranet/mcp-alkira/writes"
)

func init() {
	familySpecs = append(familySpecs, listSpecs...)
	maps.Copy(scenarioRows, map[string]map[string]string{
		"list_as_path": {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode TestAsPath_DocState; K07.1-.6",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go)",
			scImportNonexistent: "TestFamilies_NotFound",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03); TestLists_InvalidArgumentsSendNothing",
			scPreflightBlocked:  "TestFamilies_StateGuard (AS-path lists carry no state); MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
		},
		"list_community": {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode; K08.1-.6",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go)",
			scImportNonexistent: "TestFamilies_NotFound",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03); TestLists_InvalidArgumentsSendNothing",
			scPreflightBlocked:  "TestFamilies_StateGuard (community lists carry no state); MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
		},
		"list_extended_community": {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode TestExtendedCommunity_MixedUpdate; K09.1-.6",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go)",
			scImportNonexistent: "TestFamilies_NotFound",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03); TestLists_InvalidArgumentsSendNothing",
			scPreflightBlocked:  "TestFamilies_StateGuard (extended community lists carry no state); MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
		},
		"list_global_cidr": {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode TestGlobalCidr_TagsEdits TestGlobalCidr_DeleteCodeIgnoresReadTime; K11.1-.6",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go)",
			scImportNonexistent: "TestFamilies_NotFound",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03); TestLists_InvalidArgumentsSendNothing",
			scPreflightBlocked:  "TestFamilies_StateGuard (global CIDR lists carry no state); MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
		},
		"list_udr": {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode TestUdr_DuplicateFixedText; K14.1-.6",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go)",
			scImportNonexistent: "TestFamilies_NotFound",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03); TestLists_InvalidArgumentsSendNothing",
			scPreflightBlocked:  "TestFamilies_StateGuard (UDR lists carry no state); MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
		},
		"list_dns_server": {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode TestDnsServer_DanglingSegment; K12.1-.6",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go)",
			scImportNonexistent: "TestFamilies_NotFound",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03); TestLists_InvalidArgumentsSendNothing",
			scPreflightBlocked:  "TestFamilies_StateGuard (DNS server lists carry no state); MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
		},
		"list_policy_fqdn": {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode TestPolicyFqdn_Order; K13.1-.6",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go)",
			scImportNonexistent: "TestFamilies_NotFound",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03); TestLists_InvalidArgumentsSendNothing",
			scPreflightBlocked:  "TestFamilies_StateGuard (policy FQDN lists carry no state); MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
		},
	})
}

var listSpecs = []familySpec{
	{
		name: "list_as_path", noun: "AS-path list", f: family(writes.AsPathList), coll: "as-path-lists",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"asPathList with id 404 not found"}`},
		stored: func() map[string]any {
			return map[string]any{"name": "a7", "description": "seven", "values": []any{"^65001_"}, "docState": "ADDED", "lastConfigUpdatedAt": 1763380800000}
		},
		create:  func(name string) map[string]any { return map[string]any{"name": name, "values": []any{"^65001_"}} },
		creates: []createCase{{nil, map[string]any{"name": "a1", "values": []any{"^65001_", "_65002$"}}, map[string]any{"name": "a1", "values": []any{"^65001_", "_65002$"}}, nil}},
		inUse: []inUseCase{
			{fault{status: 400, body: `{"code":"AK000002","message":"Cannot delete asPathList with name 'a7', it is still associated with route policy"}`}, "still associated with route policy"},
			{fault{status: 409, body: `{"code":"AK000001","message":"AS_PATH_LIST with id 7 is still referred externally"}`}, "still referred externally"},
		},
		previews: []previewCase{{"", []string{`Delete preview for AS-path list "a7" (id 7).`, "Deployed: unknown; AS-path lists carry no state.", "The delete now: removes it at once.",
			"The next provision: tears it down if it was deployed. A provision never sweeps an AS-path list delete, and the delete cannot be selected in a provision of chosen changes"}}},
	},
	{
		name: "list_community", noun: "community list", f: family(writes.CommunityList), coll: "community-lists",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"communityList with id 404 not found"}`},
		stored: func() map[string]any {
			return map[string]any{"name": "c7", "description": "seven", "values": []any{"65001:100"}, "lastConfigUpdatedAt": 1763380800000}
		},
		create:  func(name string) map[string]any { return map[string]any{"name": name, "values": []any{"65001:100"}} },
		creates: []createCase{{nil, map[string]any{"name": "c1", "description": "", "values": []any{"65001:100"}}, map[string]any{"name": "c1", "description": "", "values": []any{"65001:100"}}, nil}},
		inUse:   []inUseCase{{fault{status: 400, body: `{"code":"AK000002","message":"Cannot delete communityList with name 'c7', it is in use by an InterCXP route policy"}`}, "in use by an InterCXP route policy"}},
		previews: []previewCase{{"", []string{`Delete preview for community list "c7" (id 7).`, "Deployed: unknown; community lists carry no state.",
			"The next provision: tears it down if it was deployed; otherwise nothing to tear down."}}},
	},
	{
		name: "list_extended_community", noun: "extended community list", f: family(writes.ExtendedCommunityList), coll: "extended-community-lists",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"extendedCommunityList with id 404 not found"}`},
		stored: func() map[string]any {
			return map[string]any{"name": "e7", "description": "seven", "values": []any{"rt:65001:100"}, "lastConfigUpdatedAt": 1763380800000}
		},
		create:   func(name string) map[string]any { return map[string]any{"name": name, "values": []any{"rt:65001:100"}} },
		creates:  []createCase{{nil, map[string]any{"name": "e1", "values": []any{"rt:65001:100", "soo:65001:2"}}, map[string]any{"name": "e1", "values": []any{"rt:65001:100", "soo:65001:2"}}, nil}},
		inUse:    []inUseCase{{fault{status: 400, body: `{"code":"AK000002","message":"Cannot delete extendedCommunityList with name 'e7', it is still associated with route policy"}`}, "still associated with route policy"}},
		previews: []previewCase{{"", []string{`Delete preview for extended community list "e7" (id 7).`, "Deployed: unknown; extended community lists carry no state."}}},
	},
	{
		name: "list_global_cidr", noun: "global CIDR list", f: family(writes.GlobalCidrList), coll: "global-cidr-lists",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"globalCidrList with id 404 not found"}`},
		stored: func() map[string]any {
			return map[string]any{"name": "gc7", "description": "seven", "cxp": "US-WEST", "values": []any{"10.250.0.0/24"}, "tags": []any{"INFOBLOX"}, "lastConfigUpdatedAt": 1763380800000}
		},
		create:   globalCidrBody,
		creates:  []createCase{{nil, withFields(globalCidrBody("gc1"), map[string]any{"tags": []any{"F5LB"}}), withFields(globalCidrBody("gc1"), map[string]any{"tags": []any{"F5LB"}}), nil}},
		inUse:    []inUseCase{{fault{status: 400, body: `{"code":"AK000002","message":"Cannot delete Global CIDR List gc7 as it is referred by INFOBLOX:ib-1"}`}, "referred by INFOBLOX:ib-1"}},
		previews: []previewCase{{"", []string{`Delete preview for global CIDR list "gc7" (id 7).`, "Deployed: unknown; global CIDR lists carry no state."}}},
	},
	{
		name: "list_udr", noun: "UDR list", f: family(writes.UdrList), coll: "udr-lists",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"udrList with id 404 not found"}`},
		stored: func() map[string]any {
			return map[string]any{"name": "u7", "description": "seven", "cloudProvider": "AZURE",
				"udrs": []any{map[string]any{"prefix": "10.250.0.0/24", "nextHopType": "INTERNET"}}, "lastConfigUpdatedAt": 1763380800000}
		},
		create:   udrBody,
		creates:  []createCase{{nil, udrBody("u1"), udrBody("u1"), nil}},
		inUse:    []inUseCase{{fault{status: 400, body: `{"code":"AK000001","message":"Validation failed","errors":[{"field":"udrList","code":"constraint.forbidden","message":"Cannot delete UDR List 'u7' as it is being referred to by [{type=AZURE_VNET_CONNECTOR, name=vnet-1}]"}]}`}, "being referred to by"}},
		previews: []previewCase{{"", []string{`Delete preview for UDR list "u7" (id 7).`, "Deployed: unknown; UDR lists carry no state."}}},
	},
	{
		name: "list_dns_server", noun: "DNS server list", f: family(writes.DnsServerList), coll: "dns-server-lists",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"dns server list with id 404 not found"}`},
		stored: func() map[string]any {
			return map[string]any{"name": "d7", "description": "seven", "dnsServerIps": []any{"10.250.0.53"}, "segment": "seg1", "segmentId": 3, "lastConfigUpdatedAt": 1763380800000}
		},
		create:   dnsServerBody,
		creates:  []createCase{{nil, dnsServerBody("d1"), dnsServerBody("d1"), nil}},
		inUse:    []inUseCase{{fault{status: 400, body: `{"code":"AK000002","message":"Cannot delete dns server list with name 'd7', it is still associated with policy fqdn list"}`}, "still associated with policy fqdn list"}},
		previews: []previewCase{{"", []string{`Delete preview for DNS server list "d7" (id 7).`, "Deployed: unknown; DNS server lists carry no state."}}},
	},
	{
		name: "list_policy_fqdn", noun: "policy FQDN list", f: family(writes.PolicyFqdnList), coll: "policy-fqdn-lists",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"policy fqdn list with id 404 not found"}`},
		stored: func() map[string]any {
			return map[string]any{"name": "f7", "description": "seven", "fqdns": []any{"b.example.com", "a.example.com"}, "dnsServerListId": 12,
				"metadata": map[string]any{"resolvedIps": map[string]any{}}, "lastConfigUpdatedAt": 1763380800000}
		},
		create:  policyFqdnBody,
		creates: []createCase{{nil, policyFqdnBody("f1"), policyFqdnBody("f1"), nil}},
		inUse: []inUseCase{
			{fault{status: 400, body: `{"code":"AK000002","message":"Policy FQDN List 'f7' is in use by internet application 'ia-1'"}`}, "in use by internet application"},
			{fault{status: 409, body: `{"code":"AK000001","message":"POLICY_FQDN_LIST with id 7 is still referred externally"}`}, "still referred externally"},
		},
		previews: []previewCase{{"", []string{`Delete preview for policy FQDN list "f7" (id 7).`, "Deployed: unknown; policy FQDN lists carry no state."}}},
	},
}

func globalCidrBody(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "US-WEST", "values": []any{"10.250.0.0/24"}}
}

func udrBody(name string) map[string]any {
	return map[string]any{"name": name, "cloudProvider": "AZURE", "udrs": []any{map[string]any{"prefix": "10.250.0.0/24", "nextHopType": "INTERNET"}}}
}

func dnsServerBody(name string) map[string]any {
	return map[string]any{"name": name, "dnsServerIps": []any{"10.250.0.53"}, "segment": "seg1"}
}

func policyFqdnBody(name string) map[string]any {
	return map[string]any{"name": name, "fqdns": []any{"c.example.com", "a.example.com"}, "dnsServerListId": 12}
}

// MA-W2: list-specific arguments the schema or the client-go rules refuse, with zero requests.
func TestLists_InvalidArgumentsSendNothing(t *testing.T) {
	noCXP := globalCidrBody("gc1")
	delete(noCXP, "cxp")
	cases := []struct {
		family, tool string
		args         map[string]any
		want         string
	}{
		{"list_as_path", "_create", map[string]any{"body": map[string]any{"name": "a1"}}, "values is required"},
		{"list_as_path", "_create", map[string]any{"body": map[string]any{"name": "a1", "values": []any{"^65001_", " _1"}}}, "values[1] has a leading or trailing space"},
		{"list_as_path", "_update", map[string]any{"id": "7", "patch": map[string]any{"docState": "ADDED"}}, "docState is read-only"},
		{"list_community", "_update", map[string]any{"id": "7", "patch": map[string]any{"values": nil}}, "values cannot be removed"},
		{"list_extended_community", "_create", map[string]any{"body": map[string]any{"name": "e1", "values": []any{"rt:65001:100", "65001:.*"}}}, "mixes rt or soo values with regex values"},
		{"list_global_cidr", "_create", map[string]any{"body": noCXP}, "cxp is required"},
		{"list_global_cidr", "_update", map[string]any{"id": "7", "patch": map[string]any{"tags": []any{"F5LB"}}}, "tags"},
		{"list_udr", "_create", map[string]any{"body": withFields(udrBody("u1"), map[string]any{"cloudProvider": "AWS"})}, "cloudProvider must be one of AZURE"},
		{"list_udr", "_create", map[string]any{"body": withFields(udrBody("u1"), map[string]any{"udrs": []any{map[string]any{"prefix": "10.1.0.0/24", "nextHopType": "VNET"}}})}, "nextHopType must be one of INTERNET"},
		{"list_udr", "_create", map[string]any{"body": withFields(udrBody("u1"), map[string]any{"udrs": []any{map[string]any{"prefix": "10.1.0.0/24", "nextHopType": "INTERNET", "nextHopValue": "1.1.1.1"}}})}, "nextHopValue is read-only"},
		{"list_dns_server", "_create", map[string]any{"body": map[string]any{"name": "d1", "dnsServerIps": []any{"10.1.0.53"}}}, "segment is required"},
		{"list_dns_server", "_update", map[string]any{"id": "7", "patch": map[string]any{"segmentId": 4}}, "segmentId is read-only"},
		{"list_policy_fqdn", "_create", map[string]any{"body": map[string]any{"name": "f1", "fqdns": []any{"a.example.com"}}}, "dnsServerListId is required"},
		{"list_policy_fqdn", "_create", map[string]any{"body": withFields(policyFqdnBody("f1"), map[string]any{"dnsServerListId": "12"})}, "dnsServerListId must be an integer"},
		{"list_policy_fqdn", "_update", map[string]any{"id": "7", "patch": map[string]any{"metadata": map[string]any{}}}, "metadata is read-only"},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("%s-%d", c.family, i), func(t *testing.T) {
			s := specNamed(c.family)
			e := newFamilyEnv(t, s)
			e.portal.put(7, s.stored())
			r := e.call(c.family+c.tool, c.args)
			mustRefuse(t, r, "invalid_arguments", c.want)
			for _, q := range e.portal.requests() {
				if q.Method != "GET" {
					t.Fatalf("write sent: %+v", q)
				}
			}
		})
	}
}

// The extended community mixing rule runs on the merged update body: a patch
// that adds a regex value to a list of rt values is refused with no PUT.
func TestExtendedCommunity_MixedUpdate(t *testing.T) {
	s := specNamed("list_extended_community")
	e := newFamilyEnv(t, s)
	e.portal.put(7, s.stored())
	mustRefuse(t, e.call("list_extended_community_update", map[string]any{"id": "7", "patch": map[string]any{"values": []any{"rt:65001:100", "65001:.*"}}}),
		"invalid_arguments", "mixes rt or soo values with regex values")
	if e.portal.count("PUT") != 0 {
		t.Fatal("PUT sent")
	}
	mustOK(t, e.call("list_extended_community_update", map[string]any{"id": "7", "patch": map[string]any{"values": []any{"soo:65001:1", "rt:65001:2"}}}))
}

// Global CIDR list tags change through add and remove; the stored tags are kept otherwise.
func TestGlobalCidr_TagsEdits(t *testing.T) {
	s := specNamed("list_global_cidr")
	e := newFamilyEnv(t, s)
	e.portal.put(7, s.stored())
	mustOK(t, e.call("list_global_cidr_update", map[string]any{"id": "7", "add": map[string]any{"tags": []any{"F5LB"}}}))
	if got := e.portal.object(7)["tags"]; !reflect.DeepEqual(got, []any{"INFOBLOX", "F5LB"}) {
		t.Fatalf("tags = %v", got)
	}
	mustOK(t, e.call("list_global_cidr_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}))
	if got := e.portal.object(7)["tags"]; !reflect.DeepEqual(got, []any{"INFOBLOX", "F5LB"}) {
		t.Fatalf("tags after a description patch = %v", got)
	}
}

// MA-W24, E4: a UDR list duplicate answers 409 AK000001 with the raw database
// error; the tool shows the fixed text, never the table or constraint name.
// A 409 that is not the database error keeps its own text.
func TestUdr_DuplicateFixedText(t *testing.T) {
	const dbText = `{"code":"AK000001","message":"ERROR: duplicate key value violates unique constraint \"udr_list_tenant_id_tenant_network_id_name\"\n  Detail: Key (tenant_id, tenant_network_id, name)=(9, 7, u1) already exists."}`
	s := specNamed("list_udr")
	e := newFamilyEnv(t, s)
	e.portal.fail("POST", e.portal.base(), fault{status: 409, body: dbText})
	r := e.call("list_udr_create", map[string]any{"body": udrBody("u1")})
	mustRefuse(t, r, "duplicate", handlers.Article("UDR list")+` named "u1" already exists.`)
	if strings.Contains(r.Text, "udr_list") || strings.Contains(r.Text, "constraint") || e.portal.count("POST") != 1 {
		t.Fatalf("text = %s, %d POSTs", r.Text, e.portal.count("POST"))
	}

	e = newFamilyEnv(t, s)
	e.portal.put(7, s.stored())
	e.portal.fail("PUT", e.portal.base()+"/7", fault{status: 409, body: dbText})
	r = e.call("list_udr_update", map[string]any{"id": "7", "patch": map[string]any{"name": "u2"}})
	mustRefuse(t, r, "duplicate", handlers.Article("UDR list")+` named "u2" already exists.`)
	if strings.Contains(r.Text, "udr_list") {
		t.Fatalf("text = %s", r.Text)
	}

	e = newFamilyEnv(t, s)
	e.portal.fail("POST", e.portal.base(), fault{status: 409, body: `{"code":"AK000001","message":"Request conflicts with a change in progress"}`})
	r = e.call("list_udr_create", map[string]any{"body": udrBody("u1")})
	if !r.IsError || r.Code == "duplicate" || !strings.Contains(r.Text, "Request conflicts with a change in progress") {
		t.Fatalf("plain 409: code %q text %s", r.Code, r.Text)
	}
}

// A DNS server list whose segment was deleted reads back with segmentId and
// without segment (TPS DNSServerListTranslator). An update that does not set
// segment skips the segment lookup and succeeds; one that sets segment
// sends it.
func TestDnsServer_DanglingSegment(t *testing.T) {
	stored := map[string]any{"id": 101, "name": "d1", "dnsServerIps": []any{"10.1.0.53"}, "segmentId": 1326}
	e := newFamilyEnv(t, specNamed("list_dns_server"))
	e.portal.put(101, stored)
	mustOK(t, e.call("list_dns_server_update", map[string]any{"id": "101", "patch": map[string]any{"description": "x"}}))
	if put := sentBody(t, e, "PUT"); put["segment"] != nil {
		t.Fatalf("PUT segment = %v", put["segment"])
	}
	for _, q := range e.portal.requests() {
		if strings.Contains(q.Path, "/segments/") {
			t.Fatalf("segment looked up: %+v", q)
		}
	}
	e = newFamilyEnv(t, specNamed("list_dns_server"))
	e.portal.put(101, stored)
	mustOK(t, e.call("list_dns_server_update", map[string]any{"id": "101", "patch": map[string]any{"segment": "seg1"}}))
	if put := sentBody(t, e, "PUT"); put["segment"] != "seg1" {
		t.Fatalf("PUT segment = %v", put["segment"])
	}
}

// MA-W16: an AS-path list body always carries docState; a body without it,
// and a marked row, are refused with no write.
func TestAsPath_DocState(t *testing.T) {
	s := specNamed("list_as_path")
	e := newFamilyEnv(t, s)
	o := s.stored()
	delete(o, "docState")
	e.portal.put(7, o)
	r := e.call("list_as_path_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}})
	if !r.IsError || r.Code != "state_refused" {
		t.Fatalf("no docState: code %q text %s", r.Code, r.Text)
	}
	o = s.stored()
	o["docState"] = "MARKED_FOR_DELETION"
	e.portal.put(8, o)
	mustRefuse(t, e.call("list_as_path_delete", map[string]any{"id": "8"}), "state_refused", "deletion pending provisioning")
	for _, q := range e.portal.requests() {
		if q.Method != "GET" {
			t.Fatalf("write sent: %+v", q)
		}
	}
}

// fqdns keeps its order on create and on an update that leaves it unchanged.
func TestPolicyFqdn_Order(t *testing.T) {
	s := specNamed("list_policy_fqdn")
	e := newFamilyEnv(t, s)
	mustOK(t, e.call("list_policy_fqdn_create", map[string]any{"body": policyFqdnBody("f1")}))
	if got := sentBody(t, e, "POST")["fqdns"]; !reflect.DeepEqual(got, []any{"c.example.com", "a.example.com"}) {
		t.Fatalf("POST fqdns = %v", got)
	}
	e.portal.put(7, s.stored())
	mustOK(t, e.call("list_policy_fqdn_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}))
	if got := sentBody(t, e, "PUT")["fqdns"]; !reflect.DeepEqual(got, []any{"b.example.com", "a.example.com"}) {
		t.Fatalf("PUT fqdns = %v", got)
	}
	if _, ok := sentBody(t, e, "PUT")["metadata"]; ok {
		t.Fatal("read-only metadata resent")
	}
}

// TPS answers every global CIDR list read with the current time as
// lastConfigUpdatedAt (GlobalCidrList.java:109-112), so the delete code
// ignores it; any other change still refuses the code.
func TestGlobalCidr_DeleteCodeIgnoresReadTime(t *testing.T) {
	s := specNamed("list_global_cidr")
	e := newFamilyEnv(t, s)
	e.portal.put(7, s.stored())
	code := previewCode(t, e.call("list_global_cidr_delete", map[string]any{"id": "7"}))
	e.portal.object(7)["lastConfigUpdatedAt"] = 1763380900000
	mustOK(t, e.call("list_global_cidr_delete", map[string]any{"id": "7", "confirmCode": code}))
	if e.portal.count("DELETE") != 1 {
		t.Fatal("want one DELETE")
	}
}
