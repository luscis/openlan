#!/bin/bash
source tools/auto.sh

show_topology() {
  cat <<'EOF'
# Topology:
# - Docker mgmt network: 172.242.0.0/24
#   sw1=172.242.0.241, sw2=172.242.0.242.
# - OpenLAN service network "example": 192.63.0.0/24
#   sw1=192.63.0.1, sw2=192.63.0.2.
# - Both service network L3 devices are enslaved to VRF "vrf-example".
# - Forwarding link:
#   sw2 -> sw1 over UDP output.
# Validation:
#   (see scenario assertions in this case)

EOF
}


# OpenLAN Switch UT: network namespace/VRF path.

export net_name=tests-net-namespace
export sw1_name=tests-sw-namespace1
export sw2_name=tests-sw-namespace2
export vrf_name=vrf-example


setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.242.0.0/24 --gateway=172.242.0.1 >/dev/null
}

setup_switch_config() {
  local name=$1

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  cat > /opt/openlan/$name/etc/openlan/switch/switch.json <<EOF
{
  "protocol": "udp",
  "crypt": {
    "algorithm": "aes-128",
    "secret": "ea64d5b0c96c"
  }
}
EOF
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.242.0.241

  setup_switch_config $name
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name example add --address 192.63.0.1/24 --namespace $vrf_name
  assert_match 1 "docker exec $name openlan network --name example" "namespace: $vrf_name"
  assert_cmd docker exec $name ip link show $vrf_name
  assert_match 5 "docker exec $name ip link show hi-example" "master $vrf_name"
  assert_cmd docker exec $name openlan user add --name t1@example --password 123456
}

setup_sw2() {
  local name="$sw2_name"
  local address=172.242.0.242

  setup_switch_config $name
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name example add --address 192.63.0.2/24 --namespace $vrf_name
  assert_match 1 "docker exec $name openlan network --name example" "namespace: $vrf_name"
  assert_cmd docker exec $name ip link show $vrf_name
  assert_match 5 "docker exec $name ip link show hi-example" "master $vrf_name"
  assert_cmd docker exec $name openlan network --name example output add --remote 172.242.0.241 --protocol udp --secret t1:123456 --crypt aes-128:ea64d5b0c96c
}

test_vrf_ping() {
  assert_match 15 "docker exec $sw2_name openlan network --name example output ls" "state: authenticated"
  assert_match 20 "docker exec $sw2_name ip vrf exec $vrf_name ping -c 3 192.63.0.1" "bytes from"
}

test_reload_persistence() {
  assert_cmd docker exec $sw1_name openlan reload --save
  assert_cmd docker exec $sw2_name openlan reload --save

  assert_match 10 "docker exec $sw1_name openlan network --name example" "namespace: $vrf_name"
  assert_match 10 "docker exec $sw2_name openlan network --name example" "namespace: $vrf_name"
  assert_match 10 "docker exec $sw1_name ip link show hi-example" "master $vrf_name"
  assert_match 10 "docker exec $sw2_name ip link show hi-example" "master $vrf_name"
  test_vrf_ping
}

setup_topology() {
  setup_net
  setup_sw1
  setup_sw2
}

setup() {
  setup_topology
  test_vrf_ping
  test_reload_persistence
}

case "$1" in
  --topology)
    show_topology
    ;;
  *)
    main
    ;;
esac
