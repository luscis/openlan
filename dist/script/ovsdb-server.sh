#!/bin/bash

set -ex

if [ ! -f "/etc/openvswitch/conf.db" ]; then
  ovsdb-tool create /etc/openvswitch/conf.db
elif ovsdb-tool needs-conversion /etc/openvswitch/conf.db | grep -s -w yes; then
  ovsdb-tool convert /etc/openvswitch/conf.db
fi

/usr/sbin/ovsdb-server /etc/openvswitch/conf.db \
  -vconsole:info -vsyslog:off -vfile:off \
  --remote=punix:/run/openvswitch/db.sock \
  --remote=db:Open_vSwitch,Open_vSwitch,manager_options \
  --pidfile
