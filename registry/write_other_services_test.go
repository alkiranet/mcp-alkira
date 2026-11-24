package registry

import (
	"encoding/json"
	"maps"
	"strings"
	"testing"

	"github.com/alkiranet/mcp-alkira/writes"
)

func init() {
	familySpecs = append(familySpecs, otherServiceSpecs...)
	rows := func(k, guard string) map[string]string {
		return map[string]string{
			scCRUD:              "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode; " + k,
			scExternalDelete:    "TestFamilies_NotFound TestFamilies_DeleteWithCode",
			scRetryOn429:        "TestWrite_ErrorMapping; CG-W5 (client-go)",
			scImportNonexistent: "TestFamilies_NotFound",
			scWaveCoalescing:    "MA-W13 (ST-03)",
			scRequestFailed:     "MA-W18 (ST-03); TestFamilies_InvalidArgumentsSendNothing",
			scPreflightBlocked:  guard + "; MA-W13 (ST-03)",
			scMissingRequestID:  "MA-W12 (ST-03)",
			scTimeoutPartial:    "TestWrite_Outcomes; MA-W18 (ST-03)",
			scSoftDeleteOff:     "TestOtherServices_MarkedRow TestFamilies_DeletePreview",
		}
	}
	zscaler := rows("K46.1-.6", "TestOtherServices_StateGuard (GET /services row state)")
	zscaler[scCRUD] = "TestFamilies_CreateRequestShape TestFamilies_UpdateMerge TestFamilies_DeleteWithCode TestOtherServices_ZscalerCreateKey; K46.1-.6"
	ra := rows("K26.1-.6", "TestOtherServices_StateGuard (TPS runs no state check on a template)")
	ra[scSoftDeleteOff] = "TestFamilies_DocState TestFamilies_DeletePreview"
	maps.Copy(scenarioRows, map[string]map[string]string{
		"service_f5_lb":           rows("K44.1-.6", "TestOtherServices_StateGuard (GET /services row state)"),
		"service_infoblox":        rows("K45.1-.6", "TestOtherServices_StateGuard (entity FAILED)"),
		"service_zscaler":         zscaler,
		"connector_remote_access": ra,
	})
}

const (
	svcNotFound = `{"code":"AK000003","message":"service entity with id: 404 is not present in the data store"}`
	svcInUse    = `{"code":"AK000001","message":"SERVICE with id 7 is still referred externally"}`
)

// noName drops the shared cases whose text the family's name rule changes.
func noName(charset string) map[string]toolCase {
	return map[string]toolCase{
		"trailing space": {}, // descriptions allow a trailing space (TPS @Pattern)
		"leading space":  {"_create", map[string]any{"body": map[string]any{"name": " a"}}, "outside the " + charset + " character set"},
		"charset":        {"_create", map[string]any{"body": map[string]any{"name": "a/b"}}, "outside the " + charset + " character set"},
	}
}

func f5Create(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "USEAST-AZURE-1", "serviceGroupName": "sg", "segments": []any{"seg"}, "globalCidrListId": 355,
		"segmentOptions": map[string]any{"seg": map[string]any{"elbNicCount": 1}},
		"instances": []any{map[string]any{"name": "i1", "credentialId": "c1", "licenseType": "BRING_YOUR_OWN", "version": "17.1.103000",
			"deployment": map[string]any{"type": "LTM_DNS"}, "hostNameFqdn": "a.example.com"}}}
}

func infobloxCreate(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "USEAST-AZURE-1", "segments": []any{"seg"}, "globalCidrListId": 353, "serviceGroupName": "sg",
		"gridMaster": map[string]any{"name": "grid1", "gridMasterCredentialId": "g1", "sharedSecretCredentialId": "s1"},
		"instances":  []any{map[string]any{"hostName": "ib1.example.com", "credentialId": "c1", "model": "TE-V926", "type": "MASTER", "version": "9.0.6"}}}
}

