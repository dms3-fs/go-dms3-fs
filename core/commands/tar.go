package commands

import (
	"io"
	"strings"

	cmds "github.com/dms3-fs/go-dms3-fs/commands"
	core "github.com/dms3-fs/go-dms3-fs/core"
	e "github.com/dms3-fs/go-dms3-fs/core/commands/e"
	"github.com/dms3-fs/go-dms3-fs/core/coreunix"
	tar "github.com/dms3-fs/go-dms3-fs/tar"
	dag "github.com/dms3-fs/go-merkledag"
	path "github.com/dms3-fs/go-path"

	"github.com/dms3-fs/go-fs-cmdkit"
)

var TarCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Utility functions for tar files in dms3fs.",
	},

	Subcommands: map[string]*cmds.Command{
		"add": tarAddCmd,
		"cat": tarCatCmd,
	},
}

var tarAddCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Import a tar file into dms3fs.",
		ShortDescription: `
'dms3fs tar add' will parse a tar file and create a merkledag structure to
represent it.
`,
	},

	Arguments: []cmdkit.Argument{
		cmdkit.FileArg("file", true, false, "Tar file to add.").EnableStdin(),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		nd, err := req.InvocContext().GetNode()
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		fi, err := req.Files().NextFile()
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		node, err := tar.ImportTar(req.Context(), fi, nd.DAG)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		c := node.Cid()

		fi.FileName()
		res.SetOutput(&coreunix.AddedObject{
			Name: fi.FileName(),
			Hash: c.String(),
		})
	},
	Type: coreunix.AddedObject{},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: func(res cmds.Response) (io.Reader, error) {
			v, err := unwrapOutput(res.Output())
			if err != nil {
				return nil, err
			}

			o, ok := v.(*coreunix.AddedObject)
			if !ok {
				return nil, e.TypeErr(o, v)
			}
			return strings.NewReader(o.Hash + "\n"), nil
		},
	},
}

var tarCatCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Export a tar file from DMS3FS.",
		ShortDescription: `
'dms3fs tar cat' will export a tar file from a previously imported one in DMS3FS.
`,
	},

	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("path", true, false, "dms3fs path of archive to export.").EnableStdin(),
	},
	Run: func(req cmds.Request, res cmds.Response) {
		nd, err := req.InvocContext().GetNode()
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		p, err := path.ParsePath(req.Arguments()[0])
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		root, err := core.Resolve(req.Context(), nd.Namesys, nd.Resolver, p)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		rootpb, ok := root.(*dag.ProtoNode)
		if !ok {
			res.SetError(dag.ErrNotProtobuf, cmdkit.ErrNormal)
			return
		}

		r, err := tar.ExportTar(req.Context(), rootpb, nd.DAG)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		res.SetOutput(r)
	},
}
