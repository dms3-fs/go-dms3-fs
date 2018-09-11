/*
Package core implements the Dms3FsNode object and related methods.

Packages underneath core/ provide a (relatively) stable, low-level API
to carry out most DMS3FS-related tasks.  For more details on the other
interfaces and how core/... fits into the bigger DMS3FS picture, see:

  $ godoc github.com/dms3-fs/go-dms3-fs
*/
package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	version "github.com/dms3-fs/go-dms3-fs"
	rp "github.com/dms3-fs/go-dms3-fs/exchange/reprovide"
	filestore "github.com/dms3-fs/go-dms3-fs/filestore"
	mount "github.com/dms3-fs/go-dms3-fs/fuse/mount"
	namesys "github.com/dms3-fs/go-dms3-fs/namesys"
	dms3nsrp "github.com/dms3-fs/go-dms3-fs/namesys/republisher"
	p2p "github.com/dms3-fs/go-dms3-fs/p2p"
	pin "github.com/dms3-fs/go-dms3-fs/pin"
	repo "github.com/dms3-fs/go-dms3-fs/repo"

	bitswap "github.com/dms3-fs/go-bitswap"
	bsnet "github.com/dms3-fs/go-bitswap/network"
	bserv "github.com/dms3-fs/go-blockservice"
	cid "github.com/dms3-fs/go-cid"
	ds "github.com/dms3-fs/go-datastore"
	bstore "github.com/dms3-fs/go-fs-blockstore"
	config "github.com/dms3-fs/go-fs-config"
	exchange "github.com/dms3-fs/go-fs-exchange-interface"
	nilrouting "github.com/dms3-fs/go-fs-routing/none"
	offroute "github.com/dms3-fs/go-fs-routing/offline"
	u "github.com/dms3-fs/go-fs-util"
	dms3ld "github.com/dms3-fs/go-ld-format"
	logging "github.com/dms3-fs/go-log"
	merkledag "github.com/dms3-fs/go-merkledag"
	mfs "github.com/dms3-fs/go-mfs"
	"github.com/dms3-fs/go-path/resolver"
	ft "github.com/dms3-fs/go-unixfs"
	goprocess "github.com/jbenet/goprocess"
	floodsub "github.com/dms3-p2p/go-floodsub"
	dms3p2p "github.com/dms3-p2p/go-p2p"
	circuit "github.com/dms3-p2p/go-p2p-circuit"
	connmgr "github.com/dms3-p2p/go-p2p-connmgr"
	ic "github.com/dms3-p2p/go-p2p-crypto"
	p2phost "github.com/dms3-p2p/go-p2p-host"
	ifconnmgr "github.com/dms3-p2p/go-p2p-interface-connmgr"
	dht "github.com/dms3-p2p/go-p2p-kad-dht"
	dhtopts "github.com/dms3-p2p/go-p2p-kad-dht/opts"
	metrics "github.com/dms3-p2p/go-p2p-metrics"
	peer "github.com/dms3-p2p/go-p2p-peer"
	pstore "github.com/dms3-p2p/go-p2p-peerstore"
	pnet "github.com/dms3-p2p/go-p2p-pnet"
	psrouter "github.com/dms3-p2p/go-p2p-pubsub-router"
	record "github.com/dms3-p2p/go-p2p-record"
	routing "github.com/dms3-p2p/go-p2p-routing"
	rhelpers "github.com/dms3-p2p/go-p2p-routing-helpers"
	discovery "github.com/dms3-p2p/go-p2p/p2p/discovery"
	p2pbhost "github.com/dms3-p2p/go-p2p/p2p/host/basic"
	rhost "github.com/dms3-p2p/go-p2p/p2p/host/routed"
	identify "github.com/dms3-p2p/go-p2p/p2p/protocol/identify"
	ping "github.com/dms3-p2p/go-p2p/p2p/protocol/ping"
	mafilter "github.com/dms3-p2p/go-maddr-filter"
	smux "github.com/dms3-p2p/go-stream-muxer"
	ma "github.com/dms3-mft/go-multiaddr"
	mplex "github.com/dms3-why/go-smux-multiplex"
	yamux "github.com/dms3-why/go-smux-yamux"
	mamask "github.com/whyrusleeping/multiaddr-filter"
)

