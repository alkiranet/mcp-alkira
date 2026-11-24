package registry

import (
	"fmt"
	"maps"
	"strings"
	"testing"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/writes"
)

func init() {
	familySpecs = append(familySpecs, policySpecs...)
	maps.Copy(scenarioRows, map[string]map[string]string{
		"list_prefix": {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode TestPrefixList_Accepted202; K10.1-.6",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode TestPolicies_DocStateDeleted",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go)",
			scImportNonexistent: "TestFamilies_NotFound",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03)",
			scPreflightBlocked:  "TestFamilies_StateGuard (prefix lists carry no state); MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
		},
		"policy_nat_rule": {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode; K18.1-.6",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode TestPolicies_DocStateDeleted",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go)",
			scImportNonexistent: "TestFamilies_NotFound",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03); E14: TestPolicies_InvalidArgumentsSendNothing",
			scPreflightBlocked:  "TestFamilies_StateGuard (NAT rules carry no state); MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
			scSoftDeleteOff:     "TestFamilies_DocState TestPolicies_DocStateDeleted",
		},
		"policy_nat": {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode; K19.1-.6",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode TestPolicies_DocStateDeleted",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go)",
			scImportNonexistent: "TestFamilies_NotFound",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03)",
			scPreflightBlocked:  "TestPolicies_StateGuard (no TPS state check on update); MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
			scSoftDeleteOff:     "TestFamilies_DocState TestPolicies_DocStateDeleted TestFamilies_DeletePreview",
		},
		"policy_route": {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode; K20.1-.6",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode TestPolicies_DocStateDeleted",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go)",
			scImportNonexistent: "TestFamilies_NotFound",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03); E11 route policy half: TestPolicies_StateGuard",
			scPreflightBlocked:  "TestPolicies_StateGuard; MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
			scSoftDeleteOff:     "TestFamilies_DocState TestPolicies_DocStateDeleted TestFamilies_DeletePreview",
		},
	})
}

