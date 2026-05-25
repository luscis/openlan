#!/bin/bash
source tools/auto.sh

# OpenLAN Switch UT: BGP global/neighbor/prefix flow.

export net_name=tests-net-bgp
export sw1_name=tests-sw-bgp1
export sw2_name=tests-sw-bgp2

# Topology:
# - Docker mgmt network: 172.244.0.0/24
#   sw1=172.244.0.241, sw2=172.244.0.242.
# - OpenLAN service network "example": 192.54.0.0/24
#   sw1=192.54.0.1, sw2=192.54.0.2.
# - BGP design:
#   sw1 local-as 65101, router-id 172.244.0.241.
#   sw2 local-as 65102, router-id 172.244.0.242.
#   peers use mgmt addresses as neighbors.
# - Validation path:
#   BGP reaches established state, prefix filters are present, and
#   config persists after reload.

setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.244.0.0/24 --gateway=172.244.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.244.0.241

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name a add --address 192.54.0.1/24
  assert_cmd docker exec $name openlan bgp enable --router-id 172.244.0.241 --local-as 65101
  assert_cmd docker exec $name openlan bgp neighbor add --address 172.244.0.242 --remote-as 65102
}

setup_sw2() {
  local name="$sw2_name"
  local address=172.244.0.242

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name b add --address 192.55.0.1/24
  assert_cmd docker exec $name openlan bgp enable --router-id 172.244.0.242 --local-as 65102
  assert_cmd docker exec $name openlan bgp neighbor add --address 172.244.0.241 --remote-as 65101
}

setup_prefix_filters() {
  assert_cmd docker exec $sw1_name openlan bgp advertis add --neighbor 172.244.0.242 --prefix 192.54.0.0/24
  assert_cmd docker exec $sw1_name openlan bgp receives add --neighbor 172.244.0.242 --prefix 192.55.0.0/24

  assert_cmd docker exec $sw2_name openlan bgp advertis add --neighbor 172.244.0.241 --prefix 192.55.0.0/24
  assert_cmd docker exec $sw2_name openlan bgp receives add --neighbor 172.244.0.241 --prefix 192.54.0.0/24
}

test_bgp_once() {
  assert_match 10 "docker exec $sw1_name openlan bgp ls" "state: established"
  assert_match 60 "docker exec $sw1_name ip route show" "192.55.0.0/24"
  assert_match 60 "docker exec $sw2_name ip route show" "192.54.0.0/24"
}

test_bgp() {
  test_bgp_once

  assert_cmd docker exec $sw1_name openlan reload --save
  assert_cmd docker exec $sw2_name openlan reload --save

  test_bgp_once
}

setup_topology() {
  setup_net
  setup_sw1
  setup_sw2
  setup_prefix_filters
}

setup() {
  setup_topology
  test_bgp
}

main
