#!/bin/bash

set -ex

pushd $(dirname $0)

source access.sh
source switch.sh

popd