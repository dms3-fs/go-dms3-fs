// +build linux darwin freebsd netbsd openbsd
// +build !nofuse

package dms3ns

import (
	core "github.com/dms3-fs/go-dms3-fs/core"
	mount "github.com/dms3-fs/go-dms3-fs/fuse/mount"
)

// Mount mounts dms3ns at a given location, and returns a mount.Mount instance.
func Mount(dms3fs *core.Dms3FsNode, dms3nsmp, dms3fsmp string) (mount.Mount, error) {
	cfg, err := dms3fs.Repo.Config()
	if err != nil {
		return nil, err
	}

	allow_other := cfg.Mounts.FuseAllowOther

	fsys, err := NewFileSystem(dms3fs, dms3fs.PrivateKey, dms3fsmp, dms3nsmp)
	if err != nil {
		return nil, err
	}

	return mount.NewMount(dms3fs.Process(), fsys, dms3nsmp, allow_other)
}
