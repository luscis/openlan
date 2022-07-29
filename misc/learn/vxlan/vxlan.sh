#!/usr/bin/env bash

set -ex

vni="100"
local="192.168.7.41"
remote="192.168.7.42"
port="4789"

ssh ${local} /bin/bash <<EOF
set -ex

ip link add vxlan${vni} type vxlan id ${vni} remote ${remote} local ${local} dstport ${port}
ip link set vxlan${vni} up

EOF

ssh ${remote} /bin/bash <<EOF
set -ex

ip link add vxlan${vni} type vxlan id ${vni} remote ${local} local ${remote} dstport ${port}
ip link set vxlan${vni} up

EOF
