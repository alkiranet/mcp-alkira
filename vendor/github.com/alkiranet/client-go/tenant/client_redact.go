// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"fmt"
	"log/slog"
	"strings"
)

const redacted = "[REDACTED]"

// redactAuthorization keeps the scheme of an Authorization value and hides the credential.
func redactAuthorization(v string) string {
	if v == "" {
		return ""
	}
	if scheme, _, ok := strings.Cut(v, " "); ok {
		return scheme + " " + redacted
	}
	return redacted
}

func redactSecret(v string) string {
	if v == "" {
		return ""
	}
	return redacted
}

// String formats the client with Authorization and ApiKey redacted. The value
// receiver makes fmt redact both AlkiraClient and *AlkiraClient.
func (ac AlkiraClient) String() string {
	return fmt.Sprintf("AlkiraClient{URI:%s TenantNetworkId:%s HostHeader:%s Authorization:%s ApiKey:%s MaxToken:%d Compression:%s}",
		ac.URI, ac.TenantNetworkId, ac.HostHeader, redactAuthorization(ac.Authorization), redactSecret(ac.ApiKey),
		ac.MaxToken, ac.Compression)
}

// GoString implements fmt.GoStringer so %#v is redacted as well.
func (ac AlkiraClient) GoString() string {
	return fmt.Sprintf("tenant.AlkiraClient{URI:%q, TenantNetworkId:%q, HostHeader:%q, Authorization:%q, ApiKey:%q, MaxToken:%d, Compression:%q}",
		ac.URI, ac.TenantNetworkId, ac.HostHeader, redactAuthorization(ac.Authorization), redactSecret(ac.ApiKey),
		ac.MaxToken, ac.Compression)
}

// LogValue implements slog.LogValuer with Authorization and ApiKey redacted.
func (ac AlkiraClient) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("uri", ac.URI),
		slog.String("tenantNetworkId", ac.TenantNetworkId),
		slog.String("hostHeader", ac.HostHeader),
		slog.String("authorization", redactAuthorization(ac.Authorization)),
		slog.String("apiKey", redactSecret(ac.ApiKey)),
		slog.Int("maxToken", ac.MaxToken),
		slog.String("compression", ac.Compression),
	)
}
