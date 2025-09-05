// Copyright (C) 2021-2025 Alkira Inc. All Rights Reserved.

package tenant

import (
	"fmt"
)

type CloudProviderAccount struct {
	Name          string `json:"name"`
	Id            string `json:"id,omitempty"`
	CredentialId  string `json:"credentialId"`
	CloudProvider string `json:"cloudProvider"`
	AutoSync      string `json:"autoSync"`
	NativeId      string `json:"nativeId"`
}

// NewCloudProviderAccounts
func NewCloudProviderAccounts(ac *AlkiraClient) *AlkiraApi[CloudProviderAccount] {
	uri := fmt.Sprintf("%s/cloud-provider-accounts", ac.URI)
	api := &AlkiraApi[CloudProviderAccount]{ac, uri, PaginationOn}
	return api
}