const Dms3NsValidatorTag = "dms3ns"

const kReprovideFrequency = time.Hour * 12
const discoveryConnTimeout = time.Second * 30

var log = logging.Logger("core")

type mode int

const (
	// zero value is not a valid mode, must be explicitly set
	localMode mode = iota
	offlineMode
	onlineMode
)

func init() {
	identify.ClientVersion = "go-dms3-fs/" + version.CurrentVersionNumber + "/" + version.CurrentCommit
}

// Dms3FsNode is DMS3FS Core module. It represents an DMS3FS instance.
type Dms3FsNode struct {

	// Self
	Identity peer.ID // the local node's identity

	Repo repo.Repo

	// Local node
	Pinning         pin.Pinner // the pinning manager
	Mounts          Mounts     // current mount state, if any.
	PrivateKey      ic.PrivKey // the local node's private Key
	PNetFingerprint []byte     // fingerprint of private network

	// Services
	Peerstore       pstore.Peerstore     // storage for other Peer instances
	Blockstore      bstore.GCBlockstore  // the block store (lower level)
	Filestore       *filestore.Filestore // the filestore blockstore
	BaseBlocks      bstore.Blockstore    // the raw blockstore, no filestore wrapping
	GCLocker        bstore.GCLocker      // the locker used to protect the blockstore during gc
	Blocks          bserv.BlockService   // the block service, get/add blocks.
	DAG             dms3ld.DAGService      // the merkle dag service, get/add objects.
	Resolver        *resolver.Resolver   // the path resolution system
	Reporter        metrics.Reporter
	Discovery       discovery.Service
	FilesRoot       *mfs.Root
	RecordValidator record.Validator

	// Online
	PeerHost     p2phost.Host        // the network host (server+client)
	Bootstrapper io.Closer           // the periodic bootstrapper
	Routing      routing.Dms3FsRouting // the routing system. recommend dms3fs-dht
	Exchange     exchange.Interface  // the block exchange + strategy (bitswap)
	Namesys      namesys.NameSystem  // the name system, resolves paths to hashes
	Ping         *ping.PingService
	Reprovider   *rp.Reprovider // the value reprovider system
	Dms3NsRepub    *dms3nsrp.Republisher

	Floodsub *floodsub.PubSub
	PSRouter *psrouter.PubsubValueStore
	DHT      *dht.Dms3FsDHT
	P2P      *p2p.P2P

	proc goprocess.Process
	ctx  context.Context

	mode         mode
	localModeSet bool
}

// Mounts defines what the node's mount state is. This should
// perhaps be moved to the daemon or mount. It's here because
// it needs to be accessible across daemon requests.
type Mounts struct {
	Dms3Fs mount.Mount
	Dms3Ns mount.Mount
}

