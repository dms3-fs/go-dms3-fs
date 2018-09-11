package main

import (
	"errors"
	_ "expvar"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sort"
	"sync"

    utilmain "github.com/dms3-fs/go-dms3-fs/cmd/dms3fs/util"
    oldcmds "github.com/dms3-fs/go-dms3-fs/commands"
    "github.com/dms3-fs/go-dms3-fs/core"
    commands "github.com/dms3-fs/go-dms3-fs/core/commands"
    corehttp "github.com/dms3-fs/go-dms3-fs/core/corehttp"
    corerepo "github.com/dms3-fs/go-dms3-fs/core/corerepo"
    nodeMount "github.com/dms3-fs/go-dms3-fs/fuse/node"
    fsrepo "github.com/dms3-fs/go-dms3-fs/repo/fsrepo"
    migrate "github.com/dms3-fs/go-dms3-fs/repo/fsrepo/migrations"

    "github.com/gxed/client_golang/prometheus"
    "github.com/dms3-fs/go-fs-cmdkit"
    cmds "github.com/dms3-fs/go-fs-cmds"
    mprome "github.com/dms3-fs/go-metrics-prometheus"
    ma "github.com/dms3-mft/go-multiaddr"
    "github.com/dms3-mft/go-multiaddr-net"
)

const (
	adjustFDLimitKwd          = "manage-fdlimit"
	enableGCKwd               = "enable-gc"
	initOptionKwd             = "init"
	initProfileOptionKwd      = "init-profile"
	dms3fsMountKwd              = "mount-dms3fs"
	dms3nsMountKwd              = "mount-dms3ns"
	migrateKwd                = "migrate"
	mountKwd                  = "mount"
	offlineKwd                = "offline"
	routingOptionKwd          = "routing"
	routingOptionSupernodeKwd = "supernode"
	routingOptionDHTClientKwd = "dhtclient"
	routingOptionDHTKwd       = "dht"
	routingOptionNoneKwd      = "none"
	routingOptionDefaultKwd   = "default"
	unencryptTransportKwd     = "disable-transport-encryption"
	unrestrictedApiAccessKwd  = "unrestricted-api"
	writableKwd               = "writable"
	enableFloodSubKwd         = "enable-pubsub-experiment"
	enableDMS3NSPubSubKwd       = "enable-namesys-pubsub"
	enableMultiplexKwd        = "enable-mplex-experiment"
	// apiAddrKwd    = "address-api"
	// swarmAddrKwd  = "address-swarm"
)

var daemonCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Run a network-connected DMS3FS node.",
		ShortDescription: `
'dms3fs daemon' runs a persistent dms3fs daemon that can serve commands
over the network. Most applications that use DMS3FS will do so by
communicating with a daemon over the HTTP API. While the daemon is
running, calls to 'dms3fs' commands will be sent over the network to
the daemon.
`,
		LongDescription: `
The daemon will start listening on ports on the network, which are
documented in (and can be modified through) 'dms3fs config Addresses'.
For example, to change the 'Gateway' port:

  dms3fs config Addresses.Gateway /ip4/127.0.0.1/tcp/8182

The API address can be changed the same way:

  dms3fs config Addresses.API /ip4/127.0.0.1/tcp/5102

Make sure to restart the daemon after changing addresses.

By default, the gateway is only accessible locally. To expose it to
other computers in the network, use 0.0.0.0 as the ip address:

  dms3fs config Addresses.Gateway /ip4/0.0.0.0/tcp/8180

Be careful if you expose the API. It is a security risk, as anyone could
control your node remotely. If you need to control the node remotely,
make sure to protect the port as you would other services or database
(firewall, authenticated proxy, etc).

HTTP Headers

dms3fs supports passing arbitrary headers to the API and Gateway. You can
do this by setting headers on the API.HTTPHeaders and Gateway.HTTPHeaders
keys:

  dms3fs config --json API.HTTPHeaders.X-Special-Header '["so special :)"]'
  dms3fs config --json Gateway.HTTPHeaders.X-Special-Header '["so special :)"]'

Note that the value of the keys is an _array_ of strings. This is because
headers can have more than one value, and it is convenient to pass through
to other libraries.

CORS Headers (for API)

You can setup CORS headers the same way:

  dms3fs config --json API.HTTPHeaders.Access-Control-Allow-Origin '["example.com"]'
  dms3fs config --json API.HTTPHeaders.Access-Control-Allow-Methods '["PUT", "GET", "POST"]'
  dms3fs config --json API.HTTPHeaders.Access-Control-Allow-Credentials '["true"]'

Shutdown

To shutdown the daemon, send a SIGINT signal to it (e.g. by pressing 'Ctrl-C')
or send a SIGTERM signal to it (e.g. with 'kill'). It may take a while for the
daemon to shutdown gracefully, but it can be killed forcibly by sending a
second signal.

DMS3FS_PATH environment variable

dms3fs uses a repository in the local file system. By default, the repo is
located at ~/.dms3-fs. To change the repo location, set the $DMS3FS_PATH
environment variable:

  export DMS3FS_PATH=/path/to/dms3fsrepo

Routing

DMS3FS by default will use a DHT for content routing. There is a highly
experimental alternative that operates the DHT in a 'client only' mode that
can be enabled by running the daemon as:

  dms3fs daemon --routing=dhtclient

This will later be transitioned into a config option once it gets out of the
'experimental' stage.

DEPRECATION NOTICE

Previously, dms3fs used an environment variable as seen below:

  export API_ORIGIN="http://localhost:8888/"

This is deprecated. It is still honored in this version, but will be removed
in a future version, along with this notice. Please move to setting the HTTP
Headers.
`,
	},

	Options: []cmdkit.Option{
		cmdkit.BoolOption(initOptionKwd, "Initialize dms3fs with default settings if not already initialized"),
		cmdkit.StringOption(initProfileOptionKwd, "Configuration profiles to apply for --init. See dms3fs init --help for more"),
		cmdkit.StringOption(routingOptionKwd, "Overrides the routing option").WithDefault(routingOptionDefaultKwd),
		cmdkit.BoolOption(mountKwd, "Mounts DMS3FS to the filesystem"),
		cmdkit.BoolOption(writableKwd, "Enable writing objects (with POST, PUT and DELETE)"),
		cmdkit.StringOption(dms3fsMountKwd, "Path to the mountpoint for DMS3FS (if using --mount). Defaults to config setting."),
		cmdkit.StringOption(dms3nsMountKwd, "Path to the mountpoint for DMS3NS (if using --mount). Defaults to config setting."),
		cmdkit.BoolOption(unrestrictedApiAccessKwd, "Allow API access to unlisted hashes"),
		cmdkit.BoolOption(unencryptTransportKwd, "Disable transport encryption (for debugging protocols)"),
		cmdkit.BoolOption(enableGCKwd, "Enable automatic periodic repo garbage collection"),
		cmdkit.BoolOption(adjustFDLimitKwd, "Check and raise file descriptor limits if needed").WithDefault(true),
		cmdkit.BoolOption(offlineKwd, "Run offline. Do not connect to the rest of the network but provide local API."),
		cmdkit.BoolOption(migrateKwd, "If true, assume yes at the migrate prompt. If false, assume no."),
		cmdkit.BoolOption(enableFloodSubKwd, "Instantiate the dms3fs daemon with the experimental pubsub feature enabled."),
		cmdkit.BoolOption(enableDMS3NSPubSubKwd, "Enable DMS3NS record distribution through pubsub; enables pubsub."),
		cmdkit.BoolOption(enableMultiplexKwd, "Add the experimental 'go-multiplex' stream muxer to dms3p2p on construction.").WithDefault(true),

		// TODO: add way to override addresses. tricky part: updating the config if also --init.
		// cmdkit.StringOption(apiAddrKwd, "Address for the daemon rpc API (overrides config)"),
		// cmdkit.StringOption(swarmAddrKwd, "Address for the swarm socket (overrides config)"),
	},
	Subcommands: map[string]*cmds.Command{},
	Run:         daemonFunc,
}

