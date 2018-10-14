
package index
import (
	"strings"
	"context"
	"errors"
	"fmt"
	"io"
	"time"
	"os"
	"encoding/xml"
//	"reflect"
	"bytes"
	"path/filepath"

    util "github.com/dms3-fs/go-fs-util"
	"github.com/facebookgo/atomicfile"

	core "github.com/dms3-fs/go-dms3-fs/core"
	cmdenv "github.com/dms3-fs/go-dms3-fs/core/commands/cmdenv"
	e "github.com/dms3-fs/go-dms3-fs/core/commands/e"

	cmds "github.com/dms3-fs/go-fs-cmds"
	cmdkit "github.com/dms3-fs/go-fs-cmdkit"
    idxconfig "github.com/dms3-fs/go-idx-config"
	path "github.com/dms3-fs/go-path"
)

type RepoPath struct{
	path string
}

var MakeIndexCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Make index repository set.",
		ShortDescription: `
Make a new searchable repository set.
`,
		LongDescription: `
Make a new searchable infostore or infospace repository set for
documents of a similar kind. The repository kind is named using
a locally unique key ex: blog.

Each created repository set can be customized with specific schema
fields to expose structure of documents it will host. The exposed
structure can be used to refine search with the supported robust
query language.

The set of fields used for a specifc kind key can be customized
using the repository configure commands.

	dms3fs index config show    # to show index configuration
	dms3fs index config --json Metadata.Kind \
		'[{"Name": "blog", "Field": ["About", "Address", \
		"Affiliation", "Author", "Brand", "Citation", \
		"Description", "Email", "Headline", "Keywords", "Language", \
		"Name", "Telephone", "Version"]}]' # to set blog fields
	dms3fs index config --json Metadata.Kind [{}] # to reset fields

Use the create document command to make an empty document template
with all the fields pre-generated.

	dms3fs index mkdoc -k=blog -x > b.xml # edit document
	dms3fs index addoc b.xml <path>       # add blog to reposet

Use --xml option to convey repository input document format.

The first form of the command (without path argument) is used to
create and infostore repository set. The second form of the
command that includes a path argument is used to create an infospace
repository set. An infospace repository stores metadata information
for documents contained in an associated infostore repository set
specified by the path.

`,
	},

	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("dms3fs-path", false, false, "path to associated repository."),
	},
	Options: []cmdkit.Option{
		cmdkit.StringOption("kind", "k", "keyword for kind of content, ex: \"blog\" ."),
		cmdkit.BoolOption("xml", "x", "xml encoding format."),
		cmdkit.BoolOption("quiet", "q", "Write just hashes of created object."),
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

		//log.Error("Running index ls : ", err)
		log.Debugf("Running command request path %s", req.Path)

		mkopts := new(makeindexOpts)

		mkopts.key, _ = req.Options["kind"].(string)
		if mkopts.key == "" {
			res.SetError(errors.New("kind of content key must be specified."), cmdkit.ErrNormal)
			return
		}
		log.Debugf("kind option value %s", mkopts.key)

		x, _ := req.Options["xml"].(bool)
		if !x {
			res.SetError(errors.New("encoding format must be specified."), cmdkit.ErrNormal)
			return
		} else {
			mkopts.enc = "xml"
		}
		log.Debugf("xml format option %s", mkopts.enc)

		ctx := req.Context

/*
		p, err := path.ParsePath(p)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}
*/
		var p path.Path

		//
		// check repo does not exists
		//	- in kvstore and local filesystem
 		// make the parameters file - on local filesystem
		// add it into Dms3Fs
		// create LD object
		// track in the kv store
		//	- link infospace to infostore
 		//

/*
		1. DMFS3 UnixFS File - Parameters
		    A parameters file is generated when creating a new repository
			using index configuration data. The parameters file is stored
			in the DMS3FS. The parameters file is shared by all repos in a
			reposet.

			- create on local fs, the add like:
			dms3fs add /home/tavit/.dms3-fs/index/reposet/blog/w1538751225-a1-c1-o0/params
			added QmWXToC1zwjJTqCS1A8g7VgGeTJDY2PaANeZwMyH8awZfF params
			 3.05 KiB / 3.05 KiB [==================================] 100.00%

		2. DAG Object - Link to parameters
		    A DAG object stores a link to the parameters file added to
			the DMS3FS. This helps locate repository metadata when
			expanding the reposet, and when publishing information.

			works like:
			dms3fs object put /home/tavit/Work/dapp/dms3/repo.json
			added QmNR4hpeHBi61fNw2xGX56cewDS9AzRdTF3fLNVNdPTUop
			    ```go
			{
			    "Data": "repar",
			    "Links": [ {
			        "Name": "params",
			        "Hash": "QmWXToC1zwjJTqCS1A8g7VgGeTJDY2PaANeZwMyH8awZfF",
			        "Size": 3139
			    } ]
			}
			```

		3. Index Datastore Entry - Reposet properties
		    The set of properties of a reposet is stored under a unique Key
			in the DMS3FS index datastore. The properties are JSON encoded and
			stored using the key name convention. Key: "_class_/_kind_", where
			_class_ is either "infostore" of "infospace", and _kind_ is a
			source provided key component that describes the kind of repository.

			    ```go
			type reposetProps struct {
			    Kind: string,       // reposet kind
			    CreatedAt: uint64,  // creation (Unix) time (seconds since 1970 epoch)
			    MaxAreas: uint8,    // max tag2 shards, default: 64
			    MaxCats: uint8,     // max tag3 shards, default: 64
			    MaxDocs: uint64,    // max # of documents in repo kind, DEFAULT: 50m
			    Params: *Cid,       // cid of reposet paramaters file
			    RepoKey: []string,  // key list of repos in reposet
			}
			```

		4. Index Datastore Entry - Repo properties
		    The set of properties of a repo is stored under a unique Key in
			the DMS3FS index datastore. The properties are JSON encoded and
			stored using the key name convention. Key: "_class_/_kind_/_n_",
			where _n_ is the key list index of the repo in the reposet.

			works like:
			dms3fs object put /home/tavit/Work/dapp/dms3/reposet.json
			added QmNPtzjKHpadJvzxEhEB5RjQwuPBGhXzvQ8aXYA1uLGpP7

			    ```go
			type repoProps struct {
			    Suffix: {           // repository folder name suffix
			        Offset: 0,      // shard tag1, repo create relative time,
			                        // seconds since reposet creation
			        Area: 0,        // shard tag2
			        Category: 0,    // shard tag3
			    },
			    Docs: 0,            // # of documents in the repo
			}
			```

		5. Local filesystem - Path to repository
		    A path is created when creating a new repository that contains
			files and sub folders used by the indexer. A copy of the
			parameters file is placed in the local file system for the
			index server to read its configuration from. The path of a
			repository is composed of the following components:
			    _root_ - index repository root
			    _class_ - repository class
			    _kind_ - repository kind
			    _shard_ - repository name

		    For example, the very first "blog" kind repository created
			at Unix time of 1538751225 (seconds since 1970 epoch) will
			create the following folder structure by default:

			    ```bash
			~/.dms3-fs/index/infostore/blog/w1538751225-o0-a1-c1/params
			~/.dms3-fs/index/infostore/blog/w1538751225-o0-a1-c1/index/
			~/.dms3-fs/index/infostore/blog/w1538751225-o0-a1-c1/corpus/
			~/.dms3-fs/index/infostore/blog/w1538751225-o0-a1-c1/metadata/
			}
			```
*/

		cfg, err := n.Repo.IdxConfig()
		if err != nil {
			res.SetError(errors.New("could not load index config."), cmdkit.ErrNormal)
			return
		}

		output, err := makeRepo(ctx, n, cfg, p, mkopts)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		cmds.EmitOnce(res, output)

	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeEncoder(func(req *cmds.Request, w io.Writer, v interface{}) error {
			repoPath, ok := v.(*RepoPath)
			if !ok {
				return e.TypeErr(repoPath, v)
			}

			_, err := fmt.Fprintf(w, "%v\n", repoPath.path)
			return err
		}),
	},
	Type: RepoPath{},
}

