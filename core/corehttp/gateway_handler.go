package corehttp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	gopath "path"
	"runtime/debug"
	"strings"
	"time"

	core "github.com/dms3-fs/go-dms3-fs/core"
	coreiface "github.com/dms3-fs/go-dms3-fs/core/coreapi/interface"
	"github.com/dms3-fs/go-dms3-fs/dagutils"
	dag "github.com/dms3-fs/go-merkledag"
	path "github.com/dms3-fs/go-path"
	resolver "github.com/dms3-fs/go-path/resolver"
	ft "github.com/dms3-fs/go-unixfs"
	"github.com/dms3-fs/go-unixfs/importer"
	uio "github.com/dms3-fs/go-unixfs/io"

	humanize "github.com/dustin/go-humanize"
	cid "github.com/dms3-fs/go-cid"
	chunker "github.com/dms3-fs/go-fs-chunker"
	dms3ld "github.com/dms3-fs/go-ld-format"
	routing "github.com/dms3-p2p/go-p2p-routing"
	multibase "github.com/dms3-mft/go-multibase"
)

const (
	dms3fsPathPrefix = "/dms3fs/"
	dms3nsPathPrefix = "/dms3ns/"
)

// gatewayHandler is a HTTP handler that serves DMS3FS objects (accessible by default at /dms3fs/<path>)
// (it serves requests like GET /dms3fs/QmVRzPKPzNtSrEzBFm2UZfxmPAgnaLke4DMcerbsGGSaFe/link)
type gatewayHandler struct {
	node   *core.Dms3FsNode
	config GatewayConfig
	api    coreiface.CoreAPI
}

func newGatewayHandler(n *core.Dms3FsNode, c GatewayConfig, api coreiface.CoreAPI) *gatewayHandler {
	i := &gatewayHandler{
		node:   n,
		config: c,
		api:    api,
	}
	return i
}

// TODO(cryptix):  find these helpers somewhere else
func (i *gatewayHandler) newDagFromReader(r io.Reader) (dms3ld.Node, error) {
	// TODO(cryptix): change and remove this helper once PR1136 is merged
	// return ufs.AddFromReader(i.node, r.Body)
	return importer.BuildDagFromReader(
		i.node.DAG,
		chunker.DefaultSplitter(r))
}

// TODO(btc): break this apart into separate handlers using a more expressive muxer
func (i *gatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(i.node.Context(), time.Hour)
	// the hour is a hard fallback, we don't expect it to happen, but just in case
	defer cancel()

	if cn, ok := w.(http.CloseNotifier); ok {
		clientGone := cn.CloseNotify()
		go func() {
			select {
			case <-clientGone:
			case <-ctx.Done():
			}
			cancel()
		}()
	}

	defer func() {
		if r := recover(); r != nil {
			log.Error("A panic occurred in the gateway handler!")
			log.Error(r)
			debug.PrintStack()
		}
	}()

	if i.config.Writable {
		switch r.Method {
		case "POST":
			i.postHandler(ctx, w, r)
			return
		case "PUT":
			i.putHandler(w, r)
			return
		case "DELETE":
			i.deleteHandler(w, r)
			return
		}
	}

	if r.Method == "GET" || r.Method == "HEAD" {
		i.getOrHeadHandler(ctx, w, r)
		return
	}

	if r.Method == "OPTIONS" {
		i.optionsHandler(w, r)
		return
	}

	errmsg := "Method " + r.Method + " not allowed: "
	if !i.config.Writable {
		w.WriteHeader(http.StatusMethodNotAllowed)
		errmsg = errmsg + "read only access"
	} else {
		w.WriteHeader(http.StatusBadRequest)
		errmsg = errmsg + "bad request for " + r.URL.Path
	}
	fmt.Fprint(w, errmsg)
}

func (i *gatewayHandler) optionsHandler(w http.ResponseWriter, r *http.Request) {
	/*
		OPTIONS is a noop request that is used by the browsers to check
		if server accepts cross-site XMLHttpRequest (indicated by the presence of CORS headers)
		https://developer.mozilla.org/en-US/docs/Web/HTTP/Access_control_CORS#Preflighted_requests
	*/
	i.addUserHeaders(w) // return all custom headers (including CORS ones, if set)
}

