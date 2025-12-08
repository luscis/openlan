#!/bin/bash

set -ex

cs_dir="/etc/openlan/switch"

sysctl -p /etc/sysctl.d/90-openlan.conf

## START: clean older files.
/usr/bin/env find /var/openlan/access -type f -delete
/usr/bin/env find /var/openlan/openvpn -name '*.status' -delete
/usr/bin/env find /var/openlan/openvpn -name '*client.ovpn' -delete
/usr/bin/env find /var/openlan/openvpn -name '*client.tmpl' -delete
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
  "provider": "ipsec"
}
EOF

[ -e $cs_dir/network/bgp.json ] || cat > $cs_dir/network/bgp.json << EOF
{
  "name": "bgp",
  "provider": "bgp"
}
EOF

[ -e $cs_dir/network/ceci.json ] || cat > $cs_dir/network/ceci.json << EOF
{
  "name": "ceci",
  "provider": "ceci"
}
EOF

[ -e $cs_dir/network/router.json ] || cat > $cs_dir/network/router.json << EOF
{
  "name": "router",
  "provider": "router"
}
EOF
## END

## START: wait ipsec ready
while true; do
  if ipsec status ; then
    break
  fi
  sleep 5
done
## END

exec /usr/bin/openlan-switch -conf:dir $cs_dir -log:level 20
