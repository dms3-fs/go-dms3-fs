package iface

import (
	dms3fspath "github.com/dms3-fs/go-path"

	cid "github.com/dms3-fs/go-cid"
)

//TODO: merge with dms3fspath so we don't depend on it

// Path is a generic wrapper for paths used in the API. A path can be resolved
// to a CID using one of Resolve functions in the API.
//
// Paths must be prefixed with a valid prefix:
//
// * /dms3fs - Immutable unixfs path (files)
// * /dms3ld - Immutable dms3ld path (data)
// * /dms3ns - Mutable names. Usually resolves to one of the immutable paths
//TODO: /local (MFS)
type Path interface {
	// String returns the path as a string.
	String() string

	// Namespace returns the first component of the path.
	//
	// For example path "/dms3fs/QmHash", calling Namespace() will return "dms3fs"
	Namespace() string

	// Mutable returns false if the data pointed to by this path in guaranteed
	// to not change.
	//
	// Note that resolved mutable path can be immutable.
	Mutable() bool
}

// ResolvedPath is a path which was resolved to the last resolvable node
type ResolvedPath interface {
	// Cid returns the CID of the node referenced by the path. Remainder of the
	// path is guaranteed to be within the node.
	//
	// Examples:
	// If you have 3 linked objects: QmRoot -> A -> B:
	//
	// cidB := {"foo": {"bar": 42 }}
	// cidA := {"B": {"/": cidB }}
	// cidRoot := {"A": {"/": cidA }}
	//
	// And resolve paths:
	// * "/dms3fs/${cidRoot}"
	//   * Calling Cid() will return `cidRoot`
	//   * Calling Root() will return `cidRoot`
	//   * Calling Remainder() will return ``
	//
	// * "/dms3fs/${cidRoot}/A"
	//   * Calling Cid() will return `cidA`
	//   * Calling Root() will return `cidRoot`
	//   * Calling Remainder() will return ``
	//
	// * "/dms3fs/${cidRoot}/A/B/foo"
	//   * Calling Cid() will return `cidB`
	//   * Calling Root() will return `cidRoot`
	//   * Calling Remainder() will return `foo`
	//
	// * "/dms3fs/${cidRoot}/A/B/foo/bar"
	//   * Calling Cid() will return `cidB`
	//   * Calling Root() will return `cidRoot`
	//   * Calling Remainder() will return `foo/bar`
	Cid() *cid.Cid

	// Root returns the CID of the root object of the path
	//
	// Example:
	// If you have 3 linked objects: QmRoot -> A -> B, and resolve path
	// "/dms3fs/QmRoot/A/B", the Root method will return the CID of object QmRoot
	//
	// For more examples see the documentation of Cid() method
	Root() *cid.Cid

	// Remainder returns unresolved part of the path
	//
	// Example:
	// If you have 2 linked objects: QmRoot -> A, where A is a CBOR node
	// containing the following data:
	//
	// {"foo": {"bar": 42 }}
	//
	// When resolving "/dms3ld/QmRoot/A/foo/bar", Remainder will return "foo/bar"
	//
	// For more examples see the documentation of Cid() method
	Remainder() string

	Path
}

// path implements coreiface.Path
type path struct {
	path dms3fspath.Path
}

// resolvedPath implements coreiface.resolvedPath
type resolvedPath struct {
	path
	cid       *cid.Cid
	root      *cid.Cid
	remainder string
}

// Dms3FsPath creates new /dms3fs path from the provided CID
func Dms3FsPath(c *cid.Cid) ResolvedPath {
	return &resolvedPath{
		path:      path{dms3fspath.Path("/dms3fs/" + c.String())},
		cid:       c,
		root:      c,
		remainder: "",
	}
}

// Dms3LdPath creates new /dms3ld path from the provided CID
func Dms3LdPath(c *cid.Cid) ResolvedPath {
	return &resolvedPath{
		path:      path{dms3fspath.Path("/dms3ld/" + c.String())},
		cid:       c,
		root:      c,
		remainder: "",
	}
}

// ParsePath parses string path to a Path
func ParsePath(p string) (Path, error) {
	pp, err := dms3fspath.ParsePath(p)
	if err != nil {
		return nil, err
	}

	return &path{path: pp}, nil
}

// NewResolvedPath creates new ResolvedPath. This function performs no checks
// and is intended to be used by resolver implementations. Incorrect inputs may
// cause panics. Handle with care.
func NewResolvedPath(ipath dms3fspath.Path, c *cid.Cid, root *cid.Cid, remainder string) ResolvedPath {
	return &resolvedPath{
		path:      path{ipath},
		cid:       c,
		root:      root,
		remainder: remainder,
	}
}

func (p *path) String() string {
	return p.path.String()
}

func (p *path) Namespace() string {
	if len(p.path.Segments()) < 1 {
		panic("path without namespace") //this shouldn't happen under any scenario
	}
	return p.path.Segments()[0]
}

func (p *path) Mutable() bool {
	//TODO: MFS: check for /local
	return p.Namespace() == "dms3ns"
}

func (p *resolvedPath) Cid() *cid.Cid {
	return p.cid
}

func (p *resolvedPath) Root() *cid.Cid {
	return p.root
}

func (p *resolvedPath) Remainder() string {
	return p.remainder
}
