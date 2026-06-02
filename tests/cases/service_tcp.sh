#!/bin/bash
source tools/auto.sh

show_topology() {
  cat <<'EOF'
# Topology:
# - Diagram:
#       client wget on sw1
#              |
#       sw1 Ceci TCP service -- output --> sw2 backend 192.56.0.2:18083
# - Docker mgmt network: 172.246.0.0/24
#   sw1=172.246.0.241 (ceci service), sw2=172.246.0.242 (service backends).
# - OpenLAN service network "example": 192.56.0.0/24
#   sw1=192.56.0.1, sw2=192.56.0.2, with sw2 output to sw1.
# Validation:
#   sw1 wget -> sw1 ceci(service tcp) -> sw2 local http server.

EOF
}

# OpenLAN Proxy UT: Ceci service tcp.

export net_name=tests-net-service-tcp
export sw1_name=tests-sw-service-tcp1
export sw2_name=tests-sw-service-tcp2
export tcp_service_listen=127.0.0.1:13082
export tcp_target_listen=18083
export tcp_target_body=ceci-service-tcp-ok

setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.246.0.0/24 --gateway=172.246.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.246.0.241

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
  assert_cmd docker exec $name openlan user add --name t1@example --password 123456
  assert_cmd docker exec $name openlan ceci service add --listen $tcp_service_listen --protocol tcp --balance roundrobin
  assert_cmd docker exec $name openlan ceci service backend add --listen $tcp_service_listen --backend 192.56.0.2:$tcp_target_listen
  assert_match 10 "docker exec $name openlan ceci service ls" "listen: $tcp_service_listen"
  assert_match 10 "docker exec $name openlan ceci service ls" "protocol: tcp"
}

setup_sw2() {
  local name="$sw2_name"
  local address=172.246.0.242

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
  assert_cmd docker exec $name openlan network --name example output add --remote 172.246.0.241 --protocol tcp --secret t1@example:123456 --crypt aes-128:ea64d5b0c96c
  assert_match 20 "docker exec $name openlan network --name example output ls" "state: authenticated"
}

setup_target_http() {
  assert_cmd docker exec $sw2_name sh -c "mkdir -p /tmp/ceci-service-tcp && echo '$tcp_target_body' > /tmp/ceci-service-tcp/index.html"
  assert_cmd docker exec $sw2_name sh -c "nohup python3 -m http.server $tcp_target_listen --bind 0.0.0.0 --directory /tmp/ceci-service-tcp >/tmp/ceci-service-tcp.log 2>&1 &"
  assert_match 10 "docker exec $sw2_name wget -q -O- http://192.56.0.2:$tcp_target_listen/" "$tcp_target_body"
}

test_ceci_service() {
  assert_match 20 "docker exec $sw1_name ping -c 3 192.56.0.2" "bytes from"
  assert_match 20 "docker exec $sw1_name wget -q -O- http://$tcp_service_listen/" "$tcp_target_body"
}

restart_ceci_service() {
  assert_cmd docker exec $sw1_name openlan ceci service restart --listen $tcp_service_listen
  assert_match 20 "docker exec $sw1_name openlan ceci service ls" "listen: $tcp_service_listen"
}

remove_ceci_service() {
  assert_cmd docker exec $sw1_name openlan ceci service rm --listen $tcp_service_listen
  assert_unmatch 10 "docker exec $sw1_name openlan ceci service ls" "$tcp_service_listen"
  assert_unmatch 15 "docker exec $sw1_name wget -q -O- http://$tcp_service_listen/" "$tcp_target_body"
}

setup_topology() {
  setup_net
  setup_sw1
  setup_sw2
  setup_target_http
}

setup() {
  setup_topology
  test_ceci_service
  restart_ceci_service
  test_ceci_service
  remove_ceci_service
}

case "$1" in
  --topology)
    show_topology
    ;;
  *)
    main
    ;;
esac
