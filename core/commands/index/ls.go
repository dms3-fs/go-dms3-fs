package index

import (
	//"context"
	"errors"
	"fmt"
	"io"

	core "github.com/dms3-fs/go-dms3-fs/core"
	cmdenv "github.com/dms3-fs/go-dms3-fs/core/commands/cmdenv"
	e "github.com/dms3-fs/go-dms3-fs/core/commands/e"

	cmds "github.com/dms3-fs/go-fs-cmds"
	cmdkit "github.com/dms3-fs/go-fs-cmdkit"
    dsquery "github.com/dms3-fs/go-datastore/query"
	idxkvs "github.com/dms3-fs/go-dms3-fs/core/coreindex/kvs"
	logging "github.com/dms3-fs/go-log"
)

// log is the command logger
var log = logging.Logger("index")

// ErrDepthLimitExceeded indicates that the max depth has been exceeded.
var ErrNotYetImplemented = fmt.Errorf("not yet implemented")

type ReposetRef struct {
	Infoclass string
	Reposetkind string
	Reposetname string
	Reposetpath string
}

type ReposetRefList []ReposetRef

var ListIndexCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "List index repositories.",
		ShortDescription: `
Returns the list of local index repositories.
`,
		LongDescription: `
Returns the list of local index repositories.
By default, all repositories are returned.

Use the '--kind' flag to match a specific repository kind.
Use the '--name' flag to match a specific repository name.
Use the '--meta' flag to list metastore repositories.
Use the '--data' flag to list infostore repositories.
Use the '--offset' flag to specify result starting page offset.
Use the '--length' flag to specify length of each result page.
`,
	},

	Arguments: []cmdkit.Argument{
	},
	Options: []cmdkit.Option{
		cmdkit.StringOption(kindOptionName, "k", "Kind of repository to list."),
		cmdkit.StringOption(nameOptionName, "n", "Name of repository to list."),
		cmdkit.BoolOption(metaOptionName, "m", "List metastore repositories.").WithDefault(true),
		cmdkit.BoolOption(dataOptionName, "d", "List infostore repositories.").WithDefault(true),
		cmdkit.IntOption(offsetOptionName, "p", "Page offset.").WithDefault(0),
		cmdkit.IntOption(lengthOptionName, "l", "Page length.").WithDefault(24),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) {
		kopt, _ := req.Options[kindOptionName].(string)
		nopt, _ := req.Options[nameOptionName].(string)
		mopt, _ := req.Options[metaOptionName].(bool)
		dopt, _ := req.Options[dataOptionName].(bool)
		popt, _ := req.Options[offsetOptionName].(int)
		lopt, _ := req.Options[lengthOptionName].(int)
		log.Debugf("kind option %s", kopt)
		log.Debugf("name option %s", nopt)
		log.Debugf("meta option %v", mopt)
		log.Debugf("data option %v", dopt)
		log.Debugf("offset option %v", popt)
		log.Debugf("length option %v", lopt)

		n, err := cmdenv.GetNode(env)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		output, err := listRepo(n, kopt, nopt, mopt, dopt, popt, lopt)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}
		cmds.EmitOnce(res, output)

		log.Debugf("output %s", output)

	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeEncoder(func(req *cmds.Request, w io.Writer, v interface{}) error {
			list, ok := v.(ReposetRefList)
			if !ok {
				return e.TypeErr(list, v)
			}

			for i, _ := range list {
				if _, err := fmt.Fprintf(w, "%v %v %v %v\n", list[i].Infoclass, list[i].Reposetkind, list[i].Reposetname, list[i].Reposetpath); err != nil {
					return err
				}
			}
			return nil
		}),
	},
	Type: ReposetRefList{},
}

func listRepo(n *core.Dms3FsNode, kind, reposetname string, wantmeta bool, wantdata bool, startpage, pagesize int) (ReposetRefList, error) {

	// set the KV store to use
	idxkvs.InitIndexKVStore(n.Repo.Datastore())
	dstore := idxkvs.GetIndexKVStore()

	nresult, readCount, readPage := 0, 0, 0
	rlist := []ReposetRef{}

    if res, err := dstore.Query(dsquery.Query{}); err != nil {
		return rlist, errors.New(fmt.Sprintf("cannot issue Query request %v\n.", err))
    } else {
        defer res.Close()
OuterLoop:
        for {
            select {
            case result, ok := <-res.Next():
                if !ok {
                    // no more left
                    break OuterLoop
                }
                if result.Error != nil {
					return rlist, errors.New(fmt.Sprintf("Query returned internal error %v\n.", err))
                }
				// count entries, pages read
				readCount += 1
				if readCount > pagesize {
					readCount = 0
					readPage += 1
				}
				// if on correct page
				if readPage >= startpage {
					// read next key, value
					//rtype, rkind, rname, err := idxkvs.DecomposeRepoSetKey(result.Key)
					rtype, rkind, rname, err := idxkvs.DecomposeRepoSetKey(result.Key)
					if err != nil {
						return rlist, errors.New(fmt.Sprintf("Query returned invalid reposet key error %v\n.", err))
					}
					// if entry matches option filters
					if ((wantmeta && rtype == "metastore") || (wantdata && rtype == "infostore")) &&
						(kind == "" || rkind == "" || kind == rkind) && (reposetname == "" || reposetname == rname) {

						// add entry to results
						r := idxkvs.NewRps()
		                if err := r.Unmarshal(result.Value); err != nil {
							return rlist, errors.New(fmt.Sprintf("cannot unmarshal Query returned value."))
		                }
						rr := ReposetRef{
							Infoclass: rtype,
							Reposetkind: rkind,
							Reposetname: rname,
							Reposetpath: r.GetCid().String(),
						}
						rlist = append(rlist, rr)
						// count results and honor requested limit
						nresult += 1
						if nresult == pagesize {
							break OuterLoop
						}
					}
				}
            }
        }
    }

	return rlist, nil
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
	Type: ReposetRefList{},
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
