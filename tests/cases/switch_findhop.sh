#!/bin/bash
source tools/auto.sh

show_topology() {
  cat <<'EOF'
# Topology:
# - Diagram:
#                    sw0 VIP 10.243.0.10
#                  ^                       ^
#                  | network a             | network b
#              sw1.0 ------------------- sw1.1
#                  ^                       ^
#                  +--------- sw2 ---------+
#                findhop active-backup chooses nexthop path
# - Docker mgmt network: 172.243.0.0/24
#   sw0=172.243.0.240, sw1.0=172.243.0.241, sw1.1=172.243.0.242, sw2=172.243.0.243.
# - Service networks:
#   network a: sw0=192.53.0.1, sw1.0=192.53.0.2, sw1.1=192.53.0.4, sw2=192.53.0.3.
#   network b: sw0=192.54.0.1, sw1.1=192.54.0.2, sw2=192.54.0.3.
# - VIP:
#   sw0 lo=10.243.0.10/32.
# Validation:
#   sw2 -> sw1.0 -> sw0 uses network a, sw2 -> sw1.1 -> sw0 uses network b,
#   then findhop on sw2 uses multi-nexthop in active-backup mode.

EOF
}

# OpenLAN Switch UT: findhop multi-nexthop active-backup with two relays.

export net_name=tests-net-findhop
export sw0_name=tests-sw-findhop0
export sw10_name=tests-sw-findhop10
export sw11_name=tests-sw-findhop11
export sw2_name=tests-sw-findhop2


setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.243.0.0/24 --gateway=172.243.0.1 >/dev/null
}

setup_sw0() {
  local name="$sw0_name"
  local address=172.243.0.240

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan crypt update --algorithm aes-128 --secret cb2ff088a34d
  assert_cmd docker exec $name openlan network --name a add --address 192.53.0.1/24
  assert_cmd docker exec $name openlan network --name b add --address 192.54.0.1/24
  assert_cmd docker exec $name openlan router address add --device lo --address 10.243.0.10/32
  assert_cmd docker exec $name openlan user add --name edgea@a --password 123456
  assert_cmd docker exec $name openlan user add --name edgeb@b --password 123457
}

setup_sw10() {
  local name="$sw10_name"
  local address=172.243.0.241

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan crypt update --algorithm aes-128 --secret cb2ff088a34d
  assert_cmd docker exec $name openlan network --name a add --address 192.53.0.2/24
  assert_cmd docker exec $name openlan user add --name edgea@a --password 123456
  assert_cmd docker exec $name openlan network --name a output add --remote 172.243.0.240 --protocol tcp --secret edgea:123456 --crypt aes-128:cb2ff088a34d
}

setup_sw11() {
  local name="$sw11_name"
  local address=172.243.0.242

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan crypt update --algorithm aes-128 --secret cb2ff088a34d
  assert_cmd docker exec $name openlan network --name a add --address 192.53.0.4/24
  assert_cmd docker exec $name openlan network --name b add --address 192.54.0.2/24
  assert_cmd docker exec $name openlan user add --name edgea@a --password 123456
  assert_cmd docker exec $name openlan user add --name edgeb@b --password 123457
  assert_cmd docker exec $name openlan network --name b output add --remote 172.243.0.240 --protocol tcp --secret edgeb:123457 --crypt aes-128:cb2ff088a34d
}

setup_sw2() {
  local name="$sw2_name"
  local address=172.243.0.243

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan crypt update --algorithm aes-128 --secret cb2ff088a34d
  assert_cmd docker exec $name openlan network --name a add --address 192.53.0.3/24
  assert_cmd docker exec $name openlan network --name b add --address 192.54.0.3/24
  assert_cmd docker exec $name openlan user add --name edgea@a --password 123456
  assert_cmd docker exec $name openlan user add --name edgeb@b --password 123457

  assert_cmd docker exec $name openlan network --name a output add --remote 172.243.0.241 --protocol tcp --secret edgea:123456 --crypt aes-128:cb2ff088a34d
  assert_cmd docker exec $name openlan network --name b output add --remote 172.243.0.242 --protocol tcp --secret edgeb:123457 --crypt aes-128:cb2ff088a34d
}

recover_sw10() {
  start_switch $sw10_name $net_name 172.243.0.241
  assert_expect 30 "docker logs -f $sw10_name" "Http.Start"
  assert_match 30 "docker exec $sw10_name openlan network --name a output ls" "state: authenticated"
}

recover_sw11() {
  start_switch $sw11_name $net_name 172.243.0.242
  assert_expect 30 "docker logs -f $sw11_name" "Http.Start"
  assert_match 30 "docker exec $sw11_name openlan network --name b output ls" "state: authenticated"
}

