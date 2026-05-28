#!/bin/bash
source tools/auto.sh

show_topology() {
  cat <<'EOF'
# Topology:
# - Docker mgmt network: 172.248.0.0/24
#   sw1=172.248.0.241 (name proxy client),
#   sw2=172.248.0.242 (upstream dns A),
#   sw3=172.248.0.243 (upstream dns B).
# - OpenLAN service network "example": 192.55.0.0/24
#   sw1=192.55.0.1, sw2=192.55.0.2, sw3=192.55.0.3,
#   with sw2/sw3 outputs to sw1.
# Validation:
#   sw1 nslookup domain_a/domain_b -> sw1 openceci(name) -> sw2/sw3 dnsmasq.

EOF
}

# OpenLAN Proxy UT: Ceci NAME proxy with multiple backends by domain match.

export net_name=tests-net-proxy-name-backends
export sw1_name=tests-sw-proxy-name-backends1
export sw2_name=tests-sw-proxy-name-backends2
export sw3_name=tests-sw-proxy-name-backends3
export name_listen=127.0.0.1:1054
export name_domain_a=proxy-name-a.test
export name_domain_b=proxy-name-b.test
export name_answer_a=192.55.0.2
export name_answer_b=192.55.0.3
export upstream_dns_a=192.55.0.2:5353
export upstream_dns_b=192.55.0.3:5353


setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.248.0.0/24 --gateway=172.248.0.1 >/dev/null
}

setup_switch() {
  local name="$1"
  local net="$2"
  local address="$3"
  local overlay="$4"

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  cat > /opt/openlan/$name/etc/openlan/switch/switch.json <<EOS
{
  "protocol": "tcp",
  "crypt": {
    "algorithm": "aes-128",
    "secret": "ea64d5b0c96c"
  }
}
EOS

  start_switch $name $net $address
  assert_expect 30 "docker logs -f $name" "Http.Start"
  assert_cmd docker exec $name openlan network --name example add --address $overlay/24
}

setup_sw1() {
  setup_switch "$sw1_name" "$net_name" "172.248.0.241" "192.55.0.1"
  assert_cmd docker exec $sw1_name openlan user add --name t1@example --password 123456
  assert_cmd docker exec $sw1_name openlan user add --name t2@example --password 123456
}

setup_sw2() {
  setup_switch "$sw2_name" "$net_name" "172.248.0.242" "192.55.0.2"
  assert_cmd docker exec $sw2_name openlan network --name example output add --remote 172.248.0.241 --protocol tcp --secret t1@example:123456 --crypt aes-128:ea64d5b0c96c
  assert_match 20 "docker exec $sw2_name openlan network --name example output ls" "state: authenticated"
}

setup_sw3() {
  setup_switch "$sw3_name" "$net_name" "172.248.0.243" "192.55.0.3"
  assert_cmd docker exec $sw3_name openlan network --name example output add --remote 172.248.0.241 --protocol tcp --secret t2@example:123456 --crypt aes-128:ea64d5b0c96c
  assert_match 20 "docker exec $sw3_name openlan network --name example output ls" "state: authenticated"
}

setup_upstream_dns() {
  assert_cmd docker exec $sw2_name sh -c "nohup dnsmasq --no-daemon --port=5353 --listen-address=192.55.0.2 --bind-interfaces --address=/$name_domain_a/$name_answer_a >/tmp/proxy-name-a-dnsmasq.log 2>&1 &"
  assert_cmd docker exec $sw3_name sh -c "nohup dnsmasq --no-daemon --port=5353 --listen-address=192.55.0.3 --bind-interfaces --address=/$name_domain_b/$name_answer_b >/tmp/proxy-name-b-dnsmasq.log 2>&1 &"
  assert_match 20 "docker exec $sw1_name ping -c 3 192.55.0.2" "bytes from"
  assert_match 20 "docker exec $sw1_name ping -c 3 192.55.0.3" "bytes from"
}

setup_name_proxy() {
  assert_cmd docker exec $sw1_name sh -c "cat > /var/openlan/ceci/$name_listen.yaml <<EOS
listen: $name_listen
nameto: 8.8.8.8
backends:
  - server: 192.55.0.2
    match:
      - $name_domain_a
    nameto: $upstream_dns_a
  - server: 192.55.0.3
    match:
      - $name_domain_b
    nameto: $upstream_dns_b
EOS"
  start_name_proxy_backends
}

start_name_proxy_backends() {
  assert_cmd docker exec $sw1_name sh -c "nohup /usr/bin/openceci -mode name -conf /var/openlan/ceci/$name_listen.yaml -log:file /var/openlan/ceci/$name_listen.log -write-pid /var/openlan/ceci/$name_listen.pid >/tmp/proxy-name-backends.log 2>&1 &"
  assert_match 20 "docker exec $sw1_name cat /var/openlan/ceci/$name_listen.log" "NameProxy.StartDNS on $name_listen"
}

test_name_proxy_backends() {
  assert_fuzzy 20 "docker exec $sw1_name nslookup -port=1054 $name_domain_a 127.0.0.1" "Address: $name_answer_a"
  assert_fuzzy 20 "docker exec $sw1_name nslookup -port=1054 $name_domain_b 127.0.0.1" "Address: $name_answer_b"
  assert_fuzzy 20 "docker exec $sw1_name ip route show" "$name_answer_a via 192.55.0.2"
  assert_fuzzy 20 "docker exec $sw1_name ip route show" "$name_answer_b via 192.55.0.3"
  assert_fuzzy 20 "docker exec $sw1_name cat /var/openlan/ceci/$name_listen.log" "NameProxy.handleDNS $upstream_dns_a <-"
  assert_fuzzy 20 "docker exec $sw1_name cat /var/openlan/ceci/$name_listen.log" "NameProxy.handleDNS $upstream_dns_b <-"
}

restart_name_proxy_backends() {
  assert_cmd docker exec $sw1_name pkill -f /usr/bin/openceci
  assert_cmd docker exec $sw1_name openlan reload --save
  start_name_proxy_backends
  assert_match 30 "docker exec $sw1_name ping -c 3 192.55.0.2" "bytes from"
  assert_match 30 "docker exec $sw1_name ping -c 3 192.55.0.3" "bytes from"
  assert_match 20 "docker exec $sw1_name cat /var/openlan/ceci/$name_listen.log" "NameProxy.StartDNS on $name_listen"
}

setup_topology() {
  setup_net
  setup_sw1
  setup_sw2
  setup_sw3
  setup_upstream_dns
  setup_name_proxy
}

setup() {
  setup_topology
  test_name_proxy_backends
  restart_name_proxy_backends
  test_name_proxy_backends
}

case "$1" in
  --topology)
    show_topology
    ;;
  *)
    main
    ;;
esac
