# OpenLAN Switch UT: findhop multi-nexthop active-backup with two relays.

net_name=tests-net-findhop
sw0_name=tests-sw-findhop0
sw10_name=tests-sw-findhop10
sw11_name=tests-sw-findhop11
sw2_name=tests-sw-findhop2

# Topology:
# - Docker mgmt network: 172.243.0.0/24
#   sw0=172.243.0.240, sw1.0=172.243.0.241, sw1.1=172.243.0.242, sw2=172.243.0.243.
# - Service networks:
#   network a: sw0=192.53.0.1, sw1.0=192.53.0.2, sw1.1=192.53.0.4, sw2=192.53.0.3.
#   network b: sw0=192.54.0.1, sw1.1=192.54.0.2, sw2=192.54.0.3.
# - VIP:
#   sw0 lo=10.243.0.10/32.
# - Validation path:
#   sw2 -> sw1.0 -> sw0 uses network a, sw2 -> sw1.1 -> sw0 uses network b,
#   then findhop on sw2 uses multi-nexthop in active-backup mode.

setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.243.0.0/24 --gateway=172.243.0.1
}

setup_sw0() {
  local name="$sw0_name"
  local address=172.243.0.240

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  wait "docker logs -f $name" Http.Start 30

  docker exec $name openlan crypt update --algorithm aes-128 --secret cb2ff088a34d
  docker exec $name openlan network --name a add --address 192.53.0.1/24
  docker exec $name openlan network --name b add --address 192.54.0.1/24
  docker exec $name openlan router address add --device lo --address 10.243.0.10/32
  docker exec $name openlan user add --name edgea@a --password 123456
  docker exec $name openlan user add --name edgeb@b --password 123457
}

setup_sw10() {
  local name="$sw10_name"
  local address=172.243.0.241

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  wait "docker logs -f $name" Http.Start 30

  docker exec $name openlan crypt update --algorithm aes-128 --secret cb2ff088a34d
  docker exec $name openlan network --name a add --address 192.53.0.2/24
  docker exec $name openlan user add --name edgea@a --password 123456
  docker exec $name openlan network --name a output add --remote 172.243.0.240 --protocol tcp --secret edgea:123456 --crypt aes-128:cb2ff088a34d
}

setup_sw11() {
  local name="$sw11_name"
  local address=172.243.0.242

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  wait "docker logs -f $name" Http.Start 30

  docker exec $name openlan crypt update --algorithm aes-128 --secret cb2ff088a34d
  docker exec $name openlan network --name a add --address 192.53.0.4/24
  docker exec $name openlan network --name b add --address 192.54.0.2/24
  docker exec $name openlan user add --name edgea@a --password 123456
  docker exec $name openlan user add --name edgeb@b --password 123457
  docker exec $name openlan network --name b output add --remote 172.243.0.240 --protocol tcp --secret edgeb:123457 --crypt aes-128:cb2ff088a34d
}

setup_sw2() {
  local name="$sw2_name"
  local address=172.243.0.243

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  wait "docker logs -f $name" Http.Start 30

  docker exec $name openlan crypt update --algorithm aes-128 --secret cb2ff088a34d
  docker exec $name openlan network --name a add --address 192.53.0.3/24
  docker exec $name openlan network --name b add --address 192.54.0.3/24
  docker exec $name openlan user add --name edgea@a --password 123456
  docker exec $name openlan user add --name edgeb@b --password 123457

  docker exec $name openlan network --name a output add --remote 172.243.0.241 --protocol tcp --secret edgea:123456 --crypt aes-128:cb2ff088a34d
  docker exec $name openlan network --name b output add --remote 172.243.0.242 --protocol tcp --secret edgeb:123457 --crypt aes-128:cb2ff088a34d
}

recover_sw10() {
  start_switch $sw10_name $net_name 172.243.0.241
  wait "docker logs -f $sw10_name" Http.Start 30
  check "docker exec $sw10_name openlan network --name a output ls" "state: authenticated" 30
}

recover_sw11() {
  start_switch $sw11_name $net_name 172.243.0.242
  wait "docker logs -f $sw11_name" Http.Start 30
  check "docker exec $sw11_name openlan network --name b output ls" "state: authenticated" 30
}

test_path_via_sw10_network_a() {
  wait "docker exec $sw2_name ping -c 6 192.53.0.1" "bytes from" 15
  docker exec $sw2_name openlan network --name a route add --prefix 10.243.0.10/32 --nexthop 192.53.0.1
  wait "docker exec $sw2_name ping -c 12 10.243.0.10" "bytes from" 20
  docker exec $sw2_name openlan network --name a route rm --prefix 10.243.0.10/32
}

