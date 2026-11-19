package registry

import (
	"maps"
	"strings"
	"testing"

	"github.com/alkiranet/mcp-alkira/writes"
)

func init() {
	familySpecs = append(familySpecs, segmentResourceSpecs...)
	maps.Copy(scenarioRows, map[string]map[string]string{
		"segment_resource": {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode; K05.1-.5",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go), H-W16 (ST-04)",
			scImportNonexistent: "TestFamilies_NotFound; K05.6",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03)",
			scPreflightBlocked:  "TestSegmentResourceFamilies_StateGuard (segment resources carry no state); MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
		},
		"segment_resource_share": {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode; K06.1-.5",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go), H-W16 (ST-04)",
			scImportNonexistent: "TestFamilies_NotFound; K06.6",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03)",
			scPreflightBlocked:  "TestSegmentResourceFamilies_StateGuard; MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
		},
		"connector_internet": {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode; K22.1-.5",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go), H-W16 (ST-04)",
			scImportNonexistent: "TestFamilies_NotFound TestConnectorInternet_FamilyCheck; K22.6",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03)",
			scPreflightBlocked:  "TestSegmentResourceFamilies_StateGuard TestConnectorInternet_FamilyCheck; MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
			scSoftDeleteOff:     "TestConnectorInternet_DocState TestFamilies_DeletePreview",
		},
		"internet_application": {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode; K21.1-.5",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go), H-W16 (ST-04)",
			scImportNonexistent: "TestFamilies_NotFound; K21.6",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03)",
			scPreflightBlocked:  "TestSegmentResourceFamilies_StateGuard; MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
			scSoftDeleteOff:     "TestFamilies_DocState TestFamilies_DeletePreview",
		},
		"byoip": {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode; K04 runs live as negative rows only (no owned prefix)",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go), H-W16 (ST-04)",
			scImportNonexistent: "TestFamilies_NotFound; K04.6",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03)",
			scPreflightBlocked:  "TestSegmentResourceFamilies_StateGuard (BYOIP carries no state); MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
		},
	})
}

