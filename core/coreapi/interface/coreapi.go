// Package iface defines DMS3FS Core API which is a set of interfaces used to
// interact with DMS3FS nodes.
package iface

import (
	"context"

	dms3ld "github.com/dms3-fs/go-ld-format"
)

// CoreAPI defines an unified interface to DMS3FS for Go programs
type CoreAPI interface {
	// Unixfs returns an implementation of Unixfs API
	Unixfs() UnixfsAPI

	// Block returns an implementation of Block API
	Block() BlockAPI

	// Dag returns an implementation of Dag API
	Dag() DagAPI

	// Name returns an implementation of Name API
	Name() NameAPI

	// Key returns an implementation of Key API
	Key() KeyAPI

	// Pin returns an implementation of Pin API
	Pin() PinAPI

	// ObjectAPI returns an implementation of Object API
	Object() ObjectAPI

	// ResolvePath resolves the path using Unixfs resolver
	ResolvePath(context.Context, Path) (ResolvedPath, error)

	// ResolveNode resolves the path (if not resolved already) using Unixfs
	// resolver, gets and returns the resolved Node
	ResolveNode(context.Context, Path) (dms3ld.Node, error)
}
