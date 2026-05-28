#!/bin/bash
source tools/auto.sh

show_topology() {
  cat <<'EOF'
# Topology:
# - Docker mgmt network: 172.254.0.0/24
#   sw1=172.254.0.241, sw2=172.254.0.242.
# - OpenLAN service network "example": 192.61.0.0/24
#   sw1=192.61.0.1, sw2=192.61.0.2.
# - sw2 VIP:
#   lo=10.254.0.12/32, tcp/80 service.
# Validation:
#   persistence across reload, and sw1 -> sw2 VIP tcp/80 and ICMP behavior.

EOF
}

# OpenLAN Switch UT: ACL add/list/save/reload/remove path.

export net_name=tests-net-acl
export sw1_name=tests-sw-acl1
export sw2_name=tests-sw-acl2
export vip_address=10.254.0.12


setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.254.0.0/24 --gateway=172.254.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.254.0.241
  local crypt_secret="cb2ff088a34d"

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan crypt update --algorithm aes-128 --secret "$crypt_secret"
  assert_cmd docker exec $name openlan network --name example add --address 192.61.0.1/24
  assert_cmd docker exec $name openlan network --name example route add --prefix $vip_address/32 --nexthop 192.61.0.2
  assert_cmd docker exec $name openlan user add --name t1@example --password 123456
  assert_cmd docker exec $name ip link show hi-example
}

setup_sw2() {
  local name="$sw2_name"
  local address=172.254.0.242
  local crypt_secret="cb2ff088a34d"

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan crypt update --algorithm aes-128 --secret "$crypt_secret"
  assert_cmd docker exec $name openlan network --name example add --address 192.61.0.2/24
  assert_cmd docker exec $name openlan router address add --device lo --address $vip_address/32
  assert_cmd docker exec $name openlan user add --name t1@example --password 123456
  assert_cmd docker exec $name openlan network --name example output add --remote 172.254.0.241 --protocol udp --secret t1@example:123456 --crypt aes-128:$crypt_secret
}

setup_vip_http() {
  assert_cmd docker exec $sw2_name sh -c "nohup sh -c 'while true; do printf \"HTTP/1.1 200 OK\\r\\nContent-Length: 10\\r\\n\\r\\nacl-vip-80\" | socat - TCP-LISTEN:80,bind=$vip_address,reuseaddr; done' >/tmp/acl-vip-80.log 2>&1 &"
}

test_vip_reachable_before_acl() {
  assert_match 15 "docker exec $sw1_name openlan network --name example access ls" "172.254.0.242"
  assert_match 5 "docker exec $sw1_name ping -c 3 $vip_address" "bytes from"
  assert_match 5 "docker exec $sw1_name wget -qO- -T 3 -t 1 http://$vip_address:80" "acl-vip-80"
}

test_acl_add() {
  assert_match 5 "docker exec $sw1_name wget -qO- -T 3 -t 1 http://$vip_address:80" "acl-vip-80"
  assert_cmd docker exec $sw2_name openlan acl --name example rule add --source 192.61.0.1 --destination $vip_address --protocol tcp --dport 80

  assert_match 10 "docker exec $sw2_name openlan acl --name example rule list" "192.61.0.1"
  assert_match 10 "docker exec $sw2_name openlan acl --name example rule list" "$vip_address"
  assert_match 10 "docker exec $sw2_name iptables -t raw -S TT_pre-example" "AT_example"
  assert_match 10 "docker exec $sw2_name iptables -t raw -S AT_example" "192.61.0.1.*$vip_address.*tcp.*--dport 80.*DROP"
  assert_unmatch 5 "docker exec $sw1_name wget -qO- -T 3 -t 1 http://$vip_address:80" "acl-vip-80"

  assert_match 5 "docker exec $sw1_name ping -c 3 $vip_address" "bytes from"
  assert_cmd docker exec $sw2_name openlan acl --name example rule add --source 192.61.0.1 --destination $vip_address --protocol icmp

  assert_match 10 "docker exec $sw2_name openlan acl --name example rule list" "192.61.0.1"
  assert_match 10 "docker exec $sw2_name openlan acl --name example rule list" "$vip_address"
  assert_match 10 "docker exec $sw2_name iptables -t raw -S TT_pre-example" "AT_example"
  assert_match 10 "docker exec $sw2_name iptables -t raw -S AT_example" "192.61.0.1.*$vip_address.*icmp.*DROP"
  assert_unmatch 3 "docker exec $sw1_name ping -c 3 $vip_address" "bytes from"
}

test_acl_save_reload() {
  assert_cmd docker exec $sw2_name openlan acl --name example rule save
  assert_cmd docker exec $sw1_name openlan reload --save
  assert_cmd docker exec $sw2_name openlan reload --save

  assert_match 15 "docker exec $sw1_name openlan network --name example access ls" "172.254.0.242"
  assert_match 10 "docker exec $sw2_name openlan acl --name example rule list" "192.61.0.1"
  assert_match 10 "docker exec $sw2_name iptables -t raw -S TT_pre-example" "AT_example"
  assert_match 10 "docker exec $sw2_name iptables -t raw -S AT_example" "192.61.0.1.*$vip_address.*tcp.*--dport 80.*DROP"
  assert_match 10 "docker exec $sw2_name iptables -t raw -S AT_example" "192.61.0.1.*$vip_address.*icmp.*DROP"
  assert_unmatch 5 "docker exec $sw1_name wget -qO- -T 3 -t 1 http://$vip_address:80" "acl-vip-80"
  assert_unmatch 3 "docker exec $sw1_name ping -c 3 $vip_address" "bytes from"
}

test_acl_flush() {
  assert_cmd docker exec $sw2_name openlan acl --name example rule flush

  assert_unmatch 3 "docker exec $sw2_name openlan acl --name example rule list" "192.61.0.1"
  assert_unmatch 3 "docker exec $sw2_name iptables -t raw -S AT_example" "dport 80"
  assert_unmatch 3 "docker exec $sw2_name iptables -t raw -S AT_example" "icmp"
  assert_match 3 "docker exec $sw1_name ping -c 3 $vip_address" "bytes from"
  assert_match 5 "docker exec $sw1_name wget -qO- -T 3 -t 1 http://$vip_address:80" "acl-vip-80"

  assert_cmd docker exec $sw2_name openlan acl --name example rule save
  assert_cmd docker exec $sw1_name openlan reload --save
  assert_cmd docker exec $sw2_name openlan reload --save

  assert_unmatch 3 "docker exec $sw2_name openlan acl --name example rule list" "192.61.0.1"
  assert_unmatch 3 "docker exec $sw2_name iptables -t raw -S AT_example" "dport 80"
  assert_unmatch 3 "docker exec $sw2_name iptables -t raw -S AT_example" "icmp"
  assert_match 3 "docker exec $sw1_name ping -c 3 $vip_address" "bytes from"
  assert_match 5 "docker exec $sw1_name wget -qO- -T 3 -t 1 http://$vip_address:80" "acl-vip-80"
}

setup_topology() {
  setup_net
  setup_sw1
  setup_sw2
  setup_vip_http
}

setup() {
  setup_topology
  test_vip_reachable_before_acl
  test_acl_add
  test_acl_save_reload
  test_acl_flush
}

case "$1" in
  --topology)
    show_topology
    ;;
  *)
    main
    ;;
esac
