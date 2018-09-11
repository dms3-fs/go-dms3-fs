package namesys

import (
	"context"
	"strings"
	"sync"
	"time"

	pin "github.com/dms3-fs/go-dms3-fs/pin"
	path "github.com/dms3-fs/go-path"
	ft "github.com/dms3-fs/go-unixfs"

	proto "github.com/gogo/protobuf/proto"
	ds "github.com/dms3-fs/go-datastore"
	dsquery "github.com/dms3-fs/go-datastore/query"
	dms3ns "github.com/dms3-fs/go-dms3ns"
	pb "github.com/dms3-fs/go-dms3ns/pb"
	ci "github.com/dms3-p2p/go-p2p-crypto"
	peer "github.com/dms3-p2p/go-p2p-peer"
	routing "github.com/dms3-p2p/go-p2p-routing"
	base32 "github.com/whyrusleeping/base32"
)

const dms3nsPrefix = "/dms3ns/"

const PublishPutValTimeout = time.Minute
const DefaultRecordTTL = 24 * time.Hour

// Dms3NsPublisher is capable of publishing and resolving names to the DMS3FS
// routing system.
type Dms3NsPublisher struct {
	routing routing.ValueStore
	ds      ds.Datastore

	// Used to ensure we assign DMS3NS records *sequential* sequence numbers.
	mu sync.Mutex
}

// NewDms3NsPublisher constructs a publisher for the DMS3FS Routing name system.
func NewDms3NsPublisher(route routing.ValueStore, ds ds.Datastore) *Dms3NsPublisher {
	if ds == nil {
		panic("nil datastore")
	}
	return &Dms3NsPublisher{routing: route, ds: ds}
}

// Publish implements Publisher. Accepts a keypair and a value,
// and publishes it out to the routing system
func (p *Dms3NsPublisher) Publish(ctx context.Context, k ci.PrivKey, value path.Path) error {
	log.Debugf("Publish %s", value)
	return p.PublishWithEOL(ctx, k, value, time.Now().Add(DefaultRecordTTL))
}

func Dms3NsDsKey(id peer.ID) ds.Key {
	return ds.NewKey("/dms3ns/" + base32.RawStdEncoding.EncodeToString([]byte(id)))
}

