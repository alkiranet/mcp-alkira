// Copyright (C) 2025 Alkira Inc. All Rights Reserved.

package alkira

import (
	"fmt"
)

// GetHealthAll get all resources health status
func (ac *AlkiraClient) GetHealthAll() (string, error) {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/health", ac.URI, ac.TenantNetworkId)
	data, _, err := ac.get(uri)

	return string(data), err
}

// GetHealthOfConnectorInstance get the health status by given
// connector instance ID
func (ac *AlkiraClient) GetHealthConnectorInstance(connectorId int, instanceId int) (string, error) {

	if connectorId == 0 || instanceId == 0 {
		return "", fmt.Errorf("Invalid connector ID %d or instance ID %d.", connectorId, instanceId)
	}

	uri := fmt.Sprintf("%s/tenantnetworks/%s/health/connector/%d/instance/%d", ac.URI, ac.TenantNetworkId, connectorId, instanceId)
	data, _, err := ac.get(uri)

	return string(data), err
}

// GetHealthOfService get the health status by given service ID
func (ac *AlkiraClient) GetHealthOfService(serviceId int) (string, error) {

	if serviceId == 0 {
		return "", fmt.Errorf("Invalid service ID %d.", serviceId)
	}

	uri := fmt.Sprintf("%s/tenantnetworks/%s/health/service/%d", ac.URI, ac.TenantNetworkId, serviceId)
	data, _, err := ac.get(uri)

	return string(data), err
}

// GetHealthOfServiceInstance get the health status by given service
// instance ID
func (ac *AlkiraClient) GetHealthOfServiceInstance(serviceId int, instanceId int) (string, error) {

	if serviceId == 0 || instanceId == 0 {
		return "", fmt.Errorf("Invalid service ID %d or instance ID %d.", serviceId, instanceId)
	}

	uri := fmt.Sprintf("%s/tenantnetworks/%s/health/service/%d/instance/%d", ac.URI, ac.TenantNetworkId, serviceId, instanceId)
	data, _, err := ac.get(uri)

	return string(data), err
}
