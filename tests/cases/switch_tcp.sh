#!/bin/bash
source tools/auto.sh

show_topology() {
  cat <<'EOF'
# Topology:
# - Diagram:
#                         sw1 192.41.0.1
#                           ^              ^
#                           | TCP output    | TCP output
#                    sw2 192.41.0.2   sw3 192.41.0.3
#                           then sw3 output is moved from sw1 to sw2
# - Docker mgmt network: 172.255.0.0/24
#   sw1=172.255.0.241, sw2=172.255.0.242, sw3=172.255.0.243.
# - OpenLAN service network "example": 192.41.0.0/24
#   sw1=192.41.0.1, sw2=192.41.0.2, sw3=192.41.0.3.
# - Forwarding links:
#   sw2 -> sw1 first, sw3 -> sw1 first, then sw3 -> sw2.
# Validation:
#   (see scenario assertions in this case)

EOF
}

# OpenLAN Access UT.

export net_name=tests-net1
export sw1_name=tests-sw1
export sw2_name=tests-sw2
export sw3_name=tests-sw3


setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.255.0.0/24 --gateway=172.255.0.1 >/dev/null
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
  assert_expect 30 "docker logs -f $name" "Http.Start"

  # Add a network.
  assert_cmd docker exec $name openlan network --name example add --address 192.41.0.1/24
  assert_match 1 "docker exec $name openlan network ls" "name: example"
  # Add users
  assert_cmd docker exec $name openlan user add --name t1@example --password 123456
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
  assert_expect 30 "docker logs -f $name" "Http.Start"

  # Add a network.
  assert_cmd docker exec $name openlan network --name example add --address 192.41.0.2/24
  # Add users
  assert_cmd docker exec $name openlan user add --name t1@example --password 123456

  # Add a output to sw1
  assert_cmd docker exec $name openlan network --name example output add --remote 172.255.0.241 --protocol tcp --secret t1:123456 --crypt aes-128:password
  assert_cmd docker exec $name openlan network --name example output ls
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
  assert_expect 30 "docker logs -f $name" "Http.Start"

  # Add a network.
  assert_cmd docker exec $name openlan network --name example add --address 192.41.0.3/24
  # Add a output to sw1
  assert_cmd docker exec $name openlan network --name example output add --remote 172.255.0.241 --protocol tcp --secret t1:123456 --crypt aes-128:ea64d5b0c96c
  assert_cmd docker exec $name openlan network --name example output ls
}

test_ping_before_sw3_sw2_output() {
  assert_match 15 "docker exec $sw2_name ping -c 3 192.41.0.1" "bytes from"
  assert_unmatch 3 "docker exec $sw3_name ping -c 3 192.41.0.1" "bytes from"
}

test_ping_after_sw3_sw2_output() {
  assert_match 15 "docker exec $sw2_name ping -c 3 192.41.0.2" "bytes from"
  assert_match 15 "docker exec $sw2_name ping -c 3 192.41.0.3" "bytes from"
}

test_ping() {
  test_ping_before_sw3_sw2_output

  # Add a output on sw3 to sw2
  local name="$sw3_name"
  assert_cmd docker exec $name openlan network --name example output add --remote 172.255.0.242 --protocol tcp --secret t1:123456 --crypt aes-128:ea64d5b0c96c
  assert_match 15 "docker exec $name openlan network --name example output ls" "state: authenticated"
  test_ping_after_sw3_sw2_output

  assert_cmd docker exec $sw1_name openlan reload --save
  assert_cmd docker exec $sw2_name openlan reload --save
  assert_cmd docker exec $sw3_name openlan reload --save

  assert_cmd docker exec $sw2_name ip neigh flush dev hi-example
  test_ping_after_sw3_sw2_output

  # Remove sw3->sw2 output and verify sw2 cannot reach sw3 anymore.
  # Output link naming for udp/tcp is "<protocol>:<remote>:<user>".
  local dev="tcp:172.255.0.242:t1"
  assert_cmd docker exec $sw3_name openlan network --name example output rm --device "$dev"
  assert_unmatch 20 "docker exec $sw2_name ping -c 3 192.41.0.3" "bytes from"
}

setup_topology() {
  setup_net
  setup_sw1
  setup_sw2
  setup_sw3
}

setup() {
  setup_topology
  test_ping
}

case "$1" in
  --topology)
    show_topology
    ;;
  *)
    main
    ;;
esac
