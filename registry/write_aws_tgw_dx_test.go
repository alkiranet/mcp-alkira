package registry

import (
	"encoding/json"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/alkiranet/mcp-alkira/writes"
)

func init() {
	familySpecs = append(familySpecs, awsTgwDxSpecs...)
	for name, k := range map[string]string{"connector_aws_tgw": "K31", "connector_aws_direct_connect": "K32"} {
		maps.Copy(scenarioRows, map[string]map[string]string{name: {
			// Live K rows run .6 only: a save needs an ACTIVE TGW peering attachment or a real Direct Connect connection.
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode; " + k + ".6 (negative only)",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go)",
			scImportNonexistent: "TestFamilies_NotFound TestAwsTgwDx_FamilyCheck; " + k + ".6",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03); TestFamilies_InvalidArgumentsSendNothing",
			scPreflightBlocked:  "TestAwsTgwDx_StateGuard TestAwsTgwDx_FamilyCheck; MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
			scSoftDeleteOff:     "TestAwsTgwDx_DocState TestFamilies_DeletePreview",
		}})
	}
}

var connectorNotFound = fault{status: 404, body: `{"code":"AK000003","message":"Could not find connector with id : 404"}`}

// Stored bodies follow the TPS GET shapes (model/awstgw/AWSTGWConnector.java,
// model/DirectConnectConnector.java).
var awsTgwDxSpecs = []familySpec{
	{
		name: "connector_aws_tgw", noun: "AWS TGW connector", f: family(writes.ConnectorAwsTgw), coll: "aws-tgw-connectors", connector: "AWS_TGW",
		notFound: connectorNotFound,
		stored:   awsTgwStored,
		create:   awsTgwBody,
		readOnly: []string{"implicitGroupId", "segmentIds", "awsTgwId", "transitGatewaySegmentGroupId", "primary", "staticRouteValues", "previousState"},
		creates: []createCase{
			{nil, awsTgwBody("tgw1"), withFields(awsTgwBody("tgw1"), map[string]any{"enabled": true}), nil},
			{[]string{"BILLING-TAGS"}, withFields(awsTgwBody("tgw1"), map[string]any{"billingTags": []any{3}, "enabled": false}),
				withFields(awsTgwBody("tgw1"), map[string]any{"billingTags": []any{3}, "enabled": false}), []string{"BILLING-TAGS"}},
			{[]string{"RELEASE-SCALE-GROUP-FOR-TGW-CONNECTOR"}, withFields(awsTgwBody("tgw1"), map[string]any{"scaleGroupId": "sg-1"}),
				withFields(awsTgwBody("tgw1"), map[string]any{"scaleGroupId": "sg-1", "enabled": true}), []string{"RELEASE-SCALE-GROUP-FOR-TGW-CONNECTOR"}},
		},
		invalid: map[string]toolCase{
			// TPS accepts a description with a trailing space (AWSTGWConnector.java:69-71).
			"trailing space": {},
			// The AWS TGW connector name holds letters, digits, - and _ only (AWSTGWConnector.java:64-67).
			"charset":       {"_create", map[string]any{"body": awsTgwBody("a.b")}, "outside the slug character set"},
			"leading space": {"_create", map[string]any{"body": awsTgwBody(" a")}, "outside the slug character set"},
			"ALK- prefix":   {"_create", map[string]any{"body": awsTgwBody("ALK-x")}, "name must not start with ALK-"},
			"two segments":  {"_create", map[string]any{"body": withFields(awsTgwBody("t1"), map[string]any{"segments": []any{"a", "b"}})}, "segments must hold exactly one segment name"},
			"no attachment": {"_create", map[string]any{"body": withFields(awsTgwBody("t1"), map[string]any{"awsTgwPeeringAttachmentId": nil})}, "awsTgwPeeringAttachmentId is required"},
			"size":          {"_create", map[string]any{"body": withFields(awsTgwBody("t1"), map[string]any{"size": "30LARGE"})}, "size must be one of"},
			"awsTgwId":      {"_update", map[string]any{"id": "7", "patch": map[string]any{"awsTgwId": 2}}, "awsTgwId is read-only"},
		},
		inUse: []inUseCase{
			{fault{status: 409, body: `{"code":"AK000001","message":"CONNECTOR with id 7 is still referred externally"}`}, "referred externally"},
		},
		previews: []previewCase{
			{"PENDING", []string{`Delete preview for AWS TGW connector "tgw7" (id 7).`, "Deployed: no, never deployed (state PENDING).", `Also deleted with it: "tgw7".`, "The delete now: removes it at once."}},
			{"SUCCESS", []string{"Deployed: yes (state SUCCESS).", `Also deleted with it: "tgw7".`, "The delete now: marks it for deletion; it keeps carrying traffic, with its name and references, until the deletion is provisioned from the Alkira portal."}},
		},
	},
	{
		name: "connector_aws_direct_connect", noun: "AWS Direct Connect connector", f: family(writes.ConnectorAwsDirectConnect), coll: "directconnectconnectors", connector: "DIRECT_CONNECT",
		notFound: connectorNotFound,
		stored:   dxStored,
		create:   dxBody,
		readOnly: []string{"implicitGroupId", "lastEnabled", "previousState"},
		creates: []createCase{
			{nil, dxBody("dx1"), withFields(dxBody("dx1"), map[string]any{"enabled": true}), nil},
			{[]string{"BILLING-TAGS"}, withFields(dxBody("dx1"), map[string]any{"billingTags": []any{3}}),
				withFields(dxBody("dx1"), map[string]any{"billingTags": []any{3}, "enabled": true}), []string{"BILLING-TAGS"}},
			{[]string{"SUPPORT-SCALE-GROUP-FOR-DX-CONNECTOR"}, withFields(dxBody("dx1"), map[string]any{"scaleGroupId": "sg-1"}),
				withFields(dxBody("dx1"), map[string]any{"scaleGroupId": "sg-1", "enabled": true}), []string{"SUPPORT-SCALE-GROUP-FOR-DX-CONNECTOR"}},
			{[]string{"RELEASE-CUSTOM-SCALE-OPTIONS"}, withFields(dxBody("dx1"), map[string]any{"loopbackPrefixes": []any{"100.64.0.0/26"}}),
				withFields(dxBody("dx1"), map[string]any{"loopbackPrefixes": []any{"100.64.0.0/26"}, "enabled": true}), []string{"RELEASE-CUSTOM-SCALE-OPTIONS"}},
		},
		invalid: map[string]toolCase{
			"trailing space": {},
			"ALK- prefix":    {"_create", map[string]any{"body": dxBody("ALK-x")}, "name must not start with ALK-"},
			"no size":        {"_create", map[string]any{"body": withFields(dxBody("d1"), map[string]any{"size": nil})}, "size is required"},
			"bgp key":        {"_create", map[string]any{"body": withFields(dxBody("d1"), map[string]any{"instances": []any{withFields(dxInstance("i1"), map[string]any{"bgpAuthKey": "k123456"})}})}, "bgpAuthKey"},
			"VXLAN":          {"_create", map[string]any{"body": withFields(dxBody("d1"), map[string]any{"tunnelProtocol": "VXLAN"})}, "vni must be set when tunnelProtocol is VXLAN"},
			"key in patch":   {"_update", map[string]any{"id": "7", "patch": map[string]any{"instances": []any{map[string]any{"id": 801, "bgpAuthKey": "k123456"}}}}, "bgpAuthKey is a secret"},
			"connectionName": {"_update", map[string]any{"id": "7", "patch": map[string]any{"instances": []any{map[string]any{"id": 801, "connectionName": "x"}}}}, "connectionName is read-only"},
		},
		inUse: []inUseCase{
			{fault{status: 409, body: `{"code":"AK000001","message":"CONNECTOR with id 7 is still referred externally"}`}, "referred externally"},
		},
		previews: []previewCase{
			{"PENDING", []string{`Delete preview for AWS Direct Connect connector "dx7" (id 7).`, "Deployed: no, never deployed (state PENDING).", `Also deleted with it: "dx7".`}},
		},
	},
}

