package registry

import (
	"encoding/json"
	"maps"
	"strings"
	"testing"

	"github.com/alkiranet/mcp-alkira/writes"
)

func init() {
	familySpecs = append(familySpecs, ipsecAdvSpecs...)
	maps.Copy(scenarioRows, map[string]map[string]string{
		"connector_ipsec_adv": {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode TestConnectorIPsecAdv_CreateKeys; K24.1-.5",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go), H-W16 (ST-04)",
			scImportNonexistent: "TestFamilies_NotFound TestConnectorIPsecAdv_Guards; K24.6",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03)",
			scPreflightBlocked:  "TestConnectorIPsecAdv_Guards; MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
			scSoftDeleteOff:     "TestConnectorIPsecAdv_Guards TestFamilies_DeletePreview",
		},
		"connector_ipsec_tunnel_profile": {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode TestIPsecTunnelProfile_ReadTools; K25.1-.6",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go)",
			scImportNonexistent: "TestFamilies_NotFound TestIPsecTunnelProfile_ReadTools; K25.6",
		},
	})
}

// Stored bodies follow the TPS GET shapes (model/ipsec/AdvancedIPSecConnector.java,
// model/ipsec/IPSecTunnelProfile.java).
var ipsecAdvSpecs = []familySpec{
	{
		name: "connector_ipsec_adv", noun: "advanced IPsec connector", f: family(writes.ConnectorIPsecAdv), coll: "adv-ipsec-connectors", connector: "ADV_IP_SEC",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"Could not find connector with id : 404"}`},
		stored:   ipsecAdvStored,
		creates: []createCase{
			{nil, ipsecAdvBody("adv1"), ipsecAdvBody("adv1"), nil},
		},
		readOnly: []string{"implicitGroupId", "type", "segmentId", "tenantId", "previousState"},
		invalid: map[string]toolCase{
			"trailing space": {},
			"key in patch": {"_update", map[string]any{"id": "7", "patch": map[string]any{"gateways": []any{map[string]any{"id": 701,
				"tunnels": []any{map[string]any{"id": "tunnel-IPSEC-a0", "preSharedKey": "k"}}}}}}, "preSharedKey"},
			"bgp key":   {"_update", map[string]any{"id": "7", "patch": map[string]any{"routingOptions": map[string]any{"dynamicRouting": map[string]any{"bgpAuthKeyAlkira": "k"}}}}, "bgpAuthKeyAlkira"},
			"vpn mode":  {"_update", map[string]any{"id": "7", "patch": map[string]any{"vpnMode": "POLICY_BASED"}}, "vpnMode must be one of ROUTE_BASED"},
			"gatewayNo": {"_update", map[string]any{"id": "7", "patch": map[string]any{"gateways": []any{map[string]any{"id": 701, "gatewayNo": 2}}}}, "gatewayNo is read-only"},
		},
		inUse: []inUseCase{
			{fault{status: 400, body: `{"code":"AK000002","message":"connector with name 'adv7' is associated with nat policy 'n'"}`}, "nat policy"},
		},
		previews: []previewCase{
			{"SUCCESS", []string{`Delete preview for advanced IPsec connector "adv7" (id 7).`, "Deployed: yes (state SUCCESS).", `Also deleted with it: "adv7".`,
				"The delete now: marks it for deletion; it keeps carrying traffic, with its name and references, until the deletion is provisioned from the Alkira portal."}},
			{"PENDING", []string{"Deployed: no, never deployed (state PENDING).", `Also deleted with it: "adv7".`, "The delete now: removes it at once."}},
		},
	},
	{
		name: "connector_ipsec_tunnel_profile", noun: "IPsec tunnel profile", f: family(writes.ConnectorIPsecTunnelProfile), coll: "ipsec-tunnel-profiles",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"Could not find IPSecTunnelProfile with id: '404'"}`},
		stored:   tunnelProfileStored,
		create:   tunnelProfileBody,
		readOnly: []string{"tenantId", "tenantNetworkId", "updatedAt"},
		creates: []createCase{
			{nil, tunnelProfileBody("tp1"), tunnelProfileBody("tp1"), nil},
			{nil, withFields(tunnelProfileBody("tp1"), map[string]any{"description": "", "ipSecConfiguration": map[string]any{"encryptionAlgorithm": "AES256GCM16", "dhGroup": "ECP384"}}),
				withFields(tunnelProfileBody("tp1"), map[string]any{"description": "", "ipSecConfiguration": map[string]any{"encryptionAlgorithm": "AES256GCM16", "dhGroup": "ECP384"}}), nil},
		},
		invalid: map[string]toolCase{
			// The profile description has no character or trim rule (IPSecTunnelProfile.java:51-52).
			"trailing space": {},
			"gcm with none": {"_create", map[string]any{"body": withFields(tunnelProfileBody("tp1"), map[string]any{"ipSecConfiguration": map[string]any{"encryptionAlgorithm": "AES128GCM16", "dhGroup": "NONE"}})},
				"not allowed with dhGroup NONE"},
			"lower case":   {"_create", map[string]any{"body": withFields(tunnelProfileBody("tp1"), map[string]any{"ikeConfiguration": map[string]any{"encryptionAlgorithm": "aes256cbc", "integrityAlgorithm": "SHA256", "dhGroup": "MODP2048"}})}, "encryptionAlgorithm must be one of"},
			"no ike":       {"_create", map[string]any{"body": map[string]any{"name": "tp1", "ipSecConfiguration": map[string]any{"encryptionAlgorithm": "AES256CBC", "dhGroup": "MODP2048"}}}, "ikeConfiguration is required"},
			"ike removed":  {"_update", map[string]any{"id": "7", "patch": map[string]any{"ikeConfiguration": nil}}, "ikeConfiguration cannot be removed"},
			"ike dh none":  {"_update", map[string]any{"id": "7", "patch": map[string]any{"ikeConfiguration": map[string]any{"dhGroup": "NONE"}}}, "dhGroup must be one of"},
			"read-only at": {"_update", map[string]any{"id": "7", "patch": map[string]any{"updatedAt": 1}}, "updatedAt is read-only"},
		},
		inUse: []inUseCase{
			{fault{status: 400, body: `{"code":"AK000002","message":"Cannot delete IPSecTunnelProfile with name: tp7 because it is still referenced by connector: adv7"}`}, "still referenced by connector: adv7"},
		},
		previews: []previewCase{
			{"", []string{`Delete preview for IPsec tunnel profile "tp7" (id 7).`, "Deployed: not applicable; this family is not provisioned.",
				"The delete now: removes it at once.", "The next provision: no effect."}},
		},
	},
}

