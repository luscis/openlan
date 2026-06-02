#!/bin/bash
source tools/auto.sh

show_topology() {
  cat <<'EOF'
# Topology:
# - Diagram:
#            sw1(center) 172.255.0.241 / 192.11.0.1
#                 ^                    ^
#                 | tcp access          | tcp access
#         ac1 192.11.0.11       ac2 192.11.0.12
#                 both access clients join example network
# - Docker mgmt network: 172.255.0.0/24
#   sw1=172.255.0.241, ac1/ac2 join the same mgmt network.
# - OpenLAN service network "example": 192.11.0.0/24
#   sw1 gateway=192.11.0.1, ac1=192.11.0.11, ac2=192.11.0.12.
# Validation:
#   (see scenario assertions in this case)

EOF
}

# OpenLAN Access UT.

export net_name=tests-net1
export sw1_name=tests-sw1
export ac1_name=tests-sw1.ac1
export ac2_name=tests-sw1.ac2
export crypt_secret_v1=ea64d5b0c96c
export crypt_secret_v2=ea64d5b0c96d


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
    "secret": "$crypt_secret_v1"
  }
}
EOF

  # Start switch: tests-sw1
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  # Add a network.
  assert_cmd docker exec $name openlan network --name example add --address 192.11.0.1/24
  # Add users
  assert_cmd docker exec $name openlan user add --name t1@example --password 123456
  assert_cmd docker exec $name openlan user add --name t2@example --password 123457
}

setup_ac1() {
  local name="$ac1_name"
  local secret="${1:-$crypt_secret_v1}"

  mkdir -p /opt/openlan/$name/etc/openlan
  # Start access: ac1
  cat > /opt/openlan/$name/etc/openlan/access.yaml <<EOF
protocol: tcp
crypt:
  algorithm: aes-128
  secret: $secret
connection: 172.255.0.241
username: t1@example
password: 123456
interface:
  address: 192.11.0.11/24
EOF
  start_access $name $net_name
}

setup_ac2() {
  local name="$ac2_name"
  local secret="${1:-$crypt_secret_v1}"

  mkdir -p /opt/openlan/$name/etc/openlan
  # Start access: ac2
  cat > /opt/openlan/$name/etc/openlan/access.yaml <<EOF
protocol: udp
crypt:
  algorithm: aes-128
  secret: $secret
connection: 172.255.0.241
username: t2@example
password: 123457
interface:
  address: 192.11.0.12/24
EOF
  start_access $name $net_name
}

test_ping() {
  assert_expect 30 "docker logs -f $ac1_name" "Worker.OnSuccess"
  assert_expect 30 "docker logs -f $ac2_name" "Worker.OnSuccess"

  assert_match 5 "docker exec $ac1_name ping -c 3 192.11.0.1" "bytes from"
  assert_match 5 "docker exec $ac2_name ping -c 3 192.11.0.12" "bytes from"
}

test_crypt_update() {
  assert_cmd docker exec $sw1_name openlan crypt update --algorithm aes-128 --secret "$crypt_secret_v2"
  assert_match 1 "docker exec $sw1_name openlan crypt ls" "secret: $crypt_secret_v2"

  docker stop $ac1_name
  docker stop $ac2_name

  setup_ac1 "$crypt_secret_v1"
  assert_expect 30 "docker logs -f $ac1_name" "SocketClientImpl.Try"

  docker stop $ac1_name
  setup_ac1 "$crypt_secret_v2"
  assert_expect 30 "docker logs -f $ac1_name" "Worker.OnSuccess"

  setup_ac2 "$crypt_secret_v2"
  assert_expect 30 "docker logs -f $ac2_name" "Worker.OnSuccess"
  test_ping
}

setup_topology() {
  setup_net
  setup_sw1
  setup_ac1
  setup_ac2
  test_ping
  test_crypt_update
}

setup() {
  setup_topology
}

case "$1" in
  --topology)
    show_topology
    ;;
  *)
    main
    ;;
esac