// Stored bodies follow the dev captures of the K05, K06, K21 and K22 probes
// (tenant 239, 2026-11-19); BYOIP follows TPS model/BYOIP.java, since the
// tenant holds no prefix.
var segmentResourceSpecs = []familySpec{
	{
		name: "segment_resource", f: family(writes.SegmentResource), coll: "segment-resources",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"Could not find segmentResource with id : 404"}`},
		stored: func() map[string]any {
			return map[string]any{"name": "r7", "description": "seven", "segment": "seg-a",
				"groupPrefixes": []any{map[string]any{"groupId": -1, "prefixListId": -1}}, "groupId": 60284,
				"systemGenerated": false, "segmentId": 690, "lastConfigUpdatedAt": 1795071692164, "subConnectors": []any{}}
		},
		create:   segmentResourceBody,
		readOnly: []string{"groupId", "systemGenerated", "segmentId"},
		creates:  []createCase{{nil, segmentResourceBody("r1"), segmentResourceBody("r1"), nil}},
		invalid: map[string]toolCase{
			"no segment":   {"_create", map[string]any{"body": map[string]any{"name": "r1", "groupPrefixes": []any{map[string]any{"groupId": -1, "prefixListId": -1}}}}, "segment is required"},
			"group -2":     {"_create", map[string]any{"body": map[string]any{"name": "r1", "segment": "seg-a", "groupPrefixes": []any{map[string]any{"groupId": -2, "prefixListId": -1}}}}, "groupPrefixes[0].groupId must be at least -1"},
			"read-only id": {"_update", map[string]any{"id": "7", "patch": map[string]any{"groupId": 5}}, "groupId is read-only"},
		},
		inUse: []inUseCase{
			{fault{status: 400, body: `{"code":"AK000001","message":"Validation failed","errors":[{"field":"segmentResourceShares","code":"constraint.forbidden","message":"Deleting a segment resource with associated segment resource shares is not allowed"}]}`}, "associated segment resource shares"},
			{fault{status: 409, body: `{"code":"AK000001","message":"Cannot delete Segment Resource 'r7' (ID: 7) because it is currently in use by the following Route Policies: 'p'."}`}, "Route Policies"},
		},
		previews: []previewCase{
			{"", []string{`Delete preview for segment resource "r7" (id 7). Nothing was deleted.`, `Also deleted with it: "r7-Group".`,
				"Deployed: unknown; segment resources carry no state.", "The delete now: removes it at once."}},
		},
	},
	{
		name: "segment_resource_share", f: family(writes.SegmentResourceShare), coll: "segment-resource-shares",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"Could not find segmentResourceShare with id : 404"}`},
		stored: func() map[string]any {
			return map[string]any{"name": "sh7", "description": "seven", "designatedSegment": "seg-a", "endAResources": []any{1408},
				"endBResources": []any{1409}, "endARouteLimit": 100, "endBRouteLimit": 100, "direction": "BIDIRECTIONAL",
				"state": "PENDING", "designatedSegmentId": 690, "endASegmentId": 690, "endBSegmentId": 1207, "serviceList": []any{0}}
		},
		create:   segmentResourceShareBody,
		readOnly: []string{"state", "designatedSegmentId"},
		creates: []createCase{
			{nil, segmentResourceShareBody("sh1"), withFields(segmentResourceShareBody("sh1"), map[string]any{"endARouteLimit": 100, "endBRouteLimit": 100}), nil},
		},
		invalid: map[string]toolCase{
			"direction":      {"_create", map[string]any{"body": withFields(segmentResourceShareBody("sh1"), map[string]any{"direction": "BOTH"})}, "direction must be one of UNIDIRECTIONAL, BIDIRECTIONAL"},
			"fromEnd":        {"_create", map[string]any{"body": withFields(segmentResourceShareBody("sh1"), map[string]any{"fromEnd": "C"})}, "fromEnd must be one of A, B"},
			"fromEnd bidir":  {"_create", map[string]any{"body": withFields(segmentResourceShareBody("sh1"), map[string]any{"fromEnd": "A"})}, "fromEnd must not be set on a BIDIRECTIONAL share"},
			"route limit":    {"_create", map[string]any{"body": withFields(segmentResourceShareBody("sh1"), map[string]any{"endARouteLimit": 0})}, "endARouteLimit must be at least 1"},
			"no endB":        {"_create", map[string]any{"body": withoutField(segmentResourceShareBody("sh1"), "endBResources")}, "endBResources is required"},
			"rule list both": {"_create", map[string]any{"body": withFields(segmentResourceShareBody("sh1"), map[string]any{"ruleListId": 3})}, "set either ruleListId or serviceId/serviceList, not both"},
		},
		inUse: []inUseCase{
			{fault{status: 409, body: `{"code":"AK000001","message":"SEGMENT_RESOURCE_SHARE with id 7 is still referred externally"}`}, "still referred externally"},
		},
		previews: []previewCase{
			{"SUCCESS", []string{`Delete preview for segment resource share "sh7" (id 7).`, "Deployed: yes (state SUCCESS).",
				"The delete now: removes it at once.", "The next provision: tears down the deployed object."}},
		},
	},
	{
		name: "connector_internet", noun: "internet connector", f: family(writes.ConnectorInternet), coll: "internetconnectors", connector: "SAAS",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"Could not find connector with id : 404"}`},
		stored: func() map[string]any {
			return map[string]any{"name": "ic7", "description": "seven", "implicitGroupId": 60281, "cxp": "USEAST-AZURE-1",
				"segments": []any{"seg-a"}, "billingTags": []any{}, "size": "SMALL",
				"trafficDistribution": map[string]any{"algorithm": "HASHING", "algorithmAttributes": map[string]any{"keys": "DEFAULT"}},
				"numOfPublicIPs":      2, "enabled": true, "egressIPTypes": []any{"ALKIRA_PUBLIC_IP"}, "notEligibleForScaling": false,
				"numOfPublicIps": 2, "egressIpTypes": []any{"ALKIRA_PUBLIC_IP"}}
		},
		create:   connectorInternetBody,
		readOnly: []string{"implicitGroupId", "numOfPublicIps", "egressIpTypes", "notEligibleForScaling", "previousState"},
		creates: []createCase{
			{nil, connectorInternetBody("ic1"), withFields(connectorInternetBody("ic1"), map[string]any{"numOfPublicIPs": 2, "enabled": true,
				"trafficDistribution": map[string]any{"algorithm": "HASHING", "algorithmAttributes": map[string]any{"keys": "DEFAULT"}}}), nil},
		},
		invalid: map[string]toolCase{
			// The connector description carries no trim rule.
			"trailing space": {},
			"ALK- prefix":    {"_create", map[string]any{"body": connectorInternetBody("ALK-x")}, "name must not start with ALK-"},
			"two segments":   {"_create", map[string]any{"body": withFields(connectorInternetBody("ic1"), map[string]any{"segments": []any{"a", "b"}})}, "segments must hold exactly one segment name"},
			"public IPs":     {"_create", map[string]any{"body": withFields(connectorInternetBody("ic1"), map[string]any{"numOfPublicIPs": 3})}, "numOfPublicIPs must be one of 2, 4, 6, 8"},
			"alias patched":  {"_update", map[string]any{"id": "7", "patch": map[string]any{"numOfPublicIps": 4}}, "numOfPublicIps is read-only"},
		},
		inUse: []inUseCase{
			{fault{status: 400, body: `{"code":"AK000002","message":"connector with name 'ic7' is actively associated with internet facing applications"}`}, "internet facing applications"},
		},
		previews: []previewCase{
			{"SUCCESS", []string{`Delete preview for internet connector "ic7" (id 7).`, "Deployed: yes (state SUCCESS).", `Also deleted with it: "ic7".`,
				"The delete now: marks it for deletion; it keeps its name and references until the next provision."}},
			{"PENDING", []string{"Deployed: no, never deployed (state PENDING).", `Also deleted with it: "ic7".`, "The delete now: removes it at once.", "The next provision: nothing to tear down."}},
		},
	},
	{
		name: "internet_application", f: family(writes.InternetApplication), coll: "internet-applications", path: "/api/v1/tenantnetworks/7/internet-applications",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"internet application with id 404 is not present in tenant with id 239 networkId 7"}`},
		stored: func() map[string]any {
			return map[string]any{"name": "ifa7", "description": "seven", "fqdnPrefix": "ifa7", "segmentName": "seg-a", "connectorType": "IP_SEC",
				"connectorId": 23295, "segmentId": 1207, "size": "SMALL", "sessionLimit": 1000, "billingTags": []any{},
				"internalName": "ifa-x", "cxp": "USEAST-AZURE-1", "appId": 16001, "state": "PENDING", "connectorGroupId": 60282,
				"inboundInternetGroupId": 60286, "inboundInternetConnectorId": 23296, "trafficEnabled": true, "docState": "ADDED",
				"targets":              []any{map[string]any{"type": "IP", "value": "10.253.7.10", "portRanges": []any{"443"}}},
				"inboundConnectorType": "DEFAULT", "internetProtocol": "IPV4"}
		},
		create:   internetApplicationBody,
		readOnly: []string{"segmentId", "sessionLimit", "internalName", "appId", "state", "connectorGroupId", "inboundInternetGroupId", "docState"},
		creates:  []createCase{{nil, internetApplicationBody("ifa1"), withFields(internetApplicationBody("ifa1"), map[string]any{"size": "SMALL"}), nil}},
		invalid: map[string]toolCase{
			"charset":       {"_create", map[string]any{"body": internetApplicationBody("a/b")}, "outside the slug character set"},
			"leading space": {"_create", map[string]any{"body": internetApplicationBody(" a")}, "outside the slug character set"},
			"dns with value": {"_create", map[string]any{"body": withFields(internetApplicationBody("ifa1"), map[string]any{
				"targets": []any{map[string]any{"type": "INTERNAL_DNS", "value": "x", "policyFqdnListId": 4, "portRanges": []any{"-1"}}}})}, "an INTERNAL_DNS target needs policyFqdnListId and no value"},
			"name space":    {"_create", map[string]any{"body": internetApplicationBody("a b")}, "outside the slug character set"},
			"size":          {"_create", map[string]any{"body": withFields(internetApplicationBody("ifa1"), map[string]any{"size": "XSMALL"})}, "size must be one of"},
			"read-only cxp": {"_update", map[string]any{"id": "7", "patch": map[string]any{"cxp": "x"}}, "cxp is read-only"},
		},
		inUse: []inUseCase{
			{fault{status: 409, body: `{"code":"AK000001","message":"Internet Application 'ifa7' is still referred by Probe(s) 'p'"}`}, "Probe(s)"},
		},
		previews: []previewCase{
			{"SUCCESS", []string{`Delete preview for internet application "ifa7" (id 7).`, "The delete now: marks it for deletion",
				`Also deleted with it: the inbound internet connector <cxp>-seg-a-INB-INT of segment "seg-a" on this application's CXP, if no other internet application uses it, and the ALK-INB-INT-GROUP group`}},
		},
	},
	{
		name: "byoip", noun: "BYOIP", f: family(writes.Byoip), coll: "byoips", nameKey: "prefix",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"byoip with id 404 not found"}`},
		stored: func() map[string]any {
			return map[string]any{"prefix": "203.0.113.0/24", "cxp": "USEAST-AZURE-1", "cloudProvider": "AZURE", "description": "seven",
				"extraAttributes": map[string]any{"message": "m", "signature": "s", "publicKey": "k"}, "onboardingStatus": "PENDING", "usageCount": 0}
		},
		create:   byoipBody,
		readOnly: []string{"usageCount"},
		creates:  []createCase{{nil, byoipBody("203.0.113.0/24"), withFields(byoipBody("203.0.113.0/24"), map[string]any{"cloudProvider": "AWS"}), nil}},
		invalid: map[string]toolCase{
			"no ROA":          {"_create", map[string]any{"body": withoutField(byoipBody("203.0.113.0/24"), "extraAttributes")}, "extraAttributes is required"},
			"provider":        {"_create", map[string]any{"body": withFields(byoipBody("203.0.113.0/24"), map[string]any{"cloudProvider": "GCP"})}, "cloudProvider must be one of AWS, AZURE"},
			"read-only usage": {"_update", map[string]any{"id": "7", "patch": map[string]any{"usageCount": 1}}, "usageCount is read-only"},
		},
		inUse: []inUseCase{
			{fault{status: 400, body: `{"code":"AK000002","message":"Cannot delete BYOIP with id 7 because it is already associated with internet connector ic"}`}, "already associated with internet connector"},
		},
		previews: []previewCase{
			{"", []string{`Delete preview for BYOIP "203.0.113.0/24" (id 7). Nothing was deleted.`, "Deployed: unknown; BYOIPs carry no state."}},
		},
	},
}

func segmentResourceBody(name string) map[string]any {
	return map[string]any{"name": name, "segment": "seg-a", "groupPrefixes": []any{map[string]any{"groupId": -1, "prefixListId": -1}}}
}

func segmentResourceShareBody(name string) map[string]any {
	return map[string]any{"name": name, "designatedSegment": "seg-a", "endAResources": []any{1}, "endBResources": []any{2},
		"direction": "BIDIRECTIONAL", "serviceList": []any{0}}
}

func connectorInternetBody(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "USEAST-AZURE-1", "segments": []any{"seg-a"}}
}

func internetApplicationBody(name string) map[string]any {
	return map[string]any{"name": name, "segmentName": "seg-a", "connectorId": 23295, "fqdnPrefix": "ifa1",
		"targets": []any{map[string]any{"type": "IP", "value": "10.1.1.1", "portRanges": []any{"443"}}}}
}

func byoipBody(prefix string) map[string]any {
	return map[string]any{"prefix": prefix, "cxp": "USEAST-AZURE-1",
		"extraAttributes": map[string]any{"message": "m", "signature": "s", "publicKey": "k"}}
}

// withoutField returns a copy of m without key.
func withoutField(m map[string]any, key string) map[string]any {
	out := withFields(m, nil)
	delete(out, key)
	return out
}

// TPS sets onboardingStatus, poolId and the Azure prefix names from the PUT
// body while no connector uses the prefix (BYOIPServiceImpl.java:233-237), so
// a description update resends the stored values of a READY_TO_USE prefix;
// prefix, cxp and cloudProvider cannot change.
func TestByoip_UpdateResendsState(t *testing.T) {
	s := specNamed("byoip")
	e := newFamilyEnv(t, s)
	ready := s.stored()
	ready["onboardingStatus"], ready["poolId"] = "READY_TO_USE", "/subscriptions/x/resourceGroups/rg/providers/Microsoft.Network/publicIPPrefixes/p1"
	ready["extraAttributes"].(map[string]any)["tenantCustomPrefixName"] = "custom-1"
	ready["extraAttributes"].(map[string]any)["tenantPublicPrefixName"] = "public-1"
	e.portal.put(7, ready)
	mustOK(t, e.call("byoip_update", map[string]any{"id": "7", "patch": map[string]any{"description": "new"}}))
	put := sentBody(t, e, "PUT")
	attrs, _ := put["extraAttributes"].(map[string]any)
	if put["onboardingStatus"] != "READY_TO_USE" || put["poolId"] != ready["poolId"] || attrs["tenantCustomPrefixName"] != "custom-1" || attrs["tenantPublicPrefixName"] != "public-1" {
		t.Fatalf("stored BYOIP state not resent: %v", put)
	}
	mustRefuse(t, e.call("byoip_update", map[string]any{"id": "7", "patch": map[string]any{"onboardingStatus": "PENDING"}}), "invalid_arguments", "onboardingStatus is read-only")

	e = newFamilyEnv(t, s)
	e.portal.put(7, s.stored())
	mustRefuse(t, e.call("byoip_update", map[string]any{"id": "7", "patch": map[string]any{"prefix": "198.51.100.0/24"}}), "invalid_arguments", "prefix cannot be changed after create")
	if e.portal.count("PUT") != 0 {
		t.Fatal("PUT sent")
	}
}

// MA-W5: the state guards: FAILED share and application refused; the
// internet connector runs the D13 guard on its GET /connectors row state;
// the segment resource carries no state.
func TestSegmentResourceFamilies_StateGuard(t *testing.T) {
	for _, name := range []string{"segment_resource_share", "internet_application"} {
		s := specNamed(name)
		e := newFamilyEnv(t, s)
		o := s.stored()
		o["state"] = "FAILED"
		e.portal.put(7, o)
		mustRefuse(t, e.call(name+"_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}), "state_refused", "state is FAILED")
		if e.portal.count("PUT") != 0 {
			t.Fatal("PUT sent")
		}
	}
	s := specNamed("connector_internet")
	for _, c := range []struct {
		row, network string
		ok           bool
	}{{"FAILED", "", false}, {"SCHEDULED", "", false}, {"IN_PROGRESS", "", false}, {"SUCCESS", "FAILED", false}, {"PENDING", "FAILED", true}, {"SUCCESS", "", true}} {
		e := newFamilyEnv(t, s)
		e.portal.networkState = c.network
		o := s.stored()
		o["state"] = c.row
		s.put(e, 7, o)
		r := e.call("connector_internet_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}})
		if c.ok != !r.IsError || c.ok != (e.portal.count("PUT") == 1) {
			t.Errorf("row %s network %s: %s", c.row, c.network, r.Text)
		}
		if !c.ok && r.Code != "state_refused" {
			t.Errorf("row %s network %s: code %s", c.row, c.network, r.Code)
		}
	}
	sr := specNamed("segment_resource")
	e := newFamilyEnv(t, sr)
	o := sr.stored()
	o["state"] = "FAILED"
	e.portal.put(7, o)
	mustOK(t, e.call("segment_resource_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}))
}

// MA-W26, MA-W41: an IPsec id through the internet connector tools is refused
// with no PUT or DELETE, by the missing numOfPublicIPs or the row type.
func TestConnectorInternet_FamilyCheck(t *testing.T) {
	s := specNamed("connector_internet")
	sibling := map[string]any{"name": "ipsec8", "implicitGroupId": 60282, "cxp": "USEAST-AZURE-1", "segments": []any{"seg-a"},
		"billingTags": []any{}, "size": "SMALL", "enabled": true, "notEligibleForScaling": false,
		"trafficDistribution": map[string]any{"algorithm": "HASHING", "algorithmAttributes": map[string]any{"keys": "DEFAULT"}}}
	for name, body := range map[string]map[string]any{"sibling shape": sibling, "row type": s.stored()} {
		t.Run(name, func(t *testing.T) {
			e := newFamilyEnv(t, s)
			e.portal.put(8, body)
			e.portal.connectors = []map[string]any{{"id": 8, "name": body["name"], "type": "IP_SEC", "state": "PENDING"}}
			mustRefuse(t, e.call("connector_internet_update", map[string]any{"id": "8", "patch": map[string]any{"description": "x"}}), "family_mismatch", "belongs to another family")
			mustRefuse(t, e.call("connector_internet_delete", map[string]any{"id": "8"}), "family_mismatch", "belongs to another family")
			if e.portal.count("PUT")+e.portal.count("DELETE") != 0 {
				t.Fatal("write sent")
			}
		})
	}
}

// MA-W16: an internet connector marked for deletion answers its GET by id
// with not found; it is found through includeMarkedForDeletion and refused
// before any write.
func TestConnectorInternet_DocState(t *testing.T) {
	s := specNamed("connector_internet")
	e := newFamilyEnv(t, s)
	e.portal.fail("GET", e.portal.base()+"/7", fault{status: 404, body: `{"code":"AK000003","message":"Could not find connector with id : 7"}`})
	o := s.stored()
	o["state"] = "SUCCESS"
	s.put(e, 7, o)
	mustRefuse(t, e.call("connector_internet_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}), "state_refused", "deletion pending provisioning")
	if e.portal.count("PUT") != 0 {
		t.Fatal("PUT sent")
	}
}

// MA-W11: a segment resource whose "<name>-Group" group exists is refused.
func TestSegmentResource_GroupNameSpace(t *testing.T) {
	e := newFamilyEnv(t, specNamed("segment_resource"))
	body := specNamed("segment_resource").createBody("r1")
	e.portal.groups = []map[string]any{{"id": 61, "name": "r1-Group"}}
	mustRefuse(t, e.call("segment_resource_create", map[string]any{"body": body}), "duplicate",
		`a group named "r1-Group" already exists (id 61); the backend creates a group of that name with the segment resource, so the name cannot be used.`)
	if e.portal.count("POST") != 0 {
		t.Fatal("POST sent")
	}
	e.portal.groups = []map[string]any{{"id": 61, "name": "r1"}}
	mustOK(t, e.call("segment_resource_create", map[string]any{"body": body}))
}

// MA-W11: an internet connector name equal to any connector or group name is
// refused; a connector row whose single GET is 404 is skipped.
func TestConnectorInternet_NameSpace(t *testing.T) {
	e := newFamilyEnv(t, specNamed("connector_internet"))
	e.portal.connectors = []map[string]any{{"id": 31, "name": "edge", "type": "IP_SEC"}, {"id": 32, "name": "gone", "type": "AWS_VPC", "gone": true}}
	e.portal.groups = []map[string]any{{"id": 61, "name": "grp"}}
	mustRefuse(t, e.call("connector_internet_create", map[string]any{"body": connectorInternetBody("edge")}), "duplicate",
		`an ip_sec connector named "edge" already exists (id 31); connector names are unique across every connector type and group.`)
	mustRefuse(t, e.call("connector_internet_create", map[string]any{"body": connectorInternetBody("grp")}), "duplicate", `an internet connector named "grp" already exists (id 61).`)
	if e.portal.count("POST") != 0 {
		t.Fatal("POST sent")
	}
	mustOK(t, e.call("connector_internet_create", map[string]any{"body": connectorInternetBody("gone")}))
}

// MA-W23: a field the create echo drops is named.
func TestInternetApplication_EchoDiff(t *testing.T) {
	e := newFamilyEnv(t, specNamed("internet_application"))
	e.portal.fail("POST", e.portal.base(), fault{status: 201, body: `{"id":101,"name":"ifa1","segmentName":"seg-a","connectorId":23295,"size":"SMALL","targets":[{"type":"IP","value":"10.1.1.1","portRanges":["443"]}]}`})
	r := e.call("internet_application_create", map[string]any{"body": internetApplicationBody("ifa1")})
	mustOK(t, r)
	if !strings.Contains(r.Text, "fqdnPrefix dropped") {
		t.Fatalf("text = %s", r.Text)
	}
}

// D16: a flag-gated field set while its flag is off is left out and named.
func TestSegmentResourceFamilies_FlagGates(t *testing.T) {
	for _, c := range []struct {
		family, field string
		value         any
		flag          string
	}{
		{"connector_internet", "billingTags", []any{7}, "BILLING-TAGS"},
		{"internet_application", "internetProtocol", "IPV6", "IFA-IPV6-SUPPORT"},
		{"byoip", "doNotAdvertise", true, "AKAMAI-PRO-PHASE-1"},
	} {
		for _, on := range []bool{false, true} {
			s := specNamed(c.family)
			if on {
				s.flags = []string{c.flag}
			}
			e := newFamilyEnv(t, s)
			body := s.createBody("n1")
			body[c.field] = c.value
			r := e.call(c.family+"_create", map[string]any{"body": body})
			mustOK(t, r)
			_, sent := sentBody(t, e, "POST")[c.field]
			if sent != on {
				t.Errorf("%s flag %v: %s sent %v\n%s", c.family, on, c.field, sent, r.Text)
			}
			if !on && !strings.Contains(r.Text, "Left out because their tenant feature is off: "+c.field+".") {
				t.Errorf("%s: drop not named:\n%s", c.family, r.Text)
			}
		}
	}
}
