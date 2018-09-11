package commands

import (
	"io"
	"strings"

	cmds "github.com/dms3-fs/go-dms3-fs/commands"
	e "github.com/dms3-fs/go-dms3-fs/core/commands/e"
	ncmd "github.com/dms3-fs/go-dms3-fs/core/commands/name"
	namesys "github.com/dms3-fs/go-dms3-fs/namesys"
	nsopts "github.com/dms3-fs/go-dms3-fs/namesys/opts"

	"github.com/dms3-fs/go-fs-cmdkit"
)

var DNSCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Resolve DNS links.",
		ShortDescription: `
Multihashes are hard to remember, but domain names are usually easy to
remember.  To create memorable aliases for multihashes, DNS TXT
records can point to other DNS links, DMS3FS objects, DMS3NS keys, etc.
This command resolves those links to the referenced object.
`,
		LongDescription: `
Multihashes are hard to remember, but domain names are usually easy to
remember.  To create memorable aliases for multihashes, DNS TXT
records can point to other DNS links, DMS3FS objects, DMS3NS keys, etc.
This command resolves those links to the referenced object.

For example, with this DNS TXT record:

	> dig +short TXT _dnslink.dms3.io
	dnslink=/dms3fs/QmRzTuh2Lpuz7Gr39stNr6mTFdqAghsZec1JoUnfySUzcy

The resolver will give:

	> dms3fs dns dms3.io
	/dms3fs/QmRzTuh2Lpuz7Gr39stNr6mTFdqAghsZec1JoUnfySUzcy

The resolver can recursively resolve:

	> dig +short TXT recursive.dms3.io
	dnslink=/dms3ns/dms3.io
	> dms3fs dns -r recursive.dms3.io
	/dms3fs/QmRzTuh2Lpuz7Gr39stNr6mTFdqAghsZec1JoUnfySUzcy
`,
	},

	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("domain-name", true, false, "The domain-name name to resolve.").EnableStdin(),
	},
	Options: []cmdkit.Option{
		cmdkit.BoolOption("recursive", "r", "Resolve until the result is not a DNS link."),
	},
	Run: func(req cmds.Request, res cmds.Response) {

		recursive, _, _ := req.Option("recursive").Bool()
		name := req.Arguments()[0]
		resolver := namesys.NewDNSResolver()

		var ropts []nsopts.ResolveOpt
		if !recursive {
			ropts = append(ropts, nsopts.Depth(1))
		}

		output, err := resolver.Resolve(req.Context(), name, ropts...)
		if err == namesys.ErrResolveFailed {
			res.SetError(err, cmdkit.ErrNotFound)
			return
		}
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}
		res.SetOutput(&ncmd.ResolvedPath{Path: output})
	},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: func(res cmds.Response) (io.Reader, error) {
			v, err := unwrapOutput(res.Output())
			if err != nil {
				return nil, err
			}

			output, ok := v.(*ncmd.ResolvedPath)
			if !ok {
				return nil, e.TypeErr(output, v)
			}
			return strings.NewReader(output.Path.String() + "\n"), nil
		},
	},
	Type: ncmd.ResolvedPath{},
}
