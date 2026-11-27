package writes

import (
	"encoding/json"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// F5LBService is the F5 load balancer service write family (TPS /f5-lb-services).
var F5LBService = handlers.WriteFamily[ak.F5LBServiceConfig]{
	Name:         "service_f5_lb",
	Noun:         "F5 load balancer service",
	Descriptor:   ak.F5LBServiceDescriptor(),
	ImplicitNote: f5ServiceGroupsNote,
	CreateNote: "Each instance references an F5 instance credential and, for BRING_YOUR_OWN, an F5 registration credential, created in the Alkira portal (no tool lists or creates credentials); " +
		"globalCidrListId names a global CIDR list tagged F5LB in the service's CXP. Billable once provisioned. " +
		"ilbServiceGroupName and segmentOptions lbType take effect only while the tenant feature RELEASE-SERVICE-F5LB-ILB is on, or RELEASE-SERVICE-F5LB-AWS-VXLAN-ELB-ILB-SNAT for an AWS CXP; with it off the backend drops the value, and the result lists it as not kept.",
	UpdateNote: "Instances are matched by id: keep every stored instance in a patch on instances, or remove one with remove.instances. " +
		"cxp, globalCidrListId and each instance's credentials, licenseType and deployment.type cannot change. " +
		"Refused with no request while the service is FAILED, SCHEDULED or IN_PROGRESS, or while the network is FAILED and the service was provisioned before. " +
		"ilbServiceGroupName and segmentOptions lbType take effect only while the tenant feature RELEASE-SERVICE-F5LB-ILB is on, or RELEASE-SERVICE-F5LB-AWS-VXLAN-ELB-ILB-SNAT for an AWS CXP; with it off the backend drops the value, and the result lists it as not kept.",
	DeleteNote: "A service with vServer endpoints, or whose service group a policy uses, is refused by the backend with its reason. A deployed service keeps its credentials until the deletion is provisioned from the Alkira portal.",
}

// serviceGroupNote names the service group, with its implicit group, that
// the backend deletes with an F5 or Infoblox service.
func serviceGroupNote(_ string, raw []byte) string {
	var s struct {
		ServiceGroupName string `json:"serviceGroupName"`
	}
	if json.Unmarshal(raw, &s) != nil || s.ServiceGroupName == "" {
		return ""
	}
	return `its service group "` + s.ServiceGroupName + `"`
}

// f5ServiceGroupsNote adds the ILB service group an F5 service with an ILB
// segment holds (F5LBServiceLogicImpl.java:1310-1324).
func f5ServiceGroupsNote(name string, raw []byte) string {
	var s struct {
		ILB string `json:"ilbServiceGroupName"`
	}
	note := serviceGroupNote(name, raw)
	if json.Unmarshal(raw, &s) == nil && s.ILB != "" && note != "" {
		note += ` and its ILB service group "` + s.ILB + `"`
	}
	return note
}
