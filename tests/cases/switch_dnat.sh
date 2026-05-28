#!/bin/bash
source tools/auto.sh

show_topology() {
  cat <<'EOF'
# Topology:
# - Docker mgmt network: 172.246.0.0/24
#   sw1=172.246.0.241, sw2=172.246.0.242.
# - OpenLAN service network "example": 192.58.0.0/24
#   sw1=192.58.0.1, sw2=192.58.0.2.
# Validation:
#   start local 127.0.0.1:8080 service on sw2, map example:80 to 8080 by dnat,
#   verify unreachable before dnat and reachable after dnat from sw1.

EOF
}

# OpenLAN Switch UT: DNAT add/remove path.

export net_name=tests-net-dnat
export sw1_name=tests-sw-dnat1
export sw2_name=tests-sw-dnat2


setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.246.0.0/24 --gateway=172.246.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.246.0.241

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan crypt update --algorithm aes-128 --secret cb2ff088a34d
  assert_cmd docker exec $name openlan network --name example add --address 192.58.0.1/24
  assert_cmd docker exec $name openlan user add --name t1@example --password 123456
}

setup_sw2() {
  local name="$sw2_name"
  local address=172.246.0.242

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan crypt update --algorithm aes-128 --secret cb2ff088a34d
  assert_cmd docker exec $name openlan network --name example add --address 192.58.0.2/24
  assert_cmd docker exec $name openlan user add --name t1@example --password 123456
  assert_cmd docker exec $name openlan network --name example output add --remote 172.246.0.241 --protocol udp --secret t1:123456 --crypt aes-128:cb2ff088a34d
}

setup_http() {
  assert_cmd docker exec $sw2_name sh -c "nohup sh -c 'while true; do printf \"HTTP/1.1 200 OK\\r\\nContent-Length: 9\\r\\n\\r\\nport-8080\" | socat - TCP-LISTEN:8080,bind=127.0.0.1,reuseaddr; done' >/tmp/dnat-8080.log 2>&1 &"
}

test_dnat_add_and_reachability() {
  assert_cmd docker exec $sw2_name sysctl -w net.ipv4.conf.all.route_localnet=1
  assert_match 15 "docker exec $sw1_name openlan network --name example access ls" "172.246.0.242"
  assert_match 5 "docker exec $sw1_name ping -c 3 192.58.0.2" "bytes from"

  # verify dnat is required for access and the rule is added after dnat add
  assert_unmatch 5 "docker exec $sw1_name wget -qO- -T 3 -t 1 http://192.58.0.2:80" "port-8080"

  # add dnat rule and verify the rule and reachability after dnat add
  assert_cmd docker exec $sw2_name openlan network --name example dnat add --protocol tcp --dest 192.58.0.2 --dport 80 --todest 127.0.0.1 --todport 8080
  assert_match 15 "docker exec $sw2_name openlan network --name example dnat ls" "todport: 8080"
  assert_match 15 "docker exec $sw2_name iptables -t nat -S TT_example_DNAT " "DNAT tcp:192.58.0.2:80"

  assert_match 15 "docker exec $sw1_name wget -qO- -T 3 -t 1 http://192.58.0.2:80" "port-8080"

  assert_cmd docker exec $sw1_name openlan reload --save
  assert_cmd docker exec $sw2_name openlan reload --save
  assert_match 15 "docker exec $sw2_name openlan network --name example dnat ls" "todport: 8080"
  assert_match 15 "docker exec $sw2_name iptables -t nat -S TT_example_DNAT " "DNAT tcp:192.58.0.2:80"
  assert_match 15 "docker exec $sw1_name wget -qO- -T 3 -t 1 http://192.58.0.2:80" "port-8080"
}

test_dnat_remove() {
  assert_cmd docker exec $sw2_name openlan network --name example dnat rm --protocol tcp --dest 192.58.0.2 --dport 80

  assert_unmatch 3 "docker exec $sw2_name openlan network --name example dnat ls" "dport: 80"
  assert_unmatch 5 "docker exec $sw1_name wget -qO- -T 3 -t 1 http://192.58.0.2:80" "port-8080"

  assert_cmd docker exec $sw1_name openlan reload --save
  assert_cmd docker exec $sw2_name openlan reload --save
  assert_unmatch 3 "docker exec $sw2_name openlan network --name example dnat ls" "dport: 80"
  assert_unmatch 5 "docker exec $sw1_name wget -qO- -T 3 -t 1 http://192.58.0.2:80" "port-8080"
}

setup_topology() {
  setup_net
  setup_sw1
  setup_sw2
  setup_http
}

setup() {
  setup_topology
  test_dnat_add_and_reachability
  test_dnat_remove
}

case "$1" in
  --topology)
    show_topology
    ;;
  *)
    main
    ;;
esac
