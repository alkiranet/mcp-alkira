package writes

import (
	"encoding/json"
	"strings"

	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// ConnectorRemoteAccessTemplate is the remote access connector template
// write family (TPS /alkira-remote-access-connector-templates).
var ConnectorRemoteAccessTemplate = handlers.WriteFamily[ak.ConnectorRemoteAccessTemplateConfig]{
	Name:       "connector_remote_access",
	Noun:       "remote access connector template",
	Descriptor: ak.ConnectorRemoteAccessTemplateDescriptor(),
	// The backend creates one connector per arguments[].cxp, named
	// remote-access-connector-<cxp>-<template name>, and deletes them with the template.
	ImplicitNote: func(name string, raw []byte) string {
		var t struct {
			Arguments []struct {
				CXP string `json:"cxp"`
			} `json:"arguments"`
		}
		if json.Unmarshal(raw, &t) != nil || len(t.Arguments) == 0 {
			return ""
		}
		names := make([]string, len(t.Arguments))
		for i, a := range t.Arguments {
			names[i] = `"remote-access-connector-` + a.CXP + "-" + name + `"`
		}
		return "its connectors " + strings.Join(names, ", ")
	},
	CreateNote: "The backend creates one remote access connector per arguments entry and generates the certificates and keys. " +
		"Names are unique ignoring case. Each user group mapping names an existing user group; an LDAP mode takes ldapSettings with the id of a credential created in the Alkira portal. Billable once provisioned. " +
		"advancedOptions cxpToNameServerMapping, fallbackToTcp, alwaysOnEnabled, preLogonEnabled and zeroTrustAccessPosture each depend on a tenant feature; with it off the backend drops the value, and the result lists it as not kept. deviceCertificate is sent only while the tenant feature ZTNA-PHASE-2 is on, and the result names it when left out. advancedOptions udpPort and tcpPort are set by the backend and cannot be given.",
	UpdateNote: "A patch on segmentOptions or arguments replaces the whole list (RFC 7386); send every entry to keep. " +
		"The name cannot change once a connector of the template has deployed. " +
		"advancedOptions cxpToNameServerMapping, fallbackToTcp, alwaysOnEnabled, preLogonEnabled and zeroTrustAccessPosture each depend on a tenant feature; with it off the backend drops the value, and the result lists it as not kept. deviceCertificate is sent only while the tenant feature ZTNA-PHASE-2 is on, and the result names it when left out. advancedOptions udpPort and tcpPort are set by the backend and cannot be given.",
	DeleteNote: "The delete also removes the template's connectors. A delete that would drop the last connector of a user group used by a policy or firewall zone is refused.",
}
