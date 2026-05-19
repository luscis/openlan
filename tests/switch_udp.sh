#!/bin/bash

# OpenLAN Switch UT: UDP transport path.

source $PWD/macro.sh

network_name=tests-net-udp

setup_net() {
  docker network inspect $network_name || {
    docker network create $network_name \
      --driver=bridge --subnet=172.254.0.0/24 --gateway=172.254.0.1
  }
}

setup_sw1() {
  local name=tests-sw-udp1
  local address=172.254.0.2

  mkdir -p /opt/openlan/$name
  mkdir -p /opt/openlan/$name/etc/openlan/switch
  cat > /opt/openlan/$name/etc/openlan/switch/switch.yaml <<EOF
protocol: udp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
EOF

  docker run -d --rm --privileged --network $network_name --ip $address \
    --volume /opt/openlan/$name/etc/openlan:/etc/openlan \
    --name $name $IMAGE /usr/bin/openlan-switch -conf:dir /etc/openlan/switch

  wait "docker logs -f $name" Http.Start 30

  docker exec $name openlan network --name example add --address 172.21.0.1/24
  docker exec $name openlan user add --name t1@example --password 123456
}

setup_sw2() {
  local name=tests-sw-udp2
  local address=172.254.0.3

  mkdir -p /opt/openlan/$name
  mkdir -p /opt/openlan/$name/etc/openlan/switch
  cat > /opt/openlan/$name/etc/openlan/switch/switch.yaml <<EOF
protocol: udp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
EOF

  docker run -d --rm --privileged --network $network_name --ip $address \
    --volume /opt/openlan/$name/etc/openlan:/etc/openlan \
    --name $name $IMAGE /usr/bin/openlan-switch -conf:dir /etc/openlan/switch

  wait "docker logs -f $name" Http.Start 30

  docker exec $name openlan network --name example add --address 172.21.0.2/24
  docker exec $name openlan network --name example output add \
    --remote 172.254.0.2 \
    --protocol udp \
    --secret t1:123456 \
    --crypt aes-128:ea64d5b0c96c
}

ping() {
  wait "docker exec tests-sw-udp2 ping -c 15 172.21.0.1" "bytes from" 20
}

setup() {
  setup_net
  setup_sw1
  setup_sw2
  ping
}

main
