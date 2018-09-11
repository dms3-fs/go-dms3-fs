# Assets loaded in with DMS3FS

## Generating docs

Do not edit the .go files directly.

Instead, edit the source files and use `go generate` from within the
assets directory:

```
go get -u github.com/jteeuwen/go-bindata/...
go generate
```

Before using `go generate`:
1. find the last published hash of package dir-index-html, see
    - the dependency section of go-dms3-fs/package.json, or
    - the file dir-index-html/.dms3-gx/lastpubver
2. verify that the last publish hash matches the hash listed in the file assets.go `go generate` line -prefix flag and the input directory list
3. verify that the last publish hash matches the hash listed on line [37] initializing "initDirPath"

After using `go generate`:
1. install dms3fs,
2. dms3fs init
    - init will seed the fsrepo with these assets using go-dms3-fs/cmd/dms3fs/init.go:assets.SeedInitDocs, which also prints local access instructions
    - the dms3fs gateway which provide access via http, see go-dms3-fs/core/corehttp/gateway_handler.go and gateway_indexPage.go