type makeindexOpts struct {
	key string
	enc string
}

func makeRepo(ctx context.Context, n *core.Dms3FsNode, cfg interface{}, ref path.Path, opts *makeindexOpts) (*RepoPath, error) {

	var err error
	var found bool
	var kind, reposet, filename string

	if kind = opts.key; kind == "" {
		err = errors.New("repo kind must not be null.")
		return nil, err
	}

	if reposet, err = reposetName(kind); err != nil {
		return nil, err
	}

	if reposetExists(reposet) {
		err = errors.New(fmt.Sprintf("repo kind \"%s\" already exists.", kind))
		return nil, err
	}

	if filename, err = repoParamFilename(reposet); err != nil {
		return nil, err
	}

	// make path to repo root and indexer params
	found, err = writeParamFile(filename, cfg, opts)

	// make subfolders after creating params file, which also create path to repo root
	if err = makeRepoSubDirs(filename); err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	} else {
		if found {
			return &RepoPath{
				path: "success",
				}, nil
		} else {
			return &RepoPath{
				path: "failed to make index repository, please use \"dms3fs index config\" command to verify configure.",
			}, nil
		}
	}
}

func reposetExists(reposet string) bool {
	//
	// TODO: check both kvstore and params file in local fs
	//
	if !util.FileExists(reposet) {
			return false
	}
	return true
}