// PublishedNames returns the latest DMS3NS records published by this node and
// their expiration times.
//
// This method will not search the routing system for records published by other
// nodes.
func (p *Dms3NsPublisher) ListPublished(ctx context.Context) (map[peer.ID]*pb.Dms3NsEntry, error) {
	query, err := p.ds.Query(dsquery.Query{
		Prefix: dms3nsPrefix,
	})
	if err != nil {
		return nil, err
	}
	defer query.Close()

	records := make(map[peer.ID]*pb.Dms3NsEntry)
	for {
		select {
		case result, ok := <-query.Next():
			if !ok {
				return records, nil
			}
			if result.Error != nil {
				return nil, result.Error
			}
			e := new(pb.Dms3NsEntry)
			if err := proto.Unmarshal(result.Value, e); err != nil {
				// Might as well return what we can.
				log.Error("found an invalid DMS3NS entry:", err)
				continue
			}
			if !strings.HasPrefix(result.Key, dms3nsPrefix) {
				log.Errorf("datastore query for keys with prefix %s returned a key: %s", dms3nsPrefix, result.Key)
				continue
			}
			k := result.Key[len(dms3nsPrefix):]
			pid, err := base32.RawStdEncoding.DecodeString(k)
			if err != nil {
				log.Errorf("dms3ns ds key invalid: %s", result.Key)
				continue
			}
			records[peer.ID(pid)] = e
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// GetPublished returns the record this node has published corresponding to the
// given peer ID.
//
// If `checkRouting` is true and we have no existing record, this method will
// check the routing system for any existing records.
func (p *Dms3NsPublisher) GetPublished(ctx context.Context, id peer.ID, checkRouting bool) (*pb.Dms3NsEntry, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	value, err := p.ds.Get(Dms3NsDsKey(id))
	switch err {
	case nil:
	case ds.ErrNotFound:
		if !checkRouting {
			return nil, nil
		}
		dms3nskey := dms3ns.RecordKey(id)
		value, err = p.routing.GetValue(ctx, dms3nskey)
		if err != nil {
			// Not found or other network issue. Can't really do
			// anything about this case.
			if err != routing.ErrNotFound {
				log.Debugf("error when determining the last published DMS3NS record for %s: %s", id, err)
			}

			return nil, nil
		}
	default:
		return nil, err
	}
	e := new(pb.Dms3NsEntry)
	if err := proto.Unmarshal(value, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (p *Dms3NsPublisher) updateRecord(ctx context.Context, k ci.PrivKey, value path.Path, eol time.Time) (*pb.Dms3NsEntry, error) {
	id, err := peer.IDFromPrivateKey(k)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// get previous records sequence number
	rec, err := p.GetPublished(ctx, id, true)
	if err != nil {
		return nil, err
	}

	seqno := rec.GetSequence() // returns 0 if rec is nil
	if rec != nil && value != path.Path(rec.GetValue()) {
		// Don't bother incrementing the sequence number unless the
		// value changes.
		seqno++
	}

	// Create record
	entry, err := dms3ns.Create(k, []byte(value), seqno, eol)
	if err != nil {
		return nil, err
	}

	// Set the TTL
	// TODO: Make this less hacky.
	ttl, ok := checkCtxTTL(ctx)
	if ok {
		entry.Ttl = proto.Uint64(uint64(ttl.Nanoseconds()))
	}

	data, err := proto.Marshal(entry)
	if err != nil {
		return nil, err
	}

	// Put the new record.
	if err := p.ds.Put(Dms3NsDsKey(id), data); err != nil {
		return nil, err
	}
	return entry, nil
}

// PublishWithEOL is a temporary stand in for the dms3ns records implementation
// see here for more details: https://github.com/dms3-fs/specs/tree/master/records
func (p *Dms3NsPublisher) PublishWithEOL(ctx context.Context, k ci.PrivKey, value path.Path, eol time.Time) error {
	record, err := p.updateRecord(ctx, k, value, eol)
	if err != nil {
		return err
	}

	return PutRecordToRouting(ctx, p.routing, k.GetPublic(), record)
}

// setting the TTL on published records is an experimental feature.
// as such, i'm using the context to wire it through to avoid changing too
// much code along the way.
func checkCtxTTL(ctx context.Context) (time.Duration, bool) {
	v := ctx.Value("dms3ns-publish-ttl")
	if v == nil {
		return 0, false
	}

	d, ok := v.(time.Duration)
	return d, ok
}

func PutRecordToRouting(ctx context.Context, r routing.ValueStore, k ci.PubKey, entry *pb.Dms3NsEntry) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errs := make(chan error, 2) // At most two errors (DMS3NS, and public key)

	if err := dms3ns.EmbedPublicKey(k, entry); err != nil {
		return err
	}

	id, err := peer.IDFromPublicKey(k)
	if err != nil {
		return err
	}

	go func() {
		errs <- PublishEntry(ctx, r, dms3ns.RecordKey(id), entry)
	}()

	// Publish the public key if a public key cannot be extracted from the ID
	// TODO: once v0.4.16 is widespread enough, we can stop doing this
	// and at that point we can even deprecate the /pk/ namespace in the dht
	//
	// NOTE: This check actually checks if the public key has been embedded
	// in the DMS3NS entry. This check is sufficient because we embed the
	// public key in the DMS3NS entry if it can't be extracted from the ID.
	if entry.PubKey != nil {
		go func() {
			errs <- PublishPublicKey(ctx, r, PkKeyForID(id), k)
		}()

		if err := waitOnErrChan(ctx, errs); err != nil {
			return err
		}
	}

	return waitOnErrChan(ctx, errs)
}

func waitOnErrChan(ctx context.Context, errs chan error) error {
	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func PublishPublicKey(ctx context.Context, r routing.ValueStore, k string, pubk ci.PubKey) error {
	log.Debugf("Storing pubkey at: %s", k)
	pkbytes, err := pubk.Bytes()
	if err != nil {
		return err
	}

	// Store associated public key
	timectx, cancel := context.WithTimeout(ctx, PublishPutValTimeout)
	defer cancel()
	return r.PutValue(timectx, k, pkbytes)
}

func PublishEntry(ctx context.Context, r routing.ValueStore, dms3nskey string, rec *pb.Dms3NsEntry) error {
	timectx, cancel := context.WithTimeout(ctx, PublishPutValTimeout)
	defer cancel()

	data, err := proto.Marshal(rec)
	if err != nil {
		return err
	}

	log.Debugf("Storing dms3ns entry at: %s", dms3nskey)
	// Store dms3ns entry at "/dms3ns/"+h(pubkey)
	return r.PutValue(timectx, dms3nskey, data)
}

// InitializeKeyspace sets the dms3ns record for the given key to
// point to an empty directory.
// TODO: this doesnt feel like it belongs here
func InitializeKeyspace(ctx context.Context, pub Publisher, pins pin.Pinner, key ci.PrivKey) error {
	emptyDir := ft.EmptyDirNode()

	// pin recursively because this might already be pinned
	// and doing a direct pin would throw an error in that case
	err := pins.Pin(ctx, emptyDir, true)
	if err != nil {
		return err
	}

	err = pins.Flush()
	if err != nil {
		return err
	}

	return pub.Publish(ctx, key, path.FromCid(emptyDir.Cid()))
}

// PkKeyForID returns the public key routing key for the given peer ID.
func PkKeyForID(id peer.ID) string {
	return "/pk/" + string(id)
}
