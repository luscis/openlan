#!/bin/bash

# OpenLAN route test: 3-node forwarding (sw3 -> sw2 -> sw1).

source $PWD/macro.sh

network_name=tests-net-route3

setup_net() {
  docker network inspect $network_name || {
    docker network create $network_name \
      --driver=bridge --subnet=172.251.0.0/24 --gateway=172.251.0.1
  }
}

setup_sw1() {
  local name=tests-sw-route1
  local address=172.251.0.2

  mkdir -p /opt/openlan/$name/etc/openlan/switch/network
  cat > /opt/openlan/$name/etc/openlan/switch/switch.yaml <<EOF
protocol: tcp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
EOF
  cat > /opt/openlan/$name/etc/openlan/switch/network/router.yaml <<EOF
name: router
provider: router
EOF

  docker run -d --rm --privileged --network $network_name --ip $address \
    --volume /opt/openlan/$name/etc/openlan:/etc/openlan \
    --name $name $IMAGE /usr/bin/openlan-switch -conf:dir /etc/openlan/switch

  wait "docker logs -f $name" Http.Start 30
  docker exec $name openlan network --name example add --address 172.51.0.1/24
  docker exec $name openlan router address add --device lo --address 10.251.0.11/32
  docker exec $name openlan user add --name edge1@example --password 123456
}

setup_sw2() {
  local name=tests-sw-route2
  local address=172.251.0.3

  mkdir -p /opt/openlan/$name/etc/openlan/switch/network
  cat > /opt/openlan/$name/etc/openlan/switch/switch.yaml <<EOF
protocol: tcp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
EOF
  cat > /opt/openlan/$name/etc/openlan/switch/network/router.yaml <<EOF
name: router
provider: router
EOF

  docker run -d --rm --privileged --network $network_name --ip $address \
    --volume /opt/openlan/$name/etc/openlan:/etc/openlan \
    --name $name $IMAGE /usr/bin/openlan-switch -conf:dir /etc/openlan/switch

  wait "docker logs -f $name" Http.Start 30
  docker exec $name openlan network --name example add --address 172.51.0.2/24
  docker exec $name openlan router address add --device lo --address 10.251.0.12/32
  docker exec $name openlan user add --name edge2@example --password 123457
  docker exec $name openlan network --name example output add \
    --remote 172.251.0.2 \
    --protocol tcp \
    --secret edge1@example:123456 \
    --crypt aes-128:ea64d5b0c96c
}

setup_sw3() {
  local name=tests-sw-route3
  local address=172.251.0.4

  mkdir -p /opt/openlan/$name/etc/openlan/switch/network
  cat > /opt/openlan/$name/etc/openlan/switch/switch.yaml <<EOF
protocol: tcp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
EOF
  cat > /opt/openlan/$name/etc/openlan/switch/network/router.yaml <<EOF
name: router
provider: router
EOF

  docker run -d --rm --privileged --network $network_name --ip $address \
    --volume /opt/openlan/$name/etc/openlan:/etc/openlan \
    --name $name $IMAGE /usr/bin/openlan-switch -conf:dir /etc/openlan/switch

  wait "docker logs -f $name" Http.Start 30
  docker exec $name openlan network --name example add --address 172.51.0.3/24
  docker exec $name openlan network --name example output add \
    --remote 172.251.0.3 \
    --protocol tcp \
    --secret edge2@example:123457 \
    --crypt aes-128:ea64d5b0c96c

  # Route VIP traffic via sw2 (172.51.0.2) for 3-node forwarding validation.
  docker exec $name openlan network --name example route add --prefix 10.251.0.11/32 --nexthop 172.51.0.1
  docker exec $name openlan network --name example route add --prefix 10.251.0.12/32 --nexthop 172.51.0.2
  # Open IP forwarding for route testing. 
  docker exec $name sysctl -p /etc/sysctl.d/90-openlan.conf
}

check_route() {
  docker exec tests-sw-route3 ip route show | grep "10.251.0.11"
  docker exec tests-sw-route3 ip route show | grep "10.251.0.12"
  wait "docker exec tests-sw-route3 ping -c 20 172.51.0.1" "bytes from" 25
  wait "docker exec tests-sw-route3 ping -c 20 172.51.0.2" "bytes from" 25
  wait "docker exec tests-sw-route3 ping -c 20 10.251.0.11" "bytes from" 25
  wait "docker exec tests-sw-route3 ping -c 20 10.251.0.12" "bytes from" 25
}

setup() {
  setup_net
  setup_sw1
  setup_sw2
  setup_sw3
  check_route
}

main