func raCreate(name string) map[string]any {
	return map[string]any{"name": name, "segments": []any{"seg"},
		"segmentOptions": []any{map[string]any{"name": "seg", "userGroupMappings": []any{map[string]any{"name": "ug", "splitTunneling": false,
			"cxpToSubnetsMapping": []any{map[string]any{"cxp": "USEAST-AZURE-1", "subnets": []any{"172.30.0.0/23"}}}}}}},
		"authenticationOptions": map[string]any{"supportedModes": []any{"LOCAL"}}, "advancedOptions": map[string]any{},
		"arguments": []any{map[string]any{"cxp": "USEAST-AZURE-1"}}}
}

var otherServiceSpecs = []familySpec{
	{
		name: "service_f5_lb", noun: "F5 load balancer service", f: family(writes.F5LBService), coll: "f5-lb-services", service: "F5LB",
		notFound: fault{status: 404, body: svcNotFound}, invalid: noName("slug"),
		stored: func() map[string]any {
			return map[string]any{"name": "f7", "description": "seven", "state": "PENDING", "cxp": "USEAST-AZURE-1", "size": "SMALL", "serviceGroupName": "sg7",
				"elbServiceGroupName": "sg7", "billingTags": []any{}, "segments": []any{"seg"}, "providerName": "AZURE", "globalCidrListId": 355,
				"segmentOptions": map[string]any{"seg": map[string]any{"segmentId": 3, "elbNicCount": 1}}, "tunnelProtocol": "VXLAN", "implicitGroupId": 60,
				"instances": []any{map[string]any{"id": 11, "name": "i1", "internalName": "svci-1", "credentialId": "c1", "licenseType": "BRING_YOUR_OWN",
					"version": "17.1.103000", "deployment": map[string]any{"option": "TWO_BOOT_LOCATION", "type": "LTM_DNS"}, "hostNameFqdn": "a.example.com"}}}
		},
		create:   f5Create,
		readOnly: []string{"providerName", "elbServiceGroupName", "implicitGroupId", "state"},
		creates: []createCase{
			{nil, f5Create("f1"), withFields(f5Create("f1"), map[string]any{"size": "SMALL"}), nil},
			{[]string{"BILLING-TAGS"}, withFields(f5Create("f1"), map[string]any{"billingTags": []any{7}}),
				withFields(f5Create("f1"), map[string]any{"size": "SMALL", "billingTags": []any{7}}), []string{"BILLING-TAGS"}},
		},
		inUse: []inUseCase{{fault{status: 409, body: svcInUse}, "is still referred externally"}},
		previews: []previewCase{
			{"PENDING", []string{`Delete preview for F5 load balancer service "f7" (id 7).`,
				"Deployed: no, never deployed (state PENDING).", `Also deleted with it: its service group "sg7".`, "The delete now: removes it at once."}},
			{"SUCCESS", []string{"Deployed: yes (state SUCCESS).", `Also deleted with it: its service group "sg7".`,
				"marks it for deletion; it keeps carrying traffic, with its name and references, until the deletion is provisioned from the Alkira portal"}},
		},
	},
	{
		name: "service_infoblox", noun: "Infoblox service", f: family(writes.ServiceInfoblox), coll: "infoblox-services", service: "INFOBLOX", nameMax: 64,
		notFound: fault{status: 404, body: svcNotFound}, invalid: noName("hyphen"),
		stored: func() map[string]any {
			return map[string]any{"name": "ib7", "description": "seven", "state": "PENDING", "size": "SMALL", "cxp": "USEAST-AZURE-1", "segments": []any{"seg"},
				"billingTags": []any{}, "globalCidrListId": 353, "anycast": map[string]any{"enabled": false, "ips": []any{}, "backupCxps": []any{}},
				"serviceGroupId": 425, "serviceGroupImplicitGroupId": 61, "serviceGroupName": "sg7", "licenseType": "BRING_YOUR_OWN",
				"gridMaster": map[string]any{"name": "grid1", "external": false, "gridMasterCredentialId": "g1", "sharedSecretCredentialId": "s1"},
				"instances": []any{map[string]any{"id": 10, "name": "h", "state": "PENDING", "hostName": "h.example.com", "credentialId": "c1",
					"model": "TE-V926", "type": "MASTER", "version": "9.0.6"}}}
		},
		create:   infobloxCreate,
		readOnly: []string{"state", "serviceGroupId", "serviceGroupImplicitGroupId", "licenseType"},
		creates: []createCase{{nil, infobloxCreate("ib1"),
			withFields(infobloxCreate("ib1"), map[string]any{"size": "SMALL", "anycast": map[string]any{"enabled": false}}), nil}},
		inUse: []inUseCase{{fault{status: 409, body: svcInUse}, "is still referred externally"}},
		previews: []previewCase{
			{"PENDING", []string{`Delete preview for Infoblox service "ib7" (id 7).`, "Deployed: no, never deployed (state PENDING).",
				`Also deleted with it: its service group "sg7".`, "The delete now: removes it at once."}},
			{"SUCCESS", []string{"Deployed: yes (state SUCCESS).", `Also deleted with it: its service group "sg7".`,
				"marks it for deletion; it keeps carrying traffic, with its name and references, until the deletion is provisioned from the Alkira portal"}},
		},
	},
	{
		name: "service_zscaler", noun: "Zscaler internet access service", f: family(writes.ServiceZscaler), coll: "zscaler-internet-access-services", service: "ZIA", nameMax: 60,
		notFound: fault{status: 404, body: svcNotFound}, invalid: noName("hyphen"),
		stored:   zscalerStored,
		create:   zscalerCreate,
		readOnly: []string{"internalName", "segmentIds", "state"},
		creates: []createCase{{nil, zscalerCreate("z1"), withFields(zscalerCreate("z1"), map[string]any{"size": "SMALL", "tunnelType": "IPSEC",
			"ipsecConfiguration": map[string]any{"localFqdnId": "a@example.com", "preSharedKey": "caller-key-1", "ikeEncryptionAlgorithm": "AES256CBC",
				"ikeIntegrityAlgorithm": "SHA256", "ikeDhGroupNumber": "MODP1024", "espEncryptionAlgorithm": "NULL", "espIntegrityAlgorithm": "MD5",
				"espDhGroupNumber": "MODP1024", "healthCheckType": "IKE_STATUS"}}), nil}},
		inUse: []inUseCase{{fault{status: 409, body: svcInUse}, "is still referred externally"}},
		previews: []previewCase{
			{"PENDING", []string{`Delete preview for Zscaler internet access service "z7" (id 7).`,
				"Deployed: no, never deployed (state PENDING).", "Also deleted with it: its firewall zones.", "The delete now: removes it at once."}},
			{"SUCCESS", []string{"Deployed: yes (state SUCCESS).", "Also deleted with it: its firewall zones.",
				"marks it for deletion; it keeps carrying traffic, with its name and references, until the deletion is provisioned from the Alkira portal"}},
		},
	},
	{
		name: "connector_remote_access", noun: "remote access connector template", f: family(writes.ConnectorRemoteAccessTemplate),
		coll:     "alkira-remote-access-connector-templates",
		notFound: fault{status: 404, body: `{"code":"AK000003","message":"Could not find connector template with id : 404"}`},
		invalid:  map[string]toolCase{"trailing space": {}, "charset": {}},
		stored: func() map[string]any {
			return map[string]any{"name": "r7", "description": "seven", "internalName": "rct-7", "segments": []any{"seg"}, "segmentIds": []any{3},
				"segmentOptions": []any{map[string]any{"name": "seg", "segmentId": 3, "userGroupMappings": []any{map[string]any{"name": "ug", "splitTunneling": false,
					"cxpToSubnetsMapping": []any{map[string]any{"cxp": "USEAST-AZURE-1", "subnets": []any{"172.30.0.0/23"}}}}}}},
				"authenticationOptions": map[string]any{"supportedModes": []any{"LOCAL"}},
				"advancedOptions":       map[string]any{"enableDynamicRegionMapping": true, "maxActiveUsersThreshold": 80, "udpPort": 1194},
				"arguments":             []any{map[string]any{"cxp": "USEAST-AZURE-1", "size": "SMALL"}}, "clientCertificateId": 31, "openVPNStaticKeysId": 30,
				"docState": "ADDED", "state": "PENDING", "certificatesExpireAt": 1890121655506, "referencedPrefixListIds": []any{}}
		},
		create:   raCreate,
		readOnly: []string{"state", "docState", "internalName", "clientCertificateId", "openVPNStaticKeysId", "certificatesExpireAt", "referencedPrefixListIds"},
		creates: []createCase{{nil, raCreate("ra1"), withFields(raCreate("ra1"), map[string]any{
			"advancedOptions": map[string]any{"enableDynamicRegionMapping": true, "maxActiveUsersThreshold": 80}}), nil}},
		inUse: []inUseCase{{fault{status: 400, body: `{"code":"AK000002","message":"Cannot remove all remote access connectors for user group 'ug'. User group 'ug' is associated with firewall zones"}`},
			"is associated with firewall zones"}},
		previews: []previewCase{{"PENDING", []string{`Delete preview for remote access connector template "r7" (id 7).`,
			"Deployed: no, never deployed (state PENDING).", `Also deleted with it: its connectors "remote-access-connector-USEAST-AZURE-1-r7".`}}},
	},
}

