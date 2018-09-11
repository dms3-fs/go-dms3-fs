#!/usr/bin/env bash
# vim: set expandtab sw=2 ts=2:

# bash safe mode
set -euo pipefail
IFS=$'\n\t'


OUTPUT=$(realpath ${1:-go-dms3-fs-source.tar.gz})

TMPDIR="$(mktemp -d)"
NEWDMS3FS="$TMPDIR/github.com/dms3-fs/go-dms3-fs"
mkdir -p "$NEWDMS3FS"
cp -r . "$NEWDMS3FS"
( cd "$NEWDMS3FS" &&
  echo $PWD &&
  GOPATH="$TMPDIR" dms3gx install --local &&
  (git rev-parse --short HEAD || true) > .tarball &&
  tar -czf "$OUTPUT" --exclude="./.git" .
)

rm -rf "$TMPDIR"
