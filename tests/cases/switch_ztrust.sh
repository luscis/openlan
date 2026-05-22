# OpenLAN Switch UT: ztrust enable/guest/knock flow.

net_name=tests-net-ztrust
sw1_name=tests-sw-ztrust
vpn1_name=tests-sw-ztrust.vpn1

# Topology:
# - Docker mgmt network: 172.245.0.0/24
#   sw1=172.245.0.241.
# - OpenLAN service network "example": 192.59.0.0/24
#   sw1=192.59.0.1.
# - OpenVPN overlay on sw1:
#   tcp/1194, subnet 10.93.0.0/24, vpn1@example fixed address 10.93.0.10.
# - Validation path:
#   vpn1 -> sw1:8081 is reachable before ztrust; blocked after ztrust enable;
#   allowed after guest+knock; blocked after knock remove; restored when disabled.

setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.245.0.0/24 --gateway=172.245.0.1
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.245.0.241

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  start_switch $name $net_name $address
  wait "docker logs -f $name" Http.Start 30

  docker exec $name openlan network --name example add --address 192.59.0.1/24
  docker exec $name openlan user add --name vpn1@example --password 123456

  docker exec $name openlan network --name example openvpn add --listen :1194 --protocol tcp --subnet 10.93.0.0/24 --dns 8.8.8.8
  docker exec $name openlan network --name example client add --user vpn1 --address 10.93.0.10
}

setup_openvpn_client() {
  local name="$sw1_name"

  mkdir -p /opt/openlan/$vpn1_name/ovpn
  docker cp $name:/var/openlan/openvpn/example/tcp1194client.ovpn /opt/openlan/$vpn1_name/ovpn/client.ovpn
  cat > /opt/openlan/$vpn1_name/ovpn/auth.txt <<EOF
vpn1@example
123456
EOF

  start_openvpn $vpn1_name $net_name
  wait "docker logs -f $vpn1_name" "Initialization Sequence Completed" 40
}

setup_local_service() {
  docker exec $sw1_name sh -c "nohup sh -c 'while true; do printf \"HTTP/1.1 200 OK\\r\\nContent-Length: 11\\r\\n\\r\\nztrust-8081\" | socat - TCP-LISTEN:8081,reuseaddr; done' >/tmp/ztrust-8081.log 2>&1 &"
}

test_ztrust_flow() {
  check "docker exec $vpn1_name wget -qO- -T 3 -t 1 http://192.59.0.1:8081" "ztrust-8081" 3

  docker exec $sw1_name openlan ztrust --network example enable
  check "docker exec $sw1_name iptables -t mangle -S TT_pre-example" "Goto Zero Trust" 1
  check "docker exec $sw1_name iptables -t mangle -S ZT_example" "ZTrust Deny All" 1
  if check "docker exec $vpn1_name wget -qO- -T 3 -t 1 http://192.59.0.1:8081" "ztrust-8081" 3; then
    echo "unexpected access success after ztrust enable without guest/knock"
    return 1
  fi

  docker exec $sw1_name openlan reload --save
  
  check "docker exec $sw1_name iptables -t mangle -S TT_pre-example" "Goto Zero Trust" 1
  check "docker exec $sw1_name iptables -t mangle -S ZT_example" "ZTrust Deny All" 1
  if check "docker exec $vpn1_name wget -qO- -T 3 -t 1 http://192.59.0.1:8081" "ztrust-8081" 3; then
    echo "unexpected access success after ztrust reload without guest/knock"
    return 1
  fi

  docker exec $sw1_name openlan ztrust --network example guest  add --user vpn1 --address 10.93.0.10
  docker exec $sw1_name openlan ztrust --network example knock add --user vpn1 --protocol tcp --socket 192.59.0.1:8081 --age 120
  
  check "docker exec $sw1_name openlan ztrust --network example guest ls" "vpn1@example" 1
  check "docker exec $sw1_name openlan ztrust --network example knock ls --user vpn1" "192.59.0.1:8081" 1
  check "docker exec $vpn1_name wget -qO- -T 3 -t 1 http://192.59.0.1:8081" "ztrust-8081" 3

  docker exec $sw1_name openlan ztrust --network example guest rm --user vpn1
  docker exec $sw1_name openlan ztrust --network example disable
  check "docker exec $vpn1_name wget -qO- -T 3 -t 1 http://192.59.0.1:8081" "ztrust-8081" 3 || {
    echo "unexpected access failure after ztrust disable"
    return 1
  }

  docker exec $sw1_name openlan reload --save
  check "docker exec $vpn1_name wget -qO- -T 3 -t 1 http://192.59.0.1:8081" "ztrust-8081" 3 || {
    echo "unexpected access failure after ztrust disable"
    return 1
  }
}

setup() {
  setup_net
  setup_sw1
  setup_openvpn_client
  setup_local_service
  test_ztrust_flow
}

main
