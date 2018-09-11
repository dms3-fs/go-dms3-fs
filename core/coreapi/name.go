package coreapi

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	core "github.com/dms3-fs/go-dms3-fs/core"
	coreiface "github.com/dms3-fs/go-dms3-fs/core/coreapi/interface"
	caopts "github.com/dms3-fs/go-dms3-fs/core/coreapi/interface/options"
	keystore "github.com/dms3-fs/go-dms3-fs/keystore"
	namesys "github.com/dms3-fs/go-dms3-fs/namesys"
	nsopts "github.com/dms3-fs/go-dms3-fs/namesys/opts"
	ipath "github.com/dms3-fs/go-path"

	offline "github.com/dms3-fs/go-fs-routing/offline"
	crypto "github.com/dms3-p2p/go-p2p-crypto"
	peer "github.com/dms3-p2p/go-p2p-peer"
)

type NameAPI CoreAPI

type dms3nsEntry struct {
	name  string
	value coreiface.Path
}

// Name returns the dms3nsEntry name.
func (e *dms3nsEntry) Name() string {
	return e.name
}

// Value returns the dms3nsEntry value.
func (e *dms3nsEntry) Value() coreiface.Path {
	return e.value
}

// Publish announces new DMS3NS name and returns the new DMS3NS entry.
func (api *NameAPI) Publish(ctx context.Context, p coreiface.Path, opts ...caopts.NamePublishOption) (coreiface.Dms3NsEntry, error) {
	options, err := caopts.NamePublishOptions(opts...)
	if err != nil {
		return nil, err
	}
	n := api.node

	if !n.OnlineMode() {
		err := n.SetupOfflineRouting()
		if err != nil {
			return nil, err
		}
	}

	if n.Mounts.Dms3Ns != nil && n.Mounts.Dms3Ns.IsActive() {
		return nil, errors.New("cannot manually publish while DMS3NS is mounted")
	}

	pth, err := ipath.ParsePath(p.String())
	if err != nil {
		return nil, err
	}

	k, err := keylookup(n, options.Key)
	if err != nil {
		return nil, err
	}

	eol := time.Now().Add(options.ValidTime)
	err = n.Namesys.PublishWithEOL(ctx, k, pth, eol)
	if err != nil {
		return nil, err
	}

	pid, err := peer.IDFromPrivateKey(k)
	if err != nil {
		return nil, err
	}

	return &dms3nsEntry{
		name:  pid.Pretty(),
		value: p,
	}, nil
}

// Resolve attempts to resolve the newest version of the specified name and
// returns its path.
func (api *NameAPI) Resolve(ctx context.Context, name string, opts ...caopts.NameResolveOption) (coreiface.Path, error) {
	options, err := caopts.NameResolveOptions(opts...)
	if err != nil {
		return nil, err
	}

	n := api.node

	if !n.OnlineMode() {
		err := n.SetupOfflineRouting()
		if err != nil {
			return nil, err
		}
	}

	var resolver namesys.Resolver = n.Namesys

	if options.Local && !options.Cache {
		return nil, errors.New("cannot specify both local and nocache")
	}

	if options.Local {
		offroute := offline.NewOfflineRouter(n.Repo.Datastore(), n.RecordValidator)
		resolver = namesys.NewDms3NsResolver(offroute)
	}

	if !options.Cache {
		resolver = namesys.NewNameSystem(n.Routing, n.Repo.Datastore(), 0)
	}

	if !strings.HasPrefix(name, "/dms3ns/") {
		name = "/dms3ns/" + name
	}

	var ropts []nsopts.ResolveOpt
	if !options.Recursive {
		ropts = append(ropts, nsopts.Depth(1))
	}

	output, err := resolver.Resolve(ctx, name, ropts...)
	if err != nil {
		return nil, err
	}

	return coreiface.ParsePath(output.String())
}

func keylookup(n *core.Dms3FsNode, k string) (crypto.PrivKey, error) {
	res, err := n.GetKey(k)
	if res != nil {
		return res, nil
	}

	if err != nil && err != keystore.ErrNoSuchKey {
		return nil, err
	}

	keys, err := n.Repo.Keystore().List()
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		privKey, err := n.Repo.Keystore().Get(key)
		if err != nil {
			return nil, err
		}

		pubKey := privKey.GetPublic()

		pid, err := peer.IDFromPublicKey(pubKey)
		if err != nil {
			return nil, err
		}

		if pid.Pretty() == k {
			return privKey, nil
		}
	}

	return nil, fmt.Errorf("no key by the given name or PeerID was found")
}