func ipsecAdvStored() map[string]any {
	tunnel := func(id, key, res string) map[string]any {
		return map[string]any{"id": id, "tunnelNo": 0, "preSharedKey": key,
			"customerEnd": map[string]any{"overlayIpReservationId": "c-" + res},
			"cxpEnd":      map[string]any{"publicIp": "20.1.1.1", "publicIpReservationId": "p-" + res, "overlayIpReservationId": "o-" + res}}
	}
	return map[string]any{"name": "adv7", "description": "seven", "tenantId": 239, "cxp": "USEAST-AZURE-1", "implicitGroupId": 61301, "size": "SMALL",
		"type": "ADV_IP_SEC", "enabled": true, "billingTags": []any{}, "advertiseDefaultRoute": false, "advertiseOnPremRoutes": false,
		"destinationType": "IPSEC_ENDPOINT", "segment": "seg-a", "segmentId": 1207, "tunnelsPerGateway": 1,
		"gateways": []any{
			map[string]any{"id": 701, "name": "gw-a", "gatewayNo": 0, "customerGwIp": "192.0.3.10", "haMode": "ACTIVE", "tunnels": []any{tunnel("tunnel-IPSEC-a0", "stored-ka0", "a0")}},
			map[string]any{"id": 702, "name": "gw-b", "gatewayNo": 1, "customerGwIp": "192.0.3.11", "haMode": "STANDBY", "tunnels": []any{tunnel("tunnel-IPSEC-b0", "stored-kb0", "b0")}},
		},
		"vpnMode": "ROUTE_BASED", "routingOptions": map[string]any{"dynamicRouting": map[string]any{"customerGwAsn": "65001", "bgpAuthKeyAlkira": "stored-bgp", "availability": "IPSEC_INTERFACE_PING"}}}
}

