package index

import (
	"errors"
	"fmt"
	"io"
	"os"
    "path"
    "path/filepath"
	"strconv"
	"strings"
	"time"

	blockservice "github.com/dms3-fs/go-blockservice"
	bstore "github.com/dms3-fs/go-fs-blockstore"
	cid "github.com/dms3-fs/go-cid"
	cidutil "github.com/dms3-fs/go-cidutil"
	cmdenv "github.com/dms3-fs/go-dms3-fs/core/commands/cmdenv"
	cmdkit "github.com/dms3-fs/go-fs-cmdkit"
	cmds "github.com/dms3-fs/go-fs-cmds"
	core "github.com/dms3-fs/go-dms3-fs/core"
	coreiface "github.com/dms3-fs/go-dms3-fs/core/coreapi/interface"
	coreunix "github.com/dms3-fs/go-dms3-fs/core/coreunix"
	dag "github.com/dms3-fs/go-merkledag"
	dagtest "github.com/dms3-fs/go-merkledag/test"
    dms3ld "github.com/dms3-fs/go-ld-format"
	ds "github.com/dms3-fs/go-datastore"
	e "github.com/dms3-fs/go-dms3-fs/core/commands/e"
	files "github.com/dms3-fs/go-fs-cmdkit/files"
	filestore "github.com/dms3-fs/go-dms3-fs/filestore"
	ft "github.com/dms3-fs/go-unixfs"
	idxlfs "github.com/dms3-fs/go-dms3-fs/core/coreindex/lfs"
	idxkvs "github.com/dms3-fs/go-dms3-fs/core/coreindex/kvs"
	idxufs "github.com/dms3-fs/go-dms3-fs/core/coreindex/ufs"
	mfs "github.com/dms3-fs/go-mfs"
	mh "github.com/dms3-mft/go-multihash"
	offline "github.com/dms3-fs/go-fs-exchange-offline"
	dms3fspath "github.com/dms3-fs/go-path"
	pb "github.com/cheggaaa/pb"
	"github.com/dms3-fs/go-dms3-fs/pin"
	resolver "github.com/dms3-fs/go-path/resolver"
	uio "github.com/dms3-fs/go-unixfs/io"
)

// ErrDepthLimitExceeded indicates that the max depth has been exceeded.
var ErrDepthLimitExceeded = fmt.Errorf("depth limit exceeded")

const (
	// cli options
	quietOptionName       = "quiet"
	progressOptionName    = "progress"
	kindOptionName        = "kind"
	nameOptionName		  = "name"
	metaOptionName		  = "meta"
	dataOptionName		  = "data"
	lengthOptionName	  = "length"
	offsetOptionName	  = "offset"
	// internal properties
	// - hack to pass parameters from the command Run to PostRun function
	infoClassName		  = "info-class"
	createdAtName		  = "created-at"
)

const adderOutChanSize = 8

var MakeIndexCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Make index repository set.",
		ShortDescription: `
Make a new searchable repository set.
`,
		LongDescription: `
Make a new searchable infostore or metastore repository set for
documents of a similar kind. The repository kind is named using
a locally unique key ex: blog.

Each created repository set can be customized with specific schema
fields to expose structure of documents it will host. The exposed
structure can be used to refine search with the robust supported
query language.

The set of fields used for a specifc kind key can be customized
using the repository configure command.

	dms3fs index config show    # to show index configuration
	dms3fs index config --json Metadata.Kind \
		'[{"Name": "blog", "Field": ["About", "Address", \
		"Affiliation", "Author", "Brand", "Citation", \
		"Description", "Email", "Headline", "Keywords", "Language", \
		"Name", "Telephone", "Version"]}]' # to set blog fields
	dms3fs index config --json Metadata.Kind [{}] # to reset fields

Use the create document command to make an empty document template
with all the fields pre-generated.

	dms3fs index mkdoc -k=blog > b.xml    # edit document
	dms3fs index addoc b.xml <path>       # add blog to reposet

The first form of the command (without path argument) is used to
create an infostore repository set. The second form of the
command that includes a path argument is used to create an metastore
repository set. An metastore repository stores metadata information
for documents contained in an associated infostore repository set
specified by the path.

`,
	},

	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("infostores", false, true, "dms3fs path to associated repository.").EnableStdin(),
	},
	Options: []cmdkit.Option{
		cmdkit.StringOption(kindOptionName, "k", "keyword for kind of content, ex: \"blog\" ."),
		cmdkit.StringOption(nameOptionName, "n", "reposet name, ex: \"foodblog\" ."),
		cmdkit.BoolOption(quietOptionName, "q", "Write just hashes of created object.").WithDefault(false),
		cmdkit.BoolOption(progressOptionName, "p", "Stream progress data.").WithDefault(true),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) {

		n, err := cmdenv.GetNode(env)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		//
		// process and verify options and arguments
		//

		kopt, _ := req.Options[kindOptionName].(string)
		if kopt == "" {
			res.SetError(errors.New("kind of content key must be specified."), cmdkit.ErrNormal)
			return
		} else {
			log.Debugf("kind option value %s", kopt)
		}

		nopt, _ := req.Options[nameOptionName].(string)
		if nopt == "" {
			res.SetError(errors.New("reposet name must be specified."), cmdkit.ErrNormal)
			return
		}
		log.Debugf("reposet name option value %s", nopt)

        if len(req.Arguments) < 1 {
			req.SetOption(infoClassName, "infostore")
        } else {
			req.SetOption(infoClassName, "metastore")
			// metastore to infostore associatioins
			paths := req.Arguments
	        out := make([]*cid.Cid, len(paths))

	        r := &resolver.Resolver{
	                DAG:         n.DAG,
	                ResolveOnce: uio.ResolveUnixfsOnce,
	        }
	        for i, fpath := range paths {
				p, err := dms3fspath.ParsePath(fpath)
				if err != nil {
					res.SetError(errors.New(fmt.Sprintf("failed to parse path to infostore. error %s", err)), cmdkit.ErrNormal)
					return
				}

				dagnode, err := core.Resolve(n.Context(), n.Namesys, r, p)
				if err != nil {
					res.SetError(errors.New(fmt.Sprintf("mkidx: %s", err)), cmdkit.ErrNormal)
					return
				}
				out[i] = dagnode.Cid()
				fmt.Printf("infostore[%v] cid %s\n", i, dagnode.Cid().String())

				pn, ok := dagnode.(*dag.ProtoNode)
			    if !ok {
					res.SetError(errors.New(fmt.Sprintf("mkidx: invalid dag node %s", dagnode.Cid().String())), cmdkit.ErrNormal)
					return
			    }

				// make index repository root with this root node
				sr, err := idxufs.NewStoreRoot(n.Context(), n.DAG, pn)
			    if err != nil {
					res.SetError(errors.New(fmt.Sprintf("mkidx: failed to create NewStoreRoot with error %s", err)), cmdkit.ErrNormal)
					return
			    }

				// make new reposetprops type
				rps := idxufs.NewReposetProps()

				// create repoprops from store root
			    ri, err := sr.GetProps("reposetprops", rps)
			    if err != nil {
					res.SetError(errors.New(fmt.Sprintf("mkidx: failed to get reposetprops with error %s", err)), cmdkit.ErrNormal)
					return
			    }

			    rps, ok = ri.(idxufs.ReposetProps)
			    if !ok {
					res.SetError(errors.New(fmt.Sprintf("mkidx: invalid reposetprops.", err)), cmdkit.ErrNormal)
					return
			    }
	        }
			//res.SetError(errors.New("test done"), cmdkit.ErrNormal)
			//return
        }
		log.Debugf("infoclass is %s", req.Options[infoClassName].(string))


 		// load the index configuration
		icfg, err := n.Repo.IdxConfig()
		if err != nil {
			res.SetError(errors.New("could not load index config."), cmdkit.ErrNormal)
			return
		}

		// check repo kind is configured
		if found, err := idxlfs.IsKindConfigured(icfg, kopt); err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		} else {
			if !found {
				res.SetError(fmt.Sprintf("metadata not configured for repo kind %s\n", kopt), cmdkit.ErrNormal)
				return
			}
		}

		// check repo does not already exists on local filesystem
		var rpath string

		if found, p, err := idxlfs.ReposetExists(kopt, nopt); err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		} else {
			if found {
				res.SetError(fmt.Sprintf("named reposet path already exists %s\n", rpath), cmdkit.ErrNormal)
				return
			} else {
				rpath = p
			}
		}

		// check repo does not already exists on in kvstore
		var key ds.Key

		iopt, _ := req.Options[infoClassName].(string)

		// set the KV store to use
		idxkvs.InitIndexKVStore(n.Repo.Datastore())
		dstore := idxkvs.GetIndexKVStore()

		if key, err = idxkvs.GetRepoSetKey(iopt, kopt, nopt); err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
        } else {
			if _, err = dstore.Get(key); err == nil {
				res.SetError(fmt.Sprintf("named reposet key already exists %s\n", key), cmdkit.ErrNormal)
				return
			}
		}
		log.Debugf("reposet key is %v\n", key)

		// now we are ready to configure the reposet

		// create the params file on local filesystem
		var paramsfile, reponame string
		if fn, rn, ct, err := idxlfs.MakeRepo(icfg, rpath, kopt); err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		} else {
			paramsfile = fn
			reponame = rn
			req.SetOption(createdAtName, fmt.Sprintf("%s", ct.UTC().Format(time.RFC3339)))
			log.Debugf("reposet create time %s", req.Options[createdAtName].(string))
		}

		// add params file into Dms3Fs
		if err := addParamsFile(req, res, env, n, paramsfile, reponame); err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

	},
	PostRun: cmds.PostRunMap{
		cmds.CLI: func(req *cmds.Request, re cmds.ResponseEmitter) cmds.ResponseEmitter {
			reNext, res := cmds.NewChanResponsePair(req)
			outChan := make(chan interface{})

			sizeChan := make(chan int64, 1)

			sizeFile, ok := req.Files.(files.SizeFile)
			if ok {
				// Could be slow.
				go func() {
					size, err := sizeFile.Size()
					if err != nil {
						log.Warningf("error getting files size: %s", err)
						// see comment above
						return
					}

					sizeChan <- size
				}()
			} else {
				// we don't need to error, the progress bar just
				// won't know how big the files are
				log.Warning("cannot determine size of input file")
			}

			progressBar := func(wait chan struct{}) {
				defer close(wait)

				quiet, _ := req.Options[quietOptionName].(bool)
				quieter := false

				progress, _ := req.Options[progressOptionName].(bool)

				var bar *pb.ProgressBar
				if progress {
					bar = pb.New64(0).SetUnits(pb.U_BYTES)
					bar.ManualUpdate = true
					bar.ShowTimeLeft = false
					bar.ShowPercent = false
					bar.Output = os.Stderr
					bar.Start()
				}

				lastFile := ""
				lastHash := ""
				var totalProgress, prevFiles, lastBytes int64

			LOOP:
				for {
					select {
					case out, ok := <-outChan:
						if !ok {
							if quieter {
								fmt.Fprintln(os.Stdout, lastHash)
							}

							break LOOP
						}
						output := out.(*coreunix.AddedObject)
						if len(output.Hash) > 0 {
							lastHash = output.Hash
							if quieter {
								continue
							}

							if progress {
								// clear progress bar line before we print "added x" output
								fmt.Fprintf(os.Stderr, "\033[2K\r")
							}
							if quiet {
								fmt.Fprintf(os.Stdout, "%s\n", output.Hash)
							} else {
								fmt.Fprintf(os.Stdout, "added %s %s\n", output.Hash, output.Name)
							}

						} else {
							if !progress {
								continue
							}

							if len(lastFile) == 0 {
								lastFile = output.Name
							}
							if output.Name != lastFile || output.Bytes < lastBytes {
								prevFiles += lastBytes
								lastFile = output.Name
							}
							lastBytes = output.Bytes
							delta := prevFiles + lastBytes - totalProgress
							totalProgress = bar.Add64(delta)
						}

						if progress {
							bar.Update()
						}
					case size := <-sizeChan:
						if progress {
							bar.Total = size
							bar.ShowPercent = true
							bar.ShowBar = true
							bar.ShowTimeLeft = true
						}
					case <-req.Context.Done():
						// don't set or print error here, that happens in the goroutine below
						return
					}
				}
			}

			go func() {
				// defer order important! First close outChan, then wait for output to finish, then close re
				defer re.Close()

				if e := res.Error(); e != nil {
					defer close(outChan)
					re.SetError(e.Message, e.Code)
					return
				}

				wait := make(chan struct{})
				go progressBar(wait)

				defer func() { <-wait }()
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
	Type: coreunix.AddedObject{},
}


func addParamsFile(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment, n *core.Dms3FsNode, fpath, reponame string) error {

	// following logic is lifted from core/commands/add

	cfg, err := n.Repo.Config()
	if err != nil {
		return err
	}
	// check if repo will exceed storage limit if added
	// TODO: this doesn't handle the case if the hashed file is already in blocks (deduplicated)
	// TODO: conditional GC is disabled due to it is somehow not possible to pass the size to the daemon
	//if err := corerepo.ConditionalGC(req.Context(), n, uint64(size)); err != nil {
	//	res.SetError(err, cmdkit.ErrNormal)
	//	return
	//}

	progress, _ := req.Options[progressOptionName].(bool)
	trickle := false
	wrap := false
	hash := false
	hidden := false
	silent := false
	chunker := "size-262144"
	dopin := true
	rawblks, rbset := false, false
	nocopy := false
	fscache := false
	cidVer, cidVerSet := 1, true
	hashFunStr := "sha2-256"
	inline := false
	inlineLimit := 32

	// The arguments are subject to the following constraints.
	//
	// nocopy -> filestoreEnabled
	// nocopy -> rawblocks
	// (hash != sha2-256) -> cidv1

	// NOTE: 'rawblocks -> cidv1' is missing. Legacy reasons.

	// nocopy -> filestoreEnabled
	if nocopy && !cfg.Experimental.FilestoreEnabled {
		return filestore.ErrFilestoreNotEnabled
	}

	// nocopy -> rawblocks
	if nocopy && !rawblks {
		// fixed?
		if rbset {
			return fmt.Errorf("nocopy option requires '--raw-leaves' to be enabled as well")
		}
		// No, satisfy mandatory constraint.
		rawblks = true
	}

	// (hash != "sha2-256") -> CIDv1
	if hashFunStr != "sha2-256" && cidVer == 0 {
		if cidVerSet {
			return errors.New("CIDv0 only supports sha2-256")
		}
		cidVer = 1
	}

	// cidV1 -> raw blocks (by default)
	if cidVer > 0 && !rbset {
		rawblks = true
	}

	prefix, err := dag.PrefixForCidVersion(cidVer)
	if err != nil {
		return err
	}

	hashFunCode, ok := mh.Names[strings.ToLower(hashFunStr)]
	if !ok {
		return fmt.Errorf("unrecognized hash function: %s", strings.ToLower(hashFunStr))
	}

	prefix.MhType = hashFunCode
	prefix.MhLength = -1

	if hash {
		nilnode, err := core.NewNode(n.Context(), &core.BuildCfg{
			//TODO: need this to be true or all files
			// hashed will be stored in memory!
			NilRepo: true,
		})
		if err != nil {
			return err
		}
		n = nilnode
	}

	addblockstore := n.Blockstore
	if !(fscache || nocopy) {
		addblockstore = bstore.NewGCBlockstore(n.BaseBlocks, n.GCLocker)
	}

	exch := n.Exchange
	local, _ := req.Options["local"].(bool)
	if local {
		exch = offline.Exchange(addblockstore)
	}

	bserv := blockservice.New(addblockstore, exch) // hash security 001
	dserv := dag.NewDAGService(bserv)

	outChan := make(chan interface{}, adderOutChanSize)

	fileAdder, err := coreunix.NewAdder(req.Context, n.Pinning, n.Blockstore, dserv)
	if err != nil {
		return err
	}

	fileAdder.Out = outChan
	fileAdder.Chunker = chunker
	fileAdder.Progress = progress
	fileAdder.Hidden = hidden
	fileAdder.Trickle = trickle
	fileAdder.Wrap = wrap
	fileAdder.Pin = dopin
	fileAdder.Silent = silent
	fileAdder.RawLeaves = rawblks
	fileAdder.NoCopy = nocopy
	fileAdder.CidBuilder = prefix

	if inline {
		fileAdder.CidBuilder = cidutil.InlineBuilder{
			Builder: fileAdder.CidBuilder,
			Limit:   inlineLimit,
		}
	}

	if hash {
		md := dagtest.Mock()
		emptyDirNode := ft.EmptyDirNode()
		// Use the same prefix for the "empty" MFS root as for the file adder.
		emptyDirNode.SetCidBuilder(fileAdder.CidBuilder)
		mr, err := mfs.NewRoot(req.Context, md, emptyDirNode, nil)
		if err != nil {
			return err
		}

		fileAdder.SetMfsRoot(mr)
	}

	addAndPin := func(f files.File) error {

		file := f

		if err := fileAdder.AddFile(file); err != nil {
			return err
		}

		// copy intermediary nodes from editor to our actual dagservice
		ldnode, err := fileAdder.Finalize()
		if err != nil {
			return err
		}

		err = fileAdder.PinRoot()
		if err != nil {
			return err
		}

		api, err := cmdenv.GetApi(env)
		if err != nil {
			return err
		}

		// create new reposet DAG tree
		err = createRepoNode(req, res, env, n, api, file.FullPath(), reponame, ldnode, fileAdder.Out)
		if err != nil {
			return err
		}

		return err
	}

	errCh := make(chan error)
	go func() {
		var err error
		defer func() { errCh <- err }()
		defer close(outChan)

        fpath = filepath.ToSlash(filepath.Clean(fpath))

        stat, err := os.Lstat(fpath)
        if err != nil {
                return
        }

        if stat.IsDir() {
			err = fmt.Errorf("Invalid params file path '%s', path must not be a directory", fpath)
            return
        }

        f, err := files.NewSerialFile(path.Base(fpath), fpath, false, stat)
		if err != nil {
                return
        }

		err = addAndPin(f)

		// TODO: instantiate and serve the index repository

	}()

	defer res.Close()

	err = res.Emit(outChan)
	if err != nil {
		log.Error(err)
		return err
	}
	err = <-errCh
	if err != nil {
		return err
	}
	return err
}

func createRepoNode(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment, n *core.Dms3FsNode, api coreiface.CoreAPI, paramsfile, reponame string, paramscid dms3ld.Node, outchan chan interface{}) error {

	ctx := req.Context

	defer n.Blockstore.PinLock().Unlock()

	ct, err := time.Parse(time.RFC3339, req.Options[createdAtName].(string))
	if err != nil {
        return err
    }
	createdAt := uint64(ct.Unix())

	rp := idxufs.NewRepoProps()

	_, pfname := path.Split(paramsfile)	// pfname is params file name

    rp.SetType(req.Options[infoClassName].(string))
    rp.SetKind(req.Options[kindOptionName].(string))
    rp.SetName(reponame)	// includes area/cat/offset matching values below
    rp.SetOffset(0)			// offset is zero at creation time
    rp.SetArea(1)			// 0 implies N/A
    rp.SetCat(1)			// 0 implies N/A
    rp.SetPath(paramsfile)

	sr, err := idxufs.NewStoreRoot(ctx, n.DAG, nil)
    if err != nil {
		return err
    }

    rpid, err := sr.AddProps(reponame, rp)
    if err != nil {
		return err
    }

	outchan <- &coreunix.AddedObject{
		Hash: rpid.String(),
    	Name: reponame,
    	Size: strconv.FormatUint(0, 10),
	}

	rps := idxufs.NewReposetProps()

    rps.SetType(req.Options[infoClassName].(string))
    rps.SetKind(req.Options[kindOptionName].(string))
    rps.SetName(req.Options[nameOptionName].(string))
    rps.SetCreatedAt(createdAt)
    rps.SetMaxAreas(64)			// TODO: sould be configurable, per indexer or kind
    rps.SetMaxCats(64)			// TODO: sould be configurable, per indexer or kind
    rps.SetMaxDocs(50000000)	// TODO: sould be configurable, per indexer or kind

	reposetName := "reposetprops"
    rpsid, err := sr.AddProps(reposetName, rps)
    if err != nil {
		return err
    }

	outchan <- &coreunix.AddedObject{
		Hash: rpsid.String(),
    	Name: reposetName,
    	Size: strconv.FormatUint(0, 10),
	}

	rootdir := sr.GetDirectory()

	err = rootdir.AddChild(pfname, paramscid)
	if err != nil {
		return nil
	}

	nd, err := rootdir.GetNode() // adds rootdir to dag
	if err != nil {
		return nil
	}

	// pin nodes
	n.Pinning.PinWithMode(nd.Cid(), pin.Recursive)
	err = n.Pinning.Flush()
	if err != nil {
		return err
	}

	outchan <- &coreunix.AddedObject{
		Hash: nd.Cid().String(),
    	Name: "reposetdir",
    	Size: strconv.FormatUint(0, 10),
	}

	// set the KV store to use, where we track reposet cids.
	// this enables repo lookup by type, kind, name, etc...
	idxkvs.InitIndexKVStore(n.Repo.Datastore())
	dstore := idxkvs.GetIndexKVStore()

	var key ds.Key
	var value []byte

	iopt, _ := req.Options[infoClassName].(string)
	kopt, _ := req.Options[kindOptionName].(string)
	nopt, _ := req.Options[nameOptionName].(string)

	if key, err = idxkvs.GetRepoSetKey(iopt, kopt, nopt); err != nil {
		return errors.New(fmt.Sprintf("could not lookup reposet key. error: %s\n", err))
	}

	v := idxkvs.NewRps()
	v.SetCid(nd.Cid())

	if value, err = v.Marshal(); err != nil {
		return errors.New(fmt.Sprintf("could not marshal reposet value. error: %s\n", err))
	}
	if err = dstore.Put(key, value); err != nil {
		return errors.New(fmt.Sprintf("could not put reposet key value. error: %s\n", err))
	}
	log.Debugf("reposet key %v value %v\n", key, value)

	return err

}



type RepoDoc struct{
	content string
}

var MakeDocumentCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Make document template for new content.",
		ShortDescription: `
Make an empty document for editing new content.
`,
		LongDescription: `
Make a document for editing new content of a similar kind.
The content kind is named using a locally unique key ex: blog.

Use the create document command to create an empty document template
with all the fields pre-generated.

	dms3fs index mkdoc -k=blog > b.xml    # edit document, then
	dms3fs index addoc b.xml <path>       # add blog to reposet

`,
	},

	Arguments: []cmdkit.Argument{
	},
	Options: []cmdkit.Option{
		cmdkit.StringOption("kind", "k", "keyword for kind of content, ex: \"blog\" ."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) {
		n, err := cmdenv.GetNode(env)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		kopt, _ := req.Options[kindOptionName].(string)
		if kopt == "" {
			res.SetError(errors.New("kind of content key must be specified."), cmdkit.ErrNormal)
			return
		}
		log.Debugf("kind option value %s", kopt)

        icfg, err := n.Repo.IdxConfig()
		if err != nil {
			res.SetError(errors.New("could not load index config."), cmdkit.ErrNormal)
			return
		}

		var repodoc *RepoDoc

		output, err := idxlfs.MakeDoc(*icfg, kopt)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		} else {
			repodoc = &RepoDoc{
				content: output,
			}
		}

		cmds.EmitOnce(res, repodoc)

	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeEncoder(func(req *cmds.Request, w io.Writer, v interface{}) error {
			repoDoc, ok := v.(*RepoDoc)
			if !ok {
				return e.TypeErr(repoDoc, v)
			}

			_, err := fmt.Fprintf(w, "%v\n",repoDoc.content)
			return err
		}),
	},
	Type: RepoDoc{},
}
