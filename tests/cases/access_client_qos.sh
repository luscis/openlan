#!/bin/bash
source tools/auto.sh

show_description() {
  echo "verify network client qos rule add-list-save-remove flow"
}

show_topology_summary() {
  cat <<'EOF'
sw1(center) 192.91.0.1 | OpenVPN tcp/1194, 10.91.0.0/24 | vpn1 10.91.0.10 | QoS rule add/update/remove on vpn1@example
EOF
}

show_topology() {
  cat <<'EOF'
# Topology:
# - Diagram:
#            sw1(center) 192.91.0.1
#                 ^
#                 | OpenVPN tcp/1194, 10.91.0.0/24
#              vpn1 10.91.0.10
#                 |
#              QoS rule add/update/remove on vpn1@example
# - Docker mgmt network: 100.100.0.0/24
#   sw1=100.100.0.241, vpn1 client joins the same mgmt network.
# - OpenLAN service network "example": 192.91.0.0/24
#   sw1 gateway=192.91.0.1.
# - OpenVPN overlay:
#   tcp/1194, subnet 10.91.0.0/24, vpn1 static address 10.91.0.10.
# Validation:
#   add/update/list/save/remove network qos rule for vpn client.
EOF
}

# OpenLAN Access UT: client qos rules.

export net_name=tests-net-client-qos
export sw1_name=tests-sw-client-qos
export vpn1_name=tests-sw-client-qos.vpn1
export vpn_user=vpn1@example

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

  assert_cmd docker exec $name openlan network --name example add --address 192.91.0.1/24
  assert_cmd docker exec $name openlan user add --name vpn1@example --password 123456
  assert_cmd docker exec $name openlan network --name example openvpn add --listen :1194 --protocol tcp --subnet 10.91.0.0/24 --dns 8.8.8.8
  assert_cmd docker exec $name openlan network --name example client add --user vpn1 --address 10.91.0.10
}

setup_openvpn_client() {
  mkdir -p /opt/openlan/$vpn1_name/ovpn
  docker cp $sw1_name:/var/openlan/openvpn/example/tcp1194client.ovpn /opt/openlan/$vpn1_name/ovpn/client.ovpn
  cat > /opt/openlan/$vpn1_name/ovpn/auth.txt <<EOF
vpn1@example
123456
EOF

  start_openvpn $vpn1_name $net_name
  assert_expect 40 "docker logs -f $vpn1_name" "Initialization Sequence Completed"
  assert_match 10 "docker exec $vpn1_name ping -c 3 192.91.0.1" "bytes from"
}

test_client_qos() {
  local name="$sw1_name"
  local qos_file=/etc/openlan/switch/qos/example.json
  local client_ip=10.91.0.10

  assert_cmd docker exec $name openlan network --name example qos rule add --client $vpn_user --inspeed 1.5
  assert_match 10 "docker exec $name openlan network --name example qos rule ls" "$vpn_user"
  assert_match 10 "docker exec $name openlan network --name example qos rule ls" "1.50"
  assert_match 10 "docker exec $name iptables -t mangle -S" "Qos_example-in"
  assert_match 10 "docker exec $name iptables -t mangle -S" "\-j Qos_example-in$"
  assert_match 10 "docker exec $name iptables -t mangle -S" "Qos Limit In $vpn_user"
  assert_match 10 "docker exec $name iptables -t mangle -S" "$client_ip"

  assert_cmd docker exec $name openlan network --name example qos rule add --client $vpn_user --inspeed 2
  assert_match 10 "docker exec $name openlan network --name example qos rule ls" "2.00"
  assert_unmatch 5 "docker exec $name openlan network --name example qos rule ls" "1.50"
  assert_match 10 "docker exec $name iptables -t mangle -S" "Qos Limit In $vpn_user"

  assert_cmd docker exec $name openlan network --name example qos rule save
  assert_match 10 "docker exec $name cat $qos_file" "$vpn_user"
  assert_match 10 "docker exec $name cat $qos_file" '"inSpeed": 2'

  assert_cmd docker exec $name openlan network --name example qos rule rm --client $vpn_user
  assert_unmatch 10 "docker exec $name openlan network --name example qos rule ls" "$vpn_user"
  assert_unmatch 10 "docker exec $name iptables -t mangle -S" "Qos Limit In $vpn_user"
  assert_unmatch 10 "docker exec $name iptables -t mangle -S" "$client_ip"
}

setup_topology() {
  setup_net
  setup_sw1
  setup_openvpn_client
  test_client_qos
}

setup() {
  setup_topology
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
