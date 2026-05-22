
# OpenLAN OpenVPN SNAT test: VPN client reaches sw2 VIP via sw1 SNAT.

net_name=tests-net-openvpn-snat-vip
sw1_name=tests-sw-openvpn-snat-vip.sw1
sw2_name=tests-sw-openvpn-snat-vip.sw2
vpn1_name=tests-sw-openvpn-snat-vip.vpn1

# Topology:
# - Docker mgmt network: 172.250.0.0/24
#   sw1=172.250.0.241, sw2=172.250.0.242.
# - OpenLAN service network "example": 192.52.0.0/24
#   sw1=192.52.0.1, sw2=192.52.0.2.
# - sw2 VIP:
#   lo=10.252.0.12/32.
# - OpenVPN overlay on sw1:
#   tcp/1194, subnet 10.96.0.0/24, vpn1@example fixed address 10.96.0.10.
# - Validation:
#   vpn client reaches sw2 VIP (10.252.0.12) through sw1 with SNAT enabled.

setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.250.0.0/24 --gateway=172.250.0.1
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.250.0.241
  local crypt_secret="ea64d5b0c96c"

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  wait "docker logs -f $name" Http.Start 30

  docker exec $name openlan crypt update --algorithm aes-128 --secret "$crypt_secret"
  docker exec $name openlan network --name example add --address 192.52.0.1/24
  docker exec $name openlan user add --name uplink@example --password 123456
  docker exec $name openlan user add --name vpn1@example --password 123456
}

setup_sw2() {
  local name="$sw2_name"
  local address=172.250.0.242
  local crypt_secret="ea64d5b0c96c"

  mkdir -p /opt/openlan/$name/etc/openlan/switch

  start_switch $name $net_name $address
  wait "docker logs -f $name" Http.Start 30

  docker exec $name openlan crypt update --algorithm aes-128 --secret "$crypt_secret"
  docker exec $name openlan network --name example add --address 192.52.0.2/24 
  docker exec $name openlan router address add --device lo --address 10.252.0.12/32
  docker exec $name openlan user add --name uplink@example --password 123456

  # sw2 connects to sw1 to build forwarding path.
  docker exec $name openlan network --name example output add --remote 172.250.0.241 --protocol tcp --secret uplink@example:123456 --crypt aes-128:$crypt_secret
}

setup_openvpn() {
  local name="$sw1_name"

  docker exec $name openlan network --name example route add --prefix 10.252.0.12/32 --nexthop 192.52.0.2
  docker exec $name openlan network --name example openvpn add --listen :1194 --protocol tcp --subnet 10.96.0.0/24 --dns 8.8.8.8
  docker exec $name openlan network --name example client add --user vpn1 --address 10.96.0.10

  mkdir -p /opt/openlan/$vpn1_name/ovpn
  docker cp $name:/var/openlan/openvpn/example/tcp1194client.ovpn /opt/openlan/$vpn1_name/ovpn/client.ovpn
  cat > /opt/openlan/$vpn1_name/ovpn/auth.txt <<EOF
vpn1@example
123456
EOF

  start_openvpn $vpn1_name $net_name
  wait "docker logs -f $vpn1_name" "Initialization Sequence Completed" 40
}

test_vpn_to_vip() {
  # Disable SNAT to verify that VIP is not reachable without SNAT.
  docker exec $sw1_name openlan network --name example snat disable
  if wait "docker exec $vpn1_name ping -c 10 10.252.0.12" "bytes from" 20; then
    echo "unexpected success pinging VIP before SNAT enabled"
    return 1
  fi

  # Enable SNAT for VPN subnet egress via sw1.
  docker exec $sw1_name openlan network --name example snat enable --scope openvpn
  wait "docker exec $vpn1_name ping -c 10 10.252.0.12" "bytes from" 20
}

setup() {
  setup_net
  setup_sw1
  setup_sw2
  setup_openvpn
  test_vpn_to_vip
}

main
