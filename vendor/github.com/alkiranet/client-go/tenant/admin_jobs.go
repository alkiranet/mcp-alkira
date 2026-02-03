// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type Jobs struct {
	Id        string    `json:"id"`
	Summary   string    `json:"summary,omitempty"`
	StartTime int       `json:"startTime"`
	EndTime   int       `json:"endTime"`
	Errors    []string  `json:"errors,omitempty"`
	Status    string    `json:"status,omitempty"`
	Tasks     []JobTask `json:"tasks,omitempty"`
}

type JobTask struct {
	Id        string `json:"id"`
	Summary   string `json:"summary,omitempty"`
	Type      string `json:"type,omitempty"`
	Progress  int    `json:"progress"`
	StartTime int    `json:"startTime"`
	EndTime   int    `json:"endTime"`
	Status    string `json:"status,omitempty"`
}

// GetJobs get jobs with optional filter with status or type
func (ac *AlkiraClient) GetJobs(jobStatus string, jobType string, offset string, limit string) (string, error) {

	baseUri := fmt.Sprintf("%s/api/jobs?paginated=true", ac.URI)

	uri, err := url.Parse(baseUri)

	if err != nil {
		return "", fmt.Errorf("GetJobs: failed to parse URI: %v", err)
	}

	// Process optional job parameters
	q := uri.Query()

	if jobStatus != "" {
		q.Add("status", jobStatus)
	}
	if jobType != "" {
		q.Add("type", jobType)
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

func (ac *AlkiraClient) GetJobsSummary(jobStatus string, jobType string, offset string, limit string) (string, error) {

	baseUri := fmt.Sprintf("%s/api/jobs?paginated=true", ac.URI)

	uri, err := url.Parse(baseUri)

	if err != nil {
		return "", fmt.Errorf("GetAlerts: failed to parse URI: %v", err)
	}

	// Process optional query parameters
	q := uri.Query()
	if jobStatus != "" {
		q.Add("status", jobStatus)
	}
	if jobType != "" {
		q.Add("type", jobType)
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

	var result DataWithPagination[Jobs]

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