func awsTgwStored() map[string]any {
	return map[string]any{"name": "tgw7", "description": "seven", "cxp": "USEAST-AZURE-1", "implicitGroupId": 61201, "segments": []any{"seg-a"},
		"segmentIds": []any{1207}, "size": "SMALL", "awsTgwPeeringAttachmentId": 1687, "awsTgwId": 1123, "billingTags": []any{}, "staticRoutes": []any{},
		"enabled": true, "transitGatewaySegmentGroupId": "5b8f3c1e-7d2a-4e7b-9c1d-2a6f0e4b8c11", "primary": true, "staticRouteValues": []any{}}
}

func awsTgwBody(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "USEAST-AZURE-1", "segments": []any{"seg-a"}, "awsTgwPeeringAttachmentId": 1687}
}

func dxSegmentOption() map[string]any {
	return map[string]any{"segmentName": "seg-a", "customerAsn": 65010, "customerLoopbackIp": "10.9.0.10", "alkLoopbackIp1": "10.9.0.11",
		"alkLoopbackIp2": "10.9.0.12", "loopbackSubnet": "10.9.0.0/24"}
}

func dxInstance(name string) map[string]any {
	return map[string]any{"name": name, "connectionId": "dxcon-" + name, "dcGatewayAsn": 64900, "underlayAsn": 65001, "customerRegion": "us-east-1",
		"credentialId": "cred-aws", "segmentOptions": []any{dxSegmentOption()}}
}

