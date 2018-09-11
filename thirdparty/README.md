thirdparty consists of Golang packages that contain no go-dms3-fs dependencies and
may be vendored dms3-fs/go-dms3-fs at a later date.

packages in under this directory _must not_ import packages under
`dms3-fs/go-dms3-fs` that are not also under `thirdparty`.
