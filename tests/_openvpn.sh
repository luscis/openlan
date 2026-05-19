#!/bin/bash

# OpenLAN OpenVPN scenario test.

source $PWD/macro.sh

network_name=tests-net-openvpn

setup_net() {
  docker network inspect $network_name || {
    docker network create $network_name \
      --driver=bridge --subnet=172.252.0.0/24 --gateway=172.252.0.1
  }
}

setup_sw1() {
  local name=tests-sw-openvpn
  local address=172.252.0.2

  mkdir -p /opt/openlan/$name
  mkdir -p /opt/openlan/$name/etc/openlan/switch
  cat > /opt/openlan/$name/etc/openlan/switch/switch.yaml <<EOF
protocol: tcp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
EOF

  docker run -d --rm --privileged --network $network_name --ip $address \
    --volume /opt/openlan/$name/etc/openlan:/etc/openlan \
    --name $name $IMAGE /usr/bin/openlan-switch -conf:dir /etc/openlan/switch

  wait "docker logs -f $name" Http.Start 30
  docker exec $name openlan network --name example add --address 172.41.0.1/24
  docker exec $name openlan user add --name t1@example --password 123456
}

setup_openvpn() {
  local name=tests-sw-openvpn
  local sm4_supported=true
  docker exec $name openlan network --name example openvpn add \
    --listen 172.252.0.2:1194 \
    --protocol tcp \
    --subnet 10.99.0.0/24 \
    --dns 8.8.8.8 \
    --cipher AES-128-GCM:AES-256-GCM

  docker exec $name test -f /var/openlan/openvpn/example/tcp1194server.conf
  docker exec $name test -f /var/openlan/openvpn/example/tcp1194client.ovpn

  docker exec $name openlan network --name example client add --user vpn1 --address 10.99.0.10
  docker exec $name test -f /var/openlan/openvpn/example/ccd/vpn1@example

  docker exec $name openlan network --name example client remove --user vpn1
  docker exec $name test ! -f /var/openlan/openvpn/example/ccd/vpn1@example

  mkdir -p /opt/openlan/tests-sw-openvpn/ovpn
  docker cp $name:/var/openlan/openvpn/example/tcp1194client.ovpn /opt/openlan/tests-sw-openvpn/ovpn/example.ovpn
  cat > /opt/openlan/tests-sw-openvpn/ovpn/auth.txt <<EOF
t1@example
123456
EOF

  docker run -d --rm --cap-add=NET_ADMIN --device /dev/net/tun --network $network_name \
    --volume /opt/openlan/tests-sw-openvpn/ovpn:/ovpn \
    --name tests-ovpn-client $IMAGE /bin/sh -c \
    "openvpn --config /ovpn/example.ovpn --auth-user-pass /ovpn/auth.txt --verb 3"

  wait "docker logs -f tests-ovpn-client" "Initialization Sequence Completed" 40
  wait "docker logs -f tests-ovpn-client" "Data Channel:" 40
  docker logs tests-ovpn-client 2>&1 | grep -E "Data Channel:.*(AES-128-GCM|AES-256-GCM)"
  wait "docker exec tests-ovpn-client ping -c 5 172.41.0.1" "bytes from" 15

  docker exec $name openlan network --name example openvpn remove
  if docker exec $name openlan network --name example client add --user vpn1 --address 10.99.0.10; then
    echo "unexpected success adding VPN client after openvpn remove"
    return 1
  fi

  # Invalid ciphers should be rejected when openvpn is disabled.
  if docker exec $name openlan network --name example openvpn add \
      --listen 172.252.0.2:1194 --protocol tcp --subnet 10.99.0.0/24 --cipher BAD-CIPHER; then
    echo "unexpected success for invalid openvpn cipher"
    return 1
  fi

  # SM4 cipher negotiation check (skip if runtime does not support SM4).
  if docker exec $name openlan network --name example openvpn add \
      --listen 172.252.0.2:1194 --protocol tcp --subnet 10.98.0.0/24 --cipher SM4-CBC; then
    docker cp $name:/var/openlan/openvpn/example/tcp1194client.ovpn /opt/openlan/tests-sw-openvpn/ovpn/example-sm4.ovpn
    cat > /opt/openlan/tests-sw-openvpn/ovpn/auth-sm4.txt <<EOF
t1@example
123456
EOF
    docker run -d --rm --cap-add=NET_ADMIN --device /dev/net/tun --network $network_name \
      --volume /opt/openlan/tests-sw-openvpn/ovpn:/ovpn \
      --name tests-ovpn-client-sm4 $IMAGE /bin/sh -c \
      "openvpn --config /ovpn/example-sm4.ovpn --auth-user-pass /ovpn/auth-sm4.txt --verb 3"

    if wait "docker logs -f tests-ovpn-client-sm4" "Initialization Sequence Completed" 40; then
      wait "docker logs -f tests-ovpn-client-sm4" "Data Channel:" 40
      docker logs tests-ovpn-client-sm4 2>&1 | grep -E "Data Channel:.*(SM4-CBC)"
    else
      sm4_supported=false
      echo "SM4 login did not complete, likely runtime OpenVPN lacks SM4 support; skip strict SM4 check."
    fi

    docker stop tests-ovpn-client-sm4 || true
    docker exec $name openlan network --name example openvpn remove || true
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
  setup_net
  setup_sw1
  setup_openvpn
}

main
