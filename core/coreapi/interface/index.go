package iface

import (
	"context"
//	"io"

	options "github.com/dms3-fs/go-dms3-fs/core/coreapi/interface/options"
//	dms3ld "github.com/dms3-fs/go-ld-format"
)

//type RepoEntry interface {
type RepoList interface {
	// Name returns repository name
	Name() string
	// Id returns repository path
	Id() string
}
//type RepoList []RepoEntry

// UnixfsAPI is the basic interface to immutable files in DMS3FS
type IndexAPI interface {
	// Add imports the data from the reader into merkledag file
	//Add(context.Context, io.Reader) (ResolvedPath, error)

	// Cat returns a reader for the file
	//Cat(context.Context, Path) (Reader, error)
	Index(ctx context.Context, path Path, opts ...options.IndexListOption) (RepoList, error)

	// Ls returns the list of links in a directory
	//Ls(context.Context, Path) ([]*dms3ld.Link, error)
}
