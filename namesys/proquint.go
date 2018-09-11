package namesys

import (
	"errors"
	"time"

	context "context"

	proquint "github.com/bren2010/proquint"
	opts "github.com/dms3-fs/go-dms3-fs/namesys/opts"
	path "github.com/dms3-fs/go-path"
)

type ProquintResolver struct{}

// Resolve implements Resolver.
func (r *ProquintResolver) Resolve(ctx context.Context, name string, options ...opts.ResolveOpt) (path.Path, error) {
	return resolve(ctx, r, name, opts.ProcessOpts(options), "/dms3ns/")
}

// resolveOnce implements resolver. Decodes the proquint string.
func (r *ProquintResolver) resolveOnce(ctx context.Context, name string, options *opts.ResolveOpts) (path.Path, time.Duration, error) {
	ok, err := proquint.IsProquint(name)
	if err != nil || !ok {
		return "", 0, errors.New("not a valid proquint string")
	}
	// Return a 0 TTL as caching this result is pointless.
	return path.FromString(string(proquint.Decode(name))), 0, nil
}
