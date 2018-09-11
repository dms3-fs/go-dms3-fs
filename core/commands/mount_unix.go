// +build !windows,!nofuse

package commands

import (
	"fmt"
	"io"
	"strings"

	cmds "github.com/dms3-fs/go-dms3-fs/commands"
	e "github.com/dms3-fs/go-dms3-fs/core/commands/e"
	nodeMount "github.com/dms3-fs/go-dms3-fs/fuse/node"

	"github.com/dms3-fs/go-fs-cmdkit"
	config "github.com/dms3-fs/go-fs-config"
)

var MountCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Mounts DMS3FS to the filesystem (read-only).",
		ShortDescription: `
Mount DMS3FS at a read-only mountpoint on the OS (default: /dms3fs and /dms3ns).
All DMS3FS objects will be accessible under that directory. Note that the
root will not be listable, as it is virtual. Access known paths directly.

You may have to create /dms3fs and /dms3ns before using 'dms3fs mount':

> sudo mkdir /dms3fs /dms3ns
> sudo chown $(whoami) /dms3fs /dms3ns
> dms3fs daemon &
> dms3fs mount
`,
		LongDescription: `
Mount DMS3FS at a read-only mountpoint on the OS. The default, /dms3fs and /dms3ns,
are set in the configuration file, but can be overriden by the options.
All DMS3FS objects will be accessible under this directory. Note that the
root will not be listable, as it is virtual. Access known paths directly.

You may have to create /dms3fs and /dms3ns before using 'dms3fs mount':

> sudo mkdir /dms3fs /dms3ns
> sudo chown $(whoami) /dms3fs /dms3ns
> dms3fs daemon &
> dms3fs mount

Example:

# setup
> mkdir foo
> echo "baz" > foo/bar
> dms3fs add -r foo
added QmWLdkp93sNxGRjnFHPaYg8tCQ35NBY3XPn6KiETd3Z4WR foo/bar
added QmSh5e7S6fdcu75LAbXNZAFY2nGyZUJXyLCJDvn2zRkWyC foo
> dms3fs ls QmSh5e7S6fdcu75LAbXNZAFY2nGyZUJXyLCJDvn2zRkWyC
QmWLdkp93sNxGRjnFHPaYg8tCQ35NBY3XPn6KiETd3Z4WR 12 bar
> dms3fs cat QmWLdkp93sNxGRjnFHPaYg8tCQ35NBY3XPn6KiETd3Z4WR
baz

# mount
> dms3fs daemon &
> dms3fs mount
DMS3FS mounted at: /dms3fs
DMS3NS mounted at: /dms3ns
> cd /dms3fs/QmSh5e7S6fdcu75LAbXNZAFY2nGyZUJXyLCJDvn2zRkWyC
> ls
bar
> cat bar
baz
> cat /dms3fs/QmSh5e7S6fdcu75LAbXNZAFY2nGyZUJXyLCJDvn2zRkWyC/bar
baz
> cat /dms3fs/QmWLdkp93sNxGRjnFHPaYg8tCQ35NBY3XPn6KiETd3Z4WR
baz
`,
	},
	Options: []cmdkit.Option{
		cmdkit.StringOption("dms3fs-path", "f", "The path where DMS3FS should be mounted."),
		cmdkit.StringOption("dms3ns-path", "n", "The path where DMS3NS should be mounted."),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		cfg, err := req.InvocContext().GetConfig()
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		node, err := req.InvocContext().GetNode()
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		// error if we aren't running node in online mode
		if node.LocalMode() {
			res.SetError(ErrNotOnline, cmdkit.ErrClient)
			return
		}

		fsdir, found, err := req.Option("f").String()
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}
		if !found {
			fsdir = cfg.Mounts.DMS3FS // use default value
		}

		// get default mount points
		nsdir, found, err := req.Option("n").String()
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}
		if !found {
			nsdir = cfg.Mounts.DMS3NS // NB: be sure to not redeclare!
		}

		err = nodeMount.Mount(node, fsdir, nsdir)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		var output config.Mounts
		output.DMS3FS = fsdir
		output.DMS3NS = nsdir
		res.SetOutput(&output)
	},
	Type: config.Mounts{},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: func(res cmds.Response) (io.Reader, error) {
			v, err := unwrapOutput(res.Output())
			if err != nil {
				return nil, err
			}

			mnts, ok := v.(*config.Mounts)
			if !ok {
				return nil, e.TypeErr(mnts, v)
			}

			s := fmt.Sprintf("DMS3FS mounted at: %s\n", mnts.DMS3FS)
			s += fmt.Sprintf("DMS3NS mounted at: %s\n", mnts.DMS3NS)
			return strings.NewReader(s), nil
		},
	},
}
