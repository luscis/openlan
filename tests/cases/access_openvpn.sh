#!/bin/bash
source tools/auto.sh

show_topology() {
  cat <<'EOF'
# Topology:
# - Docker mgmt network: 172.252.0.0/24
#   sw1=172.252.0.241, vpn containers join the same mgmt network.
# - OpenLAN service network "example": 192.41.0.0/24
#   sw1 gateway=192.41.0.1.
# - OpenVPN overlay:
#   tcp/1194 with subnet 10.99.0.0/24 for default cipher checks,
#   tcp/1194 with subnet 10.98.0.0/24 for SM4 cipher checks.
# Validation:
#   (see scenario assertions in this case)

EOF
}


# OpenLAN OpenVPN scenario test.

export net_name=tests-net-openvpn
export sw1_name=tests-sw-openvpn
export vpn1_name=tests-sw-openvpn.vpn1
export vpn2_name=tests-sw-openvpn.vpn2


setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.252.0.0/24 --gateway=172.252.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.252.0.241

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

  assert_cmd docker exec $name openlan network --name example add --address 192.41.0.1/24
  assert_cmd docker exec $name openlan user add --name t1@example --password 123456
}

setup_openvpn() {
  local name="$sw1_name"
  local sm4_supported=true

  assert_cmd docker exec $name openlan network --name example openvpn add --listen :1194 --protocol tcp --subnet 10.99.0.0/24 --dns 8.8.8.8 --cipher AES-128-GCM:AES-256-GCM

  assert_cmd docker exec $name test -f /var/openlan/openvpn/example/tcp1194server.conf
  assert_cmd docker exec $name test -f /var/openlan/openvpn/example/tcp1194client.ovpn

  assert_cmd docker exec $name openlan network --name example client add --user vpn1 --address 10.99.0.10
  assert_cmd docker exec $name test -f /var/openlan/openvpn/example/ccd/vpn1@example

  assert_cmd docker exec $name openlan network --name example client remove --user vpn1
  assert_cmd docker exec $name test ! -f /var/openlan/openvpn/example/ccd/vpn1@example

  mkdir -p /opt/openlan/$vpn1_name/ovpn
  docker cp $name:/var/openlan/openvpn/example/tcp1194client.ovpn /opt/openlan/$vpn1_name/ovpn/client.ovpn
  cat > /opt/openlan/$vpn1_name/ovpn/auth.txt <<EOF
t1@example
123456
EOF

  start_openvpn $vpn1_name $net_name
  assert_expect 40 "docker logs -f $vpn1_name" "Initialization Sequence Completed"
  assert_expect 40 "docker logs -f $vpn1_name" "Data Channel:"

  docker logs $vpn1_name 2>&1 | grep -E "Data Channel:.*(AES-128-GCM|AES-256-GCM)"
  assert_match 5 "docker exec $vpn1_name ping -c 3 192.41.0.1" "bytes from"

  assert_cmd docker exec $name openlan network --name example openvpn remove
  assert_fail_cmd docker exec $name openlan network --name example client add --user vpn1 --address 10.99.0.10

  # Invalid ciphers should be rejected when openvpn is disabled.
  assert_fail_cmd docker exec $name openlan network --name example openvpn add --listen :1194 --protocol tcp --subnet 10.99.0.0/24 --cipher BAD-CIPHER
}

setup_openvpn_sm4() {
  # SM4 cipher negotiation check.
  assert_cmd docker exec $sw1_name openlan network --name example openvpn add --listen :1194 --protocol tcp --subnet 10.98.0.0/24 --cipher SM4-CBC
  mkdir -p /opt/openlan/$vpn2_name/ovpn
  docker cp $sw1_name:/var/openlan/openvpn/example/tcp1194client.ovpn /opt/openlan/$vpn2_name/ovpn/client.ovpn
  cat > /opt/openlan/$vpn2_name/ovpn/auth.txt <<EOF
t1@example
123456
EOF
  start_openvpn $vpn2_name $net_name
  assert_expect 40 "docker logs -f $vpn2_name" "Initialization Sequence Completed"
  assert_expect 40 "docker logs -f $vpn2_name" "Data Channel:"
  assert_fuzzy 15 "docker logs $vpn2_name" "Data Channel:.*SM4-CBC"
  assert_fuzzy 15 "docker logs $vpn2_name" "Outgoing Data Channel:.*SM4-CBC.*initialized"
  assert_fuzzy 15 "docker logs $vpn2_name" "Incoming Data Channel:.*SM4-CBC.*initialized"

  assert_cmd docker exec $sw1_name openlan network --name example openvpn remove
  echo "SM4 cipher negotiation check passed."
}

setup_topology() {
  setup_net
  setup_sw1
  setup_openvpn
  setup_openvpn_sm4
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