var policySpecs = []familySpec{
	{
		name: "list_prefix", noun: "prefix list", f: family(writes.PrefixList), coll: "policy/prefixlists", nameMax: 255,
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"Prefix list 404 not found"}`},
		stored: func() map[string]any {
			return map[string]any{"name": "p7", "description": "seven", "prefixes": []any{"10.1.0.0/24"}, "type": "CIDR",
				"tenantId": 9, "tenantNetworkId": 7, "lastUpdatedDate": 1763380800000}
		},
		readOnly: []string{"tenantId"},
		// The prefix list is live only while its collection lists it, so an update pages the collection after the GET by id.
		updateGets: 2,
		creates: []createCase{
			{nil, map[string]any{"name": "p1", "prefixes": []any{"10.2.0.0/24"}, "prefixRanges": []any{map[string]any{"prefix": "10.3.0.0/16", "le": 24, "ge": 20}}},
				map[string]any{"name": "p1", "prefixes": []any{"10.2.0.0/24"}, "prefixRanges": []any{map[string]any{"prefix": "10.3.0.0/16", "le": 24, "ge": 20}}}, nil},
		},
		inUse: []inUseCase{
			{fault{status: 400, body: `{"code":"AK000002","message":"Validation failed","errors":[{"field":"id","code":"prefixlist.00000003","message":"prefix list is referred by route policy rp1"}]}`}, "prefix list is referred by route policy rp1"},
			{fault{status: 400, body: `{"code":"AK000002","message":"Validation failed","errors":[{"field":"id","code":"prefixlist.00000001","message":"prefix list is used by an active rule"}]}`}, "prefix list is used by an active rule"},
		},
		previews: []previewCase{
			{"", []string{`Delete preview for prefix list "p7" (id 7).`, "Deployed: unknown; prefix lists carry no state.",
				"The next provision: tears it down if it was deployed; otherwise nothing to tear down."}},
		},
	},
	{
		name: "policy_nat_rule", noun: "NAT rule", f: family(writes.NatRule), coll: "nat-rules", nameMax: 255,
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"NAT rule 404 not found"}`},
		stored: func() map[string]any {
			return map[string]any{"name": "r7", "description": "seven", "enabled": true, "category": "DEFAULT",
				"match":     map[string]any{"sourcePrefixes": []any{"10.1.0.0/24"}, "destPrefixes": []any{"any"}, "protocol": "any"},
				"action":    map[string]any{"sourceAddressTranslation": map[string]any{"translationType": "STATIC_IP", "translatedPrefixes": []any{"172.16.0.0/24"}}},
				"segmentId": 3, "lastConfigUpdatedAt": 1763380800000, "updatedBy": "user@example.com"}
		},
		create:  natRuleBody,
		creates: []createCase{{nil, natRuleBody("r1"), withFields(natRuleBody("r1"), map[string]any{"category": "DEFAULT"}), nil}},
		inUse: []inUseCase{
			{fault{status: 400, body: `{"code":"AK000002","message":"NAT rule 7 is being used by NAT policies [n7]"}`}, "is being used by NAT policies [n7]"},
			{fault{status: 409, body: `{"code":"AK000001","message":"Entity is still referred externally"}`}, "still referred externally"},
		},
		previews: []previewCase{{"", []string{`Delete preview for NAT rule "r7" (id 7).`, "Deployed: unknown; NAT rules carry no state."}}},
	},
	{
		name: "policy_nat", noun: "NAT policy", f: family(writes.NatPolicy), coll: "nat-policies",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"nat policy 404 not found"}`},
		stored: func() map[string]any {
			return map[string]any{"name": "n7", "description": "seven", "type": "INTRA_SEGMENT", "segment": "seg1", "segmentId": 3,
				"includedGroups": []any{1}, "excludedGroups": []any{}, "natRuleIds": []any{5}, "category": "DEFAULT", "state": "PENDING"}
		},
		readOnly: []string{"state"},
		create:   natPolicyBody,
		creates:  []createCase{{nil, natPolicyBody("n1"), withFields(natPolicyBody("n1"), map[string]any{"category": "DEFAULT", "type": "INTRA_SEGMENT"}), nil}},
		inUse:    []inUseCase{{fault{status: 409, body: `{"code":"AK000001","message":"Entity is still referred externally"}`}, "still referred externally"}},
		previews: []previewCase{
			{"PENDING", []string{`Delete preview for NAT policy "n7" (id 7).`, "Deployed: no, never deployed (state PENDING).",
				"The delete now: removes it at once.", "The next provision: nothing to tear down."}},
			{"SUCCESS", []string{"Deployed: yes (state SUCCESS).", "marks it for deletion; it keeps carrying traffic, with its name and references, until the deletion is provisioned from the Alkira portal",
				"The next provision: tears down the deployed object."}},
		},
	},
	{
		name: "policy_route", noun: "route policy", f: family(writes.RoutePolicy), coll: "route-policies",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"route policy 404 not found"}`},
		stored: func() map[string]any {
			return map[string]any{"name": "rp7", "description": "seven", "segment": "seg1", "segmentId": 3, "direction": "INBOUND", "enabled": true,
				"includedGroups": []any{1}, "excludedGroups": []any{},
				"rules": []any{map[string]any{"name": "r1", "sequenceNo": 1, "action": "ALLOW", "match": map[string]any{"all": true}}}, "state": "PENDING"}
		},
		readOnly: []string{"state"},
		create:   routePolicyBody,
		creates:  []createCase{{nil, routePolicyBody("rp1"), routePolicyBody("rp1"), nil}},
		inUse:    []inUseCase{{fault{status: 409, body: `{"code":"AK000001","message":"Entity is still referred externally"}`}, "still referred externally"}},
		previews: []previewCase{
			{"SUCCESS", []string{`Delete preview for route policy "rp7" (id 7).`, "Deployed: yes (state SUCCESS).",
				"marks it for deletion; it keeps carrying traffic, with its name and references, until the deletion is provisioned from the Alkira portal", "The next provision: tears down the deployed object."}},
		},
	},
}

func natRuleBody(name string) map[string]any {
	return map[string]any{"name": name, "enabled": true,
		"match":  map[string]any{"sourcePrefixes": []any{"10.1.0.0/24"}, "destPrefixes": []any{"any"}, "protocol": "any"},
		"action": map[string]any{"sourceAddressTranslation": map[string]any{"translationType": "STATIC_IP", "translatedPrefixes": []any{"172.16.0.0/24"}}}}
}

func natPolicyBody(name string) map[string]any {
	return map[string]any{"name": name, "segment": "seg1", "includedGroups": []any{1}, "natRuleIds": []any{5}}
}

func routePolicyBody(name string) map[string]any {
	return map[string]any{"name": name, "segment": "seg1", "direction": "INBOUND", "enabled": true, "includedGroups": []any{1},
		"rules": []any{map[string]any{"name": "r1", "action": "ALLOW", "match": map[string]any{"all": true}}}}
}

