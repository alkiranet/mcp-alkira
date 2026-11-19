// Copyright (C) 2020-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"fmt"
)

type SegmentIpBlocks struct {
	Values []string `json:"values"`
}

type SegmentSrcIpv4PoolList struct {
	StartIp string `json:"startIp,omitempty"`
	EndIp   string `json:"endIp,omitempty"`
}

type ServceTrafficDistributionAlgorthmAttributes struct {
	Keys string `json:"keys,omitempty"`
}

type ServiceTrafficDistribution struct {
	Algorithm           string                                      `json:"algorithm,omitempty"`
	AlgorithmAttributes ServceTrafficDistributionAlgorthmAttributes `json:"algorithmAttributes,omitempty"`
}

type Segment struct {
	Asn                                               int                        `json:"asn"`
	Description                                       string                     `json:"description,omitempty"`
	EnableIpv6ToIpv4Translation                       bool                       `json:"enableIpv6ToIpv4Translation"`
	EnterpriseDNSServerIP                             string                     `json:"enterpriseDNSServerIP,omitempty"`
	Id                                                json.Number                `json:"id,omitempty"` // only for response
	IpBlock                                           string                     `json:"ipBlock,omitempty"`
	IpBlocks                                          SegmentIpBlocks            `json:"ipBlocks,omitempty"`
	Name                                              string                     `json:"name"`
	OverlappingRouteValidationEnabled                 bool                       `json:"overlappingRouteValidationEnabled"`
	ReservePublicIPsForUserAndSiteConnectivity        bool                       `json:"reservePublicIPsForUserAndSiteConnectivity,omitempty"`
	ReservePublicIPsForUserAndSiteConnectivityForCXPs []string                   `json:"reservePublicIPsForUserAndSiteConnectivityForCXPs,omitempty"`
	SrcIpv4PoolList                                   []SegmentSrcIpv4PoolList   `json:"srcIpv4PoolList,omitempty"`
	ServiceTrafficDistribution                        ServiceTrafficDistribution `json:"serviceTrafficDistribution,omitempty"`
}

func NewSegment(ac *AlkiraClient) *AlkiraApi[Segment] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/segments", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[Segment]{ac, uri, PaginationOff}
	return api
}

// SegmentConfig is the write model of a segment (TPS model/Segment.java).
// Optional fields are pointers or omitempty slices, so a zero value is never
// sent; a create sends an unset field that has default= with that value, as
// the portal does.
type SegmentConfig struct {
	Name        string  `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=Segment.java:54-58"`
	Description *string `json:"description,omitempty" ak:"maxlen=512,trimmed,charset=name,src=Segment.java:60-63"`

	ID              json.Number `json:"id,omitempty" ak:"readonly"`
	Services        []int64     `json:"services,omitempty" ak:"readonly"`
	Connectors      []int64     `json:"connectors,omitempty" ak:"readonly"`
	TenantID        json.Number `json:"tenantId,omitempty" ak:"readonly"`
	TenantNetworkID json.Number `json:"tenantNetworkId,omitempty" ak:"readonly"`
	State           string      `json:"state,omitempty" ak:"readonly"`
	SegmentNo       *int64      `json:"segmentNo,omitempty" ak:"readonly"`

	IPBlock               *string                `json:"ipBlock,omitempty"`
	IPBlocks              *SegmentConfigIPBlocks `json:"ipBlocks,omitempty"`
	ASN                   *int64                 `json:"asn,omitempty" ak:"default=65514,range=1..7223|7225..8073|8076..9058|9060..10123|10125..12075|12077..17942|17944..23455|23457..31897|31899..64495|65001..65514|65552..2147483647|4200000000..4200064499|4200065001..4294967294,src=Segment.java:52|Segment.java:96-98|Defaults.java:1445-1458|SegmentASNValidator.java:41-68"`
	EnterpriseDNSServerIP *string                `json:"enterpriseDNSServerIP,omitempty"`

	// Deprecated by the backend: the portal never sends it, TPS keeps it only
	// under AKAMAI-PRO-PHASE-1 or ADVANCED-IPSEC-CONNECTOR, and an update
	// copies it from the request (SegmentServiceImpl.java:303-306, :574-577).
	// It is read-only, resent from the stored body until its expiry and
	// dropped after.
	ReservePublicIPsForUserAndSiteConnectivity        *bool    `json:"reservePublicIPsForUserAndSiteConnectivity,omitempty" ak:"readonly,keep,expired=1861919999,src=Segment.java:103-105"`
	ReservePublicIPsForUserAndSiteConnectivityForCXPs []string `json:"reservePublicIPsForUserAndSiteConnectivityForCXPs,omitempty" ak:"flag=ALLOW-USERS-AND-SITES-CONNECTIVITY-PER-CXP,src=Segment.java:111|createSegments/index.js:79-81|:384-391"`

	InternalName        string      `json:"internalName,omitempty" ak:"readonly"`
	LastConfigUpdatedAt json.Number `json:"lastConfigUpdatedAt,omitempty" ak:"readonly"`

	EnableIpv6ToIpv4Translation *bool                             `json:"enableIpv6ToIpv4Translation,omitempty" ak:"default=false,flag=IFA-IPV6-SUPPORT,src=Segment.java:124-125|SegmentServiceImpl.java:1070-1089|createSegments/index.js:410-422"`
	SrcIpv4PoolList             []SegmentConfigSrcIPv4Pool        `json:"srcIpv4PoolList,omitempty" ak:"flag=IFA-IPV6-SUPPORT,src=Segment.java:127-130|createSegments/index.js:410-422"`
	ServiceTrafficDistribution  *SegmentConfigTrafficDistribution `json:"serviceTrafficDistribution,omitempty" ak:"src=Segment.java:132-133|createSegments/index.js:403-408"`

	OverlappingRouteValidationEnabled *bool `json:"overlappingRouteValidationEnabled,omitempty" ak:"default=true,flag=OVERLAP-ROUTE-INVALIDATION-ON-SEGMENT,src=Segment.java:135-136|createSegments/index.js:424-429"`
}

