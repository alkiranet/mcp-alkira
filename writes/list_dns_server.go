package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// DnsServerList is the DNS server list write family (TPS /dns-server-lists).
var DnsServerList = handlers.WriteFamily[ak.DnsServerListConfig]{
	Name:       "list_dns_server",
	Noun:       "DNS server list",
	Descriptor: ak.DnsServerListDescriptor(),
	CreateNote: "segment is the segment name. The number of dnsServerIps is bounded by the tenant's resource limits (one by default), not by this schema.",
	UpdateNote: "A list whose segment was deleted reads back with segmentId and without segment. " +
		"An update that does not set segment succeeds and keeps that stale segmentId; set segment to an existing segment name in the patch to point the list at it.",
	DeleteNote: "A DNS server list used by a policy FQDN list is refused as in use. " +
		"A segment delete is not checked against DNS server lists, so a list can hold a segment id that no longer exists.",
}