func (n *Dms3FsNode) startOnlineServices(ctx context.Context, routingOption RoutingOption, hostOption HostOption, do DiscoveryOption, pubsub, dms3nsps, mplex bool) error {
	if n.PeerHost != nil { // already online.
		return errors.New("node already online")
	}

	// load private key
	if err := n.LoadPrivateKey(); err != nil {
		return err
	}

	// get undialable addrs from config
	cfg, err := n.Repo.Config()
	if err != nil {
		return err
	}

	var dms3p2pOpts []dms3p2p.Option
	for _, s := range cfg.Swarm.AddrFilters {
		f, err := mamask.NewMask(s)
		if err != nil {
			return fmt.Errorf("incorrectly formatted address filter in config: %s", s)
		}
		dms3p2pOpts = append(dms3p2pOpts, dms3p2p.FilterAddresses(f))
	}

	if !cfg.Swarm.DisableBandwidthMetrics {
		// Set reporter
		n.Reporter = metrics.NewBandwidthCounter()
		dms3p2pOpts = append(dms3p2pOpts, dms3p2p.BandwidthReporter(n.Reporter))
	}

	swarmkey, err := n.Repo.SwarmKey()
	if err != nil {
		return err
	}

	if swarmkey != nil {
		protec, err := pnet.NewProtector(bytes.NewReader(swarmkey))
		if err != nil {
			return fmt.Errorf("failed to configure private network: %s", err)
		}
		n.PNetFingerprint = protec.Fingerprint()
		go func() {
			t := time.NewTicker(30 * time.Second)
			<-t.C // swallow one tick
			for {
				select {
				case <-t.C:
					if ph := n.PeerHost; ph != nil {
						if len(ph.Network().Peers()) == 0 {
							log.Warning("We are in private network and have no peers.")
							log.Warning("This might be configuration mistake.")
						}
					}
				case <-n.Process().Closing():
					t.Stop()
					return
				}
			}
		}()

		dms3p2pOpts = append(dms3p2pOpts, dms3p2p.PrivateNetwork(protec))
	}

	addrsFactory, err := makeAddrsFactory(cfg.Addresses)
	if err != nil {
		return err
	}
	if !cfg.Swarm.DisableRelay {
		addrsFactory = composeAddrsFactory(addrsFactory, filterRelayAddrs)
	}
	dms3p2pOpts = append(dms3p2pOpts, dms3p2p.AddrsFactory(addrsFactory))

	connm, err := constructConnMgr(cfg.Swarm.ConnMgr)
	if err != nil {
		return err
	}
	dms3p2pOpts = append(dms3p2pOpts, dms3p2p.ConnectionManager(connm))

	dms3p2pOpts = append(dms3p2pOpts, makeSmuxTransportOption(mplex))

	if !cfg.Swarm.DisableNatPortMap {
		dms3p2pOpts = append(dms3p2pOpts, dms3p2p.NATPortMap())
	}
	if !cfg.Swarm.DisableRelay {
		var opts []circuit.RelayOpt
		if cfg.Swarm.EnableRelayHop {
			opts = append(opts, circuit.OptHop)
		}
		dms3p2pOpts = append(dms3p2pOpts, dms3p2p.EnableRelay(opts...))
	}

	peerhost, err := hostOption(ctx, n.Identity, n.Peerstore, dms3p2pOpts...)

	if err != nil {
		return err
	}

	if err := n.startOnlineServicesWithHost(ctx, peerhost, routingOption, pubsub, dms3nsps); err != nil {
		return err
	}

	// Ok, now we're ready to listen.
	if err := startListening(n.PeerHost, cfg); err != nil {
		return err
	}

	n.P2P = p2p.NewP2P(n.Identity, n.PeerHost, n.Peerstore)

	// setup local discovery
	if do != nil {
		service, err := do(ctx, n.PeerHost)
		if err != nil {
			log.Error("mdns error: ", err)
		} else {
			service.RegisterNotifee(n)
			n.Discovery = service
		}
	}

	return n.Bootstrap(DefaultBootstrapConfig)
}

func constructConnMgr(cfg config.ConnMgr) (ifconnmgr.ConnManager, error) {
	switch cfg.Type {
	case "":
		// 'default' value is the basic connection manager
		return connmgr.NewConnManager(config.DefaultConnMgrLowWater, config.DefaultConnMgrHighWater, config.DefaultConnMgrGracePeriod), nil
	case "none":
		return nil, nil
	case "basic":
		grace, err := time.ParseDuration(cfg.GracePeriod)
		if err != nil {
			return nil, fmt.Errorf("parsing Swarm.ConnMgr.GracePeriod: %s", err)
		}

		return connmgr.NewConnManager(cfg.LowWater, cfg.HighWater, grace), nil
	default:
		return nil, fmt.Errorf("unrecognized ConnMgr.Type: %q", cfg.Type)
	}
}

