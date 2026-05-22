
# OpenLAN SNAT scope matrix test for network a.

net_name=tests-net-snat-scope
sw1_name=tests-sw-snat-scope.sw1
sw2_name=tests-sw-snat-scope.sw2
aca_name=tests-sw-snat-scope.aca
acb_name=tests-sw-snat-scope.acb
vpn1_name=tests-sw-snat-scope.vpn1
vpn2_name=tests-sw-snat-scope.vpn2
target_subnet_ip=10.253.0.12
pass_uplink="pw-uplink-${RANDOM}-${RANDOM}"
pass_vpn1="pw-vpn1-${RANDOM}-${RANDOM}"
pass_vpn2="pw-vpn2-${RANDOM}-${RANDOM}"
pass_ua="pw-ua-${RANDOM}-${RANDOM}"
pass_ub="pw-ub-${RANDOM}-${RANDOM}"

# Topology:
# - Docker mgmt network: 172.249.0.0/24
#   sw1=172.249.0.241, sw2=172.249.0.242.
# - sw1 virtual networks:
#   network int: 192.55.0.1/24 (uplink to sw2)
#   network a: 192.53.0.1/24 (with OpenVPN subnet 10.95.0.0/24)
#   network b: 192.54.0.1/24
# - sw2 virtual network:
#   network int: 192.55.0.2/24, loopback subnet target 10.253.0.12/32.
# - Validation:
#   default: all networks snat disabled, including int.
#   scope openvpn: a.openvpn yes, a.access no, b.openvpn no, b.access no.
#   scope local:   a.openvpn yes, a.access yes, b.openvpn no, b.access no.
#   scope enable:  a.openvpn yes, a.access yes, b.openvpn yes, b.access yes.

setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.249.0.0/24 --gateway=172.249.0.1
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.249.0.241

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  wait "docker logs -f $name" Http.Start 30

  docker exec $name openlan crypt update --algorithm xor --secret "ea64d5b0c96c"
  docker exec $name openlan network --name int add --address 192.55.0.1/24
  docker exec $name openlan network --name a add --address 192.53.0.1/24
  docker exec $name openlan network --name b add --address 192.54.0.1/24
  docker exec $name openlan network --name int snat disable
  docker exec $name openlan network --name a snat disable
  docker exec $name openlan network --name b snat disable

  docker exec $name openlan user add --name uplink@int --password "$pass_uplink"
  docker exec $name openlan user add --name vpn1@a --password "$pass_vpn1"
  docker exec $name openlan user add --name vpn2@b --password "$pass_vpn2"
  docker exec $name openlan user add --name ua@a --password "$pass_ua"
  docker exec $name openlan user add --name ub@b --password "$pass_ub"

  docker exec $name openlan network --name a route add --prefix $target_subnet_ip/32 --nexthop 192.55.0.2
  docker exec $name openlan network --name b route add --prefix $target_subnet_ip/32 --nexthop 192.55.0.2
}

setup_sw2() {
  local name="$sw2_name"
  local address=172.249.0.242

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  wait "docker logs -f $name" Http.Start 30

  docker exec $name openlan crypt update --algorithm aes-128 --secret "ea64d5b0c96c"
  docker exec $name openlan network --name int add --address 192.55.0.2/24
  docker exec $name openlan network --name int snat disable
  docker exec $name openlan router address add --device lo --address $target_subnet_ip/32
  docker exec $name openlan network --name int output add --remote 172.249.0.241 --protocol udp --secret uplink@int:$pass_uplink --crypt xor:ea64d5b0c96c
}

setup_access_a() {
  mkdir -p /opt/openlan/$aca_name/etc/openlan
  cat > /opt/openlan/$aca_name/etc/openlan/access.yaml <<EOF
protocol: tcp
crypt:
  algorithm: xor
  secret: ea64d5b0c96c
connection: 172.249.0.241
username: ua@a
password: $pass_ua
interface:
  address: 192.53.0.11/24
forward:
- $target_subnet_ip/32
EOF
  start_access $aca_name $net_name
  wait "docker logs -f $aca_name" Worker.OnSuccess 30
}

