#!/bin/bash

echo "# Set process: "
ps -aux | grep 'openlan-switch' | grep -v grep
ps -aux | grep 'openvpn --cd' | grep -v grep

echo "# Doing taskset:"

cpus=$(nproc)
taskset -pc 0 $(pidof openlan-switch)

c=1
for pid in `pidof openvpn`; do
	taskset -pc $c $pid
	c=$(( (c + 1) % cpus ));
done
