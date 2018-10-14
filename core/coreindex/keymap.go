package coreindex

import (
    "path"
    "strconv"

    ds "github.com/dms3-fs/go-datastore"
)

const corpusDocPrefix = "/corpus"

func GetRepoSetKey(rc string, rk string) (ds.Key, error) {
    // repository properties Key: "/_class_/_kind_/"
    key := ds.NewKey(path.Join(rc, rk))
    return key, nil
}

func GetRepoKey(rc string, rk string, ri int64) (ds.Key, error) {
    // repository properties Key: "/_class_/_kind_/_n_/"
    key := ds.NewKey(path.Join(rc, rk, strconv.FormatInt(ri, 10)))
    return key, nil
}

func GetDocKey(rc string, rk string, ri int64, di int64) (ds.Key, error) {
    // Key: "/_class_/_kind_/_n_/corpus/_i_"
    key := ds.NewKey(path.Join(rc, rk, strconv.FormatInt(ri, 10), corpusDocPrefix, strconv.FormatInt(di, 10)))
    return key, nil
}
