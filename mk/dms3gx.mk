dms3gx-path = dms3gx/dms3fs/$(shell dms3gx deps find $(1))/$(1)

dms3gx-deps:
	dms3gx install --global
.PHONY: dms3gx-deps

ifneq ($(DMS3FS_GX_USE_GLOBAL),1)
dms3gx-deps: bin/dms3gx bin/dms3gx-go
endif
.PHONY: dms3gx-deps

ifeq ($(tarball-is),0)
DEPS_GO += dms3gx-deps
endif
