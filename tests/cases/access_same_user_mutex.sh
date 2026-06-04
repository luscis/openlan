#!/bin/bash
source tools/auto.sh

show_description() {
  echo "same user multiple access logins are mutually exclusive"
}

show_topology_summary() {
  cat <<'EOF'
sw1(center) 100.100.0.241 / example | tcp access | tcp access | ac1(t1) ac2(t1)
EOF
}

show_topology() {
  cat <<'EOF'
# Topology:
# - Diagram:
#            sw1(center) 100.100.0.241 / example
#                 ^                    ^
#                 | tcp access          | tcp access
#              ac1(t1)              ac2(t1)
#                 same user login is mutually exclusive
# - Docker mgmt network: 100.100.0.0/24
#   sw1=100.100.0.241, ac1/ac2 join the same mgmt network.
# - OpenLAN service network "example": 192.41.0.0/24
#   same user logs in from ac1 and ac2.
# Validation:
#   (see scenario assertions in this case)

EOF
}


# OpenLAN Access UT: same user multi-access mutual exclusion.

export net_name=tests-net-same-user
export sw1_name=tests-sw-same-user
export ac1_name=tests-sw-same-user.ac1
export ac2_name=tests-sw-same-user.ac2


setup_net() {
  docker network create $net_name --driver=bridge --subnet=100.100.0.0/24 --gateway=100.100.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=100.100.0.241

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

  assert_cmd docker exec $name openlan network --name example add --address 192.41.0.1/24
  assert_cmd docker exec $name openlan user add --name t1@example --password 123456
}

setup_ac1() {
  local name="$ac1_name"

  mkdir -p /opt/openlan/$name/etc/openlan
  cat > /opt/openlan/$name/etc/openlan/access.yaml <<YAML
alias: ac1.alias
protocol: tcp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
connection: 100.100.0.241
username: t1@example
password: 123456
interface:
  address: 192.41.0.11/24
YAML

  start_access $name $net_name
}

setup_ac2() {
  local name="$ac2_name"

  mkdir -p /opt/openlan/$name/etc/openlan
  cat > /opt/openlan/$name/etc/openlan/access.yaml <<YAML
alias: ac2.alias
protocol: udp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
connection: 100.100.0.241
username: t1@example
password: 123456
interface:
  address: 192.41.0.12/24
YAML

  start_access $name $net_name
}

test_same_user_mutex() {
  setup_ac1
  assert_expect 30 "docker logs -f $ac1_name" "Worker.OnSuccess"
  assert_match 3 "docker exec $sw1_name openlan network --name example access ls" "ac1.alias"

  setup_ac2
  assert_expect 30 "docker logs -f $ac2_name" "Worker.OnSuccess"
  assert_match 3 "docker exec $sw1_name openlan network --name example access ls" "ac2.alias"
  assert_match 3 "docker exec $sw1_name openlan network --name example access ls" "total 1"

  # stop ac1 to avoid reconnect flipping and ensure ac2 remains valid
  docker stop $ac2_name
  assert_match 60 "docker exec $sw1_name openlan network --name example access ls" "ac1.alias"
  
  docker exec $ac1_name ping -c 3 192.41.0.1 || true
  assert_match 5 "docker exec $ac1_name ping -c 3 192.41.0.1" "bytes from"
}

setup_topology() {
  setup_net
  setup_sw1
  test_same_user_mutex
}

setup() {
  setup_topology
}

case "$1" in
  --description)
    show_description
    ;;
  --summary)
    show_topology_summary
    ;;
  --topology)
    show_topology
    ;;
  *)
    main
    ;;
esac
