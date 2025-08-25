// Copyright (C) 2022-2025 Alkira Inc. All Rights Reserved.

package tenant

import (
	"fmt"
)

type UserGroup struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// NewUserGroup new user group
func NewUserGroup(ac *AlkiraClient) *AlkiraApi[UserGroup] {
	uri := fmt.Sprintf("%s/user-groups", ac.URI)
	api := &AlkiraApi[UserGroup]{ac, uri}
	return api
}
