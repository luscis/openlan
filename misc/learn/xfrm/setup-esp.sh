#!/usr/bin/env bash

set -ex

# Topo.
#
#           100.141        --        200.130
#            |                        |
#      192.168.209.141   <=====>   192.168.209.130
#

auth_key=$(dd if=/dev/urandom count=32 bs=1 2> /dev/null| xxd -p -c 64)
enc_key=$(dd if=/dev/urandom count=32 bs=1 2> /dev/null| xxd -p -c 64)

sun_spi=$(dd if=/dev/urandom count=4 bs=1 2> /dev/null| xxd -p -c 8)
moon_spi=$(dd if=/dev/urandom count=4 bs=1 2> /dev/null| xxd -p -c 8)

reqid=$(dd if=/dev/urandom count=4 bs=1 2> /dev/null| xxd -p -c 8)

sun="$1"; shift
sun_net="$1"; shift
moon="$1"; shift
moon_net="$1"; shift

if [ -z "${sun}${sun_net}${moon}${moon_net}" ]; then
    echo "$0 moon moon-net sun sun-net"
    exit 0
fi

sun_port="$1";
moon_port="$2";

if [ -z "${sun_port}" ]; then
     sun_port="22"
fi
if [ -z "${moon_port}" ]; then
     moon_port="22"
fi

ssh -p ${sun_port} ${sun} /bin/bash << EOF
    # --
    ip xfrm state flush

    ip xfrm state add src ${moon} dst ${sun} proto esp spi 0x${moon_spi} reqid 0x${reqid} mode tunnel auth sha256 0x${auth_key} enc aes 0x${enc_key}
    ip xfrm state add src ${sun} dst ${moon} proto esp spi 0x${sun_spi} reqid 0x${reqid} mode tunnel auth sha256 0x${auth_key} enc aes 0x${enc_key}
    ip xfrm state ls

    # --
    ip xfrm policy flush

    ip xfrm policy add src ${moon_net} dst ${sun_net} dir in ptype main tmpl src ${moon} dst ${sun} proto esp reqid 0x${reqid} mode tunnel
    ip xfrm policy add src ${moon_net} dst ${sun_net} dir fwd ptype main tmpl src ${moon} dst ${sun} proto esp reqid 0x${reqid} mode tunnel
    ip xfrm policy add src ${sun_net} dst ${moon_net} dir out ptype main tmpl src ${sun} dst ${moon} proto esp reqid 0x${reqid} mode tunnel
    ip xfrm policy ls
    ip link show dummy0 || ip link add type dummy
    ip link set dummy0 up
    ip addr replace ${sun_net} dev dummy0
    ip route replace ${moon_net} via ${sun_net}
EOF

ssh -p ${moon_port} ${moon} /bin/bash << EOF
    # --
    ip xfrm state flush

    ip xfrm state add src ${sun} dst ${moon} proto esp spi 0x${sun_spi} reqid 0x${reqid} mode tunnel auth sha256 0x${auth_key} enc aes 0x${enc_key}
    ip xfrm state add src ${moon} dst ${sun} proto esp spi 0x${moon_spi} reqid 0x${reqid} mode tunnel auth sha256 0x${auth_key} enc aes 0x${enc_key}
    ip xfrm state ls

    # --
    ip xfrm policy flush

    ip xfrm policy add src ${sun_net} dst ${moon_net} dir in ptype main tmpl src ${sun} dst ${moon} proto esp reqid 0x${reqid} mode tunnel
    ip xfrm policy add src ${sun_net} dst ${moon_net} dir fwd ptype main tmpl src ${sun} dst ${moon} proto esp reqid 0x${reqid} mode tunnel
    ip xfrm policy add src ${moon_net} dst ${sun_net} dir out ptype main tmpl src ${moon} dst ${sun} proto esp reqid 0x${reqid} mode tunnel
    ip xfrm policy ls
    ip link show dummy0 || ip link add type dummy
    ip link set dummy0 up
    ip addr replace ${moon_net} dev dummy0
    ip route replace ${sun_net} via ${moon_net}
EOF
