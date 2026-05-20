#!/bin/bash

pushd $(dirname $0) > /dev/null

shopt -s nullglob
scenarios=( _*.sh )
shopt -u nullglob

scenario_description() {
  local file=$1
  case "$file" in
    _access_success.sh) echo "two access clients authenticate and can communicate" ;;
    _access_fail.sh) echo "reject client authentication with wrong password" ;;
    _access_openvpn.sh) echo "add/remove OpenVPN and validate cipher negotiation" ;;
    _switch_tcp.sh) echo "build two switches and verify tcp output connectivity" ;;
    _switch_udp.sh) echo "build two switches and verify udp output connectivity" ;;
    _switch_three_node.sh) echo "validate forwarding and route reachability via sw2" ;;
    *) echo "custom scenario" ;;
  esac
}

print_scenario_header() {
  local file=$1
  local name=${file#_}
  name=${name%.sh}
  local desc
  desc=$(scenario_description "$file")

  printf "==> %-17s : %s\n" "$name" "$desc"
}

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
      print_scenario_header "$file"
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
    print_scenario_header "$file"
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
    printf "%-17s : %s\n" "${name%.sh}" "$(scenario_description "$file")"
  done
elif [[ "$1" == "help" || "$1" == "--help" || "$1" == "-h" ]]; then
  echo "usage: bash tests/start.sh [scenario ...]"
  echo "examples:"
  echo "  bash tests/start.sh"
  echo "  bash tests/start.sh switch_tcp"
  echo "  bash tests/start.sh access_openvpn access_success"
  echo "  bash tests/start.sh --list"
else
  cleanup
  set -ex
  for scenario in "$@"; do
    run_scenario "$scenario"
  done
fi

popd > /dev/null