// MA-W2, E14: policy-specific invalid arguments are refused with zero requests.
func TestPolicies_InvalidArgumentsSendNothing(t *testing.T) {
	ports := natRuleBody("r1")
	ports["match"] = map[string]any{"sourcePortList": []any{"80", "443"}}
	noMatch := natRuleBody("r1")
	delete(noMatch, "match")
	noAction := natRuleBody("r1")
	delete(noAction, "action")
	noRuleIDs := natPolicyBody("n1")
	delete(noRuleIDs, "natRuleIds")
	twoLists := routePolicyBody("rp1")
	twoLists["rules"] = []any{map[string]any{"name": "r1", "action": "ALLOW", "match": map[string]any{"prefixListIds": []any{1, 2}}}}
	cases := []struct {
		family, tool string
		args         map[string]any
		want         string
	}{
		{"list_prefix", "_create", map[string]any{"body": map[string]any{"name": "p1", "type": "CIDR"}}, "type is read-only"},
		{"list_prefix", "_update", map[string]any{"id": "7", "patch": map[string]any{"type": "CIDR"}}, "type is read-only"},
		{"policy_nat_rule", "_create", map[string]any{"body": withFields(natRuleBody("r1"), map[string]any{"name": strings.Repeat("a", 256)})}, "name is 256 characters long, at most 255 allowed"},
		{"policy_nat_rule", "_create", map[string]any{"body": withFields(natRuleBody("r1"), map[string]any{"description": strings.Repeat("a", 256)})}, "description is 256 characters long, at most 255 allowed"},
		{"policy_nat_rule", "_create", map[string]any{"body": noMatch}, "match is required"},
		{"policy_nat_rule", "_create", map[string]any{"body": noAction}, "action is required"},
		{"policy_nat_rule", "_create", map[string]any{"body": ports}, "match.sourcePortList has 2 items, at most 1 allowed"},
		{"policy_nat_rule", "_update", map[string]any{"id": "7", "patch": map[string]any{"match": map[string]any{"destPortList": []any{"1", "2"}}}}, "match.destPortList has 2 items, at most 1 allowed"},
		{"policy_nat_rule", "_create", map[string]any{"body": withFields(natRuleBody("r1"), map[string]any{"direction": "SIDEWAYS"})}, "direction must be one of"},
		{"policy_nat", "_create", map[string]any{"body": noRuleIDs}, "natRuleIds is required"},
		{"policy_nat", "_create", map[string]any{"body": withFields(natPolicyBody("n1"), map[string]any{"direction": "INBOUND"})}, "unknown field direction"},
		{"policy_nat", "_update", map[string]any{"id": "7", "patch": map[string]any{"natRuleIds": []any{"x"}}}, "natRuleIds[0] must be an integer"},
		{"policy_route", "_create", map[string]any{"body": twoLists}, "prefixListIds has 2 items, at most 1 allowed"},
		{"policy_route", "_create", map[string]any{"body": withFields(routePolicyBody("rp1"), map[string]any{"direction": nil})}, "direction is required"},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("%s-%d", c.family, i), func(t *testing.T) {
			e := newFamilyEnv(t, specNamed(c.family))
			mustRefuse(t, e.call(c.family+c.tool, c.args), "invalid_arguments", c.want)
			if n := len(e.portal.requests()); n != 0 {
				t.Fatalf("%d requests sent", n)
			}
		})
	}
}