// SegmentConfigIPBlocks is Segment.ipBlocks (Segment.java:243-253).
type SegmentConfigIPBlocks struct {
	Values                 []string `json:"values,omitempty"`
	PercentageIPBlocksUsed *float64 `json:"percentageIPBlocksUsed,omitempty" ak:"readonly"`
}

// SegmentConfigSrcIPv4Pool is one Segment.srcIpv4PoolList entry.
type SegmentConfigSrcIPv4Pool struct {
	StartIP string `json:"startIp,omitempty"`
	EndIP   string `json:"endIp,omitempty"`
}

// SegmentConfigTrafficDistribution is Segment.serviceTrafficDistribution.
// The backend stores null when it is unset (Segment.java:132-133); the portal
// always sends HASHING with keys DEFAULT on create.
type SegmentConfigTrafficDistribution struct {
	Algorithm           *string                                     `json:"algorithm,omitempty" ak:"default=HASHING,enum=HASHING,src=Segment.java:145-151|createSegments/index.js:403-408"`
	AlgorithmAttributes *SegmentConfigTrafficDistributionAttributes `json:"algorithmAttributes,omitempty"`
}

// SegmentConfigTrafficDistributionAttributes is HashingAlgorithmAttributes.
type SegmentConfigTrafficDistributionAttributes struct {
	Keys *string `json:"keys,omitempty" ak:"default=DEFAULT,enum=DEFAULT|SRC_IP,src=Segment.java:161-169|createSegments/index.js:83-87"`
}

// segmentIPBlocks refuses what TPS validateSegment refuses across the IP
// block fields: an ipBlock missing from a non-empty ipBlocks.values
// (SegmentServiceImpl.java:1045-1050) and a value listed twice
// (@UniqueElements, Segment.java:244). The count limits of ipBlocks.values
// and srcIpv4PoolList are per-tenant soft limits (@SoftLimit, Segment.java:128, :246;
// SoftLimitValidator.java), so the backend checks them.
func segmentIPBlocks(body map[string]any) error {
	blocks, _ := body["ipBlocks"].(map[string]any)
	values, _ := blocks["values"].([]any)
	seen := make(map[string]bool, len(values))
	for _, v := range values {
		s := fmt.Sprint(v)
		if seen[s] {
			return fmt.Errorf("ipBlocks.values lists %s more than once", s)
		}
		seen[s] = true
	}
	if ip, _ := body["ipBlock"].(string); ip != "" && len(values) > 0 && !seen[ip] {
		return fmt.Errorf("ipBlock %s must be one of ipBlocks.values", ip)
	}
	return nil
}

