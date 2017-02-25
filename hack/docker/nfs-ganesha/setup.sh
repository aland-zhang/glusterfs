#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(dirname "${BASH_SOURCE}")/../..
source "$REPO_ROOT/libbuild/common/lib.sh"
source "$REPO_ROOT/libbuild/common/public_image.sh"

IMG=${PWD##*/}
TAG=2.3.0
PRIVILEGED_CONTAINER='--privileged=true'

docker_sh() {
	name=$IMG-$(date +%s | sha256sum | base64 | head -c 8 ; echo)
	privileged="${PRIVILEGED_CONTAINER:-}"
	local cmd="systemctl stop rpcbind.service"
	echo $cmd; $cmd
	local cmd=$(cat <<EOM
docker run -d \
  -p 1110:1110/tcp -p 1110:1110/udp \
  -p 111:111/tcp -p 111:111/udp \
  -p 2049:2049/tcp -p 2049:2049/udp \
  -p 24007:24007 \
  -p 49152-49155:49152-49155 \
  -p 38465-38467:38465-38467 \
  -it $privileged --name=$name appscode/$IMG:$TAG
EOM
	)
	echo $cmd; $cmd
	cmd="docker exec -it $name bash"
	echo $cmd; $cmd
}

binary_repo $@
