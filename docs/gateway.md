# Gateway

An DMS3FS Gateway acts as a bridge between traditional web browsers and DMS3FS.
Through the gateway, users can browse files and websites stored in DMS3FS as if
they were stored in a traditional web server.

By default, go-dms3-fs nodes run a gateway at `http://127.0.0.1:5001/`.

We also provide a public gateway at `https://dms3.io`. If you've ever seen a
link in the form `https://dms3.io/dms3fs/Qm...`, that's being served from *our*
gateway.

## Configuration

The gateway's configuration options are (briefly) described in the
[config](https://github.com/dms3-fs/go-dms3-fs/blob/master/docs/config.md#gateway)
documentation.

## Directories

For convenience, the gateway (mostly) acts like a normal web-server when serving
a directory:

1. If the directory contains an `index.html` file:
  1. If the path does not end in a `/`, append a `/` and redirect. This helps
     avoid serving duplicate content from different paths.<sup>&dagger;</sup>
  2. Otherwise, serve the `index.html` file.
2. Dynamically build and serve a listing of the contents of the directory.

<sub><sup>&dagger;</sup>This redirect is skipped if the query string contains a
`go-get=1` parameter. See [PR#3964](https://github.com/dms3-fs/go-dms3-fs/pull/3963)
for details</sub>

## Filenames

When downloading files, browsers will usually guess a file's filename by looking
at the last component of the path. Unfortunately, when linking *directly* to a
file (with no containing directory), the final component is just a CID
(`Qm...`). This isn't exactly user-friendly.

To work around this issue, you can add a `filename=some_filename` parameter to
your query string to explicitly specify the filename. For example:

> https://dms3.io/dms3fs/QmfM2r8seH2GiRaC4esTjeraXEachRt8ZsSeGaWTPLyMoG?filename=hello_world.txt

## MIME-Types

TODO