// repoParamFilename returns the index repository parameters filename for a specified reposet.
func reposetName(kind string) (reposet string, err error) {

	//
	// local filesystem reposet path.
	//
	// index <repo root>, cfg parameter Indexer.Path, must be relative path
	// 	  - index
	// reposet root
	// 	  - index/reposet
	// reposet name
	//	  - <kind>: string,    // repository kind (ex: blog)
	// reposet root folder
	// 	  - index/reposet/<kind>
	//
	var rootpath string

	if rootpath, err = idxconfig.PathRoot(); err != nil {
		return reposet, err
	}

	reposet = filepath.Join(rootpath, "reposet", kind)
	fmt.Printf("rootpath: %v\n", rootpath)
	fmt.Printf("kind: %v\n", kind)
	fmt.Printf("reposet: %v\n", reposet)

	return reposet, err
}

// repoParamFilename returns the index repository parameters filename for a specified reposet.
func repoParamFilename(reposet string) (filename string, err error) {

	//
	// local filesystem repository file folder hierarchy.
	//
	// index <repo root>, cfg parameter Indexer.Path, must be relative path
	// 	  - index
	// reposet root
	// 	  - index/reposet
	// reposet name
	//	  - <kind>: string,    // repository kind (ex: blog)
	// reposet root folder
	// 	  - index/reposet/<kind>
	// repo name, composed as:
	//	  - window: uint64,    // creation time (Unix, seconds), sharding tag
	//	  - area: uint8,       // area number, sharding tag
	//	  - cat: uint8,        // category number, sharding tag
	//	  - offset: int64,     // time since creation (seconds), recovery tag
	// repo root folder
	// 	  - index/reposet/<kind>/<reponame>
	// repo specific files and sub-folders, cfg parameters are ignored for these
	//	  - index/reposet/<kind>/<reponame>/corpus, cfg parameter Indexer.Corpus.Path
	//	  - index/reposet/<kind>/<reponame>/metadata, cfg parameter Indexer.Corpus.Metadata
	// 	  - params, repo params file, no corresponding cfg parameter
	//
	t := time.Now()	// repo create time
	o := t.Sub(t)	// seconds since repo create (zero at creation time)
	a := 0			// current area (zero ==> none)
	c := 0			// current category (zero ==> none)

	window := fmt.Sprintf("w%d", t.Unix())
	area := fmt.Sprintf("-a%d", a+1)
	category := fmt.Sprintf("-c%d", c+1)
	offset := fmt.Sprintf("-o%d", o)

	reponame := window + area + category + offset

	filename = filepath.Join(reposet, reponame, "params")
	fmt.Printf("reposet: %v\n", reposet)
	fmt.Printf("reponame: %v\n", reponame)
	fmt.Printf("filename: %v\n", filename)

	return filename, err
}

// writeParamFile writes the index repository parameters from `cfg` into `filename`.
func writeParamFile(filename string, cfg interface{}, opts *makeindexOpts) (bool, error) {
        err := os.MkdirAll(filepath.Dir(filename), 0775)
        if err != nil {
                return false, err
        }

        f, err := atomicfile.New(filename, 0660)
        if err != nil {
                return false, err
        }
        defer f.Close()

        return encode(f, cfg, opts)
}

// given repo params file, create repo subfolders
// cfg path parameters are ignored for these
//	  - Indexer.Corpus.Path
//	  - Indexer.Corpus.Metadata
func makeRepoSubDirs(filename string) error {

	var i, c, m, r string

	r = filepath.Dir(filename)
	i = filepath.Join(r, "index")
	c = filepath.Join(r, "corpus")
	m = filepath.Join(r, "metadata")

	fmt.Printf("index: %v\n", i)
	fmt.Printf("corpus: %v\n", c)
	fmt.Printf("metadata: %v\n", m)

	if err := os.Mkdir(i, 0660); err != nil {
		return err
	}
	if err := os.Mkdir(c, 0660); err != nil {
		return err
	}
	if err := os.Mkdir(m, 0660); err != nil {
		return err
	}
	return nil
}

