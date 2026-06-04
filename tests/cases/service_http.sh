#!/bin/bash
source tools/auto.sh

show_description() {
  echo "verify ceci service http forwarding and restart"
}

show_topology_summary() {
  cat <<'EOF'
client wget with Host header on sw1 | sw1 Ceci HTTP service | ^ route groups ^ global backend
EOF
}

show_topology() {
  cat <<'EOF'
# Topology:
# - Diagram:
#       client wget with Host header on sw1
#              |
#       sw1 Ceci HTTP service
#          ^ route groups          ^ global backend
#          |                       |
#       sw2 backends            sw3 backend
#       192.56.0.2              192.56.0.3
# - Docker mgmt network: 100.100.0.0/24
#   sw1=100.100.0.241 (ceci service),
#   sw2=100.100.0.242 (hostname-route backends),
#   sw3=100.100.0.243 (global backend).
# - OpenLAN service network "example": 192.56.0.0/24
#   sw1=192.56.0.1, sw2=192.56.0.2, sw3=192.56.0.3,
#   with sw2/sw3 outputs to sw1.
# Validation:
#   sw1 wget(host header) -> sw1 ceci(service http) -> sw2 route groups or sw3 global backend.

EOF
}

# OpenLAN Proxy UT: Ceci service http.

export net_name=tests-net-service-http
export sw1_name=tests-sw-service-http1
export sw2_name=tests-sw-service-http2
export sw3_name=tests-sw-service-http3
export http_service_listen=127.0.0.1:13083
export http_target_listen_a=18084
export http_target_listen_b=18085
export http_target_listen_single=18086
export http_target_listen_c=18087
export http_target_listen_global=18088
export http_target_body_a=ceci-service-http-a
export http_target_body_b=ceci-service-http-b
export http_target_body_single=ceci-service-http-single
export http_target_body_c=ceci-service-http-c
export http_target_body_global=ceci-service-http-global
export http_global_backend=192.56.0.3:18088

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

  assert_cmd docker exec $name openlan network --name example add --address 192.56.0.1/24
  assert_cmd docker exec $name openlan user add --name t1@example --password 123456 --role admin
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

  assert_cmd docker exec $name openlan network --name example add --address 192.56.0.2/24
  assert_cmd docker exec $name openlan network --name example output add --remote 100.100.0.241 --protocol tcp --secret t1@example:123456 --crypt aes-128:ea64d5b0c96c
  assert_match 20 "docker exec $name openlan network --name example output ls" "state: authenticated"
}

setup_sw3() {
  local name="$sw3_name"
  local address=100.100.0.243

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

  assert_cmd docker exec $name openlan network --name example add --address 192.56.0.3/24
  assert_cmd docker exec $name openlan network --name example output add --remote 100.100.0.241 --protocol tcp --secret t1@example:123456 --crypt aes-128:ea64d5b0c96c
  assert_match 20 "docker exec $name openlan network --name example output ls" "state: authenticated"
}

setup_target_http() {
  assert_cmd docker exec $sw2_name sh -c "mkdir -p /tmp/ceci-service-http-a /tmp/ceci-service-http-b /tmp/ceci-service-http-single"
  assert_cmd docker exec $sw2_name sh -c "mkdir -p /tmp/ceci-service-http-c"
  assert_cmd docker exec $sw3_name sh -c "mkdir -p /tmp/ceci-service-http-global"
  assert_cmd docker exec $sw2_name sh -c "echo '$http_target_body_a' > /tmp/ceci-service-http-a/index.html"
  assert_cmd docker exec $sw2_name sh -c "echo '$http_target_body_b' > /tmp/ceci-service-http-b/index.html"
  assert_cmd docker exec $sw2_name sh -c "echo '$http_target_body_single' > /tmp/ceci-service-http-single/index.html"
  assert_cmd docker exec $sw2_name sh -c "echo '$http_target_body_c' > /tmp/ceci-service-http-c/index.html"
  assert_cmd docker exec $sw3_name sh -c "echo '$http_target_body_global' > /tmp/ceci-service-http-global/index.html"
  assert_cmd docker exec $sw2_name sh -c "nohup python3 -m http.server $http_target_listen_a --bind 0.0.0.0 --directory /tmp/ceci-service-http-a >/tmp/ceci-service-http-a.log 2>&1 &"
  assert_cmd docker exec $sw2_name sh -c "nohup python3 -m http.server $http_target_listen_b --bind 0.0.0.0 --directory /tmp/ceci-service-http-b >/tmp/ceci-service-http-b.log 2>&1 &"
  assert_cmd docker exec $sw2_name sh -c "nohup python3 -m http.server $http_target_listen_single --bind 0.0.0.0 --directory /tmp/ceci-service-http-single >/tmp/ceci-service-http-single.log 2>&1 &"
  assert_cmd docker exec $sw2_name sh -c "nohup python3 -m http.server $http_target_listen_c --bind 0.0.0.0 --directory /tmp/ceci-service-http-c >/tmp/ceci-service-http-c.log 2>&1 &"
  assert_cmd docker exec $sw3_name sh -c "nohup python3 -m http.server $http_target_listen_global --bind 0.0.0.0 --directory /tmp/ceci-service-http-global >/tmp/ceci-service-http-global.log 2>&1 &"
}

