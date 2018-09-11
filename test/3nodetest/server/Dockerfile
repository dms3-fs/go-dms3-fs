FROM zaqwsx_dms3fs-test-img

RUN dms3fs init -b=1024
ADD . /tmp/test
RUN mv -f /tmp/test/config /root/.dms3-fs/config
RUN dms3fs id
RUN chmod +x /tmp/test/run.sh

EXPOSE 4021 4022/udp

ENV DMS3FS_PROF true
ENV DMS3FS_LOGGING_FMT nocolor

ENTRYPOINT ["/bin/bash"]
CMD ["/tmp/test/run.sh"]
