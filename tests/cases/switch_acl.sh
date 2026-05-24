# OpenLAN Switch UT: ACL add/list/save/reload/remove path.

net_name=tests-net-acl
sw1_name=tests-sw-acl1
sw2_name=tests-sw-acl2
vip_address=10.254.0.12

# Topology:
# - Docker mgmt network: 172.254.0.0/24
#   sw1=172.254.0.241, sw2=172.254.0.242.
# - OpenLAN service network "example": 192.61.0.0/24
#   sw1=192.61.0.1, sw2=192.61.0.2.
# - sw2 VIP:
#   lo=10.254.0.12/32, tcp/80 service.
# - Validation: add/remove ACL rules on sw2, verify CLI list, raw table state,
#   persistence across reload, and sw1 -> sw2 VIP tcp/80 and ICMP behavior.

setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.254.0.0/24 --gateway=172.254.0.1
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.254.0.241
  local crypt_secret="cb2ff088a34d"

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  wait "docker logs -f $name" Http.Start 30

  docker exec $name openlan crypt update --algorithm aes-128 --secret "$crypt_secret"
  docker exec $name openlan network --name example add --address 192.61.0.1/24
  docker exec $name openlan network --name example route add --prefix $vip_address/32 --nexthop 192.61.0.2
  docker exec $name openlan user add --name t1@example --password 123456
  docker exec $name ip link show hi-example
}

setup_sw2() {
  local name="$sw2_name"
  local address=172.254.0.242
  local crypt_secret="cb2ff088a34d"

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  wait "docker logs -f $name" Http.Start 30

  docker exec $name openlan crypt update --algorithm aes-128 --secret "$crypt_secret"
  docker exec $name openlan network --name example add --address 192.61.0.2/24
  docker exec $name openlan router address add --device lo --address $vip_address/32
  docker exec $name openlan user add --name t1@example --password 123456
  docker exec $name openlan network --name example output add --remote 172.254.0.241 --protocol udp --secret t1@example:123456 --crypt aes-128:$crypt_secret
}

setup_vip_http() {
  docker exec $sw2_name sh -c "nohup sh -c 'while true; do printf \"HTTP/1.1 200 OK\\r\\nContent-Length: 10\\r\\n\\r\\nacl-vip-80\" | socat - TCP-LISTEN:80,bind=$vip_address,reuseaddr; done' >/tmp/acl-vip-80.log 2>&1 &"
}

test_vip_reachable_before_acl() {
  check "docker exec $sw1_name openlan network --name example access ls" "172.254.0.242" 15
  check "docker exec $sw1_name ping -c 5 $vip_address" "bytes from" 20
  check "docker exec $sw1_name wget -qO- -T 3 -t 1 http://$vip_address:80" "acl-vip-80" 20
}

test_acl_add() {
  check "docker exec $sw1_name wget -qO- -T 3 -t 1 http://$vip_address:80" "acl-vip-80" 20
  docker exec $sw2_name openlan acl --name example rule add --source 192.61.0.1 --destination $vip_address --protocol tcp --dport 80

  check "docker exec $sw2_name openlan acl --name example rule list" "192.61.0.1" 10
  check "docker exec $sw2_name openlan acl --name example rule list" "$vip_address" 10
  check "docker exec $sw2_name iptables -t raw -S TT_pre-example" "AT_example" 10
  check "docker exec $sw2_name iptables -t raw -S AT_example" "192.61.0.1.*$vip_address.*tcp.*--dport 80.*DROP" 10

  if check "docker exec $sw1_name wget -qO- -T 3 -t 1 http://$vip_address:80" "acl-vip-80" 5; then
    echo "unexpected tcp/80 access success after acl add"
    return 1
  fi

  check "docker exec $sw1_name ping -c 5 $vip_address" "bytes from" 20
  docker exec $sw2_name openlan acl --name example rule add --source 192.61.0.1 --destination $vip_address --protocol icmp

  check "docker exec $sw2_name openlan acl --name example rule list" "192.61.0.1" 10
  check "docker exec $sw2_name openlan acl --name example rule list" "$vip_address" 10
  check "docker exec $sw2_name iptables -t raw -S TT_pre-example" "AT_example" 10
  check "docker exec $sw2_name iptables -t raw -S AT_example" "192.61.0.1.*$vip_address.*icmp.*DROP" 10

  if check "docker exec $sw1_name ping -c 5 $vip_address" "bytes from" 5; then
    echo "unexpected icmp access success after acl add"
    return 1
  fi
}

test_acl_save_reload() {
  docker exec $sw2_name openlan acl --name example rule save
  docker exec $sw1_name openlan reload --save
  docker exec $sw2_name openlan reload --save

  check "docker exec $sw1_name openlan network --name example access ls" "172.254.0.242" 15
  check "docker exec $sw2_name openlan acl --name example rule list" "192.61.0.1" 10
  check "docker exec $sw2_name iptables -t raw -S TT_pre-example" "AT_example" 10
  check "docker exec $sw2_name iptables -t raw -S AT_example" "192.61.0.1.*$vip_address.*tcp.*--dport 80.*DROP" 10
  check "docker exec $sw2_name iptables -t raw -S AT_example" "192.61.0.1.*$vip_address.*icmp.*DROP" 10

  if check "docker exec $sw1_name wget -qO- -T 3 -t 1 http://$vip_address:80" "acl-vip-80" 5; then
    echo "unexpected tcp/80 access success after acl reload"
    return 1
  fi

  if check "docker exec $sw1_name ping -c 5 $vip_address" "bytes from" 5; then
    echo "unexpected icmp access success after acl reload"
    return 1
  fi
}

test_acl_flush() {
  docker exec $sw2_name openlan acl --name example rule flush

  if check "docker exec $sw2_name openlan acl --name example rule list" "192.61.0.1" 3; then
    echo "unexpected acl rule remains in list after flush"
    return 1
  fi

  if check "docker exec $sw2_name iptables -t raw -S AT_example" "dport 80" 3; then
    echo "unexpected acl tcp 80 rule remains after flush"
    return 1
  fi

  if check "docker exec $sw2_name iptables -t raw -S AT_example" "icmp" 3; then
    echo "unexpected acl icmp rule remains after flush"
    return 1
  fi

  check "docker exec $sw1_name ping -c 5 $vip_address" "bytes from" 20
  check "docker exec $sw1_name wget -qO- -T 3 -t 1 http://$vip_address:80" "acl-vip-80" 20

  docker exec $sw2_name openlan acl --name example rule save
  docker exec $sw1_name openlan reload --save
  docker exec $sw2_name openlan reload --save

  if check "docker exec $sw2_name openlan acl --name example rule list" "192.61.0.1" 3; then
    echo "unexpected acl rule remains after flush reload"
    return 1
  fi

  if check "docker exec $sw2_name iptables -t raw -S AT_example" "dport 80" 3; then
    echo "unexpected acl tcp 80 rule remains after flush reload"
    return 1
  fi

  if check "docker exec $sw2_name iptables -t raw -S AT_example" "icmp" 3; then
    echo "unexpected acl icmp rule remains after flush reload"
    return 1
  fi

  check "docker exec $sw1_name ping -c 5 $vip_address" "bytes from" 20
  check "docker exec $sw1_name wget -qO- -T 3 -t 1 http://$vip_address:80" "acl-vip-80" 20
}

setup() {
  setup_net
  setup_sw1
  setup_sw2
  setup_vip_http
  test_vip_reachable_before_acl
  test_acl_add
  test_acl_save_reload
  test_acl_flush
}

main
