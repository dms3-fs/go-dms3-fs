package name

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

    cmdenv "github.com/dms3-fs/go-dms3-fs/core/commands/cmdenv"
    e "github.com/dms3-fs/go-dms3-fs/core/commands/e"
    namesys "github.com/dms3-fs/go-dms3-fs/namesys"
    nsopts "github.com/dms3-fs/go-dms3-fs/namesys/opts"

    "github.com/dms3-fs/go-fs-cmdkit"
    cmds "github.com/dms3-fs/go-fs-cmds"
    offline "github.com/dms3-fs/go-fs-routing/offline"
    logging "github.com/dms3-fs/go-log"
    path "github.com/dms3-fs/go-path"
)

var log = logging.Logger("core/commands/dms3ns")

type ResolvedPath struct {
	Path path.Path
}

var Dms3NsCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Resolve DMS3NS names.",
		ShortDescription: `
DMS3NS is a PKI namespace, where names are the hashes of public keys, and
the private key enables publishing new (signed) values. In both publish
and resolve, the default name used is the node's own PeerID,
which is the hash of its public key.
`,
		LongDescription: `
DMS3NS is a PKI namespace, where names are the hashes of public keys, and
the private key enables publishing new (signed) values. In both publish
and resolve, the default name used is the node's own PeerID,
which is the hash of its public key.

You can use the 'dms3fs key' commands to list and generate more names and their
respective keys.

Examples:

Resolve the value of your name:

  > dms3fs name resolve
  /dms3fs/QmatmE9msSfkKxoffpHwNLNKgwZG8eT9Bud6YoPab52vpy

Resolve the value of another name:

  > dms3fs name resolve QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ
  /dms3fs/QmSiTko9JZyabH56y2fussEt1A5oDqsFXB3CkvAqraFryz

Resolve the value of a dnslink:

  > dms3fs name resolve dms3.io
  /dms3fs/QmaBvfZooxWkrv7D3r8LS9moNjzD2o525XMZze69hhoxf5

`,
	},

	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("name", false, false, "The DMS3NS name to resolve. Defaults to your node's peerID."),
	},
	Options: []cmdkit.Option{
		cmdkit.BoolOption("recursive", "r", "Resolve until the result is not an DMS3NS name."),
		cmdkit.BoolOption("nocache", "n", "Do not use cached entries."),
		cmdkit.UintOption("dht-record-count", "dhtrc", "Number of records to request for DHT resolution."),
		cmdkit.StringOption("dht-timeout", "dhtt", "Max time to collect values during DHT resolution eg \"30s\". Pass 0 for no timeout."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) {
		n, err := cmdenv.GetNode(env)
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

		nocache, _ := req.Options["nocache"].(bool)
		local, _ := req.Options["local"].(bool)

		// default to nodes namesys resolver
		var resolver namesys.Resolver = n.Namesys

		if local && nocache {
			res.SetError(errors.New("cannot specify both local and nocache"), cmdkit.ErrNormal)
			return
		}

		if local {
			offroute := offline.NewOfflineRouter(n.Repo.Datastore(), n.RecordValidator)
			resolver = namesys.NewDms3NsResolver(offroute)
		}

		if nocache {
			resolver = namesys.NewNameSystem(n.Routing, n.Repo.Datastore(), 0)
		}

		var name string
		if len(req.Arguments) == 0 {
			if n.Identity == "" {
				res.SetError(errors.New("identity not loaded"), cmdkit.ErrNormal)
				return
			}
			name = n.Identity.Pretty()

		} else {
			name = req.Arguments[0]
		}

		recursive, _ := req.Options["recursive"].(bool)
		rc, rcok := req.Options["dht-record-count"].(int)
		dhtt, dhttok := req.Options["dht-timeout"].(string)
		var ropts []nsopts.ResolveOpt
		if !recursive {
			ropts = append(ropts, nsopts.Depth(1))
		}
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

		if !strings.HasPrefix(name, "/dms3ns/") {
			name = "/dms3ns/" + name
		}

		output, err := resolver.Resolve(req.Context, name, ropts...)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		// TODO: better errors (in the case of not finding the name, we get "failed to find any peer in table")

		cmds.EmitOnce(res, &ResolvedPath{output})
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeEncoder(func(req *cmds.Request, w io.Writer, v interface{}) error {
			output, ok := v.(*ResolvedPath)
			if !ok {
				return e.TypeErr(output, v)
			}
			_, err := fmt.Fprintln(w, output.Path)
			return err
		}),
	},
	Type: ResolvedPath{},
}