func zscalerCreate(name string) map[string]any {
	return map[string]any{"name": name, "cxp": "USEAST-AZURE-1", "segments": []any{"seg"}, "primaryPublicEdgeIp": "198.51.100.10", "secondaryPublicEdgeIp": "198.51.100.11",
		"ipsecConfiguration": map[string]any{"localFqdnId": "a@example.com", "preSharedKey": "caller-key-1"}}
}

func zscalerStored() map[string]any {
	return map[string]any{"name": "z7", "description": "seven", "state": "PENDING", "internalName": "svc-zia-7", "cxp": "USEAST-AZURE-1", "size": "SMALL", "billingTags": []any{},
		"segments": []any{"seg"}, "segmentIds": []any{3}, "primaryPublicEdgeIp": "198.51.100.10", "secondaryPublicEdgeIp": "198.51.100.11", "tunnelType": "IPSEC",
		"ipsecConfiguration": map[string]any{"localFqdnId": "a@example.com", "preSharedKey": "stored-key-1", "ikeEncryptionAlgorithm": "AES256CBC",
			"ikeIntegrityAlgorithm": "SHA256", "ikeDhGroupNumber": "MODP1024", "espEncryptionAlgorithm": "NULL", "espIntegrityAlgorithm": "MD5",
			"espDhGroupNumber": "MODP1024", "healthCheckType": "IKE_STATUS"}}
}

