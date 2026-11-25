package registry

import (
	"encoding/json"
	"maps"
	"strings"
	"testing"

	"github.com/alkiranet/mcp-alkira/writes"
)

func init() {
	// The create sends disableInternetExit true in every VRF mapping that
	// leaves it out, as the portal does.
	for _, s := range sdwanSpecs {
		for i := range s.creates {
			s.creates[i].want = withVRFDefaults(s.creates[i].want)
		}
	}
	familySpecs = append(familySpecs, sdwanSpecs...)
	for name, k := range map[string]string{"connector_aruba_edge": "K35", "connector_cisco_sdwan": "K36", "connector_fortinet_sdwan": "K37",
		"connector_versa_sdwan": "K38", "connector_vmware_sdwan": "K39"} {
		maps.Copy(scenarioRows, map[string]map[string]string{name: {
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode TestSdwan_InstancesByID; " + k + ".1-.6",
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go)",
			scImportNonexistent: "TestFamilies_NotFound TestSdwan_FamilyCheck",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03); TestFamilies_InvalidArgumentsSendNothing",
			scPreflightBlocked:  "TestSdwan_StateGuard; MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
			scSoftDeleteOff:     "TestSdwan_MarkedRow TestFamilies_DeletePreview",
		}})
	}
	scenarioRows["connector_versa_sdwan"][scCRUD] = "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode TestSdwan_InstancesByID TestSdwan_VersaCreateKeys; K38"
}

var sdwanNotFound = fault{status: 404, body: `{"code":"AK000003","message":"Could not find connector with id : 404"}`}

func sdwanInUse() []inUseCase {
	return []inUseCase{
		{fault{status: 400, body: `{"code":"AK000002","message":"connector c7 is associated with segment resource sr-1"}`}, "is associated with segment resource sr-1"},
		{fault{status: 409, body: `{"code":"AK000001","message":"CONNECTOR with id 7 is still referred externally"}`}, "is still referred externally"},
	}
}

func sdwanPreviews(noun, also string) []previewCase {
	lines := func(state string, extra ...string) []string {
		return append([]string{`Delete preview for ` + noun + ` "c7" (id 7). Nothing was deleted.`, `Also deleted with it: "c7".`}, extra...)
	}
	pending := lines("PENDING", "Deployed: no, never deployed (state PENDING).", "The next provision: nothing to tear down.")
	deployed := lines("SUCCESS", "Deployed: yes (state SUCCESS).", "The next provision: tears down the deployed object.")
	if also != "" {
		pending = append(pending, "Also deleted with it: "+also+".")
		deployed = append(deployed, "Also deleted with it: "+also+".")
	}
	return []previewCase{{"PENDING", pending}, {"SUCCESS", deployed}}
}

func arubaCreate(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "USEAST-AZURE-1", "size": "SMALL", "version": "9.3.8.1", "boostMode": false, "tunnelProtocol": "IPSEC",
		"instances":            []any{map[string]any{"siteTag": "site-a", "credentialId": "cred-aruba", "hostName": "edge1", "accountName": "acct-a"}},
		"arubaEdgeVRFMappings": []any{map[string]any{"arubaEdgeConnectSegmentName": "Default", "alkiraSegmentId": 1207, "gatewayBgpAsn": 65010}}}
}

func ciscoCreate(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "USEAST-AZURE-1", "size": "SMALL", "version": "20.4.1",
		"ciscoEdgeInfo":        []any{map[string]any{"cloudInitFile": "#cloud-config", "credentialId": "cred-cisco", "hostName": "edge1"}},
		"ciscoEdgeVRFMappings": []any{map[string]any{"vrf": 10, "segmentId": 1207}}}
}

func fortinetSdwanCreate(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "USEAST-AZURE-1", "size": "SMALL",
		"instances":            []any{map[string]any{"credentialId": "c-fg1", "hostName": "fg1", "version": "7.4.11"}},
		"ftntSDWANVRFMappings": []any{map[string]any{"vrf": 1, "segmentId": 1207, "gatewayBgpAsn": 65030}}}
}

