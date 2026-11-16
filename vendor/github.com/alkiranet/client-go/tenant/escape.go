// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"net/url"
	"strings"
)

// pathSegment escapes s for use as one URL path segment. "." and ".." are
// percent-encoded so they cannot act as relative path steps.
func pathSegment(s string) string {
	if s == "." || s == ".." {
		return strings.ReplaceAll(s, ".", "%2E")
	}
	return url.PathEscape(s)
}