func (i *gatewayHandler) getOrHeadHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	escapedURLPath := r.URL.EscapedPath()

	// If the gateway is behind a reverse proxy and mounted at a sub-path,
	// the prefix header can be set to signal this sub-path.
	// It will be prepended to links in directory listings and the index.html redirect.
	prefix := ""
	if prfx := r.Header.Get("X-Dms3Fs-Gateway-Prefix"); len(prfx) > 0 {
		for _, p := range i.config.PathPrefixes {
			if prfx == p || strings.HasPrefix(prfx, p+"/") {
				prefix = prfx
				break
			}
		}
	}

	// DMS3NSHostnameOption might have constructed an DMS3NS path using the Host header.
	// In this case, we need the original path for constructing redirects
	// and links that match the requested URL.
	// For example, http://example.net would become /dms3ns/example.net, and
	// the redirects and links would end up as http://example.net/dms3ns/example.net
	originalUrlPath := prefix + urlPath
	dms3nsHostname := false
	if hdr := r.Header.Get("X-Dms3Ns-Original-Path"); len(hdr) > 0 {
		originalUrlPath = prefix + hdr
		dms3nsHostname = true
	}

	parsedPath, err := coreiface.ParsePath(urlPath)
	if err != nil {
		webError(w, "invalid dms3fs path", err, http.StatusBadRequest)
		return
	}

	// Resolve path to the final DAG node for the ETag
	resolvedPath, err := i.api.ResolvePath(ctx, parsedPath)
	if err == coreiface.ErrOffline && !i.node.OnlineMode() {
		webError(w, "dms3fs resolve -r "+escapedURLPath, err, http.StatusServiceUnavailable)
		return
	} else if err != nil {
		webError(w, "dms3fs resolve -r "+escapedURLPath, err, http.StatusNotFound)
		return
	}

	dr, err := i.api.Unixfs().Cat(ctx, resolvedPath)
	dir := false
	switch err {
	case nil:
		// Cat() worked
		defer dr.Close()
	case coreiface.ErrIsDir:
		dir = true
	default:
		webError(w, "dms3fs cat "+escapedURLPath, err, http.StatusNotFound)
		return
	}

	// Check etag send back to us
	etag := "\"" + resolvedPath.Cid().String() + "\""
	if r.Header.Get("If-None-Match") == etag || r.Header.Get("If-None-Match") == "W/"+etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	i.addUserHeaders(w) // ok, _now_ write user's headers.
	w.Header().Set("X-DMS3FS-Path", urlPath)
	w.Header().Set("Etag", etag)

	// set 'allowed' headers
	// & expose those headers
	var allowedHeadersArr = []string{
		"Content-Range",
		"X-Chunked-Output",
		"X-Stream-Output",
	}

	var allowedHeaders = strings.Join(allowedHeadersArr, ", ")

	w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
	w.Header().Set("Access-Control-Expose-Headers", allowedHeaders)

	// Suborigin header, sandboxes apps from each other in the browser (even
	// though they are served from the same gateway domain).
	//
	// Omitted if the path was treated by DMS3NSHostnameOption(), for example
	// a request for http://example.net/ would be changed to /dms3ns/example.net/,
	// which would turn into an incorrect Suborigin header.
	// In this case the correct thing to do is omit the header because it is already
	// handled correctly without a Suborigin.
	//
	// NOTE: This is not yet widely supported by browsers.
	if !dms3nsHostname {
		// e.g.: 1="dms3fs", 2="QmYuNaKwY...", ...
		pathComponents := strings.SplitN(urlPath, "/", 4)

		var suboriginRaw []byte
		cidDecoded, err := cid.Decode(pathComponents[2])
		if err != nil {
			// component 2 doesn't decode with cid, so it must be a hostname
			suboriginRaw = []byte(strings.ToLower(pathComponents[2]))
		} else {
			suboriginRaw = cidDecoded.Bytes()
		}

		base32Encoded, err := multibase.Encode(multibase.Base32, suboriginRaw)
		if err != nil {
			internalWebError(w, err)
			return
		}

		suborigin := pathComponents[1] + "000" + strings.ToLower(base32Encoded)
		w.Header().Set("Suborigin", suborigin)
	}

	// set these headers _after_ the error, for we may just not have it
	// and dont want the client to cache a 500 response...
	// and only if it's /dms3fs!
	// TODO: break this out when we split /dms3fs /dms3ns routes.
	modtime := time.Now()

	if strings.HasPrefix(urlPath, dms3fsPathPrefix) && !dir {
		w.Header().Set("Cache-Control", "public, max-age=29030400, immutable")

		// set modtime to a really long time ago, since files are immutable and should stay cached
		modtime = time.Unix(1, 0)
	}

	if !dir {
		urlFilename := r.URL.Query().Get("filename")
		var name string
		if urlFilename != "" {
			w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename*=UTF-8''%s", url.PathEscape(urlFilename)))
			name = urlFilename
		} else {
			name = gopath.Base(urlPath)
		}
		i.serveFile(w, r, name, modtime, dr)
		return
	}

	nd, err := i.api.ResolveNode(ctx, resolvedPath)
	if err != nil {
		internalWebError(w, err)
		return
	}

	dirr, err := uio.NewDirectoryFromNode(i.node.DAG, nd)
	if err != nil {
		internalWebError(w, err)
		return
	}

	ixnd, err := dirr.Find(ctx, "index.html")
	switch {
	case err == nil:
		dirwithoutslash := urlPath[len(urlPath)-1] != '/'
		goget := r.URL.Query().Get("go-get") == "1"
		if dirwithoutslash && !goget {
			// See comment above where originalUrlPath is declared.
			http.Redirect(w, r, originalUrlPath+"/", 302)
			return
		}

		dr, err := i.api.Unixfs().Cat(ctx, coreiface.Dms3FsPath(ixnd.Cid()))
		if err != nil {
			internalWebError(w, err)
			return
		}
		defer dr.Close()

		// write to request
		http.ServeContent(w, r, "index.html", modtime, dr)
		return
	default:
		internalWebError(w, err)
		return
	case os.IsNotExist(err):
	}

	if r.Method == "HEAD" {
		return
	}

	// storage for directory listing
	var dirListing []directoryItem
	dirr.ForEachLink(ctx, func(link *dms3ld.Link) error {
		// See comment above where originalUrlPath is declared.
		di := directoryItem{humanize.Bytes(link.Size), link.Name, gopath.Join(originalUrlPath, link.Name)}
		dirListing = append(dirListing, di)
		return nil
	})

	// construct the correct back link
	// https://github.com/dms3-fs/go-dms3-fs/issues/1365
	var backLink string = prefix + urlPath

	// don't go further up than /dms3fs/$hash/
	pathSplit := path.SplitList(backLink)
	switch {
	// keep backlink
	case len(pathSplit) == 3: // url: /dms3fs/$hash

	// keep backlink
	case len(pathSplit) == 4 && pathSplit[3] == "": // url: /dms3fs/$hash/

	// add the correct link depending on wether the path ends with a slash
	default:
		if strings.HasSuffix(backLink, "/") {
			backLink += "./.."
		} else {
			backLink += "/.."
		}
	}

	// strip /dms3fs/$hash from backlink if DMS3NSHostnameOption touched the path.
	if dms3nsHostname {
		backLink = prefix + "/"
		if len(pathSplit) > 5 {
			// also strip the trailing segment, because it's a backlink
			backLinkParts := pathSplit[3 : len(pathSplit)-2]
			backLink += path.Join(backLinkParts) + "/"
		}
	}

	// See comment above where originalUrlPath is declared.
	tplData := listingTemplateData{
		Listing:  dirListing,
		Path:     originalUrlPath,
		BackLink: backLink,
	}
	err = listingTemplate.Execute(w, tplData)
	if err != nil {
		internalWebError(w, err)
		return
	}
}

