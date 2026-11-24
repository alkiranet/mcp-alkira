// Copyright (C) 2021-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
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

// NatRuleConfig is the write model of a NAT rule (TPS model/NATRule.java).
// A PUT replaces the stored rule, so the merge resends the full GET body;
// the deprecated top-level fields stay unmodelled and are resent unchanged.
// The prefix and prefix-list lists of match, translations and routing
// options carry no maxitems: their limit of 1 is a tenant resource limit
// (NATRule.java:121-141, :199-233, :286-293 @SoftLimit); the port lists are
// held to 1 by @Size.
type NatRuleConfig struct {
	ID json.Number `json:"id,omitempty" ak:"readonly"`
	// The column holds 255 characters; 256 passes the bean and fails the insert with a 500.
	Name        string  `json:"name" ak:"required,maxlen=255,trimmed,charset=name,src=NATRule.java:63-66|NATRuleEntity.java:52|PHASE-NAT-PROBE-RESULTS.md:34-37"`
	Description *string `json:"description,omitempty" ak:"maxlen=255,trimmed,charset=name,src=NATRule.java:68-71|NATRuleEntity.java:54|PHASE-NAT-PROBE-RESULTS.md:169-171"`
	Enabled     *bool   `json:"enabled,omitempty" ak:"required,src=NATRuleServiceImpl.java:471-473"`
	Category    *string `json:"category,omitempty" ak:"default=DEFAULT,src=NATRule.java:75|CreateNATPolicyRule.js:168|:668"`
	// The backend clears direction while RELEASE-NAT-POLICY-OUTBOUND is off
	// (NATRuleServiceImpl.java:487-488) and refuses a direction change only on
	// a PUT that keeps the name (:482-483); a rename bypasses it.
	Direction *string              `json:"direction,omitempty" ak:"immutable,enum=INBOUND|OUTBOUND,flag=RELEASE-NAT-POLICY-OUTBOUND,src=NATRule.java:80|NATRuleServiceImpl.java:482-488|PHASE-NAT-PROBE-RESULTS.md:116-117"`
	Match     *NatRuleConfigMatch  `json:"match,omitempty" ak:"required,src=NATRule.java:94-95|PHASE-NAT-PROBE-RESULTS.md:125-126"`
	Action    *NatRuleConfigAction `json:"action,omitempty" ak:"required,src=NATRule.java:97-98|PHASE-NAT-PROBE-RESULTS.md:127"`

	SegmentID           json.Number `json:"segmentId,omitempty" ak:"readonly"`
	LastConfigUpdatedAt json.Number `json:"lastConfigUpdatedAt,omitempty" ak:"readonly"`
	UpdatedBy           string      `json:"updatedBy,omitempty" ak:"readonly"`
	DocState            string      `json:"docState,omitempty" ak:"readonly"`
}

// NatRuleConfigMatch is NATRule.Match.
type NatRuleConfigMatch struct {
	SourcePrefixes      []string `json:"sourcePrefixes,omitempty"`
	SourcePrefixListIDs []int64  `json:"sourcePrefixListIds,omitempty"`
	SourcePortList      []string `json:"sourcePortList,omitempty" ak:"maxitems=1,src=NATRule.java:131|PHASE-NAT-PROBE-RESULTS.md:105"`
	DestPrefixes        []string `json:"destPrefixes,omitempty"`
	DestPrefixListIDs   []int64  `json:"destPrefixListIds,omitempty"`
	DestPortList        []string `json:"destPortList,omitempty" ak:"maxitems=1,src=NATRule.java:145|PHASE-NAT-PROBE-RESULTS.md:106"`
	Protocol            *string  `json:"protocol,omitempty" ak:"enum=icmp|tcp|udp|any,src=NATRule.java:148-150"`
}

// NatRuleConfigAction is NATRule.Action.
type NatRuleConfigAction struct {
	SourceAddressTranslation      *NatRuleConfigSrcTranslation `json:"sourceAddressTranslation,omitempty"`
	DestinationAddressTranslation *NatRuleConfigDstTranslation `json:"destinationAddressTranslation,omitempty"`
	Egress                        *NatRuleConfigEgress         `json:"egress,omitempty"`
}

// NatRuleConfigSrcTranslation is NATRule.SourceAddressTranslation.
type NatRuleConfigSrcTranslation struct {
	TranslationType         *string                      `json:"translationType,omitempty" ak:"enum=STATIC_IP|DYNAMIC_IP_AND_PORT|NONE,src=NATRule.java:263-265"`
	TranslatedPrefixes      []string                     `json:"translatedPrefixes,omitempty"`
	TranslatedPrefixListIDs []int64                      `json:"translatedPrefixListIds,omitempty"`
	Bidirectional           *bool                        `json:"bidirectional,omitempty"`
	MatchAndInvalidate      *bool                        `json:"matchAndInvalidate,omitempty"`
	RoutingOptions          *NatRuleConfigRoutingOptions `json:"routingOptions,omitempty"`
}

