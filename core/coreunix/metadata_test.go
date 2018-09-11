package coreunix

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	bserv "github.com/dms3-fs/go-blockservice"
	core "github.com/dms3-fs/go-dms3-fs/core"
	merkledag "github.com/dms3-fs/go-merkledag"
	ft "github.com/dms3-fs/go-unixfs"
	importer "github.com/dms3-fs/go-unixfs/importer"
	uio "github.com/dms3-fs/go-unixfs/io"

	cid "github.com/dms3-fs/go-cid"
	ds "github.com/dms3-fs/go-datastore"
	dssync "github.com/dms3-fs/go-datastore/sync"
	bstore "github.com/dms3-fs/go-fs-blockstore"
	chunker "github.com/dms3-fs/go-fs-chunker"
	offline "github.com/dms3-fs/go-fs-exchange-offline"
	u "github.com/dms3-fs/go-fs-util"
	dms3ld "github.com/dms3-fs/go-ld-format"
)

func getDagserv(t *testing.T) dms3ld.DAGService {
	db := dssync.MutexWrap(ds.NewMapDatastore())
	bs := bstore.NewBlockstore(db)
	blockserv := bserv.New(bs, offline.Exchange(bs))
	return merkledag.NewDAGService(blockserv)
}

func TestMetadata(t *testing.T) {
	ctx := context.Background()
	// Make some random node
	ds := getDagserv(t)
	data := make([]byte, 1000)
	u.NewTimeSeededRand().Read(data)
	r := bytes.NewReader(data)
	nd, err := importer.BuildDagFromReader(ds, chunker.DefaultSplitter(r))
	if err != nil {
		t.Fatal(err)
	}

	c := nd.Cid()

	m := new(ft.Metadata)
	m.MimeType = "THIS IS A TEST"

	// Such effort, many compromise
	dms3fsnode := &core.Dms3FsNode{DAG: ds}

	mdk, err := AddMetadataTo(dms3fsnode, c.String(), m)
	if err != nil {
		t.Fatal(err)
	}

	rec, err := Metadata(dms3fsnode, mdk)
	if err != nil {
		t.Fatal(err)
	}
	if rec.MimeType != m.MimeType {
		t.Fatalf("something went wrong in conversion: '%s' != '%s'", rec.MimeType, m.MimeType)
	}

	cdk, err := cid.Decode(mdk)
	if err != nil {
		t.Fatal(err)
	}

	retnode, err := ds.Get(ctx, cdk)
	if err != nil {
		t.Fatal(err)
	}

	rtnpb, ok := retnode.(*merkledag.ProtoNode)
	if !ok {
		t.Fatal("expected protobuf node")
	}

	ndr, err := uio.NewDagReader(ctx, rtnpb, ds)
	if err != nil {
		t.Fatal(err)
	}

	out, err := ioutil.ReadAll(ndr)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(out, data) {
		t.Fatal("read incorrect data")
	}
}
