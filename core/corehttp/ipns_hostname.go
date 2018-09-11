package corehttp

import (
	"context"
	"net"
	"net/http"
	"strings"

	core "github.com/dms3-fs/go-dms3-fs/core"
	namesys "github.com/dms3-fs/go-dms3-fs/namesys"
	nsopts "github.com/dms3-fs/go-dms3-fs/namesys/opts"

	isd "github.com/jbenet/go-is-domain"
)

// DMS3NSHostnameOption rewrites an incoming request if its Host: header contains
// an DMS3NS name.
// The rewritten request points at the resolved name on the gateway handler.
func DMS3NSHostnameOption() ServeOption {
	return func(n *core.Dms3FsNode, _ net.Listener, mux *http.ServeMux) (*http.ServeMux, error) {
		childMux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithCancel(n.Context())
			defer cancel()

			host := strings.SplitN(r.Host, ":", 2)[0]
			if len(host) > 0 && isd.IsDomain(host) {
				name := "/dms3ns/" + host
				_, err := n.Namesys.Resolve(ctx, name, nsopts.Depth(1))
				if err == nil || err == namesys.ErrResolveRecursion {
					r.Header.Set("X-Dms3Ns-Original-Path", r.URL.Path)
					r.URL.Path = name + r.URL.Path
				}
			}
			childMux.ServeHTTP(w, r)
		})
		return childMux, nil
	}
}