func versaCreate(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "USEAST-AZURE-1", "size": "SMALL", "remoteId": "remote@example.com", "localId": "local@example.com",
		"versaControllerHost": "vc.example.com", "localPublicSharedKey": "lk-1", "remotePublicSharedKey": "rk-1",
		"instances":             []any{map[string]any{"hostName": "v1", "version": "21.2.3-B"}},
		"versaSDWANVRFMappings": []any{map[string]any{"vrfName": "vrf-a", "segmentId": 1207, "gatewayBgpAsn": 65040}}}
}

// withVRFDefaults returns body with disableInternetExit true, the portal
// default, in every VRF mapping that leaves it out.
func withVRFDefaults(body map[string]any) map[string]any {
	out := jsonRoundTrip(body)
	for k, v := range out {
		list, _ := v.([]any)
		if !strings.HasSuffix(k, "VRFMappings") {
			continue
		}
		for _, m := range list {
			if mm := m.(map[string]any); mm["disableInternetExit"] == nil {
				mm["disableInternetExit"] = true
			}
		}
	}
	return out
}

func vmwareCreate(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "USEAST-AZURE-1", "size": "SMALL", "version": "5.2.4", "orchestratorHostAddress": "https://vco.example.com",
		"instances":              []any{map[string]any{"credentialId": "cred-vmw", "hostName": "e1"}},
		"vmWareSdWanVRFMappings": []any{map[string]any{"vmWareSdWanSegmentName": "Global", "segmentId": 1207, "gatewayBgpAsn": 65050}}}
}

// sdwanStored returns a stored body: create("c7") plus the backend fields of
// a GET, each instance with its id and backend name.
func sdwanStored(create func(string) map[string]any, list string, extra map[string]any) func() map[string]any {
	return func() map[string]any {
		o := withFields(jsonRoundTrip(create("c7")), withFields(map[string]any{"description": "seven", "implicitGroupId": 61007, "billingTags": []any{}}, extra))
		items := o[list].([]any)
		for i, it := range items {
			m := it.(map[string]any)
			m["id"] = float64(5000 + i)
			m["name"] = "c7-" + m["hostName"].(string)
		}
		return o
	}
}

