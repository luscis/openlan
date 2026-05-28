#!/bin/bash
source tools/auto.sh

show_topology() {
  cat <<'EOF'
# Topology:
# - Docker mgmt network: 172.253.0.0/24
# - OpenLAN service network "example": 192.42.0.0/24
# - OpenVPN overlay: tcp/1194, subnet 10.97.0.0/24
# - Static OpenVPN client addresses:
#   vpn1@example -> 10.97.0.10
#   vpn2@example -> 10.97.0.11
# Validation:
#   (see scenario assertions in this case)

EOF
}


# OpenLAN OpenVPN client-to-client ping with static addresses.

export net_name=tests-net-openvpn-ping
export sw1_name=tests-sw-openvpn-ping
export vpn1_name=tests-sw-openvpn-ping.vpn1
export vpn2_name=tests-sw-openvpn-ping.vpn2


setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.253.0.0/24 --gateway=172.253.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.253.0.241

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name example add --address 192.42.0.1/24
  assert_cmd docker exec $name openlan user add --name vpn1@example --password 123456
  assert_cmd docker exec $name openlan user add --name vpn2@example --password 123456
}

start_vpn_client() {
  local server_name=$1
  local client_name=$2
  local username=$3
  local password=$4

  mkdir -p /opt/openlan/$client_name/ovpn
  docker cp $server_name:/var/openlan/openvpn/example/tcp1194client.ovpn /opt/openlan/$client_name/ovpn/client.ovpn
  cat > /opt/openlan/$client_name/ovpn/auth.txt <<EOF
$username
$password
EOF

  start_openvpn $client_name $net_name
  assert_expect 40 "docker logs -f $client_name" "Initialization Sequence Completed"
}

setup_openvpn_and_check_ping() {
  local name="$sw1_name"

  assert_cmd docker exec $name openlan network --name example openvpn add --listen :1194 --protocol tcp --subnet 10.97.0.0/24 --dns 8.8.8.8

  assert_cmd docker exec $name openlan network --name example client add --user vpn1 --address 10.97.0.10
  assert_cmd docker exec $name openlan network --name example client add --user vpn2 --address 10.97.0.11

  assert_cmd docker exec $name test -f /var/openlan/openvpn/example/ccd/vpn1@example
  assert_cmd docker exec $name test -f /var/openlan/openvpn/example/ccd/vpn2@example

  start_vpn_client $name $vpn1_name vpn1@example 123456
  start_vpn_client $name $vpn2_name vpn2@example 123456

  assert_check 5 "docker exec $vpn1_name ping -c 3 10.97.0.11" "bytes from"
  assert_check 5 "docker exec $vpn2_name ping -c 3 10.97.0.10" "bytes from"
}

setup_topology() {
  setup_net
  setup_sw1
  setup_openvpn_and_check_ping
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
