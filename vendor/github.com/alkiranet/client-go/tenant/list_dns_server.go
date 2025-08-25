// Copyright (C) 2023-2025 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
)

type DnsServerList struct {
	Id           json.Number `json:"id,omitempty"`
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	DnsServerIps []string    `json:"dnsServerIps"`
	Segment      string      `json:"segment"`
}

// NewDnsServerList new DNS server list
func NewDnsServerList(ac *AlkiraClient) *AlkiraApi[DnsServerList] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/dns-server-lists", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[DnsServerList]{ac, uri}
	return api
}
