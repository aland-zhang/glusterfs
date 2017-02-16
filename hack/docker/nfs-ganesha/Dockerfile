FROM appscode/base:8.7

RUN set -x \
  && apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates wget \
  && wget -O - http://download.gluster.org/pub/gluster/glusterfs/3.7/3.7.8/pub.key | apt-key add - \
  && echo deb http://download.gluster.org/pub/gluster/glusterfs/3.7/3.7.8/Debian/jessie/apt jessie main > /etc/apt/sources.list.d/gluster.list \
  && apt-get update \
  && apt-get install -y --no-install-recommends \
    bison \
    build-essential automake autoconf libtool \
    cmake \
    dbus libdbus-1-3 libdbus-1-dev \
    flex \
    git \
    glusterfs-common \
    libattr1-dev \
    libcap-dev \
    libcephfs-dev \
    libcephfs1 \
    libjemalloc-dev \
    libkrb5-dev \
    libncurses5-dev \
    libnfsidmap-dev \
    libssl-dev \
    libwbclient-dev \
    nfs-common \
    quilt \
    rpcbind \
    uuid-dev \
    xfslibs-dev \
  && git clone https://github.com/nfs-ganesha/nfs-ganesha.git \
  && cd nfs-ganesha \
  && git submodule update --init \
  && git checkout tags/V2.3.1 \
  && mkdir build \
  && cd build \
  && cmake \
    -DUSE_FSAL_GLUSTER=ON \
    -DUSE_FSAL_XFS=ON \
    -DUSE_GUI_ADMIN_TOOLS=OFF \
    -DUSE_DBUS=ON \
    -DCMAKE_BUILD_TYPE=Maintainer ../src \
  && make \
  && make install \
  && mkdir -p /var/run/dbus \
  && cp /nfs-ganesha/src/scripts/ganeshactl/org.ganesha.nfsd.conf /etc/dbus-1/system.d \
  && mkdir -p /etc/ganesha/export.d \
  && echo '%include "/etc/ganesha/export.d/*.conf"' > /etc/ganesha/ganesha.conf \
  && apt-get purge -y --auto-remove \
    build-essential automake autoconf libtool \
    ca-certificates \
    wget \
  && apt-get purge -y \
    libcap-dev \
    libcephfs-dev \
    libdbus-1-dev \
    libjemalloc-dev \
    libkrb5-dev \
    libncurses5-dev \
    libnfsidmap-dev \
    libssl-dev \
    libwbclient-dev \
    xfslibs-dev \
  && rm -rf /var/lib/apt/lists/* /usr/share/doc /usr/share/man /tmp/*

COPY manage_exports.sh /manage_exports.sh

COPY sv /etc/sv
RUN ln -s /etc/sv /etc/service

ENTRYPOINT ["/runit.sh"]
EXPOSE 111 24007 49152 49153 49154 49155 38465 38466 38467 2049 1110
