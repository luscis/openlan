#!/bin/bash

set -ex

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

/usr/bin/openlan-switch -conf:dir /etc/openlan/switch -log:level 20
