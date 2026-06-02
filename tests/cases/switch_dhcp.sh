#!/bin/bash
source tools/auto.sh

show_topology() {
  cat <<'EOF'
# Topology:
# - Docker mgmt network: 172.245.0.0/24
#   sw1=172.245.0.241.
# - OpenLAN service network "example": 192.67.0.0/24
#   sw1=192.67.0.1, DHCP range 192.67.0.100-192.67.0.120.
# - In-container DHCP client namespace:
#   veth-dhcp is attached to Linux bridge br-example, while dnsmasq listens on hi-example.
# - Access point DHCP client:
#   tests-sw-dhcp.ac1 joins the mgmt network, bridges its tunnel to br-access-dhcp,
#   then uses dhclient on that bridge after access login succeeds.
# Validation:
#   DHCP config starts dnsmasq on the OpenLAN bridge, both a namespace client and an
#   access point bridge receive leases, and reload preserves DHCP service/configuration.

EOF
}

# OpenLAN Switch UT: DHCP service on network bridge.

export net_name=tests-net-dhcp
export sw1_name=tests-sw-dhcp
export ac1_name=tests-sw-dhcp.ac1
export bridge_device=hi-example
export bridge_master=br-example
export dhcp_ns=dhcp-client
export dhcp_client=veth-client
export dhcp_bridge=veth-dhcp
export access_user=t1@example
export access_pass=123456
export dhcp_start=192.67.0.100
export dhcp_end=192.67.0.120
export dhcp_gateway=192.67.0.1
export dhcp_dns1=114.114.114.114
export dhcp_dns2=8.8.8.8
export dhcp_conf=/var/openlan/dhcp/example.conf
export dhcp_lease=/var/openlan/dhcp/example.leases

setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.245.0.0/24 --gateway=172.245.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.245.0.241

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

  assert_cmd docker exec $name sh -c "cat > /tmp/example-dhcp.yaml <<EOF
name: example
bridge:
  address: 192.67.0.1/24
EOF"
  assert_cmd docker exec $name openlan network add --file /tmp/example-dhcp.yaml
  assert_cmd docker exec $name openlan user add --name $access_user --password $access_pass
  assert_cmd docker exec $name openlan network --name example dhcp disable
  assert_unmatch 10 "docker exec $name test -f $dhcp_conf && cat $dhcp_conf" "dhcp-range="
  assert_match 10 "docker exec $name openlan network --name example" "dhcp: disable"
  assert_cmd docker exec $name openlan network --name example dhcp enable --start $dhcp_start --end $dhcp_end --gateway $dhcp_gateway --dns $dhcp_dns1,$dhcp_dns2
  assert_match 10 "docker exec $name openlan network --name example" "dhcp: enable"
}

test_dhcp_service() {
  assert_match 20 "docker exec $sw1_name cat $dhcp_conf" "interface=$bridge_device"
  assert_match 20 "docker exec $sw1_name cat $dhcp_conf" "dhcp-range=$dhcp_start,$dhcp_end,12h"
  assert_match 20 "docker exec $sw1_name cat $dhcp_conf" "dhcp-leasefile=$dhcp_lease"
  assert_match 20 "docker exec $sw1_name cat $dhcp_conf" "dhcp-option=3,$dhcp_gateway"
  assert_match 20 "docker exec $sw1_name cat $dhcp_conf" "dhcp-option=6,$dhcp_dns1,$dhcp_dns2"
  assert_match 20 "docker exec $sw1_name pgrep -f 'dnsmasq.*example.conf'" "[0-9]"
}

setup_dhcp_client_ns() {
  assert_cmd docker exec $sw1_name ip netns add $dhcp_ns
  assert_cmd docker exec $sw1_name ip link add $dhcp_bridge type veth peer name $dhcp_client
  assert_cmd docker exec $sw1_name ip link show $bridge_master
  assert_cmd docker exec $sw1_name ip link set $dhcp_bridge master $bridge_master
  assert_cmd docker exec $sw1_name ip link set $dhcp_bridge up
  assert_cmd docker exec $sw1_name ip link set $dhcp_client netns $dhcp_ns
  assert_cmd docker exec $sw1_name ip netns exec $dhcp_ns ip link set lo up
  assert_cmd docker exec $sw1_name ip netns exec $dhcp_ns ip link set $dhcp_client up
}

