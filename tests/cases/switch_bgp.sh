# OpenLAN Switch UT: BGP global/neighbor/prefix flow.

net_name=tests-net-bgp
sw1_name=tests-sw-bgp1
sw2_name=tests-sw-bgp2

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
  docker network create $net_name --driver=bridge --subnet=172.244.0.0/24 --gateway=172.244.0.1
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.244.0.241

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  wait "docker logs -f $name" Http.Start 30

  docker exec $name openlan network --name a add --address 192.54.0.1/24
  docker exec $name openlan bgp enable --router-id 172.244.0.241 --local-as 65101
  docker exec $name openlan bgp neighbor add --address 172.244.0.242 --remote-as 65102
}

setup_sw2() {
  local name="$sw2_name"
  local address=172.244.0.242

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  wait "docker logs -f $name" Http.Start 30

  docker exec $name openlan network --name b add --address 192.55.0.1/24
  docker exec $name openlan bgp enable --router-id 172.244.0.242 --local-as 65102
  docker exec $name openlan bgp neighbor add --address 172.244.0.241 --remote-as 65101
}

setup_prefix_filters() {
  docker exec $sw1_name openlan bgp advertis add --neighbor 172.244.0.242 --prefix 192.54.0.0/24
  docker exec $sw1_name openlan bgp receives add --neighbor 172.244.0.242 --prefix 192.55.0.0/24

  docker exec $sw2_name openlan bgp advertis add --neighbor 172.244.0.241 --prefix 192.55.0.0/24
  docker exec $sw2_name openlan bgp receives add --neighbor 172.244.0.241 --prefix 192.54.0.0/24
}

test_bgp_once() {
  check "docker exec $sw1_name openlan bgp ls" "state: established" 10

  check "docker exec $sw1_name ip route show" "192.55.0.0/24" 60
  check "docker exec $sw2_name ip route show" "192.54.0.0/24" 60
}

test_bgp() {
  test_bgp_once

  docker exec $sw1_name openlan reload --save
  docker exec $sw2_name openlan reload --save

  test_bgp_once
}

setup() {
  setup_net
  setup_sw1
  setup_sw2
  setup_prefix_filters
  test_bgp
}

main
