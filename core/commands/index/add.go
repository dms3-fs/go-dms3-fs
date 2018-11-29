package index

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
//	"time"

	core "github.com/dms3-fs/go-dms3-fs/core"
	cmdenv "github.com/dms3-fs/go-dms3-fs/core/commands/cmdenv"
	e "github.com/dms3-fs/go-dms3-fs/core/commands/e"

	cmds "github.com/dms3-fs/go-fs-cmds"
	cmdkit "github.com/dms3-fs/go-fs-cmdkit"
	path "github.com/dms3-fs/go-path"

	idxkvs "github.com/dms3-fs/go-dms3-fs/core/coreindex/kvs"
    cid "github.com/dms3-fs/go-cid"
	ds "github.com/dms3-fs/go-datastore"
    mh "github.com/dms3-mft/go-multihash"

)

var AddDocumentCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Add document to index repository.",
		ShortDescription: `
Add a document to index repository specified by path.
`,
		LongDescription: `
Add a document to an index repository at path containing content
of a similar kind.

Use the create document command to create an empty document template
with all the fields pre-generated. After editing the document to write
the desired content, add the document to the repository.

	dms3fs index mkdoc -k="blog" --xml > b.xml # edit document
	dms3fs index addoc b.xml <path>          # add blog to reposet

Use --xml option to convey repository input document format.

`,
	},

	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("file", true, false, "content to add to repository."),
		cmdkit.StringArg("dms3fs-path", true, false, "path to repository."),
	},
	Options: []cmdkit.Option{
		cmdkit.BoolOption("quiet", "q", "Write just hashes of created object."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) {
		var file, repo string
        if len(req.Arguments) != 2 {
			res.SetError(errors.New("file and path are both required."), cmdkit.ErrNormal)
			return
        } else {
			file = req.Arguments[0]
			repo = req.Arguments[1]
        }

		log.Debugf("file value is %s, repo path is %s", file, repo)

		n, err := cmdenv.GetNode(env)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		//log.Error("Running index ls : ", err)
		log.Debugf("Running command request path %s", req.Path)

		adopts := new(addocOpts)

		q, _ := req.Options["quiet"].(bool)
		if q {
			adopts.q = true
		} else {
			adopts.q = false
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

		output, err := addDoc(ctx, n, p, adopts)
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

type RepoPath struct{
       path string
}

type addocOpts struct {
	q bool
}

func addDoc(ctx context.Context, n *core.Dms3FsNode, ref path.Path, opts *addocOpts) (*RepoPath, error) {

	// set the KV store to use
	idxkvs.InitIndexKVStore(n.Repo.Datastore())
	dstore := idxkvs.GetIndexKVStore()

	rc := "infostore"
	rk := "blog"
	ri := int64(0)
	c1 := idxkvs.NewCorpusProps(rc, rk, ri, &cid.Cid{})
	c2 := idxkvs.NewCorpusProps("", "", 0, &cid.Cid{})

	var i int64
	var sb strings.Builder
	var key ds.Key
    var value []byte
    var err error

	for i = 0; i < 100; i++ {
		sb.Reset()
		fmt.Fprintf(&sb, "idhash%d", i)
		if id, _ := cid.NewPrefixV1(cid.Raw, mh.ID).Sum([]byte(sb.String())); id == nil {
			return nil, err
		} else {
			c1.SetRcid(id)
		}

        if value, err = c1.Marshal(); err != nil {
			return nil, err
        }

        if key, err = idxkvs.GetDocKey(c1.GetRclass(), c1.GetRkind(), c1.GetRindex(), i); err != nil {
			return nil, err
        }

        if err = dstore.Put(key, value); err != nil {
			return nil, err
        }
	}

	for i = 0; i < 100; i++ {
        sb.Reset()
		fmt.Fprintf(&sb, "idhash%d", i)
		if id, _ := cid.NewPrefixV1(cid.Raw, mh.ID).Sum([]byte(sb.String())); id == nil {
            return nil, fmt.Errorf("cannot compute cid corpus property: %v", err)
        } else {
			c1.SetRcid(id)
        }

        if key, err = idxkvs.GetDocKey(c1.GetRclass(), c1.GetRkind(), c1.GetRindex(), i); err != nil {
            return nil, fmt.Errorf("cannot get key for corpus properties: %v", err)
        }

        if value, err = dstore.Get(key); err != nil {
            return nil, fmt.Errorf("cannot get corpus properties: %v", err)
		}

        if err = c2.Unmarshal(value); err != nil {
            return nil, fmt.Errorf("cannot unmarshal corpus properties: %v", err)
        }

        if !c2.Equals(c1) {
			return nil, fmt.Errorf("put/get value mistmatch corpus property")
        }
	}

	for i = 0; i < 100; i++ {
		if key, err = idxkvs.GetDocKey(c1.GetRclass(), c1.GetRkind(), c1.GetRindex(), i); err != nil {
            return nil, fmt.Errorf("cannot get key for corpus properties: %v", err)
        }

		if err = dstore.Delete(key); err != nil {
            return nil, fmt.Errorf("cannot delete corpus properties: %v", err)
		}
	}

	return &RepoPath{
		path: "addoc test worked.",
	}, nil

}
