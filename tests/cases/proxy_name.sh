#!/bin/bash
source tools/auto.sh

show_topology() {
  cat <<'EOF'
# Topology:
# - Docker mgmt network: 172.249.0.0/24
#   sw1=172.249.0.241 (name proxy client), sw2=172.249.0.242 (upstream dns server).
# - OpenLAN service network "example": 192.54.0.0/24
#   sw1=192.54.0.1, sw2=192.54.0.2, with sw2 output to sw1.
# Validation:
#   sw1 nslookup -> sw1 openceci(name) -> sw2 dnsmasq(upstream).

EOF
}

# OpenLAN Proxy UT: Ceci NAME proxy path.

export net_name=tests-net-proxy-name
export sw1_name=tests-sw-proxy-name1
export sw2_name=tests-sw-proxy-name2
export name_listen=127.0.0.1:1053
export name_domain=proxy-name.test
export name_answer=192.54.0.2
export upstream_dns=192.54.0.2:5300


setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.249.0.0/24 --gateway=172.249.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.249.0.241

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

  assert_cmd docker exec $name openlan network --name example add --address 192.54.0.1/24
  assert_cmd docker exec $name openlan user add --name t1@example --password 123456
}

setup_sw2() {
  local name="$sw2_name"
  local address=172.249.0.242

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

  assert_cmd docker exec $name openlan network --name example add --address 192.54.0.2/24
  assert_cmd docker exec $name openlan network --name example output add --remote 172.249.0.241 --protocol tcp --secret t1@example:123456 --crypt aes-128:ea64d5b0c96c
  assert_match 20 "docker exec $name openlan network --name example output ls" "state: authenticated"
}

setup_upstream_dns() {
  assert_cmd docker exec $sw2_name sh -c "nohup dnsmasq --no-daemon --port=5300 --listen-address=192.54.0.2 --bind-interfaces --address=/$name_domain/$name_answer >/tmp/proxy-name-dnsmasq.log 2>&1 &"
  assert_match 20 "docker exec $sw1_name ping -c 3 192.54.0.2" "bytes from"
}

setup_name_proxy() {
  assert_cmd docker exec $sw1_name sh -c "cat > /var/openlan/ceci/$name_listen.yaml <<EOF
listen: $name_listen
nameto: $upstream_dns
EOF"
  start_name_proxy
}

start_name_proxy() {
  assert_cmd docker exec $sw1_name sh -c "nohup /usr/bin/openceci -mode name -conf /var/openlan/ceci/$name_listen.yaml -log:file /var/openlan/ceci/$name_listen.log -write-pid /var/openlan/ceci/$name_listen.pid >/tmp/proxy-name.log 2>&1 &"
  assert_match 20 "docker exec $sw1_name cat /var/openlan/ceci/$name_listen.log" "NameProxy.StartDNS on $name_listen"
}

test_name_proxy() {
  assert_fuzzy 20 "docker exec $sw1_name nslookup -port=1053 $name_domain 127.0.0.1" "Address: $name_answer"
  assert_fuzzy 20 "docker exec $sw1_name cat /var/openlan/ceci/$name_listen.log" "NameProxy.handleDNS $upstream_dns <-"
}

restart_name_proxy() {
  assert_cmd docker exec $sw1_name pkill -f /usr/bin/openceci
  assert_cmd docker exec $sw1_name openlan reload --save
  start_name_proxy
  assert_match 30 "docker exec $sw1_name ping -c 3 192.54.0.2" "bytes from"
  assert_match 20 "docker exec $sw1_name cat /var/openlan/ceci/$name_listen.log" "NameProxy.StartDNS on $name_listen"
}

setup_topology() {
  setup_net
  setup_sw1
  setup_sw2
  setup_upstream_dns
  setup_name_proxy
}

setup() {
  setup_topology
  test_name_proxy
  restart_name_proxy
  test_name_proxy
}

case "$1" in
  --topology)
    show_topology
    ;;
  *)
    main
    ;;
esac