test_dhcp_lease() {
  assert_cmd docker exec $sw1_name sh -c "command -v dhclient >/dev/null"
  assert_cmd docker exec $sw1_name ip netns exec $dhcp_ns dhclient -4 -1 $dhcp_client
  assert_fuzzy 10 "docker exec $sw1_name ip netns exec $dhcp_ns ip -4 addr show dev $dhcp_client" "192\\.67\\.0\\.(10[0-9]|11[0-9]|120)/24"
  assert_match 10 "docker exec $sw1_name cat $dhcp_lease" "192.67.0."
}

setup_access_client() {
  mkdir -p /opt/openlan/$ac1_name/etc/openlan
  cat > /opt/openlan/$ac1_name/etc/openlan/access.yaml <<EOF
protocol: tcp
connection: 172.245.0.241
username: $access_user
password: $access_pass
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
EOF
  start_access $ac1_name $net_name
  assert_expect 30 "docker logs -f $ac1_name" "Worker.OnSuccess"
  assert_cmd docker exec $ac1_name sh -c "command -v dhclient >/dev/null"
  assert_expect 30 "docker logs -f $ac1_name" "Access.OnTap"
}

test_access_dhcp_lease() {
  local tap_device
  local veth_ip
  tap_device=$(docker exec $ac1_name sh -c "ip -o link show | awk -F': ' '/^[0-9]+: (tap|ol|tun)/ {print \$2; exit}'")
  if [[ -z "$tap_device" ]]; then
    echo "failed to find access tap device" >&2
    exit 1
  fi
  veth_ip=$(docker exec $sw1_name sh -c "ip netns exec $dhcp_ns ip -o -4 addr show dev $dhcp_client | awk '{print \$4}' | cut -d/ -f1")
  if [[ -z "$veth_ip" ]]; then
    echo "failed to find dhcp client address on $dhcp_client" >&2
    exit 1
  fi
  assert_cmd docker exec $ac1_name dhclient -4 -1 $tap_device
  assert_fuzzy 10 "docker exec $ac1_name ip -4 addr show dev $tap_device" "192\\.67\\.0\\.(10[0-9]|11[0-9]|120)/24"
  assert_match 10 "docker exec $ac1_name ping -c 3 $veth_ip" "bytes from $veth_ip"
  assert_match 10 "docker exec $sw1_name cat $dhcp_lease" "192.67.0."
}

test_reload_persistence() {
  assert_cmd docker exec $sw1_name openlan reload --save
  assert_match 20 "docker exec $sw1_name cat $dhcp_conf" "dhcp-range=$dhcp_start,$dhcp_end,12h"
  assert_match 20 "docker exec $sw1_name pgrep -f 'dnsmasq.*example.conf'" "[0-9]"
  assert_match 10 "docker exec $sw1_name openlan network --name example" "dhcp: enable"
}

test_dhcp_disable() {
  assert_cmd docker exec $sw1_name openlan network --name example dhcp disable
  assert_unmatch 10 "docker exec $sw1_name pgrep -f 'dnsmasq.*example.conf'" "[0-9]"
  assert_unmatch 10 "docker exec $sw1_name test -f $dhcp_conf && cat $dhcp_conf" "dhcp-range="
  assert_match 10 "docker exec $sw1_name openlan network --name example" "dhcp: disable"
}

setup_topology() {
  setup_net
  setup_sw1
  test_dhcp_service
  setup_dhcp_client_ns
  test_dhcp_lease
  setup_access_client
  test_access_dhcp_lease
  test_reload_persistence
  test_dhcp_disable
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
