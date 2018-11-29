package coreindex
import (
	"strings"
	"errors"
	"fmt"
	"io"
	"time"
	"os"
	"encoding/xml"
	"bytes"
	"path/filepath"

	"github.com/facebookgo/atomicfile"

    idxconfig "github.com/dms3-fs/go-idx-config"
)

type makeindexOpts struct {
	key string
	enc string
}

func IsKindConfigured(value interface{}, kind string) (found bool, err error) {

	if iconf, ok := value.(*idxconfig.IdxConfig); ok {
		a := iconf.Metadata.Kind

		for i := range a {
			if a[i].Name == kind {
				for f := range a[i].Field {
					if a[i].Field[f] != "" {
						found = true
					}
				}
			}
		}
	}
	return
}


func MakeRepo(cfg interface{}, path, kind string) (filename, reponame string, ctime time.Time, err error) {

	var found bool

	if filename, reponame, ctime, err = repoParamFilename(path); err != nil {
		return
	}

	// make path to repo root and indexer params
	found, err = writeParamFile(filename, cfg, kind)

	// make subfolders after creating params file, which also create path to repo root
	if err = makeRepoSubDirs(filename, reponame); err != nil {
		return
	}

	if found {
		return
	} else {
		err = errors.New("failed to make index repository, please use \"dms3fs index config\" command to verify configure.")
		return
	}
}

func MakeDoc(iconf idxconfig.IdxConfig, kind string) (string, error) {

	var found bool = false

	var b []byte
	w := bytes.NewBuffer(b)

	a := iconf.Metadata.Kind
	for i := range a {

		if a[i].Name == kind {
			found = true

			enc := xml.NewEncoder(w)
			enc.Indent("  ", "    ")
			var sel xml.StartElement
			var snm xml.Name
			snm.Space = ""
			snm.Local = kind
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
		return w.String(), nil
	} else {
		return "", errors.New("specified kind is not found, please use \"dms3fs index config\" command to configure.")
	}
}

// repoParamFilename returns the index repository parameters filename for a specified reposet kind root path.
func repoParamFilename(reporoot string) (filename, reponame string, createtime time.Time, err error) {

	//
	// local filesystem repository file folder hierarchy.
	//
	// index <repo root>, cfg parameter Indexer.Path, must be relative path
	// 	  - <index>
	// reposet root
	// 	  - <index>/reposet
	// reposet kind root
	//	  - <index>/reposet/<kind>
	// reposet root folder
	// 	  - <index>/reposet/<kind>/<name>
	// reponame, composed as:
	//	  - window: uint64,    // creation time (Unix, seconds), sharding tag
	//	  - area: uint8,       // area number, sharding tag
	//	  - cat: uint8,        // category number, sharding tag
	//	  - offset: int64,     // time since creation (seconds), recovery tag
	// repo specific files and sub-folders, cfg parameters are ignored for these
	//	  - <index>/reposet/<kind>/<name>/<reponame>/corpus, cfg parameter Indexer.Corpus.Path
	//	  - <index>/reposet/<kind>/<name>/<reponame>/metadata, cfg parameter Indexer.Corpus.Metadata
	// 	  - <index>/reposet/<kind>/<name>/params, repo params file, no corresponding cfg parameter
	// 	    params file is common for all repos in a reposet
	//
	t := time.Now()	// repo create time
	o := t.Sub(t)	// seconds since repo create (zero at creation time)
	a := 0			// current area (zero ==> none)
	c := 0			// current category (zero ==> none)
	createtime = t

	window := fmt.Sprintf("w%d", t.Unix())	// seconds since Unix epoch
	area := fmt.Sprintf("-a%d", a+1)		// start at 1. 0 ==> N/A
	category := fmt.Sprintf("-c%d", c+1)	// start at 1. 0 ==> N/A
	offset := fmt.Sprintf("-o%d", o)		// 0 offset is ok

	reponame = window + area + category + offset

	filename = filepath.Join(reporoot, "params")

	return
}

// writeParamFile writes the index repository parameters from `cfg` into `filename`.
func writeParamFile(filename string, cfg interface{}, kind string) (bool, error) {
        err := os.MkdirAll(filepath.Dir(filename), 0775)
        if err != nil {
                return false, err
        }

        f, err := atomicfile.New(filename, 0660)
        if err != nil {
                return false, err
        }
        defer f.Close()

        return encode(f, cfg, kind)
}

// given repo params file, create repo subfolders
// cfg path parameters are ignored for these
//	  - Indexer.Corpus.Path
//	  - Indexer.Corpus.Metadata
func makeRepoSubDirs(filename, reponame string) error {

	var i, c, m, r string

	r = filepath.Dir(filename)
	r = filepath.Join(r, reponame)
	i = filepath.Join(r, "index")
	c = filepath.Join(r, "corpus")
	m = filepath.Join(r, "metadata")

	//fmt.Printf("index: %v\n", i)
	//fmt.Printf("corpus: %v\n", c)
	//fmt.Printf("metadata: %v\n", m)

	if err := os.Mkdir(r, 0775); err != nil {
		return err
	}
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

func encode(w io.Writer, value interface{}, kind string) (found bool, err error) {

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

	if found, err = makeIndexFieldsParams(value, enc, kind); err != nil {
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

func makeIndexFieldsParams(value interface{}, enc *xml.Encoder, kind string) (found bool, err error) {

	if iconf, ok := value.(*idxconfig.IdxConfig); ok {
		a := iconf.Metadata.Kind

		for i := range a {

			if a[i].Name == kind {
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
				if err := enc.EncodeElement(strings.ToLower(kind),knel); err != nil {
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
