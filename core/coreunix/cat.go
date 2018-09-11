package coreunix

import (
	"context"

	core "github.com/dms3-fs/go-dms3-fs/core"
	path "github.com/dms3-fs/go-path"
	resolver "github.com/dms3-fs/go-path/resolver"
	uio "github.com/dms3-fs/go-unixfs/io"
)

func Cat(ctx context.Context, n *core.Dms3FsNode, pstr string) (uio.DagReader, error) {
	r := &resolver.Resolver{
		DAG:         n.DAG,
		ResolveOnce: uio.ResolveUnixfsOnce,
	}

	dagNode, err := core.Resolve(ctx, n.Namesys, r, path.Path(pstr))
	if err != nil {
		return nil, err
	}

	return uio.NewDagReader(ctx, dagNode, n.DAG)
}
