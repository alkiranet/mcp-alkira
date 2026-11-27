// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
)

// FirewallSegmentOptions is one segmentOptions entry of a firewall service,
// keyed by segment name (TPS model/service/segmentoptions/Firewall.java:22-25).
// Zones are matched by segment and zone name; the backend fills segmentId.
type FirewallSegmentOptions struct {
	SegmentID     *int64              `json:"segmentId,omitempty" ak:"readonly"`
	ZonesToGroups map[string][]string `json:"zonesToGroups"`
}

// firewallDescriptor holds the rules the four firewall service families
// share (TPS ServiceLogicImpl, AbstractServiceController):
//   - GET, PUT and DELETE of a missing id answer 404 AK000003
//     (ServiceLogicImpl.java:1102, :1358, :865).
//   - GET /{family}/{id} does not filter by type and can answer 500 for
//     another type's id (CheckPointFWServiceLogicImpl.java:552-553), so the
//     type comes from GET /services/{id} first (ServiceType).
//   - A service name is unique across every service type in the network,
//     marked rows included: 409 AK000001 "Service with name : <n> already
//     exists in network : <id>" (ServiceLogicImpl.java:2174-2184).
//   - A DELETE is refused with 400 AK000002 while a policy rule uses the
//     service, carrying policy-service's "is in use by" text
//     (PolicyServiceProviderImpl.java:290-300; policy-service
//     PolicyValidatorImpl.java:168, :194), and with 409 for an external
//     reference (NetworkConfigEntityListener.java:83-97).
//   - An update runs the D13 state check (ProvisionedEntityValidatorImpl.java:34-76);
//     a deployed service is marked for deletion on delete and a plain GET
//     then answers 404.
//   - instances match by id on update: an element without id is a new
//     instance and a stored instance the body leaves out is removed
//     (ServiceLogicImpl.java:913-990).
func firewallDescriptor(collection, serviceType string) Descriptor {
	return Descriptor{
		Collection:  collection,
		ServiceType: serviceType,
		InUse: []Signature{
			{Status: 400, Code: "AK000002", Message: "is in use by"},
			{Status: 409, Code: "AK000001", Message: "is still referred externally"},
		},
		Duplicate:          []Signature{{Status: 409, Code: "AK000001", Message: "already exists in network"}},
		Provisionable:      true,
		MarkedRowLifecycle: true,
		DocState:           DocStatePlainGet,
		Uniqueness:         Uniqueness{Key: "name", Scope: ScopeServices},
		IDMatchedArrays:    []string{"instances"},
		StateGuard:         StateGuardD13,
		Service:            ServiceTPS,
		PrepareStored:      dropManagementZone,
	}
}

// managementZone is the zone TPS adds to the management segment
// (Defaults.java:111). With RELEASE_FIREWALL_ZONE_RESERVED_NAME_VALIDATION
// on, a body that names it is refused (FWServiceSegmentOptionValidator.java:70-91),
// so an update leaves it out, as the portal edit does
// (fortinetConfig/add/index.js:505; panConfig/EditFirewall/index.js:284).
const managementZone = "ALKIRA_MGMT_ZONE"

// dropManagementZone removes managementZone from the stored segmentOptions,
// and a segment entry left with no zone.
func dropManagementZone(m map[string]any) {
	opts, _ := m["segmentOptions"].(map[string]any)
	for seg, v := range opts {
		o, _ := v.(map[string]any)
		zones, _ := o["zonesToGroups"].(map[string]any)
		if _, ok := zones[managementZone]; !ok {
			continue
		}
		delete(zones, managementZone)
		if len(zones) == 0 {
			delete(opts, seg)
		}
	}
}

// checkInstanceCounts refuses a body whose instance list does not hold
// maxInstanceCount instances or whose minInstanceCount is above it, and,
// with sameMinMax, a minInstanceCount other than maxInstanceCount (autoscale
// off). Every firewall service validator checks these.
func checkInstanceCounts(m map[string]any, sameMinMax bool) []error {
	var errs []error
	maxN, hasMax := number(m["maxInstanceCount"])
	minN, hasMin := number(m["minInstanceCount"])
	insts, _ := m["instances"].([]any)
	if hasMax && float64(len(insts)) != maxN {
		errs = append(errs, fmt.Errorf("instances must hold maxInstanceCount (%v) instances, not %d", maxN, len(insts)))
	}
	if hasMax && hasMin && minN > maxN {
		errs = append(errs, errors.New("minInstanceCount must not be above maxInstanceCount"))
	}
	if sameMinMax && hasMax && hasMin && minN != maxN {
		errs = append(errs, errors.New("minInstanceCount must equal maxInstanceCount while autoScale is OFF"))
	}
	return errs
}

// number reads a JSON number decoded with UseNumber.
func number(v any) (float64, bool) {
	n, ok := v.(json.Number)
	if !ok {
		return 0, false
	}
	f, err := n.Float64()
	return f, err == nil
}

// checkManagementSegment refuses a managementServer.segment that is not one of segments.
func checkManagementSegment(m map[string]any) []error {
	ms, _ := m["managementServer"].(map[string]any)
	seg, ok := ms["segment"].(string)
	if !ok {
		return nil
	}
	segs, _ := m["segments"].([]any)
	if !slices.Contains(segs, any(seg)) {
		return []error{fmt.Errorf("managementServer.segment %q must be one of segments", seg)}
	}
	return nil
}

func serviceCreate[T any](ctx context.Context, ac *AlkiraClient, d Descriptor, v *T) (*T, *WriteResult, error) {
	core, err := NewCore[T](ac, d)
	if err != nil {
		return nil, nil, err
	}
	return core.Create(ctx, v, WriteOptions{})
}

func serviceGet[T any](ctx context.Context, ac *AlkiraClient, d Descriptor, id string) (*T, []byte, error) {
	core, err := NewCore[T](ac, d)
	if err != nil {
		return nil, nil, err
	}
	return core.Get(ctx, id)
}

func serviceUpdate[T any](ctx context.Context, ac *AlkiraClient, d Descriptor, id string, patch json.RawMessage) (*T, *WriteResult, error) {
	core, err := NewCore[T](ac, d)
	if err != nil {
		return nil, nil, err
	}
	return core.Update(ctx, id, patch, ListEdits{})
}

func serviceDelete[T any](ctx context.Context, ac *AlkiraClient, d Descriptor, id string) (*WriteResult, error) {
	core, err := NewCore[T](ac, d)
	if err != nil {
		return nil, err
	}
	return core.Delete(ctx, id)
}

func serviceList[T any](ctx context.Context, ac *AlkiraClient, d Descriptor) ([]T, error) {
	core, err := NewCore[T](ac, d)
	if err != nil {
		return nil, err
	}
	return core.ListPaged(ctx)
}
