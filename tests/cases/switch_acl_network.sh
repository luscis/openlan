#!/bin/bash
source tools/auto.sh

show_description() {
  echo "verify acl ebtables hook is bridge ingress only"
}

show_topology_summary() {
  cat <<'EOF'
sw1 192.63.0.1 | +-- ACL hook checks on br-example
EOF
}

show_topology() {
  cat <<'EOF'
# Topology:
# - Diagram:
#       sw1 192.63.0.1
#              |
#              +-- ACL hook checks on br-example
# - Docker mgmt network: 100.100.2.0/24
#   sw1=100.100.2.241.
# - OpenLAN service network "example": 192.63.0.0/24
# Validation:
#   ACL ebtables hook is installed only on bridge ingress FORWARD traffic.

EOF
}

# OpenLAN Switch UT: ACL ebtables ingress-only hook path.

export net_name=tests-net-acl-ingress
export sw1_name=tests-sw-acl-ingress1

setup_net() {
  docker network create $net_name --driver=bridge --subnet=100.100.2.0/24 --gateway=100.100.2.1 >/dev/null
}

setup_sw1() {
  local name="$sw1_name"
  local address=100.100.2.241
  local crypt_secret="cb2ff088a34d"

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  assert_expect 30 "docker logs -f $name" "Http.Start"

  assert_cmd docker exec $name openlan crypt update --algorithm aes-128 --secret "$crypt_secret"
  assert_cmd docker exec $name openlan network --name example add --address 192.63.0.1/24
  assert_cmd docker exec $name openlan acl --name example rule add --srcip 192.63.0.2 --dstip 192.63.0.1 --protocol icmp
}

test_acl_ingress_hook() {
  assert_match 10 "docker exec $sw1_name ebtables -t filter -L FORWARD" "logical-in br-example.*AT_example"
  assert_unmatch 3 "docker exec $sw1_name ebtables -t filter -L FORWARD" "logical-out br-example.*AT_example"
  assert_match 10 "docker exec $sw1_name ebtables -t filter -L INPUT" "logical-in br-example.*AT_example"
  assert_match 10 "docker exec $sw1_name iptables -t raw -S TT_pre-example" "hi-example.*AT_example"
  assert_unmatch 3 "docker exec $sw1_name iptables -t raw -S TT_pre-example" "br-example.*AT_example"
  assert_match 10 "docker exec $sw1_name ebtables -t filter -L AT_example" "192.63.0.2.*192.63.0.1.*icmp.*DROP"
}

setup() {
  setup_net
  setup_sw1
  test_acl_ingress_hook
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
