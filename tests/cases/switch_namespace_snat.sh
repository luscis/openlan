#!/bin/bash
source tools/auto.sh

show_topology() {
  cat <<'EOF'
# Topology:
# - Docker mgmt network: 172.241.0.0/24
#   sw1=172.241.0.241, sw2=172.241.0.242.
# - OpenLAN service network "example": 192.64.0.0/24
#   sw1=192.64.0.1, sw2=192.64.0.2.
# - sw2 service network L3 device is enslaved to VRF "vrf-snat"; sw1 is not.
# - Non-namespace network "b": 192.66.0.0/24
#   sw2=192.66.0.1.
# - Access clients:
#   ac1=192.64.0.11, connected to sw2 and forwarding 10.242.2.11/32.
#   acb=192.66.0.11, connected to sw2 network b and forwarding 10.242.2.11/32.
# - sw1 VIP:
#   lo=10.242.2.11/32, HTTP service listens on 10.242.2.11:8081.
# - Forwarding link:
#   sw2 -> sw1 over UDP output.
# Validation:
#   when SNAT is disabled, the VIP HTTP service sees ac1 address 192.64.0.11.
#   when SNAT is enabled, the VIP HTTP service sees sw2 overlay address 192.64.0.2.
#   network b is not in the namespace, so acb cannot access the VIP HTTP service
#   even when example SNAT is enabled.

EOF
}


# OpenLAN Switch UT: network namespace/VRF SNAT path.

export net_name=tests-net-namespace-snat
export sw1_name=tests-sw-namespace-snat1
export sw2_name=tests-sw-namespace-snat2
export ac1_name=tests-sw-namespace-snat.ac1
export acb_name=tests-sw-namespace-snat.acb
export vrf_name=vrf-snat
export target_vip=10.242.2.11
export access_ip=192.64.0.11
export access_b_ip=192.66.0.11


setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.241.0.0/24 --gateway=172.241.0.1 >/dev/null
}

setup_switch_config() {
  local name=$1

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  cat > /opt/openlan/$name/etc/openlan/switch/switch.json <<EOF
{
  "protocol": "udp",
  "crypt": {
    "algorithm": "aes-128",
    "secret": "ea64d5b0c96c"
  }
}
EOF
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.241.0.241

  setup_switch_config $name
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name example add --address 192.64.0.1/24
  assert_cmd docker exec $name openlan router address add --device lo --address $target_vip/32
  assert_cmd docker exec $name openlan user add --name t1@example --password 123456
}

setup_sw2() {
  local name="$sw2_name"
  local address=172.241.0.242

  setup_switch_config $name
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name example add --address 192.64.0.2/24 --namespace $vrf_name
  assert_match 1 "docker exec $name openlan network --name example" "namespace: $vrf_name"
  assert_cmd docker exec $name ip link show $vrf_name
  assert_match 5 "docker exec $name ip link show hi-example" "master $vrf_name"
  assert_cmd docker exec $name openlan network --name b add --address 192.66.0.1/24
  assert_cmd docker exec $name openlan network --name example snat disable
  assert_cmd docker exec $name openlan network --name example route add --prefix $target_vip/32 --nexthop 192.64.0.1
  assert_cmd docker exec $name openlan user add --name ac1@example --password 123456
  assert_cmd docker exec $name openlan user add --name acb@b --password 123456
  assert_cmd docker exec $name openlan network --name example output add --remote 172.241.0.241 --protocol udp --secret t1:123456 --crypt aes-128:ea64d5b0c96c
}

setup_ac1() {
  local name="$ac1_name"

  mkdir -p /opt/openlan/$name/etc/openlan
  cat > /opt/openlan/$name/etc/openlan/access.yaml <<EOF
protocol: udp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
connection: 172.241.0.242
username: ac1@example
password: 123456
interface:
  address: $access_ip/24
forward:
- $target_vip/32 to 192.64.0.2
EOF
  start_access $name $net_name
  assert_expect 30 "docker logs -f $name" "onLogin: success"
}

setup_acb() {
  local name="$acb_name"

  mkdir -p /opt/openlan/$name/etc/openlan
  cat > /opt/openlan/$name/etc/openlan/access.yaml <<EOF
protocol: udp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
connection: 172.241.0.242
username: acb@b
password: 123456
interface:
  address: $access_b_ip/24
forward:
- $target_vip/32 to 192.66.0.1
EOF
  start_access $name $net_name
  assert_expect 30 "docker logs -f $name" "onLogin: success"
}

setup_vip_http() {
  assert_cmd docker exec $sw1_name sh -c "cat > /tmp/namespace-snat-http.sh <<'EOF'
#!/bin/sh
printf 'HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nConnection: close\r\n\r\nsrc=%s\n' \"\$SOCAT_PEERADDR\"
EOF
chmod +x /tmp/namespace-snat-http.sh
nohup socat TCP-LISTEN:8081,bind=$target_vip,reuseaddr,fork EXEC:/tmp/namespace-snat-http.sh >/tmp/namespace-snat-http.log 2>&1 &"
}

assert_http_source() {
  local name=$1
  local source=$2
  assert_match 20 "docker exec $name wget -qO- -T 3 -t 1 http://$target_vip:8081" "src=$source"
}

assert_http_unreachable() {
  local name=$1
  assert_unmatch 3 "docker exec $name wget -qO- -T 3 -t 1 http://$target_vip:8081" "src="
}

assert_ping_target() {
  local name=$1
  assert_match 20 "docker exec $name ping -c 3 $target_vip" "bytes from"
}

assert_ping_target_fail() {
  local name=$1
  assert_unmatch 3 "docker exec $name ping -c 3 $target_vip" "bytes from"
}

assert_ac1_http_source() {
  local source=$1
  assert_http_source $ac1_name $source
}

test_namespace_snat() {
  assert_match 15 "docker exec $sw2_name openlan network --name example output ls" "state: authenticated"

  assert_cmd docker exec $sw2_name openlan network --name example snat disable
  assert_ping_target $ac1_name
  assert_ac1_http_source $access_ip

  assert_cmd docker exec $sw2_name openlan network --name example snat enable --scope enable
  assert_ping_target $ac1_name
  assert_ac1_http_source 192.64.0.2
  assert_ping_target_fail $acb_name
  assert_http_unreachable $acb_name
}

test_reload_persistence() {
  assert_cmd docker exec $sw1_name openlan reload --save
  assert_cmd docker exec $sw2_name openlan reload --save

  assert_match 10 "docker exec $sw2_name openlan network --name example" "namespace: $vrf_name"
  assert_match 10 "docker exec $sw2_name ip link show hi-example" "master $vrf_name"
  assert_ping_target $ac1_name
  assert_ac1_http_source 192.64.0.2
  assert_ping_target_fail $acb_name
  assert_http_unreachable $acb_name
}

setup_topology() {
  setup_net
  setup_sw1
  setup_sw2
  setup_ac1
  setup_acb
  setup_vip_http
}

setup() {
  setup_topology
  test_namespace_snat
  test_reload_persistence
}

case "$1" in
  --topology)
    show_topology
    ;;
  *)
    main
    ;;
esac
