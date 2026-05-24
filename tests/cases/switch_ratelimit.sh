# OpenLAN RateLimit UT.

net_name=tests-net-ratelimit
sw1_name=tests-sw-ratelimit
bridge_device=hi-example
openvpn_device=tun1194

# Topology:
# - Docker mgmt network: 172.253.0.0/24
#   sw1=172.253.0.241.
# - OpenLAN service network "example": 192.60.0.0/24
#   sw1=192.60.0.1.
# - OpenVPN overlay:
#   tcp/1194, subnet 10.60.0.0/24.
# - Validation: add/update/remove ratelimit on the OpenLAN bridge and OpenVPN
#   devices, and verify Linux tc qdisc/filter state is updated.

setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.253.0.0/24 --gateway=172.253.0.1
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.253.0.241

  start_switch $name $net_name $address
  wait "docker logs -f $name" Http.Start 30

  docker exec $name openlan network --name example add --address 192.60.0.1/24
  docker exec $name ip link show $bridge_device

  docker exec $name openlan network --name example openvpn add --listen :1194 --protocol tcp --subnet 10.60.0.0/24 --dns 8.8.8.8
  check "docker exec $name ip link show $openvpn_device" "$openvpn_device" 20
}

test_ratelimit_add() {
  local device=$1

  docker exec $sw1_name openlan ratelimit add --device $device --speed 1

  check "docker exec $sw1_name tc qdisc show dev $device" "rate 1Mbit" 10
  check "docker exec $sw1_name tc filter show dev $device parent ffff:" "rate 1Mbit" 10

  docker exec $sw1_name openlan ratelimit add --device $device --speed 2

  check "docker exec $sw1_name tc qdisc show dev $device" "rate 2Mbit" 10
  check "docker exec $sw1_name tc filter show dev $device parent ffff:" "rate 2Mbit" 10

  if check "docker exec $sw1_name tc qdisc show dev $device" "rate 1Mbit" 3; then
    echo "unexpected root tbf qdisc still uses 1Mbit on $device after ratelimit add 2Mbit"
    return 1
  fi

  if check "docker exec $sw1_name tc filter show dev $device parent ffff:" "rate 1Mbit" 3; then
    echo "unexpected ingress filter still uses 1Mbit on $device after ratelimit add 2Mbit"
    return 1
  fi
}

test_ratelimit_remove() {
  local device=$1

  docker exec $sw1_name openlan ratelimit remove --device $device

  if check "docker exec $sw1_name tc qdisc show dev $device" "tbf" 3; then
    echo "unexpected root tbf qdisc remains on $device after ratelimit remove"
    return 1
  fi

  if check "docker exec $sw1_name tc qdisc show dev $device" "ingress" 3; then
    echo "unexpected ingress qdisc remains on $device after ratelimit remove"
    return 1
  fi
}

setup() {
  setup_net
  setup_sw1
  test_ratelimit_add $bridge_device
  test_ratelimit_remove $bridge_device
  test_ratelimit_add $openvpn_device
  test_ratelimit_remove $openvpn_device
}

main
