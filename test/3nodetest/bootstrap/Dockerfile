FROM zaqwsx_dms3fs-test-img

RUN dms3fs init -b=1024
ADD . /tmp/id
RUN mv -f /tmp/id/config /root/.dms3-fs/config
RUN dms3fs id

ENV DMS3FS_PROF true
ENV DMS3FS_LOGGING_FMT nocolor

EXPOSE 4011 4012/udp
