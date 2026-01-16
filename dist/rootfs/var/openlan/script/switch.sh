#!/bin/bash

set -ex

cs_dir="/etc/openlan/switch"

function prepare() {
    sysctl -p /etc/sysctl.d/90-openlan.conf

    ## START: clean older files.
    /usr/bin/env find /var/openlan/access -type f -delete
    /usr/bin/env find /var/openlan/openvpn -type f -mindepth 2 -maxdepth 2 -delete
    ## END

    ## START: prepare external dir.
    for dir in network acl findhop output route qos dnat; do
        [ -e "$cs_dir/$dir" ] || mkdir -p "$cs_dir/$dir"
    done
    ## END

    [ -e $cs_dir/switch.json ] || cat > $cs_dir/switch.json << EOF
{
    "crypt": {
        "secret": "cb2ff088a34d"
    }
}
EOF

    ## START: install default network
    [ -e $cs_dir/network/ipsec.json ] || cat > $cs_dir/network/ipsec.json << EOF
{
    "name": "ipsec",
    "provider": "ipsec",
    "snat": "disable"
}
EOF

    [ -e $cs_dir/network/bgp.json ] || cat > $cs_dir/network/bgp.json << EOF
{
    "name": "bgp",
    "provider": "bgp",
    "snat": "disable"
}
EOF

    [ -e $cs_dir/network/ceci.json ] || cat > $cs_dir/network/ceci.json << EOF
{
    "name": "ceci",
    "provider": "ceci",
    "snat": "disable"
}
EOF

    [ -e $cs_dir/network/router.json ] || cat > $cs_dir/network/router.json << EOF
{
    "name": "router",
    "provider": "router",
    "snat": "disable"
}
EOF
    ## END
}

function wait_ipsec() {
    while ! ipsec status; do
        sleep 5
    done
}

child=0
jobs=0
running="yes"
last=0

function handler_exit() {
    kill $child & wait $child
    running="no"
}

options="-conf:dir $cs_dir -log:level 20"

function start_switch {
    echo "exec openlan-switch $options"
    exec /usr/bin/openlan-switch $options & child=$!
    trap handler_exit SIGINT SIGTERM
    last=$(date +%s)
    wait $child
}

function set_cpus() {
    local lastpids=""
    local nowpids=""
    if [ -e /tmp/lastpids ]; then
        lastpids=$(cat /tmp/lastpids)
    fi

    local pid=$(pidof openlan-switch)
    if [ "$pid"x == ""x ]; then
        return
    fi

    local pids=$(pidof openvpn | xargs -n1 | sort -n | xargs)
    if [ "$pids"x == ""x ]; then
        return
    fi

    nowpids="$pid $pids"
    if [ "$lastpids"x == "$nowpids"x ]; then
       return
    fi

    # Set openlan-switch.
    echo "$nowpids" > /tmp/lastpids
    taskset -pc 0 $pid # set switch affinity to cpu0
    local offset=1

    # Set openvpn daemon.
    local c=0
    local cpus=$(nproc)
    local cpus=$(( cpus - offset ))
    for pid in $pids; do
        taskset -pc $((c + offset )) $pid
        c=$(( (c + 1) % cpus ));
    done
}

function start_jobs() {
    local cpus=$(nproc)
    if [ $cpus -lt 4 ]; then
        return
    fi

    if [ -e /tmp/lastpids ]; then
        rm -vf /tmp/lastpids
    fi

    while [ true ]; do
        sleep 10
        set_cpus
    done
}

prepare
wait_ipsec

set +ex
start_jobs & jobs=$!
start_switch
while [ "$running"x == "yes"x ]; do
    now=$(date +%s)
    during=$(( now - last ))
    if [ $during -lt 5 ]; then
        echo "Supress booting switch after 5s."
        sleep 5
    fi
    start_switch
done