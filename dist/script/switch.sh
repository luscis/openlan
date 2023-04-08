#!/bin/bash

set -ex

## set this bridge as a root
# ip link show  br-hello || brctl addbr br-hello
# brctl setbridgeprio br-hello 0

# clean older files.
/usr/bin/env find /var/openlan/point -type f -delete
/usr/bin/env find /var/openlan/openvpn -name '*.status' -delete

if [ ! -e "/etc/openlan/switch/switch.json" ]; then
cat >> /etc/openlan/switch/switch.json << EOF
{
    "cert": {
        "directory": "/var/openlan/cert"
    },
    "http": {
        "public": "/var/openlan/public"
    },
    "crypt": {
        "secret": "cb2ff088a34d"
    }
}
EOF
fi

exec /usr/bin/openlan-switch -conf:dir /etc/openlan/switch -log:level 20
