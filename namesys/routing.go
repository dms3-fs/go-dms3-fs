package namesys

import (
	"context"
	"strings"
	"time"

	opts "github.com/dms3-fs/go-dms3-fs/namesys/opts"
	path "github.com/dms3-fs/go-path"

	proto "github.com/gogo/protobuf/proto"
	cid "github.com/dms3-fs/go-cid"
	dms3ns "github.com/dms3-fs/go-dms3ns"
	pb "github.com/dms3-fs/go-dms3ns/pb"
	logging "github.com/dms3-fs/go-log"
	dht "github.com/dms3-p2p/go-p2p-kad-dht"
	peer "github.com/dms3-p2p/go-p2p-peer"
	routing "github.com/dms3-p2p/go-p2p-routing"
	mh "github.com/dms3-mft/go-multihash"
)

var log = logging.Logger("namesys")

// Dms3NsResolver implements NSResolver for the main DMS3FS SFS-like naming
type Dms3NsResolver struct {
	routing routing.ValueStore
}

// NewDms3NsResolver constructs a name resolver using the DMS3FS Routing system
// to implement SFS-like naming on top.
func NewDms3NsResolver(route routing.ValueStore) *Dms3NsResolver {
	if route == nil {
		panic("attempt to create resolver with nil routing system")
	}
	return &Dms3NsResolver{
		routing: route,
	}
}

// Resolve implements Resolver.
func (r *Dms3NsResolver) Resolve(ctx context.Context, name string, options ...opts.ResolveOpt) (path.Path, error) {
	return resolve(ctx, r, name, opts.ProcessOpts(options), "/dms3ns/")
}

// resolveOnce implements resolver. Uses the DMS3FS routing system to
// resolve SFS-like names.
func (r *Dms3NsResolver) resolveOnce(ctx context.Context, name string, options *opts.ResolveOpts) (path.Path, time.Duration, error) {
	log.Debugf("RoutingResolver resolving %s", name)

	if options.DhtTimeout != 0 {
		// Resolution must complete within the timeout
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.DhtTimeout)
		defer cancel()
	}

	name = strings.TrimPrefix(name, "/dms3ns/")
	hash, err := mh.FromB58String(name)
	if err != nil {
		// name should be a multihash. if it isn't, error out here.
		log.Debugf("RoutingResolver: bad input hash: [%s]\n", name)
		return "", 0, err
	}

	pid, err := peer.IDFromBytes(hash)
	if err != nil {
		log.Debugf("RoutingResolver: could not convert public key hash %s to peer ID: %s\n", name, err)
		return "", 0, err
	}

	// Name should be the hash of a public key retrievable from dms3fs.
	// We retrieve the public key here to make certain that it's in the peer
	// store before calling GetValue() on the DHT - the DHT will call the
	// dms3ns validator, which in turn will get the public key from the peer
	// store to verify the record signature
	_, err = routing.GetPublicKey(r.routing, ctx, pid)
	if err != nil {
		log.Debugf("RoutingResolver: could not retrieve public key %s: %s\n", name, err)
		return "", 0, err
	}

	// Use the routing system to get the name.
	// Note that the DHT will call the dms3ns validator when retrieving
	// the value, which in turn verifies the dms3ns record signature
	dms3nsKey := dms3ns.RecordKey(pid)
	val, err := r.routing.GetValue(ctx, dms3nsKey, dht.Quorum(int(options.DhtRecordCount)))
	if err != nil {
		log.Debugf("RoutingResolver: dht get for name %s failed: %s", name, err)
		return "", 0, err
	}

	entry := new(pb.Dms3NsEntry)
	err = proto.Unmarshal(val, entry)
	if err != nil {
		log.Debugf("RoutingResolver: could not unmarshal value for name %s: %s", name, err)
		return "", 0, err
	}

	var p path.Path
	// check for old style record:
	if valh, err := mh.Cast(entry.GetValue()); err == nil {
		// Its an old style multihash record
		log.Debugf("encountered CIDv0 dms3ns entry: %s", valh)
		p = path.FromCid(cid.NewCidV0(valh))
	} else {
		// Not a multihash, probably a new record
		p, err = path.ParsePath(string(entry.GetValue()))
		if err != nil {
			return "", 0, err
		}
	}

	ttl := DefaultResolverCacheTTL
	if entry.Ttl != nil {
		ttl = time.Duration(*entry.Ttl)
	}
	switch eol, err := dms3ns.GetEOL(entry); err {
	case dms3ns.ErrUnrecognizedValidity:
		// No EOL.
	case nil:
		ttEol := eol.Sub(time.Now())
		if ttEol < 0 {
			// It *was* valid when we first resolved it.
			ttl = 0
		} else if ttEol < ttl {
			ttl = ttEol
		}
	default:
		log.Errorf("encountered error when parsing EOL: %s", err)
		return "", 0, err
	}

	return p, ttl, nil
}