// NatRuleConfigDstTranslation is NATRule.DestinationAddressTranslation.
type NatRuleConfigDstTranslation struct {
	TranslationType            *string                      `json:"translationType,omitempty" ak:"enum=STATIC_IP|STATIC_IP_AND_PORT|STATIC_PORT|NONE,src=NATRule.java:266-267"`
	TranslatedPrefixes         []string                     `json:"translatedPrefixes,omitempty"`
	TranslatedPrefixListIDs    []int64                      `json:"translatedPrefixListIds,omitempty"`
	TranslatedPortList         []string                     `json:"translatedPortList,omitempty"`
	TranslatedPolicyFqdnListID *int64                       `json:"translatedPolicyFqdnListId,omitempty"`
	Bidirectional              *bool                        `json:"bidirectional,omitempty"`
	AdvertiseToConnector       *bool                        `json:"advertiseToConnector,omitempty"`
	RoutingOptions             *NatRuleConfigRoutingOptions `json:"routingOptions,omitempty"`
}

// NatRuleConfigRoutingOptions is NATRule.RoutingOptions.
type NatRuleConfigRoutingOptions struct {
	TrackPrefixes                  []string `json:"trackPrefixes,omitempty"`
	TrackPrefixListIDs             []int64  `json:"trackPrefixListIds,omitempty"`
	InvalidateRoutingTrackPrefixes *bool    `json:"invalidateRoutingTrackPrefixes,omitempty"`
}

// NatRuleConfigEgress is NATRule.EgressAction.
type NatRuleConfigEgress struct {
	IPType *string `json:"ipType,omitempty"`
}

// natRuleDescriptor: GET and PUT answer a missing id with 404 AK000003 "NAT
// rule <id> not found" and DELETE with 404 AK000003 (NATRuleServiceImpl.java:322,
// :330, :438; PHASE-NAT-PROBE-RESULTS.md:21-23). A DELETE of a rule held by an
// ADDED policy is 400 AK000002 "is being used by NAT policies" (:425), one
// still referred externally 409 (NetworkConfigEntityListener.java:96). A
// duplicate name is 400 field name constraint.invalid (:481), the same code
// as the direction refusal and other rule checks, so it has no signature; the
// create guard reads the collection instead. A rule held only by marked
// policies is marked DELETED, which a plain GET hides (:327-330).
var natRuleDescriptor = Descriptor{
	Collection: "nat-rules",
	InUse:      []Signature{{Status: 400, Code: "AK000002", Message: "is being used by NAT policies"}, {Status: 409, Code: "AK000001"}},

	Provisionable:      true,
	MarkedRowLifecycle: true,
	DocState:           DocStateMissingIsAdded,
	Uniqueness:         Uniqueness{Key: "name", Scope: ScopeCollection},
	Service:            ServiceTPS,
	Flags: []FlagRule{
		{Flag: "RELEASE-NAT-POLICY-OUTBOUND", Effect: FlagIncludeField, Field: "direction", UISource: "src/views/network/provision/policies/natPolicies/CreateNATPolicyRule.js:59-61|:530"},
	},
}

// NatRuleDescriptor returns the NAT rule write descriptor.
func NatRuleDescriptor() Descriptor { return natRuleDescriptor.clone() }

// NewNatRuleCore returns the write core for NAT rules.
func NewNatRuleCore(ac *AlkiraClient) (*Core[NatRuleConfig], error) {
	return NewCore[NatRuleConfig](ac, natRuleDescriptor)
}

// CreateNatRule creates a NAT rule.
func (ac *AlkiraClient) CreateNatRule(ctx context.Context, r *NatRuleConfig) (*NatRuleConfig, *WriteResult, error) {
	c, err := NewNatRuleCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Create(ctx, r, WriteOptions{})
}

// GetNatRule reads one NAT rule and its raw body.
func (ac *AlkiraClient) GetNatRule(ctx context.Context, id string) (*NatRuleConfig, []byte, error) {
	c, err := NewNatRuleCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Get(ctx, id)
}

// UpdateNatRule applies a merge patch to a NAT rule.
func (ac *AlkiraClient) UpdateNatRule(ctx context.Context, id string, patch json.RawMessage) (*NatRuleConfig, *WriteResult, error) {
	c, err := NewNatRuleCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Update(ctx, id, patch, ListEdits{})
}

// DeleteNatRule deletes a NAT rule.
func (ac *AlkiraClient) DeleteNatRule(ctx context.Context, id string) (*WriteResult, error) {
	c, err := NewNatRuleCore(ac)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, id)
}

// ListNatRulePaged reads every NAT rule through paged reads.
func (ac *AlkiraClient) ListNatRulePaged(ctx context.Context) ([]NatRuleConfig, error) {
	c, err := NewNatRuleCore(ac)
	if err != nil {
		return nil, err
	}
	return c.ListPaged(ctx)
}
