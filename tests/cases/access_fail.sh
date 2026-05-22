
# OpenLAN Access UT: authentication failure path.

net_name=tests-net-authfail
sw1_name=tests-sw-authfail
ac1_badpass_name=tests-sw-authfail.acbad

# Topology:
# - Docker mgmt network: 172.253.0.0/24
#   sw1=172.253.0.241, bad access client joins the same mgmt network.
# - OpenLAN service network "example": 192.31.0.0/24
#   sw1 gateway=192.31.0.1, client config asks for 192.31.0.11.
# - Validation path: authentication must fail with wrong password.

setup_net() {
  docker network create $net_name --driver=bridge --subnet=172.253.0.0/24 --gateway=172.253.0.1
}

setup_sw1() {
  local name="$sw1_name"
  local address=172.253.0.241

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
  wait "docker logs -f $name" Http.Start 30

  docker exec $name openlan network --name example add --address 192.31.0.1/24
  docker exec $name openlan user add --name t1@example --password 123456
}

setup_ac_badpass() {
  local name="$ac1_badpass_name"

  mkdir -p /opt/openlan/$name/etc/openlan
  cat > /opt/openlan/$name/etc/openlan/access.yaml <<EOF
protocol: tcp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
connection: 172.253.0.241
username: t1@example
password: wrong-password
interface:
  address: 192.31.0.11/24
EOF

  start_access $name $net_name
  if wait "docker logs -f $name" "Worker.OnSuccess" 15; then
    echo "unexpected success with wrong password"
    return 1
  fi
}

setup() {
  setup_net
  setup_sw1
  setup_ac_badpass
}

main
