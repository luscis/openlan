#!/bin/bash
source tools/auto.sh

show_topology() {
  cat <<'EOF'
# Topology:
# - Docker mgmt network: 172.248.0.0/24
#   sw1=172.248.0.241, sw2=172.248.0.242.
# - OpenLAN service network "example": 192.56.0.0/24
#   sw1=192.56.0.1, sw2=192.56.0.2.
# - IPSec tunnel:
#   sw1 <-> sw2 over mgmt addresses with shared PSK.
# - Output link:
#   sw2 -> sw1 by vxlan output.
# Validation:
#   sw2 can ping/perf to sw1 on plain vxlan output (no ipsec tunnel),
#   then repeat ping/perf after enabling ipsec tunnel on the same path.

EOF
}

# OpenLAN Switch UT: IPSec output path.

export net_name=tests-net-ipsec-output
export sw1_name=tests-sw-ipsec1
export sw2_name=tests-sw-ipsec2
export ipsec_secret=ea64d5b0c96c


setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.248.0.0/24 --gateway=172.248.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.248.0.241

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  cat > /opt/openlan/$name/etc/openlan/switch/switch.json <<EOF
{
  "protocol": "udp",
  "crypt": {
    "algorithm": "aes-128",
    "secret": "$ipsec_secret"
  }
}
EOF

  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name example add --address 192.56.0.1/24
  assert_cmd docker exec $name openlan user add --name edge@example --password 123456
}

setup_sw2() {
  local name="$sw2_name"
  local address=172.248.0.242

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  cat > /opt/openlan/$name/etc/openlan/switch/switch.json <<EOF
{
  "protocol": "udp",
  "crypt": {
    "algorithm": "aes-128",
    "secret": "$ipsec_secret"
  }
}
EOF

  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name example add --address 192.56.0.2/24
  assert_cmd docker exec $name openlan user add --name edge@example --password 123456
}

setup_output() {
  assert_cmd docker exec $sw1_name openlan network --name example output add --remote 172.248.0.242 --protocol vxlan --segment 1056
  assert_cmd docker exec $sw2_name openlan network --name example output add --remote 172.248.0.241 --protocol vxlan --segment 1056
}

test_vxlan_output_ping_without_ipsec() {
  assert_match 20 "docker exec $sw2_name ping -c 3 192.56.0.1" "bytes from"

  assert_cmd docker exec $sw1_name openlan reload --save
  assert_cmd docker exec $sw2_name openlan reload --save
  assert_match 20 "docker exec $sw2_name ping -c 3 192.56.0.1" "bytes from"
}

test_vxlan_output_ping_with_ipsec() {
  assert_cmd docker exec $sw1_name openlan ipsec tunnel add --remote 172.248.0.242 --protocol vxlan --secret $ipsec_secret --localid sw1.ipsec.test --remoteid sw2.ipsec.test
  assert_cmd docker exec $sw2_name openlan ipsec tunnel add --remote 172.248.0.241 --protocol vxlan --secret $ipsec_secret --localid sw2.ipsec.test --remoteid sw1.ipsec.test
  assert_match 20 "docker exec $sw1_name openlan ipsec tunnel ls | grep 172.248.0.242" "erouted"
  assert_match 20 "docker exec $sw2_name openlan ipsec tunnel ls | grep 172.248.0.241" "erouted"
  assert_match 20 "docker exec $sw2_name ping -c 3 192.56.0.1" "bytes from"

  assert_cmd docker exec $sw1_name openlan reload --save
  assert_cmd docker exec $sw2_name openlan reload --save

  assert_match 20 "docker exec $sw1_name openlan ipsec tunnel ls | grep 172.248.0.242" "erouted"
  assert_match 20 "docker exec $sw2_name openlan ipsec tunnel ls | grep 172.248.0.241" "erouted"
  assert_match 20 "docker exec $sw2_name ping -c 3 192.56.0.1" "bytes from"
}

test_vxlan_output_perf() {
  assert_match 30 "docker exec $sw2_name ping -q -c 20 -i 0.05 -s 1200 192.56.0.1" "0% packet loss"
  assert_match 5 "docker exec $sw2_name ping -q -c 20 -i 0.05 -s 1200 192.56.0.1" "rtt min/avg/max"

  assert_cmd docker exec $sw1_name iperf3 -s -D -p 5206
  assert_match 20 "docker exec $sw2_name iperf3 -c 192.56.0.1 -p 5206 -t 5" "receiver"
  assert_cmd docker exec $sw1_name pkill -f "iperf3 -s -D -p 5206" || true

  assert_cmd docker exec $sw1_name iperf3 -s -D -p 5207
  assert_match 20 "docker exec $sw2_name iperf3 -u -c 192.56.0.1 -p 5207 -b 100M -t 5" "receiver"
  assert_cmd docker exec $sw1_name pkill -f "iperf3 -s -D -p 5207" || true
}

setup_topology() {
  setup_net
  setup_sw1
  setup_sw2
  setup_output
  test_vxlan_output_ping_without_ipsec
  test_vxlan_output_perf
  test_vxlan_output_ping_with_ipsec
  test_vxlan_output_perf
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
