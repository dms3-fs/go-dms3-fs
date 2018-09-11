package coredag

import (
	"io"
	"io/ioutil"

	dms3ldcbor "github.com/dms3-fs/go-ld-cbor"
	dms3ld "github.com/dms3-fs/go-ld-format"
)

func cborJSONParser(r io.Reader, mhType uint64, mhLen int) ([]dms3ld.Node, error) {
	nd, err := dms3ldcbor.FromJson(r, mhType, mhLen)
	if err != nil {
		return nil, err
	}

	return []dms3ld.Node{nd}, nil
}

func cborRawParser(r io.Reader, mhType uint64, mhLen int) ([]dms3ld.Node, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	nd, err := dms3ldcbor.Decode(data, mhType, mhLen)
	if err != nil {
		return nil, err
	}

	return []dms3ld.Node{nd}, nil
}
