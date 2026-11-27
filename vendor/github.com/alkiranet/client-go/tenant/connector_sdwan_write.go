// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"errors"
	"strings"
)

// sdwanInUse and sdwanDuplicate are the in-use and duplicate answers the
// five SD-WAN connector families share with every TPS connector: a DELETE
// of a connector a segment resource or policy uses is refused with 400
// AK000002, an external reference with 409 (NetworkConfigEntityListener.java:83-97),
// and a connector name is unique across every connector type and group,
// marked rows included (ConnectorServiceImpl.java:5598-5603, :938-940).
var (
	sdwanInUse = []Signature{
		{Status: 400, Code: "AK000002", Message: "associated with"},
		{Status: 409, Code: "AK000001", Message: "is still referred externally"},
	}
	sdwanDuplicate = []Signature{
		{Status: 400, FieldCode: "constraint.notunique"},
		{Status: 400, FieldCode: "connector.00000003"},
	}
)

// checkConnectorName refuses a connector name starting with ALK- in any case
// (GroupServiceImpl.java:1205-1207; ConnectorServiceImpl.java:949-952). The
// SD-WAN families carry their segments in VRF mappings, so the one-segment
// rule of checkConnector does not apply to them.
func checkConnectorName(m map[string]any) error {
	if name, _ := m["name"].(string); strings.HasPrefix(strings.ToUpper(name), "ALK-") {
		return errors.New("name must not start with ALK-")
	}
	return nil
}
