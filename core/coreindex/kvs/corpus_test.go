package coreindex

import (
    "fmt"
    "strings"
    "testing"

    cid "github.com/dms3-fs/go-cid"
    ds "github.com/dms3-fs/go-datastore"
    dsquery "github.com/dms3-fs/go-datastore/query"
    mh "github.com/dms3-mft/go-multihash"
)

func TestCorpusMarshal(t *testing.T) {

    data := []byte("this is some test content")
    hash, _ := mh.Sum(data, mh.SHA2_256, -1)
    id := cid.NewCidV1(cid.Raw, hash)

    c1 := corpusProps{
        Rclass: "testclass",
        Rkind:  "testkind",
        Rindex: int64(55),
        Rcid:   id,
    }

    if b, err := c1.Marshal(); err != nil {
        t.Fatal(err)
    } else {

        c2 := corpusProps{
            Rclass: "",
            Rkind:  "",
            Rindex: 0,
            Rcid:   nil,
        }

        if err := c2.Unmarshal(b); err != nil {
            t.Fatal(err)
        } else {
            if !c2.Equals(&c1) {
                t.Fatal(err)
            }
        }
    }
}

func TestCorpusPutGetDel(t *testing.T) {

    dstore := GetIndexKVStore()

    c1 := corpusProps{
        Rclass: "testkeystore",
        Rkind:  "testkeykind",
        Rindex: int64(35),
        Rcid:   nil,
    }

    c2 := corpusProps{
        Rclass: "",
        Rkind:  "",
        Rindex: 0,
        Rcid:   nil,
    }

	var i int64
	var sb strings.Builder
    var key ds.Key
    var value []byte
    var err error

	for i = 0; i < 100; i++ {
        sb.Reset()
		fmt.Fprintf(&sb, "this is some test key %d", i)
		if id, _ := cid.NewPrefixV1(cid.Raw, mh.ID).Sum([]byte(sb.String())); id == nil {
            t.Fatal(fmt.Errorf("cannot compute cid corpus property"))
        } else {
            c1.Rcid = id
        }

        if value, err = c1.Marshal(); err != nil {
            t.Fatal(err)
        }

        if key, err = GetDocKey(c1.Rclass, c1.Rkind, c1.Rindex, i); err != nil {
            t.Fatal(err)
        }

        if err = dstore.Put(key, value); err != nil {
            t.Fatal(err)
        }
	}

	for i = 0; i < 100; i++ {
        sb.Reset()
		fmt.Fprintf(&sb, "this is some test key %d", i)
		if id, _ := cid.NewPrefixV1(cid.Raw, mh.ID).Sum([]byte(sb.String())); id == nil {
            //return fmt.Errorf("cannot compute cid corpus property: %v", err)
            t.Fatal(fmt.Errorf("cannot compute cid corpus property"))
        } else {
            c1.Rcid = id
        }

        if key, err = GetDocKey(c1.Rclass, c1.Rkind, c1.Rindex, i); err != nil {
            //return fmt.Errorf("cannot get key for corpus properties: %v", err)
            t.Fatal(err)
        }

        if value, err = dstore.Get(key); err != nil {
            //return fmt.Errorf("cannot get corpus properties: %v", err)
            t.Fatal(err)
		}

        if err = c2.Unmarshal(value); err != nil {
            //return nil, fmt.Errorf("cannot unmarshal corpus properties: %v", err)
            t.Fatal(err)
        }

        if !c2.Equals(&c1) {
            t.Fatal(fmt.Errorf("put/get value mistmatch corpus property"))
        }
	}

	for i = 0; i < 100; i++ {
        if key, err = GetDocKey(c1.Rclass, c1.Rkind, c1.Rindex, i); err != nil {
            //return fmt.Errorf("cannot get key for corpus properties: %v", err)
            t.Fatal(err)
        }

		if err = dstore.Delete(key); err != nil {
            //return fmt.Errorf("cannot delete corpus properties: %v", err)
            t.Fatal(err)
		}
	}
}

func TestCorpusHasQuery(t *testing.T) {

    dstore := GetIndexKVStore()

    c1 := corpusProps{
        Rclass: "testkeystore",
        Rkind:  "testkeykind",
        Rindex: int64(35),
        Rcid:   nil,
    }

	var i int64
	var sb strings.Builder
    var key ds.Key
    var value []byte
    var err error
    var exists bool

    // remember what we store so we later verify testing the query interface
    var keys map[ds.Key][]byte = make(map[ds.Key][]byte,100)

	for i = 0; i < 100; i++ {
        sb.Reset()
		fmt.Fprintf(&sb, "this is some test key %d", i)
		if id, _ := cid.NewPrefixV1(cid.Raw, mh.ID).Sum([]byte(sb.String())); id == nil {
            t.Fatal(fmt.Errorf("cannot compute cid corpus property"))
        } else {
            c1.Rcid = id
        }

        if value, err = c1.Marshal(); err != nil {
            t.Fatal(err)
        }

        if key, err = GetDocKey(c1.Rclass, c1.Rkind, c1.Rindex, i); err != nil {
            t.Fatal(err)
        }

        if err = dstore.Put(key, value); err != nil {
            t.Fatal(err)
        }
        // remember for later verification
        keys[key] = value
	}

    // verify store properly responds to Has requests
	for i = 0; i < 100; i++ {
        if key, err = GetDocKey(c1.Rclass, c1.Rkind, c1.Rindex, i); err != nil {
            t.Fatal(err)
        }

        if exists, err = dstore.Has(key); err != nil {
            t.Fatal(fmt.Errorf("cannot issue Has request %v\n.", err))
		}

        if !exists {
            t.Fatal(fmt.Errorf("Has fails after put to corpus property."))
        }
	}

    // verify store properly responds to Query requests
    if res, err := dstore.Query(dsquery.Query{}); err != nil {
        t.Fatal(fmt.Errorf("cannot issue Query request %v\n.", err))
    } else {
        defer res.Close()
OuterLoop:
        for {
            select {
            case result, ok := <-res.Next():
                if !ok {
                    // no more result left
                    //fmt.Printf("results Done!\n")
                    break OuterLoop
                }
                if result.Error != nil {
                    t.Fatal(fmt.Errorf("Query returned internal error %v\n.", err))
                }
                r2 := corpusProps{}
                if err := r2.Unmarshal(result.Value); err != nil {
                    t.Fatal(err)
                    t.Fatal(fmt.Errorf("cannot unmarshal Query returned value."))
                }
                //fmt.Printf("result key %v value %v\n", result.Key, rp2)
                switch keys[ds.NewKey(result.Key)] {
                case nil:
                    t.Fatal(fmt.Errorf("Query returned value has nil."))
                default:
                    v := keys[ds.NewKey(result.Key)]
                    if len(v) != len(result.Value) {
                        t.Fatal(fmt.Errorf("Query returned value has wrong length."))
                    }
                    for i := 0; i < len(v); i++ {
                        if v[i] != result.Value[i] {
                            t.Fatal(fmt.Errorf("Query returned value has wrong content."))
                        }
                    }
                    delete(keys, ds.NewKey(result.Key))
                    //fmt.Printf("length of keys %v\n", len(keys))
                }
            }
        }
    }
    if len(keys) > 0 {
        t.Fatal(fmt.Errorf("Query did not return all stored records."))
    }

    for i = 0; i < 100; i++ {
        if key, err = GetDocKey(c1.Rclass, c1.Rkind, c1.Rindex, i); err != nil {
            t.Fatal(err)
        }
		if err = dstore.Delete(key); err != nil {
            t.Fatal(err)
		}
	}
}
