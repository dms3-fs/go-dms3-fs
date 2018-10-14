package coreindex

import (
	"fmt"
    "sync"

    ds "github.com/dms3-fs/go-datastore"
    query "github.com/dms3-fs/go-datastore/query"
)

type kvstore struct {
    lock    sync.RWMutex
    d       ds.Datastore
}

type KVStore interface {
    ds.Datastore
/*
    Put(i int64, value []byte) error
    Get(i int64) (value []byte, err error)
    Delete(i int64) error
*/
}

var kvs *kvstore

func init() {
    kvs = &kvstore{
        d:      ds.NewMapDatastore(), // tests use go map []byte datastore
    }
}

// call to change datastore to use
func InitIndexKVStore(d ds.Datastore) {
    kvs.d = d
}

func GetIndexKVStore() KVStore{
    return kvs
}

func (kvs *kvstore) Put(key ds.Key, value []byte) error {
    kvs.lock.Lock()
	defer kvs.lock.Unlock()

    if err := kvs.d.Put(key, value); err != nil {
        return fmt.Errorf("cannot store key value properties: %v", err)
    }
    return nil
}

func (kvs *kvstore) Get(key ds.Key) (value []byte, err error) {
    kvs.lock.Lock()
	defer kvs.lock.Unlock()

    //log.Debugf("Get key %v\n", key)
    if value, err := kvs.d.Get(key); err != nil {
        return nil, fmt.Errorf("cannot store key value properties: %v", err)
    } else {
        return value, nil
    }
}

func (kvs *kvstore) Has(key ds.Key) (exists bool, err error) {
    return kvs.d.Has(key)
}

func (kvs *kvstore) Delete(key ds.Key) error {
    kvs.lock.Lock()
	defer kvs.lock.Unlock()

    //log.Debugf("Delete key %v\n", key)
    if err := kvs.d.Delete(key); err != nil {
        return fmt.Errorf("cannot store key value properties: %v", err)
    }
    return nil
}

func (kvs *kvstore) Query(q query.Query) (query.Results, error) {
    return kvs.d.Query(q)

    //var r []query.Entry
    //return  query.ResultsWithEntries(q, r), nil

    //var c <-chan query.Result = make(chan query.Result, 1)
    //return  query.ResultsWithChan(q, c), nil
}