var sdwanSpecs = []familySpec{
	{
		name: "connector_aruba_edge", noun: "Aruba EdgeConnect connector", f: family(writes.ConnectorArubaEdge), coll: "aruba-edge-connectors",
		connector: "ARUBA_EDGE_CONNECT", notFound: sdwanNotFound, create: arubaCreate,
		stored:   sdwanStored(arubaCreate, "instances", map[string]any{"enabled": true, "lastEnabled": 1763600000}),
		readOnly: []string{"implicitGroupId", "lastEnabled", "state"},
		creates: []createCase{
			{nil, withFields(arubaCreate("a1"), map[string]any{"billingTags": []any{3}}), withFields(arubaCreate("a1"), map[string]any{"billingTags": []any{3}}), nil},
			{[]string{"SCALE-GROUPS-FOR-USER-AND-SITES"}, withFields(arubaCreate("a1"), map[string]any{"scaleGroupId": "sg-1"}),
				withFields(arubaCreate("a1"), map[string]any{"scaleGroupId": "sg-1"}), []string{"SCALE-GROUPS-FOR-USER-AND-SITES"}},
			{[]string{"RELEASE-CONNECTOR-ARUBA-EDGE-CONNECT-ALLOW-LIST"}, withFields(arubaCreate("a1"), map[string]any{"allowList": []any{"198.51.100.0/24"}}),
				withFields(arubaCreate("a1"), map[string]any{"allowList": []any{"198.51.100.0/24"}}), []string{"RELEASE-CONNECTOR-ARUBA-EDGE-CONNECT-ALLOW-LIST"}},
		},
		invalid: map[string]toolCase{
			// TPS accepts a description with a trailing space.
			"trailing space": {},
			"ALK- prefix":    {"_create", map[string]any{"body": arubaCreate("ALK-a")}, "name must not start with ALK-"},
			"no boostMode":   {"_create", map[string]any{"body": withFields(arubaCreate("a"), map[string]any{"boostMode": nil})}, "boostMode"},
			"same hostName":  {"_create", map[string]any{"body": withFields(arubaCreate("a"), map[string]any{"instances": []any{arubaCreate("a")["instances"].([]any)[0], arubaCreate("a")["instances"].([]any)[0]}})}, "hostName"},
			"instance name":  {"_update", map[string]any{"id": "7", "patch": map[string]any{"instances": []any{map[string]any{"id": 5000, "name": "x"}}}}, "name is read-only"},
		},
		inUse:    sdwanInUse(),
		previews: sdwanPreviews("Aruba EdgeConnect connector", ""),
	},
	{
		name: "connector_cisco_sdwan", noun: "Cisco SD-WAN connector", f: family(writes.ConnectorCiscoSdwan), coll: "ciscosdwaningresses",
		connector: "SD_WAN", notFound: sdwanNotFound, create: ciscoCreate,
		stored:   sdwanStored(ciscoCreate, "ciscoEdgeInfo", map[string]any{"type": "VEDGE", "tunnelProtocol": "IPSEC", "enabled": true, "lastEnabled": 1763600000, "validationState": map[string]any{"state": "SUCCESS"}}),
		readOnly: []string{"implicitGroupId", "lastEnabled", "validationState", "state"},
		creates: []createCase{
			{nil, ciscoCreate("c1"), ciscoCreate("c1"), nil},
			{[]string{"BILLING-TAGS"}, withFields(ciscoCreate("c1"), map[string]any{"billingTags": []any{3}}),
				withFields(ciscoCreate("c1"), map[string]any{"billingTags": []any{3}}), []string{"BILLING-TAGS"}},
			{[]string{"RELEASE-CONNECTOR-CISCO-SDWAN-ALLOW-LIST"}, withFields(ciscoCreate("c1"), map[string]any{"type": "CAT8000V", "allowList": []any{"198.51.100.0/24"}}),
				withFields(ciscoCreate("c1"), map[string]any{"type": "CAT8000V", "allowList": []any{"198.51.100.0/24"}}), []string{"RELEASE-CONNECTOR-CISCO-SDWAN-ALLOW-LIST"}},
			// The portal never sends ?async=true for this family, so the flag changes nothing.
			{[]string{"RELEASE-ASYNC-NETWORK-ENTITY-VALIDATION-PHASE-1"}, ciscoCreate("c1"), ciscoCreate("c1"), nil},
		},
		invalid: map[string]toolCase{
			// TPS accepts a description with a trailing space.
			"trailing space": {},
			"ALK- prefix":    {"_create", map[string]any{"body": ciscoCreate("alk-c")}, "name must not start with ALK-"},
			"vrf 512":        {"_create", map[string]any{"body": withFields(ciscoCreate("a"), map[string]any{"ciscoEdgeVRFMappings": []any{map[string]any{"vrf": 512, "segmentId": 1}}})}, "vrf"},
			"GRE on VEDGE":   {"_create", map[string]any{"body": withFields(ciscoCreate("a"), map[string]any{"tunnelProtocol": "GRE"})}, "tunnelProtocol GRE needs type CAT8000V"},
			"key pair CSR":   {"_create", map[string]any{"body": withFields(ciscoCreate("a"), map[string]any{"type": "CSR", "ciscoEdgeInfo": []any{map[string]any{"cloudInitFile": "#c", "credentialId": "c", "hostName": "e1", "sshKeyPairCredentialId": "kp"}}})}, "sshKeyPairCredentialId is not supported on type CSR"},
			"no cloud-init":  {"_create", map[string]any{"body": withFields(ciscoCreate("a"), map[string]any{"ciscoEdgeInfo": []any{map[string]any{"credentialId": "c", "hostName": "e1"}}})}, "cloudInitFile"},
		},
		inUse:    sdwanInUse(),
		previews: sdwanPreviews("Cisco SD-WAN connector", ""),
	},
	{
		name: "connector_fortinet_sdwan", noun: "Fortinet SD-WAN connector", f: family(writes.ConnectorFortinetSdwan), coll: "ftnt-sdwan-connectors",
		connector: "FORTINET_SDWAN", notFound: sdwanNotFound, create: fortinetSdwanCreate,
		stored:   sdwanStored(fortinetSdwanCreate, "instances", map[string]any{"tunnelProtocol": "IPSEC", "enabled": true}),
		readOnly: []string{"implicitGroupId", "state"},
		creates: []createCase{
			{nil, withFields(fortinetSdwanCreate("f1"), map[string]any{"billingTags": []any{3}, "allowList": []any{"198.51.100.0/24"}}),
				withFields(fortinetSdwanCreate("f1"), map[string]any{"billingTags": []any{3}, "allowList": []any{"198.51.100.0/24"}}), nil},
		},
		invalid: map[string]toolCase{
			// TPS accepts a description with a trailing space.
			"trailing space": {},
			"ALK- prefix":    {"_create", map[string]any{"body": fortinetSdwanCreate("ALK-f")}, "name must not start with ALK-"},
			"BYOL no serial": {"_create", map[string]any{"body": withFields(fortinetSdwanCreate("a"), map[string]any{"instances": []any{map[string]any{"credentialId": "c", "hostName": "fg1", "version": "7.4.11", "licenseType": "BRING_YOUR_OWN"}}})}, "serialNumber must be set when licenseType is BRING_YOUR_OWN"},
			"vrf 32":         {"_create", map[string]any{"body": withFields(fortinetSdwanCreate("a"), map[string]any{"ftntSDWANVRFMappings": []any{map[string]any{"vrf": 32, "segmentId": 1, "gatewayBgpAsn": 1}}})}, "vrf"},
			"bucketNameId":   {"_update", map[string]any{"id": "7", "patch": map[string]any{"instances": []any{map[string]any{"id": 5000, "bucketNameId": "b"}}}}, "bucketNameId is read-only"},
		},
		inUse:    sdwanInUse(),
		previews: sdwanPreviews("Fortinet SD-WAN connector", ""),
	},
	{
		name: "connector_versa_sdwan", noun: "Versa SD-WAN connector", f: family(writes.ConnectorVersaSdwan), coll: "versa-sdwan-connectors",
		connector: "VERSA_SDWAN", notFound: sdwanNotFound, create: versaCreate,
		creates: []createCase{
			{nil, versaCreate("v1"), versaCreate("v1"), nil},
		},
		stored: sdwanStored(versaCreate, "instances", map[string]any{"remotePublicSharedKey": "rk-stored", "localPublicSharedKey": "lk-stored",
			"globalTenantId": 1, "tunnelProtocol": "IPSEC", "enabled": true}),
		readOnly: []string{"implicitGroupId", "state"},
		invalid: map[string]toolCase{
			// TPS accepts a description with a trailing space.
			"trailing space": {},
			"patch key":      {"_update", map[string]any{"id": "7", "patch": map[string]any{"remotePublicSharedKey": "k"}}, "remotePublicSharedKey"},
		},
		inUse:    sdwanInUse(),
		previews: sdwanPreviews("Versa SD-WAN connector", ""),
	},
	{
		name: "connector_vmware_sdwan", noun: "VMware SD-WAN connector", f: family(writes.ConnectorVmwareSdwan), coll: "vmware-sdwan-connectors",
		connector: "VMWARE_SDWAN", notFound: sdwanNotFound, create: vmwareCreate,
		stored:   sdwanStored(vmwareCreate, "instances", map[string]any{"tunnelProtocol": "IPSEC", "enabled": true, "lastEnabled": 1763600000}),
		readOnly: []string{"implicitGroupId", "lastEnabled", "state"},
		creates: []createCase{
			{nil, withFields(vmwareCreate("w1"), map[string]any{"billingTags": []any{3}}),
				withFields(vmwareCreate("w1"), map[string]any{"billingTags": []any{3}, "tunnelProtocol": "IPSEC"}), nil},
			{[]string{"RELEASE-CONNECTOR-VMWARE-SDWAN-ALLOW-LIST"}, withFields(vmwareCreate("w1"), map[string]any{"allowList": []any{"198.51.100.0/24"}}),
				withFields(vmwareCreate("w1"), map[string]any{"allowList": []any{"198.51.100.0/24"}, "tunnelProtocol": "IPSEC"}), []string{"RELEASE-CONNECTOR-VMWARE-SDWAN-ALLOW-LIST"}},
		},
		invalid: map[string]toolCase{
			// TPS accepts a description with a trailing space.
			"trailing space":  {},
			"ALK- prefix":     {"_create", map[string]any{"body": vmwareCreate("ALK-w")}, "name must not start with ALK-"},
			"two mappings":    {"_create", map[string]any{"body": withFields(vmwareCreate("a"), map[string]any{"vmWareSdWanVRFMappings": append(vmwareCreate("a")["vmWareSdWanVRFMappings"].([]any), map[string]any{"vmWareSdWanSegmentName": "B", "segmentId": 1, "gatewayBgpAsn": 1})})}, "vmWareSdWanVRFMappings"},
			"no orchestrator": {"_create", map[string]any{"body": withFields(vmwareCreate("a"), map[string]any{"orchestratorHostAddress": nil})}, "orchestratorHostAddress"},
			"tunnel GRE":      {"_create", map[string]any{"body": withFields(vmwareCreate("a"), map[string]any{"tunnelProtocol": "GRE"})}, "tunnelProtocol"},
		},
		inUse:    sdwanInUse(),
		previews: sdwanPreviews("VMware SD-WAN connector", ""),
	},
}