func encode(w io.Writer, value interface{}, opts *makeindexOpts) (found bool, err error) {

	// encode the index parameters file from configured properties
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")

	var sel xml.StartElement
	var snm xml.Name
	snm.Space = ""
	snm.Local = "parameters"
	sel.Name = snm
	if err := enc.EncodeToken(sel); err != nil {
		fmt.Printf("error: %v\n", err)
		found = false
		return found, err
	}

	if found, err = makeIndexPathCorpusParams(value, enc); err != nil {
		fmt.Printf("error: %v\n", err)
		found = false
		return found, err
	}

	if found, err = makeIndexMetadataParams(value, enc); err != nil {
		fmt.Printf("error: %v\n", err)
		found = false
		return found, err
	}

	if found, err = makeIndexFieldsParams(value, enc, opts); err != nil {
		fmt.Printf("error: %v\n", err)
		found = false
		return found, err
	}

	if found, err = makeIndexMemStemNormStopParams(value, enc); err != nil {
		fmt.Printf("error: %v\n", err)
		found = false
		return found, err
	}

	if err := enc.EncodeToken(sel.End()); err != nil {
		fmt.Printf("error: %v\n", err)
		found = false
		return found, err
	}

	if err := enc.Flush();  err != nil {
		fmt.Printf("error: %v\n", err)
		found = false
		return found, err
	}
	return found, err
}

func makeIndexPathCorpusParams(value interface{}, enc *xml.Encoder) (found bool, err error) {

	if iconf, ok := value.(*idxconfig.IdxConfig); ok {
		ix := iconf.Indexer

		if ix.Path != "" {
			found = true

			var el xml.StartElement
			var en xml.Name
			en.Space = ""
			en.Local = "index"
			el.Name = en
			if err := enc.EncodeElement(strings.ToLower(ix.Path),el); err != nil {
				fmt.Printf("error: %v\n", err)
				found = false
				return found, err
			}
		} else {
			err = errors.New("missing index path parameters, please use \"dms3fs index config\" command to verify configure.")
			found = false
			return found, err
		}

		if ix.Corpus != (idxconfig.Corpus{}) {
			found = true

			var cel xml.StartElement
			var cen xml.Name
			cen.Space = ""
			cen.Local = "corpus"
			cel.Name = cen
			if err := enc.EncodeToken(cel); err != nil {
				fmt.Printf("error: %v\n", err)
				found = false
				return found, err
			}

			if ix.Corpus.Path != "" {
				var cpel xml.StartElement
				var cpen xml.Name
				cpen.Space = ""
				cpen.Local = "path"
				cpel.Name = cpen
				if err := enc.EncodeElement(strings.ToLower(ix.Corpus.Path),cpel); err != nil {
					fmt.Printf("error: %v\n", err)
					found = false
					return found, err
				}
			}

			if ix.Corpus.Class != "" {
				var ccel xml.StartElement
				var ccen xml.Name
				ccen.Space = ""
				ccen.Local = "class"
				ccel.Name = ccen
				if err := enc.EncodeElement(strings.ToLower(ix.Corpus.Class),ccel); err != nil {
					fmt.Printf("error: %v\n", err)
					found = false
					return found, err
				}
			}

			if ix.Corpus.Metadata != "" {
				var cmel xml.StartElement
				var cmen xml.Name
				cmen.Space = ""
				cmen.Local = "metadata"
				cmel.Name = cmen
				if err := enc.EncodeElement(strings.ToLower(ix.Corpus.Metadata),cmel); err != nil {
					fmt.Printf("error: %v\n", err)
					found = false
					return found, err
				}
			}

			if err := enc.EncodeToken(cel.End()); err != nil {
				fmt.Printf("error: %v\n", err)
				found = false
				return found, err
			}
		} else {
			err = errors.New("missing index corpus parameters, please use \"dms3fs index config\" command to verify configure.")
			found = false
			return found, err
		}
	}

	return found, err
}

