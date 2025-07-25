# OpenLAN Access UT.

source $PWD/macro.sh

network_name=net1

setup_net() {
  docker network inspect $network_name|| {
    docker network create $network_name\
      --driver=bridge --subnet=172.255.0.0/24 --gateway=172.255.0.1
  }
}

setup_sw1() {
  local name=sw1
  local address=172.255.0.2

  mkdir -p /opt/openlan/$name
  mkdir -p /opt/openlan/$name/etc/openlan/switch
  cat > /opt/openlan/$name/etc/openlan/switch/switch.yaml <<EOF
protocol: tcp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
EOF

  # Start switch:
  docker run -d --rm --privileged --network $network_name --ip $address \
    --volume /opt/openlan/$name/etc/openlan:/etc/openlan \
    --name $name $IMAGE /usr/bin/openlan-switch -conf:dir /etc/openlan/switch

  wait "docker logs -f $name" Http.Start 30

  # Add a network.
  docker exec $name openlan network add --name example --address 172.11.0.1/24
  # Add users
  docker exec $name openlan user add --name t1@example --password 123456
}

setup_sw2() {
  local name=sw2
  local address=172.255.0.3

  mkdir -p /opt/openlan/$name
  mkdir -p /opt/openlan/$name/etc/openlan/switch
  cat > /opt/openlan/$name/etc/openlan/switch/switch.yaml <<EOF
protocol: tcp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
EOF

  # Start switch:
  docker run -d --rm --privileged --network $network_name --ip $address \
    --volume /opt/openlan/$name/etc/openlan:/etc/openlan \
    --name $name $IMAGE /usr/bin/openlan-switch -conf:dir /etc/openlan/switch

  wait "docker logs -f $name" Http.Start 30

  # Add a network.
  docker exec $name openlan network add --name example --address 172.11.0.2/24
  # Add a output
  docker exec $name openlan network --name example output add --remote 172.255.0.2 --protocol tcp --secret t1:123456 --crypt aes-128:ea64d5b0c96c
  docker exec $name openlan network --name example output ls
}


ping() {
  wait "docker exec sw2 ping -c 15 172.11.0.1" "bytes from" 15
}

setup() {
  setup_net
  setup_sw1
  setup_sw2
  ping
}

cleanup() {
  # Stop containd
  docker stop sw1
  docker stop sw2
  # Cleanup files
  rm -rvf /opt/openlan/sw1
  rm -rvf /opt/openlan/sw2
}

main