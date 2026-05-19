#!/bin/bash

# OpenLAN Access UT: authentication failure path.

source $PWD/macro.sh

network_name=tests-net-authfail

setup_net() {
  docker network inspect $network_name || {
    docker network create $network_name \
      --driver=bridge --subnet=172.253.0.0/24 --gateway=172.253.0.1
  }
}

setup_sw1() {
  local name=tests-sw-authfail
  local address=172.253.0.2

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

  docker exec $name openlan network --name example add --address 172.31.0.1/24
  docker exec $name openlan user add --name t1@example --password 123456
}

setup_ac_badpass() {
  mkdir -p /opt/openlan/tests-sw-authfail/etc/openlan/access
  cat > /opt/openlan/tests-sw-authfail/etc/openlan/access/t1.badpass.yaml <<EOF
protocol: tcp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
connection: 172.253.0.2
username: t1@example
password: wrong-password
interface:
  address: 172.31.0.11/24
EOF

  docker run -d --rm --privileged --network $network_name \
    --volume /opt/openlan/tests-sw-authfail/etc/openlan:/etc/openlan \
    --name tests-sw-authfail.acbad $IMAGE /usr/bin/openlan-access -conf /etc/openlan/access/t1.badpass.yaml

  if wait "docker logs -f tests-sw-authfail.acbad" "Worker.OnSuccess" 15; then
    echo "unexpected success with wrong password"
    return 1
  fi
}

setup() {
  setup_net
  setup_sw1
  setup_ac_badpass
}

main
