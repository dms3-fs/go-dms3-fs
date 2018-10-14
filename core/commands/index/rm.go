package index

import (
//	"strings"
	"context"
	"errors"
	"fmt"
	"io"
//	"time"

	core "github.com/dms3-fs/go-dms3-fs/core"
	cmdenv "github.com/dms3-fs/go-dms3-fs/core/commands/cmdenv"
	e "github.com/dms3-fs/go-dms3-fs/core/commands/e"

	cmds "github.com/dms3-fs/go-fs-cmds"
	cmdkit "github.com/dms3-fs/go-fs-cmdkit"
	path "github.com/dms3-fs/go-path"
)

var RemoveDocumentCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Remove document from index repository.",
		ShortDescription: `
Remove document from an index repository specified by path.
`,
		LongDescription: `
Remove document specified by cid from index repository specified by path.
`,
	},

	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("cid", true, false, "content to remove from repository."),
		cmdkit.StringArg("dms3fs-path", true, false, "repository to remove from."),
	},
	Options: []cmdkit.Option{
		cmdkit.BoolOption("quiet", "q", "Write just hashes of created object."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) {
		var cid, repo string
        if len(req.Arguments) != 2 {
			res.SetError(errors.New("cid and path are both required."), cmdkit.ErrNormal)
			return
        } else {
			cid = req.Arguments[0]
			repo = req.Arguments[1]
        }

		log.Debugf("cid value is %s, repo path is %s", cid, repo)

		n, err := cmdenv.GetNode(env)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		//log.Error("Running index ls : ", err)
		log.Debugf("Running command request path %s", req.Path)

		rmopts := new(rmdocOpts)

		q, _ := req.Options["quiet"].(bool)
		if q {
			rmopts.q = true
		} else {
			rmopts.q = false
		}
		log.Debugf("quiet option value %t", q)

		ctx := req.Context

/*
		p, err := path.ParsePath(p)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}
*/
		var p path.Path

		output, err := rmDoc(ctx, n, p, rmopts)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}
		cmds.EmitOnce(res, output)

		log.Debugf("output %s", output)

	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeEncoder(func(req *cmds.Request, w io.Writer, v interface{}) error {
			repoPath, ok := v.(*RepoPath)
			if !ok {
				return e.TypeErr(repoPath, v)
			}

			_, err := fmt.Fprintf(w, "path %s \n", repoPath.path)
			return err
		}),
	},
	Type: RepoPath{},
}

type rmdocOpts struct {
	q bool
}

func rmDoc(ctx context.Context, n *core.Dms3FsNode, ref path.Path, opts *rmdocOpts) (*RepoPath, error) {

	return &RepoPath{
		path: "rmdoc test worked.",
	}, nil

}
