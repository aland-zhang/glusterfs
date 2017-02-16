#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -x

apt-get install -y --no-install-recommends \
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
    xfslibs-dev

    # dh-python \
    # debhelper \
    # pyqt4-dev-tools \
    # python-qt4 \

# ref: http://download.gluster.org/pub/gluster/glusterfs/LATEST/Debian/jessie/
wget -O - http://download.gluster.org/pub/gluster/glusterfs/3.7/3.7.8/pub.key | apt-key add -
echo deb http://download.gluster.org/pub/gluster/glusterfs/3.7/3.7.8/Debian/jessie/apt jessie main > /etc/apt/sources.list.d/gluster.list
apt-get update
apt-get -y --no-install-recommends install glusterfs-common glusterfs-client

rm -rf nfs-ganesha
clone https://github.com/nfs-ganesha/nfs-ganesha.git
cd nfs-ganesha
git submodule update --init
git checkout tags/V2.3.1

mkdir build
cd build
cmake \
	-DUSE_FSAL_GLUSTER=ON \
	-DUSE_FSAL_XFS=ON \
	-DUSE_GUI_ADMIN_TOOLS=OFF \
	-DUSE_DBUS=ON \
	-DCMAKE_BUILD_TYPE=Maintainer ../src
make
make install

# cd ~/nfs-ganesha/src
# dpkg-buildpackage -uc -us

# https://github.com/nfs-ganesha/nfs-ganesha/wiki/DBusExports
# https://github.com/nfs-ganesha/nfs-ganesha/wiki/Dbusinterface
# https://github.com/tianon/debian-golang-dbus
# https://github.com/docker/docker/issues/7459#issuecomment-158319306

# python2 /nfs-ganesha/src/scripts/ganeshactl/manage_exports.py add /etc/ganesha/ganesha.conf