// defaultMux tells mux to serve path using the default muxer. This is
// mostly useful to hook up things that register in the default muxer,
// and don't provide a convenient http.Handler entry point, such as
// expvar and http/pprof.
func defaultMux(path string) corehttp.ServeOption {
	return func(node *core.Dms3FsNode, _ net.Listener, mux *http.ServeMux) (*http.ServeMux, error) {
		mux.Handle(path, http.DefaultServeMux)
		return mux, nil
	}
}

func daemonFunc(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) {
	// Inject metrics before we do anything
	err := mprome.Inject()
	if err != nil {
		log.Errorf("Injecting prometheus handler for metrics failed with message: %s\n", err.Error())
	}

	// let the user know we're going.
	fmt.Printf("Initializing daemon...\n")

	managefd, _ := req.Options[adjustFDLimitKwd].(bool)
	if managefd {
		if err := utilmain.ManageFdLimit(); err != nil {
			log.Errorf("setting file descriptor limit: %s", err)
		}
	}

	cctx := env.(*oldcmds.Context)

	go func() {
		<-req.Context.Done()
		fmt.Println("Received interrupt signal, shutting down...")
		fmt.Println("(Hit ctrl-c again to force-shutdown the daemon.)")
	}()

	// check transport encryption flag.
	unencrypted, _ := req.Options[unencryptTransportKwd].(bool)
	if unencrypted {
		log.Warningf(`Running with --%s: All connections are UNENCRYPTED.
		You will not be able to connect to regular encrypted networks.`, unencryptTransportKwd)
	}

	// first, whether user has provided the initialization flag. we may be
	// running in an uninitialized state.
	initialize, _ := req.Options[initOptionKwd].(bool)
	if initialize {

		cfg := cctx.ConfigRoot
		if !fsrepo.IsInitialized(cfg) {
			profiles, _ := req.Options[initProfileOptionKwd].(string)

			err := initWithDefaults(os.Stdout, cfg, profiles)
			if err != nil {
				re.SetError(err, cmdkit.ErrNormal)
				return
			}
		}
	}

	// acquire the repo lock _before_ constructing a node. we need to make
	// sure we are permitted to access the resources (datastore, etc.)
	repo, err := fsrepo.Open(cctx.ConfigRoot)
	switch err {
	default:
		re.SetError(err, cmdkit.ErrNormal)
		return
	case fsrepo.ErrNeedMigration:
		domigrate, found := req.Options[migrateKwd].(bool)
		fmt.Println("Found outdated fs-repo, migrations need to be run.")

		if !found {
			domigrate = YesNoPrompt("Run migrations now? [y/N]")
		}

		if !domigrate {
			fmt.Println("Not running migrations of fs-repo now.")
			fmt.Println("Please get fs-repo-migrations from https://dist.dms3.io")
			re.SetError(fmt.Errorf("fs-repo requires migration"), cmdkit.ErrNormal)
			return
		}

		err = migrate.RunMigration(fsrepo.RepoVersion)
		if err != nil {
			fmt.Println("The migrations of fs-repo failed:")
			fmt.Printf("  %s\n", err)
			fmt.Println("If you think this is a bug, please file an issue and include this whole log output.")
			fmt.Println("  https://github.com/dms3-fs/fs-repo-migrations")
			re.SetError(err, cmdkit.ErrNormal)
			return
		}

		repo, err = fsrepo.Open(cctx.ConfigRoot)
		if err != nil {
			re.SetError(err, cmdkit.ErrNormal)
			return
		}
	case nil:
		break
	}

	cfg, err := cctx.GetConfig()
	if err != nil {
		re.SetError(err, cmdkit.ErrNormal)
		return
	}

	offline, _ := req.Options[offlineKwd].(bool)
	dms3nsps, _ := req.Options[enableDMS3NSPubSubKwd].(bool)
	pubsub, _ := req.Options[enableFloodSubKwd].(bool)
	mplex, _ := req.Options[enableMultiplexKwd].(bool)

	// Start assembling node config
	ncfg := &core.BuildCfg{
		Repo:      repo,
		Permanent: true, // It is temporary way to signify that node is permanent
		Online:    !offline,
		DisableEncryptedConnections: unencrypted,
		ExtraOpts: map[string]bool{
			"pubsub": pubsub,
			"dms3nsps": dms3nsps,
			"mplex":  mplex,
		},
		//TODO(Kubuxu): refactor Online vs Offline by adding Permanent vs Ephemeral
	}

	routingOption, _ := req.Options[routingOptionKwd].(string)
	if routingOption == routingOptionDefaultKwd {
		cfg, err := repo.Config()
		if err != nil {
			re.SetError(err, cmdkit.ErrNormal)
			return
		}

		routingOption = cfg.Routing.Type
		if routingOption == "" {
			routingOption = routingOptionDHTKwd
		}
	}
	switch routingOption {
	case routingOptionSupernodeKwd:
		re.SetError(errors.New("supernode routing was never fully implemented and has been removed"), cmdkit.ErrNormal)
		return
	case routingOptionDHTClientKwd:
		ncfg.Routing = core.DHTClientOption
	case routingOptionDHTKwd:
		ncfg.Routing = core.DHTOption
	case routingOptionNoneKwd:
		ncfg.Routing = core.NilRouterOption
	default:
		re.SetError(fmt.Errorf("unrecognized routing option: %s", routingOption), cmdkit.ErrNormal)
		return
	}

	node, err := core.NewNode(req.Context, ncfg)
	if err != nil {
		log.Error("error from node construction: ", err)
		re.SetError(err, cmdkit.ErrNormal)
		return
	}
	node.SetLocal(false)

	if node.PNetFingerprint != nil {
		fmt.Println("Swarm is limited to private network of peers with the swarm key")
		fmt.Printf("Swarm key fingerprint: %x\n", node.PNetFingerprint)
	}

	printSwarmAddrs(node)

	defer func() {
		// We wait for the node to close first, as the node has children
		// that it will wait for before closing, such as the API server.
		node.Close()

		select {
		case <-req.Context.Done():
			log.Info("Gracefully shut down daemon")
		default:
		}
	}()

	cctx.ConstructNode = func() (*core.Dms3FsNode, error) {
		return node, nil
	}

	// construct api endpoint - every time
	apiErrc, err := serveHTTPApi(req, cctx)
	if err != nil {
		re.SetError(err, cmdkit.ErrNormal)
		return
	}

	// construct fuse mountpoints - if the user provided the --mount flag
	mount, _ := req.Options[mountKwd].(bool)
	if mount && offline {
		re.SetError(errors.New("mount is not currently supported in offline mode"),
			cmdkit.ErrClient)
		return
	}
	if mount {
		if err := mountFuse(req, cctx); err != nil {
			re.SetError(err, cmdkit.ErrNormal)
			return
		}
	}

	// repo blockstore GC - if --enable-gc flag is present
	gcErrc, err := maybeRunGC(req, node)
	if err != nil {
		re.SetError(err, cmdkit.ErrNormal)
		return
	}

	// construct http gateway - if it is set in the config
	var gwErrc <-chan error
	if len(cfg.Addresses.Gateway) > 0 {
		var err error
		gwErrc, err = serveHTTPGateway(req, cctx)
		if err != nil {
			re.SetError(err, cmdkit.ErrNormal)
			return
		}
	}

	// initialize metrics collector
	prometheus.MustRegister(&corehttp.Dms3FsNodeCollector{Node: node})

	fmt.Printf("Daemon is ready\n")
	// collect long-running errors and block for shutdown
	// TODO(cryptix): our fuse currently doesnt follow this pattern for graceful shutdown
	for err := range merge(apiErrc, gwErrc, gcErrc) {
		if err != nil {
			log.Error(err)
			re.SetError(err, cmdkit.ErrNormal)
		}
	}
}