setup_access_b() {
  mkdir -p /opt/openlan/$acb_name/etc/openlan
  cat > /opt/openlan/$acb_name/etc/openlan/access.yaml <<EOF
protocol: udp
crypt:
  algorithm: xor
  secret: ea64d5b0c96c
connection: 172.249.0.241
username: ub@b
password: $pass_ub
interface:
  address: 192.54.0.11/24
forward:
- $target_subnet_ip/32
EOF
  start_access $acb_name $net_name
  wait "docker logs -f $acb_name" Worker.OnSuccess 30
}

setup_openvpn_clients() {
  local name="$sw1_name"

  docker exec $name openlan network --name a openvpn add --listen :1194 --protocol tcp --subnet 10.95.0.0/24 --dns 8.8.8.8
  docker exec $name openlan network --name a client add --user vpn1 --address 10.95.0.10
  docker exec $name openlan network --name b openvpn add --listen :1195 --protocol tcp --subnet 10.94.0.0/24 --dns 8.8.8.8
  docker exec $name openlan network --name b client add --user vpn2 --address 10.94.0.10

  mkdir -p /opt/openlan/$vpn1_name/ovpn
  docker cp $name:/var/openlan/openvpn/a/tcp1194client.ovpn /opt/openlan/$vpn1_name/ovpn/client.ovpn
  cat > /opt/openlan/$vpn1_name/ovpn/auth.txt <<EOF
vpn1@a
$pass_vpn1
EOF
  start_openvpn $vpn1_name $net_name
  wait "docker logs -f $vpn1_name" "Initialization Sequence Completed" 40

  mkdir -p /opt/openlan/$vpn2_name/ovpn
  docker cp $name:/var/openlan/openvpn/b/tcp1195client.ovpn /opt/openlan/$vpn2_name/ovpn/client.ovpn
  cat > /opt/openlan/$vpn2_name/ovpn/auth.txt <<EOF
vpn2@b
$pass_vpn2
EOF
  start_openvpn $vpn2_name $net_name
  wait "docker logs -f $vpn2_name" "Initialization Sequence Completed" 40
}

expect_ping_success() {
  local name=$1
  local dst=$2
  wait "docker exec $name ping -c 5 $dst" "bytes from" 15
}

expect_ping_fail() {
  local name=$1
  local dst=$2
  if wait "docker exec $name ping -c 5 $dst" "bytes from" 15; then
    echo "unexpected ping success: $name -> $dst"
    return 1
  fi
}

test_scope_openvpn_on_a() {
  docker exec $sw1_name openlan network --name a snat disable
  docker exec $sw1_name openlan network --name a snat enable --scope openvpn
  expect_ping_success $vpn1_name $target_subnet_ip
  expect_ping_fail $aca_name $target_subnet_ip
  expect_ping_fail $vpn2_name $target_subnet_ip
  expect_ping_fail $acb_name $target_subnet_ip
}

test_scope_local_on_a() {
  docker exec $sw1_name openlan network --name a snat disable
  docker exec $sw1_name openlan network --name a snat enable --scope local
  expect_ping_success $vpn1_name $target_subnet_ip
  expect_ping_success $aca_name $target_subnet_ip
  expect_ping_fail $vpn2_name $target_subnet_ip
  expect_ping_fail $acb_name $target_subnet_ip
}

test_scope_enable_on_a() {
  docker exec $sw1_name openlan network --name a snat disable
  docker exec $sw1_name openlan network --name a snat enable --scope enable
  expect_ping_success $vpn1_name $target_subnet_ip
  expect_ping_success $aca_name $target_subnet_ip
  expect_ping_success $vpn2_name $target_subnet_ip
  expect_ping_success $acb_name $target_subnet_ip
}

setup() {
  setup_net
  setup_sw1
  setup_sw2
  setup_access_a
  setup_access_b
  setup_openvpn_clients
  test_scope_openvpn_on_a
  test_scope_local_on_a
  test_scope_enable_on_a
}

main
