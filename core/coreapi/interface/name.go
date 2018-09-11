package iface

import (
	"context"

	options "github.com/dms3-fs/go-dms3-fs/core/coreapi/interface/options"
)

// Dms3NsEntry specifies the interface to Dms3NsEntries
type Dms3NsEntry interface {
	// Name returns Dms3NsEntry name
	Name() string
	// Value returns Dms3NsEntry value
	Value() Path
}

// NameAPI specifies the interface to DMS3NS.
//
// DMS3NS is a PKI namespace, where names are the hashes of public keys, and the
// private key enables publishing new (signed) values. In both publish and
// resolve, the default name used is the node's own PeerID, which is the hash of
// its public key.
//
// You can use .Key API to list and generate more names and their respective keys.
type NameAPI interface {
	// Publish announces new DMS3NS name
	Publish(ctx context.Context, path Path, opts ...options.NamePublishOption) (Dms3NsEntry, error)

	// Resolve attempts to resolve the newest version of the specified name
	Resolve(ctx context.Context, name string, opts ...options.NameResolveOption) (Path, error)
}
