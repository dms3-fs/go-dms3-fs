/*
Package namesys implements resolvers and publishers for the DMS3FS
naming system (DMS3NS).

The core of DMS3FS is an immutable, content-addressable Merkle graph.
That works well for many use cases, but doesn't allow you to answer
questions like "what is Alice's current homepage?".  The mutable name
system allows Alice to publish information like:

  The current homepage for alice.example.com is
  /dms3fs/Qmcqtw8FfrVSBaRmbWwHxt3AuySBhJLcvmFYi3Lbc4xnwj

or:

  The current homepage for node
  QmatmE9msSfkKxoffpHwNLNKgwZG8eT9Bud6YoPab52vpy
  is
  /dms3fs/Qmcqtw8FfrVSBaRmbWwHxt3AuySBhJLcvmFYi3Lbc4xnwj

The mutable name system also allows users to resolve those references
to find the immutable DMS3FS object currently referenced by a given
mutable name.

For command-line bindings to this functionality, see:

  dms3fs name
  dms3fs dns
  dms3fs resolve
*/
package namesys

import (
	"errors"
	"time"

	context "context"

	opts "github.com/dms3-fs/go-dms3-fs/namesys/opts"
	path "github.com/dms3-fs/go-path"

	ci "github.com/dms3-p2p/go-p2p-crypto"
)

// ErrResolveFailed signals an error when attempting to resolve.
var ErrResolveFailed = errors.New("could not resolve name")

// ErrResolveRecursion signals a recursion-depth limit.
var ErrResolveRecursion = errors.New(
	"could not resolve name (recursion limit exceeded)")

// ErrPublishFailed signals an error when attempting to publish.
var ErrPublishFailed = errors.New("could not publish name")

// Namesys represents a cohesive name publishing and resolving system.
//
// Publishing a name is the process of establishing a mapping, a key-value
// pair, according to naming rules and databases.
//
// Resolving a name is the process of looking up the value associated with the
// key (name).
type NameSystem interface {
	Resolver
	Publisher
}

// Resolver is an object capable of resolving names.
type Resolver interface {

	// Resolve performs a recursive lookup, returning the dereferenced
	// path.  For example, if dms3.io has a DNS TXT record pointing to
	//   /dms3ns/QmatmE9msSfkKxoffpHwNLNKgwZG8eT9Bud6YoPab52vpy
	// and there is a DHT DMS3NS entry for
	//   QmatmE9msSfkKxoffpHwNLNKgwZG8eT9Bud6YoPab52vpy
	//   -> /dms3fs/Qmcqtw8FfrVSBaRmbWwHxt3AuySBhJLcvmFYi3Lbc4xnwj
	// then
	//   Resolve(ctx, "/dms3ns/dms3.io")
	// will resolve both names, returning
	//   /dms3fs/Qmcqtw8FfrVSBaRmbWwHxt3AuySBhJLcvmFYi3Lbc4xnwj
	//
	// There is a default depth-limit to avoid infinite recursion.  Most
	// users will be fine with this default limit, but if you need to
	// adjust the limit you can specify it as an option.
	Resolve(ctx context.Context, name string, options ...opts.ResolveOpt) (value path.Path, err error)
}

// Publisher is an object capable of publishing particular names.
type Publisher interface {

	// Publish establishes a name-value mapping.
	// TODO make this not PrivKey specific.
	Publish(ctx context.Context, name ci.PrivKey, value path.Path) error

	// TODO: to be replaced by a more generic 'PublishWithValidity' type
	// call once the records spec is implemented
	PublishWithEOL(ctx context.Context, name ci.PrivKey, value path.Path, eol time.Time) error
}
