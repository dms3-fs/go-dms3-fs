package loader

import (
	"github.com/dms3-fs/go-dms3-fs/core/coredag"
	"github.com/dms3-fs/go-dms3-fs/plugin"
	"github.com/opentracing/opentracing-go"

	dms3ld "github.com/dms3-fs/go-ld-format"
)

func initialize(plugins []plugin.Plugin) error {
	for _, p := range plugins {
		err := p.Init()
		if err != nil {
			return err
		}
	}

	return nil
}

func run(plugins []plugin.Plugin) error {
	for _, pl := range plugins {
		switch pl := pl.(type) {
		case plugin.PluginDMS3LD:
			err := runDMS3LDPlugin(pl)
			if err != nil {
				return err
			}
		case plugin.PluginTracer:
			err := runTracerPlugin(pl)
			if err != nil {
				return err
			}
		default:
			panic(pl)
		}
	}
	return nil
}

func runDMS3LDPlugin(pl plugin.PluginDMS3LD) error {
	err := pl.RegisterBlockDecoders(dms3ld.DefaultBlockDecoder)
	if err != nil {
		return err
	}
	return pl.RegisterInputEncParsers(coredag.DefaultInputEncParsers)
}

func runTracerPlugin(pl plugin.PluginTracer) error {
	tracer, err := pl.InitTracer()
	if err != nil {
		return err
	}
	opentracing.SetGlobalTracer(tracer)
	return nil
}
