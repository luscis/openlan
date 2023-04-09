#!/bin/bash

set -ex

command=$1; shift
options=$@;

dir=$(dirname $0)

OVSDB_SERVER_BIN="/usr/bin/env ovsdb-server"
OVSDB_TOOL_BIN="/usr/bin/env ovsdb-tool"
[ "$OVSDB_DATABASE_SCH" == "" ] && OVSDB_DATABASE_SCH="/etc/openlan/switch/confd.schema.json"
[ "$OVSDB_DATABASE" == "" ] && OVSDB_DATABASE="/etc/openlan/switch/confd.db"
[ "$OVSDB_LOG_FILE" == "" ] && OVSDB_LOG_FILE="/var/openlan/confd/confd.log"
[ "$OVSDB_SOCK" == "" ] && OVSDB_SOCK="/var/openlan/confd/confd.sock"
[ "$OVSDB_PID_FILE" == "" ] && OVSDB_PID_FILE="/var/openlan/confd/confd.pid"

function stop() {
    [ -e "$OVSDB_PID_FILE" ] && kill "$(cat $OVSDB_PID_FILE)"
}

function start() {
    [ -e "$OVSDB_DATABASE" ] || {
        $OVSDB_TOOL_BIN create $OVSDB_DATABASE $OVSDB_DATABASE_SCH
    }

    set +x
    set $OVSDB_SERVER_BIN $OVSDB_DATABASE
    set "$@" -vconsole:info -vsyslog:off -vfile:off
    set "$@" --remote=punix:"$OVSDB_SOCK"
    set "$@" --log-file="$OVSDB_LOG_FILE"
    set "$@" --pidfile="$OVSDB_PID_FILE"
    [ "$OVSDB_OPTIONS" != "" ] && set "$@" $OVSDB_OPTIONS
    for opt in $options; do
        set "$@" $opt
    done
    set -x
    export OVS_RUNDIR="/var/openlan/confd"
    exec "$@"
}

case $command in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    *)
        echo >&2 "$0: unknown command \"$command\" (start/stop/restart)"
        exit 1
        ;;
esac