func sdwanUpdate(e *writeEnv, s familySpec, id string) writeResult {
	return e.call(s.name+"_update", map[string]any{"id": id, "patch": map[string]any{"description": "x"}})
}

// MA-W11: a connector of any type holding the name refuses the create with
// no POST; the guard reads GET /connectors?name= and the paged groups, never
// the family list. A group holding the name refuses it too.
func TestSdwan_DuplicateGuard(t *testing.T) {
	for _, s := range sdwanSpecs {
		t.Run(s.name, func(t *testing.T) {
			e := newFamilyEnv(t, s)
			e.portal.connectors = []map[string]any{{"id": 31, "name": "dup", "type": "IP_SEC", "state": "PENDING"}}
			mustRefuse(t, e.call(s.name+"_create", map[string]any{"body": s.createBody("dup")}), "duplicate",
				`an ip_sec connector named "dup" already exists (id 31); connector names are unique across every connector type and group.`)
			e.portal.groups = []map[string]any{{"id": 41, "name": "grp"}}
			mustRefuse(t, e.call(s.name+"_create", map[string]any{"body": s.createBody("grp")}), "duplicate", `"grp"`)
			if e.portal.count("POST") != 0 {
				t.Fatal("POST sent")
			}
			for _, r := range e.portal.requests() {
				if r.Path == e.portal.base() || strings.Contains(r.Query, "name=") && r.Path != "/api/tenantnetworks/7/connectors" {
					t.Fatalf("guard read %+v", r)
				}
			}
		})
	}
}