// MA-W2, MA-W4: a direction change is refused with no PUT, with and without a rename; the same direction passes.
func TestPolicies_ImmutableDirection(t *testing.T) {
	for _, c := range []struct {
		family, from, to string
		flags            []string
	}{
		{"policy_nat_rule", "INBOUND", "OUTBOUND", []string{"RELEASE-NAT-POLICY-OUTBOUND"}},
		{"policy_route", "INBOUND", "OUTBOUND", nil},
	} {
		for _, rename := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/rename=%v", c.family, rename), func(t *testing.T) {
				s := specNamed(c.family)
				s.flags = c.flags
				e := newFamilyEnv(t, s)
				o := s.stored()
				o["direction"] = c.from
				e.portal.put(7, o)
				patch := map[string]any{"direction": c.to}
				if rename {
					patch["name"] = "renamed"
				}
				mustRefuse(t, e.call(c.family+"_update", map[string]any{"id": "7", "patch": patch}), "invalid_arguments", "direction cannot be changed")
				if e.portal.count("PUT") != 0 {
					t.Fatal("PUT sent")
				}
			})
		}
	}
	e := newFamilyEnv(t, specNamed("policy_route"))
	e.portal.put(7, specNamed("policy_route").stored())
	mustOK(t, e.call("policy_route_update", map[string]any{"id": "7", "patch": map[string]any{"direction": "INBOUND", "description": "x"}}))

	// With RELEASE-NAT-POLICY-OUTBOUND off the direction patch is left out, so no direction change is sent.
	n := newFamilyEnv(t, specNamed("policy_nat_rule"))
	o := specNamed("policy_nat_rule").stored()
	o["direction"] = "INBOUND"
	n.portal.put(7, o)
	r := n.call("policy_nat_rule_update", map[string]any{"id": "7", "patch": map[string]any{"direction": "OUTBOUND", "name": "renamed"}})
	mustOK(t, r)
	if !strings.Contains(r.Text, "Left out because their tenant feature is off: direction.") {
		t.Fatalf("text = %s", r.Text)
	}
	if body := sentBody(t, n, "PUT"); body["direction"] != "INBOUND" {
		t.Fatalf("PUT body = %v", body)
	}
}

