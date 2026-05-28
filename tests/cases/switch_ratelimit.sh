#!/bin/bash
source tools/auto.sh

show_topology() {
  cat <<'EOF'
# Topology:
# - Docker mgmt network: 172.253.0.0/24
#   sw1=172.253.0.241.
# - OpenLAN service network "example": 192.60.0.0/24
#   sw1=192.60.0.1.
# - OpenVPN overlay:
#   tcp/1194, subnet 10.60.0.0/24.
# Validation:
#   devices, and verify Linux tc qdisc/filter state is updated.

EOF
}

# OpenLAN RateLimit UT.

export net_name=tests-net-ratelimit
export sw1_name=tests-sw-ratelimit
export bridge_device=hi-example
export openvpn_device=tun1194


setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.253.0.0/24 --gateway=172.253.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.253.0.241

  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name example add --address 192.60.0.1/24
  assert_cmd docker exec $name ip link show $bridge_device

  assert_cmd docker exec $name openlan network --name example openvpn add --listen :1194 --protocol tcp --subnet 10.60.0.0/24 --dns 8.8.8.8
  assert_match 20 "docker exec $name ip link show $openvpn_device" "$openvpn_device"
}

test_ratelimit_add() {
  local device=$1

  assert_cmd docker exec $sw1_name openlan ratelimit add --device $device --speed 1

  assert_match 10 "docker exec $sw1_name tc qdisc show dev $device" "rate 1Mbit"
  assert_match 10 "docker exec $sw1_name tc filter show dev $device parent ffff:" "rate 1Mbit"

  assert_cmd docker exec $sw1_name openlan ratelimit add --device $device --speed 2

  assert_match 10 "docker exec $sw1_name tc qdisc show dev $device" "rate 2Mbit"
  assert_match 10 "docker exec $sw1_name tc filter show dev $device parent ffff:" "rate 2Mbit"

  assert_unmatch 3 "docker exec $sw1_name tc qdisc show dev $device" "rate 1Mbit"
  assert_unmatch 3 "docker exec $sw1_name tc filter show dev $device parent ffff:" "rate 1Mbit"
}

test_ratelimit_remove() {
  local device=$1

  assert_cmd docker exec $sw1_name openlan ratelimit remove --device $device

  assert_unmatch 3 "docker exec $sw1_name tc qdisc show dev $device" "tbf"
  assert_unmatch 3 "docker exec $sw1_name tc qdisc show dev $device" "ingress"
}

setup_topology() {
  setup_net
  setup_sw1
  test_ratelimit_add $bridge_device
  test_ratelimit_remove $bridge_device
  test_ratelimit_add $openvpn_device
  test_ratelimit_remove $openvpn_device
}

setup() {
  setup_topology
}

case "$1" in
  --topology)
    show_topology
    ;;
  *)
    main
    ;;
esac
