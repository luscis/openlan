#!/bin/bash
source tools/auto.sh

show_description() {
  echo "verify sw1 network a openvpn client reaches sw2 network a and b addresses through snat"
}

show_topology_summary() {
  cat <<'EOF'
vpn1@sw1/a 10.84.0.10 | sw1 a 192.84.0.1 -- TCP output + SNAT --> sw2 a 192.84.0.2 + sw2 b 192.85.0.2
EOF
}

show_topology() {
  cat <<'EOF'
# Topology:
# - Diagram:
#       vpn1@a 10.84.0.10
#              |
#       sw1 network a 192.84.0.1
#              |
#              | TCP output + OpenVPN SNAT
#              v
#       sw2 network a 192.84.0.2
#       sw2 network b 192.85.0.2
# - Docker mgmt network: 100.100.0.0/24
#   sw1=100.100.0.241, sw2=100.100.0.242, vpn1 joins the same mgmt network.
# - OpenLAN service networks:
#   network a: sw1=192.84.0.1/24, sw2=192.84.0.2/24.
#   network b: sw2=192.85.0.2/24.
# - OpenVPN overlay:
#   sw1 network a tcp/1194, subnet 10.84.0.0/24, vpn1@a fixed address 10.84.0.10.
# - Routing:
#   sw1/a routes sw2/a and sw2/b subnets via sw2/a; sw2 has no OpenVPN return route.
# Validation:
#   Without SNAT, the OpenVPN client cannot reach sw2 network a or b because
#   sw2 has no return route for the OpenVPN subnet. After enabling OpenVPN SNAT
#   on sw1/a, the client can reach both sw2 network a and b addresses.

EOF
}

export net_name=tests-net-openvpn-multi-snat
export sw1_name=tests-sw-openvpn-multi-snat.sw1
export sw2_name=tests-sw-openvpn-multi-snat.sw2
export vpn1_name=tests-sw-openvpn-multi-snat.vpn1
export pass_a="pw-a-${RANDOM}-${RANDOM}"
export vpn_subnet=10.84.0.0/24
export vpn1_ip=10.84.0.10
export sw1_a_ip=192.84.0.1
export sw2_a_ip=192.84.0.2
export sw2_b_ip=192.85.0.2

setup_net() {
  docker network create $net_name --driver=bridge --subnet=100.100.0.0/24 --gateway=100.100.0.1 >/dev/null
}

setup_switch_config() {
  local name=$1

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
}

setup_sw1() {
  local name="$sw1_name"
  local address=100.100.0.241

  setup_switch_config $name
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name a add --address $sw1_a_ip/24
  assert_cmd docker exec $name openlan network --name a snat disable
  assert_cmd docker exec $name openlan network --name a route add --prefix 192.85.0.0/24 --nexthop $sw2_a_ip
  assert_cmd docker exec $name openlan user add --name vpn1@a --password "$pass_a"
  assert_cmd docker exec $name openlan network --name a output add --remote 100.100.0.242 --protocol tcp --secret link@a:$pass_a --crypt aes-128:ea64d5b0c96c
}

setup_sw2() {
  local name="$sw2_name"
  local address=100.100.0.242

  setup_switch_config $name
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name a add --address $sw2_a_ip/24
  assert_cmd docker exec $name openlan network --name b add --address $sw2_b_ip/24
  assert_cmd docker exec $name openlan network --name a snat disable
  assert_cmd docker exec $name openlan network --name b snat disable
  assert_cmd docker exec $name openlan user add --name link@a --password "$pass_a"
}

setup_openvpn() {
  local name="$sw1_name"

  assert_cmd docker exec $name openlan network --name a openvpn add --listen :1194 --protocol tcp --subnet $vpn_subnet --dns 8.8.8.8
  assert_cmd docker exec $name openlan network --name a client add --user vpn1 --address $vpn1_ip
  assert_cmd docker exec $name test -f /var/openlan/openvpn/a/tcp1194client.ovpn
  assert_match 10 "docker exec $name cat /var/openlan/openvpn/a/tcp1194server.conf" 'push "route 192.84.0.0 255.255.255.0"'
  assert_match 10 "docker exec $name cat /var/openlan/openvpn/a/tcp1194server.conf" 'push "route 192.85.0.0 255.255.255.0"'

  mkdir -p /opt/openlan/$vpn1_name/ovpn
  docker cp $name:/var/openlan/openvpn/a/tcp1194client.ovpn /opt/openlan/$vpn1_name/ovpn/client.ovpn
  cat > /opt/openlan/$vpn1_name/ovpn/auth.txt <<EOF
vpn1@a
$pass_a
EOF

  start_openvpn $vpn1_name $net_name
  assert_expect 40 "docker logs -f $vpn1_name" "Initialization Sequence Completed"
}

test_openvpn_multi_snat() {
  assert_match 20 "docker exec $sw1_name openlan network --name a output ls" "state: authenticated"
  assert_match 1 "docker exec $vpn1_name ping -c 3 $sw1_a_ip" "bytes from"
  assert_unmatch 1 "docker exec $vpn1_name ping -c 3 $sw2_a_ip" "bytes from"
  assert_unmatch 1 "docker exec $vpn1_name ping -c 3 $sw2_b_ip" "bytes from"
  
  assert_cmd docker exec $sw1_name openlan network --name a snat enable --scope openvpn
  assert_match 10 "docker exec $sw1_name iptables -t nat -S TT_a_SNAT" "$vpn_subnet"
  assert_match 1 "docker exec $vpn1_name ping -c 3 $sw1_a_ip" "bytes from"
  assert_unmatch 1 "docker exec $vpn1_name ping -c 3 $sw2_a_ip" "bytes from"
  assert_match 1 "docker exec $vpn1_name ping -c 3 $sw2_b_ip" "bytes from"

  assert_cmd docker exec $sw1_name openlan network --name a route add --prefix 192.84.0.0/24
  assert_match 1 "docker exec $vpn1_name ping -c 3 $sw1_a_ip" "bytes from"
  assert_match 1 "docker exec $vpn1_name ping -c 3 $sw2_a_ip" "bytes from"
  assert_match 1 "docker exec $vpn1_name ping -c 3 $sw2_b_ip" "bytes from"
}

setup() {
  setup_net
  setup_sw2
  setup_sw1
  setup_openvpn
  test_openvpn_multi_snat
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