// MA-W5: the D13 guard reads the state from the GET /connectors row and
// refuses FAILED, SCHEDULED and IN_PROGRESS with no PUT; on a FAILED
// network only a PENDING connector may change.
func TestSdwan_StateGuard(t *testing.T) {
	for _, s := range sdwanSpecs {
		t.Run(s.name, func(t *testing.T) {
			e := newFamilyEnv(t, s)
			for i, state := range []string{"FAILED", "IN_PROGRESS", "SCHEDULED"} {
				o := s.stored()
				o["state"] = state
				s.put(e, 7+i, o)
				mustRefuse(t, sdwanUpdate(e, s, string(rune('7'+i))), "state_refused", "its state is "+state)
			}
			if e.portal.count("PUT") != 0 {
				t.Fatal("PUT sent on a refused update")
			}
			o := s.stored()
			o["state"] = "SUCCESS"
			s.put(e, 20, o)
			e.portal.networkState = "FAILED"
			mustRefuse(t, sdwanUpdate(e, s, "20"), "state_refused", "")
			pending := s.stored()
			s.put(e, 21, pending)
			mustOK(t, sdwanUpdate(e, s, "21"))
		})
	}
}

// BC-R2: an id of another connector type is refused before any write,
// whether its body lacks the family field or carries it under another row type.
func TestSdwan_FamilyCheck(t *testing.T) {
	for _, s := range sdwanSpecs {
		t.Run(s.name, func(t *testing.T) {
			e := newFamilyEnv(t, s)
			e.portal.put(40, map[string]any{"name": "ipsec", "segments": []any{"seg-a"}, "sites": []any{}})
			e.portal.connectors = append(e.portal.connectors, map[string]any{"id": 40, "name": "ipsec", "type": "IP_SEC", "state": "PENDING", "previousState": "PENDING"})
			mustRefuse(t, sdwanUpdate(e, s, "40"), "family_mismatch", "")
			mustRefuse(t, e.call(s.name+"_delete", map[string]any{"id": "40"}), "family_mismatch", "")
			o := s.stored()
			e.portal.put(41, o)
			e.portal.connectors = append(e.portal.connectors, map[string]any{"id": 41, "name": "c7", "type": "IP_SEC", "state": "PENDING", "previousState": "PENDING"})
			mustRefuse(t, sdwanUpdate(e, s, "41"), "family_mismatch", "")
			if e.portal.count("PUT")+e.portal.count("DELETE") != 0 {
				t.Fatal("write sent")
			}
		})
	}
}

