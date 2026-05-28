#!/bin/bash
source tools/auto.sh

show_topology() {
  cat <<'EOF'
# Topology:
# - Docker mgmt network: 172.254.0.0/24
#   sw1=172.254.0.241, sw2=172.254.0.242.
# - OpenLAN service network "example": 192.51.0.0/24
#   sw1=192.51.0.1, sw2=192.51.0.2.
# - Forwarding link:
#   sw2 -> sw1 over UDP output.
# Validation:
#   (see scenario assertions in this case)

EOF
}


# OpenLAN Switch UT: UDP transport path.

export net_name=tests-net-udp
export sw1_name=tests-sw-udp1
export sw2_name=tests-sw-udp2


setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.254.0.0/24 --gateway=172.254.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.254.0.241

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  cat > /opt/openlan/$name/etc/openlan/switch/switch.json <<EOF
{
  "protocol": "udp",
  "crypt": {
    "algorithm": "aes-128",
    "secret": "ea64d5b0c96c"
  }
}
EOF
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name example add --address 192.51.0.1/24
  assert_match 1 "docker exec $name openlan network ls" "name: example"
  assert_cmd docker exec $name openlan user add --name t1@example --password 123456
}

setup_sw2() {
  local name="$sw2_name"
  local address=172.254.0.242

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  cat > /opt/openlan/$name/etc/openlan/switch/switch.json <<EOF
{
  "protocol": "udp",
  "crypt": {
    "algorithm": "aes-128",
    "secret": "ea64d5b0c96c"
  }
}
EOF
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name example add --address 192.51.0.2/24
  # add a output to sw1
  assert_cmd docker exec $name openlan network --name example output add --remote 172.254.0.241 --protocol udp --secret t1:123456 --crypt aes-128:ea64d5b0c96c
}

test_ping() {
  assert_match 15 "docker exec $sw2_name openlan network --name example output ls" "state: authenticated"
  assert_match 20 "docker exec $sw2_name ping -c 3 192.51.0.1" "bytes from"

  assert_cmd docker exec $sw1_name openlan reload --save
  assert_cmd docker exec $sw2_name openlan reload --save

  assert_cmd docker exec $sw2_name ip neigh flush dev hi-example
  assert_match 15 "docker exec $sw2_name openlan network --name example output ls" "state: authenticated"
  assert_match 20 "docker exec $sw2_name ping -c 3 192.51.0.1" "bytes from"
}

setup_topology() {
  setup_net
  setup_sw1
  setup_sw2
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
