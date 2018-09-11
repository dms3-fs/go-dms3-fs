FROM golang:1.10-stretch
MAINTAINER <support@dms3.io>

# There is a copy of this Dockerfile called Dockerfile.fast,
# which is optimized for build time, instead of image size.
#
# Please keep these two Dockerfiles in sync.

ENV GX_DMS3-FS ""
ENV SRC_DIR /go/src/github.com/dms3-fs/go-dms3-fs

COPY . $SRC_DIR

# Build the thing.
# Also: fix getting HEAD commit hash via git rev-parse.
# Also: allow using a custom DMS3-FS API endpoint.
RUN cd $SRC_DIR \
  && mkdir .git/objects \
  && ([ -z "$GX_DMS3-FS" ] || echo $GX_DMS3-FS > /root/.dms3-fs/api) \
  && make build

# Get su-exec, a very minimal tool for dropping privileges,
# and tini, a very minimal init daemon for containers
ENV SUEXEC_VERSION v0.2
ENV TINI_VERSION v0.16.1
RUN set -x \
  && cd /tmp \
  && git clone https://github.com/ncopa/su-exec.git \
  && cd su-exec \
  && git checkout -q $SUEXEC_VERSION \
  && make \
  && cd /tmp \
  && wget -q -O tini https://github.com/krallin/tini/releases/download/$TINI_VERSION/tini \
  && chmod +x tini

# Get the TLS CA certificates, they're not provided by busybox.
RUN apt-get update && apt-get install -y ca-certificates

# Now comes the actual target image, which aims to be as small as possible.
FROM busybox:1-glibc
MAINTAINER support@dms3.io>

# Get the dms3-fs binary, entrypoint script, and TLS CAs from the build container.
ENV SRC_DIR /go/src/github.com/dms3-fs/go-dms3-fs
COPY --from=0 $SRC_DIR/cmd/dms3-fs/dms3-fs /usr/local/bin/dms3-fs
COPY --from=0 $SRC_DIR/bin/container_daemon /usr/local/bin/start_dms3fs
COPY --from=0 /tmp/su-exec/su-exec /sbin/su-exec
COPY --from=0 /tmp/tini /sbin/tini
COPY --from=0 /etc/ssl/certs /etc/ssl/certs

# This shared lib (part of glibc) doesn't seem to be included with busybox.
COPY --from=0 /lib/x86_64-linux-gnu/libdl-2.24.so /lib/libdl.so.2

# Ports for Swarm TCP, Swarm uTP, API, Gateway, Swarm Websockets
EXPOSE 4001
EXPOSE 4002/udp
EXPOSE 5001
EXPOSE 8080
EXPOSE 8081

# Create the fs-repo directory and switch to a non-privileged user.
ENV DMS3-FS_PATH /data/dms3-fs
RUN mkdir -p $DMS3-FS_PATH \
  && adduser -D -h $DMS3-FS_PATH -u 1000 -G users dms3-fs \
  && chown dms3-fs:users $DMS3-FS_PATH

# Expose the fs-repo as a volume.
# start_dms3fs initializes an dms3fs-repo if none is mounted.
# Important this happens after the USER directive so permission are correct.
VOLUME $DMS3-FS_PATH

# The default logging level
ENV DMS3-FS_LOGGING ""

# This just makes sure that:
# 1. There's an fs-repo, and initializes one if there isn't.
# 2. The API and Gateway are accessible from outside the container.
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/start_dms3fs"]

# Execute the daemon subcommand by default
CMD ["daemon", "--migrate=true"]