// serveHTTPApi collects options, creates listener, prints status message and starts serving requests
func serveHTTPApi(req *cmds.Request, cctx *oldcmds.Context) (<-chan error, error) {
	cfg, err := cctx.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("serveHTTPApi: GetConfig() failed: %s", err)
	}

	apiAddr, _ := req.Options[commands.ApiOption].(string)
	if apiAddr == "" {
		apiAddr = cfg.Addresses.API
	}
	apiMaddr, err := ma.NewMultiaddr(apiAddr)
	if err != nil {
		return nil, fmt.Errorf("serveHTTPApi: invalid API address: %q (err: %s)", apiAddr, err)
	}

	apiLis, err := manet.Listen(apiMaddr)
	if err != nil {
		return nil, fmt.Errorf("serveHTTPApi: manet.Listen(%s) failed: %s", apiMaddr, err)
	}
	// we might have listened to /tcp/0 - lets see what we are listing on
	apiMaddr = apiLis.Multiaddr()
	fmt.Printf("API server listening on %s\n", apiMaddr)

	// by default, we don't let you load arbitrary dms3fs objects through the api,
	// because this would open up the api to scripting vulnerabilities.
	// only the webui objects are allowed.
	// if you know what you're doing, go ahead and pass --unrestricted-api.
	unrestricted, _ := req.Options[unrestrictedApiAccessKwd].(bool)
	gatewayOpt := corehttp.GatewayOption(false, corehttp.WebUIPaths...)
	if unrestricted {
		gatewayOpt = corehttp.GatewayOption(true, "/dms3fs", "/dms3ns")
	}

	var opts = []corehttp.ServeOption{
		corehttp.MetricsCollectionOption("api"),
		corehttp.CheckVersionOption(),
		corehttp.CommandsOption(*cctx),
		corehttp.WebUIOption,
		gatewayOpt,
		corehttp.VersionOption(),
		defaultMux("/debug/vars"),
		defaultMux("/debug/pprof/"),
		corehttp.MetricsScrapingOption("/debug/metrics/prometheus"),
		corehttp.LogOption(),
	}

	if len(cfg.Gateway.RootRedirect) > 0 {
		opts = append(opts, corehttp.RedirectOption("", cfg.Gateway.RootRedirect))
	}

	node, err := cctx.ConstructNode()
	if err != nil {
		return nil, fmt.Errorf("serveHTTPApi: ConstructNode() failed: %s", err)
	}

	if err := node.Repo.SetAPIAddr(apiMaddr); err != nil {
		return nil, fmt.Errorf("serveHTTPApi: SetAPIAddr() failed: %s", err)
	}

	errc := make(chan error)
	go func() {
		errc <- corehttp.Serve(node, manet.NetListener(apiLis), opts...)
		close(errc)
	}()
	return errc, nil
}

