# OpenLAN Access UT.

source $PWD/macro.sh

net_name=tests-net1
sw1_name=tests-sw1
ac1_name=tests-sw1.ac1
ac2_name=tests-sw1.ac2
crypt_secret_v1=ea64d5b0c96c
crypt_secret_v2=ea64d5b0c96d

# Topology:
# - Docker mgmt network: 172.255.0.0/24
#   sw1=172.255.0.241, ac1/ac2 join the same mgmt network.
# - OpenLAN service network "example": 192.11.0.0/24
#   sw1 gateway=192.11.0.1, ac1=192.11.0.11, ac2=192.11.0.12.
# - Validation path: ac1 -> sw1 and ac2 -> sw1 connectivity by ping.

describe() {
  echo "==> scenario: access_success"
  echo "    description: access success path: two access clients authenticate and can communicate"
}

setup_net() {
  docker network inspect $net_name || {
    docker network create $net_name \
      --driver=bridge --subnet=172.255.0.0/24 --gateway=172.255.0.1
  }
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.255.0.241

  mkdir -p /opt/openlan/$name/etc/openlan/switch
  cat > /opt/openlan/$name/etc/openlan/switch/switch.json <<EOF
{
  "protocol": "tcp",
  "crypt": {
    "algorithm": "aes-128",
    "secret": "$crypt_secret_v1"
  }
}
EOF

  # Start switch: tests-sw1
  start_switch $name $net_name $address
  wait "docker logs -f $name" Http.Start 30

  # Add a network.
  docker exec $name openlan network --name example add --address 192.11.0.1/24
  # Add users
  docker exec $name openlan user add --name t1@example --password 123456
  docker exec $name openlan user add --name t2@example --password 123457
}

setup_ac1() {
  local name="$ac1_name"
  local secret="${1:-$crypt_secret_v1}"

  mkdir -p /opt/openlan/$name/etc/openlan
  # Start access: ac1
  cat > /opt/openlan/$name/etc/openlan/access.yaml <<EOF
protocol: tcp
crypt:
  algorithm: aes-128
  secret: $secret
connection: 172.255.0.241
username: t1@example
password: 123456
interface:
  address: 192.11.0.11/24
EOF
  start_access $name $net_name
}

setup_ac2() {
  local name="$ac2_name"
  local secret="${1:-$crypt_secret_v1}"

  mkdir -p /opt/openlan/$name/etc/openlan
  # Start access: ac2
  cat > /opt/openlan/$name/etc/openlan/access.yaml <<EOF
protocol: udp
crypt:
  algorithm: aes-128
  secret: $secret
connection: 172.255.0.241
username: t2@example
password: 123457
interface:
  address: 192.11.0.12/24
EOF
  start_access $name $net_name
}

test_ping() {
  wait "docker logs -f $ac1_name" Worker.OnSuccess 30
  wait "docker logs -f $ac2_name" Worker.OnSuccess 30

  wait "docker exec $ac1_name ping -c 3 192.11.0.1" "3 received" 5
  wait "docker exec $ac2_name ping -c 3 192.11.0.12" "3 received" 5
}

test_crypt_update() {
  docker exec $sw1_name openlan crypt update --algorithm aes-128 --secret "$crypt_secret_v2"
  docker exec $sw1_name openlan crypt ls | grep "secret: $crypt_secret_v2"

  docker stop $ac1_name
  docker stop $ac2_name

  setup_ac1 "$crypt_secret_v1"
  wait "docker logs -f $ac1_name" SocketClientImpl.Reset 30

  docker stop $ac1_name
  setup_ac1 "$crypt_secret_v2"
  wait "docker logs -f $ac1_name" Worker.OnSuccess 30

  setup_ac2 "$crypt_secret_v2"
  wait "docker logs -f $ac2_name" Worker.OnSuccess 30
  test_ping
}

setup() {
  describe
  setup_net
  setup_sw1
  setup_ac1
  setup_ac2
  test_ping
  test_crypt_update
}

main