// MA-W16: a connector marked for deletion (plain GET not found, the
// includeMarkedForDeletion GET found) is refused before any write.
func TestSdwan_MarkedRow(t *testing.T) {
	for _, s := range sdwanSpecs {
		t.Run(s.name, func(t *testing.T) {
			e := newFamilyEnv(t, s)
			s.put(e, 7, s.stored())
			e.portal.fail("GET", e.portal.base()+"/7", sdwanNotFound)
			mustRefuse(t, sdwanUpdate(e, s, "7"), "state_refused", "deletion pending provisioning")
			if e.portal.count("PUT") != 0 {
				t.Fatal("PUT sent")
			}
		})
	}
}

// MA-W4: a patch on the instances (Cisco edges) lists every stored entry by
// id; a listed entry keeps its stored fields, one without id is added, and
// remove drops a stored one.
func TestSdwan_InstancesByID(t *testing.T) {
	added := map[string]map[string]any{
		"connector_aruba_edge":     {"siteTag": "site-b", "credentialId": "cred-2", "hostName": "edge2", "accountName": "acct-b"},
		"connector_cisco_sdwan":    {"cloudInitFile": "#c2", "credentialId": "cred-2", "hostName": "edge2"},
		"connector_fortinet_sdwan": {"credentialId": "c-fg2", "hostName": "fg2", "version": "7.4.11"},
		"connector_versa_sdwan":    {"hostName": "v2", "version": "21.2.3-B"},
		"connector_vmware_sdwan":   {"credentialId": "cred-2", "hostName": "e2"},
	}
	for _, s := range sdwanSpecs {
		t.Run(s.name, func(t *testing.T) {
			list := s.f.Descriptor.IDMatchedArrays[0]
			e := newFamilyEnv(t, s)
			s.put(e, 7, s.stored())
			mustOK(t, e.call(s.name+"_update", map[string]any{"id": "7", "patch": map[string]any{list: []any{map[string]any{"id": 5000}, added[s.name]}}}))
			items := sentBody(t, e, "PUT")[list].([]any)
			first := items[0].(map[string]any)
			if len(items) != 2 || first["id"] != float64(5000) || first["hostName"] == nil || first["name"] != nil || items[1].(map[string]any)["id"] != nil {
				t.Fatalf("%s = %v", list, items)
			}
			e = newFamilyEnv(t, s)
			o := s.stored()
			o[list] = append(o[list].([]any), withFields(added[s.name], map[string]any{"id": float64(5001)}))
			s.put(e, 7, o)
			mustOK(t, e.call(s.name+"_update", map[string]any{"id": "7", "remove": map[string]any{list: []any{"5001"}}}))
			if items := sentBody(t, e, "PUT")[list].([]any); len(items) != 1 || items[0].(map[string]any)["id"] != float64(5000) {
				t.Fatalf("after remove %s = %v", list, items)
			}
		})
	}
}

