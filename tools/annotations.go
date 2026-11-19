package tools

import "github.com/mark3labs/mcp-go/mcp"

// Write tool annotations by verb (01-api-surface.md section 5). Tools built
// with mcp.NewToolWithRawSchema get no defaults, so every hint is set.
var (
	CreateAnnotations = annotations(false, false, false)
	UpdateAnnotations = annotations(false, true, true)
	DeleteAnnotations = annotations(false, true, true)
)

func annotations(readOnly, destructive, idempotent bool) mcp.ToolAnnotation {
	return mcp.ToolAnnotation{
		ReadOnlyHint:    mcp.ToBoolPtr(readOnly),
		DestructiveHint: mcp.ToBoolPtr(destructive),
		IdempotentHint:  mcp.ToBoolPtr(idempotent),
		OpenWorldHint:   mcp.ToBoolPtr(true),
	}
}