// printSwarmAddrs prints the addresses of the host
func printSwarmAddrs(node *core.Dms3FsNode) {
	if !node.OnlineMode() {
		fmt.Println("Swarm not listening, running in offline mode.")
		return
	}

	var lisAddrs []string
	ifaceAddrs, err := node.PeerHost.Network().InterfaceListenAddresses()
	if err != nil {
		log.Errorf("failed to read listening addresses: %s", err)
	}
	for _, addr := range ifaceAddrs {
		lisAddrs = append(lisAddrs, addr.String())
	}
	sort.Sort(sort.StringSlice(lisAddrs))
	for _, addr := range lisAddrs {
		fmt.Printf("Swarm listening on %s\n", addr)
	}

	var addrs []string
	for _, addr := range node.PeerHost.Addrs() {
		addrs = append(addrs, addr.String())
	}
	sort.Sort(sort.StringSlice(addrs))
	for _, addr := range addrs {
		fmt.Printf("Swarm announcing %s\n", addr)
	}

}

// serveHTTPGateway collects options, creates listener, prints status message and starts serving requests
func serveHTTPGateway(req *cmds.Request, cctx *oldcmds.Context) (<-chan error, error) {
	cfg, err := cctx.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("serveHTTPGateway: GetConfig() failed: %s", err)
	}

	gatewayMaddr, err := ma.NewMultiaddr(cfg.Addresses.Gateway)
	if err != nil {
		return nil, fmt.Errorf("serveHTTPGateway: invalid gateway address: %q (err: %s)", cfg.Addresses.Gateway, err)
	}

	writable, writableOptionFound := req.Options[writableKwd].(bool)
	if !writableOptionFound {
		writable = cfg.Gateway.Writable
	}

	gwLis, err := manet.Listen(gatewayMaddr)
	if err != nil {
		return nil, fmt.Errorf("serveHTTPGateway: manet.Listen(%s) failed: %s", gatewayMaddr, err)
	}
	// we might have listened to /tcp/0 - lets see what we are listing on
	gatewayMaddr = gwLis.Multiaddr()

	if writable {
		fmt.Printf("Gateway (writable) server listening on %s\n", gatewayMaddr)
	} else {
		fmt.Printf("Gateway (readonly) server listening on %s\n", gatewayMaddr)
	}

	var opts = []corehttp.ServeOption{
		corehttp.MetricsCollectionOption("gateway"),
		corehttp.CheckVersionOption(),
		corehttp.CommandsROOption(*cctx),
		corehttp.VersionOption(),
		corehttp.DMS3NSHostnameOption(),
		corehttp.GatewayOption(writable, "/dms3fs", "/dms3ns"),
	}

	if len(cfg.Gateway.RootRedirect) > 0 {
		opts = append(opts, corehttp.RedirectOption("", cfg.Gateway.RootRedirect))
	}

	node, err := cctx.ConstructNode()
	if err != nil {
		return nil, fmt.Errorf("serveHTTPGateway: ConstructNode() failed: %s", err)
	}

	errc := make(chan error)
	go func() {
		errc <- corehttp.Serve(node, manet.NetListener(gwLis), opts...)
		close(errc)
	}()
	return errc, nil
}