func (n *Dms3FsNode) startLateOnlineServices(ctx context.Context) error {
	cfg, err := n.Repo.Config()
	if err != nil {
		return err
	}

	var keyProvider rp.KeyChanFunc

	switch cfg.Reprovider.Strategy {
	case "all":
		fallthrough
	case "":
		keyProvider = rp.NewBlockstoreProvider(n.Blockstore)
	case "roots":
		keyProvider = rp.NewPinnedProvider(n.Pinning, n.DAG, true)
	case "pinned":
		keyProvider = rp.NewPinnedProvider(n.Pinning, n.DAG, false)
	default:
		return fmt.Errorf("unknown reprovider strategy '%s'", cfg.Reprovider.Strategy)
	}
	n.Reprovider = rp.NewReprovider(ctx, n.Routing, keyProvider)

	reproviderInterval := kReprovideFrequency
	if cfg.Reprovider.Interval != "" {
		dur, err := time.ParseDuration(cfg.Reprovider.Interval)
		if err != nil {
			return err
		}

		reproviderInterval = dur
	}

	go n.Reprovider.Run(reproviderInterval)

	return nil
}

func makeAddrsFactory(cfg config.Addresses) (p2pbhost.AddrsFactory, error) {
	var annAddrs []ma.Multiaddr
	for _, addr := range cfg.Announce {
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		annAddrs = append(annAddrs, maddr)
	}

	filters := mafilter.NewFilters()
	noAnnAddrs := map[string]bool{}
	for _, addr := range cfg.NoAnnounce {
		f, err := mamask.NewMask(addr)
		if err == nil {
			filters.AddDialFilter(f)
			continue
		}
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		noAnnAddrs[maddr.String()] = true
	}

	return func(allAddrs []ma.Multiaddr) []ma.Multiaddr {
		var addrs []ma.Multiaddr
		if len(annAddrs) > 0 {
			addrs = annAddrs
		} else {
			addrs = allAddrs
		}

		var out []ma.Multiaddr
		for _, maddr := range addrs {
			// check for exact matches
			ok, _ := noAnnAddrs[maddr.String()]
			// check for /ipcidr matches
			if !ok && !filters.AddrBlocked(maddr) {
				out = append(out, maddr)
			}
		}
		return out
	}, nil
}

func makeSmuxTransportOption(mplexExp bool) dms3p2p.Option {
	const yamuxID = "/yamux/1.0.0"
	const mplexID = "/mplex/6.7.0"

	ymxtpt := &yamux.Transport{
		AcceptBacklog:          512,
		ConnectionWriteTimeout: time.Second * 10,
		KeepAliveInterval:      time.Second * 30,
		EnableKeepAlive:        true,
		MaxStreamWindowSize:    uint32(1024 * 512),
		LogOutput:              ioutil.Discard,
	}

	if os.Getenv("YAMUX_DEBUG") != "" {
		ymxtpt.LogOutput = os.Stderr
	}

	muxers := map[string]smux.Transport{yamuxID: ymxtpt}
	if mplexExp {
		muxers[mplexID] = mplex.DefaultTransport
	}

	// Allow muxer preference order overriding
	order := []string{yamuxID, mplexID}
	if prefs := os.Getenv("LIBP2P_MUX_PREFS"); prefs != "" {
		order = strings.Fields(prefs)
	}

	opts := make([]dms3p2p.Option, 0, len(order))
	for _, id := range order {
		tpt, ok := muxers[id]
		if !ok {
			log.Warning("unknown or duplicate muxer in LIBP2P_MUX_PREFS: %s", id)
			continue
		}
		delete(muxers, id)
		opts = append(opts, dms3p2p.Muxer(id, tpt))
	}

	return dms3p2p.ChainOptions(opts...)
}

func setupDiscoveryOption(d config.Discovery) DiscoveryOption {
	if d.MDNS.Enabled {
		return func(ctx context.Context, h p2phost.Host) (discovery.Service, error) {
			if d.MDNS.Interval == 0 {
				d.MDNS.Interval = 5
			}
			return discovery.NewMdnsService(ctx, h, time.Duration(d.MDNS.Interval)*time.Second, discovery.ServiceTag)
		}
	}
	return nil
}

