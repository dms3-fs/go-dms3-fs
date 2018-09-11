package dms3ns

import (
	"context"

    "github.com/dms3-fs/go-dms3-fs/core"
    nsys "github.com/dms3-fs/go-dms3-fs/namesys"
    path "github.com/dms3-fs/go-path"
    ft "github.com/dms3-fs/go-unixfs"
    ci "github.com/dms3-p2p/go-p2p-crypto"
)

// InitializeKeyspace sets the dms3ns record for the given key to
// point to an empty directory.
func InitializeKeyspace(n *core.Dms3FsNode, key ci.PrivKey) error {
	ctx, cancel := context.WithCancel(n.Context())
	defer cancel()

	emptyDir := ft.EmptyDirNode()

	err := n.Pinning.Pin(ctx, emptyDir, false)
	if err != nil {
		return err
	}

	err = n.Pinning.Flush()
	if err != nil {
		return err
	}

	pub := nsys.NewDms3NsPublisher(n.Routing, n.Repo.Datastore())

	return pub.Publish(ctx, key, path.FromCid(emptyDir.Cid()))
}
