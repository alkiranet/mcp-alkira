// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
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

// secretKeys holds, lower-cased, the JSON keys whose values DEBUG logs
// replace, matched without regard to case (TPS spells presharedKey both
// ways): the
// credential and inline secret fields of the write API, plus every field
// tagged ak:"secret" on a type passed to FieldRules.
var secretKeys sync.Map

func init() {
	for _, k := range []string{
		// Credential payloads.
		"ec2SecretKey", "ec2ExternalId", "secretKey", "private_key", "private_key_id", "key",
		"adminPassword", "sicKey", "password", "fmcRegistrationKey", "alkiraAdminPassword",
		"licenseKey", "ionToken", "ionSecret", "authToken", "sharedSecret", "joinToken",
		"activationKey", "configData", "privateKey", "bindPassword", "authKey", "authCode",
		"registrationKey", "masterKey", "registrationPinValue", "activationCode",
		"bgpAuthenticationKey", "accountKey", "apiKey",
		// Secrets carried inline in resource bodies.
		"presharedKeys", "preSharedKey", "localAuthValue", "remoteAuthValue",
		"bgpAuthKeyAlkira", "bgpAuthKey", "localPublicSharedKey", "remotePublicSharedKey",
		"awsBgpAuthKey", "underlayBgpAuthKey", "customerSecret", "gcpBgpAuthKey", "akAdminPassword",
	} {
		secretKeys.Store(strings.ToLower(k), true)
	}
}

// redactBody returns body for logging with the value of every secret key
// replaced, at any depth. A body that is not JSON is replaced whole when it
// mentions a secret key. It returns "" unless AK_LOG is DEBUG, the only level
// that logs bodies, so other levels skip the parse.
func redactBody(body []byte) string {
	if len(body) == 0 || os.Getenv("AK_LOG") != "DEBUG" {
		return ""
	}
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		found := false
		lower := bytes.ToLower(body)
		secretKeys.Range(func(k, _ any) bool {
			found = bytes.Contains(lower, []byte(k.(string)))
			return !found
		})
		if found {
			return redacted
		}
		return string(body)
	}
	out, err := json.Marshal(redactValue(v))
	if err != nil {
		return redacted
	}
	return string(out)
}

func redactValue(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, e := range t {
			if _, ok := secretKeys.Load(strings.ToLower(k)); ok && e != nil {
				t[k] = redacted
				continue
			}
			t[k] = redactValue(e)
		}
	case []any:
		for i, e := range t {
			t[i] = redactValue(e)
		}
	}
	return v
}
