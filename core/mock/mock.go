package coremock

import (
	"context"

	commands "github.com/dms3-fs/go-dms3-fs/commands"
	core "github.com/dms3-fs/go-dms3-fs/core"
	"github.com/dms3-fs/go-dms3-fs/repo"

	datastore "github.com/dms3-fs/go-datastore"
	syncds "github.com/dms3-fs/go-datastore/sync"
	config "github.com/dms3-fs/go-fs-config"
	dms3p2p "github.com/dms3-p2p/go-p2p"
	host "github.com/dms3-p2p/go-p2p-host"
	peer "github.com/dms3-p2p/go-p2p-peer"
	pstore "github.com/dms3-p2p/go-p2p-peerstore"
	mocknet "github.com/dms3-p2p/go-p2p/p2p/net/mock"
	testutil "github.com/dms3-p2p/go-testutil"
)

// NewMockNode constructs an Dms3FsNode for use in tests.
func NewMockNode() (*core.Dms3FsNode, error) {
	ctx := context.Background()

	// effectively offline, only peer in its network
	return core.NewNode(ctx, &core.BuildCfg{
		Online: true,
		Host:   MockHostOption(mocknet.New(ctx)),
	})
}

func MockHostOption(mn mocknet.Mocknet) core.HostOption {
	return func(ctx context.Context, id peer.ID, ps pstore.Peerstore, _ ...dms3p2p.Option) (host.Host, error) {
		return mn.AddPeerWithPeerstore(id, ps)
	}
}

func MockCmdsCtx() (commands.Context, error) {
	// Generate Identity
	ident, err := testutil.RandIdentity()
	if err != nil {
		return commands.Context{}, err
	}
	p := ident.ID()

	conf := config.Config{
		Identity: config.Identity{
			PeerID: p.String(),
		},
	}

	r := &repo.Mock{
		D: syncds.MutexWrap(datastore.NewMapDatastore()),
		C: conf,
	}

	node, err := core.NewNode(context.Background(), &core.BuildCfg{
		Repo: r,
	})
	if err != nil {
		return commands.Context{}, err
	}

	return commands.Context{
		Online:     true,
		ConfigRoot: "/tmp/.mockdms3fsconfig",
		LoadConfig: func(path string) (*config.Config, error) {
			return &conf, nil
		},
		ConstructNode: func() (*core.Dms3FsNode, error) {
			return node, nil
		},
	}, nil
}
