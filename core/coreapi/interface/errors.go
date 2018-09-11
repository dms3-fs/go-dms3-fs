package iface

import "errors"

var (
	ErrIsDir   = errors.New("object is a directory")
	ErrOffline = errors.New("can't resolve, dms3fs node is offline")
)
