package tenant

import (
	"encoding/json"
	"fmt"
)

type ConnectorGcpInterconnect struct {
	Id               json.Number                        `json:"id"`
	Name             string                             `json:"name"`
	Description      string                             `json:"description,omitempty"`
	ScaleGroupId     string                             `json:"scaleGroupId,omitempty"`
	Cxp              string                             `json:"cxp"`
	Group            string                             `json:"group,omitempty"`
	ImplicitGroupId  int                                `json:"implicitGroupId,omitempty"` // RESPONSE ONLY
	Size             string                             `json:"size"`
	TunnelProtocol   string                             `json:"tunnelProtocol"`
}

func NewConnectorGcpInterconnect(ac *AlkiraClient) *AlkiraApi[ConnectorGcpInterconnect] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/gcp-interconnect-connectors", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[ConnectorGcpInterconnect]{ac, uri, PaginationOn}
	return api
}
