//go:build writetest

package registry

import "github.com/alkiranet/mcp-alkira/internal/testfamily"

func init() {
	writeFamilies = append(writeFamilies, family(testfamily.Family))
}
