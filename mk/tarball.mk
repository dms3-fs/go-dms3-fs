

ifeq (,$(wildcard .tarball))
tarball-is:=0
else
tarball-is:=1
# override git hash
git-hash:=$(shell cat .tarball)
endif


go-dms3-fs-source.tar.gz: distclean
	bin/maketarball.sh $@
