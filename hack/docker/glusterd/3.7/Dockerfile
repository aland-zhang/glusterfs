FROM appscode/base:8.6

ENV PVROOT /storage

RUN set -x \
  && apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates wget \
  && wget -O - http://download.gluster.org/pub/gluster/glusterfs/3.7/LATEST/rsa.pub | apt-key add - \
  && echo deb http://download.gluster.org/pub/gluster/glusterfs/3.7/LATEST/Debian/jessie/apt jessie main > /etc/apt/sources.list.d/gluster.list \
  && apt-get update \
  && apt-get install -y --no-install-recommends glusterfs-server attr \
  && apt-get purge -y --auto-remove ca-certificates wget \
  && rm -rf /var/lib/apt/lists/* /usr/share/doc /usr/share/man /tmp/* \
  && mkdir -p /var/run

ADD sv /etc/sv/
RUN ln -s /etc/sv /etc/service

COPY runit.sh /runit.sh
ENTRYPOINT ["/runit.sh"]
EXPOSE 111 24007 49152 49153 49154 49155 38465 38466 38467 2049 1110