func makeIndexMemStemNormStopParams(value interface{}, enc *xml.Encoder) (found bool, err error) {

	if iconf, ok := value.(*idxconfig.IdxConfig); ok {
		var n1, n2 xml.CharData
		var c1 xml.Comment
		var n1b, n2b, c1b []byte
		n1s := bytes.NewBuffer(n1b)
		n2s := bytes.NewBuffer(n2b)
		c1s := bytes.NewBuffer(c1b)
		n1s.WriteString("\n")
		n2s.WriteString("\n\n")
		c1s.WriteString(" optional index parameters ")
		n1 = n1s.Bytes()
		n2 = n2s.Bytes()
		c1 = c1s.Bytes()

		if err := enc.EncodeToken(n2); err != nil {
			fmt.Printf("error: %v\n", err)
			found = false
			return found, err
		}
		if err := enc.EncodeToken(c1); err != nil {
			fmt.Printf("error: %v\n", err)
			found = false
			return found, err
		}
		if err := enc.EncodeToken(n1); err != nil {
			fmt.Printf("error: %v\n", err)
			found = false
			return found, err
		}

		ix := iconf.Indexer

		if ix.Memory != "" {
			var imel xml.StartElement
			var imen xml.Name
			imen.Space = ""
			imen.Local = "memory"
			imel.Name = imen
			if err := enc.EncodeElement(strings.ToLower(ix.Memory),imel); err != nil {
				fmt.Printf("error: %v\n", err)
				found = false
				return found, err
			}
		}

		if ix.Stemmer != "" {
			var isel xml.StartElement
			var isen xml.Name
			isen.Space = ""
			isen.Local = "stemmer"
			isel.Name = isen
			if err := enc.EncodeToken(isel); err != nil {
				fmt.Printf("error: %v\n", err)
				found = false
				return found, err
			}

			var isnel xml.StartElement
			var isnen xml.Name
			isnen.Space = ""
			isnen.Local = "name"
			isnel.Name = isnen
			if err := enc.EncodeElement(strings.ToLower(ix.Stemmer),isnel); err != nil {
				fmt.Printf("error: %v\n", err)
				found = false
				return found, err
			}

			if err := enc.EncodeToken(isel.End()); err != nil {
				fmt.Printf("error: %v\n", err)
				found = false
				return found, err
			}
		}

		var inel xml.StartElement
		var inen xml.Name
		inen.Space = ""
		inen.Local = "normalize"
		inel.Name = inen
		if err := enc.EncodeElement(ix.Normalize,inel); err != nil {
			fmt.Printf("error: %v\n", err)
			found = false
			return found, err
		}

		if len(ix.Stopper) > 0 {
			found = true

			var istel xml.StartElement
			var isten xml.Name
			isten.Space = ""
			isten.Local = "stopper"
			istel.Name = isten
			if err := enc.EncodeToken(istel); err != nil {
				fmt.Printf("error: %v\n", err)
				found = false
				return found, err
			}

			for swd := range ix.Stopper {
				var swel xml.StartElement
				var swen xml.Name
				swen.Space = ""
				swen.Local = "word"
				swel.Name = swen
				if err := enc.EncodeElement(strings.ToLower(ix.Stopper[swd]),swel); err != nil {
					fmt.Printf("error: %v\n", err)
					found = false
					return found, err
				}
			}

			if err := enc.EncodeToken(istel.End()); err != nil {
				fmt.Printf("error: %v\n", err)
				found = false
				return found, err
			}
		}
	}

	return found, err
}

