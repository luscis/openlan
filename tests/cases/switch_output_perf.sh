#!/bin/bash
source tools/auto.sh

show_topology() {
  cat <<'EOF'
# Topology:
# - Diagram:
#                         sw1 center 192.53.0.1
#                           ^                 ^
#                           | UDP output       | TCP output
#                    sw2 192.53.0.2     sw3 192.53.0.3
#                         mixed output auth, ping RTT, and bandwidth samples
# - One center switch sw1 accepts mixed output dial-ins.
# - sw2 -> sw1 over UDP output.
# - sw3 -> sw1 over TCP output.
# Validation:
# - Both outputs reach authenticated state on the same sw1.
# - UDP/TCP paths both pass connectivity checks.
# - Performance sample shows 0% loss and RTT summary.
# - Bandwidth sample is collected with iperf/iperf3 for UDP and TCP.

EOF
}

# OpenLAN Switch UT: one sw supports TCP/UDP output dial-ins.

export net_name="tests-net-output-perf-mixed"
export sw1_name="tests-sw-output-mix1"
export sw2_name="tests-sw-output-mix2"
export sw3_name="tests-sw-output-mix3"
export sw1_ip="172.253.0.241"
export sw2_ip="172.253.0.242"
export sw3_ip="172.253.0.243"
export sw1_svc="192.53.0.1"
export sw2_svc="192.53.0.2"
export sw3_svc="192.53.0.3"
export crypt_secret="ea64d5b0c96c"

cleanup_lab() {
  local net_name="$1"
  local sw1="$2"
  local sw2="$3"
  local sw3="$4"

  docker rm -f "$sw1" "$sw1-pause" "$sw1-frr" "$sw1-ipsec" >/dev/null 2>&1 || true
  docker rm -f "$sw2" "$sw2-pause" "$sw2-frr" "$sw2-ipsec" >/dev/null 2>&1 || true
  docker rm -f "$sw3" "$sw3-pause" "$sw3-frr" "$sw3-ipsec" >/dev/null 2>&1 || true
  docker network rm "$net_name" >/dev/null 2>&1 || true
}

setup_sw() {
  local name="$1"
  local net_name="$2"
  local address="$3"
  local protocol="$4"
  local secret="$5"

  mkdir -p "/opt/openlan/$name/etc/openlan/switch"
  cat > "/opt/openlan/$name/etc/openlan/switch/switch.json" <<EOF
{
  "protocol": "${protocol}",
  "crypt": {
    "algorithm": "aes-128",
    "secret": "${secret}"
  }
}
EOF
  start_switch "$name" "$net_name" "$address"
  assert_expect 30 "docker logs -f $name" "Http.Start"
}

test_bandwidth_iperf() {
  local server_node="$1"
  local client_node="$2"
  local server_ip="$3"
  local protocol="$4"
  local port="$5"

  local server_start
  local client_cmd
  local stop_cmd

  server_start="docker exec $server_node iperf3 -s -D -p $port"
  stop_cmd="docker exec $server_node pkill -f perf3"
  if [[ "$protocol" == "udp" ]]; then
    client_cmd="docker exec $client_node iperf3 -u -c $server_ip -p $port -b 500M -t 5"
  else
    client_cmd="docker exec $client_node iperf3 -c $server_ip -p $port -t 5"
  fi

  assert_cmd $server_start
  assert_match 20 "$client_cmd" "receiver"
  assert_cmd $stop_cmd || true
}

setup_net() {
  cleanup_lab "$net_name" "$sw1_name" "$sw2_name" "$sw3_name"
  assert_cmd docker network create "$net_name" --driver=bridge --subnet=172.253.0.0/24 --gateway=172.253.0.1
}

setup_sw1() {
  setup_sw "$sw1_name" "$net_name" "$sw1_ip" tcp "$crypt_secret"
  assert_cmd docker exec "$sw1_name" openlan network --name example add --address "$sw1_svc/24"
  assert_cmd docker exec "$sw1_name" openlan user add --name t1@example --password 123456 --role admin
}

setup_sw2() {
  setup_sw "$sw2_name" "$net_name" "$sw2_ip" udp "$crypt_secret"
  assert_cmd docker exec "$sw2_name" openlan network --name example add --address "$sw2_svc/24"
}

setup_sw3() {
  setup_sw "$sw3_name" "$net_name" "$sw3_ip" tcp "$crypt_secret"
  assert_cmd docker exec "$sw3_name" openlan network --name example add --address "$sw3_svc/24"
}

setup_outputs() {
  assert_cmd docker exec "$sw2_name" openlan network --name example output add --remote "$sw1_ip" --protocol udp --secret t1:123456 --crypt "aes-128:${crypt_secret}"
  assert_cmd docker exec "$sw3_name" openlan network --name example output add --remote "$sw1_ip" --protocol tcp --secret t1:123456 --crypt "aes-128:${crypt_secret}"
  assert_match 15 "docker exec $sw2_name openlan network --name example output ls" "state: authenticated"
  assert_match 15 "docker exec $sw3_name openlan network --name example output ls" "state: authenticated"
}

test_connectivity_and_latency() {
  assert_match 20 "docker exec $sw2_name ping -c 3 $sw1_svc" "bytes from"
  assert_match 20 "docker exec $sw3_name ping -c 3 $sw1_svc" "bytes from"
  assert_match 30 "docker exec $sw2_name ping -q -c 20 -i 0.05 -s 1200 $sw1_svc" "0% packet loss"
  assert_match 5 "docker exec $sw2_name ping -q -c 20 -i 0.05 -s 1200 $sw1_svc" "rtt min/avg/max"
  assert_match 30 "docker exec $sw3_name ping -q -c 20 -i 0.05 -s 1200 $sw1_svc" "0% packet loss"
  assert_match 5 "docker exec $sw3_name ping -q -c 20 -i 0.05 -s 1200 $sw1_svc" "rtt min/avg/max"
}

test_bandwidth() {
  test_bandwidth_iperf "$sw1_name" "$sw2_name" "$sw1_svc" udp 5201
  test_bandwidth_iperf "$sw1_name" "$sw2_name" "$sw1_svc" tcp 5201
  test_bandwidth_iperf "$sw1_name" "$sw3_name" "$sw1_svc" udp 5202
  test_bandwidth_iperf "$sw1_name" "$sw3_name" "$sw1_svc" tcp 5202
}

test_reload_persistence() {
  assert_cmd docker exec "$sw1_name" openlan reload --save
  assert_cmd docker exec "$sw2_name" openlan reload --save
  assert_cmd docker exec "$sw3_name" openlan reload --save
  assert_cmd docker exec "$sw2_name" ip neigh flush dev hi-example
  assert_cmd docker exec "$sw3_name" ip neigh flush dev hi-example
  assert_match 20 "docker exec $sw2_name ping -c 3 $sw1_svc" "bytes from"
  assert_match 20 "docker exec $sw3_name ping -c 3 $sw1_svc" "bytes from"
}

setup() {
  setup_net
  setup_sw1
  setup_sw2
  setup_sw3
  setup_outputs
  test_connectivity_and_latency
  test_bandwidth
  test_reload_persistence
  cleanup_lab "$net_name" "$sw1_name" "$sw2_name" "$sw3_name"
}

case "$1" in
  --topology)
    show_topology
    ;;
  *)
    main
    ;;
esac
