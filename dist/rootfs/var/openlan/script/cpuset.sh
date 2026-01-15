#!/bin/bash

echo "# Set process: "

ps -aux | grep 'openlan-switch' | grep -v grep
ps -aux | grep 'openvpn --cd' | grep -v grep

echo "# Doing taskset:"

cpus=$(nproc)
taskset -pc 0 $(pidof openlan-switch)
offset=1 # move 1 cpu already for switch.

c=0
cpus=$(( cpus - offset ))
pids=$(pidof openvpn | tr ' ' '\n' | sort -n)
for pid in $pids; do
	taskset -pc $((c + offset )) $pid
	c=$(( (c + 1) % cpus ));
done