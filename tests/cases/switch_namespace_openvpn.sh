#!/bin/bash
source tools/auto.sh

show_topology() {
  cat <<'EOF'
# Topology:
# - Diagram:
#       vpn1 10.241.0.10
#             |
#       sw1 example [vrf-vpn]  -- TCP output -->  sw2 example VIP 10.240.2.12
#       sw1 network b 192.66.0.1 -- acb 192.66.0.11
#             VRF + OpenVPN SNAT path is isolated from network b
# - Docker mgmt network: 172.240.0.0/24
#   sw1=172.240.0.241, sw2=172.240.0.242, vpn1 joins the same mgmt network.
# - OpenLAN service network "example": 192.65.0.0/24
#   sw1=192.65.0.1, sw2=192.65.0.2.
# - sw1 service network L3 device and OpenVPN device are enslaved to VRF
#   "vrf-vpn"; sw2 is not.
# - sw1 non-namespace network "b": 192.66.0.0/24
#   sw1=192.66.0.1, acb=192.66.0.11.
# - OpenVPN overlay:
#   tcp/1194, subnet 10.241.0.0/24, vpn1@example fixed address 10.241.0.10.
# - sw2 VIP:
#   lo=10.240.2.12/32, declared in sw2 example and sw1 b networks, HTTP service
#   listens on 10.240.2.12:8081.
# - Forwarding link:
#   sw1 -> sw2 over TCP output.
# Validation:
#   vpn1 connects, server tun device is bound to the VRF. Without OpenVPN SNAT,
#   vpn1 cannot reach sw2 VIP; after enabling OpenVPN SNAT, sw2 HTTP sees sw1
#   overlay address as the source. acb, connected to sw1 non-namespace network
#   b, cannot reach the same VIP because b SNAT is disabled, even though b has
#   a route for the VIP.

EOF
}


# OpenLAN Switch UT: network namespace/VRF OpenVPN path.

export net_name=tests-net-namespace-openvpn
export sw1_name=tests-sw-namespace-openvpn1
export sw2_name=tests-sw-namespace-openvpn2
export vpn1_name=tests-sw-namespace-openvpn.vpn1
export acb_name=tests-sw-namespace-openvpn.acb
export vrf_name=vrf-vpn
export vpn_device=tun1194
export vpn_subnet=10.241.0.0/24
export vpn1_ip=10.241.0.10
export target_vip=10.240.2.12
export access_b_ip=192.66.0.11


setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.240.0.0/24 --gateway=172.240.0.1 >/dev/null
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
  local address=172.240.0.241

  setup_switch_config $name
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name example add --address 192.65.0.1/24 --namespace $vrf_name
  assert_match 1 "docker exec $name openlan network --name example" "namespace: $vrf_name"
  assert_cmd docker exec $name ip link show $vrf_name
  assert_match 5 "docker exec $name ip link show hi-example" "master $vrf_name"
  assert_cmd docker exec $name openlan network --name b add --address 192.66.0.1/24
  assert_cmd docker exec $name openlan network --name b snat disable
  assert_cmd docker exec $name openlan network --name b route add --prefix $target_vip/32 --nexthop 192.65.0.2
  assert_cmd docker exec $name openlan network --name example snat disable
  assert_cmd docker exec $name openlan network --name example route add --prefix $target_vip/32 --nexthop 192.65.0.2
  assert_cmd docker exec $name openlan user add --name vpn1@example --password 123456
  assert_cmd docker exec $name openlan user add --name acb@b --password 123456
  assert_cmd docker exec $name openlan network --name example output add --remote 172.240.0.242 --protocol tcp --secret t1@example:123456 --crypt aes-128:ea64d5b0c96c
}

setup_sw2() {
  local name="$sw2_name"
  local address=172.240.0.242

  setup_switch_config $name
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name example add --address 192.65.0.2/24
  assert_cmd docker exec $name openlan router address add --device lo --address $target_vip/32
  assert_cmd docker exec $name openlan network --name example route add --prefix $target_vip/32
  assert_cmd docker exec $name openlan user add --name t1@example --password 123456
}

setup_openvpn() {
  local name="$sw1_name"

  assert_cmd docker exec $name openlan network --name example openvpn add --listen :1194 --protocol tcp --subnet $vpn_subnet --dns 8.8.8.8
  assert_cmd docker exec $name openlan network --name example client add --user vpn1 --address $vpn1_ip
  assert_cmd docker exec $name test -f /var/openlan/openvpn/example/tcp1194server.conf
  assert_cmd docker exec $name test -f /var/openlan/openvpn/example/tcp1194client.ovpn
  assert_cmd docker exec $name test -f /var/openlan/openvpn/example/ccd/vpn1@example

  mkdir -p /opt/openlan/$vpn1_name/ovpn
  docker cp $name:/var/openlan/openvpn/example/tcp1194client.ovpn /opt/openlan/$vpn1_name/ovpn/client.ovpn
  cat > /opt/openlan/$vpn1_name/ovpn/auth.txt <<EOF
vpn1@example
123456
EOF

  start_openvpn $vpn1_name $net_name
  assert_expect 40 "docker logs -f $vpn1_name" "Initialization Sequence Completed"
}

