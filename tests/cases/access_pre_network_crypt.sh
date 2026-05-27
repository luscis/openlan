#!/bin/bash
source tools/auto.sh

# OpenLAN Access scenario: mixed crypt modes on two networks.
#
# Validate:
# 1) network a uses pre-network crypt.
# 2) network b uses switch global crypt.

export net_name=tests-net-pre-crypt
export sw1_name=tests-sw-pre-crypt
export ac_network_name=tests-sw-pre-crypt.ac-network
export ac_network_wrong_name=tests-sw-pre-crypt.ac-network-wrong
export ac_global_name=tests-sw-pre-crypt.ac-global
export ac_default_name=tests-sw-pre-crypt.ac-default
export ac_default_wrong_name=tests-sw-pre-crypt.ac-default-wrong
export ac_b_global_name=tests-sw-pre-crypt.ac-b-global
export ac_b_network_name=tests-sw-pre-crypt.ac-b-network
export global_secret=global-secret-9a0b
export network_secret=network-secret-7c1d
export network_secret_v2=network-secret-8d2e
export ac_network_old_after_update_name=tests-sw-pre-crypt.ac-network-old-after-update
export ac_network_new_after_update_name=tests-sw-pre-crypt.ac-network-new-after-update
export ac_b_global_after_update_name=tests-sw-pre-crypt.ac-b-global-after-update

setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.251.0.0/24 --gateway=172.251.0.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.251.0.241

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  cat > /opt/openlan/$name/etc/openlan/switch/switch.json <<JSON
{
  "protocol": "tcp",
  "crypt": {
    "algorithm": "aes-128",
    "secret": "$global_secret"
  }
}
JSON

  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan network --name a add --address 192.61.0.1/24
  assert_cmd docker exec $name openlan network --name b add --address 192.62.0.1/24
  assert_cmd docker exec $name openlan network --name a crypt update --algorithm aes-128 --secret "$network_secret"
  assert_cmd docker exec $name openlan network --name a crypt list | grep "secret: $network_secret"

  assert_cmd docker exec $name openlan user add --name t1@a --password 123456
  assert_cmd docker exec $name openlan user add --name t2@b --password 123457
}

setup_access_network_level() {
  local name="$ac_network_name"
  mkdir -p /opt/openlan/$name/etc/openlan
  cat > /opt/openlan/$name/etc/openlan/access.yaml <<YAML
protocol: tcp
crypt:
  algorithm: aes-128
  secret: $network_secret
  level: network
connection: 172.251.0.241
username: t1@a
password: 123456
interface:
  address: 192.61.0.11/24
YAML
  start_access $name $net_name
  assert_expect 30 "docker logs -f $name" "Worker.OnSuccess"
}

setup_access_network_level_with_global_secret_should_fail() {
  local name="$ac_network_wrong_name"
  mkdir -p /opt/openlan/$name/etc/openlan
  cat > /opt/openlan/$name/etc/openlan/access.yaml <<YAML
protocol: tcp
crypt:
  algorithm: aes-128
  secret: $global_secret
  level: network
connection: 172.251.0.241
username: t1@a
password: 123456
interface:
  address: 192.61.0.12/24
YAML
  start_access $name $net_name
  assert_unexpect 15 "docker logs -f $name" "Worker.OnSuccess"
}

setup_access_global_level_with_network_secret_should_fail() {
  local name="$ac_global_name"
  mkdir -p /opt/openlan/$name/etc/openlan
  cat > /opt/openlan/$name/etc/openlan/access.yaml <<YAML
protocol: tcp
crypt:
  algorithm: aes-128
  secret: $network_secret
  level: global
connection: 172.251.0.241
username: t1@a
password: 123456
interface:
  address: 192.61.0.13/24
YAML
  start_access $name $net_name
  assert_unexpect 15 "docker logs -f $name" "Worker.OnSuccess"
}

setup_access_default_level_with_switch_secret() {
  local name="$ac_default_name"
  mkdir -p /opt/openlan/$name/etc/openlan
  cat > /opt/openlan/$name/etc/openlan/access.yaml <<YAML
protocol: tcp
crypt:
  algorithm: aes-128
  secret: $global_secret
connection: 172.251.0.241
username: t1@a
password: 123456
interface:
  address: 192.61.0.14/24
YAML
  start_access $name $net_name
  assert_expect 30 "docker logs -f $name" "Worker.OnSuccess"
}

