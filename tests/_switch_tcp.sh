# OpenLAN Access UT.

source $PWD/macro.sh

net_name=tests-net1
sw1_name=tests-sw1
sw2_name=tests-sw2
sw3_name=tests-sw3

# Topology:
# - Docker mgmt network: 172.255.0.0/24
#   sw1=172.255.0.241, sw2=172.255.0.242, sw3=172.255.0.243.
# - OpenLAN service network "example": 192.41.0.0/24
#   sw1=192.41.0.1, sw2=192.41.0.2, sw3=192.41.0.3.
# - Forwarding links:
#   sw2 -> sw1 first, sw3 -> sw1 first, then sw3 -> sw2.
# - Validation path: verify tcp forwarding reachability between switches.

describe() {
  echo "==> scenario: switch_tcp"
  echo "    description: switch tcp path: build switches and verify tcp forwarding connectivity"
}

setup_net() {
  docker network inspect $net_name|| {
    docker network create $net_name --driver=bridge --subnet=172.255.0.0/24 --gateway=172.255.0.1
  }
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.255.0.241

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  cat > /opt/openlan/$name/etc/openlan/switch/switch.json <<EOF
{
  "protocol": "tcp",
  "crypt": {
    "algorithm": "aes-128",
    "secret": "password"
  }
}
EOF

  # Start switch:
  start_switch $name $net_name $address
  wait "docker logs -f $name" Http.Start 30

  # Add a network.
  docker exec $name openlan network --name example add --address 192.41.0.1/24
  docker exec $name openlan network ls | grep example
  # Add users
  docker exec $name openlan user add --name t1@example --password 123456
}

setup_sw2() {
  local name="$sw2_name"
  local address=172.255.0.242

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

  # Start switch:
  start_switch $name $net_name $address
  wait "docker logs -f $name" Http.Start 30

  # Add a network.
  docker exec $name openlan network --name example add --address 192.41.0.2/24
  # Add users
  docker exec $name openlan user add --name t1@example --password 123456

  # Add a output to sw1
  docker exec $name openlan network --name example output add --remote 172.255.0.241 --protocol tcp --secret t1:123456 --crypt aes-128:password
  docker exec $name openlan network --name example output ls
}


setup_sw3() {
  local name="$sw3_name"
  local address=172.255.0.243

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

  # Start switch:
  start_switch $name $net_name $address
  wait "docker logs -f $name" Http.Start 30

  # Add a network.
  docker exec $name openlan network --name example add --address 192.41.0.3/24
  # Add a output to sw1
  docker exec $name openlan network --name example output add --remote 172.255.0.241 --protocol tcp --secret t1:123456 --crypt aes-128:ea64d5b0c96c
  docker exec $name openlan network --name example output ls
}

test_ping() {
  wait "docker exec $sw2_name ping -c 15 192.41.0.1" "bytes from" 15
  if wait "docker exec $sw3_name ping -c 15 192.41.0.1" "bytes from" 15; then
    echo "unexpected ping failure from sw3 to sw1"
    return 1
  fi
  # Add a output on sw3 to sw2
  local name="$sw3_name"
  docker exec $name openlan network --name example output add --remote 172.255.0.242 --protocol tcp --secret t1:123456 --crypt aes-128:ea64d5b0c96c
  docker exec $name openlan network --name example output ls

  wait "docker exec $sw2_name ping -c 15 192.41.0.2" "bytes from" 15
  wait "docker exec $sw2_name ping -c 15 192.41.0.3" "bytes from" 15
}

setup() {
  describe
  setup_net
  setup_sw1
  setup_sw2
  setup_sw3
  test_ping
}

main
