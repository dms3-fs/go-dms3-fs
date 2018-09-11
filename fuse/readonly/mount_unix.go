// +build linux darwin freebsd netbsd openbsd
// +build !nofuse

package readonly

import (
	core "github.com/dms3-fs/go-dms3-fs/core"
	mount "github.com/dms3-fs/go-dms3-fs/fuse/mount"
)

// Mount mounts DMS3FS at a given location, and returns a mount.Mount instance.
func Mount(dms3fs *core.Dms3FsNode, mountpoint string) (mount.Mount, error) {
	cfg, err := dms3fs.Repo.Config()
	if err != nil {
		return nil, err
	}
	allow_other := cfg.Mounts.FuseAllowOther
	fsys := NewFileSystem(dms3fs)
	return mount.NewMount(dms3fs.Process(), fsys, mountpoint, allow_other)
}
