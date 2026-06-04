#!/bin/bash
source tools/auto.sh

show_description() {
  echo "verify ceci http proxy forwarding to http target"
}

show_topology_summary() {
  cat <<'EOF'
sw1 proxy client 192.52.0.1 | wget via local Ceci HTTP proxy | sw1 openceci(http) -- output --> sw2 192.52.0.2:18081
EOF
}

show_topology() {
  cat <<'EOF'
# Topology:
# - Diagram:
#       sw1 proxy client 192.52.0.1
#              | wget via local Ceci HTTP proxy
#              v
#       sw1 openceci(http) -- output --> sw2 192.52.0.2:18081
# - Docker mgmt network: 100.100.0.0/24
#   sw1=100.100.0.241 (ceci http proxy), sw2=100.100.0.242 (http target/client).
# - OpenLAN service network "example": 192.52.0.0/24
#   sw1=192.52.0.1, sw2=192.52.0.2, with sw2 output to sw1.
# Validation:
#   sw1 wget -> sw1 ceci(http proxy) -> sw2(192.52.0.2) local http server.

EOF
}

# OpenLAN Proxy UT: Ceci HTTP forward proxy path.

export net_name=tests-net-proxy-http
export sw1_name=tests-sw-proxy-http1
export sw2_name=tests-sw-proxy-http2
export proxy_listen=127.0.0.1:11082
export target_listen=18081
export target_body=proxy-http-ok


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

  assert_cmd docker exec $name openlan network --name example add --address 192.52.0.1/24
  assert_cmd docker exec $name openlan user add --name t1@example --password 123456
  assert_cmd docker exec $name openlan ceci proxy add --mode http --listen $proxy_listen
  assert_match 5 "docker exec $name openlan ceci ls" "mode: http"
  assert_match 5 "docker exec $name openlan ceci ls" "listen: $proxy_listen"
}

setup_sw2() {
  local name="$sw2_name"
  local address=100.100.0.242

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

  assert_cmd docker exec $name openlan network --name example add --address 192.52.0.2/24
  assert_cmd docker exec $name openlan network --name example output add --remote 100.100.0.241 --protocol tcp --secret t1@example:123456 --crypt aes-128:ea64d5b0c96c
  assert_match 20 "docker exec $name openlan network --name example output ls" "state: authenticated"
}

setup_target_http() {
  assert_cmd docker exec $sw2_name sh -c "mkdir -p /tmp/proxy-http && echo '$target_body' > /tmp/proxy-http/index.html"
  assert_cmd docker exec $sw2_name sh -c "nohup python3 -m http.server $target_listen --bind 0.0.0.0 --directory /tmp/proxy-http >/tmp/proxy-http.log 2>&1 &"
  assert_match 10 "docker exec $sw2_name wget -q -O- http://192.52.0.2:$target_listen/" "$target_body"
}

test_http_proxy() {
  assert_match 20 "docker exec $sw1_name ping -c 3 192.52.0.2" "bytes from"
  assert_match 20 "docker exec $sw1_name wget -q -O- -e use_proxy=yes -e http_proxy=http://127.0.0.1:11082 http://192.52.0.2:$target_listen/" "$target_body"
  assert_fuzzy 20 "docker exec $sw1_name cat /var/openlan/ceci/$proxy_listen.log" "HttpProxy.ServeHTTP .* 192.52.0.2:$target_listen"
}

restart_http_proxy() {
  assert_cmd docker exec $sw1_name pkill -f /usr/bin/openceci
  assert_cmd docker exec $sw1_name openlan reload --save
  assert_match 20 "docker exec $sw1_name openlan ceci ls" "listen: $proxy_listen"
  assert_match 30 "docker exec $sw1_name ping -c 3 192.52.0.2" "bytes from"
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
  test_http_proxy
  restart_http_proxy
  test_http_proxy
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
