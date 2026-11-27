package registry

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/alkiranet/mcp-alkira/writes"
)

func init() {
	familySpecs = append(familySpecs, expressRouteSpec)
	scenarioRows["connector_azure_express_route"] = map[string]string{
		scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode; K33.1-.5 unit rows (no circuit on dev)",
		scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
		scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go), H-W16 (ST-04)",
		scImportNonexistent: "TestFamilies_NotFound TestConnectorAzureExpressRoute_Guards; K33.6 (fake id, live)",
		scWaveCoalescing:    "MA-W13 (ST-03)",
		scRequestFailed:     "MA-W18 (ST-03)",
		scPreflightBlocked:  "TestConnectorAzureExpressRoute_Guards; MA-W13 (ST-03)",
		scMissingRequestID:  "MA-W12 (ST-03)",
		scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
		scSoftDeleteOff:     "TestConnectorAzureExpressRoute_Guards TestFamilies_DeletePreview",
	}
}

const erCircuitTest = "/subscriptions/sub-1/resourceGroups/rg-1/providers/Microsoft.Network/expressRouteCircuits/erc-1"

// The stored body follows the TPS GET shape (AzureExpressRouteConnector.java)
// with IPSEC tunnels, one holding a stored pre-shared key.
var expressRouteSpec = familySpec{
	name: "connector_azure_express_route", noun: "Azure ExpressRoute connector", f: family(writes.ConnectorAzureExpressRoute),
	coll: "azure-express-route-connectors", path: "/api/v1/tenantnetworks/7/azure-express-route-connectors", connector: "EXPRESS_ROUTE",
	notFound: fault{status: 404, body: `{"code":"AK000003","message":"Could not find connector with id : 404"}`},
	stored:   expressRouteStored,
	create:   expressRouteBody,
	readOnly: []string{"implicitGroupId", "primary", "lastEnabled", "previousState"},
	creates: []createCase{
		{nil, expressRouteBody("er1"), withFields(expressRouteBody("er1"), map[string]any{"size": "LARGE", "enabled": true, "tunnelProtocol": "VXLAN_GPE"}), nil},
	},
	invalid: map[string]toolCase{
		"trailing space": {},
		"ALK- prefix":    {"_create", map[string]any{"body": expressRouteBody("ALK-x")}, "name must not start with ALK-"},
		"size":           {"_create", map[string]any{"body": withFields(expressRouteBody("er1"), map[string]any{"size": "4LARGE"})}, "size must be one of"},
		"macs with GPE": {"_create", map[string]any{"body": withFields(expressRouteBody("er1"), map[string]any{"instances": []any{map[string]any{
			"name": "i1", "expressRouteCircuitId": erCircuitTest, "credentialId": "cred-az", "loopbackSubnet": "10.251.0.0/26",
			"gatewayMacAddresses": []any{"00:11:22:33:44:55"}}}})}, "tunnelProtocol VXLAN only"},
		"key in patch": {"_update", map[string]any{"id": "7", "patch": erTunnelPatch(`{"id":"tun-er-1","preSharedKey":"k"},{"id":"tun-er-2"}`)}, "preSharedKey is a secret"},
		"azNo":         {"_update", map[string]any{"id": "7", "patch": erTunnelPatch(`{"id":"tun-er-1","tunnelAzNo":1},{"id":"tun-er-2"}`)}, "tunnelAzNo is read-only"},
	},
	inUse: []inUseCase{
		{fault{status: 409, body: `{"code":"AK000001","message":"CONNECTOR with id 7 is still referred externally"}`}, "referred externally"},
	},
	previews: []previewCase{
		{"PENDING", []string{`Delete preview for Azure ExpressRoute connector "er7" (id 7).`, "Deployed: no, never deployed (state PENDING).", `Also deleted with it: "er7".`, "The delete now: removes it at once."}},
		{"SUCCESS", []string{"Deployed: yes (state SUCCESS).", `Also deleted with it: "er7".`, "The delete now: marks it for deletion; it keeps carrying traffic, with its name and references, until the deletion is provisioned from the Alkira portal."}},
	},
}

func expressRouteStored() map[string]any {
	return map[string]any{"name": "er7", "description": "seven", "cxp": "USEAST-AZURE-1", "implicitGroupId": 61301, "size": "LARGE",
		"billingTags": []any{}, "enabled": true, "primary": true, "vhubPrefix": "10.250.0.0/23",
		"instances": []any{map[string]any{"id": 801, "name": "i1", "expressRouteCircuitId": erCircuitTest, "expressRouteCircuitName": "erc-1",
			"redundantRouter": false, "loopbackSubnet": "10.251.0.0/26", "credentialId": "cred-az",
			"segmentOptions": []any{map[string]any{"segmentName": "seg-a", "segmentId": 3, "customerGateways": []any{map[string]any{
				"name": "gw1", "gatewayNo": 0, "tunnels": []any{
					map[string]any{"name": "t1", "id": "tun-er-1", "tunnelNo": 0, "tunnelAzNo": 0, "preSharedKey": "stored-psk"},
					map[string]any{"name": "t2", "id": "tun-er-2", "tunnelNo": 1, "tunnelAzNo": 1}}}}}}}},
		"segmentOptions": []any{map[string]any{"segmentName": "seg-a", "segmentId": 3, "customerAsn": 65010, "disableInternetExit": true, "advertiseOnPremRoutes": false}},
		"lastEnabled":    1763600000, "tunnelProtocol": "IPSEC"}
}

