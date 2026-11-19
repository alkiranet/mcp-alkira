// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// ErrHealthInstanceNotFound is returned by GetHealthOfConnectorInstance and
// GetHealthOfServiceInstance when the connector or service has no instance
// with the given id that reports health. It wraps ErrNotFound.
var ErrHealthInstanceNotFound = fmt.Errorf("%w: no instance with that id reports health in the connector or service", ErrNotFound)

// healthInstanceOwner is the part of a connector or service document that
// lists its instances.
type healthInstanceOwner struct {
	Instances []struct {
		Id      int  `json:"id"`
		Dormant bool `json:"dormant"`
	} `json:"instances"`
}

// checkHealthInstance reads the connector or service document at uri and
// returns ErrHealthInstanceNotFound unless it lists instanceId. Dormant
// instances count as missing when skipDormant is set.
//
// The instance health endpoint does not answer for an instance id its owner
// does not list (or, for services, a dormant one): the request stays open
// until the gateway read timeout and is retried, so the check runs first.
func (ac *AlkiraClient) checkHealthInstance(uri string, instanceId string, skipDormant bool) error {
	id, err := strconv.Atoi(instanceId)
	if err != nil {
		return ErrHealthInstanceNotFound
	}

	data, err := ac.Get(uri)
	if err != nil {
		return err
	}

	var owner healthInstanceOwner
	if err := json.Unmarshal(data, &owner); err != nil {
		return fmt.Errorf("health instance lookup: failed to unmarshal: %w", err)
	}

	for _, in := range owner.Instances {
		if in.Id == id && !(skipDormant && in.Dormant) {
			return nil
		}
	}
	return ErrHealthInstanceNotFound
}

// GetHealthAll get all resources health status
func (ac *AlkiraClient) GetHealthAll() (string, error) {
	uri := fmt.Sprintf("%s/tenantnetworks/%s/health", ac.URI, ac.TenantNetworkId)
	data, err := ac.Get(uri)

	return string(data), err
}

// GetHealthOfConnector get the health status by given connector ID
func (ac *AlkiraClient) GetHealthOfConnector(connectorId string) (string, error) {

	if connectorId == "" {
		return "", fmt.Errorf("Invalid connector ID %s.", connectorId)
	}

	uri := fmt.Sprintf("%s/tenantnetworks/%s/health/connector/%s", ac.URI, ac.TenantNetworkId, pathSegment(connectorId))
	data, err := ac.Get(uri)

	return string(data), err
}

// GetHealthOfConnectorInstance get the health status by given
// connector instance ID. It returns an *HTTPError with status 404 for an
// unknown connector and ErrHealthInstanceNotFound for an unknown instance.
func (ac *AlkiraClient) GetHealthOfConnectorInstance(connectorId string, instanceId string) (string, error) {

	if connectorId == "" || instanceId == "" {
		return "", fmt.Errorf("Invalid connector ID %s or instance ID %s.", connectorId, instanceId)
	}

	owner := fmt.Sprintf("%s/tenantnetworks/%s/connectors/%s?includeMarkedForDeletion=true", ac.URI, ac.TenantNetworkId, pathSegment(connectorId))
	if err := ac.checkHealthInstance(owner, instanceId, false); err != nil {
		return "", err
	}

	uri := fmt.Sprintf("%s/tenantnetworks/%s/health/connector/%s/instance/%s", ac.URI, ac.TenantNetworkId, pathSegment(connectorId), pathSegment(instanceId))
	data, err := ac.Get(uri)

	return string(data), err
}

// GetHealthOfService get the health status by given service ID
func (ac *AlkiraClient) GetHealthOfService(serviceId string) (string, error) {

	if serviceId == "" {
		return "", fmt.Errorf("Invalid service ID %s.", serviceId)
	}

	uri := fmt.Sprintf("%s/tenantnetworks/%s/health/service/%s", ac.URI, ac.TenantNetworkId, pathSegment(serviceId))
	data, err := ac.Get(uri)

	return string(data), err
}

// GetHealthOfServiceInstance get the health status by given service
// instance ID. It returns an *HTTPError with status 404 for an unknown
// service and ErrHealthInstanceNotFound for an unknown or dormant instance.
func (ac *AlkiraClient) GetHealthOfServiceInstance(serviceId string, instanceId string) (string, error) {

	if serviceId == "" || instanceId == "" {
		return "", fmt.Errorf("Invalid service ID %s or instance ID %s.", serviceId, instanceId)
	}

	owner := fmt.Sprintf("%s/tenantnetworks/%s/services/%s?includeMarkedForDeletion=true", ac.URI, ac.TenantNetworkId, pathSegment(serviceId))
	if err := ac.checkHealthInstance(owner, instanceId, true); err != nil {
		return "", err
	}

	uri := fmt.Sprintf("%s/tenantnetworks/%s/health/service/%s/instance/%s", ac.URI, ac.TenantNetworkId, pathSegment(serviceId), pathSegment(instanceId))
	data, err := ac.Get(uri)

	return string(data), err
}
