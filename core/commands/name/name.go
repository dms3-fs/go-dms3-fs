package name

import (
	"github.com/dms3-fs/go-fs-cmdkit"
	"github.com/dms3-fs/go-fs-cmds"
)

type Dms3NsEntry struct {
	Name  string
	Value string
}

var NameCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Publish and resolve DMS3NS names.",
		ShortDescription: `
DMS3NS is a PKI namespace, where names are the hashes of public keys, and
the private key enables publishing new (signed) values. In both publish
and resolve, the default name used is the node's own PeerID,
which is the hash of its public key.
`,
		LongDescription: `
DMS3NS is a PKI namespace, where names are the hashes of public keys, and
the private key enables publishing new (signed) values. In both publish
and resolve, the default name used is the node's own PeerID,
which is the hash of its public key.

You can use the 'dms3fs key' commands to list and generate more names and their
respective keys.

Examples:

Publish an <dms3fs-path> with your default name:

  > dms3fs name publish /dms3fs/QmatmE9msSfkKxoffpHwNLNKgwZG8eT9Bud6YoPab52vpy
  Published to QmbCMUZw6JFeZ7Wp9jkzbye3Fzp2GGcPgC3nmeUjfVF87n: /dms3fs/QmatmE9msSfkKxoffpHwNLNKgwZG8eT9Bud6YoPab52vpy

Publish an <dms3fs-path> with another name, added by an 'dms3fs key' command:

  > dms3fs key gen --type=rsa --size=2048 mykey
  > dms3fs name publish --key=mykey /dms3fs/QmatmE9msSfkKxoffpHwNLNKgwZG8eT9Bud6YoPab52vpy
  Published to QmSrPmbaUKA3ZodhzPWZnpFgcPMFWF4QsxXbkWfEptTBJd: /dms3fs/QmatmE9msSfkKxoffpHwNLNKgwZG8eT9Bud6YoPab52vpy

Resolve the value of your name:

  > dms3fs name resolve
  /dms3fs/QmatmE9msSfkKxoffpHwNLNKgwZG8eT9Bud6YoPab52vpy

Resolve the value of another name:

  > dms3fs name resolve QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ
  /dms3fs/QmSiTko9JZyabH56y2fussEt1A5oDqsFXB3CkvAqraFryz

Resolve the value of a dnslink:

  > dms3fs name resolve dms3.io
  /dms3fs/QmaBvfZooxWkrv7D3r8LS9moNjzD2o525XMZze69hhoxf5

`,
	},

	Subcommands: map[string]*cmds.Command{
		"publish": PublishCmd,
		"resolve": Dms3NsCmd,
		"pubsub":  Dms3NsPubsubCmd,
	},
}
