// +build !windows,nofuse

package commands

import (
	cmds "github.com/dms3-fs/go-dms3-fs/commands"

	"github.com/dms3-fs/go-fs-cmdkit"
)

var MountCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Mounts dms3fs to the filesystem (disabled).",
		ShortDescription: `
This version of dms3fs is compiled without fuse support, which is required
for mounting. If you'd like to be able to mount, please use a version of
dms3fs compiled with fuse.

For the latest instructions, please check the project's repository:
  http://github.com/dms3-fs/go-dms3-fs
`,
	},
}
