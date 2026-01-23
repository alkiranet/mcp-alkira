// Copyright (C) 2020-2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type TenantNetworkFirewallZone struct {
	ID                  int    `json:"id"`
	Name                string `json:"name"`
	TagId               int    `json:"tagId"`
	SegmentID           int    `json:"segmentId"`
	ServiceID           int    `json:"serviceId"`
	CXP                 string `json:"cxp"`
	ZoneType            string `json:"zoneType"`
	NetworkEntityID     string `json:"networkEntityId"`
	NetworkEntityType   string `json:"networkEntityType"`
	FtntPolicyID        int    `json:"ftntPolicyId"`
	LastConfigUpdatedAt int    `json:"lastConfigUpdatedAt"`
}

type TenantNetworkFirewallZones struct {
	FirewallZones []TenantNetworkFirewallZone `json:"firewallZones"`
}

// A simplifed struct that describes a tenant network.
type TenantNetwork struct {
	ID            int                         `json:"id"`
	Name          string                      `json:"name"`
	State         string                      `json:"state"`
	AwsExternalId string                      `json:"awsExternalId,omitempty"`
	CXPPairs      []interface{}               `json:"cxpPairs,omitempty"`
	FirewallZones []TenantNetworkFirewallZone `json:"firewallZones,omitempty"`
	ByoIPs        []interface{}               `json:"byoips,omitempty"`
}

// NewTenantNetwork
func NewTenantNetwork(ac *AlkiraClient) *AlkiraApi[TenantNetwork] {
	uri := fmt.Sprintf("%s/api/tenantnetworks", ac.URI)
	api := &AlkiraApi[TenantNetwork]{ac, uri, PaginationOff}
	return api
}

// NewTenantnetworkFirewallZones
func NewTenantNetworkFirewallZones(ac *AlkiraClient) *AlkiraApi[TenantNetworkFirewallZones] {
	uri := fmt.Sprintf("%s/api/tenantnetworks", ac.URI)
	api := &AlkiraApi[TenantNetworkFirewallZones]{ac, uri, PaginationOff}
	return api
}

// GetTenantNetworkSummary get the tenant networks of the current tenant
func (ac *AlkiraClient) GetTenantNetworkSummary() (string, error) {
	uri := fmt.Sprintf("%s/tenantnetworksummaries", ac.URI)

	data, err := ac.Get(uri)

	if err != nil {
		return "", err
	}

	return string(data), nil
}

type TenantNetworkId struct {
	Id int `json:"id"`
}

// GetTenantNetworkId get the tenant network Id of the current tenant
func (ac *AlkiraClient) GetTenantNetworkId() (string, error) {

	uri := fmt.Sprintf("%s/tenantnetworks", ac.URI)

	data, err := ac.Get(uri)

	if err != nil {
		return "", err
	}

	var result []TenantNetworkId
	err = json.Unmarshal([]byte(data), &result)

	if err != nil {
		return "", fmt.Errorf("GetTenantNetworkId: failed to unmarshal: %v", err)
	}

	return strconv.Itoa(result[0].Id), nil
}

type TenantNetworkState struct {
	State string `json:"state"`
}

// GetTenantNetworkState get the tenant network state
func (ac *AlkiraClient) GetTenantNetworkState() (string, error) {
	uri := fmt.Sprintf("%s/tenantnetworks/%s", ac.URI, ac.TenantNetworkId)

	data, err := ac.Get(uri)

	if err != nil {
		return "", err
	}

	var result TenantNetworkState
	err = json.Unmarshal([]byte(data), &result)

	if err != nil {
		return "", fmt.Errorf("GetTenantNetworkState: failed to unmarshal: %v", err)
	}

	return result.State, nil
}

type TenantNetworkConnectorState struct {
	State    string `json:"state"`
	DocState string `json:"docState"`
}

// GetTenantNetworkConnectorState get the tenant network connector state by its Id
func (ac *AlkiraClient) GetTenantNetworkConnectorState(id string) (string, error) {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/connectors/%s", ac.URI, ac.TenantNetworkId, id)

	data, err := ac.Get(uri)

	if err != nil {
		return "", err
	}

	var result TenantNetworkConnectorState
	err = json.Unmarshal([]byte(data), &result)

	if err != nil {
		return "", fmt.Errorf("GetTenantNetworkConnectorState: failed to unmarshal: %v", err)
	}

	return result.State, nil
}

type TenantNetworkServiceState struct {
	State    string `json:"state"`
	DocState string `json:"docState"`
}

// GetTenantNetworkServiceState get the tenant network service state by its Id
func (ac *AlkiraClient) GetTenantNetworkServiceState(id string) (string, error) {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/services/%s", ac.URI, ac.TenantNetworkId, id)

	data, err := ac.Get(uri)

	if err != nil {
		return "", err
	}

	var result TenantNetworkServiceState
	err = json.Unmarshal([]byte(data), &result)

	if err != nil {
		return "", fmt.Errorf("GetTenantNetworkConnectorState: failed to unmarshal: %v", err)
	}

	return result.State, nil
}

type TenantNetworkProvisionRequest struct {
	Id    string `json:"id"`
	State string `json:"state"`
}

// GetTenantNetworkProvisionRequest get the tenant network provision request
func (ac *AlkiraClient) GetTenantNetworkProvisionRequest(id string) (*TenantNetworkProvisionRequest, error) {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/provision-requests/%s", ac.URI, ac.TenantNetworkId, id)

	data, err := ac.Get(uri)

	if err != nil {
		return nil, err
	}

	var result TenantNetworkProvisionRequest
	err = json.Unmarshal([]byte(data), &result)

	if err != nil {
		return nil, fmt.Errorf("GetTenantNetworkProvisionRequest: failed to unmarshal: %v", err)
	}

	return &result, nil
}

// ProvisionTenantNetwork provisioning the current tenant network by its Id
func (ac *AlkiraClient) ProvisionTenantNetwork() (string, error) {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/provision", ac.URI, ac.TenantNetworkId)

	data, err := ac.Create(uri, nil)

	if err != nil {
		return "", err
	}

	var result TenantNetworkState
	err = json.Unmarshal([]byte(data), &result)

	if err != nil {
		return "", fmt.Errorf("ProvisionTenantNetwork: failed to unmarshal: %v", err)
	}

	return result.State, nil
}
