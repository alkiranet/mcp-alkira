// Copyright (C) 2020-2025 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
)

//
// The concept of group is not clearly defined in the system. It seems
// like the group API also lists all other "types" of groups as well.
//

// A summary of a group (other fields are ignored from the original
// response)
type Group struct {
	Id           json.Number `json:"id"`
	Name         string      `json:"name"`
	Description  string      `json:"description,omitempty"`
	InternalName string      `json:"internalName"`
	Type         string      `json:"type"`
}

func NewGroup(ac *AlkiraClient) *AlkiraApi[Group] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/groups", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[Group]{ac, uri, PaginationOn}
	return api
}

//
// User Group
//

type UserGroup struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewUserGroup(ac *AlkiraClient) *AlkiraApi[UserGroup] {
	uri := fmt.Sprintf("%s/user-groups", ac.URI)
	api := &AlkiraApi[UserGroup]{ac, uri, PaginationOff}
	return api
}

//
// Inter Connector Communication Group
//

type InterConnectorCommunicationGroup struct {
	Id                           json.Number `json:"id"`
	Name                         string      `json:"name"`
	Description                  string      `json:"description"`
	Segment                      string      `json:"segment"`
	Cxp                          string      `json:"cxp"`
	ConnectorProviderRegion      string      `json:"connectorProviderRegion"`
	ConnectorType                string      `json:"connectorType"`
	VirtualNetworkManagerAzureId int         `json:"azureVirtualNetworkManagerId"`
}

func NewInterConnectorCommunicationGroup(ac *AlkiraClient) *AlkiraApi[InterConnectorCommunicationGroup] {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/inter-connector-communication-groups", ac.URI, ac.TenantNetworkId)
	api := &AlkiraApi[InterConnectorCommunicationGroup]{ac, uri, PaginationOn}
	return api
}
