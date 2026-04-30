#!/bin/bash

set -ex

cs_dir="/etc/openlan/switch"

function prepare() {
    sysctl -p /etc/sysctl.d/90-openlan.conf || true

    local t=""
    local c=""
    local chains=""
    local sets=""
    local s=""
    local dir=""

    for t in raw mangle filter nat; do
        chains="$(iptables -t "$t" -S 2>/dev/null | awk '/^-N (TT_|AT_|ZT_|Qos_)/ {print $2}')"
        for c in $chains; do
            iptables --wait -t "$t" -F "$c" || true
            iptables --wait -t "$t" -X "$c" || true
        done
    done

    # new directory.
    mkdir -p /var/openlan/{cert,openvpn,access,dhcp,ceci}

    sets="$(ipset list 2>/dev/null | awk '/^Name: tt/ {print $2}')"
    for s in $sets; do
        ipset destroy "$s" || true
    done

    ## START: clean older files.
    /usr/bin/env find /var/openlan/access -type f -delete
    /usr/bin/env find /var/openlan/openvpn -mindepth 2 -maxdepth 2 -type f -delete
    ## END

    ## START: prepare external dir.
    for dir in network acl findhop output route qos dnat; do
        [ -e "$cs_dir/$dir" ] || mkdir -p "$cs_dir/$dir"
    done
    ## END

    [ -e "$cs_dir/switch.json" ] || cat > "$cs_dir/switch.json" << EOF
{
    "crypt": {
        "secret": "cb2ff088a34d"
    }
}
EOF

    ## START: install default network
    [ -e "$cs_dir/network/ipsec.json" ] || cat > "$cs_dir/network/ipsec.json" << EOF
{
    "name": "ipsec",
    "provider": "ipsec",
    "snat": "disable"
}
EOF

    [ -e "$cs_dir/network/bgp.json" ] || cat > "$cs_dir/network/bgp.json" << EOF
{
    "name": "bgp",
    "provider": "bgp",
    "snat": "disable"
}
EOF

    [ -e "$cs_dir/network/ceci.json" ] || cat > "$cs_dir/network/ceci.json" << EOF
{
    "name": "ceci",
    "provider": "ceci",
    "snat": "disable"
}
EOF

    [ -e "$cs_dir/network/router.json" ] || cat > "$cs_dir/network/router.json" << EOF
{
    "name": "router",
    "provider": "router",
    "snat": "disable"
}
EOF
    ## END
}

# wai 120s for ipsec to be ready, otherwise switch may fail to start and cause a reboot loop.
max_wait=120
function wait_ipsec() {
    local begin=$(date +%s)
    local now=0
    local waited=0
    local interval=5

    while ! ipsec status > /dev/null 2>&1; do
        now=$(date +%s)
        waited=$(( now - begin ))
        if [ "$waited" -ge "$max_wait" ]; then
            echo "WARN: ipsec is still unavailable after ${max_wait}s, continue startup."
            return 0
        fi
        echo "INFO: waiting for ipsec to be ready, waited ${waited}s so far ..."
        sleep "$interval"
    done

    echo "INFO: ipsec is ready after ${waited}s."
}

child=0
jobs=0
running="yes"
last=0

function handler_exit() {
    running="no"
    if [ "$child" -gt 0 ] && kill -0 "$child" 2>/dev/null; then
        kill "$child" 2>/dev/null || true
        wait "$child" 2>/dev/null || true
    fi
    if [ "$jobs" -gt 0 ] && kill -0 "$jobs" 2>/dev/null; then
        kill "$jobs" 2>/dev/null || true
        wait "$jobs" 2>/dev/null || true
    fi
}

options="-conf:dir $cs_dir -log:level 20"

function start_switch {
    last=$(date +%s)
    echo "INFO: start openlan-switch $options"
    /usr/bin/openlan-switch $options & child=$!
    local pid=$child
    wait "$pid"
    local rc=$?
    child=0
    echo "WARN: openlan-switch exited with code $rc"
    return "$rc"
}

function set_cpus() {
    local state_file="/tmp/lastpids"
    local lastpids=""
    local switch_pid=""
    local pids=""
    local nowpids=""
    local offset=1
    local total_cpus=0
    local worker_cpus=0
    local c=0
    local pid=""

    if [ -e "$state_file" ]; then
        lastpids="$(<"$state_file")"
    fi

    switch_pid="$(pidof openlan-switch 2>/dev/null | awk '{print $1}')"
    [ -n "$switch_pid" ] || return

    pids="$(pidof openvpn 2>/dev/null | tr ' ' '\n' | awk '/^[0-9]+$/' | sort -n | xargs)"
    [ -n "$pids" ] || return

    nowpids="$switch_pid $pids"
    [ "$lastpids" != "$nowpids" ] || return

    # Set openlan-switch to cpu0.
    echo "$nowpids" > "$state_file"
    taskset -pc 0 "$switch_pid"

    # Spread openvpn daemons on remaining CPUs.
    total_cpus="$(nproc)"
    worker_cpus=$(( total_cpus - offset ))
    if [ "$worker_cpus" -le 0 ]; then
        echo "WARN: skip openvpn cpu binding, no cpu left after reserving cpu0."
        return
    fi

    for pid in $pids; do
        taskset -pc $(( c + offset )) "$pid"
        c=$(( (c + 1) % worker_cpus ))
    done
}

function start_jobs() {
    local cpus=0
    local interval=10

    cpus="$(nproc)"
    if [ "$cpus" -lt 4 ]; then
        echo "INFO: skip cpu affinity jobs, cpu cores=$cpus (<4)."
        return
    fi

    if [ -e /tmp/lastpids ]; then
        rm -f /tmp/lastpids
    fi

    while [ "$running" = "yes" ]; do
        set_cpus
        sleep "$interval"
    done
}

prepare

set +ex
wait_ipsec

trap handler_exit SIGINT SIGTERM
start_jobs & jobs=$!
while [ "$running"x == "yes"x ]; do
    start_switch
    if [ "$running"x != "yes"x ]; then
        break
    fi
    now=$(date +%s)
    during=$(( now - last ))
    if [ $during -lt 5 ]; then
        echo "INFO: suppress reboot loop, wait 5s before restart."
        sleep 5
    fi
done
