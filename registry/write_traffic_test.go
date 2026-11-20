package registry

import (
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"testing"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/writes"
)

const ztaFlag = "RELEASE-ZTNA-POLICY-WITH-PROFILES"

func init() {
	familySpecs = append(familySpecs, trafficSpecs...)
	maps.Copy(scenarioRows, map[string]map[string]string{
		"policy_traffic_rule": {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode TestTraffic_Accepted202; K15.1-.6",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode TestTraffic_DocStateMarked",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go)",
			scImportNonexistent: "TestFamilies_NotFound",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03); TestTraffic_InvalidArgumentsSendNothing",
			scPreflightBlocked:  "TestTraffic_NoStateGuard (policy-service has no state check on update); MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
			scSoftDeleteOff:     "TestFamilies_DocState TestTraffic_DocStateMarked TestFamilies_DeletePreview",
		},
		"policy_traffic_rule_list": {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode TestTraffic_Accepted202; K16.1-.6",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go)",
			scImportNonexistent: "TestFamilies_NotFound",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03); TestTraffic_InvalidArgumentsSendNothing",
			scPreflightBlocked:  "TestTraffic_NoStateGuard (rule lists carry no state); MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
		},
		"policy_traffic": {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode TestTraffic_Accepted202; K17.1-.6",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode TestTraffic_DocStateMarked",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go)",
			scImportNonexistent: "TestFamilies_NotFound",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03); TestTraffic_InvalidArgumentsSendNothing",
			scPreflightBlocked:  "TestTraffic_NoStateGuard (policy-service has no state check on update); MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
			scSoftDeleteOff:     "TestFamilies_DocState TestTraffic_DocStateMarked TestFamilies_DeletePreview",
		},
	})
}

