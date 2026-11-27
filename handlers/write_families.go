package handlers

import "strings"

// WithWriteFamilies limits the write tools to the named families, given by
// their tool stems such as "segment" or "connector_aws_vpc". Without it
// every family is served. Read tools are not affected.
func WithWriteFamilies(names ...string) WriteOption {
	return func(p *WritePipeline) error {
		p.families = append([]string(nil), names...)
		return nil
	}
}

// Serves reports whether the pipeline serves the write tools of family name.
func (p *WritePipeline) Serves(name string) bool {
	if p.families == nil {
		return true
	}
	for _, f := range p.families {
		if f == name {
			return true
		}
	}
	return false
}

// WriteFamilies returns the WithWriteFamilies names, nil when every family is served.
func (p *WritePipeline) WriteFamilies() []string { return p.families }

// ParseWriteFamilies splits a comma-separated family list, as given to
// -writeFamilies or AK_MCP_WRITE_FAMILIES; blank entries are dropped and an
// empty list means every family.
func ParseWriteFamilies(s string) []string {
	var out []string
	for _, f := range strings.Split(s, ",") {
		if f = strings.TrimSpace(f); f != "" {
			out = append(out, f)
		}
	}
	return out
}
