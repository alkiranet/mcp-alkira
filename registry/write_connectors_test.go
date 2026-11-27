package registry

import (
	"encoding/json"
	"maps"
	"strings"
	"testing"

	"github.com/alkiranet/mcp-alkira/writes"
)

func init() {
	familySpecs = append(familySpecs, connectorSpecs...)
	rows := func(k string) map[string]string {
		return map[string]string{
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode; " + k + ".1-.5",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go), H-W16 (ST-04)",
			scImportNonexistent: "TestFamilies_NotFound TestConnectorFamilies_FamilyCheck; " + k + ".6",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03)",
			scPreflightBlocked:  "TestConnectorFamilies_StateGuard TestConnectorFamilies_FamilyCheck; MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
			scSoftDeleteOff:     "TestConnectorFamilies_DocState TestFamilies_DeletePreview",
		}
	}
	ipsec := rows("K23")
	ipsec[scCRUD] = "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode TestConnectorIPsec_CreateKeys; K23.1-.5"
	maps.Copy(scenarioRows, map[string]map[string]string{
		"connector_ipsec":      ipsec,
		"connector_aws_vpc":    rows("K27"),
		"connector_azure_vnet": rows("K28"),
	})
}

// Stored bodies follow the TPS GET shapes (model/IPsecConnector.java,
// AWSVPCConnector.java, AzureVNETConnector.java).
var connectorSpecs = []familySpec{
	{
		name: "connector_ipsec", noun: "IPsec connector", f: family(writes.ConnectorIPsec), coll: "ipsecconnectors", connector: "IP_SEC",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"Could not find connector with id : 404"}`},
		stored:   ipsecStored,
		creates: []createCase{
			{nil, ipsecBody("ipsec1"), ipsecBody("ipsec1"), nil},
		},
		readOnly: []string{"implicitGroupId", "primary", "primaryConnectorId", "type", "previousState"},
		invalid: map[string]toolCase{
			"trailing space": {},
			"key in patch":   {"_update", map[string]any{"id": "7", "patch": map[string]any{"sites": []any{map[string]any{"id": 500, "presharedKeys": []any{"k"}}}}}, "presharedKeys"},
			"bgp key":        {"_update", map[string]any{"id": "7", "patch": map[string]any{"routingOptions": map[string]any{"dynamicRouting": map[string]any{"bgpAuthKeyAlkira": "k"}}}}, "bgpAuthKeyAlkira"},
			"vpn mode":       {"_update", map[string]any{"id": "7", "patch": map[string]any{"vpnMode": "OTHER"}}, "vpnMode must be one of ROUTE_BASED, POLICY_BASED"},
		},
		inUse: []inUseCase{
			{fault{status: 400, body: `{"code":"AK000002","message":"connector with name 'ipsec7' is associated with nat policy 'n'"}`}, "nat policy"},
		},
		previews: []previewCase{
			{"SUCCESS", []string{`Delete preview for IPsec connector "ipsec7" (id 7).`, "Deployed: yes (state SUCCESS).", `Also deleted with it: "ipsec7".`,
				"The delete now: marks it for deletion; it keeps carrying traffic, with its name and references, until the deletion is provisioned from the Alkira portal."}},
		},
	},
	{
		name: "connector_aws_vpc", noun: "AWS VPC connector", f: family(writes.ConnectorAwsVpc), coll: "awsvpcconnectors", connector: "AWS_VPC",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"Could not find connector with id : 404"}`},
		stored:   awsVpcStored,
		create:   awsVpcBody,
		readOnly: []string{"implicitGroupId", "vgwId", "primary", "primaryConnectorId", "lastEnabled", "previousState"},
		creates: []createCase{
			{nil, awsVpcBody("aws1"), awsVpcBody("aws1"), nil},
			{[]string{"CXP-FAILOVER-REGION"}, withFields(awsVpcBody("aws1"), map[string]any{"enabled": false}), withFields(awsVpcBody("aws1"), map[string]any{"enabled": false}), []string{"CXP-FAILOVER-REGION"}},
		},
		invalid: map[string]toolCase{
			"trailing space": {},
			// The AWS VPC connector name allows "/" (AWSVPCConnector.java:61).
			"charset":     {},
			"ALK- prefix": {"_create", map[string]any{"body": awsVpcBody("ALK-x")}, "name must not start with ALK-"},
			"size":        {"_create", map[string]any{"body": withFields(awsVpcBody("aws1"), map[string]any{"size": "4LARGE"})}, "size must be one of"},
			"vgwId":       {"_update", map[string]any{"id": "7", "patch": map[string]any{"vgwId": "vgw-2"}}, "vgwId is read-only"},
		},
		inUse: []inUseCase{
			{fault{status: 409, body: `{"code":"AK000001","message":"CONNECTOR with id 7 is still referred externally"}`}, "referred externally"},
		},
		previews: []previewCase{
			{"PENDING", []string{"Deployed: no, never deployed (state PENDING).", `Also deleted with it: "aws7".`, "The delete now: removes it at once."}},
		},
	},
	{
		name: "connector_azure_vnet", noun: "Azure VNet connector", f: family(writes.ConnectorAzureVnet), coll: "azurevnetconnectors", connector: "AZURE_VNET",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"Could not find connector with id : 404"}`},
		stored:   azureVnetStored,
		create:   azureVnetBody,
		readOnly: []string{"implicitGroupId", "customerRegion", "routingMode", "subscriptionId", "primary", "primaryConnectorId", "lastEnabled", "previousState"},
		creates: []createCase{
			{nil, azureVnetBody("az1"), withFields(azureVnetBody("az1"), map[string]any{"enabled": true, "connectionMode": "VNET_GATEWAY"}), nil},
		},
		invalid: map[string]toolCase{
			"trailing space": {},
			"other rg":       {"_create", map[string]any{"body": withFields(azureVnetBody("az1"), map[string]any{"resourceGroupName": "rg-9"})}, "resourceGroupName"},
			"vnetId":         {"_update", map[string]any{"id": "7", "patch": map[string]any{"subscriptionId": "s"}}, "subscriptionId is read-only"},
		},
		previews: []previewCase{
			{"PENDING", []string{"Deployed: no, never deployed (state PENDING).", `Also deleted with it: "az7".`}},
		},
	},
}

func ipsecStored() map[string]any {
	return map[string]any{"name": "ipsec7", "description": "seven", "segments": []any{"seg-a"}, "cxp": "USEAST-AZURE-1", "implicitGroupId": 61201, "size": "SMALL",
		"sites": []any{map[string]any{"id": 500, "name": "s1", "connectorId": 7, "gatewayIpType": "STATIC", "customerGwIp": "203.0.113.1",
			"presharedKeys": []any{"stored-k1", "stored-k2"}, "state": "PENDING", "billingTags": []any{}}},
		"type": "IP_SEC", "vpnMode": "ROUTE_BASED", "routingOptions": map[string]any{"staticRouting": map[string]any{"prefixListId": 11, "availability": "IPSEC_INTERFACE_PING"}},
		"enabled": true, "primary": true, "secondaryCXPs": []any{}, "primaryConnectorId": 7,
		"segmentOptions": map[string]any{"seg-a": map[string]any{"disableInternetExit": false, "advertiseOnPremRoutes": false}}}
}

func awsVpcStored() map[string]any {
	return map[string]any{"name": "aws7", "description": "seven", "cxp": "USEAST-AZURE-1", "segments": []any{"seg-a"}, "vpcId": "vpc-1",
		"vpcOwnerId": "111122223333", "customerName": "aws7", "customerRegion": "us-east-1", "credentialId": "cred-aws", "implicitGroupId": 61001,
		"vgwId": "vgw-1", "billingTags": []any{}, "size": "SMALL", "enabled": true, "primary": true, "secondaryCXPs": []any{}, "primaryConnectorId": 7,
		"lastEnabled": 1763600000, "directInterVPCCommunicationEnabled": false, "tgwConnectEnabled": false}
}

func awsVpcBody(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "USEAST-AZURE-1", "segments": []any{"seg-a"}, "vpcId": "vpc-1", "customerRegion": "us-east-1", "credentialId": "cred-aws"}
}

const azureVnetIDTest = "/subscriptions/sub-1/resourceGroups/rg-1/providers/Microsoft.Network/virtualNetworks/vnet-1"

func azureVnetStored() map[string]any {
	return map[string]any{"name": "az7", "description": "seven", "implicitGroupId": 61101, "vnetId": azureVnetIDTest, "cxp": "USEAST-AZURE-1",
		"customerRegion": "eastus", "credentialId": "cred-az", "resourceGroupName": "rg-1", "enabled": true, "primary": true, "secondaryCXPs": []any{},
		"lastEnabled": 1763600000, "primaryConnectorId": 7, "connectionMode": "VNET_PEERING", "routingMode": "GLOBAL", "customerAsn": 64521,
		"segments": []any{"seg-a"}, "billingTags": []any{}, "size": "SMALL", "subscriptionId": "sub-1",
		"vnetRouting": map[string]any{"exportToCXPOptions": map[string]any{"userInputPrefixes": []any{}},
			"importFromCXPOptions": map[string]any{"prefixListIds": []any{}, "routeImportMode": "ADVERTISE_DEFAULT_ROUTE"}}}
}

func azureVnetBody(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "USEAST-AZURE-1", "segments": []any{"seg-a"}, "vnetId": azureVnetIDTest, "resourceGroupName": "rg-1", "credentialId": "cred-az"}
}

func ipsecBody(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "USEAST-AZURE-1", "segments": []any{"seg-a"},
		"sites":          []any{map[string]any{"name": "s1", "customerGwIp": "203.0.113.1"}},
		"routingOptions": map[string]any{"staticRouting": map[string]any{"prefixListId": 11}}}
}

// D20: connector_ipsec_create takes pre-shared keys and the BGP key as
// optional write-only fields: sent as given, left out when not given, never
// shown in output. The update schema holds no key field, and an update
// resends the stored keys while the output masks them.
func TestConnectorIPsec_CreateKeys(t *testing.T) {
	s := specNamed("connector_ipsec")
	e := newFamilyEnv(t, s)
	var res struct {
		Tools []struct {
			Name        string          `json:"name"`
			InputSchema json.RawMessage `json:"inputSchema"`
		} `json:"tools"`
	}
	json.Unmarshal(call(t, e.srv, "tools/list", nil), &res)
	for _, tool := range res.Tools {
		keys := strings.Contains(string(tool.InputSchema), "presharedKeys") && strings.Contains(string(tool.InputSchema), "bgpAuthKeyAlkira")
		switch tool.Name {
		case "connector_ipsec_create":
			if !keys {
				t.Errorf("create schema lacks the key fields: %s", tool.InputSchema)
			}
		case "connector_ipsec_update", "connector_ipsec_delete":
			if strings.Contains(string(tool.InputSchema), "presharedKeys") || strings.Contains(string(tool.InputSchema), "bgpAuthKeyAlkira") {
				t.Errorf("%s schema names a key field", tool.Name)
			}
		}
	}

	body := ipsecBody("ipsec1")
	body["sites"].([]any)[0].(map[string]any)["presharedKeys"] = []any{"caller-k1", "caller-k2"}
	body["routingOptions"] = map[string]any{"dynamicRouting": map[string]any{"customerGwAsn": "65001", "bgpAuthKeyAlkira": "caller-bgp"}}
	r := e.call("connector_ipsec_create", map[string]any{"body": body})
	mustOK(t, r)
	post := sentBody(t, e, "POST")
	site := post["sites"].([]any)[0].(map[string]any)
	if fmtJSON(site["presharedKeys"]) != `["caller-k1","caller-k2"]` || fmtJSON(post["routingOptions"]) != `{"dynamicRouting":{"bgpAuthKeyAlkira":"caller-bgp","customerGwAsn":"65001"}}` {
		t.Fatalf("POST %v", post)
	}
	if strings.Contains(r.Text, "caller-k1") || strings.Contains(r.Text, "caller-bgp") {
		t.Fatalf("output shows a key:\n%s", r.Text)
	}

	e = newFamilyEnv(t, s)
	mustOK(t, e.call("connector_ipsec_create", map[string]any{"body": ipsecBody("ipsec2")}))
	if site := sentBody(t, e, "POST")["sites"].([]any)[0].(map[string]any); site["presharedKeys"] != nil {
		t.Fatalf("keys sent without being given: %v", site)
	}

	s.put(e, 7, s.stored())
	r = e.call("connector_ipsec_update", map[string]any{"id": "7", "patch": map[string]any{"sites": []any{map[string]any{"id": 500, "customerGwIp": "203.0.113.9"}}}})
	mustOK(t, r)
	put := sentBody(t, e, "PUT")
	site = put["sites"].([]any)[0].(map[string]any)
	if site["customerGwIp"] != "203.0.113.9" || fmtJSON(site["presharedKeys"]) != `["stored-k1","stored-k2"]` {
		t.Fatalf("PUT site %v", site)
	}
	if strings.Contains(r.Text, "stored-k1") {
		t.Fatalf("output shows a key:\n%s", r.Text)
	}
}

// ipsecBGPStored is ipsecStored with BGP routing and a stored BGP key.
func ipsecBGPStored() map[string]any {
	o := ipsecStored()
	o["routingOptions"] = map[string]any{
		"staticRouting":  map[string]any{"prefixListId": 11, "availability": "IPSEC_INTERFACE_PING"},
		"dynamicRouting": map[string]any{"customerGwAsn": "65001", "bgpAuthKeyAlkira": "stored-bgp", "availability": "IPSEC_INTERFACE_PING"},
	}
	return o
}

// With a BGP key stored, connector_ipsec_update switches to POLICY_BASED
// when the patch sets vpnMode and policyOptions together, and moves to
// static routing when it nulls dynamicRouting; both drop the key. A patch
// whose options side does not match vpnMode is refused with no write.
func TestConnectorIPsec_ModeSwitchWithBGPKey(t *testing.T) {
	s := specNamed("connector_ipsec")
	for name, c := range map[string]struct {
		patch map[string]any
		check func(put map[string]any) bool
	}{
		"route to policy": {
			map[string]any{"vpnMode": "POLICY_BASED", "policyOptions": map[string]any{"branchTSPrefixListIds": []any{1}, "cxpTSPrefixListIds": []any{2}}},
			func(put map[string]any) bool {
				return put["vpnMode"] == "POLICY_BASED" && put["routingOptions"] == nil && fmtJSON(put["policyOptions"]) == `{"branchTSPrefixListIds":[1],"cxpTSPrefixListIds":[2]}`
			},
		},
		"dynamic to static": {
			map[string]any{"routingOptions": map[string]any{"dynamicRouting": nil}},
			func(put map[string]any) bool {
				return fmtJSON(put["routingOptions"]) == `{"staticRouting":{"availability":"IPSEC_INTERFACE_PING","prefixListId":11}}`
			},
		},
		"static change keeps the key": {
			map[string]any{"routingOptions": map[string]any{"staticRouting": map[string]any{"prefixListId": 12}}},
			func(put map[string]any) bool {
				ro, _ := put["routingOptions"].(map[string]any)
				dr, _ := ro["dynamicRouting"].(map[string]any)
				return dr["bgpAuthKeyAlkira"] == "stored-bgp"
			},
		},
	} {
		e := newFamilyEnv(t, s)
		s.put(e, 7, ipsecBGPStored())
		r := e.call("connector_ipsec_update", map[string]any{"id": "7", "patch": c.patch})
		mustOK(t, r)
		if put := sentBody(t, e, "PUT"); !c.check(put) {
			t.Errorf("%s: PUT %s", name, fmtJSON(put))
		}
		if strings.Contains(r.Text, "stored-bgp") {
			t.Errorf("%s: output shows the BGP key:\n%s", name, r.Text)
		}
	}
	for name, c := range map[string]struct {
		patch map[string]any
		text  string
	}{
		"policy without vpnMode": {map[string]any{"policyOptions": map[string]any{"branchTSPrefixListIds": []any{1}, "cxpTSPrefixListIds": []any{2}}}, "vpnMode is ROUTE_BASED"},
		"vpnMode alone":          {map[string]any{"vpnMode": "POLICY_BASED"}, "policyOptions must be set when vpnMode is POLICY_BASED"},
	} {
		e := newFamilyEnv(t, s)
		s.put(e, 7, ipsecBGPStored())
		r := e.call("connector_ipsec_update", map[string]any{"id": "7", "patch": c.patch})
		if !r.IsError || !strings.Contains(r.Text, c.text) || e.portal.count("PUT") != 0 {
			t.Errorf("%s: isError=%v PUT=%d text=%q", name, r.IsError, e.portal.count("PUT"), r.Text)
		}
	}
}

// MA-W5: the D13 guard reads the GET /connectors row state of each family.
func TestConnectorFamilies_StateGuard(t *testing.T) {
	for _, name := range []string{"connector_ipsec", "connector_aws_vpc", "connector_azure_vnet"} {
		s := specNamed(name)
		for _, c := range []struct {
			row, network string
			ok           bool
		}{{"FAILED", "", false}, {"IN_PROGRESS", "", false}, {"SUCCESS", "FAILED", false}, {"PENDING", "FAILED", true}, {"SUCCESS", "", true}} {
			e := newFamilyEnv(t, s)
			e.portal.networkState = c.network
			o := s.stored()
			o["state"] = c.row
			s.put(e, 7, o)
			r := e.call(name+"_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}})
			if c.ok != !r.IsError || c.ok != (e.portal.count("PUT") == 1) {
				t.Errorf("%s row %s network %s: %s", name, c.row, c.network, r.Text)
			}
		}
	}
}

// MA-W26, MA-W41: an id of another connector family is refused with no write.
func TestConnectorFamilies_FamilyCheck(t *testing.T) {
	for _, name := range []string{"connector_ipsec", "connector_aws_vpc", "connector_azure_vnet"} {
		s := specNamed(name)
		e := newFamilyEnv(t, s)
		e.portal.put(8, s.stored())
		e.portal.connectors = []map[string]any{{"id": 8, "name": s.stored()["name"], "type": "SAAS", "state": "PENDING"}}
		mustRefuse(t, e.call(name+"_update", map[string]any{"id": "8", "patch": map[string]any{"description": "x"}}), "family_mismatch", "belongs to another family")
		mustRefuse(t, e.call(name+"_delete", map[string]any{"id": "8"}), "family_mismatch", "belongs to another family")
		if e.portal.count("PUT")+e.portal.count("DELETE") != 0 {
			t.Fatalf("%s: write sent", name)
		}
	}
}

// MA-W16: a connector marked for deletion is refused before any write.
func TestConnectorFamilies_DocState(t *testing.T) {
	for _, name := range []string{"connector_ipsec", "connector_aws_vpc", "connector_azure_vnet"} {
		s := specNamed(name)
		e := newFamilyEnv(t, s)
		e.portal.fail("GET", e.portal.base()+"/7", fault{status: 404, body: `{"code":"AK000003","message":"Could not find connector with id : 7"}`})
		o := s.stored()
		o["state"] = "SUCCESS"
		s.put(e, 7, o)
		mustRefuse(t, e.call(name+"_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}), "state_refused", "deletion pending provisioning")
		if e.portal.count("PUT") != 0 {
			t.Fatalf("%s: PUT sent", name)
		}
	}
}

// BC-R1, BC-14: a re-derived field the backend changes on save is named as
// re-derived, not as a change the request made.
func TestConnectorAwsVpc_Rederived(t *testing.T) {
	s := specNamed("connector_aws_vpc")
	e := newFamilyEnv(t, s)
	s.put(e, 7, s.stored())
	saved := s.stored()
	saved["id"], saved["description"], saved["vgwId"] = 7, "x", "vgw-2"
	b, _ := json.Marshal(saved)
	e.portal.fail("PUT", e.portal.base()+"/7", fault{status: 200, body: string(b)})
	r := e.call("connector_aws_vpc_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}})
	mustOK(t, r)
	if !strings.Contains(r.Text, `Re-derived by the backend on save: vgwId "vgw-1" to "vgw-2".`) {
		t.Fatalf("text:\n%s", r.Text)
	}
}

// BC-R4: setting per-subnet import entries drops the stored VNet-wide import settings.
func TestConnectorAzureVnet_ImportExclusive(t *testing.T) {
	s := specNamed("connector_azure_vnet")
	e := newFamilyEnv(t, s)
	s.put(e, 7, s.stored())
	patch := map[string]any{"vnetRouting": map[string]any{"importFromCXPOptions": map[string]any{
		"subnets": []any{map[string]any{"id": "subnet-1", "value": "10.1.0.0/24"}}}}}
	mustOK(t, e.call("connector_azure_vnet_update", map[string]any{"id": "7", "patch": patch}))
	imp := sentBody(t, e, "PUT")["vnetRouting"].(map[string]any)["importFromCXPOptions"].(map[string]any)
	if _, ok := imp["routeImportMode"]; ok || imp["subnets"] == nil {
		t.Fatalf("import %v", imp)
	}
}

// D16: flag-gated connector fields set while their flag is off are left out.
func TestConnectorFamilies_FlagGates(t *testing.T) {
	for _, c := range []struct {
		family, field string
		value         any
		flag          string
	}{
		{"connector_aws_vpc", "billingTags", []any{7}, "BILLING-TAGS"},
		{"connector_aws_vpc", "scaleGroupId", "sg-1", "SCALE-GROUPS-FOR-CLOUD-CONNECTORS"},
		{"connector_azure_vnet", "cxpPeeringGatewayId", 3, "ATH-ON-AWS-CXP"},
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
			if _, sent := sentBody(t, e, "POST")[c.field]; sent != on {
				t.Errorf("%s flag %v: %s sent %v\n%s", c.family, on, c.field, sent, r.Text)
			}
		}
	}
}

func fmtJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// GAP-09: a flag-gated field the backend drops on save is named in the
// result, not reported as a clean success.
func TestConnectorIPsec_DroppedFlagFieldNamed(t *testing.T) {
	e := newFamilyEnv(t, specNamed("connector_ipsec"))
	body := ipsecBody("ipsec1")
	body["routingOptions"] = map[string]any{"dynamicRouting": map[string]any{"customerGwAsn": "65001", "localAsn": "65100"}}
	e.portal.fail("POST", e.portal.base(), fault{status: 201, body: `{"id":101,"name":"ipsec1","cxp":"USEAST-AZURE-1","segments":["seg-a"],` +
		`"sites":[{"id":1,"name":"s1","customerGwIp":"203.0.113.1"}],"routingOptions":{"dynamicRouting":{"customerGwAsn":"65001"}}}`})
	r := e.call("connector_ipsec_create", map[string]any{"body": body})
	mustOK(t, r)
	if !strings.Contains(r.Text, "routingOptions.dynamicRouting.localAsn dropped") {
		t.Fatalf("text = %s", r.Text)
	}
}