func ipsecAdvBody(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "USEAST-AZURE-1", "size": "SMALL", "advertiseDefaultRoute": false, "advertiseOnPremRoutes": false,
		"destinationType": "IPSEC_ENDPOINT", "segment": "seg-a", "tunnelsPerGateway": 1, "vpnMode": "ROUTE_BASED",
		"routingOptions": map[string]any{"staticRouting": map[string]any{"prefixListId": 11}},
		"gateways": []any{map[string]any{"name": "gw-1", "customerGwIp": "203.0.113.9", "tunnels": []any{map[string]any{
			"customerEnd": map[string]any{"overlayIp": "169.254.0.2"},
			"cxpEnd":      map[string]any{"publicIpReservationId": "r-1", "overlayIpReservationId": "r-2"}}}}}}
}

func tunnelProfileStored() map[string]any {
	return map[string]any{"name": "tp7", "description": "seven", "tenantId": 239, "tenantNetworkId": 7, "updatedAt": 1794930274874,
		"ipSecConfiguration": map[string]any{"encryptionAlgorithm": "AES256CBC", "integrityAlgorithm": "SHA256", "dhGroup": "MODP2048"},
		"ikeConfiguration":   map[string]any{"encryptionAlgorithm": "AES256CBC", "integrityAlgorithm": "SHA256", "dhGroup": "MODP2048"}}
}

func tunnelProfileBody(name string) map[string]any {
	return map[string]any{"name": name,
		"ipSecConfiguration": map[string]any{"encryptionAlgorithm": "AES256CBC", "integrityAlgorithm": "SHA256", "dhGroup": "MODP2048"},
		"ikeConfiguration":   map[string]any{"encryptionAlgorithm": "AES256CBC", "integrityAlgorithm": "SHA256", "dhGroup": "MODP2048"}}
}

// D20: connector_ipsec_adv_create takes tunnel pre-shared keys as optional
// write-only fields, never shown in output. The update schema holds no key
// field, and an update resends the stored tunnel keys and BGP key while the
// output masks them.
func TestConnectorIPsecAdv_CreateKeys(t *testing.T) {
	s := specNamed("connector_ipsec_adv")
	e := newFamilyEnv(t, s)
	var res struct {
		Tools []struct {
			Name        string          `json:"name"`
			InputSchema json.RawMessage `json:"inputSchema"`
		} `json:"tools"`
	}
	json.Unmarshal(call(t, e.srv, "tools/list", nil), &res)
	for _, tool := range res.Tools {
		keys := strings.Contains(string(tool.InputSchema), "preSharedKey") || strings.Contains(string(tool.InputSchema), "bgpAuthKeyAlkira")
		switch tool.Name {
		case "connector_ipsec_adv_create":
			if !keys {
				t.Errorf("create schema lacks the key fields")
			}
		case "connector_ipsec_adv_update", "connector_ipsec_adv_delete":
			if keys {
				t.Errorf("%s schema names a key field", tool.Name)
			}
		}
	}
	body := ipsecAdvBody("adv1")
	body["gateways"].([]any)[0].(map[string]any)["tunnels"].([]any)[0].(map[string]any)["preSharedKey"] = "caller-key"
	r := e.call("connector_ipsec_adv_create", map[string]any{"body": body})
	mustOK(t, r)
	tunnel := sentBody(t, e, "POST")["gateways"].([]any)[0].(map[string]any)["tunnels"].([]any)[0].(map[string]any)
	if tunnel["preSharedKey"] != "caller-key" {
		t.Fatalf("POST tunnel %v", tunnel)
	}
	if strings.Contains(r.Text, "caller-key") {
		t.Fatalf("output shows a key:\n%s", r.Text)
	}

	e = newFamilyEnv(t, s)
	s.put(e, 7, s.stored())
	patch := map[string]any{"gateways": []any{
		map[string]any{"id": 702, "tunnels": []any{map[string]any{"id": "tunnel-IPSEC-b0"}}},
		map[string]any{"id": 701, "customerGwIp": "192.0.3.20", "tunnels": []any{map[string]any{"id": "tunnel-IPSEC-a0"}}},
	}}
	r = e.call("connector_ipsec_adv_update", map[string]any{"id": "7", "patch": patch})
	mustOK(t, r)
	put := sentBody(t, e, "PUT")
	gws := put["gateways"].([]any)
	keyOf := func(i int) any {
		return gws[i].(map[string]any)["tunnels"].([]any)[0].(map[string]any)["preSharedKey"]
	}
	if len(gws) != 2 || keyOf(0) != "stored-kb0" || keyOf(1) != "stored-ka0" || gws[1].(map[string]any)["customerGwIp"] != "192.0.3.20" ||
		put["routingOptions"].(map[string]any)["dynamicRouting"].(map[string]any)["bgpAuthKeyAlkira"] != "stored-bgp" {
		t.Fatalf("PUT %s", fmtJSON(put))
	}
	if strings.Contains(r.Text, "stored-ka0") || strings.Contains(r.Text, "stored-bgp") {
		t.Fatalf("output shows a key:\n%s", r.Text)
	}
}

