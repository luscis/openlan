#!/bin/bash

set -ex

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

if [ ! -e "/etc/openlan/switch/network/example.json" ]; then
cat >> /etc/openlan/switch/network/example.json << EOF
{
    "name": "example",
    "bridge": {
        "address": "172.32.100.40/24"
    }
}
EOF
fi

exec /usr/bin/openlan-switch -conf:dir /etc/openlan/switch -log:level 20