func makeIndexMetadataParams(value interface{}, enc *xml.Encoder) (found bool, err error) {

	var n1, n2 xml.CharData
	var c, c1, c2, c3, c4 xml.Comment
	var n1b, n2b, c1b, c2b, c3b, c4b []byte
	n1s := bytes.NewBuffer(n1b)
	n2s := bytes.NewBuffer(n2b)
	c1s := bytes.NewBuffer(c1b)
	c2s := bytes.NewBuffer(c2b)
	c3s := bytes.NewBuffer(c3b)
	c4s := bytes.NewBuffer(c4b)
	n1s.WriteString("\n")
	n2s.WriteString("\n\n")
	c1s.WriteString(" read-only life-cycle [system] metadata fields ")
	c2s.WriteString(" read-write life-cycle [document] metadata fields ")
	c3s.WriteString(" start of [document] kind common metadata fields ")
	c4s.WriteString(" start of [document] kind specific metadata fields ")
	n1 = n1s.Bytes()
	n2 = n2s.Bytes()
	c1 = c1s.Bytes()
	c2 = c2s.Bytes()
	c3 = c3s.Bytes()
	c4 = c4s.Bytes()

	rs := []string{
		// read-only [system] metadata
		"odmver",
		"schver",
		"kind",
		"basetime",
		"maxareas",
		"maxcats",
		"offset",
		"app",
		// read-write [application] metadata fields in the index paramaters file
		"docno",
		"docver",
		// start of template specific fields - generated elsewhere
	}

	found = true

	var imel xml.StartElement
	var imnm xml.Name
	imnm.Space = ""
	imnm.Local = "metadata"
	imel.Name = imnm
	if err := enc.EncodeToken(imel); err != nil {
		fmt.Printf("error: %v\n", err)
		found = false
		return found, err
	}

	for i := range rs {
		switch strings.ToLower(rs[i]) {
		case "odmver", "docno":
			if err := enc.EncodeToken(n2); err != nil {
				fmt.Printf("error: %v\n", err)
				found = false
				return found, err
			}
			if strings.ToLower(rs[i]) == "odmver" {
				c = c1
			} else {
				c = c2
			}
			if err := enc.EncodeToken(c); err != nil {
				fmt.Printf("error: %v\n", err)
				found = false
				return found, err
			}
			if err := enc.EncodeToken(n1); err != nil {
				fmt.Printf("error: %v\n", err)
				found = false
				return found, err
			}
		}
		var fel xml.StartElement
		var fen xml.Name
		fen.Space = ""
		fen.Local = "forward"
		fel.Name = fen
		if err := enc.EncodeElement(strings.ToLower(rs[i]),fel); err != nil {
			fmt.Printf("error: %v\n", err)
			found = false
			return found, err
		}
	}
	if err := enc.EncodeToken(n2); err != nil {
		fmt.Printf("error: %v\n", err)
		found = false
		return found, err
	}
	if err := enc.EncodeToken(c3); err != nil {
		fmt.Printf("error: %v\n", err)
		found = false
		return found, err
	}
	if err := enc.EncodeToken(n1); err != nil {
		fmt.Printf("error: %v\n", err)
		found = false
		return found, err
	}

	for i := range rs {
		var bel xml.StartElement
		var ben xml.Name
		ben.Space = ""
		ben.Local = "backward"
		bel.Name = ben
		if err := enc.EncodeElement(strings.ToLower(rs[i]),bel); err != nil {
			fmt.Printf("error: %v\n", err)
			found = false
			return found, err
		}
	}

	for i := range rs {

		var fdel xml.StartElement
		var fdnm xml.Name
		fdnm.Space = ""
		fdnm.Local = "field"
		fdel.Name = fdnm
		if err := enc.EncodeToken(fdel); err != nil {
			fmt.Printf("error: %v\n", err)
			found = false
			return found, err
		}

		var nel xml.StartElement
		var nnm xml.Name
		nnm.Space = ""
		nnm.Local = "name"
		nel.Name = nnm
		if err := enc.EncodeElement(strings.ToLower(rs[i]),nel); err != nil {
			fmt.Printf("error: %v\n", err)
			found = false
			return found, err
		}

		if err := enc.EncodeToken(fdel.End()); err != nil {
			fmt.Printf("error: %v\n", err)
			found = false
			return found, err
		}
	}

	if err := enc.EncodeToken(imel.End()); err != nil {
		fmt.Printf("error: %v\n", err)
		found = false
		return found, err
	}

	if err := enc.EncodeToken(n2); err != nil {
		fmt.Printf("error: %v\n", err)
		found = false
		return found, err
	}
	if err := enc.EncodeToken(c4); err != nil {
		fmt.Printf("error: %v\n", err)
		found = false
		return found, err
	}
	if err := enc.EncodeToken(n1); err != nil {
		fmt.Printf("error: %v\n", err)
		found = false
		return found, err
	}
	return found, err
}