// D3, Q-N12: the Versa shared keys are resent in the PUT and masked in the output.
func TestSdwan_VersaSecretsMasked(t *testing.T) {
	s := specNamed("connector_versa_sdwan")
	e := newFamilyEnv(t, s)
	s.put(e, 7, s.stored())
	r := sdwanUpdate(e, s, "7")
	mustOK(t, r)
	body := sentBody(t, e, "PUT")
	if body["localPublicSharedKey"] != "lk-stored" || body["remotePublicSharedKey"] != "rk-stored" {
		t.Fatalf("PUT body = %v", body)
	}
	if strings.Contains(r.Text, "lk-stored") || strings.Contains(r.Text, "rk-stored") {
		t.Fatalf("output shows a shared key:\n%s", r.Text)
	}
}

// D20: connector_versa_sdwan_create requires both shared keys as write-only
// fields, sent as given and never shown in output; no other schema names them.
func TestSdwan_VersaCreateKeys(t *testing.T) {
	e := newFamilyEnv(t, specNamed("connector_versa_sdwan"))
	var res struct {
		Tools []struct {
			Name        string          `json:"name"`
			InputSchema json.RawMessage `json:"inputSchema"`
		} `json:"tools"`
	}
	json.Unmarshal(call(t, e.srv, "tools/list", nil), &res)
	for _, tool := range res.Tools {
		if strings.HasPrefix(tool.Name, "connector_versa_sdwan_") && (tool.Name == "connector_versa_sdwan_create") != strings.Contains(string(tool.InputSchema), "SharedKey") {
			t.Errorf("%s schema: %s", tool.Name, tool.InputSchema)
		}
	}
	body := versaCreate("v1")
	delete(body, "remotePublicSharedKey")
	mustRefuse(t, e.call("connector_versa_sdwan_create", map[string]any{"body": body}), "invalid_arguments", "remotePublicSharedKey is required")
	r := e.call("connector_versa_sdwan_create", map[string]any{"body": versaCreate("v1")})
	mustOK(t, r)
	post := sentBody(t, e, "POST")
	if post["localPublicSharedKey"] != "lk-1" || post["remotePublicSharedKey"] != "rk-1" {
		t.Fatalf("POST %v", post)
	}
	if strings.Contains(r.Text, "lk-1") || strings.Contains(r.Text, "rk-1") {
		t.Fatalf("output shows a shared key:\n%s", r.Text)
	}
}

// type cannot change after create (CiscoSDWANIngressServiceImpl.java:205-207).
func TestSdwan_CiscoTypeImmutable(t *testing.T) {
	s := specNamed("connector_cisco_sdwan")
	e := newFamilyEnv(t, s)
	s.put(e, 7, s.stored())
	mustRefuse(t, e.call(s.name+"_update", map[string]any{"id": "7", "patch": map[string]any{"type": "CSR"}}), "invalid_arguments", "type cannot be changed")
	if e.portal.count("PUT") != 0 {
		t.Fatal("PUT sent")
	}
}

// GAP-06: the Cisco cloudInitFile carries the device registration token, so
// the delete preview and the update output mask it.
func TestSdwan_CiscoCloudInitMasked(t *testing.T) {
	s := specNamed("connector_cisco_sdwan")
	e := newFamilyEnv(t, s)
	o := s.stored()
	o["ciscoEdgeInfo"].([]any)[0].(map[string]any)["cloudInitFile"] = "otp-token-123"
	s.put(e, 7, o)
	for _, r := range []writeResult{sdwanUpdate(e, s, "7"), e.call(s.name+"_delete", map[string]any{"id": "7"})} {
		mustOK(t, r)
		if strings.Contains(r.Text, "otp-token-123") || !strings.Contains(r.Text, `"cloudInitFile": "***"`) {
			t.Fatalf("cloudInitFile not masked:\n%s", r.Text)
		}
	}
}
