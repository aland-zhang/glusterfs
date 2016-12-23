#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -x

# ref: http://download.gluster.org/pub/gluster/glusterfs/LATEST/Debian/jessie/
wget -O - http://download.gluster.org/pub/gluster/glusterfs/3.7/3.7.8/pub.key | apt-key add -
echo deb http://download.gluster.org/pub/gluster/glusterfs/3.7/3.7.8/Debian/jessie/apt jessie main > /etc/apt/sources.list.d/gluster.list 
apt-get update
apt-get -y install glusterfs-server
apt-get -y install ufw vim ntp haveged fail2ban curl attr tree
apt-get -y autoremove