func dxBody(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "USEAST-AZURE-1", "size": "SMALL", "instances": []any{dxInstance("i1")}}
}

func dxStoredInstance(id int, name, key string) map[string]any {
	so := withFields(dxSegmentOption(), map[string]any{"disableInternetExit": false, "advertiseOnPremRoutes": false, "generatedTunnelLoopbackIps": []any{}})
	return withFields(dxInstance(name), map[string]any{"id": id, "connectionName": "conn-" + name, "vlan": 101,
		"bgpAuthKey": key + "-1", "bgpAuthKeyAlkira": key + "-2", "segmentOptions": []any{so}})
}

func dxStored() map[string]any {
	return map[string]any{"name": "dx7", "description": "seven", "cxp": "USEAST-AZURE-1", "implicitGroupId": 61301, "size": "SMALL",
		"instances": []any{dxStoredInstance(801, "i1", "stored-dx")}, "billingTags": []any{}, "lastEnabled": 1763600000, "enabled": true, "tunnelProtocol": "GRE"}
}

// MA-W5: the D13 guard reads the GET /connectors row state of each family.
func TestAwsTgwDx_StateGuard(t *testing.T) {
	for _, s := range awsTgwDxSpecs {
		for _, c := range []struct {
			row, network string
			ok           bool
		}{{"FAILED", "", false}, {"SCHEDULED", "", false}, {"IN_PROGRESS", "", false}, {"SUCCESS", "FAILED", false}, {"PENDING", "FAILED", true}, {"SUCCESS", "", true}} {
			e := newFamilyEnv(t, s)
			e.portal.networkState = c.network
			o := s.stored()
			o["state"] = c.row
			s.put(e, 7, o)
			r := e.call(s.name+"_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}})
			if c.ok != !r.IsError || c.ok != (e.portal.count("PUT") == 1) {
				t.Errorf("%s row %s network %s: %s", s.name, c.row, c.network, r.Text)
			}
		}
	}
}

// MA-W26, MA-W41, BC-R2: an id whose GET /connectors row is another connector
// type is refused with no write, although its body has the family's shape.
func TestAwsTgwDx_FamilyCheck(t *testing.T) {
	for _, s := range awsTgwDxSpecs {
		e := newFamilyEnv(t, s)
		e.portal.put(8, s.stored())
		e.portal.connectors = []map[string]any{{"id": 8, "name": s.stored()["name"], "type": "AWS_VPC", "state": "PENDING"}}
		mustRefuse(t, e.call(s.name+"_update", map[string]any{"id": "8", "patch": map[string]any{"description": "x"}}), "family_mismatch", "belongs to another family")
		mustRefuse(t, e.call(s.name+"_delete", map[string]any{"id": "8"}), "family_mismatch", "belongs to another family")
		if e.portal.count("PUT")+e.portal.count("DELETE") != 0 {
			t.Fatalf("%s: write sent", s.name)
		}
	}
}

// MA-W16: a connector marked for deletion is refused before any write.
func TestAwsTgwDx_DocState(t *testing.T) {
	for _, s := range awsTgwDxSpecs {
		e := newFamilyEnv(t, s)
		e.portal.fail("GET", e.portal.base()+"/7", fault{status: 404, body: `{"code":"AK000003","message":"Could not find connector with id : 7"}`})
		o := s.stored()
		o["state"] = "SUCCESS"
		s.put(e, 7, o)
		mustRefuse(t, e.call(s.name+"_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}), "state_refused", "deletion pending provisioning")
		if e.portal.count("PUT") != 0 {
			t.Fatalf("%s: PUT sent", s.name)
		}
	}
}

// MA-W11: a connector of another type or a group holding the name refuses the create with no POST.
func TestAwsTgwDx_DuplicateGuard(t *testing.T) {
	for _, s := range awsTgwDxSpecs {
		e := newFamilyEnv(t, s)
		e.portal.connectors = []map[string]any{{"id": 31, "name": "dup", "type": "IP_SEC", "state": "SUCCESS"}}
		e.portal.put(31, map[string]any{"name": "dup", "sites": []any{}})
		mustRefuse(t, e.call(s.name+"_create", map[string]any{"body": s.createBody("dup")}), "duplicate",
			`an ip_sec connector named "dup" already exists (id 31); connector names are unique across every connector type and group.`)
		if e.portal.count("POST") != 0 {
			t.Fatalf("%s: POST sent", s.name)
		}
	}
}

// The attachment an AWS TGW connector uses cannot change after create
// (AWSTGWConnectorServiceImpl.java:333-337); a patch that keeps it is sent.
func TestAwsTgw_AttachmentImmutable(t *testing.T) {
	s := specNamed("connector_aws_tgw")
	e := newFamilyEnv(t, s)
	s.put(e, 7, s.stored())
	mustRefuse(t, e.call(s.name+"_update", map[string]any{"id": "7", "patch": map[string]any{"awsTgwPeeringAttachmentId": 1688}}),
		"invalid_arguments", "awsTgwPeeringAttachmentId cannot be changed")
	if e.portal.count("PUT") != 0 {
		t.Fatal("PUT sent")
	}
	mustOK(t, e.call(s.name+"_update", map[string]any{"id": "7", "patch": map[string]any{"awsTgwPeeringAttachmentId": 1687, "staticRoutes": []any{11}}}))
	if put := sentBody(t, e, "PUT"); put["awsTgwPeeringAttachmentId"] != float64(1687) || len(put["staticRoutes"].([]any)) != 1 {
		t.Fatalf("PUT %v", put)
	}
}

// MA-W15, CG-P32.6: the Direct Connect schemas hold no BGP key; an update
// resends the stored keys of the instances it keeps, the output masks them,
// remove.instances drops an instance with its keys, and a patch that leaves
// a stored instance out is refused.
func TestAwsDirectConnect_Secrets(t *testing.T) {
	s := specNamed("connector_aws_direct_connect")
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
		if strings.HasPrefix(tool.Name, "connector_aws_direct_connect_") {
			n++
			if strings.Contains(string(tool.InputSchema), "bgpAuthKey") {
				t.Errorf("%s schema names a BGP key", tool.Name)
			}
		}
	}
	if n != 3 {
		t.Fatalf("%d tools", n)
	}
	o := s.stored()
	o["instances"] = append(o["instances"].([]any), dxStoredInstance(802, "i2", "other-dx"))
	s.put(e, 7, o)
	r := e.call(s.name+"_update", map[string]any{"id": "7", "patch": map[string]any{"instances": []any{map[string]any{"id": 801, "vlan": 102}, map[string]any{"id": 802}}}})
	mustOK(t, r)
	put := sentBody(t, e, "PUT")
	in := put["instances"].([]any)[0].(map[string]any)
	if in["vlan"] != float64(102) || in["bgpAuthKey"] != "stored-dx-1" || in["bgpAuthKeyAlkira"] != "stored-dx-2" || in["connectionName"] != nil {
		t.Fatalf("PUT instance %v", in)
	}
	if strings.Contains(r.Text, "stored-dx") {
		t.Fatalf("output shows a key:\n%s", r.Text)
	}
	e = newFamilyEnv(t, s)
	s.put(e, 7, o)
	mustOK(t, e.call(s.name+"_update", map[string]any{"id": "7", "remove": map[string]any{"instances": []any{"802"}}}))
	if got := sentBody(t, e, "PUT")["instances"].([]any); len(got) != 1 || got[0].(map[string]any)["bgpAuthKey"] != "stored-dx-1" {
		t.Fatalf("PUT instances %v", got)
	}
	e = newFamilyEnv(t, s)
	s.put(e, 7, o)
	r = e.call(s.name+"_update", map[string]any{"id": "7", "patch": map[string]any{"instances": []any{map[string]any{"id": 801}}}})
	if !r.IsError || !strings.Contains(r.Text, "instances must carry the id of every stored element; id 802 is missing") || e.portal.count("PUT") != 0 {
		t.Fatalf("instance dropped by a patch: %s", r.Text)
	}
}

// The peering gateway reads that name the attachment an AWS TGW connector uses.
func TestRegister_PeeringGatewayAwsTgwGetById_HTTP(t *testing.T) {
	client, seen := fakePortal(t, map[string]string{
		"/api/tenantnetworks/7/aws-tgws/1123":                    `{"id":1123,"name":"tgw-a","cxp":"USEAST-AZURE-1","asn":64512,"state":"ACTIVE"}`,
		"/api/tenantnetworks/7/aws-tgw-peering-attachments/1687": `{"id":1687,"name":"att-a","awsTgwId":1123,"state":"ACTIVE"}`,
	})
	srv := newServer(providerFor(client))
	for tool, args := range map[string]map[string]any{
		"peering_gateway_aws_tgw_get_by_id":            {"peeringGatewayId": "1123"},
		"peering_gateway_aws_tgw_attachment_get_by_id": {"attachmentId": "1687"},
	} {
		res := callTool(t, srv, tool, args)
		if res.IsError || !strings.Contains(res.Content[0].Text, `"state":"ACTIVE"`) {
			t.Fatalf("%s: %+v", tool, res.Content)
		}
	}
	paths := seen.get()
	slices.Sort(paths)
	if strings.Join(paths, ",") != "/api/tenantnetworks/7/aws-tgw-peering-attachments/1687,/api/tenantnetworks/7/aws-tgws/1123" {
		t.Fatalf("requests = %v", paths)
	}
}
