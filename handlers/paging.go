package handlers

import (
	"fmt"
	"math"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
)

// pagingArgs returns the offset and limit arguments as query values, using defOffset and
// defLimit when an argument is absent. An empty default leaves the parameter off the request.
func pagingArgs(request mcp.CallToolRequest, defOffset, defLimit string) (offset, limit string, err error) {
	if offset, err = pagingArg(request, "offset", defOffset); err != nil {
		return "", "", err
	}
	if limit, err = pagingArg(request, "limit", defLimit); err != nil {
		return "", "", err
	}
	return offset, limit, nil
}

// pagingArg accepts a non-negative integer given as a JSON number or a decimal string.
func pagingArg(request mcp.CallToolRequest, name, def string) (string, error) {
	switch v := request.GetArguments()[name].(type) {
	case nil:
		return def, nil
	case float64:
		if v >= 0 && v <= math.MaxInt32 && v == math.Trunc(v) {
			return strconv.Itoa(int(v)), nil
		}
	case string:
		if v == "" {
			return def, nil
		}
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return strconv.Itoa(n), nil
		}
	}
	return "", fmt.Errorf("%s must be a non-negative integer, got %#v", name, request.GetArguments()[name])
}
