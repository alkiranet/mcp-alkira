// Copyright (C) 2021-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
)

// A summary of a NAT rule (other fields are ignored)
type NATPolicyRuleSummary struct {
	Id        json.Number `json:"id"`
	Name      string      `json:"name"`
	Enabled   bool        `json:"enabled"`
	Direction string      `json:"direction,omitempty"`
}

type NATPolicyRule struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Id          json.Number   `json:"id,omitempty"`
	Enabled     bool          `json:"enabled"`
	Match       NATRuleMatch  `json:"match"`
	Action      NATRuleAction `json:"action"`
	Category    string        `json:"category"`
	Direction   string        `json:"direction,omitempty"`
}

type NATRuleMatch struct {
	SourcePrefixes      []string `json:"sourcePrefixes,omitempty"`
	SourcePrefixListIds []int    `json:"sourcePrefixListIds,omitempty"`
	DestPrefixes        []string `json:"destPrefixes,omitempty"`
	DestPrefixListIds   []int    `json:"destPrefixListIds,omitempty"`
	SourcePortList      []string `json:"sourcePortList,omitempty"`
	DestPortList        []string `json:"destPortList,omitempty"`
	Protocol            string   `json:"protocol"`
}

type NATRuleAction struct {
	SourceAddressTranslation      NATRuleActionSrcTranslation `json:"sourceAddressTranslation"`
	DestinationAddressTranslation NATRuleActionDstTranslation `json:"destinationAddressTranslation"`
	Egress                        EgressAction                `json:"egress"`
}

type NATRuleActionSrcTranslation struct {
	TranslationType         string                `json:"translationType"`
	TranslatedPrefixes      []string              `json:"translatedPrefixes,omitempty"`
	TranslatedPrefixListIds []int                 `json:"translatedPrefixListIds,omitempty"`
	Bidirectional           *bool                 `json:"bidirectional,omitempty"`
	MatchAndInvalidate      *bool                 `json:"matchAndInvalidate,omitempty"`
	RoutingOptions          NATRuleRoutingOptions `json:"routingOptions,omitempty"`
}

type NATRuleActionDstTranslation struct {
	TranslationType            string                `json:"translationType"`
	TranslatedPrefixes         []string              `json:"translatedPrefixes,omitempty"`
	TranslatedPrefixListIds    []int                 `json:"translatedPrefixListIds,omitempty"`
	TranslatedPortList         []string              `json:"translatedPortList,omitempty"`
	TranslatedPolicyFqdnListId int                   `json:"translatedPolicyFqdnListId,omitempty"`
	Bidirectional              *bool                 `json:"bidirectional,omitempty"`
	AdvertiseToConnector       *bool                 `json:"advertiseToConnector,omitempty"`
	RoutingOptions             NATRuleRoutingOptions `json:"routingOptions,omitempty"`
}

type NATRuleRoutingOptions struct {
	TrackPrefixes                  []string `json:"trackPrefixes,omitempty"`
	TrackPrefixListIds             []int    `json:"trackPrefixListIds,omitempty"`
	InvalidateRoutingTrackPrefixes *bool    `json:"invalidateRoutingTrackPrefixes,omitempty"`
}

type EgressAction struct {
	IpType string `json:"ipType"`
}

// NewNATRule new NAT rule
func NewNATRule(ac *AlkiraClient) *AlkiraApi[NATPolicyRuleSummary] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/nat-rules", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[NATPolicyRuleSummary]{ac, uri, PaginationOn}
	return api
}
