package coreindex

import (
    "bytes"
    "context"
    "fmt"
    "testing"
    "time"

    dstest "github.com/dms3-fs/go-merkledag/test"
)

// Test Reposet DAG node Add/Get using Directory and File objects
func TestReposetAddGetDirFile(t *testing.T) {
    ctx := context.Background()

    ds := dstest.Mock()

    rps := NewReposetProps()

    rps.SetType("infostore")
    rps.SetKind("blog")
    rps.SetName("mytestblog")
    rps.SetCreatedAt(uint64(time.Now().Unix()))
    rps.SetMaxAreas(0)
    rps.SetMaxCats(0)
    rps.SetMaxDocs(50000000)

    sr, err := NewStoreRoot(ctx, ds, nil)
    if err != nil {
         t.Fatal(err)
    }

    rpsid, err := sr.AddProps("reposetprops", rps)
    if err != nil {
         t.Fatal(err)
    }

    rrpsid, err := sr.HasProps("reposetprops", rps)
    if err != nil {
         t.Fatal(err)
    }

    if rrpsid == nil {
         t.Fatal("added reposetprops does not exist.")
    }

    if rrpsid.String() != rpsid.String() {
        t.Fatal(fmt.Errorf("repoprops cid mismatch, expected %v found %v", rpsid.String(), rrpsid.String()))
    }

    // validate the bytes we wrote
    b, err := rps.Marshal()
    if err != nil {
        t.Fatal(err)
    }

    bb, size, err := sr.bytesFromFile("reposetprops", int64(len(b)))
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
    ri, err := sr.GetProps("reposetprops", rps)
    if err != nil {
         t.Fatal(err)
    }

    rps2, ok := ri.(ReposetProps)
    if !ok {
        t.Fatal("invalid return type, expected.")
    }

    // assert object from file content matches object we wrote
    if !rps.Equal(rps2) {
        t.Fatal("data read was different than data written")
    }
}