// MA-W4 (gateways, tunnels), MA-W5, MA-W15, MA-W16, MA-W20, MA-W26, MA-W41:
// a tunnel left out or with an unknown id, the D13 state guard, an id of
// another connector family and a connector marked for deletion are refused
// with no write; with the async flag on the update sends ?async=true.
func TestConnectorIPsecAdv_Guards(t *testing.T) {
	s := specNamed("connector_ipsec_adv")
	for name, c := range map[string]struct {
		patch map[string]any
		text  string
	}{
		"tunnel left out": {map[string]any{"gateways": []any{map[string]any{"id": 701, "tunnels": []any{}}, map[string]any{"id": 702}}}, "tunnel-IPSEC-a0 is missing"},
		"unknown tunnel": {map[string]any{"gateways": []any{map[string]any{"id": 701, "tunnels": []any{map[string]any{"id": "tunnel-IPSEC-a0"}, map[string]any{"id": "tunnel-IPSEC-zz"}}},
			map[string]any{"id": 702}}}, "unknown id tunnel-IPSEC-zz"},
		"gateway left out": {map[string]any{"gateways": []any{map[string]any{"id": 701}}}, "id 702 is missing"},
		"half customized": {map[string]any{"gateways": []any{map[string]any{"id": 701}, map[string]any{"id": 702},
			map[string]any{"name": "gw-c", "customerGwIp": "192.0.3.12", "tunnels": []any{map[string]any{}}}}}, "every tunnel carries cxpEnd and customerEnd, or none does"},
		"profile alone": {map[string]any{"gateways": []any{map[string]any{"id": 701, "tunnels": []any{map[string]any{"id": "tunnel-IPSEC-a0", "profileId": 3}}}, map[string]any{"id": 702}}},
			"advanced must be set when profileId is set"},
	} {
		for _, flags := range [][]string{nil, {"RELEASE-ASYNC-NETWORK-ENTITY-VALIDATION-PHASE-1"}} {
			s.flags = flags
			e := newFamilyEnv(t, s)
			s.put(e, 7, s.stored())
			r := e.call("connector_ipsec_adv_update", map[string]any{"id": "7", "patch": c.patch})
			if !r.IsError || !strings.Contains(r.Text, c.text) || e.portal.count("PUT") != 0 {
				t.Errorf("%s flags %v: isError=%v PUT=%d text=%q", name, flags, r.IsError, e.portal.count("PUT"), r.Text)
			}
		}
	}
	s.flags = nil

	for _, c := range []struct {
		row, network string
		ok           bool
	}{{"FAILED", "", false}, {"SCHEDULED", "", false}, {"IN_PROGRESS", "", false}, {"SUCCESS", "FAILED", false}, {"PENDING", "FAILED", true}, {"SUCCESS", "", true}} {
		e := newFamilyEnv(t, s)
		e.portal.networkState = c.network
		o := s.stored()
		o["state"] = c.row
		s.put(e, 7, o)
		r := e.call("connector_ipsec_adv_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}})
		if c.ok != !r.IsError || c.ok != (e.portal.count("PUT") == 1) {
			t.Errorf("row %s network %s: %s", c.row, c.network, r.Text)
		}
	}

	e := newFamilyEnv(t, s)
	e.portal.put(8, s.stored())
	e.portal.connectors = []map[string]any{{"id": 8, "name": "adv7", "type": "IP_SEC", "state": "PENDING"}}
	mustRefuse(t, e.call("connector_ipsec_adv_update", map[string]any{"id": "8", "patch": map[string]any{"description": "x"}}), "family_mismatch", "belongs to another family")
	mustRefuse(t, e.call("connector_ipsec_adv_delete", map[string]any{"id": "8"}), "family_mismatch", "belongs to another family")
	if e.portal.count("PUT")+e.portal.count("DELETE") != 0 {
		t.Fatal("write sent for another family's id")
	}

	e = newFamilyEnv(t, s)
	e.portal.fail("GET", e.portal.base()+"/7", fault{status: 404, body: `{"code":"AK000003","message":"Could not find connector with id : 7"}`})
	o := s.stored()
	o["state"] = "SUCCESS"
	s.put(e, 7, o)
	mustRefuse(t, e.call("connector_ipsec_adv_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}), "state_refused", "deletion pending provisioning")
	if e.portal.count("PUT") != 0 {
		t.Fatal("PUT sent for a marked connector")
	}

	s.flags = []string{"RELEASE-ASYNC-NETWORK-ENTITY-VALIDATION-PHASE-1"}
	e = newFamilyEnv(t, s)
	e.portal.validation = map[string]string{"WIDGET:DEFAULT:7": `{"id":1,"state":"SUCCESS"}`}
	s.put(e, 7, s.stored())
	mustOK(t, e.call("connector_ipsec_adv_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}))
	for _, req := range e.portal.requests() {
		if req.Method == "PUT" && !strings.Contains(req.Query, "async=true") {
			t.Errorf("PUT without async: %+v", req)
		}
	}
}

// K25.2, HS5: connector_ipsec_tunnel_profile_get_by_id reads one profile and
// _get_by_name pages the collection and returns every profile with the name.
func TestIPsecTunnelProfile_ReadTools(t *testing.T) {
	s := specNamed("connector_ipsec_tunnel_profile")
	e := newFamilyEnv(t, s)
	s.put(e, 7, s.stored())
	twin := s.stored()
	twin["description"] = "eight"
	s.put(e, 8, twin)
	other := s.stored()
	other["name"] = "tp9"
	s.put(e, 9, other)
	srv := newServer(providerFor(e.portal.client()))

	res := callTool(t, srv, "connector_ipsec_tunnel_profile_get_by_name", map[string]any{"profileName": "tp7"})
	if res.IsError {
		t.Fatalf("by name: %s", res.Content[0].Text)
	}
	var rows []map[string]any
	if err := json.Unmarshal([]byte(res.Content[0].Text), &rows); err != nil || len(rows) != 2 {
		t.Fatalf("by name: %v %s", err, res.Content[0].Text)
	}
	for _, req := range e.portal.requests() {
		if strings.Contains(req.Query, "name=") || !strings.Contains(req.Query, "paginated=true") {
			t.Errorf("by-name read %+v", req)
		}
	}
	if res := callTool(t, srv, "connector_ipsec_tunnel_profile_get_by_name", map[string]any{"profileName": "none"}); !res.IsError || !strings.Contains(res.Content[0].Text, "not found") {
		t.Errorf("no match: %v %s", res.IsError, res.Content[0].Text)
	}
	if res := callTool(t, srv, "connector_ipsec_tunnel_profile_get_by_id", map[string]any{"profileId": "9"}); res.IsError || !strings.Contains(res.Content[0].Text, `"tp9"`) {
		t.Errorf("by id: %v %s", res.IsError, res.Content[0].Text)
	}
}
