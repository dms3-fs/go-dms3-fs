package loader

import (
	"github.com/dms3-fs/go-dms3-fs/plugin"
	plugindms3ldgit "github.com/dms3-fs/go-dms3-fs/plugin/plugins/git"
)

// DO NOT EDIT THIS FILE
// This file is being generated as part of plugin build process
// To change it, modify the plugin/loader/preload.sh

var preloadPlugins = []plugin.Plugin{
	plugindms3ldgit.Plugins[0],
}
