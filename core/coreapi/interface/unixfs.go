package iface

import (
	"context"
	"io"

	dms3ld "github.com/dms3-fs/go-ld-format"
)

// UnixfsAPI is the basic interface to immutable files in DMS3FS
type UnixfsAPI interface {
	// Add imports the data from the reader into merkledag file
	Add(context.Context, io.Reader) (ResolvedPath, error)

	// Cat returns a reader for the file
	Cat(context.Context, Path) (Reader, error)

	// Ls returns the list of links in a directory
	Ls(context.Context, Path) ([]*dms3ld.Link, error)
}