func expressRouteBody(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "USEAST-AZURE-1", "vhubPrefix": "10.250.0.0/23",
		"instances":      []any{map[string]any{"name": "i1", "expressRouteCircuitId": erCircuitTest, "credentialId": "cred-az", "loopbackSubnet": "10.251.0.0/26"}},
		"segmentOptions": []any{map[string]any{"segmentName": "seg-a", "customerAsn": 65010}}}
}

// erTunnelPatch is a patch that sets the tunnels of the stored gateway gw1.
func erTunnelPatch(tunnels string) map[string]any {
	var p map[string]any
	json.Unmarshal([]byte(`{"instances":[{"id":801,"segmentOptions":[{"segmentName":"seg-a","customerGateways":[{"name":"gw1","tunnels":[`+tunnels+`]}]}]}]}`), &p)
	return p
}

// MA-W3, CG-P33.6: no tool schema names the pre-shared key; a tunnel patch
// resends the stored key, the output masks it, and a patch that drops a
// keyed tunnel is refused with no write.
func TestConnectorAzureExpressRoute_Secrets(t *testing.T) {
	s := specNamed("connector_azure_express_route")
	e := newFamilyEnv(t, s)
	var res struct {
		Tools []struct {
			Name        string          `json:"name"`
			InputSchema json.RawMessage `json:"inputSchema"`
		} `json:"tools"`
	}
	json.Unmarshal(call(t, e.srv, "tools/list", nil), &res)
	n := 0
	for _, tool := range res.Tools {
		if strings.HasPrefix(tool.Name, "connector_azure_express_route_") {
			n++
			if strings.Contains(string(tool.InputSchema), "preSharedKey") {
				t.Errorf("%s schema names preSharedKey", tool.Name)
			}
		}
	}
	if n != 3 {
		t.Fatalf("%d tools", n)
	}
	s.put(e, 7, s.stored())
	r := e.call("connector_azure_express_route_update", map[string]any{"id": "7", "patch": erTunnelPatch(`{"id":"tun-er-1","initiator":true},{"id":"tun-er-2"},{"name":"t3"}`)})
	mustOK(t, r)
	put := sentBody(t, e, "PUT")
	tunnels := put["instances"].([]any)[0].(map[string]any)["segmentOptions"].([]any)[0].(map[string]any)["customerGateways"].([]any)[0].(map[string]any)["tunnels"]
	if fmtJSON(tunnels) != `[{"id":"tun-er-1","initiator":true,"name":"t1","preSharedKey":"stored-psk"},{"id":"tun-er-2","name":"t2"},{"name":"t3"}]` {
		t.Fatalf("PUT tunnels %s", fmtJSON(tunnels))
	}
	if strings.Contains(r.Text, "stored-psk") {
		t.Fatalf("output shows the key:\n%s", r.Text)
	}
	e = newFamilyEnv(t, s)
	s.put(e, 7, s.stored())
	r = e.call("connector_azure_express_route_update", map[string]any{"id": "7", "patch": erTunnelPatch(`{"id":"tun-er-2"}`)})
	if !r.IsError || !strings.Contains(r.Text, "stored secret") || strings.Contains(r.Text, "stored-psk") || e.portal.count("PUT") != 0 {
		t.Fatalf("drop keyed tunnel: PUT=%d %s", e.portal.count("PUT"), r.Text)
	}
}

// MA-W5, MA-W16, MA-W26, MA-W41: the D13 guard reads the GET /connectors
// row, a connector marked for deletion and an id of another connector type
// are refused, and an instances patch must keep every stored instance id;
// none of them writes.
func TestConnectorAzureExpressRoute_Guards(t *testing.T) {
	s := specNamed("connector_azure_express_route")
	for _, c := range []struct {
		row, network string
		ok           bool
	}{{"FAILED", "", false}, {"IN_PROGRESS", "", false}, {"SUCCESS", "FAILED", false}, {"PENDING", "FAILED", true}, {"SUCCESS", "", true}} {
		e := newFamilyEnv(t, s)
		e.portal.networkState = c.network
		o := s.stored()
		o["state"] = c.row
		s.put(e, 7, o)
		r := e.call(s.name+"_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}})
		if c.ok != !r.IsError || c.ok != (e.portal.count("PUT") == 1) {
			t.Errorf("row %s network %s: %s", c.row, c.network, r.Text)
		}
	}

	e := newFamilyEnv(t, s)
	e.portal.put(8, s.stored())
	e.portal.connectors = []map[string]any{{"id": 8, "name": "er7", "type": "AZURE_VNET", "state": "PENDING"}}
	mustRefuse(t, e.call(s.name+"_update", map[string]any{"id": "8", "patch": map[string]any{"description": "x"}}), "family_mismatch", "belongs to another family")
	mustRefuse(t, e.call(s.name+"_delete", map[string]any{"id": "8"}), "family_mismatch", "belongs to another family")

	e = newFamilyEnv(t, s)
	e.portal.fail("GET", e.portal.base()+"/7", fault{status: 404, body: `{"code":"AK000003","message":"Could not find connector with id : 7"}`})
	o := s.stored()
	o["state"] = "SUCCESS"
	s.put(e, 7, o)
	mustRefuse(t, e.call(s.name+"_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}), "state_refused", "deletion pending provisioning")

	e = newFamilyEnv(t, s)
	s.put(e, 7, s.stored())
	r := e.call(s.name+"_update", map[string]any{"id": "7", "patch": map[string]any{"instances": []any{map[string]any{
		"name": "i2", "expressRouteCircuitId": "c2", "credentialId": "cred-az", "loopbackSubnet": "10.251.0.64/26"}}}})
	if !r.IsError || !strings.Contains(r.Text, "id 801 is missing") || e.portal.count("PUT") != 0 {
		t.Fatalf("instance id: %s", r.Text)
	}
}