// HandlePeerFound attempts to connect to peer from `PeerInfo`, if it fails
// logs a warning log.
func (n *Dms3FsNode) HandlePeerFound(p pstore.PeerInfo) {
	log.Warning("trying peer info: ", p)
	ctx, cancel := context.WithTimeout(n.Context(), discoveryConnTimeout)
	defer cancel()
	if err := n.PeerHost.Connect(ctx, p); err != nil {
		log.Warning("Failed to connect to peer found by discovery: ", err)
	}
}

// startOnlineServicesWithHost  is the set of services which need to be
// initialized with the host and _before_ we start listening.
func (n *Dms3FsNode) startOnlineServicesWithHost(ctx context.Context, host p2phost.Host, routingOption RoutingOption, pubsub bool, dms3nsps bool) error {
	// setup diagnostics service
	n.Ping = ping.NewPingService(host)

	if pubsub || dms3nsps {
		cfg, err := n.Repo.Config()
		if err != nil {
			return err
		}

		var service *floodsub.PubSub

		switch cfg.Pubsub.Router {
		case "":
			fallthrough
		case "floodsub":
			service, err = floodsub.NewFloodSub(ctx, host)

		case "gossipsub":
			service, err = floodsub.NewGossipSub(ctx, host)

		default:
			err = fmt.Errorf("Unknown pubsub router %s", cfg.Pubsub.Router)
		}

		if err != nil {
			return err
		}
		n.Floodsub = service
	}

	// setup routing service
	r, err := routingOption(ctx, host, n.Repo.Datastore(), n.RecordValidator)
	if err != nil {
		return err
	}
	n.Routing = r

	// TODO: I'm not a fan of type assertions like this but the
	// `RoutingOption` system doesn't currently provide access to the
	// Dms3FsNode.
	//
	// Ideally, we'd do something like:
	//
	// 1. Add some fancy method to introspect into tiered routers to extract
	//    things like the pubsub router or the DHT (complicated, messy,
	//    probably not worth it).
	// 2. Pass the Dms3FsNode into the RoutingOption (would also remove the
	//    PSRouter case below.
	// 3. Introduce some kind of service manager? (my personal favorite but
	//    that requires a fair amount of work).
	if dht, ok := r.(*dht.Dms3FsDHT); ok {
		n.DHT = dht
	}

	if dms3nsps {
		n.PSRouter = psrouter.NewPubsubValueStore(
			ctx,
			host,
			n.Routing,
			n.Floodsub,
			n.RecordValidator,
		)
		n.Routing = rhelpers.Tiered{
			// Always check pubsub first.
			&rhelpers.Compose{
				ValueStore: &rhelpers.LimitedValueStore{
					ValueStore: n.PSRouter,
					Namespaces: []string{"dms3ns"},
				},
			},
			n.Routing,
		}
	}

	// Wrap standard peer host with routing system to allow unknown peer lookups
	n.PeerHost = rhost.Wrap(host, n.Routing)

	// setup exchange service
	bitswapNetwork := bsnet.NewFromDms3FsHost(n.PeerHost, n.Routing)
	n.Exchange = bitswap.New(ctx, bitswapNetwork, n.Blockstore)

	size, err := n.getCacheSize()
	if err != nil {
		return err
	}

	// setup name system
	n.Namesys = namesys.NewNameSystem(n.Routing, n.Repo.Datastore(), size)

	// setup dms3ns republishing
	return n.setupDms3NsRepublisher()
}

// getCacheSize returns cache life and cache size
func (n *Dms3FsNode) getCacheSize() (int, error) {
	cfg, err := n.Repo.Config()
	if err != nil {
		return 0, err
	}

	cs := cfg.Dms3Ns.ResolveCacheSize
	if cs == 0 {
		cs = 128
	}
	if cs < 0 {
		return 0, fmt.Errorf("cannot specify negative resolve cache size")
	}
	return cs, nil
}

