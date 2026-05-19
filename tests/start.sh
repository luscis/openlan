#!/bin/bash

set -ex

pushd $(dirname $0)

cleanup() {
  local containers
  local networks

  containers=$(docker ps -aq --filter "name=^tests-" 2>/dev/null)
  if [[ -n "$containers" ]]; then
    docker rm -f $containers
  fi
  networks=$(docker network ls -q --filter "name=^tests-" 2>/dev/null)
  if [[ -n "$networks" ]]; then
    docker network rm $networks
  fi
  rm -rf /opt/openlan/tests-*
}

cleanup

source access_success.sh
source switch_tcp.sh
source switch_udp.sh
source openvpn.sh
source access_fail.sh

popd
