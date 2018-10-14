package coreindex

import (
    "fmt"
    "testing"

    ds "github.com/dms3-fs/go-datastore"
    dsquery "github.com/dms3-fs/go-datastore/query"
)

func TestRepoMarshal(t *testing.T) {

    rp := repoProps{
        Suffix: suffix{
            Offset: 60,
            Area: 51,
            Category: 17,
        },
        Docs: 123456,
    }

    if b, err := rp.Marshal(); err != nil {
        t.Fatal(err)
    } else {
        //fmt.Printf("marshalled repo props %+v", b)

        rp2 := repoProps{
            Suffix: suffix{
                Offset: 0,
                Area: 0,
                Category: 0,
            },
            Docs: 0,
        }

        if err := rp2.Unmarshal(b); err != nil {
            t.Fatal(err)
        } else {
            if !rp.Equals(&rp2) {
                t.Fatal(err)
            } else {
                //fmt.Printf("unmarshalled repo props %+v", rp2)
            }
        }
    }
}

func TestRepoPutGetDel(t *testing.T) {

    const testclass string = "testclass"
    const testkind string = "testkind"

    dstore := GetIndexKVStore()

    rp := repoProps{
        Suffix: suffix{
            Offset: 60,
            Area: 51,
            Category: 17,
        },
        Docs: 123456,
    }

    rp2 := repoProps{
        Suffix: suffix{
            Offset: 0,
            Area: 0,
            Category: 0,
        },
        Docs: 0,
    }

	var i int64
    var key ds.Key
    var value []byte
    var err error

	for i = 0; i < 100; i++ {

        if key, err = GetRepoKey(testclass, testkind, i); err != nil {
            t.Fatal(err)
        }

        if value, err = rp.Marshal(); err != nil {
            t.Fatal(err)
        }

        if err = dstore.Put(key, value); err != nil {
            t.Fatal(err)
        }
	}

	for i = 0; i < 100; i++ {

        if key, err = GetRepoKey(testclass, testkind, i); err != nil {
            t.Fatal(err)
        }

        if value, err = dstore.Get(key); err != nil {
            t.Fatal(err)
		}

        if err = rp2.Unmarshal(value); err != nil {
            t.Fatal(err)
        }

        if !rp.Equals(&rp2) {
            t.Fatal(fmt.Errorf("put/get value mistmatch corpus property"))
        }
	}

	for i = 0; i < 100; i++ {
        if key, err = GetRepoKey(testclass, testkind, i); err != nil {
            t.Fatal(err)
        }

		if err = dstore.Delete(key); err != nil {
            t.Fatal(err)
		}
	}
}

func TestRepoHasQuery(t *testing.T) {

    const testclass string = "testclass"
    const testkind string = "testkind"

    dstore := GetIndexKVStore()

    rp := repoProps{
        Suffix: suffix{
            Offset: 160,
            Area: 21,
            Category: 7,
        },
        Docs: 654321,
    }

	var i int64
    var key ds.Key
    var value []byte
    var err error
    var exists bool

    // remember what we store so we later verify testing the query interface
    var keys map[ds.Key][]byte = make(map[ds.Key][]byte,100)

	for i = 0; i < 100; i++ {

        if key, err = GetRepoKey(testclass, testkind, i); err != nil {
            t.Fatal(err)
        }

        if value, err = rp.Marshal(); err != nil {
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

        if key, err = GetRepoKey(testclass, testkind, i); err != nil {
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
                rp2 := repoProps{}
                if err := rp2.Unmarshal(result.Value); err != nil {
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
        if key, err = GetRepoKey(testclass, testkind, i); err != nil {
            t.Fatal(err)
        }

		if err = dstore.Delete(key); err != nil {
            t.Fatal(err)
		}
	}
}
