package coreapi

import (
	"context"
//	"io"

	coreiface "github.com/dms3-fs/go-dms3-fs/core/coreapi/interface"

	options "github.com/dms3-fs/go-dms3-fs/core/coreapi/interface/options"
//	ipath "github.com/dms3-fs/go-path"
	logging "github.com/dms3-fs/go-log"
)

// log is the command logger
var log = logging.Logger("command/index")

type IndexAPI CoreAPI

type repoList struct {
	name 	string
	id 		string
}

// Name returns the repository name.
func (r *repoList) Name() string {
	return r.name
}

// Id returns the repository id.
func (r *repoList) Id() string {
	return r.id
}

//type repoList []repoEntry


// Index returns the data contained by an DMS3FS or DMS3NS object(s) at path `p`.
//func (api *IndexAPI) Index(ctx context.Context, r io.Reader) (coreiface.Reader, error) {
func (api *IndexAPI) Index(ctx context.Context, p coreiface.Path, opts ...options.IndexListOption) (coreiface.RepoList, error) {
	iopts, err := options.IndexListOptions(opts...)
	if err != nil {
		return nil, err
	}
	n := api.node
/*
	pth, err := ipath.ParsePath(p.String())
	if err != nil {
		return nil, err
	}
*/
	log.Debugf("meta %s", iopts.Meta)

	return &repoList{
		name:  "testing name",
		id: n.Identity.Pretty(),
	}, nil
}

/*

* local command, cannot run on daemon

1) configuration management of draft parameters file *
	syntax: index gen params
	arguments:
		path
		-- the path to a file to be generated
	options:
	outputs: draft default parameters file
	effects:
	- generates a draft parameters file with built-in default paramters

	futures:
	- overwrite entries in draft paramaters file
	- invoke editor to modify draft paramaters file
	- verify [read/parse/validate] draft parameters file
		- prints messages showing invalid parameter that needs to be fixed

2) same as above for metadata definitions file *
	options:
	-template/t=<template-name>
		-- name of template file, ex: blogs,...

3) create a new index data repository *
	syntax:
	index mk name
	arguments:
	-name/-n=<repo-name> -- repository name, required
	options:
	-param-file=<params-file>
		-- defines index parameters, or use system defaults
	-meta-file=<meta-file>
		-- defines index metadata, or use system defaults
	-max-area/-ma=<int>
		-- number of logical areas, defaults to 63 (same as physical)
	-max-cat/-mc=<int>
		-- number of logical categories, defaults to 63 (same as physical)
	outputs: hash referring to index DAG object
	effects:
	- verify draft parameters file
	- verify metadef file
	- create config director
	- add files to config directory
	- create index director
	- create repo director
	- stage above in tmp folder then add -r to Unixfs
	- print out last hash


*/