// MA-W5: route policy runs the D13 state check before a PUT; NAT policy has none in TPS.
func TestPolicies_StateGuard(t *testing.T) {
	cases := []struct {
		state, network string
		refused        bool
	}{
		{"FAILED", "", true},
		{"SCHEDULED", "", true},
		{"IN_PROGRESS", "", true},
		{"SUCCESS", "FAILED", true},
		{"PENDING", "FAILED", false},
		{"SUCCESS", "SUCCESS", false},
	}
	for _, c := range cases {
		t.Run(c.state+"/"+c.network, func(t *testing.T) {
			s := specNamed("policy_route")
			e := newFamilyEnv(t, s)
			e.portal.networkState = c.network
			o := s.stored()
			o["state"] = c.state
			e.portal.put(7, o)
			r := e.call("policy_route_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}})
			networkReads := 0
			for _, req := range e.portal.requests() {
				if req.Path == "/api/tenantnetworks/7" {
					networkReads++
				}
			}
			if c.refused {
				mustRefuse(t, r, "state_refused", "route policy id 7 cannot change")
				if e.portal.count("PUT") != 0 {
					t.Fatal("PUT sent")
				}
			} else {
				mustOK(t, r)
				if e.portal.count("PUT") != 1 {
					t.Fatal("want one PUT")
				}
			}
			// The network is read only when the entity state alone does not decide.
			if want := map[bool]int{true: 1, false: 0}[c.state == "SUCCESS"]; networkReads != want {
				t.Fatalf("%d network reads, want %d", networkReads, want)
			}
		})
	}

	s := specNamed("policy_nat")
	e := newFamilyEnv(t, s)
	o := s.stored()
	o["state"] = "FAILED"
	e.portal.put(7, o)
	mustOK(t, e.call("policy_nat_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}))
	if e.portal.count("PUT") != 1 {
		t.Fatal("want one PUT for a FAILED NAT policy")
	}
	if ak.NatPolicyDescriptor().StateGuard != 0 || ak.RoutePolicyDescriptor().StateGuard != ak.StateGuardD13 {
		t.Fatal("state guards differ from TPS")
	}
}

// A NAT rule body carries no state, but TPS hard-deletes a rule no live
// policy holds, so its delete says it removes the rule at once.
func TestPolicies_NatRuleDeleteRemovesAtOnce(t *testing.T) {
	s := specNamed("policy_nat_rule")
	e := newFamilyEnv(t, s)
	s.put(e, 7, s.stored())
	r := e.call("policy_nat_rule_delete", map[string]any{"id": "7"})
	code := previewCode(t, r)
	if !strings.Contains(r.Text, "The delete now: removes it at once.\n") || strings.Contains(r.Text, "marks it for deletion") {
		t.Fatalf("preview = %s", r.Text)
	}
	r = e.call("policy_nat_rule_delete", map[string]any{"id": "7", "confirmCode": code})
	mustOK(t, r)
	if !strings.HasPrefix(r.Text, `Deleted NAT rule "r7" (id 7).`) || strings.Contains(r.Text, "marked for deletion") {
		t.Fatalf("text = %s", r.Text)
	}
}

// MA-W16: a DELETED row from the include-marked GET is a residue row; a policy body without docState is ADDED.
// list_prefix is left out: policy-service sends no docState for a prefix list.
func TestPolicies_DocStateDeleted(t *testing.T) {
	for _, name := range []string{"policy_nat_rule", "policy_nat", "policy_route"} {
		t.Run(name, func(t *testing.T) {
			s := specNamed(name)
			e := newFamilyEnv(t, s)
			o := s.stored()
			o["docState"] = "DELETED"
			e.portal.put(7, o)
			mustRefuse(t, e.call(name+"_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}), "state_refused", "already deleted; the backend keeps a residue row")
			mustRefuse(t, e.call(name+"_delete", map[string]any{"id": "7"}), "state_refused", "already deleted")
			if e.portal.count("PUT")+e.portal.count("DELETE") != 0 {
				t.Fatal("write sent")
			}
			e.portal.put(8, s.stored())
			mustOK(t, e.call(name+"_update", map[string]any{"id": "8", "patch": map[string]any{"description": "x"}}))
		})
	}
}

// MA-W16: a prefix list GET by id also answers a deleted residue row, which
// carries no docState; the collection lists live rows only, so an id it does
// not list is refused as already deleted with no write.
func TestPrefixList_ResidueRow(t *testing.T) {
	s := specNamed("list_prefix")
	e := newFamilyEnv(t, s)
	e.portal.put(7, s.stored())
	empty := fault{status: 200, body: `{"data":[],"pagination":{"hits":0}}`}
	e.portal.fail("GET", e.portal.base(), empty)
	e.portal.fail("GET", e.portal.base(), empty)
	mustRefuse(t, e.call("list_prefix_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}), "state_refused", "already deleted")
	mustRefuse(t, e.call("list_prefix_delete", map[string]any{"id": "7"}), "state_refused", "already deleted")
	if e.portal.count("PUT")+e.portal.count("DELETE") != 0 {
		t.Fatal("write sent")
	}
}

// MA-W11, MA-W12: the backend duplicate answers of prefix list and route policy read as duplicate.
func TestPolicies_BackendDuplicate(t *testing.T) {
	for name, f := range map[string]fault{
		"list_prefix":  {status: 400, body: `{"code":"AK000002","message":"Validation failed","errors":[{"field":"name","code":"prefixlist.00000002","message":"prefix list p1 already exists with id 9"}]}`},
		"policy_route": {status: 400, body: `{"code":"AK000001","message":"Validation failed","errors":[{"field":"name","code":"constraint.exists","message":"route policy rp1 already exists"}]}`},
	} {
		t.Run(name, func(t *testing.T) {
			s := specNamed(name)
			e := newFamilyEnv(t, s)
			e.portal.fail("POST", e.portal.base(), f)
			body := s.createBody("p1")
			mustRefuse(t, e.call(name+"_create", map[string]any{"body": body}), "duplicate", "already exists")
			if e.portal.count("POST") != 1 {
				t.Fatal("want one POST")
			}
		})
	}
}

// MA-W23: natRuleIds the backend drops without an error are named in the output.
func TestPolicyNat_EchoDroppedRuleIDs(t *testing.T) {
	e := newFamilyEnv(t, specNamed("policy_nat"))
	e.portal.fail("POST", e.portal.base(), fault{status: 201, body: `{"id":101,"name":"n1","segment":"seg1","includedGroups":[1],"natRuleIds":[5],"category":"DEFAULT","type":"INTRA_SEGMENT"}`})
	r := e.call("policy_nat_create", map[string]any{"body": withFields(natPolicyBody("n1"), map[string]any{"natRuleIds": []any{5, 999}})})
	mustOK(t, r)
	if !strings.Contains(r.Text, "natRuleIds: 1 of 2 elements dropped") {
		t.Fatalf("text = %s", r.Text)
	}
}

// D16: allowOverlappingOutboundDestinationPrefixes needs both outbound flags; allowOverlappingTranslatedPrefixes its own.
func TestPolicyNat_FlagGatedFields(t *testing.T) {
	body := withFields(natPolicyBody("n1"), map[string]any{"allowOverlappingOutboundDestinationPrefixes": true, "allowOverlappingTranslatedPrefixes": true})
	cases := []struct {
		flags             []string
		outbound, overlap bool
	}{
		{nil, false, false},
		{[]string{"RELEASE-NAT-POLICY-OUTBOUND-DNAT"}, false, false},
		{[]string{"RELEASE-NAT-POLICY-OUTBOUND", "RELEASE-NAT-POLICY-OUTBOUND-DNAT"}, true, false},
		{[]string{"NAT-POLICY-ALLOW-OVERLAPPING-TRANSLATED-PREFIXES"}, false, true},
	}
	for i, c := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			s := specNamed("policy_nat")
			s.flags = c.flags
			e := newFamilyEnv(t, s)
			mustOK(t, e.call("policy_nat_create", map[string]any{"body": body}))
			sent := sentBody(t, e, "POST")
			if _, ok := sent["allowOverlappingOutboundDestinationPrefixes"]; ok != c.outbound {
				t.Errorf("allowOverlappingOutboundDestinationPrefixes sent = %v: %v", ok, sent)
			}
			if _, ok := sent["allowOverlappingTranslatedPrefixes"]; ok != c.overlap {
				t.Errorf("allowOverlappingTranslatedPrefixes sent = %v: %v", ok, sent)
			}
		})
	}
}

// policy-service answers PUT and DELETE with 202; both are success.
func TestPrefixList_Accepted202(t *testing.T) {
	s := specNamed("list_prefix")
	e := newFamilyEnv(t, s)
	e.portal.put(7, s.stored())
	e.portal.fail("PUT", e.portal.base()+"/7", fault{kind: faultRewrite, status: 202})
	mustOK(t, e.call("list_prefix_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}))
	code := previewCode(t, e.call("list_prefix_delete", map[string]any{"id": "7"}))
	e.portal.fail("DELETE", e.portal.base()+"/7", fault{kind: faultRewrite, status: 202})
	r := e.call("list_prefix_delete", map[string]any{"id": "7", "confirmCode": code})
	mustOK(t, r)
	if !strings.Contains(r.Text, `Deleted prefix list "p7" (id 7).`) || e.portal.object(7) != nil {
		t.Fatalf("text = %s", r.Text)
	}
}

// MA-W37 (BC-R11): the stored segment was deleted, so the PUT fails on segment, a field the patch does not set.
func TestPolicyNat_DanglingSegment(t *testing.T) {
	const blank = `{"code":"AK000002","message":"Validation failed","errors":[{"field":"segment","code":"NotBlank","message":"segment must not be blank"}]}`
	segPath := "/api/tenantnetworks/7/segments/1326"
	stored := map[string]any{"id": 101, "name": "n1", "type": "INTRA_SEGMENT", "segmentId": 1326, "includedGroups": []any{1},
		"natRuleIds": []any{5}, "category": "DEFAULT", "state": "PENDING", "docState": "ADDED"}
	for _, c := range []struct {
		name    string
		segment fault
		dangles bool
	}{
		{"segment gone", fault{status: 400, body: `{"code":"AK000002","message":"Segment with id: '1326' was not found"}`}, true},
		{"segment present", fault{status: 200, body: `{"id":1326,"name":"seg1"}`}, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			e := newFamilyEnv(t, specNamed("policy_nat"))
			e.portal.put(101, stored)
			e.portal.fail("PUT", e.portal.base()+"/101", fault{status: 400, body: blank})
			e.portal.fail("GET", segPath, c.segment)
			r := e.call("policy_nat_update", map[string]any{"id": "101", "patch": map[string]any{"description": "x"}})
			if !r.IsError || !strings.Contains(r.Text, "segment must not be blank") {
				t.Fatalf("text = %s", r.Text)
			}
			note := "The stored segmentId 1326 names a segment that no longer exists"
			if strings.Contains(r.Text, note) != c.dangles {
				t.Errorf("dangling note shown = %v: %s", !c.dangles, r.Text)
			}
			if strings.Contains(r.Refusal, note) != c.dangles {
				t.Errorf("dangling note in refusal meta = %v: %s", !c.dangles, r.Refusal)
			}
			if n := e.portal.count("PUT"); n != 1 {
				t.Errorf("%d PUTs", n)
			}
		})
	}
	// A patch that sets segment is about the new value: no lookup of the stored one.
	e := newFamilyEnv(t, specNamed("policy_nat"))
	e.portal.put(101, stored)
	e.portal.fail("PUT", e.portal.base()+"/101", fault{status: 400, body: blank})
	r := e.call("policy_nat_update", map[string]any{"id": "101", "patch": map[string]any{"segment": "seg2"}})
	if !r.IsError || strings.Contains(r.Text, "no longer exists") {
		t.Fatalf("text = %s", r.Text)
	}
	for _, q := range e.portal.requests() {
		if q.Path == segPath {
			t.Errorf("stored segment looked up for a patch that sets segment")
		}
	}
}