var trafficSpecs = []familySpec{
	{
		name: "policy_traffic_rule", noun: "traffic policy rule", f: family(writes.TrafficRule), coll: "policy/rules", nameMax: 255,
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"rule with id 404 is not present"}`},
		stored: func() map[string]any {
			return map[string]any{"name": "r7", "description": "seven", "tenantId": 9, "tenantNetworkId": 7, "internalName": "rul-7",
				"matchCondition": map[string]any{"srcIp": "any", "dstIp": "10.1.0.0/24", "dscp": "any", "protocol": "tcp",
					"srcPortList": []any{"any"}, "dstPortList": []any{"8443", "443"}, "applicationList": []any{}, "applicationFamilyList": []any{12}},
				"ruleAction": map[string]any{"action": "DROP", "serviceTypeList": []any{}, "serviceList": []any{}},
				"state":      "PENDING", "docState": "ADDED", "type": "USER_CONFIGURED", "lastConfigUpdatedAt": 1763380800000}
		},
		create:   trafficRuleBody,
		readOnly: []string{"state", "internalName", "tenantId"},
		creates: []createCase{{nil, trafficRuleBody("r1"), withFields(trafficRuleBody("r1"), map[string]any{
			"matchCondition": withFields(trafficRuleBody("r1")["matchCondition"].(map[string]any), map[string]any{"dscp": "any"})}), nil}},
		inUse: []inUseCase{{fault{status: 400, body: `{"code":"AK000001","message":"Validation failed","errors":[{"field":"id","code":"rule.00000003","message":"Rule 'r7' with ID: 7 cannot be deleted as it is currently being used in rule list"}]}`}, "currently being used in rule list"}},
		previews: []previewCase{
			{"PENDING", []string{`Delete preview for traffic policy rule "r7" (id 7).`, "Deployed: no, never deployed (state PENDING).",
				"The delete now: removes it at once.", "The next provision: nothing to tear down."}},
			{"SUCCESS", []string{"Deployed: yes (state SUCCESS).", "marks it for deletion; it keeps its name and references until the next provision",
				"The next provision: tears down the deployed object."}},
		},
	},
	{
		name: "policy_traffic_rule_list", noun: "traffic policy rule list", f: family(writes.TrafficRuleList), coll: "policy/rulelists", nameMax: 255,
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"rule list with id 404 is not present"}`},
		stored: func() map[string]any {
			return map[string]any{"name": "l7", "description": "seven", "tenantId": 9, "tenantNetworkId": 7,
				"rules": []any{map[string]any{"ruleId": 5, "priority": 1}}, "type": "USER_CONFIGURED", "lastUpdatedDate": 1763380800000}
		},
		create:   trafficRuleListBody,
		readOnly: []string{"tenantId"},
		creates:  []createCase{{nil, trafficRuleListBody("l1"), trafficRuleListBody("l1"), nil}},
		inUse:    []inUseCase{{fault{status: 400, body: `{"code":"AK000001","message":"Validation failed","errors":[{"field":"id","code":"rulelist.00000004","message":"rule list with id 7 is being used in tenant with id 9 network 7"}]}`}, "rule list with id 7 is being used"}},
		previews: []previewCase{
			{"", []string{`Delete preview for traffic policy rule list "l7" (id 7).`, "Deployed: unknown; traffic policy rule lists carry no state.",
				"The next provision: tears it down if it was deployed; otherwise nothing to tear down."}},
		},
	},
	{
		name: "policy_traffic", noun: "traffic policy", f: family(writes.TrafficPolicy), coll: "policy/policies", nameMax: 255,
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"policy with id 404 id not present"}`},
		stored: func() map[string]any {
			return map[string]any{"name": "p7", "description": "seven", "tenantId": 9, "tenantNetworkId": 7, "segmentIds": []any{3},
				"fromGroups": []any{-1}, "toGroups": []any{-1}, "ruleListId": 6, "enabled": true, "bidirectional": false,
				"state": "PENDING", "docState": "ADDED", "type": "USER_CONFIGURED", "internalName": "pol-7"}
		},
		create:   trafficPolicyBody,
		readOnly: []string{"state", "internalName", "tenantId"},
		creates: []createCase{
			{nil, trafficPolicyBody("p1"), withFields(trafficPolicyBody("p1"), map[string]any{"enabled": true}), nil},
			{[]string{ztaFlag}, withFields(trafficPolicyBody("p1"), map[string]any{"enabled": false}), withFields(trafficPolicyBody("p1"), map[string]any{"enabled": false}), nil},
		},
		previews: []previewCase{
			{"PENDING", []string{`Delete preview for traffic policy "p7" (id 7).`, "Deployed: no, never deployed (state PENDING).", "The delete now: removes it at once."}},
			{"SUCCESS", []string{"Deployed: yes (state SUCCESS).", "marks it for deletion; it keeps its name and references until the next provision"}},
		},
	},
}

func trafficRuleBody(name string) map[string]any {
	return map[string]any{"name": name,
		"matchCondition": map[string]any{"srcIp": "any", "dstIp": "10.1.0.0/24", "protocol": "tcp", "srcPortList": []any{"any"}, "dstPortList": []any{"443"}},
		"ruleAction":     map[string]any{"action": "ALLOW"}}
}

func trafficRuleListBody(name string) map[string]any {
	return map[string]any{"name": name, "rules": []any{map[string]any{"ruleId": 5, "priority": 1}}}
}

func trafficPolicyBody(name string) map[string]any {
	return map[string]any{"name": name, "segmentIds": []any{3}, "fromGroups": []any{-1}, "toGroups": []any{-1}, "ruleListId": 6}
}

// MA-W2: traffic policy invalid arguments are refused with zero requests.
func TestTraffic_InvalidArgumentsSendNothing(t *testing.T) {
	noMatch := trafficRuleBody("r1")
	delete(noMatch, "matchCondition")
	noAction := trafficRuleBody("r1")
	delete(noAction, "ruleAction")
	noPorts := trafficRuleBody("r1")
	noPorts["matchCondition"] = withFields(noPorts["matchCondition"].(map[string]any), map[string]any{"srcPortList": []any{}})
	badProto := trafficRuleBody("r1")
	badProto["matchCondition"] = withFields(badProto["matchCondition"].(map[string]any), map[string]any{"protocol": "TCP"})
	twoApps := trafficRuleBody("r1")
	twoApps["matchCondition"] = withFields(twoApps["matchCondition"].(map[string]any), map[string]any{"applicationList": []any{1, 2}})
	noRuleList := trafficPolicyBody("p1")
	delete(noRuleList, "ruleListId")
	samePriority := []any{map[string]any{"ruleId": 5, "priority": 1}, map[string]any{"ruleId": 6, "priority": 1}}
	sameRule := []any{map[string]any{"ruleId": 5, "priority": 1}, map[string]any{"ruleId": 5, "priority": 2}}
	cases := []struct {
		family, tool string
		args         map[string]any
		want         string
	}{
		{"policy_traffic_rule", "_create", map[string]any{"body": noMatch}, "matchCondition is required"},
		{"policy_traffic_rule", "_create", map[string]any{"body": noAction}, "ruleAction is required"},
		{"policy_traffic_rule", "_create", map[string]any{"body": noPorts}, "matchCondition.srcPortList has 0 items, at least 1 needed"},
		{"policy_traffic_rule", "_create", map[string]any{"body": badProto}, "matchCondition.protocol must be one of icmp, tcp, udp, any"},
		{"policy_traffic_rule", "_create", map[string]any{"body": twoApps}, "matchCondition.applicationList has 2 items, at most 1 allowed"},
		{"policy_traffic_rule", "_update", map[string]any{"id": "7", "patch": map[string]any{"ruleAction": map[string]any{"action": "DENY"}}}, "ruleAction.action must be one of ALLOW, DROP"},
		{"policy_traffic_rule", "_update", map[string]any{"id": "7", "patch": map[string]any{"type": "IMPLICIT_DROP"}}, "type is read-only"},
		{"policy_traffic_rule_list", "_create", map[string]any{"body": map[string]any{"name": "l1"}}, "rules is required"},
		{"policy_traffic_rule_list", "_create", map[string]any{"body": map[string]any{"name": "l1", "rules": samePriority}}, "rules[0] and rules[1] have the same priority 1"},
		{"policy_traffic_rule_list", "_create", map[string]any{"body": map[string]any{"name": "l1", "rules": sameRule}}, "rules[0] and rules[1] have the same ruleId 5"},
		{"policy_traffic_rule_list", "_update", map[string]any{"id": "7", "patch": map[string]any{"rules": samePriority}}, "have the same priority 1"},
		{"policy_traffic_rule_list", "_update", map[string]any{"id": "7", "patch": map[string]any{"rules": []any{}}}, "rules has 0 items, at least 1 needed"},
		{"policy_traffic_rule_list", "_create", map[string]any{"body": map[string]any{"name": "l1", "rules": []any{map[string]any{"ruleId": 5, "priority": 0}}}}, "rules[0].priority must be at least 1"},
		{"policy_traffic", "_create", map[string]any{"body": withFields(trafficPolicyBody("p1"), map[string]any{"ruleListId": 0})}, "ruleListId must be at least 1"},
		{"policy_traffic", "_update", map[string]any{"id": "7", "patch": map[string]any{"ruleListId": 0}}, "ruleListId must be at least 1"},
		{"policy_traffic", "_create", map[string]any{"body": noRuleList}, "ruleListId is required"},
		{"policy_traffic", "_create", map[string]any{"body": withFields(trafficPolicyBody("p1"), map[string]any{"segmentIds": []any{3, 4}})}, "segmentIds has 2 items, at most 1 allowed"},
		{"policy_traffic", "_create", map[string]any{"body": withFields(trafficPolicyBody("p1"), map[string]any{"fromGroups": []any{}})}, "fromGroups has 0 items, at least 1 needed"},
		{"policy_traffic", "_update", map[string]any{"id": "7", "patch": map[string]any{"toGroups": []any{}}}, "toGroups has 0 items, at least 1 needed"},
		{"policy_traffic", "_update", map[string]any{"id": "7", "patch": map[string]any{"enabled": "yes"}}, "enabled must be a boolean"},
		{"policy_traffic", "_update", map[string]any{"id": "7", "patch": map[string]any{"ruleListId": nil}}, "ruleListId cannot be removed"},
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

// MA-W19: the traffic policy enabled default and the rule dscp default equal
// the backend and portal values (Policy.java:77-79; createRule/index.js:198-200),
// and a create with them unset sends them.
func TestTraffic_SchemaDefaults(t *testing.T) {
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
	e := newFamilyEnv(t, specNamed("policy_traffic"))
	json.Unmarshal(call(t, e.srv, "tools/list", nil), &res)
	found := false
	for _, tool := range res.Tools {
		if tool.Name != "policy_traffic_create" {
			continue
		}
		found = true
		var p struct {
			Default json.RawMessage `json:"default"`
		}
		json.Unmarshal(tool.InputSchema.Properties.Body.Properties["enabled"], &p)
		if string(p.Default) != "true" {
			t.Errorf("enabled default = %s", p.Default)
		}
	}
	if !found {
		t.Fatal("policy_traffic_create not listed")
	}
	mustOK(t, e.call("policy_traffic_create", map[string]any{"body": trafficPolicyBody("p1")}))
	if body := sentBody(t, e, "POST"); body["enabled"] != true {
		t.Fatalf("POST body = %v", body)
	}

	r := newFamilyEnv(t, specNamed("policy_traffic_rule"))
	mustOK(t, r.call("policy_traffic_rule_create", map[string]any{"body": trafficRuleBody("r1")}))
	if mc := sentBody(t, r, "POST")["matchCondition"].(map[string]any); mc["dscp"] != "any" {
		t.Fatalf("matchCondition = %v", mc)
	}
	if def := ak.TrafficPolicyDescriptor(); len(def.Flags) != 1 || def.Flags[0].Field != "ztaProfileIds" || def.Flags[0].Flag != ztaFlag {
		t.Fatalf("policy flags = %+v", def.Flags)
	}
}

// MA-W23, D16: ztaProfileIds is sent and its flag named with the flag on; with it
// off the value is left out and named; an echo that drops the ids names them.
func TestTraffic_ZtaProfileIds(t *testing.T) {
	const zp = "2f1c4e7a-9d1e-4b4f-9a43-1d0c6a1b2e33"
	body := withFields(trafficPolicyBody("p1"), map[string]any{"ztaProfileIds": []any{zp}})

	s := specNamed("policy_traffic")
	s.flags = []string{ztaFlag}
	e := newFamilyEnv(t, s)
	r := e.call("policy_traffic_create", map[string]any{"body": body})
	mustOK(t, r)
	if got := sentBody(t, e, "POST")["ztaProfileIds"]; fmt.Sprint(got) != "["+zp+"]" ||
		!strings.Contains(r.Text, "Tenant feature flags that shaped the request: "+ztaFlag+".") {
		t.Fatalf("flag on: %v\n%s", got, r.Text)
	}

	e = newFamilyEnv(t, specNamed("policy_traffic"))
	r = e.call("policy_traffic_create", map[string]any{"body": body})
	mustOK(t, r)
	if _, sent := sentBody(t, e, "POST")["ztaProfileIds"]; sent || !strings.Contains(r.Text, "Left out because their tenant feature is off: ztaProfileIds.") {
		t.Fatalf("flag off:\n%s", r.Text)
	}

	// A tenant without the feature on the backend side drops the ids without an error.
	e = newFamilyEnv(t, s)
	echo, _ := json.Marshal(withFields(trafficPolicyBody("p1"), map[string]any{"id": 101, "enabled": true, "state": "PENDING"}))
	e.portal.fail("POST", e.portal.base(), fault{status: 201, body: string(echo)})
	r = e.call("policy_traffic_create", map[string]any{"body": body})
	mustOK(t, r)
	if !strings.Contains(r.Text, "ztaProfileIds dropped") {
		t.Fatalf("echo diff:\n%s", r.Text)
	}
}

// MA-W25, E18: the backend returns dstPortList in its own order; the update
// echo diff compares by value and reports nothing, while a dropped port is named.
func TestTraffic_PortListOrder(t *testing.T) {
	s := specNamed("policy_traffic_rule")
	e := newFamilyEnv(t, s)
	e.portal.put(7, s.stored())
	reordered := s.stored()
	reordered["description"] = "x"
	reordered["matchCondition"].(map[string]any)["dstPortList"] = []any{"443", "8443", "80"}
	raw, _ := json.Marshal(reordered)
	e.portal.fail("PUT", e.portal.base()+"/7", fault{status: 202, body: string(raw)})
	patch := map[string]any{"description": "x", "matchCondition": map[string]any{"dstPortList": []any{"80", "8443", "443"}}}
	r := e.call("policy_traffic_rule_update", map[string]any{"id": "7", "patch": patch})
	mustOK(t, r)
	if strings.Contains(r.Text, "dstPortList") && strings.Contains(r.Text, "changed by the backend") {
		t.Fatalf("reorder reported:\n%s", r.Text)
	}
	put := sentBody(t, e, "PUT")["matchCondition"].(map[string]any)
	if fmt.Sprint(put["dstPortList"]) != "[80 8443 443]" || fmt.Sprint(put["applicationFamilyList"]) != "[12]" {
		t.Fatalf("PUT matchCondition = %v", put)
	}

	short := s.stored()
	short["matchCondition"].(map[string]any)["dstPortList"] = []any{"443"}
	raw, _ = json.Marshal(short)
	e.portal.fail("PUT", e.portal.base()+"/7", fault{status: 202, body: string(raw)})
	r = e.call("policy_traffic_rule_update", map[string]any{"id": "7", "patch": patch})
	mustOK(t, r)
	if !strings.Contains(r.Text, "matchCondition.dstPortList: 2 of 3 elements dropped") {
		t.Fatalf("dropped port not named:\n%s", r.Text)
	}
}

// The PUT is the stored rule plus the patch: a patch that switches the
// source to srcPrefixListId and nulls srcIp sends one source, as
// policy-service requires (RuleServiceImpl validateRule, rule.00000006).
func TestTraffic_SwitchSourceNullsOldKey(t *testing.T) {
	s := specNamed("policy_traffic_rule")
	e := newFamilyEnv(t, s)
	e.portal.put(7, s.stored())
	patch := map[string]any{"matchCondition": map[string]any{"srcPrefixListId": 9, "srcIp": nil}}
	mustOK(t, e.call("policy_traffic_rule_update", map[string]any{"id": "7", "patch": patch}))
	put := sentBody(t, e, "PUT")["matchCondition"].(map[string]any)
	if _, ok := put["srcIp"]; ok || fmt.Sprint(put["srcPrefixListId"]) != "9" {
		t.Fatalf("PUT matchCondition = %v", put)
	}
	if !strings.Contains(writes.TrafficRule.UpdateNote, "set the old key to null") {
		t.Fatal("update note does not tell the model to null the old key")
	}
}

// While DEPRECATE-QOSMOS-APPLICATION-FAMILY is on, a create or patch that
// sets applicationFamilyList is refused with no write, naming the flag; an
// update sends a stored list empty, the output names the flag and the
// cleared list, and the echo diff does not report the emptied list.
func TestTraffic_QosmosFamilyCleared(t *testing.T) {
	const qosmos = "DEPRECATE-QOSMOS-APPLICATION-FAMILY"
	s := specNamed("policy_traffic_rule")
	s.flags = []string{qosmos}
	e := newFamilyEnv(t, s)
	e.portal.put(7, s.stored())
	const refusal = "Invalid input: matchCondition.applicationFamilyList must be empty while the tenant feature " + qosmos + " is on"
	patch := map[string]any{"matchCondition": map[string]any{"applicationFamilyList": []any{5}}}
	mustRefuse(t, e.call("policy_traffic_rule_update", map[string]any{"id": "7", "patch": patch}), "invalid_arguments", refusal)
	body := trafficRuleBody("r1")
	body["matchCondition"].(map[string]any)["applicationFamilyList"] = []any{5}
	mustRefuse(t, e.call("policy_traffic_rule_create", map[string]any{"body": body}), "invalid_arguments", refusal)
	if n := e.portal.count("POST") + e.portal.count("PUT"); n != 0 {
		t.Fatalf("%d writes sent", n)
	}

	r := e.call("policy_traffic_rule_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}})
	mustOK(t, r)
	if got := sentBody(t, e, "PUT")["matchCondition"].(map[string]any)["applicationFamilyList"]; fmt.Sprint(got) != "[]" {
		t.Fatalf("PUT applicationFamilyList = %v", got)
	}
	for _, want := range []string{"Tenant feature flags that shaped the request: " + qosmos + ".",
		"Stored values sent empty or false, as the portal does, because the backend refuses them under the current tenant features: matchCondition.applicationFamilyList (tenant feature " + qosmos + " is on)."} {
		if !strings.Contains(r.Text, want) {
			t.Fatalf("output lacks %q:\n%s", want, r.Text)
		}
	}
	if strings.Contains(r.Text, "did not keep") {
		t.Fatalf("emptied list reported as dropped:\n%s", r.Text)
	}
}

// MA-W16: the rule 1514 pattern (marked for deletion after its delete
// request succeeded, state SUCCESS) and a DELETED row are refused on update
// and delete with no write; a body with docState ADDED proceeds.
func TestTraffic_DocStateMarked(t *testing.T) {
	for _, name := range []string{"policy_traffic_rule", "policy_traffic"} {
		for ds, want := range map[string]string{"MARKED_FOR_DELETION": "deletion pending provisioning", "DELETED": "already deleted; the backend keeps a residue row"} {
			t.Run(name+"/"+ds, func(t *testing.T) {
				s := specNamed(name)
				e := newFamilyEnv(t, s)
				o := s.stored()
				o["docState"], o["state"] = ds, "SUCCESS"
				e.portal.put(7, o)
				mustRefuse(t, e.call(name+"_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}), "state_refused", want)
				mustRefuse(t, e.call(name+"_delete", map[string]any{"id": "7"}), "state_refused", want)
				if e.portal.count("PUT")+e.portal.count("DELETE") != 0 {
					t.Fatal("write sent")
				}
				for _, r := range e.portal.requests() {
					if r.Method == "GET" && r.Query != "includeMarkedForDeletion=true" {
						t.Fatalf("GET without includeMarkedForDeletion: %+v", r)
					}
				}
				e.portal.put(8, s.stored())
				mustOK(t, e.call(name+"_update", map[string]any{"id": "8", "patch": map[string]any{"description": "x"}}))
			})
		}
	}
}

// MA-W5: policy-service has no state check on update, so FAILED and
// IN_PROGRESS objects are updated with one PUT and the network is not read.
func TestTraffic_NoStateGuard(t *testing.T) {
	for _, name := range []string{"policy_traffic_rule", "policy_traffic_rule_list", "policy_traffic"} {
		for _, state := range []string{"FAILED", "IN_PROGRESS"} {
			t.Run(name+"/"+state, func(t *testing.T) {
				s := specNamed(name)
				e := newFamilyEnv(t, s)
				e.portal.networkState = "FAILED"
				o := s.stored()
				o["state"] = state
				e.portal.put(7, o)
				mustOK(t, e.call(name+"_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}))
				if e.portal.count("PUT") != 1 || e.portal.count("GET") != 1 {
					t.Fatalf("requests = %+v", e.portal.requests())
				}
			})
		}
	}
	for _, d := range []ak.Descriptor{ak.TrafficPolicyRuleDescriptor(), ak.TrafficPolicyRuleListDescriptor(), ak.TrafficPolicyDescriptor()} {
		if d.StateGuard != ak.StateGuardNone || d.Service != ak.ServicePolicy {
			t.Fatalf("descriptor %s: %+v", d.Collection, d)
		}
	}
}

// policy-service answers PUT and DELETE with 202; both read as success.
func TestTraffic_Accepted202(t *testing.T) {
	for _, s := range []familySpec{specNamed("policy_traffic_rule"), specNamed("policy_traffic_rule_list"), specNamed("policy_traffic")} {
		t.Run(s.name, func(t *testing.T) {
			e := newFamilyEnv(t, s)
			e.portal.put(7, s.stored())
			o := s.stored()
			o["description"] = "x"
			raw, _ := json.Marshal(o)
			e.portal.fail("PUT", e.portal.base()+"/7", fault{status: 202, body: string(raw)})
			mustOK(t, e.call(s.name+"_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}))
			code := previewCode(t, e.call(s.name+"_delete", map[string]any{"id": "7"}))
			e.portal.fail("DELETE", e.portal.base()+"/7", fault{status: 202, body: string(raw)})
			r := e.call(s.name+"_delete", map[string]any{"id": "7", "confirmCode": code})
			mustOK(t, r)
			if e.portal.count("DELETE") != 1 || !strings.Contains(r.Text, "(id 7)") {
				t.Fatalf("delete: %s", r.Text)
			}
		})
	}
}

// R rows: the zero trust access profile read tools read the tenant-level
// collection, paged, and one profile by its UUID.
func TestRegister_ZtaProfile_HTTP(t *testing.T) {
	const id = "2f1c4e7a-9d1e-4b4f-9a43-1d0c6a1b2e33"
	client, seen := fakePortal(t, map[string]string{
		"/api/zero-trust-access-profiles":       `{"data":[{"id":"` + id + `","name":"zp-1"}],"pagination":{"offset":0,"limit":50,"hits":1}}`,
		"/api/zero-trust-access-profiles/" + id: `{"id":"` + id + `","name":"zp-1","enabled":true}`,
	})
	srv := newServer(providerFor(client))
	if res := callTool(t, srv, "zta_profile_get_all", map[string]any{}); res.IsError || !strings.Contains(res.Content[0].Text, "zp-1") {
		t.Fatalf("get_all: %+v", res)
	}
	if res := callTool(t, srv, "zta_profile_get_by_id", map[string]any{"profileId": id}); res.IsError || !strings.Contains(res.Content[0].Text, "zp-1") {
		t.Fatalf("get_by_id: %+v", res)
	}
	q := seen.getQueries()
	if paths := seen.get(); len(paths) != 2 || paths[0] != "/api/zero-trust-access-profiles" || paths[1] != "/api/zero-trust-access-profiles/"+id ||
		q[0].Get("paginated") != "true" || q[0].Get("limit") != "50" {
		t.Fatalf("requests = %v %v", paths, q)
	}
	if res := callTool(t, srv, "zta_profile_get_by_id", map[string]any{}); !res.IsError {
		t.Fatal("missing profileId accepted")
	}
}
