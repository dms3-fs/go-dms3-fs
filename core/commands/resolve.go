package commands

import (
	"errors"
	"io"
	"strings"
	"time"

	cmds "github.com/dms3-fs/go-dms3-fs/commands"
	"github.com/dms3-fs/go-dms3-fs/core"
	e "github.com/dms3-fs/go-dms3-fs/core/commands/e"
	ncmd "github.com/dms3-fs/go-dms3-fs/core/commands/name"
	ns "github.com/dms3-fs/go-dms3-fs/namesys"
	nsopts "github.com/dms3-fs/go-dms3-fs/namesys/opts"
	path "github.com/dms3-fs/go-path"

	"github.com/dms3-fs/go-fs-cmdkit"
)

var ResolveCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Resolve the value of names to DMS3FS.",
		ShortDescription: `
There are a number of mutable name protocols that can link among
themselves and into DMS3NS. This command accepts any of these
identifiers and resolves them to the referenced item.
`,
		LongDescription: `
There are a number of mutable name protocols that can link among
themselves and into DMS3NS. For example DMS3NS references can (currently)
point at an DMS3FS object, and DNS links can point at other DNS links, DMS3NS
entries, or DMS3FS objects. This command accepts any of these
identifiers and resolves them to the referenced item.

EXAMPLES

Resolve the value of your identity:

  $ dms3fs resolve /dms3ns/QmatmE9msSfkKxoffpHwNLNKgwZG8eT9Bud6YoPab52vpy
  /dms3fs/Qmcqtw8FfrVSBaRmbWwHxt3AuySBhJLcvmFYi3Lbc4xnwj

Resolve the value of another name:

  $ dms3fs resolve /dms3ns/QmbCMUZw6JFeZ7Wp9jkzbye3Fzp2GGcPgC3nmeUjfVF87n
  /dms3ns/QmatmE9msSfkKxoffpHwNLNKgwZG8eT9Bud6YoPab52vpy

Resolve the value of another name recursively:

  $ dms3fs resolve -r /dms3ns/QmbCMUZw6JFeZ7Wp9jkzbye3Fzp2GGcPgC3nmeUjfVF87n
  /dms3fs/Qmcqtw8FfrVSBaRmbWwHxt3AuySBhJLcvmFYi3Lbc4xnwj

Resolve the value of an DMS3FS DAG path:

  $ dms3fs resolve /dms3fs/QmeZy1fGbwgVSrqbfh9fKQrAWgeyRnj7h8fsHS1oy3k99x/beep/boop
  /dms3fs/QmYRMjyvAiHKN9UTi8Bzt1HUspmSRD8T8DwxfSMzLgBon1

`,
	},

	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("name", true, false, "The name to resolve.").EnableStdin(),
	},
	Options: []cmdkit.Option{
		cmdkit.BoolOption("recursive", "r", "Resolve until the result is an DMS3FS name."),
		cmdkit.UintOption("dht-record-count", "dhtrc", "Number of records to request for DHT resolution."),
		cmdkit.StringOption("dht-timeout", "dhtt", "Max time to collect values during DHT resolution eg \"30s\". Pass 0 for no timeout."),
	},
	Run: func(req cmds.Request, res cmds.Response) {

		n, err := req.InvocContext().GetNode()
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		if !n.OnlineMode() {
			err := n.SetupOfflineRouting()
			if err != nil {
				res.SetError(err, cmdkit.ErrNormal)
				return
			}
		}

		name := req.Arguments()[0]
		recursive, _, _ := req.Option("recursive").Bool()

		// the case when dms3ns is resolved step by step
		if strings.HasPrefix(name, "/dms3ns/") && !recursive {
			rc, rcok, _ := req.Option("dht-record-count").Int()
			dhtt, dhttok, _ := req.Option("dht-timeout").String()
			ropts := []nsopts.ResolveOpt{nsopts.Depth(1)}
			if rcok {
				ropts = append(ropts, nsopts.DhtRecordCount(uint(rc)))
			}
			if dhttok {
				d, err := time.ParseDuration(dhtt)
				if err != nil {
					res.SetError(err, cmdkit.ErrNormal)
					return
				}
				if d < 0 {
					res.SetError(errors.New("DHT timeout value must be >= 0"), cmdkit.ErrNormal)
					return
				}
				ropts = append(ropts, nsopts.DhtTimeout(d))
			}
			p, err := n.Namesys.Resolve(req.Context(), name, ropts...)
			// ErrResolveRecursion is fine
			if err != nil && err != ns.ErrResolveRecursion {
				res.SetError(err, cmdkit.ErrNormal)
				return
			}
			res.SetOutput(&ncmd.ResolvedPath{Path: p})
			return
		}

		// else, dms3fs path or dms3ns with recursive flag
		p, err := path.ParsePath(name)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		node, err := core.Resolve(req.Context(), n.Namesys, n.Resolver, p)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		c := node.Cid()

		res.SetOutput(&ncmd.ResolvedPath{Path: path.FromCid(c)})
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
