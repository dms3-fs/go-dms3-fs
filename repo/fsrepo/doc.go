// package fsrepo
//
// TODO explain the package roadmap...
//
//   .dms3-fs/
//   ├── client/
//   |   ├── client.lock          <------ protects client/ + signals its own pid
//   │   ├── dms3fs-client.cpuprof
//   │   └── dms3fs-client.memprof
//   ├── config
//   ├── daemon/
//   │   ├── daemon.lock          <------ protects daemon/ + signals its own address
//   │   ├── dms3fs-daemon.cpuprof
//   │   └── dms3fs-daemon.memprof
//   ├── datastore/
//   ├── repo.lock                <------ protects datastore/ and config
//   └── version
package fsrepo

// TODO prevent multiple daemons from running
