#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(go env GOPATH)/src/github.com/appscode/glusterfs
source "$REPO_ROOT/hack/libbuild/common/lib.sh"
source "$REPO_ROOT/hack/libbuild/common/public_image.sh"

IMG=glusterd
TAG=3.8
PRIVILEGED_CONTAINER='--privileged=true'

binary_repo $@
