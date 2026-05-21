#!/bin/bash

# OpenLAN OpenVPN scenario test.

source $PWD/macro.sh

net_name=tests-net-openvpn
sw1_name=tests-sw-openvpn
vpn1_name=tests-sw-openvpn.vpn1
vpn2_name=tests-sw-openvpn.vpn2

# Topology:
# - Docker mgmt network: 172.252.0.0/24
#   sw1=172.252.0.241, vpn containers join the same mgmt network.
# - OpenLAN service network "example": 192.41.0.0/24
#   sw1 gateway=192.41.0.1.
# - OpenVPN overlay:
#   tcp/1194 with subnet 10.99.0.0/24 for default cipher checks,
#   tcp/1194 with subnet 10.98.0.0/24 for SM4 cipher checks.
# - Validation path: OpenVPN add/remove lifecycle and data channel cipher negotiation.

describe() {
  echo "==> scenario: access_openvpn"
  echo "    description: openvpn management path: add/remove OpenVPN and validate cipher negotiation"
}

setup_net() {
  docker network inspect $net_name || {
    docker network create $net_name --driver=bridge --subnet=172.252.0.0/24 --gateway=172.252.0.1
  }
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
  wait "docker logs -f $name" Http.Start 30

  docker exec $name openlan network --name example add --address 192.41.0.1/24
  docker exec $name openlan user add --name t1@example --password 123456
}

setup_openvpn() {
  local name="$sw1_name"
  local sm4_supported=true

  docker exec $name openlan network --name example openvpn add --listen :1194 --protocol tcp --subnet 10.99.0.0/24 --dns 8.8.8.8 --cipher AES-128-GCM:AES-256-GCM

  docker exec $name test -f /var/openlan/openvpn/example/tcp1194server.conf
  docker exec $name test -f /var/openlan/openvpn/example/tcp1194client.ovpn

  docker exec $name openlan network --name example client add --user vpn1 --address 10.99.0.10
  docker exec $name test -f /var/openlan/openvpn/example/ccd/vpn1@example

  docker exec $name openlan network --name example client remove --user vpn1
  docker exec $name test ! -f /var/openlan/openvpn/example/ccd/vpn1@example

  mkdir -p /opt/openlan/$vpn1_name/ovpn
  docker cp $name:/var/openlan/openvpn/example/tcp1194client.ovpn /opt/openlan/$vpn1_name/ovpn/client.ovpn
  cat > /opt/openlan/$vpn1_name/ovpn/auth.txt <<EOF
t1@example
123456
EOF

  start_openvpn $vpn1_name $net_name
  wait "docker logs -f $vpn1_name" "Initialization Sequence Completed" 40
  wait "docker logs -f $vpn1_name" "Data Channel:" 40

  docker logs $vpn1_name 2>&1 | grep -E "Data Channel:.*(AES-128-GCM|AES-256-GCM)"
  wait "docker exec $vpn1_name ping -c 5 192.41.0.1" "bytes from" 15

  docker exec $name openlan network --name example openvpn remove
  if docker exec $name openlan network --name example client add --user vpn1 --address 10.99.0.10; then
    echo "unexpected success adding VPN client after openvpn remove"
    return 1
  fi

  # Invalid ciphers should be rejected when openvpn is disabled.
  if docker exec $name openlan network --name example openvpn add \
      --listen :1194 --protocol tcp --subnet 10.99.0.0/24 --cipher BAD-CIPHER; then
    echo "unexpected success for invalid openvpn cipher"
    return 1
  fi
}

setup_openvpn_sm4() {
  local name="$sw1_name"
  local sm4_supported=true

  # SM4 cipher negotiation check (skip if runtime does not support SM4).
  if docker exec $name openlan network --name example openvpn add --listen :1194 --protocol tcp --subnet 10.98.0.0/24 --cipher SM4-CBC; then
    mkdir -p /opt/openlan/$vpn2_name/ovpn
    docker cp $name:/var/openlan/openvpn/example/tcp1194client.ovpn /opt/openlan/$vpn2_name/ovpn/client.ovpn
    cat > /opt/openlan/$vpn2_name/ovpn/auth.txt <<EOF
t1@example
123456
EOF
    start_openvpn $vpn2_name $net_name
    if wait "docker logs -f $vpn2_name" "Initialization Sequence Completed" 40; then
      wait "docker logs -f $vpn2_name" "Data Channel:" 40
      docker logs $vpn2_name 2>&1 | grep -E "Data Channel:.*(SM4-CBC)"
    else
      sm4_supported=false
      echo "SM4 login did not complete, likely runtime OpenVPN lacks SM4 support; skip strict SM4 check."
    fi

    docker stop $vpn2_name || true
    docker exec $vpn2_name openlan network --name example openvpn remove || true
  else
    sm4_supported=false
    echo "openvpn add with SM4 cipher failed, likely runtime OpenVPN lacks SM4 support; skip strict SM4 check."
  fi

  if [[ "$sm4_supported" == "true" ]]; then
    echo "SM4 cipher negotiation check passed."
  else
    echo "SM4 cipher negotiation check skipped due to lack of SM4 support."
    return 1
  fi
}

setup() {
  describe
  setup_net
  setup_sw1
  setup_openvpn
  setup_openvpn_sm4
}

main
