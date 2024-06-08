#!/bin/bash

set -ex

## set this bridge as a root
# ip link show  br-hello || brctl addbr br-hello
# brctl setbridgeprio br-hello 0

sysctl -p /etc/sysctl.d/90-openlan.conf

# clean older files.
/usr/bin/env find /var/openlan/point -type f -delete
/usr/bin/env find /var/openlan/openvpn -name '*.status' -delete
/usr/bin/env find /var/openlan/openvpn -name '*client.ovpn' -delete
/usr/bin/env find /var/openlan/openvpn -name '*client.tmpl' -delete

if [ ! -e "/etc/openlan/switch/switch.json" ]; then
cat >> /etc/openlan/switch/switch.json << EOF
{
    "crypt": {
        "secret": "cb2ff088a34d"
    }
}
EOF
fi

# wait ipsec service
while true; do
  if ipsec status ; then
      break
  fi
  sleep 5
done

exec /usr/bin/openlan-switch -conf:dir /etc/openlan/switch -log:level 20
