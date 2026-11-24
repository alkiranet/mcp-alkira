// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"
)

// WriteDisabledFlag in a tenant's enabled features makes the core refuse
// every write before any write request.
const WriteDisabledFlag = "BACKEND-MCP-WRITE-DISABLED"

// ErrWritesDisabled is returned by every write while WriteDisabledFlag is on.
var ErrWritesDisabled = errors.New("changes through MCP are disabled for this tenant")

// AsyncValidationFlag turns on async validation for AsyncCapable families.
const AsyncValidationFlag = "RELEASE-ASYNC-NETWORK-ENTITY-VALIDATION-PHASE-1"

// SelectiveProvFlag selects the TPS state-guard branch the D13 guard mirrors.
const SelectiveProvFlag = "SELECTIVE-PROV"

// ErrNotEnabled refuses a write whose type or field needs a flag that is off
// for the tenant.
var ErrNotEnabled = errors.New("not enabled for this tenant")

// ErrFlagsUnavailable refuses a write when the tenant's flags cannot be read
// and no earlier read for the tenant network is cached. The returned error
// also wraps the cause of the failed read.
var ErrFlagsUnavailable = errors.New("cannot check whether changes through MCP are allowed for this tenant; try again")

// featureCacheTTL bounds how long a flag change takes to reach writes.
const featureCacheTTL = 60 * time.Second

// featureReadTimeout bounds the flag read so a failing endpoint cannot use up
// the write deadline on read retries.
var featureReadTimeout = 5 * time.Second

type featureEntry struct {
	flags []string
	at    time.Time
}

var (
	featureCache sync.Map // portal URI|host|tenant network id -> featureEntry
)

// clockNow and clockAfter are the package clock; tests replace them.
var (
	clockNow   = time.Now
	clockAfter = time.After
)

func (ac *AlkiraClient) featureKey() string {
	return ac.URI + "|" + ac.HostHeader + "|" + ac.TenantNetworkId
}

// EnabledFeatures returns the tenant's enabled feature flags from
// GET /api/tenantenabledfeatures, the list the portal reads. Results are
// cached per tenant network for 60 s; failures are not cached.
func (ac *AlkiraClient) EnabledFeatures(ctx context.Context) ([]string, error) {
	key := ac.featureKey()
	if e, ok := featureCache.Load(key); ok && clockNow().Sub(e.(featureEntry).at) < featureCacheTTL {
		return slices.Clone(e.(featureEntry).flags), nil
	}
	ctx, cancel := context.WithTimeout(ctx, featureReadTimeout)
	defer cancel()
	raw, err := ac.doGet(ctx, ac.URI+"/tenantenabledfeatures")
	if err != nil {
		return nil, err
	}
	var flags []string
	if err := json.Unmarshal(raw, &flags); err != nil {
		return nil, fmt.Errorf("enabled features: %w", err)
	}
	// An empty tenant network id cannot tell tenants apart, so nothing is cached.
	if ac.TenantNetworkId != "" {
		featureCache.Store(key, featureEntry{slices.Clone(flags), clockNow()})
	}
	return flags, nil
}

// CheckWritesAllowed returns ErrWritesDisabled when WriteDisabledFlag is on,
// and ErrFlagsUnavailable when the flags cannot be read and none are cached.
func (ac *AlkiraClient) CheckWritesAllowed(ctx context.Context) error {
	_, err := ac.WriteFeatures(ctx)
	return err
}

// WriteFeatures returns the enabled feature flags that writes use: a failed
// read falls back to the last flags read for the tenant network, whatever
// their age. It returns ErrFlagsUnavailable when none are cached and
// ErrWritesDisabled while WriteDisabledFlag is on.
func (ac *AlkiraClient) WriteFeatures(ctx context.Context) ([]string, error) {
	flags, err := ac.EnabledFeatures(ctx)
	if err != nil {
		e, ok := featureCache.Load(ac.featureKey())
		if !ok {
			return nil, fmt.Errorf("%w: %w", ErrFlagsUnavailable, err)
		}
		logf("DEBUG", "enabled features unavailable, write uses the flags read at %s: %v", e.(featureEntry).at.Format(time.RFC3339), err)
		flags = slices.Clone(e.(featureEntry).flags)
	}
	if contains(flags, WriteDisabledFlag) {
		return nil, ErrWritesDisabled
	}
	return flags, nil
}
