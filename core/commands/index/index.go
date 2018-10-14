package index

import (
	"strings"
	"context"
//	"errors"
	"fmt"
	"io"
//	"time"

	core "github.com/dms3-fs/go-dms3-fs/core"
	cmdenv "github.com/dms3-fs/go-dms3-fs/core/commands/cmdenv"
	e "github.com/dms3-fs/go-dms3-fs/core/commands/e"

	cmds "github.com/dms3-fs/go-fs-cmds"
	cmdkit "github.com/dms3-fs/go-fs-cmdkit"
	path "github.com/dms3-fs/go-path"
	logging "github.com/dms3-fs/go-log"
)

// log is the command logger
var log = logging.Logger("index")

// ErrDepthLimitExceeded indicates that the max depth has been exceeded.
var ErrNotYetImplemented = fmt.Errorf("not yet implemented")

type RepoList struct {
	Name string
	Id   string
}

var ListIndexCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "List index data repositories.",
		ShortDescription: `
Returns a list of data repositories at the given path.
By default, all repositories are returned, but the '--meta' flag or
arguments can restrict that to a specific repositories.
`,
		LongDescription: `
Returns a list of data repositories at the given path.
By default, all repositories are returned, but the '--meta' flag or
arguments can restrict that to a specific repositories.

Use --meta=<keyword> to specify repositories with metadata
containing keyword to list.
`,
	},

	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("dms3fs-path", true, false, "dms3fs path to repository to be listed..").EnableStdin(),
	},
	Options: []cmdkit.Option{
		cmdkit.StringOption("meta", "m", "The metadata keyword of repositories to list."),
		cmdkit.BoolOption("quiet", "q", "Write just hashes of objects."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) {
		n, err := cmdenv.GetNode(env)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		//log.Error("Running index ls : ", err)
		log.Debugf("Running command request path %s", req.Path)

		details := commandDetails(req.Path)
		log.Debugf("cmdDetails cannotRunOnClient %t", details.cannotRunOnClient)
		log.Debugf("cmdDetails cannotRunOnDaemon %t", details.cannotRunOnDaemon)
		log.Debugf("cmdDetails doesNotUseRepo %t", details.doesNotUseRepo)
		log.Debugf("cmdDetails usesConfigAsInput() %t", details.usesConfigAsInput())

		pstr := req.Arguments[0]
		log.Debugf("path param %s", pstr)

		lopts := new(listOpts)

		lopts.meta, _ = req.Options["meta"].(string)
		log.Debugf("meta option %s", pstr)

		ctx := req.Context

		pth, err := path.ParsePath(pstr)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		output, err := listRepo(ctx, n, pth, lopts)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}
		cmds.EmitOnce(res, output)

		log.Debugf("output %s", output)

	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeEncoder(func(req *cmds.Request, w io.Writer, v interface{}) error {
			list, ok := v.(*RepoList)
			if !ok {
				return e.TypeErr(list, v)
			}

			_, err := fmt.Fprintf(w, "Listed %s: %s\n", list.Name, list.Id)
			return err
		}),
	},
	Type: RepoList{},
}

type listOpts struct {
	meta string
}

func listRepo(ctx context.Context, n *core.Dms3FsNode, ref path.Path, opts *listOpts) (*RepoList, error) {

	return &RepoList{
		Name:  "test name",
		Id: "test id",
	}, nil

}


// commandDetails returns a command's details for the command given by |path|.
func commandDetails(path []string) *cmdDetails {
	var details cmdDetails
	// find the last command in path that has a cmdDetailsMap entry
	for i := range path {
		if cmdDetails, found := cmdDetailsMap[strings.Join(path[:i+1], "/")]; found {
			details = cmdDetails
		}
	}
	return &details
}

// NB: when necessary, properties are described using negatives in order to
// provide desirable defaults
type cmdDetails struct {
	cannotRunOnClient bool
	cannotRunOnDaemon bool
	doesNotUseRepo    bool

	// doesNotUseConfigAsInput describes commands that do not use the config as
	// input. These commands either initialize the config or perform operations
	// that don't require access to the config.
	//
	// pre-command hooks that require configs must not be run before these
	// commands.
	doesNotUseConfigAsInput bool

	// preemptsAutoUpdate describes commands that must be executed without the
	// auto-update pre-command hook
	preemptsAutoUpdate bool
}

func (d *cmdDetails) String() string {
	return fmt.Sprintf("on client? %t, on daemon? %t, uses repo? %t",
		d.canRunOnClient(), d.canRunOnDaemon(), d.usesRepo())
}

func (d *cmdDetails) Loggable() map[string]interface{} {
	return map[string]interface{}{
		"canRunOnClient":     d.canRunOnClient(),
		"canRunOnDaemon":     d.canRunOnDaemon(),
		"preemptsAutoUpdate": d.preemptsAutoUpdate,
		"usesConfigAsInput":  d.usesConfigAsInput(),
		"usesRepo":           d.usesRepo(),
	}
}

func (d *cmdDetails) usesConfigAsInput() bool { return !d.doesNotUseConfigAsInput }
func (d *cmdDetails) canRunOnClient() bool    { return !d.cannotRunOnClient }
func (d *cmdDetails) canRunOnDaemon() bool    { return !d.cannotRunOnDaemon }
func (d *cmdDetails) usesRepo() bool          { return !d.doesNotUseRepo }

// "What is this madness!?" you ask. Our commands have the unfortunate problem of
// not being able to run on all the same contexts. This map describes these
// properties so that other code can make decisions about whether to invoke a
// command or return an error to the user.
var cmdDetailsMap = map[string]cmdDetails{
	"init":        {doesNotUseConfigAsInput: true, cannotRunOnDaemon: true, doesNotUseRepo: true},
	"daemon":      {doesNotUseConfigAsInput: true, cannotRunOnDaemon: true},
	"commands":    {doesNotUseRepo: true},
	"version":     {doesNotUseConfigAsInput: true, doesNotUseRepo: true}, // must be permitted to run before init
	"log":         {cannotRunOnClient: true},
	"diag/cmds":   {cannotRunOnClient: true},
	"repo/fsck":   {cannotRunOnDaemon: true},
	"config/edit": {cannotRunOnDaemon: true, doesNotUseRepo: true},
}

var NotyetIndexCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Not yet implemented.",
		ShortDescription: `
Not yet implemented.
`,
		LongDescription: `
Not yet implemented.
`,
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) {

		defer res.Close()

		res.Emit(fmt.Sprintf("repos: %s", "not implemented yet..."))

	},
	Type: RepoList{},
	PostRun: cmds.PostRunMap{
		cmds.CLI: func(req *cmds.Request, re cmds.ResponseEmitter) cmds.ResponseEmitter {
			reNext, res := cmds.NewChanResponsePair(req)
			outChan := make(chan interface{})

			go func() {
				// defer order important! First close outChan, then wait for output to finish, then close re
				defer re.Close()

				if e := res.Error(); e != nil {
					defer close(outChan)
					re.SetError(e.Message, e.Code)
					return
				}

				defer close(outChan)

				for {
					v, err := res.Next()
					if !cmds.HandleError(err, res, re) {
						break
					}

					select {
					case outChan <- v:
					case <-req.Context.Done():
						re.SetError(req.Context.Err(), cmdkit.ErrNormal)
						return
					}
				}
			}()

			return reNext
		},
	},
}
