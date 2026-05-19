#!/bin/bash

pushd $(dirname $0)

shopt -s nullglob
scenarios=( _*.sh )
shopt -u nullglob

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

run_scenario() {
  local key=$1
  local file
  local name

  for file in "${scenarios[@]}"; do
    name=${file#_}
    name=${name%.sh}
    if [[ "$key" == "$name" || "$key" == "$file" || "$key" == "_${name}" || "$key" == "${name}.sh" ]]; then
      source "$file"
      return 0
    fi
  done

  echo "unknown scenario: $key"
  echo "available scenarios:"
  for file in "${scenarios[@]}"; do
    name=${file#_}
    echo "  - ${name%.sh}"
  done
  return 1
}

run_all() {
  local file
  for file in "${scenarios[@]}"; do
    source "$file"
  done
}

if [[ $# -eq 0 ]]; then
  cleanup
  set -ex
  run_all
elif [[ "$1" == "list" || "$1" == "--list" || "$1" == "-l" ]]; then
  for file in "${scenarios[@]}"; do
    name=${file#_}
    echo "${name%.sh}"
  done
elif [[ "$1" == "help" || "$1" == "--help" || "$1" == "-h" ]]; then
  echo "usage: bash tests/start.sh [scenario ...]"
  echo "examples:"
  echo "  bash tests/start.sh"
  echo "  bash tests/start.sh switch_tcp"
  echo "  bash tests/start.sh switch_tcp switch_three_node"
  echo "  bash tests/start.sh --list"
else
  cleanup
  set -ex
  for scenario in "$@"; do
    run_scenario "$scenario"
  done
fi

popd