setup_access_default_level_with_network_secret_should_fail() {
  local name="$ac_default_wrong_name"
  mkdir -p /opt/openlan/$name/etc/openlan
  cat > /opt/openlan/$name/etc/openlan/access.yaml <<YAML
protocol: tcp
crypt:
  algorithm: aes-128
  secret: $network_secret
connection: 172.251.0.241
username: t1@a
password: 123456
interface:
  address: 192.61.0.15/24
YAML
  start_access $name $net_name
  assert_unexpect 15 "docker logs -f $name" "Worker.OnSuccess"
}

setup_access_network_b_default_level_with_switch_secret() {
  local name="$ac_b_global_name"
  mkdir -p /opt/openlan/$name/etc/openlan
  cat > /opt/openlan/$name/etc/openlan/access.yaml <<YAML
protocol: tcp
crypt:
  algorithm: aes-128
  secret: $global_secret
connection: 172.251.0.241
username: t2@b
password: 123457
interface:
  address: 192.62.0.11/24
YAML
  start_access $name $net_name
  assert_expect 30 "docker logs -f $name" "Worker.OnSuccess"
}

setup_access_network_b_level_network_should_fail() {
  local name="$ac_b_network_name"
  mkdir -p /opt/openlan/$name/etc/openlan
  cat > /opt/openlan/$name/etc/openlan/access.yaml <<YAML
protocol: tcp
crypt:
  algorithm: aes-128
  secret: $network_secret
  level: network
connection: 172.251.0.241
username: t2@b
password: 123457
interface:
  address: 192.62.0.12/24
YAML
  start_access $name $net_name
  assert_unexpect 15 "docker logs -f $name" "Worker.OnSuccess"
}

test_update_network_a_crypt() {
  assert_cmd docker exec $sw1_name openlan network --name a crypt update --algorithm aes-128 --secret "$network_secret_v2"
  assert_cmd docker exec $sw1_name openlan network --name a crypt list | grep "secret: $network_secret_v2"

  mkdir -p /opt/openlan/$ac_network_old_after_update_name/etc/openlan
  cat > /opt/openlan/$ac_network_old_after_update_name/etc/openlan/access.yaml <<YAML
protocol: tcp
crypt:
  algorithm: aes-128
  secret: $network_secret
  level: network
connection: 172.251.0.241
username: t1@a
password: 123456
interface:
  address: 192.61.0.16/24
YAML
  start_access $ac_network_old_after_update_name $net_name
  assert_unexpect 15 "docker logs -f $ac_network_old_after_update_name" "Worker.OnSuccess"

  mkdir -p /opt/openlan/$ac_network_new_after_update_name/etc/openlan
  cat > /opt/openlan/$ac_network_new_after_update_name/etc/openlan/access.yaml <<YAML
protocol: tcp
crypt:
  algorithm: aes-128
  secret: $network_secret_v2
  level: network
connection: 172.251.0.241
username: t1@a
password: 123456
interface:
  address: 192.61.0.17/24
YAML
  start_access $ac_network_new_after_update_name $net_name
  assert_expect 30 "docker logs -f $ac_network_new_after_update_name" "Worker.OnSuccess"

  mkdir -p /opt/openlan/$ac_b_global_after_update_name/etc/openlan
  cat > /opt/openlan/$ac_b_global_after_update_name/etc/openlan/access.yaml <<YAML
protocol: tcp
crypt:
  algorithm: aes-128
  secret: $global_secret
connection: 172.251.0.241
username: t2@b
password: 123457
interface:
  address: 192.62.0.13/24
YAML
  start_access $ac_b_global_after_update_name $net_name
  assert_expect 30 "docker logs -f $ac_b_global_after_update_name" "Worker.OnSuccess"
}

setup_topology() {
  setup_net
  setup_sw1

  # Network a (pre-network crypt) matrix.
  setup_access_network_level
  setup_access_network_level_with_global_secret_should_fail
  setup_access_global_level_with_network_secret_should_fail
  setup_access_default_level_with_switch_secret
  setup_access_default_level_with_network_secret_should_fail

  # Network b (global crypt only) checks.
  setup_access_network_b_default_level_with_switch_secret
  setup_access_network_b_level_network_should_fail

  # Update network a crypt and verify new key takes effect.
  test_update_network_a_crypt
}

setup() {
  setup_topology
}

main