func (n *Dms3FsNode) setupDms3NsRepublisher() error {
	cfg, err := n.Repo.Config()
	if err != nil {
		return err
	}

	n.Dms3NsRepub = dms3nsrp.NewRepublisher(n.Namesys, n.Repo.Datastore(), n.PrivateKey, n.Repo.Keystore())

	if cfg.Dms3Ns.RepublishPeriod != "" {
		d, err := time.ParseDuration(cfg.Dms3Ns.RepublishPeriod)
		if err != nil {
			return fmt.Errorf("failure to parse config setting DMS3NS.RepublishPeriod: %s", err)
		}

		if !u.Debug && (d < time.Minute || d > (time.Hour*24)) {
			return fmt.Errorf("config setting DMS3NS.RepublishPeriod is not between 1min and 1day: %s", d)
		}

		n.Dms3NsRepub.Interval = d
	}

	if cfg.Dms3Ns.RecordLifetime != "" {
		d, err := time.ParseDuration(cfg.Dms3Ns.RepublishPeriod)
		if err != nil {
			return fmt.Errorf("failure to parse config setting DMS3NS.RecordLifetime: %s", err)
		}

		n.Dms3NsRepub.RecordLifetime = d
	}

	n.Process().Go(n.Dms3NsRepub.Run)

	return nil
}

// Process returns the Process object
func (n *Dms3FsNode) Process() goprocess.Process {
	return n.proc
}

// Close calls Close() on the Process object
func (n *Dms3FsNode) Close() error {
	return n.proc.Close()
}

// Context returns the Dms3FsNode context
func (n *Dms3FsNode) Context() context.Context {
	if n.ctx == nil {
		n.ctx = context.TODO()
	}
	return n.ctx
}

