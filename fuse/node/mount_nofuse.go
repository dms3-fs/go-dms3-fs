// +build !windows,nofuse

package node

import (
	"errors"

	core "github.com/dms3-fs/go-dms3-fs/core"
)

func Mount(node *core.Dms3FsNode, fsdir, nsdir string) error {
	return errors.New("not compiled in")
}
