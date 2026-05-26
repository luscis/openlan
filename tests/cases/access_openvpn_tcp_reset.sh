#!/bin/bash
source tools/auto.sh


# OpenLAN OpenVPN TCP reset test:
# - before reject-with tcp-reset, vpn client can access sw1 service.
# - after reject-with tcp-reset, vpn client gets connection reset.

export net_name=tests-net-openvpn-rst
export sw1_name=tests-sw-openvpn-rst.sw1
export vpn1_name=tests-sw-openvpn-rst.vpn1

export sw1_overlay_ip=192.54.0.1
export rst_port=8082

setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.248.0.0/24 --gateway=172.248.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.248.0.241
  local crypt_secret="ea64d5b0c96c"

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan crypt update --algorithm aes-128 --secret "$crypt_secret"
  assert_cmd docker exec $name openlan network --name example add --address $sw1_overlay_ip/24
  assert_cmd docker exec $name openlan user add --name vpn1@example --password 123456
}

setup_openvpn_client() {
  local name="$sw1_name"

  assert_cmd docker exec $name openlan network --name example openvpn add \
    --listen :1194 --protocol tcp --subnet 10.92.0.0/24 --dns 8.8.8.8
  assert_cmd docker exec $name openlan network --name example client add --user vpn1 --address 10.92.0.10

  mkdir -p /opt/openlan/$vpn1_name/ovpn
  docker cp $name:/var/openlan/openvpn/example/tcp1194client.ovpn /opt/openlan/$vpn1_name/ovpn/client.ovpn
  cat > /opt/openlan/$vpn1_name/ovpn/auth.txt <<EOF
vpn1@example
123456
EOF

  start_openvpn $vpn1_name $net_name
  assert_expect 40 "docker logs -f $vpn1_name" "Initialization Sequence Completed"
}

setup_service() {
  assert_cmd docker exec $sw1_name sh -c \
    "nohup sh -c 'while true; do printf \"HTTP/1.1 200 OK\\r\\nContent-Length: 7\\r\\n\\r\\nrst-ok1\" | socat - TCP-LISTEN:$rst_port,bind=$sw1_overlay_ip,reuseaddr; done' >/tmp/rst-$rst_port.log 2>&1 &"
}

test_tcp_reset() {
  # Baseline: reachable before reset rule.
  assert_match 5 "docker exec $vpn1_name wget -qO- -T 3 -t 1 http://$sw1_overlay_ip:$rst_port" "rst-ok1"

  # Add explicit tcp reset reject on sw1.
  assert_cmd docker exec $sw1_name iptables -A INPUT -p tcp -d $sw1_overlay_ip --dport $rst_port -j REJECT --reject-with tcp-reset
  assert_match 3 "docker exec $sw1_name iptables -L INPUT -v -n" "tcp-reset"

  # Expect reset from client side.
  assert_fuzzy 5 "docker exec $vpn1_name wget -O- -T 3 -t 1 http://$sw1_overlay_ip:$rst_port" "Connection refused|Connection reset"
}

setup_topology() {
  setup_net
  setup_sw1
  setup_openvpn_client
  setup_service
  test_tcp_reset
}

setup() {
  setup_topology
}

main
