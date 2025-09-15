// Copyright (C) 2025 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type Alert struct {
	Id        string `json:"id"`
	Summary   string `json:"summary"`
	EventName string `json:"eventName"`
	Priority  string `json:"priority"`
	CreatedAt int    `json:"createdAt"`
	Status    string `json:"status,omitempty"`
}

// GetAlerts get alerts with optional query parameters
func (ac *AlkiraClient) GetAlerts(alertStatus string, alertType string, alertPriority string, offset string, limit string) (string, error) {

	baseUri := fmt.Sprintf("%s/alerts?paginated=true", ac.URI)

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
	if offset != "" {
		q.Add("offset", offset)
	}
	if limit != "" {
		q.Add("limit", limit)
	}

	// GET
	uri.RawQuery = q.Encode()
	data, err := ac.Get(uri.String())

	return string(data), err
}

func (ac *AlkiraClient) GetAlertsSummary(alertStatus string, alertType string, alertPriority string, offset string, limit string) (string, error) {

	baseUri := fmt.Sprintf("%s/alerts?paginated=true", ac.URI)

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
	if offset != "" {
		q.Add("offset", offset)
	}
	if limit != "" {
		q.Add("limit", limit)
	}

	// GET
	uri.RawQuery = q.Encode()
	data, err := ac.Get(uri.String())

	var result DataWithPagination[Alert]

	err = json.Unmarshal([]byte(data), &result)

	if err != nil {
		return "", fmt.Errorf("GetAlertsSummary: failed to unmarshal: %v", err)
	}

	// Marshal the summary data
	summary, err := json.Marshal(result)

	if err != nil {
		return "", fmt.Errorf("GetAlertsSummary: failed to marshal: %v", err)
	}

	return string(summary), err
}