setup_service() {
  local name="$sw1_name"
  assert_cmd docker exec $name openlan ceci service add --listen $http_service_listen --protocol http --balance roundrobin
  # split-form route add: --hostname <host> --backend <server1|server2>
  assert_cmd docker exec $name openlan ceci service backend add --listen $http_service_listen --hostname group.test --backend 192.56.0.2:$http_target_listen_a\|192.56.0.2:$http_target_listen_b
  assert_cmd docker exec $name openlan ceci service backend add --listen $http_service_listen --hostname single.test --backend 192.56.0.2:$http_target_listen_single
  assert_match 10 "docker exec $name openlan ceci service ls" "listen: $http_service_listen"
  assert_match 10 "docker exec $name openlan ceci service ls" "protocol: http"
  assert_match 10 "docker exec $name openlan ceci service ls" "balance: roundrobin"
  assert_match 10 "docker exec $name openlan ceci service ls" "group.test"
  assert_match 10 "docker exec $name openlan ceci service ls" "single.test"
  assert_cmd docker exec $name openlan ceci service backend add --listen $http_service_listen --hostname group.test --backend 192.56.0.2:$http_target_listen_c
  assert_cmd docker exec $name openlan ceci service backend add --listen $http_service_listen --backend $http_global_backend
  assert_match 10 "docker exec $name openlan ceci service ls" "$http_target_listen_c"
  assert_match 10 "docker exec $name openlan ceci service ls" "$http_global_backend"
  assert_cmd docker exec $name sh -c "echo '127.0.0.1 single.test group.test unknown.test' >> /etc/hosts"
}

test_ping() {
  assert_match 20 "docker exec $sw1_name ping -c 3 192.56.0.2" "bytes from"
  assert_match 20 "docker exec $sw1_name ping -c 3 192.56.0.3" "bytes from"
  assert_match 10 "docker exec $sw1_name wget -q -O- http://192.56.0.2:$http_target_listen_a/" "$http_target_body_a"
  assert_match 10 "docker exec $sw1_name wget -q -O- http://192.56.0.2:$http_target_listen_b/" "$http_target_body_b"
  assert_match 10 "docker exec $sw1_name wget -q -O- http://192.56.0.2:$http_target_listen_single/" "$http_target_body_single"
  assert_match 10 "docker exec $sw1_name wget -q -O- http://192.56.0.2:$http_target_listen_c/" "$http_target_body_c"
  assert_match 10 "docker exec $sw1_name wget -q -O- http://192.56.0.3:$http_target_listen_global/" "$http_target_body_global"
}

test_ceci_service() {
  assert_match 20 "docker exec $sw1_name wget -q -O- http://single.test:13083/" "$http_target_body_single"
  assert_match 20 "docker exec $sw1_name wget -q -O- http://group.test:13083/" "$http_target_body_a"
  assert_match 20 "docker exec $sw1_name wget -q -O- http://group.test:13083/" "$http_target_body_b"
  assert_match 20 "docker exec $sw1_name wget -q -O- http://group.test:13083/" "$http_target_body_c"
  assert_match 20 "docker exec $sw1_name wget -q -O- http://unknown.test:13083/" "$http_target_body_global"
}

restart_ceci_service() {
  assert_cmd docker exec $sw1_name openlan ceci service restart --listen $http_service_listen
  assert_match 20 "docker exec $sw1_name openlan ceci service ls" "listen: $http_service_listen"
}

remove_ceci_service() {
  assert_cmd docker exec $sw1_name openlan ceci service rm --listen $http_service_listen
  assert_unmatch 10 "docker exec $sw1_name openlan ceci service ls" "$http_service_listen"
  assert_unmatch 15 "docker exec $sw1_name wget -q -O- http://group.test:13083/" "$http_target_body_a"
}

setup_topology() {
  setup_net
  setup_sw1
  setup_sw2
  setup_sw3
  setup_target_http
}

setup() {
  setup_topology
  test_ping
  setup_service
  test_ceci_service
  restart_ceci_service
  test_ceci_service
  remove_ceci_service
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
