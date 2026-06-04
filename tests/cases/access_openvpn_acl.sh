#!/bin/bash
source tools/auto.sh

show_description() {
  echo "verify OpenVPN ACL uses iptables while bridge ACL uses ebtables"
}

show_topology_summary() {
  cat <<'EOF'
sw1(center) 192.64.0.1 | OpenVPN tcp/1194, 10.88.0.0/24 | vpn1 10.88.0.10 | ACL drop rule is enforced by original iptables tun hook
EOF
}

show_topology() {
  cat <<'EOF'
# Topology:
# - Diagram:
#            sw1(center) 192.64.0.1
#                 ^
#                 | OpenVPN tcp/1194, 10.88.0.0/24
#              vpn1 10.88.0.10
#                 |
#              ACL drop rule is enforced by original iptables tun hook
# - Docker mgmt network: 100.100.0.0/24
#   sw1=100.100.0.241, vpn1 client joins the same mgmt network.
# - OpenLAN service network "example": 192.64.0.0/24
#   sw1 gateway=192.64.0.1.
# - OpenVPN overlay:
#   tcp/1194, subnet 10.88.0.0/24, vpn1 static address 10.88.0.10.
# Validation:
#   OpenVPN client is blocked by the original iptables ACL path on tun1194,
#   while bridge traffic uses ebtables and ebtables ACL never hooks tun1194.

EOF
}

# OpenLAN OpenVPN ACL scenario: iptables ACL still filters tun ingress.

export net_name=tests-net-openvpn-acl
export sw1_name=tests-sw-openvpn-acl.sw1
export vpn1_name=tests-sw-openvpn-acl.vpn1
export sw1_overlay_ip=192.64.0.1
export vpn1_ip=10.88.0.10

setup_net() {
  docker network create $net_name --driver=bridge --subnet=100.100.0.0/24 --gateway=100.100.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=100.100.0.241

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name example add --address $sw1_overlay_ip/24
  assert_cmd docker exec $name openlan user add --name vpn1@example --password 123456
  assert_cmd docker exec $name openlan network --name example openvpn add --listen :1194 --protocol tcp --subnet 10.88.0.0/24 --dns 8.8.8.8
  assert_cmd docker exec $name openlan network --name example client add --user vpn1 --address $vpn1_ip
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
  assert_match 10 "docker exec $vpn1_name ping -c 3 $sw1_overlay_ip" "bytes from"
}

test_openvpn_acl_scope() {
  assert_cmd docker exec $sw1_name openlan acl --name example rule add --srcip $vpn1_ip --dstip $sw1_overlay_ip --protocol icmp

  assert_match 10 "docker exec $sw1_name openlan acl --name example rule list" "$vpn1_ip"
  assert_match 10 "docker exec $sw1_name iptables -t raw -S TT_pre-example" "tun1194.*AT_example"
  assert_unmatch 3 "docker exec $sw1_name iptables -t raw -S TT_pre-example" "br-example.*AT_example"
  assert_match 10 "docker exec $sw1_name iptables -t raw -S TT_pre-example" "hi-example.*AT_example"
  assert_match 10 "docker exec $sw1_name iptables -t raw -S AT_example" "$vpn1_ip.*$sw1_overlay_ip.*icmp.*DROP"
  assert_match 10 "docker exec $sw1_name ebtables -t filter -L AT_example" "$vpn1_ip.*$sw1_overlay_ip.*icmp.*DROP"
  assert_match 10 "docker exec $sw1_name ebtables -t filter -L FORWARD" "logical-in br-example.*AT_example"
  assert_unmatch 3 "docker exec $sw1_name ebtables -t filter -L FORWARD" "logical-in tun1194.*AT_example"

  assert_unmatch 3 "docker exec $vpn1_name ping -c 3 $sw1_overlay_ip" "bytes from"
}

setup() {
  setup_net
  setup_sw1
  setup_openvpn_client
  test_openvpn_acl_scope
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
