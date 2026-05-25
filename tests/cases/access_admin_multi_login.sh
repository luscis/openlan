#!/bin/bash
source tools/auto.sh


# OpenLAN Access UT: admin user can login from multiple access clients.

export net_name=tests-net-admin-multi
export sw1_name=tests-sw-admin-multi
export ac1_name=tests-sw-admin-multi.ac1
export ac2_name=tests-sw-admin-multi.ac2

# Topology:
# - Docker mgmt network: 172.251.0.0/24
#   sw1=172.251.0.241, ac1/ac2 join the same mgmt network.
# - OpenLAN service network "example": 192.51.0.0/24
#   same admin user logs in from ac1 and ac2.
# - Validation path: admin multi-login should be allowed at the same time.

setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.251.0.0/24 --gateway=172.251.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.251.0.241

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  cat > /opt/openlan/$name/etc/openlan/switch/switch.json <<JSON
{
  "protocol": "tcp",
  "crypt": {
    "algorithm": "aes-128",
    "secret": "ea64d5b0c96c"
  }
}
JSON

  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name example add --address 192.51.0.1/24
  assert_cmd docker exec $name openlan user add --name admin1@example --password 123456 --role admin
}

setup_ac1() {
  local name="$ac1_name"

  mkdir -p /opt/openlan/$name/etc/openlan
  cat > /opt/openlan/$name/etc/openlan/access.yaml <<YAML
protocol: tcp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
connection: 172.251.0.241
username: admin1@example
password: 123456
interface:
  address: 192.51.0.11/24
YAML

  start_access $name $net_name
}

setup_ac2() {
  local name="$ac2_name"

  mkdir -p /opt/openlan/$name/etc/openlan
  cat > /opt/openlan/$name/etc/openlan/access.yaml <<YAML
protocol: udp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
connection: 172.251.0.241
username: admin1@example
password: 123456
interface:
  address: 192.51.0.12/24
YAML

  start_access $name $net_name
}

test_admin_multi_login() {
  setup_ac1
  assert_expect 30 "docker logs -f $ac1_name" "Worker.OnSuccess"
  assert_match 10 "docker exec $sw1_name openlan network --name example access ls" "total 1"

  setup_ac2
  assert_expect 30 "docker logs -f $ac2_name" "Worker.OnSuccess"
  assert_match 30 "docker exec $ac2_name ping -c 3 192.51.0.1" "bytes from"

  # admin role should allow concurrent sessions for the same user
  assert_match 10 "docker exec $sw1_name openlan network --name example access ls" "total 2"
  assert_match 10 "docker exec $ac1_name ping -c 3 192.51.0.1" "bytes from"
}

setup_topology() {
  setup_net
  setup_sw1
  test_admin_multi_login
}

setup() {
  setup_topology
}

main
