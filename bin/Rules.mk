include mk/header.mk

dist_root_$(d)="/dms3fs/QmYpvspyyUWQTE226NFWteXYJF3x3br25xmB6XzEoqfzyv"

###$(d)/dms3gx: $(d)/dms3gx-v0.13.0
###$(d)/dms3gx-go: $(d)/dms3gx-go-v1.7.0
$(d)/dms3gx: $(d)/dms3gx-v0.12.1
$(d)/dms3gx-go: $(d)/dms3gx-go-v1.8.0

TGTS_$(d) := $(d)/dms3gx $(d)/dms3gx-go
DISTCLEAN += $(wildcard $(d)/dms3gx-v*) $(wildcard $(d)/dms3gx-go-v*) $(d)/tmp

PATH := $(realpath $(d)):$(PATH)

$(TGTS_$(d)):
	rm -f $@$(?exe)
ifeq ($(WINDOWS),1)
	cp $^$(?exe) $@$(?exe)
else
	ln -s $(notdir $^) $@
endif

bin/dms3gx-v%:
	@echo "installing dms3gx $(@:bin/dms3gx-%=%)"
	bin/dist_get $(dist_root_bin) dms3gx $@ $(@:bin/dms3gx-%=%)

bin/dms3gx-go-v%:
	@echo "installing dms3gx-go $(@:bin/dms3gx-go-%=%)"
	@bin/dist_get $(dist_root_bin) dms3gx-go $@ $(@:bin/dms3gx-go-%=%)

CLEAN += $(TGTS_$(d))
include mk/footer.mk
