#!/bin/bash
source tools/auto.sh

show_description() {
  echo "verify ztrust enable/disable with guest and token-derived knock controls"
}

show_topology_summary() {
  cat <<'EOF'
vpn1 10.93.0.10 | v OpenVPN tcp/1194 | sw1 192.59.0.1:8081 | ZTrust guest + knock gates service access
EOF
}

show_topology() {
  cat <<'EOF'
# Topology:
# - Diagram:
#       vpn1 10.93.0.10
#             |
#             v OpenVPN tcp/1194
#       sw1 192.59.0.1:8081
#             ZTrust guest + knock gates service access
# - Docker mgmt network: 100.100.0.0/24
#   sw1=100.100.0.241.
# - OpenLAN service network "example": 192.59.0.0/24
#   sw1=192.59.0.1.
# - OpenVPN overlay on sw1:
#   tcp/1194, subnet 10.93.0.0/24, vpn1@example fixed address 10.93.0.10.
# Validation:
#   vpn1 -> sw1:8081 is reachable before ztrust; blocked after ztrust enable;
#   allowed after guest+knock; blocked after knock remove; restored when disabled.

EOF
}

# OpenLAN Switch UT: ztrust enable/guest/knock flow.

export net_name=tests-net-ztrust
export sw1_name=tests-sw-ztrust
export vpn1_name=tests-sw-ztrust.vpn1


setup_net() {
  docker network create $net_name --driver=bridge --subnet=100.100.0.0/24 --gateway=100.100.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=100.100.0.241

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name example add --address 192.59.0.1/24
  assert_match 1 "docker exec $name openlan network ls" "name: example"
  assert_cmd docker exec $name openlan user add --name vpn1@example --password 123456
  assert_cmd docker exec $name openlan user add --name vpn2@example --password 123457
  assert_cmd docker exec $name openlan user add --name vpn3@example --password 123458
  assert_cmd docker exec $name openlan user add --name vpn4@example --password 123459
  assert_cmd docker exec $name openlan network --name example openvpn add --listen :1194 --protocol tcp --subnet 10.93.0.0/24
  assert_cmd docker exec $name openlan network --name example client add --user vpn1@example --address 10.93.0.10
  assert_cmd docker exec $name openlan network --name example client add --user vpn2@example --address 10.93.0.11
}

setup_openvpn_client() {
  local name="$sw1_name"

  mkdir -p /opt/openlan/$vpn1_name/ovpn
  docker cp $name:/var/openlan/openvpn/example/tcp1194client.ovpn /opt/openlan/$vpn1_name/ovpn/client.ovpn
  cat > /opt/openlan/$vpn1_name/ovpn/auth.txt <<EOF
vpn1@example
123456
EOF

  start_openvpn $vpn1_name $net_name
  assert_expect 40 "docker logs -f $vpn1_name" "Initialization Sequence Completed"
}

setup_local_service() {
  assert_cmd docker exec $sw1_name sh -c "nohup sh -c 'while true; do printf \"HTTP/1.1 200 OK\\r\\nContent-Length: 11\\r\\n\\r\\nztrust-8081\" | socat - TCP-LISTEN:8081,reuseaddr; done' >/tmp/ztrust-8081.log 2>&1 &"
}

test_ztrust_flow() {
  assert_match 3 "docker exec $vpn1_name wget -qO- -T 3 -t 1 http://192.59.0.1:8081" "ztrust-8081"

  assert_cmd docker exec $sw1_name openlan ztrust --network example enable
  assert_match 1 "docker exec $sw1_name iptables -t mangle -S TT_pre-example" "Goto Zero Trust"
  assert_match 1 "docker exec $sw1_name iptables -t mangle -S ZT_example" "ZTrust Deny All"
  assert_unmatch 3 "docker exec $vpn1_name wget -qO- -T 3 -t 1 http://192.59.0.1:8081" "ztrust-8081"

  assert_cmd docker exec $sw1_name openlan reload --save
  
  assert_match 1 "docker exec $sw1_name iptables -t mangle -S TT_pre-example" "Goto Zero Trust"
  assert_match 1 "docker exec $sw1_name iptables -t mangle -S ZT_example" "ZTrust Deny All"
  assert_unmatch 3 "docker exec $vpn1_name wget -qO- -T 3 -t 1 http://192.59.0.1:8081" "ztrust-8081"

  assert_cmd docker exec $sw1_name openlan ztrust --network example guest add --user vpn1 --address 10.93.0.10
  assert_cmd docker exec $sw1_name openlan --token vpn2@example:123457 ztrust guest add --address 10.93.0.11
  assert_match 1 "docker exec $sw1_name openlan ztrust --network example guest ls" "vpn2@example"
  assert_match 1 "docker exec $sw1_name openlan ztrust --network example guest ls" "10.93.0.11"
  assert_match 1 "docker exec $sw1_name sh -c 'openlan --token vpn4@example:123459 ztrust guest add 2>&1 || true'" "can't find address"
  assert_fail_cmd docker exec $sw1_name openlan --token vpn3@example:123458 ztrust knock add --protocol tcp --socket 192.59.0.1:8081 --age 120
  assert_cmd docker exec $sw1_name openlan --token vpn1@example:123456 ztrust knock add --protocol tcp --socket 192.59.0.1:8081 --age 120
  
  assert_match 1 "docker exec $sw1_name openlan ztrust --network example guest ls" "vpn1@example"
  assert_match 1 "docker exec $sw1_name openlan ztrust --network example knock ls --user vpn1" "192.59.0.1:8081"
  assert_match 1 "docker exec $sw1_name openlan --token vpn1@example:123456 ztrust knock ls" "192.59.0.1:8081"
  assert_unmatch 1 "docker exec $sw1_name openlan --token vpn2@example:123457 ztrust knock ls" "192.59.0.1:8081"
  assert_match 3 "docker exec $vpn1_name wget -qO- -T 3 -t 1 http://192.59.0.1:8081" "ztrust-8081"

  assert_cmd docker exec $sw1_name openlan ztrust --network example guest rm --user vpn1
  assert_cmd docker exec $sw1_name openlan ztrust --network example disable
  assert_match 3 "docker exec $vpn1_name wget -qO- -T 3 -t 1 http://192.59.0.1:8081" "ztrust-8081"

  assert_cmd docker exec $sw1_name openlan reload --save
  assert_match 3 "docker exec $vpn1_name wget -qO- -T 3 -t 1 http://192.59.0.1:8081" "ztrust-8081"
}

setup_topology() {
  setup_net
  setup_sw1
  setup_openvpn_client
  setup_local_service
  test_ztrust_flow
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
