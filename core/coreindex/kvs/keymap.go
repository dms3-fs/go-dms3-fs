package coreindex

import (
    "errors"
    "fmt"
    "path"
    "strconv"

    ds "github.com/dms3-fs/go-datastore"
)

//
// reposet key convention
// 	  - <index>/reposet/<type>/<kind>/<name>
//
const rootPrefix = "/index/reposet"
//
// corpus key convention
// 	  - <index>/reposet/<type>/<kind>/<name>/<reponame>/corpus
//
const corpusDocPrefix = "/corpus"

func GetRepoSetKey(t, k, n string) (ds.Key, error) {
    key := ds.NewKey(path.Join(rootPrefix, t, k, n))
    return key, nil
}

func DecomposeRepoSetKey(k string) (rtype, rkind, rname string, err error) {
    // verify key length
    key := ds.NewKey(k)
    kl := key.List()
    if len(kl) < 4 {
        err = errors.New(fmt.Sprintf("invalid reposet key length %v\n", key))
        return
    }
    // verify key prefix
    rootKey := ds.NewKey(rootPrefix)
    rl := rootKey.List()
    for i, _ := range rl {
        if rl[i] != kl[i] {
            err = errors.New(fmt.Sprintf("invalid reposet key prefix %v\n", key))
            return
        }
    }
    // extract and return reposet class and name
    switch len(kl) {
    case 4:
        // depricated
        rtype = kl[len(rl)]
        rkind = ""
        rname = kl[len(rl)+1]
    case 5:
        rtype = kl[len(rl)]
        rkind = kl[len(rl)+1]
        rname = kl[len(rl)+2]
    default:
        err = errors.New(fmt.Sprintf("invalid reposet key length %v\n", key))
    }
    return
}

func GetDocKey(rc string, rn string, ri int64, di int64) (ds.Key, error) {
    // Key: rootPrefix + "/_class_/_name_/_n_/corpus/_i_"
    key := ds.NewKey(path.Join(rootPrefix, rc, rn, strconv.FormatInt(ri, 10), corpusDocPrefix, strconv.FormatInt(di, 10)))
    return key, nil
}
