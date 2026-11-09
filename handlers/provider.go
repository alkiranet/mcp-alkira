package handlers

import (
	"context"

	ak "github.com/alkiranet/client-go/tenant"
)

// ClientProvider returns the Alkira client a tool call runs with.
type ClientProvider func(ctx context.Context) (*ak.AlkiraClient, error)
