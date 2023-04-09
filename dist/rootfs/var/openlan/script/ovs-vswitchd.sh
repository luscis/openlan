#!/bin/bash

set -ex

exec /usr/sbin/ovs-vswitchd unix:/run/openvswitch/db.sock \
  -vconsole:info -vsyslog:off -vfile:off --mlockall \
  --pidfile
