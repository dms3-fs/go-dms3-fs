/*
Package coreapi provides direct access to the core commands in DMS3FS. If you are
embedding DMS3FS directly in your Go program, this package is the public
interface you should use to read and write files or otherwise control DMS3FS.

If you are running DMS3FS as a separate process, you should use `go-dms3fs-api` to
work with it via HTTP. As we finalize the interfaces here, `go-dms3fs-api` will
transparently adopt them so you can use the same code with either package.

**NOTE: this package is experimental.** `go-dms3-fs` has mainly been developed
as a standalone application and library-style use of this package is still new.
Interfaces here aren't yet completely stable.
*/
package coreapi

import (
	core "github.com/dms3-fs/go-dms3-fs/core"
	coreiface "github.com/dms3-fs/go-dms3-fs/core/coreapi/interface"
)

type CoreAPI struct {
	node *core.Dms3FsNode
}

// NewCoreAPI creates new instance of DMS3FS CoreAPI backed by go-dms3fs Node.
func NewCoreAPI(n *core.Dms3FsNode) coreiface.CoreAPI {
	api := &CoreAPI{n}
	return api
}

// Unixfs returns the UnixfsAPI interface implementation backed by the go-dms3fs node
func (api *CoreAPI) Unixfs() coreiface.UnixfsAPI {
	return (*UnixfsAPI)(api)
}

// Block returns the BlockAPI interface implementation backed by the go-dms3fs node
func (api *CoreAPI) Block() coreiface.BlockAPI {
	return (*BlockAPI)(api)
}

// Dag returns the DagAPI interface implementation backed by the go-dms3fs node
func (api *CoreAPI) Dag() coreiface.DagAPI {
	return (*DagAPI)(api)
}

// Name returns the NameAPI interface implementation backed by the go-dms3fs node
func (api *CoreAPI) Name() coreiface.NameAPI {
	return (*NameAPI)(api)
}

// Key returns the KeyAPI interface implementation backed by the go-dms3fs node
func (api *CoreAPI) Key() coreiface.KeyAPI {
	return (*KeyAPI)(api)
}

// Object returns the ObjectAPI interface implementation backed by the go-dms3fs node
func (api *CoreAPI) Object() coreiface.ObjectAPI {
	return (*ObjectAPI)(api)
}

// Pin returns the PinAPI interface implementation backed by the go-dms3fs node
func (api *CoreAPI) Pin() coreiface.PinAPI {
	return (*PinAPI)(api)
}
