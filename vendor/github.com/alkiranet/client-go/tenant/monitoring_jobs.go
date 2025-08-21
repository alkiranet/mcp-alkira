// Copyright (C) 2025 Alkira Inc. All Rights Reserved.

package tenant

import (
	"fmt"
	"net/url"
)

// GetJobs get jobs with optional filter with status or type
func (ac *AlkiraClient) GetJobs(jobStatus string, jobType string) (string, error) {

	baseUri := fmt.Sprintf("%s/api/jobs", ac.URI)

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

	// GET
	uri.RawQuery = q.Encode()
	data, _, err := ac.get(uri.String())

	return string(data), err
}
