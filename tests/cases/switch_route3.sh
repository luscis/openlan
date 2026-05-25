#!/bin/bash
source tools/auto.sh


# OpenLAN route test: 3-node forwarding (sw3 -> sw2 -> sw1).

export net_name=tests-net-route3
export sw1_name=tests-sw-route1
export sw2_name=tests-sw-route2
export sw3_name=tests-sw-route3

# Topology:
# - Docker mgmt network: 172.251.0.0/24
#   sw1=172.251.0.241, sw2=172.251.0.242, sw3=172.251.0.243.
# - OpenLAN service network "example": 192.51.0.0/24
#   sw1=192.51.0.1, sw2=192.51.0.2, sw3=192.51.0.3.
# - Loopback VIPs:
#   sw1 lo=10.251.0.11/32, sw2 lo=10.251.0.12/32.
# - Forwarding and route design:
#   sw2 -> sw1 output, sw3 -> sw2 output;
#   sw3 routes 10.251.0.11 via 192.51.0.1 and 10.251.0.12 via 192.51.0.2.
# - Validation path: sw3 reaches sw1/sw2 service IPs and both VIPs.

setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.251.0.0/24 --gateway=172.251.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.251.0.241

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  cat > /opt/openlan/$name/etc/openlan/switch/switch.json <<EOF
{
  "protocol": "tcp",
  "crypt": {
    "algorithm": "aes-128",
    "secret": "ea64d5b0c96c"
  }
}
EOF
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name example add --address 192.51.0.1/24
  assert_cmd docker exec $name openlan router address add --device lo --address 10.251.0.11/32
  assert_cmd docker exec $name openlan user add --name edge1@example --password 123456
}

setup_sw2() {
  local name="$sw2_name"
  local address=172.251.0.242

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  cat > /opt/openlan/$name/etc/openlan/switch/switch.json <<EOF
{
  "protocol": "tcp",
  "crypt": {
    "algorithm": "aes-128",
    "secret": "ea64d5b0c96c"
  }
}
EOF
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name example add --address 192.51.0.2/24
  assert_cmd docker exec $name openlan router address add --device lo --address 10.251.0.12/32
  assert_cmd docker exec $name openlan user add --name edge2@example --password 123457
  # Add a output to sw1 for 3-node forwarding validation.
  assert_cmd docker exec $name openlan network --name example output add --remote 172.251.0.241 --protocol tcp --secret edge1@example:123456 --crypt aes-128:ea64d5b0c96c
}

setup_sw3() {
  local name="$sw3_name"
  local address=172.251.0.243

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  cat > /opt/openlan/$name/etc/openlan/switch/switch.json <<EOF
{
  "protocol": "tcp",
  "crypt": {
    "algorithm": "aes-128",
    "secret": "ea64d5b0c96c"
  }
}
EOF
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"
  
  assert_cmd docker exec $name openlan network --name example add --address 192.51.0.3/24
  # Add outputs to sw2 for 3-node forwarding validation.
  assert_cmd docker exec $name openlan network --name example output add --remote 172.251.0.242 --protocol tcp --secret edge2@example:123457 --crypt aes-128:ea64d5b0c96c

  # Route VIP traffic via sw2 (192.51.0.2) for 3-node forwarding validation.
  assert_cmd docker exec $name openlan network --name example route add --prefix 10.251.0.11/32 --nexthop 192.51.0.1
  assert_cmd docker exec $name openlan network --name example route add --prefix 10.251.0.12/32 --nexthop 192.51.0.2
}

test_route_once() {
  assert_match 1 "docker exec $sw3_name ip route show" "10.251.0.11"
  assert_match 1 "docker exec $sw3_name ip route show" "10.251.0.12"

  assert_cmd docker exec $sw3_name ip neigh flush dev hi-example
  assert_match 25 "docker exec $sw3_name ping -c 3 192.51.0.1" "bytes from"
  assert_match 25 "docker exec $sw3_name ping -c 3 192.51.0.2" "bytes from"
  assert_match 25 "docker exec $sw3_name ping -c 3 10.251.0.11" "bytes from"
  assert_match 25 "docker exec $sw3_name ping -c 3 10.251.0.12" "bytes from"
}

test_route() {
  assert_match 60 "docker exec $sw2_name openlan network --name example output ls" "state: authenticated"
  assert_match 60 "docker exec $sw3_name openlan network --name example output ls" "state: authenticated"
  
  test_route_once

  assert_cmd docker exec $sw1_name openlan reload --save
  assert_cmd docker exec $sw2_name openlan reload --save
  assert_cmd docker exec $sw3_name openlan reload --save

  assert_match 60 "docker exec $sw2_name openlan network --name example output ls" "state: authenticated"
  assert_match 60 "docker exec $sw3_name openlan network --name example output ls" "state: authenticated"

  test_route_once
}

setup_topology() {
  setup_net
  setup_sw1
  setup_sw2
  setup_sw3
  test_route
}

setup() {
  setup_topology
}

main
