package coredag

import (
	"io"
	"io/ioutil"
	"math"

	"github.com/dms3-fs/go-merkledag"

	cid "github.com/dms3-fs/go-cid"
	dms3ld "github.com/dms3-fs/go-ld-format"
	mh "github.com/dms3-mft/go-multihash"
)

func dagpbJSONParser(r io.Reader, mhType uint64, mhLen int) ([]dms3ld.Node, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	nd := &merkledag.ProtoNode{}

	err = nd.UnmarshalJSON(data)
	if err != nil {
		return nil, err
	}

	nd.SetCidBuilder(cidPrefix(mhType, mhLen))

	return []dms3ld.Node{nd}, nil
}

func dagpbRawParser(r io.Reader, mhType uint64, mhLen int) ([]dms3ld.Node, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	nd, err := merkledag.DecodeProtobuf(data)
	if err != nil {
		return nil, err
	}

	nd.SetCidBuilder(cidPrefix(mhType, mhLen))

	return []dms3ld.Node{nd}, nil
}

func cidPrefix(mhType uint64, mhLen int) *cid.Prefix {
	if mhType == math.MaxUint64 {
		mhType = mh.SHA2_256
	}

	prefix := &cid.Prefix{
		MhType:   mhType,
		MhLength: mhLen,
		Version:  1,
		Codec:    cid.DagProtobuf,
	}

	if mhType == mh.SHA2_256 {
		prefix.Version = 0
	}

	return prefix
}
