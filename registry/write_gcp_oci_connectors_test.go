package registry

import (
	"encoding/json"
	"maps"
	"strings"
	"testing"

	"github.com/alkiranet/mcp-alkira/writes"
)

func init() {
	familySpecs = append(familySpecs, gcpOciConnectorSpecs...)
	rows := func(k, crud, nonexistent string) map[string]string {
		return map[string]string{
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode; " + crud,
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go), H-W16 (ST-04)",
			scImportNonexistent: "TestFamilies_NotFound TestGcpOciConnectorFamilies_FamilyCheck; " + nonexistent,
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03)",
			scPreflightBlocked:  "TestGcpOciConnectorFamilies_StateGuard TestGcpOciConnectorFamilies_FamilyCheck; MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
			scSoftDeleteOff:     "TestGcpOciConnectorFamilies_DocState TestFamilies_DeletePreview",
		}
	}
	// GCP Interconnect and OCI VCN have no cloud fixture on dev: K34 and K30
	// run the not-found negative only (03-TEST-PLAN.md section 6).
	maps.Copy(scenarioRows, map[string]map[string]string{
		"connector_gcp_vpc":          rows("K29", "K29.1-.5", "K29.6"),
		"connector_gcp_interconnect": rows("K34", "unit only (no VLAN attachment on dev)", "K34.6"),
		"connector_oci_vcn":          rows("K30", "unit only (no OCI fixture on dev)", "K30.6"),
	})
}

