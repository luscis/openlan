#!/bin/bash
source tools/auto.sh

show_description() {
  echo "sample OpenVPN tcp/udp latency and bandwidth performance"
}

show_topology_summary() {
  cat <<'EOF'
sw1(center) 192.55.0.1 | OpenVPN tcp/1194 + udp/1195 | vpn1 10.95.0.0/24 | ping RTT and iperf3 TCP/UDP samples
EOF
}

show_topology() {
  cat <<'EOF'
# Topology:
# - Diagram:
#            sw1(center) 192.55.0.1
#                 ^
#                 | OpenVPN tcp/1194 + udp/1195
#              vpn1 10.95.0.0/24
#                 | ping RTT and iperf3 TCP/UDP samples
# - Docker mgmt network: 100.100.0.0/24
#   sw1=100.100.0.241, vpn1 joins the same mgmt network.
# - OpenLAN service network "example": 192.55.0.0/24
#   sw1 gateway=192.55.0.1.
# - OpenVPN overlay on sw1:
#   tcp/1194 and udp/1195, subnet 10.95.0.0/24.
# Validation:
# - vpn1 reaches sw1 service IP through OpenVPN.
# - RTT sample shows 0% loss and summary line.
# - Bandwidth sample is collected with iperf3 for both TCP and UDP.

EOF
}

export net_name="tests-net-openvpn-perf"
export sw1_name="tests-sw-openvpn-perf.sw1"
export vpn1_name="tests-sw-openvpn-perf.vpn1"
export sw1_ip="100.100.0.241"
export sw1_svc="192.55.0.1"
export crypt_secret="ea64d5b0c96c"

setup_net() {
  docker network rm "$net_name" >/dev/null 2>&1 || true
  assert_cmd docker network create "$net_name" --driver=bridge --subnet=100.100.0.0/24 --gateway=100.100.0.1
}

setup_sw1() {
  mkdir -p /opt/openlan/$sw1_name/etc/openlan/switch
  start_switch "$sw1_name" "$net_name" "$sw1_ip"
  assert_expect 30 "docker logs -f $sw1_name" "Http.Start"
  assert_cmd docker exec "$sw1_name" openlan crypt update --algorithm aes-128 --secret "$crypt_secret"
  assert_cmd docker exec "$sw1_name" openlan network --name example add --address "$sw1_svc/24"
  assert_cmd docker exec "$sw1_name" openlan user add --name vpn1@example --password 123456
}

setup_openvpn() {
  local protocol="$1"
  local listen="$2"
  assert_cmd docker exec "$sw1_name" openlan network --name example openvpn add --listen "$listen" --protocol "$protocol" --subnet 10.95.0.0/24 --dns 8.8.8.8
  mkdir -p /opt/openlan/$vpn1_name/ovpn
  assert_cmd docker cp "$sw1_name:/var/openlan/openvpn/example/${protocol}${listen#:}client.ovpn" "/opt/openlan/$vpn1_name/ovpn/client.ovpn"
  cat > /opt/openlan/$vpn1_name/ovpn/auth.txt <<EOF
vpn1@example
123456
EOF
  docker rm -f "$vpn1_name" >/dev/null 2>&1 || true
  start_openvpn "$vpn1_name" "$net_name"
  assert_expect 40 "docker logs -f $vpn1_name" "Initialization Sequence Completed"
}

test_openvpn_latency() {
  assert_match 20 "docker exec $vpn1_name ping -c 3 $sw1_svc" "bytes from"
  assert_match 30 "docker exec $vpn1_name ping -q -c 20 -i 0.05 -s 1200 $sw1_svc" "0% packet loss"
  assert_match 5 "docker exec $vpn1_name ping -q -c 20 -i 0.05 -s 1200 $sw1_svc" "rtt min/avg/max"
}

test_openvpn_bandwidth() {
  local server_start
  local client_cmd
  local stop_cmd

  server_start="docker exec $sw1_name iperf3 -s -D -p 5203"
  client_cmd="docker exec $vpn1_name iperf3 -c $sw1_svc -p 5203 -t 5"
  stop_cmd="docker exec $sw1_name pkill -f iperf3"

  assert_cmd $server_start
  assert_match 20 "$client_cmd" "receiver"
  assert_cmd $stop_cmd || true
}

test_reload_persistence() {
  assert_cmd docker exec "$sw1_name" openlan reload --save
  assert_match 20 "docker exec $vpn1_name ping -c 3 $sw1_svc" "bytes from"
}

test_openvpn_protocol_perf() {
  local protocol="$1"
  local listen="$2"
  setup_openvpn "$protocol" "$listen"
  test_openvpn_latency
  test_openvpn_bandwidth
  test_reload_persistence
  assert_cmd docker exec "$sw1_name" openlan network --name example openvpn remove
  docker rm -f "$vpn1_name" >/dev/null 2>&1 || true
}

setup() {
  setup_net
  setup_sw1
  test_openvpn_protocol_perf tcp :1194
  test_openvpn_protocol_perf udp :1195
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
