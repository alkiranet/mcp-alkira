package writes

import (
	ak "github.com/alkiranet/client-go/tenant"
	"github.com/alkiranet/mcp-alkira/handlers"
)

// SegmentResource is the segment resource write family (TPS /segment-resources).
var SegmentResource = handlers.WriteFamily[ak.SegmentResourceConfig]{
	Name:       "segment_resource",
	Noun:       "segment resource",
	Descriptor: ak.SegmentResourceDescriptor(),
	// TPS creates the "<name>-Group" group with the resource and deletes it
	// with it (SegmentResourceServiceImpl.java:405-426, :1391).
	Implicit: func(name string) []string { return []string{name + "-Group"} },
	CreateNote: "segment is a segment name. groupPrefixes maps groups to prefix lists; -1 means any group or any prefix list, and a -1/-1 entry must be the only one. " +
		"The backend creates a group named <name>-Group with the resource, so a name whose -Group group exists is refused too.",
	UpdateNote: "A system-generated resource (<segment>-ALL) cannot be changed. The segment cannot change while a segment resource share uses the resource.",
	DeleteNote: "The delete also removes the resource's -Group group. A resource used by a segment resource share or a route policy is refused as in use; a system-generated resource cannot be deleted.",
}