// D20, MA-W15, MA-W20: the Zscaler create requires the pre-shared key as a
// write-only field; no other schema names it, an update resends the stored
// key, and every output masks it.
func TestOtherServices_ZscalerCreateKey(t *testing.T) {
	e := newFamilyEnv(t, specNamed("service_zscaler"))
	for name, schema := range toolSchemas(t, e.srv) {
		b, _ := json.Marshal(schema.Schema)
		if strings.HasPrefix(name, "service_zscaler_") && (name == "service_zscaler_create") != strings.Contains(string(b), "preSharedKey") {
			t.Errorf("%s schema: %s", name, b)
		}
	}
	body := zscalerCreate("z1")
	delete(body["ipsecConfiguration"].(map[string]any), "preSharedKey")
	mustRefuse(t, e.call("service_zscaler_create", map[string]any{"body": body}), "invalid_arguments", "ipsecConfiguration.preSharedKey is required")
	r := e.call("service_zscaler_create", map[string]any{"body": zscalerCreate("z1")})
	mustOK(t, r)
	if k := sentBody(t, e, "POST")["ipsecConfiguration"].(map[string]any)["preSharedKey"]; k != "caller-key-1" || strings.Contains(r.Text, "caller-key-1") {
		t.Fatalf("POST key %v, output:\n%s", k, r.Text)
	}

	e = newFamilyEnv(t, specNamed("service_zscaler"))
	e.portal.put(7, zscalerStored())
	r = e.call("service_zscaler_update", map[string]any{"id": "7", "patch": map[string]any{"ipsecConfiguration": map[string]any{"ikeDhGroupNumber": "MODP2048"}}})
	mustOK(t, r)
	ipsec := sentBody(t, e, "PUT")["ipsecConfiguration"].(map[string]any)
	if ipsec["preSharedKey"] != "stored-key-1" || ipsec["ikeDhGroupNumber"] != "MODP2048" {
		t.Fatalf("PUT ipsecConfiguration = %v", ipsec)
	}
	p := e.call("service_zscaler_delete", map[string]any{"id": "7"})
	for _, out := range []string{r.Text, p.Text} {
		if strings.Contains(out, "stored-key-1") || !strings.Contains(out, "***") {
			t.Fatalf("key not masked:\n%s", out)
		}
	}
	mustRefuse(t, e.call("service_zscaler_update", map[string]any{"id": "7", "patch": map[string]any{"ipsecConfiguration": map[string]any{"preSharedKey": "x"}}}),
		"invalid_arguments", "preSharedKey")
	if e.portal.count("PUT") != 1 {
		t.Fatal("key patch sent")
	}
}

