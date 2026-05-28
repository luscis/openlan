#!/bin/bash
source tools/auto.sh

show_topology() {
  cat <<'EOF'
# Topology:
# - Docker mgmt network: 172.253.0.0/24
#   sw1=172.253.0.241, bad access client joins the same mgmt network.
# - OpenLAN service network "example": 192.31.0.0/24
#   sw1 gateway=192.31.0.1, client config asks for 192.31.0.11.
# Validation:
#   (see scenario assertions in this case)

EOF
}


# OpenLAN Access UT: authentication failure path.

export net_name=tests-net-authfail
export sw1_name=tests-sw-authfail
export ac1_badpass_name=tests-sw-authfail.acbad


setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.253.0.0/24 --gateway=172.253.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.253.0.241

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

  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name example add --address 192.31.0.1/24
  assert_cmd docker exec $name openlan user add --name t1@example --password 123456
}

setup_ac_badpass() {
  local name="$ac1_badpass_name"

  mkdir -p /opt/openlan/$name/etc/openlan
  cat > /opt/openlan/$name/etc/openlan/access.yaml <<EOF
protocol: tcp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
connection: 172.253.0.241
username: t1@example
password: wrong-password
interface:
  address: 192.31.0.11/24
EOF

  start_access $name $net_name
  assert_unexpect 15 "docker logs -f $name" "Worker.OnSuccess"
}

setup_topology() {
  setup_net
  setup_sw1
  setup_ac_badpass
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
