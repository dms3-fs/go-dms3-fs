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

var PublishIndexCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Publish index repository.",
		ShortDescription: `
Publish index repository specified by path.
`,
		LongDescription: `
Publish index repository specified by path.
`,
	},

	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("dms3fs-path", true, false, "repository to publish."),
	},
	Options: []cmdkit.Option{
		cmdkit.BoolOption("quiet", "q", "Write just hashes of created object."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) {
		var repo string
        if len(req.Arguments) != 1 {
			res.SetError(errors.New("path is required."), cmdkit.ErrNormal)
			return
        } else {
			repo = req.Arguments[0]
        }

		log.Debugf("repo path is %s", repo)

		n, err := cmdenv.GetNode(env)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		//log.Error("Running index ls : ", err)
		log.Debugf("Running command request path %s", req.Path)

		pubopts := new(publishOpts)

		q, _ := req.Options["quiet"].(bool)
		if q {
			pubopts.q = true
		} else {
			pubopts.q = false
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

		output, err := pubRepo(ctx, n, p, pubopts)
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

type publishOpts struct {
	q bool
}

func pubRepo(ctx context.Context, n *core.Dms3FsNode, ref path.Path, opts *publishOpts) (*RepoPath, error) {

	return &RepoPath{
		path: "publish test worked.",
	}, nil

}