setup_acb() {
  local name="$acb_name"

  mkdir -p /opt/openlan/$name/etc/openlan
  cat > /opt/openlan/$name/etc/openlan/access.yaml <<EOF
protocol: udp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
connection: 172.240.0.241
username: acb@b
password: 123456
interface:
  address: $access_b_ip/24
forward:
- $target_vip/32 to 192.66.0.1
EOF
  start_access $name $net_name
  assert_expect 30 "docker logs -f $name" "onLogin: success"
}

setup_vip_http() {
  assert_cmd docker exec $sw2_name sh -c "cat > /tmp/namespace-openvpn-http.sh <<'EOF'
#!/bin/sh
printf 'HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nConnection: close\r\n\r\nsrc=%s\n' \"\$SOCAT_PEERADDR\"
EOF
chmod +x /tmp/namespace-openvpn-http.sh
nohup socat TCP-LISTEN:8081,bind=$target_vip,reuseaddr,fork EXEC:/tmp/namespace-openvpn-http.sh >/tmp/namespace-openvpn-http.log 2>&1 &"
}

assert_acb_http_source() {
  local source=$1
  assert_match 20 "docker exec $acb_name wget -qO- -T 3 -t 1 http://$target_vip:8081" "src=$source"
}

assert_acb_http_unreachable() {
  assert_unmatch 3 "docker exec $acb_name wget -qO- -T 3 -t 1 http://$target_vip:8081" "src="
}

assert_vpn_http_source() {
  local source=$1
  assert_match 20 "docker exec $vpn1_name wget -qO- -T 3 -t 1 http://$target_vip:8081" "src=$source"
}

assert_vpn_http_unreachable() {
  assert_unmatch 3 "docker exec $vpn1_name wget -qO- -T 3 -t 1 http://$target_vip:8081" "src="
}

test_openvpn_namespace() {
  assert_match 15 "docker exec $sw1_name openlan network --name example output ls" "state: authenticated"
  assert_match 20 "docker exec $sw1_name ip link show $vpn_device" "master $vrf_name"
  assert_match 20 "docker exec $vpn1_name ping -c 3 192.65.0.1" "bytes from"
  assert_match 5 "docker exec $sw1_name ip route show vrf $vrf_name" "$vpn_subnet"
  assert_match 5 "docker exec $sw2_name openlan network --name example route ls" "$target_vip/32"
  assert_match 5 "docker exec $sw1_name openlan network --name b route ls" "$target_vip/32"

  assert_unmatch 3 "docker exec $vpn1_name ping -c 3 $target_vip" "bytes from"
  assert_vpn_http_unreachable
  assert_cmd docker exec $sw1_name openlan network --name example snat enable --scope openvpn
  assert_match 20 "docker exec $vpn1_name ping -c 3 $target_vip" "bytes from"
  assert_vpn_http_source 192.65.0.1
  assert_unmatch 3 "docker exec $acb_name ping -c 3 $target_vip" "bytes from"
  assert_acb_http_unreachable
}

test_reload_persistence() {
  assert_cmd docker exec $sw1_name openlan reload --save
  assert_cmd docker exec $sw2_name openlan reload --save

  assert_match 10 "docker exec $sw1_name openlan network --name example" "namespace: $vrf_name"
  assert_match 15 "docker exec $sw1_name openlan network --name example output ls" "state: authenticated"
  assert_match 20 "docker exec $sw1_name ip link show hi-example" "master $vrf_name"
  assert_match 20 "docker exec $sw1_name ip link show $vpn_device" "master $vrf_name"
  assert_match 20 "docker exec $vpn1_name ping -c 3 192.65.0.1" "bytes from"
  assert_match 5 "docker exec $sw2_name openlan network --name example route ls" "$target_vip/32"
  assert_match 5 "docker exec $sw1_name openlan network --name b route ls" "$target_vip/32"
  assert_match 20 "docker exec $vpn1_name ping -c 3 $target_vip" "bytes from"
  assert_vpn_http_source 192.65.0.1
  assert_unmatch 3 "docker exec $acb_name ping -c 3 $target_vip" "bytes from"
  assert_acb_http_unreachable
}

setup_topology() {
  setup_net
  setup_sw2
  setup_sw1
  setup_openvpn
  setup_acb
  setup_vip_http
}

setup() {
  setup_topology
  test_openvpn_namespace
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
