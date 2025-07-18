#!/bin/bash

set -ex

pushd $(dirname $0)

source access.ut
source switch.ut

popd