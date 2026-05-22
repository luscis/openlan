# OpenLAN Switch UT: IPSec output path.

net_name=tests-net-ipsec-output
sw1_name=tests-sw-ipsec1
sw2_name=tests-sw-ipsec2
ipsec_secret=ea64d5b0c96c

# Topology:
# - Docker mgmt network: 172.248.0.0/24
#   sw1=172.248.0.241, sw2=172.248.0.242.
# - OpenLAN service network "example": 192.56.0.0/24
#   sw1=192.56.0.1, sw2=192.56.0.2.
# - IPSec tunnel:
#   sw1 <-> sw2 over mgmt addresses with shared PSK.
# - Output link:
#   sw2 -> sw1 by vxlan output.
# - Validation path:
#   sw2 can ping sw1 service address through ipsec-protected output path.

setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.248.0.0/24 --gateway=172.248.0.1
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
  wait "docker logs -f $name" Http.Start 30

  docker exec $name openlan network --name example add --address 192.56.0.1/24
  docker exec $name openlan user add --name edge@example --password 123456
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
  wait "docker logs -f $name" Http.Start 30

  docker exec $name openlan network --name example add --address 192.56.0.2/24
  docker exec $name openlan user add --name edge@example --password 123456
}

setup_output() {
  docker exec $sw1_name openlan network --name example output add --remote 172.248.0.242 --protocol vxlan --segment 1056
  docker exec $sw2_name openlan network --name example output add --remote 172.248.0.241 --protocol vxlan --segment 1056
}

test_ipsec_output_ping() {
  docker exec $sw1_name openlan ipsec tunnel add --remote 172.248.0.242 --protocol vxlan --secret $ipsec_secret --localid sw1.ipsec.test --remoteid sw2.ipsec.test
  docker exec $sw2_name openlan ipsec tunnel add --remote 172.248.0.241 --protocol vxlan --secret $ipsec_secret --localid sw2.ipsec.test --remoteid sw1.ipsec.test
  check "docker exec $sw1_name openlan ipsec tunnel ls | grep 172.248.0.242" "erouted" 20
  check "docker exec $sw2_name openlan ipsec tunnel ls | grep 172.248.0.241" "erouted" 20
  wait "docker exec $sw2_name ping -c 15 192.56.0.1" "bytes from" 20

  docker exec $sw1_name openlan reload --save
  docker exec $sw2_name openlan reload --save

  check "docker exec $sw1_name openlan ipsec tunnel ls | grep 172.248.0.242" "erouted" 20
  check "docker exec $sw2_name openlan ipsec tunnel ls | grep 172.248.0.241" "erouted" 20
  wait "docker exec $sw2_name ping -c 15 192.56.0.1" "bytes from" 20
}

setup() {
  setup_net
  setup_sw1
  setup_sw2
  setup_output
  test_ipsec_output_ping
}

main