test_path_via_sw10_network_a() {
  assert_match 15 "docker exec $sw2_name ping -c 3 192.53.0.1" "bytes from"
  assert_cmd docker exec $sw2_name openlan network --name a route add --prefix 10.243.0.10/32 --nexthop 192.53.0.1
  assert_match 20 "docker exec $sw2_name ping -c 3 10.243.0.10" "bytes from"
  assert_cmd docker exec $sw2_name openlan network --name a route rm --prefix 10.243.0.10/32
}

test_path_via_sw11_network_b() {
  # Warm up b-path handshake before route validation.
  assert_match 15 "docker exec $sw2_name ping -c 3 192.54.0.1" "bytes from"
  assert_cmd docker exec $sw2_name openlan network --name b route add --prefix 10.243.0.10/32 --nexthop 192.54.0.1
  assert_match 20 "docker exec $sw2_name ping -c 3 10.243.0.10" "bytes from"
  assert_cmd docker exec $sw2_name openlan network --name b route rm --prefix 10.243.0.10/32
}

test_findhop_active_backup() {
  assert_cmd docker exec $sw2_name openlan network --name a findhop add --findhop sw0-hop --nexthop 192.53.0.1,192.54.0.1 --check ping --mode active-backup
  assert_match 20 "docker exec $sw2_name openlan network --name a findhop ls" "192.53.0.1,192.54.0.1"

  assert_cmd docker exec $sw2_name openlan network --name a route add --prefix 10.243.0.10/32 --findhop sw0-hop
  assert_match 60 "docker exec $sw2_name ip r get 10.243.0.10" "via 192"
  assert_match 20 "docker exec $sw2_name ping -c 3 10.243.0.10" "bytes from"

  # stop sw1.0, sw2 should reach vip via sw1.1
  stop_switch $sw10_name
  assert_match 60 "docker exec $sw2_name ip r get 10.243.0.10" "192.54.0.1"
  assert_match 30 "docker exec $sw2_name ping -c 3 10.243.0.10" "bytes from"

  # stop sw1.1, start sw1.0, sw2 should reach vip via sw1.0
  stop_switch $sw11_name
  recover_sw10
  assert_match 60 "docker exec $sw2_name ip r get 10.243.0.10" "192.53.0.1"
  assert_match 30 "docker exec $sw2_name ping -c 3 10.243.0.10" "bytes from"

  assert_cmd docker exec $sw0_name openlan reload --save
  assert_cmd docker exec $sw10_name openlan reload --save
  assert_cmd docker exec $sw2_name openlan reload --save

  assert_cmd docker exec $sw2_name ip neigh flush all
  assert_match 20 "docker exec $sw2_name ping -c 3 10.243.0.10" "bytes from"
  assert_match 5 "docker exec $sw2_name openlan network --name a findhop rm --findhop sw0-hop" "checker has route"
  echo "findhop remove is blocked while route is bound, as expected":

  assert_cmd docker exec $sw2_name openlan network --name a route rm --prefix 10.243.0.10/32
  assert_unmatch 5 "docker exec $sw2_name ip r get 10.243.0.10" "via 192"
  assert_cmd docker exec $sw2_name openlan network --name a findhop rm --findhop sw0-hop
}

test_findhop_loadbalance() {
  recover_sw11
  assert_match 30 "docker exec $sw2_name openlan network --name b output ls" "state: authenticated"

  assert_cmd docker exec $sw2_name openlan network --name a findhop add --findhop sw0-hop-lb --nexthop 192.53.0.1,192.54.0.1 --check ping --mode load-balance
  assert_match 20 "docker exec $sw2_name openlan network --name a findhop ls" "load-balance"

  assert_cmd docker exec $sw2_name openlan network --name a route add --prefix 10.243.0.10/32 --findhop sw0-hop-lb
  assert_match 60 "docker exec $sw2_name ip route show" "nexthop via 192.53.0.1"
  assert_match 60 "docker exec $sw2_name ip route show" "nexthop via 192.54.0.1"
  assert_match 30 "docker exec $sw2_name ping -c 3 10.243.0.10" "bytes from"

  assert_cmd docker exec $sw2_name openlan network --name a route rm --prefix 10.243.0.10/32
  assert_cmd docker exec $sw2_name openlan network --name a findhop rm --findhop sw0-hop-lb
}

setup_topology() {
  setup_net
  setup_sw0
  setup_sw10
  setup_sw11
  setup_sw2

  assert_match 30 "docker exec $sw10_name openlan network --name a output ls" "state: authenticated"
  assert_match 30 "docker exec $sw11_name openlan network --name b output ls" "state: authenticated"

  assert_match 30 "docker exec $sw2_name openlan network --name a output ls" "state: authenticated"
  assert_match 30 "docker exec $sw2_name openlan network --name b output ls" "state: authenticated"

  test_path_via_sw10_network_a
  test_path_via_sw11_network_b
  test_findhop_active_backup
  test_findhop_loadbalance
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