//collects options and opens the fuse mountpoint
func mountFuse(req *cmds.Request, cctx *oldcmds.Context) error {
	cfg, err := cctx.GetConfig()
	if err != nil {
		return fmt.Errorf("mountFuse: GetConfig() failed: %s", err)
	}

	fsdir, found := req.Options[dms3fsMountKwd].(string)
	if !found {
		fsdir = cfg.Mounts.DMS3FS
	}

	nsdir, found := req.Options[dms3nsMountKwd].(string)
	if !found {
		nsdir = cfg.Mounts.DMS3NS
	}

	node, err := cctx.ConstructNode()
	if err != nil {
		return fmt.Errorf("mountFuse: ConstructNode() failed: %s", err)
	}

	err = nodeMount.Mount(node, fsdir, nsdir)
	if err != nil {
		return err
	}
	fmt.Printf("DMS3FS mounted at: %s\n", fsdir)
	fmt.Printf("DMS3NS mounted at: %s\n", nsdir)
	return nil
}

func maybeRunGC(req *cmds.Request, node *core.Dms3FsNode) (<-chan error, error) {
	enableGC, _ := req.Options[enableGCKwd].(bool)
	if !enableGC {
		return nil, nil
	}

	errc := make(chan error)
	go func() {
		errc <- corerepo.PeriodicGC(req.Context, node)
		close(errc)
	}()
	return errc, nil
}

// merge does fan-in of multiple read-only error channels
// taken from http://blog.golang.org/pipelines
func merge(cs ...<-chan error) <-chan error {
	var wg sync.WaitGroup
	out := make(chan error)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan error) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	for _, c := range cs {
		if c != nil {
			wg.Add(1)
			go output(c)
		}
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func YesNoPrompt(prompt string) bool {
	var s string
	for i := 0; i < 3; i++ {
		fmt.Printf("%s ", prompt)
		fmt.Scanf("%s", &s)
		switch s {
		case "y", "Y":
			return true
		case "n", "N":
			return false
		case "":
			return false
		}
		fmt.Println("Please press either 'y' or 'n'")
	}

	return false
}
