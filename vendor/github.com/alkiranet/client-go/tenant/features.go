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
const WriteDisabledFlag = "MCP-WRITE-DISABLED"

// ErrWritesDisabled is returned by every write while WriteDisabledFlag is on.
var ErrWritesDisabled = errors.New("changes through MCP are disabled for this tenant")

// featureCacheTTL bounds how long a flag change takes to reach writes.
const featureCacheTTL = 60 * time.Second

// featureReadTimeout bounds the flag read so a failing endpoint cannot use up
// the write deadline on read retries.
const featureReadTimeout = 5 * time.Second

type featureEntry struct {
	flags []string
	at    time.Time
}

var (
	featureCache sync.Map // portal URI|host|tenant network id -> featureEntry
	featureNow   = time.Now
)

// EnabledFeatures returns the tenant's enabled feature flags from
// GET /api/tenantenabledfeatures, the list the portal reads. Results are
// cached per tenant network for 60 s; failures are not cached.
func (ac *AlkiraClient) EnabledFeatures(ctx context.Context) ([]string, error) {
	key := ac.URI + "|" + ac.HostHeader + "|" + ac.TenantNetworkId
	if e, ok := featureCache.Load(key); ok && featureNow().Sub(e.(featureEntry).at) < featureCacheTTL {
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
		featureCache.Store(key, featureEntry{slices.Clone(flags), featureNow()})
	}
	return flags, nil
}

// CheckWritesAllowed returns ErrWritesDisabled when WriteDisabledFlag is on.
// A failed flag read does not block the write: it returns
// flagsUnavailable true and a nil error.
func (ac *AlkiraClient) CheckWritesAllowed(ctx context.Context) (flagsUnavailable bool, err error) {
	flags, err := ac.EnabledFeatures(ctx)
	if err != nil {
		logf("DEBUG", "enabled features unavailable, write proceeds: %v", err)
		return true, nil
	}
	if contains(flags, WriteDisabledFlag) {
		return false, ErrWritesDisabled
	}
	return false, nil
}
