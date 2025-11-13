#!/bin/bash

set -ex

## set this bridge as a root
# ip link show  br-hello || brctl addbr br-hello
# brctl setbridgeprio br-hello 0

sysctl -p /etc/sysctl.d/90-openlan.conf

# clean older files.
/usr/bin/env find /var/openlan/access -type f -delete
/usr/bin/env find /var/openlan/openvpn -name '*.status' -delete
/usr/bin/env find /var/openlan/openvpn -name '*client.ovpn' -delete
/usr/bin/env find /var/openlan/openvpn -name '*client.tmpl' -delete

cs_dir="/etc/openlan/switch"

if [ ! -e $cs_dir/switch.json ]; then
  cat > $cs_dir/switch.json << EOF
{
  "crypt": {
    "secret": "cb2ff088a34d"
  }
}
EOF
fi

if [ ! -e $cs_dir/network/ipsec.json ]; then
  cat > $cs_dir/network/ipsec.json << EOF
{
  "name": "ipsec",
  "provider": "ipsec"
}
EOF
fi

if [ ! -e $cs_dir/network/bgp.json ]; then
  cat > $cs_dir/network/bgp.json << EOF
{
  "name": "bgp",
  "provider": "bgp"
}
EOF
fi

if [ ! -e $cs_dir/network/proxy.json ]; then
  cat > $cs_dir/network/proxy.json << EOF
{
  "name": "proxy",
  "provider": "proxy"
}
EOF
fi

for dir in acl findhop link output route network qos dnat; do
  if [ -e "$cs_dir/$dir" ]; then
    continue
  fi
  mkdir -p "$cs_dir/$dir"
done

# wait ipsec service
while true; do
  if ipsec status ; then
      break
  fi
  sleep 5
done

exec /usr/bin/openlan-switch -conf:dir $cs_dir -log:level 20
