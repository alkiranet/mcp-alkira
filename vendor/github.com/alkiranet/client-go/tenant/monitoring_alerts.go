// Copyright (C) 2025 Alkira Inc. All Rights Reserved.

package tenant

import (
	"fmt"
	"net/url"
)

// GetAlerts get alerts with optional query parameters
func (ac *AlkiraClient) GetAlerts(alertStatus string, alertType string, alertPriority string) (string, error) {

	baseUri := fmt.Sprintf("%s/api/alerts", ac.URI)

	uri, err := url.Parse(baseUri)

	if err != nil {
		return "", fmt.Errorf("GetAlerts: failed to parse URI: %v", err)
	}

	// Process optional query parameters
	q := uri.Query()

	if alertStatus != "" {
		q.Add("status", alertStatus)
	}

	if alertType != "" {
		q.Add("type", alertType)
	}

	if alertPriority != "" {
		q.Add("priority", alertPriority)
	}

	// GET
	uri.RawQuery = q.Encode()
	data, _, err := ac.get(uri.String())

	return string(data), err
}
