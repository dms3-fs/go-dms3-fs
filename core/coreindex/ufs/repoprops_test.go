package coreindex

import (
    "bytes"
    "context"
    "fmt"
    "testing"

    dstest "github.com/dms3-fs/go-merkledag/test"
)

// Test Repo DAG node Add/Get using Directory and File objects
func TestRepoAddGetDirFile(t *testing.T) {
    ctx := context.Background()

    ds := dstest.Mock()

    rp := NewRepoProps()

    rp.SetType("infostore")
    rp.SetKind("blog")
    rp.SetName("mytestblog")
    rp.SetOffset(0)
    rp.SetArea(0)
    rp.SetCat(0)
    rp.SetPath("/index/reposet/mytestblog")

    sr, err := NewStoreRoot(ctx, ds, nil)
    if err != nil {
         t.Fatal(err)
    }

    rpid, err := sr.AddProps("repoprops", rp)
    if err != nil {
         t.Fatal(err)
    }

    rrpid, err := sr.HasProps("repoprops", rp)
    if err != nil {
         t.Fatal(err)
    }

    if rrpid == nil {
         t.Fatal("added repoprops does not exist.")
    }

    if rrpid.String() != rpid.String() {
        t.Fatal(fmt.Errorf("repoprops cid mismatch, expected %v found %v", rpid.String(), rrpid.String()))
    }

    // validate the bytes we wrote
    b, err := rp.Marshal()
    if err != nil {
        t.Fatal(err)
    }

    bb, size, err := sr.bytesFromFile("repoprops", int64(len(b)))
    if err != nil {
         t.Fatal(err)
    }

    // assert size is as expected
    if size != int64(len(b)) {
        t.Fatal("size is incorrect")
    }

    // assert valid bytes
    if !bytes.Equal(bb, b) {
        t.Fatal("data read was different than data written")
    }

    // create repoprops from file in directory
    ri, err := sr.GetProps("repoprops", rp)
    if err != nil {
         t.Fatal(err)
    }

    rp2, ok := ri.(RepoProps)
    if !ok {
        t.Fatal("invalid return type, expected.")
    }

    // assert object from file content matches object we wrote
    if !rp.Equal(rp2) {
        t.Fatal("data read was different than data written")
    }

}