// Stored bodies follow the TPS GET shapes (model/GCPVPCConnector.java,
// gcpinterconnect/GCPInterconnectConnector.java, oci/OCIVCNConnector.java).
var gcpOciConnectorSpecs = []familySpec{
	{
		name: "connector_gcp_vpc", noun: "GCP VPC connector", f: family(writes.ConnectorGcpVpc), coll: "gcpvpcconnectors", connector: "GCP_VPC",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"Could not find connector with id : 404"}`},
		stored:   gcpVpcStored,
		create:   gcpVpcBody,
		readOnly: []string{"implicitGroupId", "vpcId", "routingMode", "primary", "primaryConnectorId", "lastEnabled", "previousState"},
		creates: []createCase{
			{nil, gcpVpcBody("gcp1"), gcpVpcBody("gcp1"), nil},
			{nil, withFields(gcpVpcBody("gcp1"), map[string]any{"enabled": false, "customerAsn": 64512}),
				withFields(gcpVpcBody("gcp1"), map[string]any{"enabled": false, "customerAsn": 64512}), nil},
		},
		invalid: map[string]toolCase{
			// TPS does not trim-check descriptions (GCPVPCConnector.java:67-69).
			"trailing space": {},
			"ALK- prefix":    {"_create", map[string]any{"body": gcpVpcBody("ALK-x")}, "name must not start with ALK-"},
			"no region":      {"_create", map[string]any{"body": withoutField(gcpVpcBody("gcp1"), "customerRegion")}, "customerRegion"},
			"size":           {"_create", map[string]any{"body": withFields(gcpVpcBody("gcp1"), map[string]any{"size": "20LARGE"})}, "size must be one of"},
			"asn":            {"_create", map[string]any{"body": withFields(gcpVpcBody("gcp1"), map[string]any{"customerAsn": 65001 - 1000})}, "customerAsn"},
			"subnet type":    {"_create", map[string]any{"body": withFields(gcpVpcBody("gcp1"), map[string]any{"gcpRouting": gcpExport(false, map[string]any{"id": "1", "value": "10.0.0.0/24", "type": "CIDR"})})}, "type must be one of"},
			"all and list":   {"_create", map[string]any{"body": withFields(gcpVpcBody("gcp1"), map[string]any{"gcpRouting": gcpExport(true, map[string]any{"id": "1", "value": "10.0.0.0/24", "type": "SUBNET"})})}, "exportAllSubnets is true"},
			"routingMode":    {"_update", map[string]any{"id": "7", "patch": map[string]any{"routingMode": "REGIONAL"}}, "routingMode is read-only"},
			"region null":    {"_update", map[string]any{"id": "7", "patch": map[string]any{"customerRegion": nil}}, "customerRegion cannot be removed"},
		},
		inUse: []inUseCase{
			{fault{status: 409, body: `{"code":"AK000001","message":"CONNECTOR with id 7 is still referred externally"}`}, "referred externally"},
		},
		previews: []previewCase{
			{"PENDING", []string{"Deployed: no, never deployed (state PENDING).", `Also deleted with it: "gcp7".`, "The delete now: removes it at once."}},
			{"SUCCESS", []string{"Deployed: yes (state SUCCESS).", `Also deleted with it: "gcp7".`,
				"The delete now: marks it for deletion; it keeps carrying traffic, with its name and references, until the deletion is provisioned from the Alkira portal."}},
		},
	},
	{
		name: "connector_gcp_interconnect", noun: "GCP Interconnect connector", f: family(writes.ConnectorGcpInterconnect), coll: "gcp-interconnect-connectors", connector: "GCP_INTERCONNECT",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"Could not find connector with id : 404"}`},
		stored:   gciStored,
		create:   gciBody,
		readOnly: []string{"implicitGroupId", "primary", "lastEnabled", "previousState"},
		creates: []createCase{
			{nil, gciBody("gci1"), gciBody("gci1"), nil},
		},
		invalid: map[string]toolCase{
			// TPS does not trim-check descriptions (GCPInterconnectConnector.java:77-79).
			"trailing space": {},
			"ALK- prefix":    {"_create", map[string]any{"body": gciBody("ALK-x")}, "name must not start with ALK-"},
			"size":           {"_create", map[string]any{"body": withFields(gciBody("gci1"), map[string]any{"size": "XSMALL"})}, "size must be one of"},
			"no instances":   {"_create", map[string]any{"body": withoutField(gciBody("gci1"), "instances")}, "instances"},
			"key in body":    {"_create", map[string]any{"body": withFields(gciBody("gci1"), map[string]any{"instances": []any{withFields(gciInstance(nil), map[string]any{"bgpAuthKeyAlkira": "k"})}})}, "bgpAuthKeyAlkira"},
			"vni with GRE":   {"_create", map[string]any{"body": withFields(gciBody("gci1"), map[string]any{"instances": []any{withFields(gciInstance(nil), map[string]any{"vni": 16773023})}})}, "vni is set only when"},
			"key in patch":   {"_update", map[string]any{"id": "7", "patch": map[string]any{"instances": []any{withFields(gciInstance(500), map[string]any{"bgpAuthKeyAlkira": "k"})}}}, "bgpAuthKeyAlkira"},
			"segmentId":      {"_update", map[string]any{"id": "7", "patch": map[string]any{"instances": []any{gciInstance(500, map[string]any{"segmentId": 9})}}}, "segmentId is read-only"},
		},
		inUse: []inUseCase{
			{fault{status: 409, body: `{"code":"AK000001","message":"CONNECTOR with id 7 is still referred externally"}`}, "referred externally"},
		},
		previews: []previewCase{
			{"PENDING", []string{"Deployed: no, never deployed (state PENDING).", `Also deleted with it: "gci7".`}},
		},
	},
	{
		name: "connector_oci_vcn", noun: "OCI VCN connector", f: family(writes.ConnectorOciVcn), coll: "oci-vcn-connectors", connector: "OCI_VCN",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"Could not find connector with id : 404"}`},
		stored:   ociVcnStored,
		create:   ociVcnBody,
		readOnly: []string{"implicitGroupId", "drgId", "primary", "primaryConnectorId", "vcnCompartmentName", "lastEnabled", "previousState"},
		creates: []createCase{
			{nil, ociVcnBody("oci1"), ociVcnBody("oci1"), nil},
		},
		invalid: map[string]toolCase{
			// The OCI VCN connector name allows "/" (OCIVCNConnector.java:63), and
			// TPS does not trim-check descriptions (:66-68).
			"charset":        {},
			"trailing space": {},
			"ALK- prefix":    {"_create", map[string]any{"body": ociVcnBody("ALK-x")}, "name must not start with ALK-"},
			"size":           {"_create", map[string]any{"body": withFields(ociVcnBody("oci1"), map[string]any{"size": "2LARGE"})}, "size must be one of"},
			"primary":        {"_update", map[string]any{"id": "7", "patch": map[string]any{"primary": false}}, "primary is read-only"},
		},
		inUse: []inUseCase{
			{fault{status: 409, body: `{"code":"AK000001","message":"CONNECTOR with id 7 is still referred externally"}`}, "referred externally"},
		},
		previews: []previewCase{
			{"PENDING", []string{"Deployed: no, never deployed (state PENDING).", `Also deleted with it: "oci7".`}},
		},
	},
}

func gcpVpcStored() map[string]any {
	return map[string]any{"name": "gcp7", "description": "seven", "size": "SMALL", "segments": []any{"seg-a"}, "cxp": "USEAST-AZURE-1",
		"customerRegion": "us-east1", "vpcId": "5858523136380003535", "vpcName": "net-1", "projectId": "proj-1", "implicitGroupId": 61101,
		"credentialId": "cred-gcp", "routingMode": "GLOBAL", "billingTags": []any{}, "subnetRegions": []any{"us-east1"},
		"gcpRouting": map[string]any{"importFromCXPOptions": map[string]any{"prefixListIds": []any{}, "routeImportMode": "ADVERTISE_DEFAULT_ROUTE"}},
		"enabled":    true, "primary": true, "secondaryCXPs": []any{}, "lastEnabled": 1763600000, "primaryConnectorId": 7, "customerAsn": 64522}
}

func gcpVpcBody(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "USEAST-AZURE-1", "segments": []any{"seg-a"}, "customerRegion": "us-east1", "vpcName": "net-1", "credentialId": "cred-gcp"}
}

func gcpExport(all bool, prefixes ...any) map[string]any {
	return map[string]any{"exportToCXPOptions": map[string]any{"exportAllSubnets": all, "userInputPrefixes": prefixes}}
}

// gciInstance is one instance on seg-a; id nil leaves the id out, and seg
// adds fields to its segment option.
func gciInstance(id any, seg ...map[string]any) map[string]any {
	so := map[string]any{"segmentName": "seg-a", "customerGateways": []any{map[string]any{"tunnelCount": 2}}, "disableInternetExit": false, "advertiseOnPremRoutes": false}
	for _, s := range seg {
		maps.Copy(so, s)
	}
	in := map[string]any{"name": "inst-1", "segmentOptions": []any{so}, "customerAsn": 65001}
	if id != nil {
		in["id"] = id
	}
	return in
}

func gciStored() map[string]any {
	in := gciInstance(500, map[string]any{"segmentId": 1207})
	in["gcpEdgeAvailabilityDomain"], in["bgpAuthKeyAlkira"] = "AVAILABILITY_DOMAIN_1", "stored-bgp"
	return map[string]any{"name": "gci7", "description": "seven", "cxp": "USEAST-AZURE-1", "implicitGroupId": 61201, "size": "LARGE",
		"billingTags": []any{}, "enabled": true, "loopbackPrefixes": []any{"10.240.0.0/24"}, "instances": []any{in}, "tunnelProtocol": "GRE"}
}

func gciBody(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "USEAST-AZURE-1", "loopbackPrefixes": []any{"10.240.0.0/24"}, "instances": []any{gciInstance(nil)}}
}

func ociVcnStored() map[string]any {
	return map[string]any{"name": "oci7", "description": "seven", "cxp": "USEAST-AZURE-1", "segments": []any{"seg-a"}, "vcnId": "ocid1.vcn.oc1.iad.aaaa",
		"customerRegion": "us-ashburn-1", "credentialId": "cred-oci", "implicitGroupId": 61301, "drgId": "ocid1.drg.oc1.iad.bbbb", "billingTags": []any{}, "size": "SMALL",
		"vcnRouting": map[string]any{
			"exportToCXPOptions":   map[string]any{"userInputPrefixes": []any{map[string]any{"value": "10.234.0.0/24", "type": "CIDR", "defaultRouting": true}}, "routeExportMode": "USER_INPUT_PREFIXES"},
			"importFromCXPOptions": map[string]any{"routeTables": []any{map[string]any{"id": "rt-1", "prefixListIds": []any{}, "routeImportMode": "ADVERTISE_DEFAULT_ROUTE", "defaultRouting": true}}}},
		"enabled": true, "primary": true, "secondaryCXPs": []any{}, "lastEnabled": 1763600000, "primaryConnectorId": 7, "vcnCompartmentName": "dev"}
}

func ociVcnBody(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "USEAST-AZURE-1", "segments": []any{"seg-a"}, "vcnId": "ocid1.vcn.oc1.iad.aaaa", "customerRegion": "us-ashburn-1", "credentialId": "cred-oci"}
}

var gcpOciConnectorNames = []string{"connector_gcp_vpc", "connector_gcp_interconnect", "connector_oci_vcn"}

// MA-W5: the D13 guard reads the GET /connectors row state of each family.
func TestGcpOciConnectorFamilies_StateGuard(t *testing.T) {
	for _, name := range gcpOciConnectorNames {
		s := specNamed(name)
		for _, c := range []struct {
			row, network string
			ok           bool
		}{{"FAILED", "", false}, {"SCHEDULED", "", false}, {"IN_PROGRESS", "", false}, {"SUCCESS", "FAILED", false}, {"PENDING", "FAILED", true}, {"SUCCESS", "", true}} {
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
func TestGcpOciConnectorFamilies_FamilyCheck(t *testing.T) {
	for _, name := range gcpOciConnectorNames {
		s := specNamed(name)
		e := newFamilyEnv(t, s)
		e.portal.put(8, s.stored())
		e.portal.connectors = []map[string]any{{"id": 8, "name": s.stored()["name"], "type": "AWS_VPC", "state": "PENDING"}}
		mustRefuse(t, e.call(name+"_update", map[string]any{"id": "8", "patch": map[string]any{"description": "x"}}), "family_mismatch", "belongs to another family")
		mustRefuse(t, e.call(name+"_delete", map[string]any{"id": "8"}), "family_mismatch", "belongs to another family")
		if e.portal.count("PUT")+e.portal.count("DELETE") != 0 {
			t.Fatalf("%s: write sent", name)
		}
	}
}

// MA-W16: a connector marked for deletion is refused before any write.
func TestGcpOciConnectorFamilies_DocState(t *testing.T) {
	for _, name := range gcpOciConnectorNames {
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

// MA-W2: vpcName and projectId cannot change, and the import rules apply
// to the merged body; each refusal sends no PUT.
func TestConnectorGcpVpc_UpdateRefusals(t *testing.T) {
	s := specNamed("connector_gcp_vpc")
	for name, c := range map[string]struct {
		patch map[string]any
		text  string
	}{
		"vpcName":   {map[string]any{"vpcName": "other-net"}, "vpcName cannot be changed"},
		"projectId": {map[string]any{"projectId": "other"}, "projectId cannot be changed"},
		"custom without lists": {map[string]any{"gcpRouting": map[string]any{"importFromCXPOptions": map[string]any{"routeImportMode": "ADVERTISE_CUSTOM_PREFIX"}}},
			"prefixListIds must be set"},
	} {
		e := newFamilyEnv(t, s)
		s.put(e, 7, s.stored())
		mustRefuse(t, e.call("connector_gcp_vpc_update", map[string]any{"id": "7", "patch": c.patch}), "invalid_arguments", c.text)
		if e.portal.count("PUT") != 0 {
			t.Fatalf("%s: PUT sent", name)
		}
	}
}

// MA-W25, BC-14: routingMode re-stamped on an unrelated PUT is named as
// re-derived, not as a change the request made.
func TestConnectorGcpVpc_Rederived(t *testing.T) {
	s := specNamed("connector_gcp_vpc")
	e := newFamilyEnv(t, s)
	s.put(e, 7, s.stored())
	saved := s.stored()
	saved["id"], saved["description"], saved["routingMode"] = 7, "x", "REGIONAL"
	b, _ := json.Marshal(saved)
	e.portal.fail("PUT", e.portal.base()+"/7", fault{status: 200, body: string(b)})
	r := e.call("connector_gcp_vpc_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}})
	mustOK(t, r)
	if !strings.Contains(r.Text, `Re-derived by the backend on save: routingMode "GLOBAL" to "REGIONAL".`) {
		t.Fatalf("text:\n%s", r.Text)
	}
}

// MA-W4 (routing provenance): an OCI update that leaves vcnRouting out
// leaves the backend-built routing out of the PUT; a user-set entry keeps it.
func TestConnectorOciVcn_RoutingProvenance(t *testing.T) {
	s := specNamed("connector_oci_vcn")
	e := newFamilyEnv(t, s)
	s.put(e, 7, s.stored())
	mustOK(t, e.call("connector_oci_vcn_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}))
	if _, ok := sentBody(t, e, "PUT")["vcnRouting"]; ok {
		t.Fatal("backend-built vcnRouting resent")
	}
	e = newFamilyEnv(t, s)
	o := s.stored()
	o["vcnRouting"].(map[string]any)["exportToCXPOptions"].(map[string]any)["userInputPrefixes"] = []any{map[string]any{"value": "10.234.0.0/24", "type": "CIDR"}}
	s.put(e, 7, o)
	mustOK(t, e.call("connector_oci_vcn_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}))
	if put := sentBody(t, e, "PUT"); put["vcnRouting"] == nil || strings.Contains(fmtJSON(put["vcnRouting"]), "defaultRouting") {
		t.Fatalf("user routing: %s", fmtJSON(put["vcnRouting"]))
	}
}

// MA-W3, MA-W15: the Interconnect schemas hold no BGP key; an update resends
// the stored key of each kept instance, and the output masks it.
func TestConnectorGcpInterconnect_StoredKey(t *testing.T) {
	s := specNamed("connector_gcp_interconnect")
	e := newFamilyEnv(t, s)
	var res struct {
		Tools []struct {
			Name        string          `json:"name"`
			InputSchema json.RawMessage `json:"inputSchema"`
		} `json:"tools"`
	}
	json.Unmarshal(call(t, e.srv, "tools/list", nil), &res)
	for _, tool := range res.Tools {
		if strings.HasPrefix(tool.Name, "connector_gcp_interconnect_") && strings.Contains(string(tool.InputSchema), "bgpAuthKeyAlkira") {
			t.Errorf("%s schema names the key field", tool.Name)
		}
	}
	s.put(e, 7, s.stored())
	r := e.call("connector_gcp_interconnect_update", map[string]any{"id": "7", "patch": map[string]any{"instances": []any{
		withFields(gciInstance(500), map[string]any{"customerAsn": 65002})}}})
	mustOK(t, r)
	in := sentBody(t, e, "PUT")["instances"].([]any)[0].(map[string]any)
	if in["bgpAuthKeyAlkira"] != "stored-bgp" || fmtJSON(in["customerAsn"]) != "65002" || in["gcpEdgeAvailabilityDomain"] != "AVAILABILITY_DOMAIN_1" {
		t.Fatalf("PUT instance %v", in)
	}
	if strings.Contains(r.Text, "stored-bgp") {
		t.Fatalf("output shows the key:\n%s", r.Text)
	}
	// A patch that leaves a stored instance out is refused with no write.
	e = newFamilyEnv(t, s)
	s.put(e, 7, s.stored())
	r = e.call("connector_gcp_interconnect_update", map[string]any{"id": "7", "patch": map[string]any{"instances": []any{gciInstance(nil)}}})
	if !r.IsError || e.portal.count("PUT") != 0 {
		t.Fatalf("instance dropped: %s", r.Text)
	}
}

// D16: flag-gated connector fields set while their flag is off are left out.
func TestGcpOciConnectorFamilies_FlagGates(t *testing.T) {
	for _, c := range []struct {
		family, field string
		value         any
		flag          string
	}{
		{"connector_gcp_vpc", "billingTags", []any{7}, "BILLING-TAGS"},
		{"connector_gcp_vpc", "scaleGroupId", "sg-1", "SCALE-GROUPS-FOR-CLOUD-CONNECTORS"},
		{"connector_gcp_interconnect", "billingTags", []any{7}, "BILLING-TAGS"},
		{"connector_gcp_interconnect", "scaleGroupId", "sg-1", "SUPPORT-SCALE-GROUP-FOR-DX-CONNECTOR"},
		{"connector_oci_vcn", "billingTags", []any{7}, "BILLING-TAGS"},
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