type sizeReadSeeker interface {
	Size() uint64

	io.ReadSeeker
}

type sizeSeeker struct {
	sizeReadSeeker
}

func (s *sizeSeeker) Seek(offset int64, whence int) (int64, error) {
	if whence == io.SeekEnd && offset == 0 {
		return int64(s.Size()), nil
	}

	return s.sizeReadSeeker.Seek(offset, whence)
}

func (i *gatewayHandler) serveFile(w http.ResponseWriter, req *http.Request, name string, modtime time.Time, content io.ReadSeeker) {
	if sp, ok := content.(sizeReadSeeker); ok {
		content = &sizeSeeker{
			sizeReadSeeker: sp,
		}
	}

	http.ServeContent(w, req, name, modtime, content)
}

func (i *gatewayHandler) postHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	p, err := i.api.Unixfs().Add(ctx, r.Body)
	if err != nil {
		internalWebError(w, err)
		return
	}

	i.addUserHeaders(w) // ok, _now_ write user's headers.
	w.Header().Set("DMS3FS-Hash", p.Cid().String())
	http.Redirect(w, r, p.String(), http.StatusCreated)
}

func (i *gatewayHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	// TODO(cryptix): move me to ServeHTTP and pass into all handlers
	ctx, cancel := context.WithCancel(i.node.Context())
	defer cancel()

	rootPath, err := path.ParsePath(r.URL.Path)
	if err != nil {
		webError(w, "putHandler: DMS3FS path not valid", err, http.StatusBadRequest)
		return
	}

	rsegs := rootPath.Segments()
	if rsegs[0] == dms3nsPathPrefix {
		webError(w, "putHandler: updating named entries not supported", errors.New("WritableGateway: dms3ns put not supported"), http.StatusBadRequest)
		return
	}

	var newnode dms3ld.Node
	if rsegs[len(rsegs)-1] == "QmUNLLsPACCz1vLxQVkXqqLX5R1X345qqfHbsf67hvA3Nn" {
		newnode = ft.EmptyDirNode()
	} else {
		putNode, err := i.newDagFromReader(r.Body)
		if err != nil {
			webError(w, "putHandler: Could not create DAG from request", err, http.StatusInternalServerError)
			return
		}
		newnode = putNode
	}

	var newPath string
	if len(rsegs) > 1 {
		newPath = path.Join(rsegs[2:])
	}

	var newcid *cid.Cid
	rnode, err := core.Resolve(ctx, i.node.Namesys, i.node.Resolver, rootPath)
	switch ev := err.(type) {
	case resolver.ErrNoLink:
		// ev.Node < node where resolve failed
		// ev.Name < new link
		// but we need to patch from the root
		c, err := cid.Decode(rsegs[1])
		if err != nil {
			webError(w, "putHandler: bad input path", err, http.StatusBadRequest)
			return
		}

		rnode, err := i.node.DAG.Get(ctx, c)
		if err != nil {
			webError(w, "putHandler: Could not create DAG from request", err, http.StatusInternalServerError)
			return
		}

		pbnd, ok := rnode.(*dag.ProtoNode)
		if !ok {
			webError(w, "Cannot read non protobuf nodes through gateway", dag.ErrNotProtobuf, http.StatusBadRequest)
			return
		}

		e := dagutils.NewDagEditor(pbnd, i.node.DAG)
		err = e.InsertNodeAtPath(ctx, newPath, newnode, ft.EmptyDirNode)
		if err != nil {
			webError(w, "putHandler: InsertNodeAtPath failed", err, http.StatusInternalServerError)
			return
		}

		nnode, err := e.Finalize(ctx, i.node.DAG)
		if err != nil {
			webError(w, "putHandler: could not get node", err, http.StatusInternalServerError)
			return
		}

		newcid = nnode.Cid()

	case nil:
		pbnd, ok := rnode.(*dag.ProtoNode)
		if !ok {
			webError(w, "Cannot read non protobuf nodes through gateway", dag.ErrNotProtobuf, http.StatusBadRequest)
			return
		}

		pbnewnode, ok := newnode.(*dag.ProtoNode)
		if !ok {
			webError(w, "Cannot read non protobuf nodes through gateway", dag.ErrNotProtobuf, http.StatusBadRequest)
			return
		}

		// object set-data case
		pbnd.SetData(pbnewnode.Data())

		newcid = pbnd.Cid()
		err = i.node.DAG.Add(ctx, pbnd)
		if err != nil {
			nnk := newnode.Cid()
			webError(w, fmt.Sprintf("putHandler: Could not add newnode(%q) to root(%q)", nnk.String(), newcid.String()), err, http.StatusInternalServerError)
			return
		}
	default:
		webError(w, "could not resolve root DAG", ev, http.StatusInternalServerError)
		return
	}

	i.addUserHeaders(w) // ok, _now_ write user's headers.
	w.Header().Set("DMS3FS-Hash", newcid.String())
	http.Redirect(w, r, gopath.Join(dms3fsPathPrefix, newcid.String(), newPath), http.StatusCreated)
}