func makeIndexFieldsParams(value interface{}, enc *xml.Encoder, opts *makeindexOpts) (found bool, err error) {

	if iconf, ok := value.(*idxconfig.IdxConfig); ok {
		a := iconf.Metadata.Kind

		for i := range a {

			if a[i].Name == opts.key {
				found = true

				var kfel xml.StartElement
				var kfnm xml.Name
				kfnm.Space = ""
				kfnm.Local = "field"
				kfel.Name = kfnm
				if err := enc.EncodeToken(kfel); err != nil {
					fmt.Printf("error: %v\n", err)
					found = false
					return found, err
				}

				var knel xml.StartElement
				var knnm xml.Name
				knnm.Space = ""
				knnm.Local = "name"
				knel.Name = knnm
				if err := enc.EncodeElement(strings.ToLower(opts.key),knel); err != nil {
					fmt.Printf("error: %v\n", err)
					found = false
					return found, err
				}

				if err := enc.EncodeToken(kfel.End()); err != nil {
					fmt.Printf("error: %v\n", err)
					found = false
					return found, err
				}

				for f := range a[i].Field {

					var fdel xml.StartElement
					var fdnm xml.Name
					fdnm.Space = ""
					fdnm.Local = "field"
					fdel.Name = fdnm
					if err := enc.EncodeToken(fdel); err != nil {
						fmt.Printf("error: %v\n", err)
						found = false
						return found, err
					}

					var nel xml.StartElement
					var nnm xml.Name
					nnm.Space = ""
					nnm.Local = "name"
					nel.Name = nnm
					if err := enc.EncodeElement(strings.ToLower(a[i].Field[f]),nel); err != nil {
						fmt.Printf("error: %v\n", err)
						found = false
						return found, err
					}

					if err := enc.EncodeToken(fdel.End()); err != nil {
						fmt.Printf("error: %v\n", err)
						found = false
						return found, err
					}
				}
			}
		}
	}

	return found, err
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

	dms3fs index mkdoc -k=blog -x > b.xml # edit document, then
	dms3fs index addoc b.xml <path>       # add blog to reposet

Use --xml option to convey repository input document format.

`,
	},

	Arguments: []cmdkit.Argument{
	},
	Options: []cmdkit.Option{
		cmdkit.StringOption("kind", "k", "keyword for kind of content, ex: \"blog\" ."),
		cmdkit.BoolOption("xml", "x", "xml encoding format."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) {
		n, err := cmdenv.GetNode(env)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		mkopts := new(makeindexOpts)

		mkopts.key, _ = req.Options["kind"].(string)
		if mkopts.key == "" {
			res.SetError(errors.New("kind of content key must be specified."), cmdkit.ErrNormal)
			return
		}
		log.Debugf("kind option value %s", mkopts.key)

		x, _ := req.Options["xml"].(bool)
		if !x {
			res.SetError(errors.New("encoding format must be specified."), cmdkit.ErrNormal)
			return
		} else {
			mkopts.enc = "xml"
		}
		log.Debugf("xml format option %s", mkopts.enc)

		ctx := req.Context

        cfg, err := n.Repo.IdxConfig()
		if err != nil {
			res.SetError(errors.New("could not load index config."), cmdkit.ErrNormal)
			return
		}

		output, err := makeDoc(ctx, n, *cfg, mkopts)
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
			return
		}

		cmds.EmitOnce(res, output)

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

type makedocOpts struct {
	key string
	enc string
}

func makeDoc(ctx context.Context, n *core.Dms3FsNode, iconf idxconfig.IdxConfig, opts *makeindexOpts) (*RepoDoc, error) {

	var found bool = false

	var b []byte
	w := bytes.NewBuffer(b)

	a := iconf.Metadata.Kind
	for i := range a {

		if a[i].Name == opts.key {
			found = true

			enc := xml.NewEncoder(w)
			enc.Indent("  ", "    ")
			var sel xml.StartElement
			var snm xml.Name
			snm.Space = ""
			snm.Local = opts.key
			sel.Name = snm

			if err := enc.EncodeToken(sel); err != nil {
				fmt.Printf("error: %v\n", err)
			}
			for f := range a[i].Field {
				var el xml.StartElement
				var en xml.Name
				en.Space = ""
				en.Local = strings.ToLower(a[i].Field[f])
				el.Name = en
				if err := enc.EncodeElement("",el); err != nil {
					fmt.Printf("error: %v\n", err)
				}
			}
			if err := enc.EncodeToken(sel.End()); err != nil {
				fmt.Printf("error: %v\n", err)
			}
			if err := enc.Flush();  err != nil {
				fmt.Printf("error: %v\n", err)
			}
		}
	}

	if found {
		return &RepoDoc{
			content: w.String(),
		}, nil
	} else {
		return &RepoDoc{
			content: "specified kind is not found, please use \"dms3fs index config\" command to configure.",
		}, nil
	}
}