// segmentDescriptor: GET answers a missing id with 400 AK000002 "Segment
// with id: '<id>' was not found" (SegmentServiceImpl.java:469,
// messages.properties:3), PUT with 404 AK000003 (:541) and DELETE with the
// 400 (:384); the message keeps other 400s with that code, a missing tenant
// network included, from reading as not found. A DELETE refused for a connector,
// service or user segment resource is 400 AK000002 (:409, :428, :431), told
// apart from other 400 AK000002 answers by its message; one held by a policy,
// firewall zone or peering gateway is 409 (:390, :394, :403).
var segmentDescriptor = Descriptor{
	Collection:     "segments",
	NotFound:       []Signature{{Status: 400, Code: "AK000002", Message: "Segment with id: '"}, {Status: 404, Code: "AK000003"}},
	DeleteNotFound: []Signature{{Status: 400, Code: "AK000002", Message: "Segment with id: '"}},
	InUse: []Signature{
		{Status: 400, Code: "AK000002", Message: "is associated with user created segment resources"},
		{Status: 400, Code: "AK000002", Message: "not deleted as it is still being used by"},
		{Status: 409, Code: "AK000001"},
	},
	Duplicate: []Signature{{Status: 400, Code: "AK000001", FieldCode: "constraint.exists"}},

	Provisionable: true,
	DocState:      DocStateMissingIsAdded,
	Uniqueness:    Uniqueness{Key: "name", Scope: ScopeCollection},
	StateGuard:    StateGuardFailedOnly,
	Service:       ServiceTPS,
	Flags: []FlagRule{
		{Flag: "OVERLAP-ROUTE-INVALIDATION-ON-SEGMENT", Effect: FlagIncludeField, Field: "overlappingRouteValidationEnabled", UISource: "src/views/segments/createSegments/index.js:92-94|:424-429"},
		{Flag: "ALLOW-USERS-AND-SITES-CONNECTIVITY-PER-CXP", AnyOf: []string{"AKAMAI-PRO-PHASE-1", "ADVANCED-IPSEC-CONNECTOR"}, Effect: FlagIncludeField, Field: "reservePublicIPsForUserAndSiteConnectivityForCXPs", UISource: "src/views/segments/createSegments/index.js:79-81|:88-90|:384-391"},
		{Flag: "IFA-IPV6-SUPPORT", Effect: FlagIncludeField, Field: "srcIpv4PoolList", ClearStored: true, UISource: "src/views/segments/createSegments/index.js:82|:410-422"},
		{Flag: "IFA-IPV6-SUPPORT", Effect: FlagIncludeField, Field: "enableIpv6ToIpv4Translation", ClearStored: true, UISource: "src/views/segments/createSegments/index.js:82|:410-422"},
	},
	CrossCheck: segmentIPBlocks,
}

// SegmentDescriptor returns the segment write descriptor.
func SegmentDescriptor() Descriptor { return segmentDescriptor.clone() }

// NewSegmentCore returns the write core for segments.
func NewSegmentCore(ac *AlkiraClient) (*Core[SegmentConfig], error) {
	return NewCore[SegmentConfig](ac, segmentDescriptor)
}

// CreateSegment creates a segment.
func (ac *AlkiraClient) CreateSegment(ctx context.Context, s *SegmentConfig) (*SegmentConfig, *WriteResult, error) {
	c, err := NewSegmentCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Create(ctx, s, WriteOptions{})
}

// GetSegment reads one segment and its raw body.
func (ac *AlkiraClient) GetSegment(ctx context.Context, id string) (*SegmentConfig, []byte, error) {
	c, err := NewSegmentCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Get(ctx, id)
}

// UpdateSegment applies a merge patch to a segment.
func (ac *AlkiraClient) UpdateSegment(ctx context.Context, id string, patch json.RawMessage) (*SegmentConfig, *WriteResult, error) {
	c, err := NewSegmentCore(ac)
	if err != nil {
		return nil, nil, err
	}
	return c.Update(ctx, id, patch, ListEdits{})
}

// DeleteSegment deletes a segment.
func (ac *AlkiraClient) DeleteSegment(ctx context.Context, id string) (*WriteResult, error) {
	c, err := NewSegmentCore(ac)
	if err != nil {
		return nil, err
	}
	return c.Delete(ctx, id)
}

// ListSegmentPaged reads every segment through paged reads.
func (ac *AlkiraClient) ListSegmentPaged(ctx context.Context) ([]SegmentConfig, error) {
	c, err := NewSegmentCore(ac)
	if err != nil {
		return nil, err
	}
	return c.ListPaged(ctx)
}
