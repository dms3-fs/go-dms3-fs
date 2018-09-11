package unixfs

import (
	cmds "github.com/dms3-fs/go-dms3-fs/commands"
	e "github.com/dms3-fs/go-dms3-fs/core/commands/e"

	"github.com/dms3-fs/go-fs-cmdkit"
)

var UnixFSCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Interact with DMS3FS objects representing Unix filesystems.",
		ShortDescription: `
'dms3fs file' provides a familiar interface to file systems represented
by DMS3FS objects, which hides dms3fs implementation details like layout
objects (e.g. fanout and chunking).
`,
		LongDescription: `
'dms3fs file' provides a familiar interface to file systems represented
by DMS3FS objects, which hides dms3fs implementation details like layout
objects (e.g. fanout and chunking).
`,
	},

	Subcommands: map[string]*cmds.Command{
		"ls": LsCmd,
	},
}

// copy+pasted from ../commands.go
func unwrapOutput(i interface{}) (interface{}, error) {
	var (
		ch <-chan interface{}
		ok bool
	)

	if ch, ok = i.(<-chan interface{}); !ok {
		return nil, e.TypeErr(ch, i)
	}

	return <-ch, nil
}