// teardown closes owned children. If any errors occur, this function returns
// the first error.
func (n *Dms3FsNode) teardown() error {
	log.Debug("core is shutting down...")
	// owned objects are closed in this teardown to ensure that they're closed
	// regardless of which constructor was used to add them to the node.
	var closers []io.Closer

	// NOTE: The order that objects are added(closed) matters, if an object
	// needs to use another during its shutdown/cleanup process, it should be
	// closed before that other object

	if n.FilesRoot != nil {
		closers = append(closers, n.FilesRoot)
	}

	if n.Exchange != nil {
		closers = append(closers, n.Exchange)
	}

	if n.Mounts.Dms3Fs != nil && !n.Mounts.Dms3Fs.IsActive() {
		closers = append(closers, mount.Closer(n.Mounts.Dms3Fs))
	}
	if n.Mounts.Dms3Ns != nil && !n.Mounts.Dms3Ns.IsActive() {
		closers = append(closers, mount.Closer(n.Mounts.Dms3Ns))
	}

	if n.DHT != nil {
		closers = append(closers, n.DHT.Process())
	}

	if n.Blocks != nil {
		closers = append(closers, n.Blocks)
	}

	if n.Bootstrapper != nil {
		closers = append(closers, n.Bootstrapper)
	}

	if n.PeerHost != nil {
		closers = append(closers, n.PeerHost)
	}

	// Repo closed last, most things need to preserve state here
	closers = append(closers, n.Repo)

	var errs []error
	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// OnlineMode returns whether or not the Dms3FsNode is in OnlineMode.
func (n *Dms3FsNode) OnlineMode() bool {
	return n.mode == onlineMode
}

// SetLocal will set the Dms3FsNode to local mode
func (n *Dms3FsNode) SetLocal(isLocal bool) {
	if isLocal {
		n.mode = localMode
	}
	n.localModeSet = true
}

// LocalMode returns whether or not the Dms3FsNode is in LocalMode
func (n *Dms3FsNode) LocalMode() bool {
	if !n.localModeSet {
		// programmer error should not happen
		panic("local mode not set")
	}
	return n.mode == localMode
}

// Bootstrap will set and call the Dms3FsNodes bootstrap function.
func (n *Dms3FsNode) Bootstrap(cfg BootstrapConfig) error {
	// TODO what should return value be when in offlineMode?
	if n.Routing == nil {
		return nil
	}

	if n.Bootstrapper != nil {
		n.Bootstrapper.Close() // stop previous bootstrap process.
	}

	// if the caller did not specify a bootstrap peer function, get the
	// freshest bootstrap peers from config. this responds to live changes.
	if cfg.BootstrapPeers == nil {
		cfg.BootstrapPeers = func() []pstore.PeerInfo {
			ps, err := n.loadBootstrapPeers()
			if err != nil {
				log.Warning("failed to parse bootstrap peers from config")
				return nil
			}
			return ps
		}
	}

	var err error
	n.Bootstrapper, err = Bootstrap(n, cfg)
	return err
}

func (n *Dms3FsNode) loadID() error {
	if n.Identity != "" {
		return errors.New("identity already loaded")
	}

	cfg, err := n.Repo.Config()
	if err != nil {
		return err
	}

	cid := cfg.Identity.PeerID
	if cid == "" {
		return errors.New("identity was not set in config (was 'dms3fs init' run?)")
	}
	if len(cid) == 0 {
		return errors.New("no peer ID in config! (was 'dms3fs init' run?)")
	}

	id, err := peer.IDB58Decode(cid)
	if err != nil {
		return fmt.Errorf("peer ID invalid: %s", err)
	}

	n.Identity = id
	return nil
}

// GetKey will return a key from the Keystore with name `name`.
func (n *Dms3FsNode) GetKey(name string) (ic.PrivKey, error) {
	if name == "self" {
		return n.PrivateKey, nil
	} else {
		return n.Repo.Keystore().Get(name)
	}
}

func (n *Dms3FsNode) LoadPrivateKey() error {
	if n.Identity == "" || n.Peerstore == nil {
		return errors.New("loaded private key out of order")
	}

	if n.PrivateKey != nil {
		return errors.New("private key already loaded")
	}

	cfg, err := n.Repo.Config()
	if err != nil {
		return err
	}

	sk, err := loadPrivateKey(&cfg.Identity, n.Identity)
	if err != nil {
		return err
	}

	n.PrivateKey = sk
	n.Peerstore.AddPrivKey(n.Identity, n.PrivateKey)
	n.Peerstore.AddPubKey(n.Identity, sk.GetPublic())
	return nil
}

func (n *Dms3FsNode) loadBootstrapPeers() ([]pstore.PeerInfo, error) {
	cfg, err := n.Repo.Config()
	if err != nil {
		return nil, err
	}

	parsed, err := cfg.BootstrapPeers()
	if err != nil {
		return nil, err
	}
	return toPeerInfos(parsed), nil
}

func (n *Dms3FsNode) loadFilesRoot() error {
	dsk := ds.NewKey("/local/filesroot")
	pf := func(ctx context.Context, c *cid.Cid) error {
		return n.Repo.Datastore().Put(dsk, c.Bytes())
	}

	var nd *merkledag.ProtoNode
	val, err := n.Repo.Datastore().Get(dsk)

	switch {
	case err == ds.ErrNotFound || val == nil:
		nd = ft.EmptyDirNode()
		err := n.DAG.Add(n.Context(), nd)
		if err != nil {
			return fmt.Errorf("failure writing to dagstore: %s", err)
		}
	case err == nil:
		c, err := cid.Cast(val)
		if err != nil {
			return err
		}

		rnd, err := n.DAG.Get(n.Context(), c)
		if err != nil {
			return fmt.Errorf("error loading filesroot from DAG: %s", err)
		}

		pbnd, ok := rnd.(*merkledag.ProtoNode)
		if !ok {
			return merkledag.ErrNotProtobuf
		}

		nd = pbnd
	default:
		return err
	}

	mr, err := mfs.NewRoot(n.Context(), n.DAG, nd, pf)
	if err != nil {
		return err
	}

	n.FilesRoot = mr
	return nil
}

// SetupOfflineRouting instantiates a routing system in offline mode. This is
// primarily used for offline dms3ns modifications.
func (n *Dms3FsNode) SetupOfflineRouting() error {
	if n.Routing != nil {
		// Routing was already set up
		return nil
	}

	// TODO: move this somewhere else.
	err := n.LoadPrivateKey()
	if err != nil {
		return err
	}

	n.Routing = offroute.NewOfflineRouter(n.Repo.Datastore(), n.RecordValidator)

	size, err := n.getCacheSize()
	if err != nil {
		return err
	}

	n.Namesys = namesys.NewNameSystem(n.Routing, n.Repo.Datastore(), size)

	return nil
}

func loadPrivateKey(cfg *config.Identity, id peer.ID) (ic.PrivKey, error) {
	sk, err := cfg.DecodePrivateKey("passphrase todo!")
	if err != nil {
		return nil, err
	}

	id2, err := peer.IDFromPrivateKey(sk)
	if err != nil {
		return nil, err
	}

	if id2 != id {
		return nil, fmt.Errorf("private key in config does not match id: %s != %s", id, id2)
	}

	return sk, nil
}

func listenAddresses(cfg *config.Config) ([]ma.Multiaddr, error) {
	var listen []ma.Multiaddr
	for _, addr := range cfg.Addresses.Swarm {
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, fmt.Errorf("failure to parse config.Addresses.Swarm: %s", cfg.Addresses.Swarm)
		}
		listen = append(listen, maddr)
	}

	return listen, nil
}

