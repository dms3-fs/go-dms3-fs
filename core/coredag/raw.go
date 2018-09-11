package coredag

import (
	"io"
	"io/ioutil"
	"math"

	"github.com/dms3-fs/go-merkledag"

	block "github.com/dms3-fs/go-block-format"
	cid "github.com/dms3-fs/go-cid"
	dms3ld "github.com/dms3-fs/go-ld-format"
	mh "github.com/dms3-mft/go-multihash"
)

func rawRawParser(r io.Reader, mhType uint64, mhLen int) ([]dms3ld.Node, error) {
	if mhType == math.MaxUint64 {
		mhType = mh.SHA2_256
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	h, err := mh.Sum(data, mhType, mhLen)
	if err != nil {
		return nil, err
	}
	c := cid.NewCidV1(cid.Raw, h)
	blk, err := block.NewBlockWithCid(data, c)
	if err != nil {
		return nil, err
	}
	nd := &merkledag.RawNode{Block: blk}
	return []dms3ld.Node{nd}, nil
}
