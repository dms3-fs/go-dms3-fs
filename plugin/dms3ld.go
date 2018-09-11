package plugin

import (
	"github.com/dms3-fs/go-dms3-fs/core/coredag"

	dms3ld "github.com/dms3-fs/go-ld-format"
)

// PluginDMS3LD is an interface that can be implemented to add handlers for
// for different DMS3LD formats
type PluginDMS3LD interface {
	Plugin

	RegisterBlockDecoders(dec dms3ld.BlockDecoder) error
	RegisterInputEncParsers(iec coredag.InputEncParsers) error
}
