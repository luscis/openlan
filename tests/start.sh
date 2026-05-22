#!/bin/bash

pushd $(dirname $0) > /dev/null

source ./macro.sh

shopt -s nullglob
scenarios=( cases/*.sh )
shopt -u nullglob

scenario_description() {
  local file=$1
  case "$file" in
    access_success.sh) echo "two access clients authenticate and can communicate" ;;
    access_fail.sh) echo "reject client authentication with wrong password" ;;
    access_admin_multi_login.sh) echo "admin user can login concurrently from multiple access clients" ;;
    access_same_user_mutex.sh) echo "same user multiple access logins are mutually exclusive" ;;
    access_openvpn.sh) echo "add/remove OpenVPN and validate cipher negotiation" ;;
    access_openvpn_client_ping.sh) echo "two OpenVPN clients with static addresses can ping each other" ;;
    access_openvpn_snat_vip.sh) echo "openvpn client reaches sw2 vip through sw1 snat" ;;
    access_snat_scope_matrix.sh) echo "verify snat scope matrix for openvpn, network a access, and network b access" ;;
    switch_bgp.sh) echo "verify bgp peer establishment and prefix filter persistence" ;;
    switch_dnat.sh) echo "verify dnat add-list-remove and nat table rule updates" ;;
    switch_findhop.sh) echo "verify findhop route binding, remove guard, and reload state" ;;
    switch_ztrust.sh) echo "verify ztrust enable/disable with guest and knock controls" ;;
    switch_tcp.sh) echo "build two switches and verify tcp output connectivity" ;;
    switch_udp.sh) echo "build two switches and verify udp output connectivity" ;;
    switch_ipsec_vxlan.sh) echo "build two switches and verify ipsec vxlan output connectivity" ;;
    switch_ipsec_gre.sh) echo "build two switches and verify ipsec gre output connectivity" ;;
    switch_route3.sh) echo "validate forwarding and route reachability via sw2" ;;
    *) echo "custom scenario" ;;
  esac
}

print_scenario_header() {
  local file=$1
  local name
  name=$(basename "$file")
  name=${name%.sh}
  local desc
  desc=$(scenario_description "$(basename "$file")")

  printf "==> %-17s : %s\n" "$name" "$desc"
}

run_scenario() {
  set +x
  local key=$1
  local file
  local name

  for file in "${scenarios[@]}"; do
    name=$(basename "$file")
    name=${name%.sh}
    if [[ "$key" == "$name" || "$key" == "$file" || "$key" == "${name}.sh" ]]; then
      echo "========================================"
      printf "[RUN ] %s\n" "$name"
      echo "----------------------------------------"
      print_scenario_header "$file"
      set -x
      source "$file"
      set +x
      printf "[PASS] %s\n" "$name"
      echo "========================================"
      return 0
    fi
  done

  echo "========================================"
  printf "[FAIL] unknown scenario: %s\n" "$key"
  echo "Available scenarios:"
  for file in "${scenarios[@]}"; do
    name=$(basename "$file")
    echo "  - ${name%.sh}"
  done
  echo "========================================"
  return 1
}

run_all() {
  set +x
  run_batch "OpenLAN test run started" "${scenarios[@]}"
}

run_batch() {
  set +x
  local title=$1
  shift
  local total=$#
  local count=0
  local index=0
  local key
  local file
  local found
  local name
  local scenario=""

  echo "========================================"
  printf "%s: %d scenario(s)\n" "$title" "$total"
  echo "========================================"

  for key in "$@"; do
    found=0
    for file in "${scenarios[@]}"; do
      name=$(basename "$file")
      name=${name%.sh}
      if [[ "$key" == "$name" || "$key" == "$file" || "$key" == "${name}.sh" ]]; then
        found=1
        index=$((index + 1))
        printf "\n[%02d/%02d] %s\n" "$index" "$total" "$name"
        echo "----------------------------------------"
        print_scenario_header "$file"
        set -x
        source "$file"
        set +x
        printf "[PASS] %s\n" "$name"
        count=$((count + 1))
        scenario="$scenario $name"
        break
      fi
    done

    if [[ $found -eq 0 ]]; then
      index=$((index + 1))
      printf "\n[%02d/%02d] %s\n" "$index" "$total" "$key"
      echo "----------------------------------------"
      printf "[FAIL] unknown scenario: %s\n" "$key"
      echo "Available scenarios:"
      for file in "${scenarios[@]}"; do
        name=$(basename "$file")
        echo "  - ${name%.sh}"
      done
      echo "========================================"
      return 1
    fi
  done

  echo
  echo "========================================"
  printf "Summary: %d/%d scenario(s) passed\n" "$count" "$total"
  echo "Passed scenarios:"
  echo "$scenario" | xargs printf "  - %s\n"
  echo "========================================"
}

run_selected() {
  set +x
  run_batch "OpenLAN selected run started" "$@"
}

if [[ $# -eq 0 ]]; then
  cleanup
  set -ex
  run_all
elif [[ "$1" == "list" || "$1" == "--list" || "$1" == "-l" ]]; then
  for file in "${scenarios[@]}"; do
    name=$(basename "$file")
    printf "%-17s : %s\n" "${name%.sh}" "$(scenario_description "$name")"
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
  if [[ $# -eq 1 ]]; then
    run_scenario "$1"
  else
    run_selected "$@"
  fi
fi

popd > /dev/null
