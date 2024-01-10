#!/bin/bash

set -ex

name=$1

if [ "$name"x == ""x ]; then
    exit 0
fi

if ip link show dev $name > /dev/null; then
    ip link set master $name dev $dev
fi
