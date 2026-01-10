#!/bin/bash

set -ex

cs_dir="/etc/openlan/switch"

function prepare() {
  sysctl -p /etc/sysctl.d/90-openlan.conf

  ## START: clean older files.
  /usr/bin/env find /var/openlan/access -type f -delete
  /usr/bin/env find /var/openlan/openvpn -type f -mindepth 2 -maxdepth 2 -delete
  ## END

  ## START: prepare external dir.
  for dir in network acl findhop output route qos dnat; do
    [ -e "$cs_dir/$dir" ] || mkdir -p "$cs_dir/$dir"
  done
  ## END

  [ -e $cs_dir/switch.json ] || cat > $cs_dir/switch.json << EOF
{
  "crypt": {
    "secret": "cb2ff088a34d"
  }
}
EOF

  ## START: install default network
  [ -e $cs_dir/network/ipsec.json ] || cat > $cs_dir/network/ipsec.json << EOF
{
  "name": "ipsec",
  "provider": "ipsec",
  "snat": "disable"
}
EOF

  [ -e $cs_dir/network/bgp.json ] || cat > $cs_dir/network/bgp.json << EOF
{
  "name": "bgp",
  "provider": "bgp",
  "snat": "disable"
}
EOF

  [ -e $cs_dir/network/ceci.json ] || cat > $cs_dir/network/ceci.json << EOF
{
  "name": "ceci",
  "provider": "ceci",
  "snat": "disable"
}
EOF

  [ -e $cs_dir/network/router.json ] || cat > $cs_dir/network/router.json << EOF
{
  "name": "router",
  "provider": "router",
  "snat": "disable"
}
EOF
  ## END
}

function wait_ipsec() {
  ## START: wait ipsec ready
  while true; do
    if ipsec status ; then
      break
    fi
    sleep 5
  done
  ## END
}

function start_switch {
  exec /usr/bin/openlan-switch -conf:dir $cs_dir -log:level 20 & child=$!
  trap 'kill ${child:-}; wait ${child:-}' SIGINT SIGTERM
  wait $child
}

prepare
wait_ipsec
start_switch

# Wait reloading.
while [ true ]; do
  start_switch
done