type ConstructPeerHostOpts struct {
	AddrsFactory      p2pbhost.AddrsFactory
	DisableNatPortMap bool
	DisableRelay      bool
	EnableRelayHop    bool
	ConnectionManager ifconnmgr.ConnManager
}

type HostOption func(ctx context.Context, id peer.ID, ps pstore.Peerstore, options ...dms3p2p.Option) (p2phost.Host, error)

var DefaultHostOption HostOption = constructPeerHost

// isolates the complex initialization steps
func constructPeerHost(ctx context.Context, id peer.ID, ps pstore.Peerstore, options ...dms3p2p.Option) (p2phost.Host, error) {
	pkey := ps.PrivKey(id)
	if pkey == nil {
		return nil, fmt.Errorf("missing private key for node ID: %s", id.Pretty())
	}
	options = append([]dms3p2p.Option{dms3p2p.Identity(pkey), dms3p2p.Peerstore(ps)}, options...)
	return dms3p2p.New(ctx, options...)
}

func filterRelayAddrs(addrs []ma.Multiaddr) []ma.Multiaddr {
	var raddrs []ma.Multiaddr
	for _, addr := range addrs {
		_, err := addr.ValueForProtocol(circuit.P_CIRCUIT)
		if err == nil {
			continue
		}
		raddrs = append(raddrs, addr)
	}
	return raddrs
}

func composeAddrsFactory(f, g p2pbhost.AddrsFactory) p2pbhost.AddrsFactory {
	return func(addrs []ma.Multiaddr) []ma.Multiaddr {
		return f(g(addrs))
	}
}

// startListening on the network addresses
func startListening(host p2phost.Host, cfg *config.Config) error {
	listenAddrs, err := listenAddresses(cfg)
	if err != nil {
		return err
	}

	// Actually start listening:
	if err := host.Network().Listen(listenAddrs...); err != nil {
		return err
	}

	// list out our addresses
	addrs, err := host.Network().InterfaceListenAddresses()
	if err != nil {
		return err
	}
	log.Infof("Swarm listening at: %s", addrs)
	return nil
}

func constructDHTRouting(ctx context.Context, host p2phost.Host, dstore ds.Batching, validator record.Validator) (routing.Dms3FsRouting, error) {
	return dht.New(
		ctx, host,
		dhtopts.Datastore(dstore),
		dhtopts.Validator(validator),
	)
}

func constructClientDHTRouting(ctx context.Context, host p2phost.Host, dstore ds.Batching, validator record.Validator) (routing.Dms3FsRouting, error) {
	return dht.New(
		ctx, host,
		dhtopts.Client(true),
		dhtopts.Datastore(dstore),
		dhtopts.Validator(validator),
	)
}

type RoutingOption func(context.Context, p2phost.Host, ds.Batching, record.Validator) (routing.Dms3FsRouting, error)

type DiscoveryOption func(context.Context, p2phost.Host) (discovery.Service, error)

var DHTOption RoutingOption = constructDHTRouting
var DHTClientOption RoutingOption = constructClientDHTRouting
var NilRouterOption RoutingOption = nilrouting.ConstructNilRouting