test_path_via_sw11_network_b() {
  # Warm up b-path handshake before route validation.
  wait "docker exec $sw2_name ping -c 6 192.54.0.1" "bytes from" 15
  docker exec $sw2_name openlan network --name b route add --prefix 10.243.0.10/32 --nexthop 192.54.0.1
  wait "docker exec $sw2_name ping -c 12 10.243.0.10" "bytes from" 20
  docker exec $sw2_name openlan network --name b route rm --prefix 10.243.0.10/32
}

test_findhop_active_backup() {
  docker exec $sw2_name openlan network --name a findhop add --findhop sw0-hop --nexthop 192.53.0.1,192.54.0.1 --check ping --mode active-backup
  check "docker exec $sw2_name openlan network --name a findhop ls" "192.53.0.1,192.54.0.1" 20

  docker exec $sw2_name openlan network --name a route add --prefix 10.243.0.10/32 --findhop sw0-hop
  check "docker exec $sw2_name ip r get 10.243.0.10" "via 192" 60
  wait "docker exec $sw2_name ping -c 12 10.243.0.10" "bytes from" 20

  # stop sw1.0, sw2 should reach vip via sw1.1
  stop_switch $sw10_name
  check "docker exec $sw2_name ip r get 10.243.0.10" "192.54.0.1" 60
  wait "docker exec $sw2_name ping -c 12 10.243.0.10" "bytes from" 30

  # stop sw1.1, start sw1.0, sw2 should reach vip via sw1.0
  stop_switch $sw11_name
  recover_sw10
  check "docker exec $sw2_name ip r get 10.243.0.10" "192.53.0.1" 60
  wait "docker exec $sw2_name ping -c 12 10.243.0.10" "bytes from" 30

  docker exec $sw0_name openlan reload --save
  docker exec $sw10_name openlan reload --save
  docker exec $sw2_name openlan reload --save

  docker exec $sw2_name ip neigh flush all
  wait "docker exec $sw2_name ping -c 12 10.243.0.10" "bytes from" 20
  if check "docker exec $sw2_name openlan network --name a findhop rm --findhop sw0-hop" "checker has route" 5; then
    echo "findhop remove is blocked while route is bound, as expected":
  else
    echo "unexpected findhop remove behavior while route is bound"
    return 1
  fi

  docker exec $sw2_name openlan network --name a route rm --prefix 10.243.0.10/32
  if check "docker exec $sw2_name ip r get 10.243.0.10" "via 192" 10; then
    echo "unexpected route to VIP still exists after route removal"
    return 1
  fi
  docker exec $sw2_name openlan network --name a findhop rm --findhop sw0-hop
}

test_findhop_loadbalance() {
  recover_sw11
  check "docker exec $sw2_name openlan network --name b output ls" "state: authenticated" 30

  docker exec $sw2_name openlan network --name a findhop add --findhop sw0-hop-lb --nexthop 192.53.0.1,192.54.0.1 --check ping --mode load-balance
  check "docker exec $sw2_name openlan network --name a findhop ls" "load-balance" 20

  docker exec $sw2_name openlan network --name a route add --prefix 10.243.0.10/32 --findhop sw0-hop-lb
  check "docker exec $sw2_name ip route show" "nexthop via 192.53.0.1" 60
  check "docker exec $sw2_name ip route show" "nexthop via 192.54.0.1" 60
  wait "docker exec $sw2_name ping -c 20 10.243.0.10" "bytes from" 30

  docker exec $sw2_name openlan network --name a route rm --prefix 10.243.0.10/32
  docker exec $sw2_name openlan network --name a findhop rm --findhop sw0-hop-lb
}

setup() {
  setup_net
  setup_sw0
  setup_sw10
  setup_sw11
  setup_sw2

  check "docker exec $sw10_name openlan network --name a output ls" "state: authenticated" 30
  check "docker exec $sw11_name openlan network --name b output ls" "state: authenticated" 30

  check "docker exec $sw2_name openlan network --name a output ls" "state: authenticated" 30
  check "docker exec $sw2_name openlan network --name b output ls" "state: authenticated" 30

  test_path_via_sw10_network_a
  test_path_via_sw11_network_b
  test_findhop_active_backup
  test_findhop_loadbalance
}

main
