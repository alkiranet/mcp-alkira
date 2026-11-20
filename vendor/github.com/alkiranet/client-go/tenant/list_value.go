// Copyright (C) 2026 Alkira Inc. All Rights Reserved.

package tenant

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ValueListConfig is the write model shared by the AS-path, community and
// extended community lists (TPS model/AsPathList.java, CommunityList.java,
// ExtendedCommunityList.java); the three differ only in their descriptor.
// The backend checks each value's syntax per family; values carry no
// element rule here. Only AS-path list bodies carry docState.
type ValueListConfig struct {
	ID          json.Number `json:"id,omitempty" ak:"readonly"`
	Name        string      `json:"name" ak:"required,maxlen=256,trimmed,charset=name,src=AsPathList.java:35-39|CommunityList.java:35-39|ExtendedCommunityList.java:35-39"`
	Description *string     `json:"description,omitempty" ak:"maxlen=512,trimmed,charset=name,src=AsPathList.java:41-44|CommunityList.java:41-44|ExtendedCommunityList.java:41-44"`
	Values      []string    `json:"values,omitempty" ak:"required,trimmed,src=AsPathList.java:46-51|CommunityList.java:46-51|ExtendedCommunityList.java:46-52"`

	DocState            string      `json:"docState,omitempty" ak:"readonly"`
	LastConfigUpdatedAt json.Number `json:"lastConfigUpdatedAt,omitempty" ak:"readonly"`
}

// extCommunityKind matches the rt and soo values of an extended community
// list; every other value is a regex value.
var extCommunityKind = regexp.MustCompile(`rt|soo`)

// checkExtCommunityMix refuses an extended community list that mixes rt or
// soo values with regex values, as TPS does after the request
// (ExtendedCommunityListServiceImpl.java:136-144).
func checkExtCommunityMix(body map[string]any) error {
	values, _ := body["values"].([]any)
	var first bool
	for i, v := range values {
		s, _ := v.(string)
		typed := extCommunityKind.MatchString(strings.ToLower(s))
		if i == 0 {
			first = typed
			continue
		}
		if typed != first {
			return fmt.Errorf("%w: values mixes rt or soo values with regex values; an extended community list holds one kind", ErrInvalidInput)
		}
	}
	return nil
}
