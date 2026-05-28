#!/bin/bash
source tools/auto.sh

show_topology() {
  cat <<'EOF'
# Topology:
# - Docker mgmt network: 172.249.0.0/24
#   sw1=172.249.0.241, sw2=172.249.0.242, vpn1 joins the same mgmt network.
# - OpenLAN service network "example": 192.53.0.0/24
#   sw1=192.53.0.1, sw2=192.53.0.2.
# - VIP services:
#   sw1 VIP=10.253.0.11:8081, sw2 VIP=10.253.0.12:8081.
# - Forwarding design:
#   sw2 has output to sw1;
#   sw2 adds return route for 10.97.0.0/24 via 192.53.0.1;
#   sw1 hosts OpenVPN tcp/1194 with subnet 10.97.0.0/24, vpn1 fixed IP 10.97.0.10.
# Validation:
#   before redirect, vpn1 reaches sw1 VIP and cannot reach sw2 VIP;
#   after redirecting source 10.97.0.10 to nexthop 192.53.0.2, vpn1 reaches sw2 VIP and cannot reach sw1 VIP.
EOF
}


# OpenLAN OpenVPN redirect test:
# before redirect, VPN reaches sw1 VIP; after redirect, only sw2 VIP is reachable.

export net_name=tests-net-openvpn-redirect
export sw1_name=tests-sw-openvpn-redirect.sw1
export sw2_name=tests-sw-openvpn-redirect.sw2
export vpn1_name=tests-sw-openvpn-redirect.vpn1

export sw1_vip=10.253.0.11
export sw2_vip=10.253.0.12

setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.249.0.0/24 --gateway=172.249.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.249.0.241
  local crypt_secret="ea64d5b0c96c"

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan crypt update --algorithm aes-128 --secret "$crypt_secret"
  assert_cmd docker exec $name openlan network --name example add --address 192.53.0.1/24
  assert_cmd docker exec $name openlan router address add --device lo --address $sw1_vip/32
  assert_cmd docker exec $name openlan user add --name uplink@example --password 123456
  assert_cmd docker exec $name openlan user add --name vpn1@example --password 123456
}

setup_sw2() {
  local name="$sw2_name"
  local address=172.249.0.242
  local crypt_secret="ea64d5b0c96c"

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan crypt update --algorithm aes-128 --secret "$crypt_secret"
  assert_cmd docker exec $name openlan network --name example add --address 192.53.0.2/24
  assert_cmd docker exec $name openlan router address add --device lo --address $sw2_vip/32
  assert_cmd docker exec $name openlan user add --name uplink@example --password 123456

  # Build forwarding path sw2 -> sw1.
  assert_cmd docker exec $name openlan network --name example output add --remote 172.249.0.241 --protocol tcp --secret uplink@example:123456 --crypt aes-128:$crypt_secret

  # Return path for VPN subnet when redirected through sw2.
  assert_cmd docker exec $name openlan network --name example route add --prefix 10.97.0.0/24 --nexthop 192.53.0.1
}

setup_vip_http() {
  assert_cmd docker exec $sw1_name sh -c "nohup sh -c 'while true; do printf \"HTTP/1.1 200 OK\\r\\nContent-Length: 7\\r\\n\\r\\nsw1-vip\" | socat - TCP-LISTEN:8081,bind=$sw1_vip,reuseaddr; done' >/tmp/sw1-vip.log 2>&1 &"
  assert_cmd docker exec $sw2_name sh -c "nohup sh -c 'while true; do printf \"HTTP/1.1 200 OK\\r\\nContent-Length: 7\\r\\n\\r\\nsw2-vip\" | socat - TCP-LISTEN:8081,bind=$sw2_vip,reuseaddr; done' >/tmp/sw2-vip.log 2>&1 &"
}

setup_openvpn_client() {
  local name="$sw1_name"

  assert_cmd docker exec $name openlan network --name example route add --prefix $sw2_vip/32
  assert_cmd docker exec $name openlan network --name example route add --prefix $sw1_vip/32
  assert_cmd docker exec $name openlan network --name example openvpn add --listen :1194 --protocol tcp --subnet 10.97.0.0/24 --dns 8.8.8.8
  assert_cmd docker exec $name openlan network --name example client add --user vpn1 --address 10.97.0.10

  mkdir -p /opt/openlan/$vpn1_name/ovpn
  docker cp $name:/var/openlan/openvpn/example/tcp1194client.ovpn /opt/openlan/$vpn1_name/ovpn/client.ovpn
  cat > /opt/openlan/$vpn1_name/ovpn/auth.txt <<EOF
vpn1@example
123456
EOF

  start_openvpn $vpn1_name $net_name
  assert_expect 40 "docker logs -f $vpn1_name" "Initialization Sequence Completed"
}

test_redirect() {
  # Before redirect: sw1 VIP is reachable.
  assert_match 10 "docker exec $vpn1_name wget -qO- -T 3 -t 1 http://$sw1_vip:8081" "sw1-vip"
  assert_unmatch 3 "docker exec $vpn1_name wget -qO- -T 3 -t 1 http://$sw2_vip:8081" "sw2-vip"
  # Redirect VPN source traffic to sw2 via table 100.
  assert_cmd docker exec $sw1_name openlan router redirect add --source 10.97.0.0/24 --nexthop 192.53.0.2 --table 100

  # After redirect: only sw2 VIP is reachable; sw1 VIP should fail.
  assert_match 10 "docker exec $vpn1_name wget -qO- -T 3 -t 1 http://$sw2_vip:8081" "sw2-vip"
  assert_match 10 "docker exec $vpn1_name wget -qO- -T 3 -t 1 http://$sw1_vip:8081" "sw1-vip"
}

setup_topology() {
  setup_net
  setup_sw1
  setup_sw2
  setup_vip_http
  setup_openvpn_client
  test_redirect
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
