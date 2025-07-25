# OpenLAN Access UT.

source $PWD/macro.sh

setup_net() {
  docker network inspect net1 || {
    docker network create net1 \
      --driver=bridge --subnet=172.255.0.0/24 --gateway=172.255.0.1
  }
}

setup_sw1() {
  mkdir -p /opt/openlan/sw1
  mkdir -p /opt/openlan/sw1/etc/openlan/switch
  cat > /opt/openlan/sw1/etc/openlan/switch/switch.yaml <<EOF
protocol: tcp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
EOF

  # Start switch: sw1
  docker run -d --rm --privileged --network net1 --ip 172.255.0.2 \
    --volume /opt/openlan/sw1/etc/openlan:/etc/openlan \
    --name sw1 $IMAGE /usr/bin/openlan-switch -conf:dir /etc/openlan/switch

  wait "docker logs -f sw1" Http.Start 30

  # Add a network.
  docker exec sw1 openlan network add --name example --address 172.11.0.1/24
  # Add users
  docker exec sw1 openlan user add --name t1@example --password 123456
  docker exec sw1 openlan user add --name t2@example --password 123457
}

setup_ac1() {
  mkdir -p /opt/openlan/sw1/etc/openlan/access
  # Start access: ac1
  cat > /opt/openlan/sw1/etc/openlan/access/t1.yaml <<EOF
protocol: tcp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
connection: 172.255.0.2
username: t1@example
password: 123456
interface:
  address: 172.11.0.11/24
EOF
  docker run -d --rm --privileged --network net1 \
    --volume /opt/openlan/sw1/etc/openlan:/etc/openlan \
    --name sw1.ac1 $IMAGE /usr/bin/openlan-access -conf /etc/openlan/access/t1.yaml

  wait "docker logs -f sw1.ac1" Worker.OnSuccess 30
}

setup_ac2() {
  # Start access: ac2
  cat > /opt/openlan/sw1/etc/openlan/access/t2.yaml <<EOF
protocol: tcp
crypt:
  algorithm: aes-128
  secret: ea64d5b0c96c
connection: 172.255.0.2
username: t2@example
password: 123457
interface:
  address: 172.11.0.12/24
EOF
  docker run -d --rm --privileged --network net1 \
    --volume /opt/openlan/sw1/etc/openlan:/etc/openlan \
    --name sw1.ac2 $IMAGE /usr/bin/openlan-access -conf /etc/openlan/access/t2.yaml

  wait "docker logs -f sw1.ac2" Worker.OnSuccess
}

ping() {
  wait "docker exec sw1.ac1 ping -c 3 172.11.0.1" "3 received" 5
  wait "docker exec sw1.ac1 ping -c 3 172.11.0.12" "3 received" 5
}

setup() {
  setup_net
  setup_sw1
  setup_ac1
  setup_ac2
  ping
}

cleanup() {
  # Stop containd
  docker stop sw1.ac1
  docker stop sw1.ac2
  docker stop sw1

  # Cleanup files
  rm -rvf /opt/openlan/sw1
}

main