// MA-W5: F5, Zscaler and Infoblox services are refused on the state of
// their GET /services row with no PUT; a remote access template is not
// checked, as TPS does not check it.
func TestOtherServices_StateGuard(t *testing.T) {
	for _, name := range []string{"service_f5_lb", "service_zscaler", "service_infoblox", "connector_remote_access"} {
		t.Run(name, func(t *testing.T) {
			s := specNamed(name)
			e := newFamilyEnv(t, s)
			o := s.stored()
			o["state"] = "FAILED"
			e.portal.put(7, o)
			r := e.call(name+"_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}})
			if name == "connector_remote_access" {
				mustOK(t, r)
				return
			}
			mustRefuse(t, r, "state_refused", "FAILED")
			if e.portal.count("PUT") != 0 {
				t.Fatalf("%d PUTs", e.portal.count("PUT"))
			}
		})
	}
}

// BC-R2: a sibling family's id answers 200 in this family's shape without its
// family field; update and delete refuse it with no write.
func TestOtherServices_FamilyCheck(t *testing.T) {
	for _, s := range otherServiceSpecs {
		t.Run(s.name, func(t *testing.T) {
			e := newFamilyEnv(t, s)
			e.portal.put(7, map[string]any{"name": "pan7", "cxp": "USEAST-AZURE-1", "segments": []any{"seg"}, "docState": "ADDED"})
			mustRefuse(t, e.call(s.name+"_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}), "family_mismatch", "belongs to another family")
			mustRefuse(t, e.call(s.name+"_delete", map[string]any{"id": "7"}), "family_mismatch", "belongs to another family")
			if e.portal.count("PUT")+e.portal.count("DELETE") != 0 {
				t.Fatal("write sent")
			}
		})
	}
}

// MA-W12, MA-W50: a name another service type holds is refused by the
// backend with 409; the tool reports a duplicate without the backend text.
func TestOtherServices_BackendDuplicate(t *testing.T) {
	e := newFamilyEnv(t, specNamed("service_f5_lb"))
	e.portal.fail("POST", e.portal.base(), fault{status: 409, body: `{"code":"AK000001","message":"Service with name : f1 already exists in network : 7"}`})
	mustRefuse(t, e.call("service_f5_lb_create", map[string]any{"body": f5Create("f1")}), "duplicate", `an F5 load balancer service named "f1" already exists.`)
}

// MA-W4: F5 instances are matched by id; a patch that leaves out a stored
// instance or changes its deployment type is refused with no PUT.
func TestOtherServices_F5Instances(t *testing.T) {
	s := specNamed("service_f5_lb")
	e := newFamilyEnv(t, s)
	e.portal.put(7, s.stored())
	inst := f5Create("x")["instances"]
	mustRefuse(t, e.call("service_f5_lb_update", map[string]any{"id": "7", "patch": map[string]any{"instances": inst}}), "invalid_arguments", "id 11 is missing")
	mustRefuse(t, e.call("service_f5_lb_update", map[string]any{"id": "7", "patch": map[string]any{"instances": []any{
		map[string]any{"id": 11, "deployment": map[string]any{"type": "BEST"}}}}}), "invalid_arguments", "deployment.type")
	if e.portal.count("PUT") != 0 {
		t.Fatal("PUT sent")
	}
	mustOK(t, e.call("service_f5_lb_update", map[string]any{"id": "7", "patch": map[string]any{"instances": []any{
		map[string]any{"id": 11, "version": "17.1.103001"}}}}))
	b, _ := json.Marshal(sentBody(t, e, "PUT")["instances"])
	if !strings.Contains(string(b), `"id":11`) || !strings.Contains(string(b), `"credentialId":"c1"`) || strings.Contains(string(b), "TWO_BOOT_LOCATION") {
		t.Fatalf("PUT instances = %s", b)
	}
}

// MA-W16: a service marked for deletion answers the plain GET with not found
// and the include-marked GET with the row; update and delete refuse it with
// no write. The F5 preview also names the ILB service group.
func TestOtherServices_MarkedRow(t *testing.T) {
	for _, name := range []string{"service_f5_lb", "service_infoblox", "service_zscaler"} {
		t.Run(name, func(t *testing.T) {
			s := specNamed(name)
			e := newFamilyEnv(t, s)
			e.portal.put(7, s.stored())
			for _, c := range []struct {
				tool string
				args map[string]any
			}{{"_update", map[string]any{"id": "7", "patch": map[string]any{"description": "x"}}}, {"_delete", map[string]any{"id": "7"}}} {
				e.portal.fail("GET", e.portal.base()+"/7", fault{status: 404, body: svcNotFound})
				mustRefuse(t, e.call(name+c.tool, c.args), "state_refused", "deletion pending provisioning")
			}
			if e.portal.count("PUT")+e.portal.count("DELETE") != 0 {
				t.Fatal("write sent")
			}
		})
	}
	s := specNamed("service_f5_lb")
	e := newFamilyEnv(t, s)
	e.portal.put(7, withFields(s.stored(), map[string]any{"ilbServiceGroupName": "ilb7"}))
	if r := e.call("service_f5_lb_delete", map[string]any{"id": "7"}); !strings.Contains(r.Text, `its service group "sg7" and its ILB service group "ilb7"`) {
		t.Fatalf("preview:\n%s", r.Text)
	}
}

// INT-04, MA-W50: a name any service type holds refuses the F5, Infoblox and
// Zscaler create with no POST; the guard reads GET /services.
func TestOtherServices_CrossTypeDuplicate(t *testing.T) {
	for _, name := range []string{"service_f5_lb", "service_infoblox", "service_zscaler"} {
		t.Run(name, func(t *testing.T) {
			s := specNamed(name)
			e := newFamilyEnv(t, s)
			e.portal.otherServices = []map[string]any{{"id": 31, "name": "dup", "type": "PAN", "state": "SUCCESS"}}
			mustRefuse(t, e.call(name+"_create", map[string]any{"body": s.createBody("dup")}), "duplicate",
				`a service named "dup" already exists (id 31, type PAN); service names are unique across every service type in the network.`)
			if e.portal.count("POST") != 0 {
				t.Fatal("POST sent")
			}
		})
	}
}
