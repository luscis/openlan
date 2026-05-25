#!/bin/bash

_version() {
  ../dist/version.sh
}

export ARCH=amd64
export VERSION=$(_version)
export IMAGE="luscis/openlan:$VERSION.$ARCH.deb"


source ./tools/color.sh
source ./tools/common.sh
source ./tools/report.sh