#!/bin/bash
source tools/auto.sh

show_topology() {
  cat <<'EOF'
# Topology:
# - Docker mgmt network: 172.247.0.0/24
#   sw1=172.247.0.241, sw2=172.247.0.242.
# - OpenLAN service network "example": 192.57.0.0/24
#   sw1=192.57.0.1, sw2=192.57.0.2.
# - Output link:
#   sw2 -> sw1 by vxlan output.
# Validation:
# - Phase1: no ipsec tunnel, run ping + iperf3 TCP/UDP.
# - Phase2: enable ipsec tunnel, run ping + iperf3 TCP/UDP again.

EOF
}

export net_name=tests-net-ipsec-vxlan-perf
export sw1_name=tests-sw-ipsec-vxlan-perf1
export sw2_name=tests-sw-ipsec-vxlan-perf2
export ipsec_secret=ea64d5b0c96c

setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.247.0.0/24 --gateway=172.247.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.247.0.241
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
  assert_cmd docker exec $name openlan network --name example add --address 192.57.0.1/24
  assert_cmd docker exec $name openlan user add --name edge@example --password 123456
}

setup_sw2() {
  local name="$sw2_name"
  local address=172.247.0.242
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
  assert_cmd docker exec $name openlan network --name example add --address 192.57.0.2/24
  assert_cmd docker exec $name openlan user add --name edge@example --password 123456
}

setup_output() {
  assert_cmd docker exec $sw1_name openlan network --name example output add --remote 172.247.0.242 --protocol vxlan --segment 1057
  assert_cmd docker exec $sw2_name openlan network --name example output add --remote 172.247.0.241 --protocol vxlan --segment 1057
}

setup_ipsec_tunnel() {
  assert_cmd docker exec $sw1_name openlan ipsec tunnel add --remote 172.247.0.242 --protocol vxlan --secret $ipsec_secret --localid sw1.ipsec.perf --remoteid sw2.ipsec.perf
  assert_cmd docker exec $sw2_name openlan ipsec tunnel add --remote 172.247.0.241 --protocol vxlan --secret $ipsec_secret --localid sw2.ipsec.perf --remoteid sw1.ipsec.perf
  assert_match 20 "docker exec $sw1_name openlan ipsec tunnel ls | grep 172.247.0.242" "erouted"
  assert_match 20 "docker exec $sw2_name openlan ipsec tunnel ls | grep 172.247.0.241" "erouted"
}

test_ping() {
  assert_match 20 "docker exec $sw2_name ping -c 3 192.57.0.1" "bytes from"
  assert_match 30 "docker exec $sw2_name ping -q -c 20 -i 0.05 -s 1200 192.57.0.1" "0% packet loss"
  assert_match 5 "docker exec $sw2_name ping -q -c 20 -i 0.05 -s 1200 192.57.0.1" "rtt min/avg/max"
}

test_bandwidth() {
  assert_cmd docker exec $sw1_name iperf3 -s -D -p 5210
  assert_match 20 "docker exec $sw2_name iperf3 -c 192.57.0.1 -p 5210 -t 5" "receiver"
  assert_cmd docker exec $sw1_name pkill -f iperf3 || true
  assert_cmd docker exec $sw1_name iperf3 -s -D -p 5211
  assert_match 20 "docker exec $sw2_name iperf3 -u -c 192.57.0.1 -p 5211 -b 500M -t 5" "receiver"
  assert_cmd docker exec $sw1_name pkill -f iperf3 || true
}

test_phase_without_ipsec() {
  test_ping
  test_bandwidth
  assert_cmd docker exec $sw1_name openlan reload --save
  assert_cmd docker exec $sw2_name openlan reload --save
  test_ping
}

test_phase_with_ipsec() {
  setup_ipsec_tunnel
  test_ping
  test_bandwidth
  assert_cmd docker exec $sw1_name openlan reload --save
  assert_cmd docker exec $sw2_name openlan reload --save
  assert_match 20 "docker exec $sw1_name openlan ipsec tunnel ls | grep 172.247.0.242" "erouted"
  assert_match 20 "docker exec $sw2_name openlan ipsec tunnel ls | grep 172.247.0.241" "erouted"
  test_ping
}

setup_topology() {
  setup_net
  setup_sw1
  setup_sw2
  setup_output
  test_phase_without_ipsec
  test_phase_with_ipsec
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
