#!/bin/bash
source tools/auto.sh

# OpenLAN Proxy UT: Ceci TCP proxy path.

export net_name=tests-net-proxy-tcp
export sw1_name=tests-sw-proxy-tcp1
export sw2_name=tests-sw-proxy-tcp2
export proxy_listen=127.0.0.1:12082
export target_listen=18082
export target_body=proxy-tcp-ok

# Topology:
# - Docker mgmt network: 172.250.0.0/24
#   sw1=172.250.0.241 (ceci tcp proxy), sw2=172.250.0.242 (tcp target).
# - OpenLAN service network "example": 192.53.0.0/24
#   sw1=192.53.0.1, sw2=192.53.0.2, with sw2 output to sw1.
# - Validation path:
#   sw1 wget -> sw1 ceci(tcp proxy) -> sw2(192.53.0.2) http server.

setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.250.0.0/24 --gateway=172.250.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.250.0.241

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

  assert_cmd docker exec $name openlan network --name example add --address 192.53.0.1/24
  assert_cmd docker exec $name openlan user add --name t1@example --password 123456
  assert_cmd docker exec $name openlan ceci proxy add --mode tcp --listen $proxy_listen --target 192.53.0.2:$target_listen
  assert_match 5 "docker exec $name openlan ceci ls" "mode: tcp"
  assert_match 5 "docker exec $name openlan ceci ls" "listen: $proxy_listen"
}

setup_sw2() {
  local name="$sw2_name"
  local address=172.250.0.242

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

  assert_cmd docker exec $name openlan network --name example add --address 192.53.0.2/24
  assert_cmd docker exec $name openlan network --name example output add --remote 172.250.0.241 --protocol tcp --secret t1@example:123456 --crypt aes-128:ea64d5b0c96c
  assert_match 20 "docker exec $name openlan network --name example output ls" "state: authenticated"
}

setup_target_http() {
  assert_cmd docker exec $sw2_name sh -c "mkdir -p /tmp/proxy-tcp && echo '$target_body' > /tmp/proxy-tcp/index.html"
  assert_cmd docker exec $sw2_name sh -c "nohup python3 -m http.server $target_listen --bind 0.0.0.0 --directory /tmp/proxy-tcp >/tmp/proxy-tcp.log 2>&1 &"
  assert_match 10 "docker exec $sw2_name wget -q -O- http://192.53.0.2:$target_listen/" "$target_body"
}

test_tcp_proxy() {
  assert_match 20 "docker exec $sw1_name ping -c 3 192.53.0.2" "bytes from"
  assert_match 20 "docker exec $sw1_name wget -q -O- http://127.0.0.1:12082/" "$target_body"
  assert_fuzzy 20 "docker exec $sw1_name cat /var/openlan/ceci/$proxy_listen.log" "TcpProxy.tunnel .* -> .*192.53.0.2:$target_listen"
}

restart_tcp_proxy() {
  assert_cmd docker exec $sw1_name pkill -f /usr/bin/openceci
  assert_cmd docker exec $sw1_name openlan reload --save
  assert_match 20 "docker exec $sw1_name openlan ceci ls" "listen: $proxy_listen"
  assert_match 30 "docker exec $sw1_name ping -c 3 192.53.0.2" "bytes from"
  assert_match 30 "docker exec $sw2_name openlan network --name example output ls" "state: authenticated"
}

setup_topology() {
  setup_net
  setup_sw1
  setup_sw2
  setup_target_http
}

setup() {
  setup_topology
  test_tcp_proxy
  restart_tcp_proxy
  test_tcp_proxy
}

main
