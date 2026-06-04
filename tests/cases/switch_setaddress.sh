#!/bin/bash
source tools/auto.sh

show_description() {
  echo "verify openvpn pushed route follows bridge address update"
}

show_topology_summary() {
  cat <<'EOF'
sw1 changes example bridge address from 192.72.0.1/24 to 192.73.0.1/24 and openvpn restart pushes the new route range
EOF
}

show_topology() {
  cat <<'EOF'
# Topology:
# - Diagram:
#       sw1 example bridge 192.72.0.1/24 -> 192.73.0.1/24
#       openvpn restart refreshes pushed route 192.72.0.0/24 -> 192.73.0.0/24
# - Docker mgmt network: 100.100.0.0/24
#   sw1=100.100.0.241.
# - OpenLAN service network "example": 192.72.0.0/24
#   initial bridge address 192.72.0.1/24, updated bridge address 192.73.0.1/24.
# Validation:
#   After `network address add` changes the bridge address and OpenVPN restarts,
#   the generated OpenVPN server config pushes the new bridge route range.

EOF
}

export net_name=tests-net-setaddress
export sw1_name=tests-sw-setaddress

setup_net() {
  docker network create $net_name --driver=bridge --subnet=100.100.0.0/24 --gateway=100.100.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=100.100.0.241

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

  assert_cmd docker exec $name openlan network --name example add --address 192.72.0.1/24
  assert_cmd docker exec $name openlan network --name example openvpn add --listen :1194 --protocol tcp --subnet 10.72.0.0/24
  assert_cmd docker exec $name openlan network --name example snat enable
}

test_setaddress_openvpn_route() {
  assert_match 10 "docker exec $sw1_name cat /var/openlan/openvpn/example/tcp1194server.conf" 'push "route 192.72.0.0 255.255.255.0"'
  assert_match 10 "docker exec $sw1_name iptables -t nat -S TT_example_SNAT" "192.72.0.0/24"
  assert_cmd docker exec $sw1_name openlan network --name example address add --address 192.73.0.1/24
  assert_match 10 "docker exec $sw1_name ip addr show dev hi-example" "inet 192.73.0.1/24"
  assert_unmatch 3 "docker exec $sw1_name ip addr show dev hi-example" "inet 192.72.0.1/24"
  assert_match 10 "docker exec $sw1_name iptables -t nat -S TT_example_SNAT" "192.73.0.0/24"
  assert_unmatch 3 "docker exec $sw1_name iptables -t nat -S TT_example_SNAT" "192.72.0.0/24"
  assert_cmd docker exec $sw1_name openlan network --name example openvpn restart
  assert_match 10 "docker exec $sw1_name cat /var/openlan/openvpn/example/tcp1194server.conf" 'push "route 192.73.0.0 255.255.255.0"'
  assert_unmatch 3 "docker exec $sw1_name cat /var/openlan/openvpn/example/tcp1194server.conf" 'push "route 192.72.0.0 255.255.255.0"'
}

setup() {
  setup_net
  setup_sw1
  test_setaddress_openvpn_route
}

case "$1" in
  --description)
    show_description
    ;;
  --summary)
    show_topology_summary
    ;;
  --topology)
    show_topology
    ;;
  *)
    main
    ;;
esac