func (i *gatewayHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	ctx, cancel := context.WithCancel(i.node.Context())
	defer cancel()

	p, err := path.ParsePath(urlPath)
	if err != nil {
		webError(w, "failed to parse path", err, http.StatusBadRequest)
		return
	}

	c, components, err := path.SplitAbsPath(p)
	if err != nil {
		webError(w, "Could not split path", err, http.StatusInternalServerError)
		return
	}

	tctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	rootnd, err := i.node.Resolver.DAG.Get(tctx, c)
	if err != nil {
		webError(w, "Could not resolve root object", err, http.StatusBadRequest)
		return
	}

	pathNodes, err := i.node.Resolver.ResolveLinks(tctx, rootnd, components[:len(components)-1])
	if err != nil {
		webError(w, "Could not resolve parent object", err, http.StatusBadRequest)
		return
	}

	pbnd, ok := pathNodes[len(pathNodes)-1].(*dag.ProtoNode)
	if !ok {
		webError(w, "Cannot read non protobuf nodes through gateway", dag.ErrNotProtobuf, http.StatusBadRequest)
		return
	}

	// TODO(cyrptix): assumes len(pathNodes) > 1 - not found is an error above?
	err = pbnd.RemoveNodeLink(components[len(components)-1])
	if err != nil {
		webError(w, "Could not delete link", err, http.StatusBadRequest)
		return
	}

	var newnode *dag.ProtoNode = pbnd
	for j := len(pathNodes) - 2; j >= 0; j-- {
		if err := i.node.DAG.Add(ctx, newnode); err != nil {
			webError(w, "Could not add node", err, http.StatusInternalServerError)
			return
		}

		pathpb, ok := pathNodes[j].(*dag.ProtoNode)
		if !ok {
			webError(w, "Cannot read non protobuf nodes through gateway", dag.ErrNotProtobuf, http.StatusBadRequest)
			return
		}

		newnode, err = pathpb.UpdateNodeLink(components[j], newnode)
		if err != nil {
			webError(w, "Could not update node links", err, http.StatusInternalServerError)
			return
		}
	}

	if err := i.node.DAG.Add(ctx, newnode); err != nil {
		webError(w, "Could not add root node", err, http.StatusInternalServerError)
		return
	}

	// Redirect to new path
	ncid := newnode.Cid()

	i.addUserHeaders(w) // ok, _now_ write user's headers.
	w.Header().Set("DMS3FS-Hash", ncid.String())
	http.Redirect(w, r, gopath.Join(dms3fsPathPrefix+ncid.String(), path.Join(components[:len(components)-1])), http.StatusCreated)
}

func (i *gatewayHandler) addUserHeaders(w http.ResponseWriter) {
	for k, v := range i.config.Headers {
		w.Header()[k] = v
	}
}

func webError(w http.ResponseWriter, message string, err error, defaultCode int) {
	if _, ok := err.(resolver.ErrNoLink); ok {
		webErrorWithCode(w, message, err, http.StatusNotFound)
	} else if err == routing.ErrNotFound {
		webErrorWithCode(w, message, err, http.StatusNotFound)
	} else if err == context.DeadlineExceeded {
		webErrorWithCode(w, message, err, http.StatusRequestTimeout)
	} else {
		webErrorWithCode(w, message, err, defaultCode)
	}
}

func webErrorWithCode(w http.ResponseWriter, message string, err error, code int) {
	w.WriteHeader(code)

	fmt.Fprintf(w, "%s: %s\n", message, err)
	if code >= 500 {
		log.Warningf("server error: %s: %s", err)
	}
}

// return a 500 error and log
func internalWebError(w http.ResponseWriter, err error) {
	webErrorWithCode(w, "internalWebError", err, http.StatusInternalServerError)
}
