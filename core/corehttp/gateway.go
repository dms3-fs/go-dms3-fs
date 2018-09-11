package corehttp

import (
	"fmt"
	"net"
	"net/http"

	version "github.com/dms3-fs/go-dms3-fs"
	core "github.com/dms3-fs/go-dms3-fs/core"
	coreapi "github.com/dms3-fs/go-dms3-fs/core/coreapi"

	id "github.com/dms3-p2p/go-p2p/p2p/protocol/identify"
)

type GatewayConfig struct {
	Headers      map[string][]string
	Writable     bool
	PathPrefixes []string
}

func GatewayOption(writable bool, paths ...string) ServeOption {
	return func(n *core.Dms3FsNode, _ net.Listener, mux *http.ServeMux) (*http.ServeMux, error) {
		cfg, err := n.Repo.Config()
		if err != nil {
			return nil, err
		}

		gateway := newGatewayHandler(n, GatewayConfig{
			Headers:      cfg.Gateway.HTTPHeaders,
			Writable:     writable,
			PathPrefixes: cfg.Gateway.PathPrefixes,
		}, coreapi.NewCoreAPI(n))

		for _, p := range paths {
			mux.Handle(p+"/", gateway)
		}
		return mux, nil
	}
}

func VersionOption() ServeOption {
	return func(_ *core.Dms3FsNode, _ net.Listener, mux *http.ServeMux) (*http.ServeMux, error) {
		mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Commit: %s\n", version.CurrentCommit)
			fmt.Fprintf(w, "Client Version: %s\n", id.ClientVersion)
			fmt.Fprintf(w, "Protocol Version: %s\n", id.LibP2PVersion)
		})
		return mux, nil
	}
}
